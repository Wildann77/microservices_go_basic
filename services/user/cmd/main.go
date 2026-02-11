package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"github.com/microservices-go/shared/config"
	"github.com/microservices-go/shared/logger"
	sharedMiddleware "github.com/microservices-go/shared/middleware"
	"github.com/microservices-go/shared/rabbitmq"

	"github.com/microservices-go/services/user/internal/rabbit"
	"github.com/microservices-go/services/user/internal/user"
)

func main() {
	// Initialize logger
	log := logger.New("user-service")
	log.Info("Starting User Service...")

	// Load configuration
	dbConfig := config.LoadDatabaseConfig("user")
	serverConfig := config.LoadServerConfig("user")
	jwtConfig := config.LoadJWTConfig()
	rabbitConfig := config.LoadRabbitMQConfig()

	// Connect to database
	db, err := sql.Open("postgres", dbConfig.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database: " + err.Error())
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database: " + err.Error())
	}
	log.Info("Connected to database")

	// Connect to RabbitMQ
	rabbitClient, err := rabbitmq.NewClient(rabbitConfig.URL())
	if err != nil {
		log.Warn("Failed to connect to RabbitMQ: " + err.Error())
		// Continue without RabbitMQ
	} else {
		defer rabbitClient.Close()
		
		// Declare exchange
		if err := rabbitClient.DeclareExchange("microservices.events"); err != nil {
			log.Warn("Failed to declare exchange: " + err.Error())
		} else {
			log.Info("Connected to RabbitMQ")
		}
	}

	// Initialize repository
	userRepo := user.NewRepository(db)

	// Initialize publisher
	var publisher user.EventPublisher
	if rabbitClient != nil {
		publisher = rabbit.NewPublisher(rabbitClient)
	}

	// Initialize service
	userService := user.NewService(userRepo, jwtConfig, publisher)

	// Initialize handler
	userHandler := user.NewHandler(userService)

	// Initialize auth middleware
	authMiddleware := sharedMiddleware.NewAuthMiddleware(jwtConfig)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(sharedMiddleware.LoggingMiddleware)
	r.Use(sharedMiddleware.RecoveryMiddleware)
	r.Use(sharedMiddleware.SecurityHeadersMiddleware)
	r.Use(sharedMiddleware.CORSMiddleware([]string{"*"}))

	// Health check
	r.Get("/health", userHandler.HealthCheck)

	// API routes
	userHandler.RegisterRoutes(r, authMiddleware)

	// Create server
	srv := &http.Server{
		Addr:         ":" + getEnv("USER_PORT", "8080"),
		Handler:      r,
		ReadTimeout:  time.Duration(serverConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(serverConfig.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Server starting on port %s", getEnv("USER_PORT", "8080"))
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