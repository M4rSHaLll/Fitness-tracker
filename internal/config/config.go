package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAppEnv        = "local"
	defaultAppHost       = ""
	defaultAppPort       = 8080
	defaultStorageDriver = "memory"
	defaultDatabasePort  = 5432
	defaultDatabaseSSL   = "disable"
)

type Config struct {
	App      AppConfig
	Storage  StorageConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Env  string
	Host string
	Port int
}

func (c AppConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type StorageConfig struct {
	Driver string
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c DatabaseConfig) ConnectionString() string {
	if c.URL != "" {
		return c.URL
	}
	if c.Host == "" || c.User == "" || c.Name == "" {
		return ""
	}

	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Name,
	}

	query := dsn.Query()
	query.Set("sslmode", c.SSLMode)
	dsn.RawQuery = query.Encode()

	return dsn.String()
}

func Load() (*Config, error) {
	appPort, err := getEnvPort("APP_PORT", defaultAppPort)
	if err != nil {
		return nil, err
	}

	dbPort, err := getEnvPort("DB_PORT", defaultDatabasePort)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", defaultAppEnv),
			Host: getEnv("APP_HOST", defaultAppHost),
			Port: appPort,
		},
		Storage: StorageConfig{
			Driver: getEnv("STORAGE_DRIVER", defaultStorageDriver),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", ""),
			Host:     getEnv("DB_HOST", ""),
			Port:     dbPort,
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", defaultDatabaseSSL),
		},
	}

	return cfg, nil
}

func getEnv(name, fallback string) string {
	value, exists := os.LookupEnv(name)
	if !exists || value == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func getEnvInt(name string, fallback int) (int, error) {
	value := getEnv(name, "")
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return parsed, nil
}

func getEnvPort(name string, fallback int) (int, error) {
	port, err := getEnvInt(name, fallback)
	if err != nil {
		return 0, err
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid %s: must be between 1 and 65535", name)
	}
	return port, nil
}
