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
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/microservices-go/shared/cache"
	"github.com/microservices-go/shared/config"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/middleware"
	"github.com/microservices-go/shared/migrate"
	"github.com/microservices-go/shared/rabbitmq"
	"github.com/microservices-go/shared/redis"

	"github.com/microservices-go/services/order/internal/order"
	"github.com/microservices-go/services/order/internal/rabbit"
)

func main() {
	// Initialize logger
	log := logger.New("order-service")
	log.Info("Starting Order Service...")

	// Load configuration
	dbConfig := config.LoadDatabaseConfig("order")
	serverConfig := config.LoadServerConfig("order")
	jwtConfig := config.LoadJWTConfig()
	rabbitConfig := config.LoadRabbitMQConfig()
	redisConfig := config.LoadRedisConfig()

	// Connect to Redis
	redisClient, err := redis.NewClient(redisConfig)
	if err != nil {
		log.Warn("Failed to connect to Redis: " + err.Error())
		log.Warn("Rate limiting will be disabled")
		redisClient = nil
	} else {
		defer redisClient.Close()
		log.Info("Connected to Redis")
	}

	// Load rate limit config
	rateLimitConfig := config.LoadRateLimitConfig("order")

	// Create rate limiter and cache
	var rateLimiter *middleware.RateLimiter
	var cacheClient *cache.Cache
	if redisClient != nil {
		rateLimiter = middleware.NewRateLimiter(redisClient, rateLimitConfig, "order")
		cacheClient = cache.NewCache(redisClient.GetClient(), "order")
		log.Infof("Rate limiting enabled: %d req/min", rateLimitConfig.RequestsPerMinute)
		log.Info("Caching enabled")
	}

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

	// Run database migrations (if enabled)
	if shouldAutoMigrate() {
		if err := runMigrations(sqlDB, log); err != nil {
			log.Fatal("Failed to run migrations: " + err.Error())
		}
	} else {
		log.Info("Auto-migrate disabled. Skipping migrations.")
	}

	// Connect to RabbitMQ
	var rabbitClient *rabbitmq.Client
	var publisher order.EventPublisher
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

	// Initialize repository
	orderRepo := order.NewRepository(db)

	// Initialize service
	orderService := order.NewService(orderRepo, publisher, cacheClient)

	// Initialize consumer
	if rabbitClient != nil {
		consumer = rabbit.NewConsumer(orderService)
		if err := consumer.Start(rabbitClient); err != nil {
			log.Warn("Failed to start consumer: " + err.Error())
		} else {
			log.Info("Started RabbitMQ consumer")
		}
	}

	// Initialize handler
	orderHandler := order.NewHandler(orderService)

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

	// Rate limiting middleware
	if rateLimiter != nil {
		r.Use(rateLimiter.Middleware)
	}

	// Health check
	r.Get("/health", orderHandler.HealthCheck)

	// API routes
	orderHandler.RegisterRoutes(r, authMiddleware)

	// Create server
	srv := &http.Server{
		Addr:         ":" + getEnv("ORDER_PORT", "8080"),
		Handler:      r,
		ReadTimeout:  time.Duration(serverConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(serverConfig.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Server starting on port %s", getEnv("ORDER_PORT", "8080"))
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

func shouldAutoMigrate() bool {
	return getEnv("ORDER_AUTO_MIGRATE", "true") == "true"
}

func runMigrations(sqlDB *sql.DB, log *logger.Logger) error {
	migrator := migrate.NewMigrator(sqlDB, log.WithField("component", "migrator"))
	return migrator.Run("./migrations")
}
