package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	const key = "TEST_CONFIG_GET_ENV"
	os.Unsetenv(key)

	if got := getEnv(key, "default"); got != "default" {
		t.Fatalf("expected default value, got %q", got)
	}

	if err := os.Setenv(key, "value"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	defer os.Unsetenv(key)

	if got := getEnv(key, "default"); got != "value" {
		t.Fatalf("expected overridden value, got %q", got)
	}
}

func TestGetEnvInt(t *testing.T) {
	const key = "TEST_CONFIG_GET_ENV_INT"
	os.Unsetenv(key)

	if got := getEnvInt(key, 42); got != 42 {
		t.Fatalf("expected default value, got %d", got)
	}

	if err := os.Setenv(key, "100"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	if got := getEnvInt(key, 42); got != 100 {
		t.Fatalf("expected parsed value, got %d", got)
	}

	if err := os.Setenv(key, "not-a-number"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	if got := getEnvInt(key, 7); got != 7 {
		t.Fatalf("expected fallback on parse error, got %d", got)
	}

	os.Unsetenv(key)
}

func TestGetEnvDuration(t *testing.T) {
	const key = "TEST_CONFIG_GET_ENV_DURATION"
	os.Unsetenv(key)

	defaultDuration := 30 * time.Second
	if got := getEnvDuration(key, defaultDuration); got != defaultDuration {
		t.Fatalf("expected default duration, got %s", got)
	}

	if err := os.Setenv(key, "45s"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	if got := getEnvDuration(key, defaultDuration); got != 45*time.Second {
		t.Fatalf("expected parsed duration, got %s", got)
	}

	if err := os.Setenv(key, "not-a-duration"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	if got := getEnvDuration(key, defaultDuration); got != defaultDuration {
		t.Fatalf("expected fallback on parse error, got %s", got)
	}

	os.Unsetenv(key)
}

func TestIsRunningInDocker(t *testing.T) {
	originalPath := cgroupFilePath
	t.Cleanup(func() { cgroupFilePath = originalPath })

	dir := t.TempDir()
	file := filepath.Join(dir, "cgroup")

	if err := os.WriteFile(file, []byte("12:freezer:/docker/123"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cgroupFilePath = file
	if !isRunningInDocker() {
		t.Fatal("expected docker detection to return true")
	}

	if err := os.WriteFile(file, []byte("12:freezer:/container"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if isRunningInDocker() {
		t.Fatal("expected docker detection to return false")
	}
}

func TestLoad(t *testing.T) {
	originalPath := cgroupFilePath
	t.Cleanup(func() { cgroupFilePath = originalPath })

	dir := t.TempDir()
	file := filepath.Join(dir, "cgroup")
	if err := os.WriteFile(file, []byte("docker"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	cgroupFilePath = file

	envs := map[string]string{
		"APP_ENV":         "prod",
		"APP_PORT":        "9090",
		"APP_JWT_SECRET":  "supersecret",
		"JWT_TTL":         "1h",
		"FRONTEND_URL":    "https://example.com",
		"DB_URL":          "postgres://localhost:5432/app",
		"DB_MAX_CONNS":    "20",
		"DB_MIN_CONNS":    "5",
		"DB_CONN_TIMEOUT": "10s",
		"DB_IDLE_TIMEOUT": "2m",
		"TG_TOKEN":        "token",
		"TG_CHAT":         "chat",
	}

	for k, v := range envs {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("setenv %s: %v", k, err)
		}
		t.Cleanup(func(key string) func() {
			return func() { os.Unsetenv(key) }
		}(k))
	}

	cfg := Load()

	if cfg.AppEnv != "prod" {
		t.Fatalf("expected AppEnv prod, got %s", cfg.AppEnv)
	}
	if cfg.AppPort != "9090" {
		t.Fatalf("expected AppPort 9090, got %s", cfg.AppPort)
	}
	if cfg.JWTSecret != "supersecret" {
		t.Fatalf("expected JWTSecret supersecret, got %s", cfg.JWTSecret)
	}
	if cfg.JWTTTL != time.Hour {
		t.Fatalf("expected JWTTTL 1h, got %s", cfg.JWTTTL)
	}
	if cfg.FrontendURL != "https://example.com" {
		t.Fatalf("expected FrontendURL https://example.com, got %s", cfg.FrontendURL)
	}

	if !strings.Contains(cfg.DB.URL, "host.docker.internal") {
		t.Fatalf("expected DB URL adjusted for docker, got %s", cfg.DB.URL)
	}
	if cfg.DB.MaxConns != 20 {
		t.Fatalf("expected DB MaxConns 20, got %d", cfg.DB.MaxConns)
	}
	if cfg.DB.MinConns != 5 {
		t.Fatalf("expected DB MinConns 5, got %d", cfg.DB.MinConns)
	}
	if cfg.DB.ConnTimeout != 10*time.Second {
		t.Fatalf("expected DB ConnTimeout 10s, got %s", cfg.DB.ConnTimeout)
	}
	if cfg.DB.IdleTimeout != 2*time.Minute {
		t.Fatalf("expected DB IdleTimeout 2m, got %s", cfg.DB.IdleTimeout)
	}
	if cfg.TG.TelegramToken != "token" {
		t.Fatalf("expected TG token token, got %s", cfg.TG.TelegramToken)
	}
	if cfg.TG.TelegramChat != "chat" {
		t.Fatalf("expected TG chat chat, got %s", cfg.TG.TelegramChat)
	}
}
