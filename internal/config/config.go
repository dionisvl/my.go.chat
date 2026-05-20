package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

var Version = "dev"

type Config struct {
	App  AppConfig
	DB   DBConfig
	Chat ChatConfig
}

type AppConfig struct {
	Env             string
	Port            string // e.g. ":8080"
	TrustedOrigins  []string
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	DSN             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
}

type ChatConfig struct {
	WelcomeMessage string
	WelcomeTimeout time.Duration
	Profanities    []string
	HistoryLimit   int
}

func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:             getEnv("APP_ENV", "dev"),
			Port:            getEnv("APP_PORT", ":8080"),
			TrustedOrigins:  getEnvCSV("CORS_TRUSTED_ORIGINS", nil),
			ShutdownTimeout: getEnvDuration("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		DB: DBConfig{
			DSN:             getEnv("DB_DSN", "postgres://chat:chat@db:5432/chat?sslmode=disable"),
			MaxConns:        int32(getEnvInt("DB_MAX_CONNS", 10)),
			MinConns:        int32(getEnvInt("DB_MIN_CONNS", 2)),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", 5*time.Minute),
		},
		Chat: ChatConfig{
			WelcomeMessage: getEnv("WELCOME_MESSAGE", ""),
			WelcomeTimeout: time.Duration(getEnvInt("WELCOME_TIMEOUT", 0)) * time.Second,
			Profanities:    getEnvCSV("PROFANITIES", nil),
			HistoryLimit:   getEnvInt("CHAT_HISTORY_LIMIT", 50),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		slog.Warn("invalid int env var, using fallback", "key", key, "value", v, "fallback", fallback)
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		slog.Warn("invalid duration env var, using fallback", "key", key, "value", v, "fallback", fallback, "error", err)
		return fallback
	}
	return d
}

func getEnvCSV(key string, fallback []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
