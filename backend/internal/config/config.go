package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	AppPort            string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecret          string
	JWTIssuer          string
	JWTTTL             time.Duration
	BcryptCost         int
	CORSAllowedOrigins []string
	InitialBalance     int64
	IdempotencyKeyTTL  time.Duration
}

func Load() (*Config, error) {
	// Optionally load .env from root or current dir, ignore error if missing
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg := &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", "8080"),
		DBHost:            os.Getenv("DB_HOST"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            os.Getenv("DB_USER"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBName:            os.Getenv("DB_NAME"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTIssuer:         getEnv("JWT_ISSUER", "jwt-auth-wallet"),
		JWTTTL:            15 * time.Minute,
		BcryptCost:        12,
		InitialBalance:    100000,
		IdempotencyKeyTTL: 24 * time.Hour,
	}

	if ttlStr := os.Getenv("JWT_TTL"); ttlStr != "" {
		if d, err := time.ParseDuration(ttlStr); err == nil {
			cfg.JWTTTL = d
		} else {
			return nil, fmt.Errorf("invalid JWT_TTL: %w", err)
		}
	}

	if costStr := os.Getenv("BCRYPT_COST"); costStr != "" {
		if cost, err := strconv.Atoi(costStr); err == nil {
			cfg.BcryptCost = cost
		} else {
			return nil, fmt.Errorf("invalid BCRYPT_COST: %w", err)
		}
	}

	originsStr := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	parts := strings.Split(originsStr, ",")
	cfg.CORSAllowedOrigins = make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, trimmed)
		}
	}

	if balStr := os.Getenv("INITIAL_BALANCE"); balStr != "" {
		if bal, err := strconv.ParseInt(balStr, 10, 64); err == nil {
			cfg.InitialBalance = bal
		} else {
			return nil, fmt.Errorf("invalid INITIAL_BALANCE: %w", err)
		}
	}

	if ikTTLStr := os.Getenv("IDEMPOTENCY_KEY_TTL"); ikTTLStr != "" {
		if d, err := time.ParseDuration(ikTTLStr); err == nil {
			cfg.IdempotencyKeyTTL = d
		} else {
			return nil, fmt.Errorf("invalid IDEMPOTENCY_KEY_TTL: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
