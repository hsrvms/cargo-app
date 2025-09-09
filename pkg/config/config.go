package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	SafeCubeAPI    SafeCubeAPIConfig
	BackgroundJobs BackgroundJobsConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type SafeCubeAPIConfig struct {
	BaseURL string
	APIKey  string
}

type BackgroundJobsConfig struct {
	ShipmentRefreshInterval     time.Duration
	ShipmentRefreshWorkers      int
	ShipmentMaxPerRun           int
	ShipmentSkipRecentlyUpdated time.Duration
}

func New() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "go-starter"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		SafeCubeAPI: SafeCubeAPIConfig{
			BaseURL: getEnv("SAFECUBE_API_BASE_URL", ""),
			APIKey:  getEnv("SAFECUBE_API_KEY", ""),
		},
		BackgroundJobs: BackgroundJobsConfig{
			ShipmentRefreshInterval:     getEnvAsDuration("SHIPMENT_REFRESH_INTERVAL", 3*time.Hour),
			ShipmentRefreshWorkers:      getEnvAsInt("SHIPMENT_REFRESH_WORKERS", 5),
			ShipmentMaxPerRun:           getEnvAsInt("SHIPMENT_MAX_PER_RUN", 0),
			ShipmentSkipRecentlyUpdated: getEnvAsDuration("SHIPMENT_SKIP_RECENTLY_UPDATED", 30*time.Minute),
		},
	}
}

func (c *Config) GetDBConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
