package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const envFile = "api/config.env"

type Config struct {
	DB                 DBConfig
	Keycloak           KeycloakConfig
	WebServerPort      string
	CORSAllowedOrigins []string
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
	Issuer   string
	Audience string
	BaseURL  string
	Realm    string
	Timeout  int
	JWKSURI  string
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.ConnectTimeout,
	)
}

func Load() (*Config, error) {
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config: failed to load env file %q: %w", envFile, err)
	}
	keycloak, err := loadKeycloak()
	if err != nil {
		return nil, err
	}
	origins, err := loadCORSOrigins()
	if err != nil {
		return nil, err
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
		Keycloak:           keycloak,
		WebServerPort:      getEnv("WEBSERVER_PORT", "8081"),
		CORSAllowedOrigins: origins,
	}

	return cfg, nil
}

func loadCORSOrigins() ([]string, error) {
	origins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	invalid := errors.New("config: CORS_ALLOWED_ORIGINS must contain only complete HTTP(S) origins")
	for i, origin := range origins {
		origin = strings.TrimSpace(origin)
		u, err := url.Parse(origin)
		if err != nil {
			return nil, invalid
		}
		hasHTTP := u.Scheme == "http" || u.Scheme == "https"
		hasHost := u.Hostname() != "" && !strings.ContainsAny(u.Host, "* \t\n\r")
		hasExtra := u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || strings.Contains(origin, "#")
		if !hasHTTP || !hasHost || hasExtra || strings.HasSuffix(u.Host, ":") {
			return nil, invalid
		}
		if strings.Contains(u.Hostname(), ":") && !strings.HasPrefix(u.Host, "[") {
			return nil, invalid
		}
		if port := u.Port(); port != "" {
			n, err := strconv.Atoi(port)
			if err != nil || n < 1 || n > 65535 {
				return nil, invalid
			}
		}
		origins[i] = u.String()
	}
	return origins, nil
}

func loadKeycloak() (KeycloakConfig, error) {
	issuer := os.Getenv("KEYCLOAK_ISSUER")
	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return KeycloakConfig{}, fmt.Errorf("config: invalid KEYCLOAK_ISSUER: %w", err)
	}
	hasHTTP := issuerURL.Scheme == "http" || issuerURL.Scheme == "https"
	hasAuthority := issuerURL.Hostname() != "" && issuerURL.User == nil
	hasExtra := issuerURL.RawQuery != "" || issuerURL.Fragment != "" || issuerURL.ForceQuery
	if !hasHTTP || !hasAuthority || hasExtra {
		return KeycloakConfig{}, errors.New("config: KEYCLOAK_ISSUER must be an absolute HTTP(S) URL")
	}
	audience := os.Getenv("API_AUDIENCE")
	if strings.TrimSpace(audience) == "" {
		return KeycloakConfig{}, errors.New("config: API_AUDIENCE is required")
	}
	timeout, err := strconv.Atoi(getEnv("KEYCLOAK_TIMEOUT_SECONDS", "5"))
	if err != nil {
		return KeycloakConfig{}, fmt.Errorf("config: invalid KEYCLOAK_TIMEOUT_SECONDS: %w", err)
	}
	if timeout <= 0 || int64(timeout) > int64((1<<63-1)/time.Second) {
		return KeycloakConfig{}, errors.New("config: KEYCLOAK_TIMEOUT_SECONDS must be positive and fit a duration")
	}
	return KeycloakConfig{
		Issuer: issuer, Audience: audience, Timeout: timeout,
		BaseURL: os.Getenv("KEYCLOAK_BASE_URL"), Realm: os.Getenv("KEYCLOAK_REALM"),
		JWKSURI: os.Getenv("KEYCLOAK_JWKS_URI"),
	}, nil
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
