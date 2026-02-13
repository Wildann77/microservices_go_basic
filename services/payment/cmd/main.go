package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/microservices-go/shared/config"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/middleware"
	"github.com/microservices-go/shared/rabbitmq"

	"github.com/microservices-go/services/payment/internal/payment"
	"github.com/microservices-go/services/payment/internal/rabbit"
)

func main() {
	// Initialize logger
	log := logger.New("payment-service")
	log.Info("Starting Payment Service...")

	// Load configuration
	dbConfig := config.LoadDatabaseConfig("payment")
	serverConfig := config.LoadServerConfig("payment")
	jwtConfig := config.LoadJWTConfig()
	rabbitConfig := config.LoadRabbitMQConfig()

	// Connect to database using GORM
	db, err := gorm.Open(postgres.Open(dbConfig.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: " + err.Error())
	}

	// Get underlying sql.DB for closing and pinging
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB: " + err.Error())
	}
	defer sqlDB.Close()

	// Test database connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database: " + err.Error())
	}
	log.Info("Connected to database via GORM")

	// Connect to RabbitMQ
	var rabbitClient *rabbitmq.Client
	var publisher payment.EventPublisher
	var consumer *rabbit.Consumer

	rabbitClient, err = rabbitmq.NewClient(rabbitConfig.URL())
	if err != nil {
		log.Warn("Failed to connect to RabbitMQ: " + err.Error())
	} else {
		defer rabbitClient.Close()

		// Declare exchange
		if err := rabbitClient.DeclareExchange("microservices.events"); err != nil {
			log.Warn("Failed to declare exchange: " + err.Error())
		} else {
			log.Info("Connected to RabbitMQ")

			// Initialize publisher
			publisher = rabbit.NewPublisher(rabbitClient)
		}
	}

	// Initialize payment provider (Stripe)
	provider := payment.NewStripeProvider()

	// Initialize repository
	paymentRepo := payment.NewRepository(db)

	// Initialize service
	paymentService := payment.NewService(paymentRepo, provider, publisher)

	// Initialize consumer
	if rabbitClient != nil {
		consumer = rabbit.NewConsumer(paymentService)
		if err := consumer.Start(rabbitClient); err != nil {
			log.Warn("Failed to start consumer: " + err.Error())
		} else {
			log.Info("Started RabbitMQ consumer")
		}
	}

	// Initialize handler
	paymentHandler := payment.NewHandler(paymentService)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtConfig)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.SecurityHeadersMiddleware)
	r.Use(middleware.CORSMiddleware([]string{"*"}))

	// Health check
	r.Get("/health", paymentHandler.HealthCheck)

	// API routes
	paymentHandler.RegisterRoutes(r, authMiddleware)

	// Create server
	srv := &http.Server{
		Addr:         ":" + getEnv("PAYMENT_PORT", "8080"),
		Handler:      r,
		ReadTimeout:  time.Duration(serverConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(serverConfig.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Server starting on port %s", getEnv("PAYMENT_PORT", "8080"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed: " + err.Error())
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: " + err.Error())
	}

	log.Info("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
