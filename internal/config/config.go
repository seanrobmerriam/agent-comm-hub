// Package config provides application configuration management.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Server  ServerConfig
	Redis   RedisConfig
	Memory  MemoryConfig
	Logging LoggingConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host string
	Port string
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	StandardURL string
	PubSubURL   string
	PoolSize    int
	MinIdleConn int
	Timeout     time.Duration
}

// MemoryConfig holds agent memory server configuration.
type MemoryConfig struct {
	URL     string
	Timeout time.Duration
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level string
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Redis: RedisConfig{
			StandardURL: getEnv("REDIS_STANDARD_URL", "redis://localhost:6379"),
			PubSubURL:   getEnv("REDIS_PUBSUB_URL", "redis://localhost:6380"),
			PoolSize:    getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConn: getEnvInt("REDIS_MIN_IDLE_CONN", 5),
			Timeout:     getEnvDuration("REDIS_TIMEOUT", 5*time.Second),
		},
		Memory: MemoryConfig{
			URL:     getEnv("AGENT_MEMORY_URL", "http://localhost:8081"),
			Timeout: getEnvDuration("AGENT_MEMORY_TIMEOUT", 10*time.Second),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
