package config

import (
	"os"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				DBHost:     "localhost",
				DBUser:     "user",
				DBPassword: "password",
				DBName:     "wallet_db",
				JWTSecret:  "12345678901234567890123456789012",
			},
			wantErr: false,
		},
		{
			name: "missing DBHost",
			cfg: Config{
				DBUser:     "user",
				DBPassword: "password",
				DBName:     "wallet_db",
				JWTSecret:  "12345678901234567890123456789012",
			},
			wantErr: true,
		},
		{
			name: "missing DBUser",
			cfg: Config{
				DBHost:     "localhost",
				DBPassword: "password",
				DBName:     "wallet_db",
				JWTSecret:  "12345678901234567890123456789012",
			},
			wantErr: true,
		},
		{
			name: "missing DBPassword",
			cfg: Config{
				DBHost:    "localhost",
				DBUser:    "user",
				DBName:    "wallet_db",
				JWTSecret: "12345678901234567890123456789012",
			},
			wantErr: true,
		},
		{
			name: "missing DBName",
			cfg: Config{
				DBHost:     "localhost",
				DBUser:     "user",
				DBPassword: "password",
				JWTSecret:  "12345678901234567890123456789012",
			},
			wantErr: true,
		},
		{
			name: "short JWTSecret",
			cfg: Config{
				DBHost:     "localhost",
				DBUser:     "user",
				DBPassword: "password",
				DBName:     "wallet_db",
				JWTSecret:  "too-short",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Load(t *testing.T) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "wallet_user")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "wallet_db")
	os.Setenv("JWT_SECRET", "super-secret-jwt-key-with-at-least-32-characters")
	os.Setenv("JWT_TTL", "30m")
	os.Setenv("BCRYPT_COST", "10")
	os.Setenv("INITIAL_BALANCE", "50000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.JWTTTL != 30*time.Minute {
		t.Errorf("expected JWTTTL=30m, got %v", cfg.JWTTTL)
	}
	if cfg.BcryptCost != 10 {
		t.Errorf("expected BcryptCost=10, got %d", cfg.BcryptCost)
	}
	if cfg.InitialBalance != 50000 {
		t.Errorf("expected InitialBalance=50000, got %d", cfg.InitialBalance)
	}
}
