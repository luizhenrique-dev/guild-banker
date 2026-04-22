package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const envFile = "config.env"

type Config struct {
	DB       DBConfig
	Keycloak KeycloakConfig
}

type DBConfig struct {
	Host                   string
	Port                   string
	User                   string
	Password               string
	Name                   string
	SSLMode                string
	ConnectTimeout         int
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
}

type KeycloakConfig struct {
	BaseURL string
	Realm   string
	Timeout int
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.ConnectTimeout,
	)
}

func Load() (*Config, error) {
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("config: failed to load env file %q: %w", envFile, err)
	}

	cfg := &Config{
		DB: DBConfig{
			Host:                   requireEnv("DB_HOST"),
			Port:                   requireEnv("DB_PORT"),
			User:                   requireEnv("DB_USER"),
			Password:               requireEnv("DB_PASSWORD"),
			Name:                   requireEnv("DB_NAME"),
			SSLMode:                getEnv("DB_SSLMODE", "disable"),
			ConnectTimeout:         mustInt("DB_CONNECT_TIMEOUT", 5),
			MaxOpenConns:           mustInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:           mustInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetimeMinutes: mustInt("DB_CONN_MAX_LIFETIME_MINUTES", 30),
		},
		Keycloak: KeycloakConfig{
			BaseURL: requireEnv("KEYCLOAK_BASE_URL"),
			Realm:   requireEnv("KEYCLOAK_REALM"),
			Timeout: mustInt("KEYCLOAK_TIMEOUT_SECONDS", 5),
		},
	}

	return cfg, nil
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: required env var %q is not set", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("config: env var %q must be an integer, got %q", key, v))
	}
	return n
}
