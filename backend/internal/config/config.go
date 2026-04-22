package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	API      APIConfig
	Scraper  ScraperConfig
}

type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	Issuer        string
}

type APIConfig struct {
	OddsAPIKey    string
	AbacusAPIKey  string
	AbacusAPIURL  string
	NBAStatsURL   string
	RateLimitIP   int
	RateLimitUser int
}

type ScraperConfig struct {
	DailySchedule  string
	ResultSchedule string
}

func Load() (*Config, error) {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	accessTTL, _ := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	refreshTTL, _ := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "better"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "better"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessTTL:     accessTTL,
			RefreshTTL:    refreshTTL,
			Issuer:        "better-platform",
		},
		API: APIConfig{
			OddsAPIKey:    getEnv("ODDS_API_KEY", ""),
			AbacusAPIKey:  getEnv("ABACUS_API_KEY", ""),
			AbacusAPIURL:  getEnv("ABACUS_API_URL", ""),
			NBAStatsURL:   getEnv("NBA_STATS_URL", "https://stats.nba.com/stats"),
			RateLimitIP:   getEnvInt("RATE_LIMIT_IP", 100),
			RateLimitUser: getEnvInt("RATE_LIMIT_USER", 1000),
		},
		Scraper: ScraperConfig{
			DailySchedule:  getEnv("SCRAPER_DAILY_SCHEDULE", "0 6 * * *"),
			ResultSchedule: getEnv("SCRAPER_RESULT_SCHEDULE", "0 8 * * *"),
		},
	}

	if cfg.JWT.AccessSecret == "" || cfg.JWT.RefreshSecret == "" {
		return nil, fmt.Errorf("JWT secrets must be configured")
	}

	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("database password must be configured")
	}

	return cfg, nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
