package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBPath           string
	JWTSecret        string
	JWTRefreshSecret string
	AbacusAPIKey     string
	AbacusAPIURL     string
	ScraperUserAgent string
	RateLimitRPM     int
	AppEnv           string
	AppPort          string
	AllowedOrigins   string
	OddsAPIKey       string
	OddsAPIURL       string
}

func Load() *Config {
	return &Config{
		DBPath:           getEnv("DB_PATH", "./data/nba_bet_insights.db"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		AbacusAPIKey:     getEnv("ABACUS_API_KEY", ""),
		AbacusAPIURL:     getEnv("ABACUS_API_URL", "https://apps.abacus.ai/api/v0"),
		ScraperUserAgent: getEnv("SCRAPER_USER_AGENT", "Mozilla/5.0"),
		RateLimitRPM:     getEnvInt("RATE_LIMIT_RPM", 100),
		AppEnv:           getEnv("APP_ENV", "development"),
		AppPort:          getEnv("APP_PORT", "8080"),
		AllowedOrigins:   getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		OddsAPIKey:       getEnv("ODDS_API_KEY", ""),
		OddsAPIURL:       getEnv("ODDS_API_URL", "https://api.the-odds-api.com/v4"),
	}
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return fallback
}
