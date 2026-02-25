package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr          string
	DBUser        string
	DBPass        string
	DBHost        string
	DBPort        string
	DBName        string
	MemcachedAddr string
	AdminHost     string
	SessionTTL    time.Duration
}

func Load() Config {
	ttlHours := getEnvInt("SESSION_TTL_HOURS", 24)
	return Config{
		Addr:          getEnv("ADDR", ":8080"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPass:        getEnv("DB_PASS", "root"),
		DBHost:        getEnv("DB_HOST", "127.0.0.1"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBName:        getEnv("DB_NAME", "gasha_system"),
		MemcachedAddr: getEnv("MEMCACHED_ADDR", "127.0.0.1:11211"),
		AdminHost:     getEnv("ADMIN_HOST", ""),
		SessionTTL:    time.Duration(ttlHours) * time.Hour,
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
