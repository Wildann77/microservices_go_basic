package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	// Try to load .env from current directory or parent directories
	// This is useful for local development
	currDir, _ := os.Getwd()
	for {
		envPath := filepath.Join(currDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			break
		}
		parentDir := filepath.Dir(currDir)
		if parentDir == currDir {
			break
		}
		currDir = parentDir
	}
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN returns PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

// URL returns RabbitMQ connection URL
func (c *RabbitMQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.User, c.Password, c.Host, c.Port, c.VHost)
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret    string
	ExpiresIn int // hours
	Issuer    string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
}

// LoadDatabaseConfig loads database config from environment
func LoadDatabaseConfig(service string) *DatabaseConfig {
	prefix := strings.ToUpper(service)
	return &DatabaseConfig{
		Host:     getEnv(fmt.Sprintf("%s_DB_HOST", prefix), "localhost"),
		Port:     getEnvAsInt(fmt.Sprintf("%s_DB_PORT", prefix), 5432),
		User:     getEnv(fmt.Sprintf("%s_DB_USER", prefix), "postgres"),
		Password: getEnv(fmt.Sprintf("%s_DB_PASSWORD", prefix), "password"),
		DBName:   getEnv(fmt.Sprintf("%s_DB_NAME", prefix), service),
		SSLMode:  getEnv(fmt.Sprintf("%s_DB_SSLMODE", prefix), "disable"),
	}
}

// LoadServerConfig loads server config from environment
func LoadServerConfig(service string) *ServerConfig {
	prefix := strings.ToUpper(service)
	return &ServerConfig{
		Port:         getEnvAsInt(fmt.Sprintf("%s_PORT", prefix), 8080),
		ReadTimeout:  getEnvAsInt(fmt.Sprintf("%s_READ_TIMEOUT", prefix), 10),
		WriteTimeout: getEnvAsInt(fmt.Sprintf("%s_WRITE_TIMEOUT", prefix), 10),
	}
}

// LoadRabbitMQConfig loads RabbitMQ config from environment
func LoadRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		Host:     getEnv("RABBITMQ_HOST", "localhost"),
		Port:     getEnvAsInt("RABBITMQ_PORT", 5672),
		User:     getEnv("RABBITMQ_USER", "guest"),
		Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		VHost:    getEnv("RABBITMQ_VHOST", "/"),
	}
}

// LoadJWTConfig loads JWT config from environment
func LoadJWTConfig() *JWTConfig {
	return &JWTConfig{
		Secret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		ExpiresIn: getEnvAsInt("JWT_EXPIRES_IN", 24),
		Issuer:    getEnv("JWT_ISSUER", "microservices-go"),
	}
}

// LoadRateLimitConfig loads rate limit config from environment
func LoadRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerSecond: getEnvAsInt("RATE_LIMIT_RPS", 100),
		BurstSize:         getEnvAsInt("RATE_LIMIT_BURST", 150),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
