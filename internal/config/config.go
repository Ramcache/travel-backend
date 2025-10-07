package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	AppPort     string
	JWTSecret   string
	JWTTTL      time.Duration
	DB          DBConfig
	TG          TelegramConfig
	FrontendURL string

	AppBaseURL  string
	UploadDir   string
	MaxUploadMB int
}

type DBConfig struct {
	URL         string
	MaxConns    int32
	MinConns    int32
	ConnTimeout time.Duration
	IdleTimeout time.Duration
}

type TelegramConfig struct {
	TelegramToken string
	TelegramChat  string
}

func Load() *Config {
	// .env не обязателен, просто пробуем загрузить
	if err := godotenv.Load(); err == nil {
		log.Println("ℹ️ .env loaded")
	} else {
		log.Println("ℹ️ .env not found, using system env only")
	}

	ttl := getEnvDuration("JWT_TTL", 24*time.Hour)
	log.Println("👉 JWT_TTL loaded as:", ttl)

	dbURL := getEnv("DB_URL", "")
	// Автоматическая подмена localhost → host.docker.internal внутри контейнера
	if isRunningInDocker() && strings.Contains(dbURL, "localhost") {
		dbURL = strings.ReplaceAll(dbURL, "localhost", "host.docker.internal")
		log.Println("🔄 DB_URL adjusted for Docker:", dbURL)
	}

	return &Config{
		AppEnv:      getEnv("APP_ENV", "dev"),
		AppPort:     getEnv("APP_PORT", "8080"),
		JWTSecret:   getEnv("APP_JWT_SECRET", "changeme"),
		JWTTTL:      ttl,
		FrontendURL: getEnv("FRONTEND_URL", ""),
		AppBaseURL:  getEnv("APP_BASE_URL", "http://localhost:8080"),
		UploadDir:   getEnv("UPLOAD_DIR", "uploads"),
		MaxUploadMB: int(getEnvInt("MAX_UPLOAD_MB", 20)),
		DB: DBConfig{
			URL:         dbURL,
			MaxConns:    getEnvInt("DB_MAX_CONNS", 10),
			MinConns:    getEnvInt("DB_MIN_CONNS", 2),
			ConnTimeout: getEnvDuration("DB_CONN_TIMEOUT", 5*time.Second),
			IdleTimeout: getEnvDuration("DB_IDLE_TIMEOUT", 5*time.Minute),
		},
		TG: TelegramConfig{
			TelegramToken: getEnv("TG_TOKEN", ""),
			TelegramChat:  getEnv("TG_CHAT", ""),
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

var cgroupFilePath = "/proc/1/cgroup"

func isRunningInDocker() bool {
	if f, err := os.ReadFile(cgroupFilePath); err == nil {
		return strings.Contains(string(f), "docker")
	}
	return false
}
