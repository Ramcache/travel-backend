package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv    string
	AppPort   string
	JWTSecret string
	DB        DBConfig
}

type DBConfig struct {
	URL         string
	MaxConns    int32
	MinConns    int32
	ConnTimeout time.Duration
	IdleTimeout time.Duration
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, using system env")
	}

	return &Config{
		AppEnv:    getEnv("APP_ENV", "dev"),
		AppPort:   getEnv("APP_PORT", "8080"),
		JWTSecret: getEnv("APP_JWT_SECRET", "changeme"),
		DB: DBConfig{
			URL:         getEnv("DB_URL", ""),
			MaxConns:    getEnvInt("DB_MAX_CONNS", 10),
			MinConns:    getEnvInt("DB_MIN_CONNS", 2),
			ConnTimeout: getEnvDuration("DB_CONN_TIMEOUT", 5*time.Second),
			IdleTimeout: getEnvDuration("DB_IDLE_TIMEOUT", 5*time.Minute),
		},
	}
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func getEnvInt(key string, def int32) int32 {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(val); err == nil {
			return int32(parsed)
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		if parsed, err := time.ParseDuration(val); err == nil {
			return parsed
		}
	}
	return def
}
