package config

import (
	"os"
	"strconv"
)

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime int64
	MaxConnIdleTime int64
}

type Config struct {
	DB          DBConfig
	AnalysisAPI string
}

func LoadConfig() *Config {
	cfg := &Config{}
	cfg.DB = DBConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "user"),
		Password:        getEnv("DB_PASSWORD", "password"),
		Name:            getEnv("DB_NAME", "db"),
		MaxConns:        getEnvAsInt32("DB_MAX_CONNS", 10),
		MinConns:        getEnvAsInt32("DB_MIN_CONNS", 2),
		MaxConnLifetime: getEnvAsInt64("DB_MAX_CONN_LIFETIME", 3600), // seconds
		MaxConnIdleTime: getEnvAsInt64("DB_MAX_CONN_IDLE_TIME", 300), // seconds
	}
	cfg.AnalysisAPI = getEnv("ANALYSIS_API_URL", "")
	if cfg.AnalysisAPI == "" {
		panic("ANALYSIS_API_URL is not set")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt32(key string, fallback int32) int32 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return fallback
	}
	if value, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
		return int32(value)
	}
	return fallback
}

func getEnvAsInt64(key string, fallback int64) int64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return fallback
	}
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return fallback
}
