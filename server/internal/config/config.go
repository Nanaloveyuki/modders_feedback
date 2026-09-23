package config

import (
	"errors"
	"log"
	"os"
	"strings"
)

type Config struct {
	ListenAddr    string
	DataDir       string
	WebDir        string
	AdminUsername string
	AdminPassword string
	JWTSecret     string
	CookieSecure  bool
}

func Load() (Config, error) {
	cfg := Config{
		ListenAddr:    env("LISTEN_ADDR", ":8080"),
		DataDir:       env("DATA_DIR", "./data"),
		WebDir:        env("WEB_DIR", "/app/web"),
		AdminUsername: env("ADMIN_USERNAME", ""),
		AdminPassword: env("ADMIN_PASSWORD", ""),
		JWTSecret:     env("JWT_SECRET", ""),
		CookieSecure:  os.Getenv("COOKIE_SECURE") == "true",
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	path := os.Getenv(key + "_FILE")
	if path == "" {
		return fallback
	}
	value, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s: %v", key+"_FILE", err)
	}
	return strings.TrimRight(string(value), "\r\n")
}
