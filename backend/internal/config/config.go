package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config アプリケーション設定
type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DBDSN string

	// JWT
	JWTSecret string

	// CORS
	AllowedOrigins string

	// Request
	MaxRequestBodySize string

	// GCS
	GCSBucketName              string
	GoogleApplicationCredentials string

	// SMTP
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFromName string

	// Frontend
	FrontendURL string

	// Logging
	LogLevel string
}

// Load 設定を読み込む
func Load() (*Config, error) {
	// .envファイルの読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := &Config{
		Port:                       getEnv("PORT", "8080"),
		Env:                        getEnv("ENV", "development"),
		DBDSN:                      getEnv("DB_DSN", "root:password@tcp(localhost:3306)/Go-Next-WebRTC?parseTime=true"),
		JWTSecret:                  os.Getenv("JWT_SECRET"),
		AllowedOrigins:             getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		MaxRequestBodySize:         getEnv("MAX_REQUEST_BODY_SIZE", "10485760"),
		GCSBucketName:              os.Getenv("GCS_BUCKET_NAME"),
		GoogleApplicationCredentials: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		SMTPHost:                   os.Getenv("SMTP_HOST"),
		SMTPPort:                   os.Getenv("SMTP_PORT"),
		SMTPUser:                   os.Getenv("SMTP_USER"),
		SMTPPassword:               os.Getenv("SMTP_PASSWORD"),
		SMTPFromName:               os.Getenv("SMTP_FROM_NAME"),
		FrontendURL:                getEnv("FRONTEND_URL", "http://localhost:3000"),
		LogLevel:                   getEnv("LOG_LEVEL", "info"),
	}

	// 設定の検証
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate 設定の検証
func (c *Config) Validate() error {
	required := map[string]string{
		"JWT_SECRET": c.JWTSecret,
		"DB_DSN":     c.DBDSN,
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	// JWT_SECRETの長さチェック
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
