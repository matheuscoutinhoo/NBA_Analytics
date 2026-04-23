package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/matheuscoutinhoo/better/internal/abacus"
	"github.com/matheuscoutinhoo/better/internal/auth"
	"github.com/matheuscoutinhoo/better/internal/bankroll"
	"github.com/matheuscoutinhoo/better/internal/bets"
	"github.com/matheuscoutinhoo/better/internal/config"
	"github.com/matheuscoutinhoo/better/internal/database"
	"github.com/matheuscoutinhoo/better/internal/games"
	"github.com/matheuscoutinhoo/better/internal/middleware"
	"github.com/matheuscoutinhoo/better/internal/odds"
	"github.com/matheuscoutinhoo/better/internal/predictions"
	"github.com/matheuscoutinhoo/better/internal/scraper"
)

func main() {
	_ = godotenv.Load() // Load .env if present

	cfg := config.Load()

	if cfg.JWTSecret == "" || cfg.JWTRefreshSecret == "" {
		log.Println("WARNING: JWT secrets not configured. Using default dev values.")
		if cfg.JWTSecret == "" {
			cfg.JWTSecret = "dev-jwt-secret-change-in-production-32chars!"
		}
		if cfg.JWTRefreshSecret == "" {
			cfg.JWTRefreshSecret = "dev-refresh-secret-change-in-production-32!"
		}
	}

	// Database
	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Repositories
	authRepo := auth.NewRepository(database.DB)
	gamesRepo := games.NewRepository(database.DB)
	oddsRepo := odds.NewRepository(database.DB)
	betsRepo := bets.NewRepository(database.DB)
	bankrollRepo := bankroll.NewRepository(database.DB)

	// Services
	aiClient := abacus.NewClient(cfg.AbacusAPIKey, cfg.AbacusAPIURL)
	insightSvc := games.NewInsightService(gamesRepo, oddsRepo, aiClient, database.DB)
	bankrollSvc := bankroll.NewService(bankrollRepo)
	predictionSvc := predictions.NewService(gamesRepo, oddsRepo, aiClient, database.DB)

	// Handlers
	authHandler := auth.NewHandler(authRepo, cfg.JWTSecret, cfg.JWTRefreshSecret)
	gamesHandler := games.NewHandler(gamesRepo, insightSvc)
	betsHandler := bets.NewHandler(betsRepo, bankrollSvc)
	bankrollHandler := bankroll.NewHandler(bankrollSvc)
	predictionsHandler := predictions.NewHandler(predictionSvc)

	// Middleware
	secMiddleware := middleware.NewSecurityMiddleware(cfg.AllowedOrigins, cfg.IsProduction())
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPM)
	authMiddleware := auth.AuthMiddleware(cfg.JWTSecret)

	// Router
	mux := http.NewServeMux()

	// Auth routes (public)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.RefreshTokenHandler)

	// Auth routes (protected)
	mux.Handle("POST /api/v1/auth/logout", authMiddleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("DELETE /api/v1/auth/account", authMiddleware(http.HandlerFunc(authHandler.DeleteAccount)))

	// Games routes (protected)
	mux.Handle("GET /api/v1/games/recent", authMiddleware(http.HandlerFunc(gamesHandler.GetRecentGames)))
	mux.Handle("GET /api/v1/games/upcoming", authMiddleware(http.HandlerFunc(gamesHandler.GetUpcomingGames)))
	mux.Handle("GET /api/v1/games/{id}/stats", authMiddleware(http.HandlerFunc(gamesHandler.GetGameStats)))
	mux.Handle("GET /api/v1/games/{id}/odds", authMiddleware(http.HandlerFunc(gamesHandler.GetGameOdds)))
	mux.Handle("POST /api/v1/games/{id}/insights", authMiddleware(http.HandlerFunc(gamesHandler.GenerateInsights)))

	// Bets routes (protected)
	mux.Handle("GET /api/v1/bets", authMiddleware(http.HandlerFunc(betsHandler.GetBets)))
	mux.Handle("POST /api/v1/bets", authMiddleware(http.HandlerFunc(betsHandler.CreateBet)))
	mux.Handle("PATCH /api/v1/bets/{id}/result", authMiddleware(http.HandlerFunc(betsHandler.UpdateResult)))
	mux.Handle("DELETE /api/v1/bets/{id}", authMiddleware(http.HandlerFunc(betsHandler.DeleteBet)))

	// Bankroll routes (protected)
	mux.Handle("GET /api/v1/bankroll", authMiddleware(http.HandlerFunc(bankrollHandler.GetBankroll)))
	mux.Handle("POST /api/v1/bankroll/deposit", authMiddleware(http.HandlerFunc(bankrollHandler.Deposit)))
	mux.Handle("POST /api/v1/bankroll/withdraw", authMiddleware(http.HandlerFunc(bankrollHandler.Withdraw)))
	mux.Handle("GET /api/v1/bankroll/stats", authMiddleware(http.HandlerFunc(bankrollHandler.GetStats)))

	// Account routes (protected)
	mux.Handle("GET /api/v1/account/profile", authMiddleware(http.HandlerFunc(authHandler.GetProfile)))
	mux.Handle("PATCH /api/v1/account/profile", authMiddleware(http.HandlerFunc(authHandler.UpdateProfile)))
	mux.Handle("PATCH /api/v1/account/password", authMiddleware(http.HandlerFunc(authHandler.ChangePassword)))

	// Predictions routes (protected)
	mux.Handle("GET /api/v1/predictions", authMiddleware(http.HandlerFunc(predictionsHandler.GetPredictions)))

	// Health check
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply middleware chain
	handler := secMiddleware.SecurityHeaders(
		secMiddleware.CORS(
			rateLimiter.Limit(
				middleware.MaxBodySize(1 << 20)(mux),
			),
		),
	)

	// Start scraper cron job
	scraperSvc := scraper.NewScraper(gamesRepo, oddsRepo, cfg.ScraperUserAgent, cfg.OddsAPIKey, cfg.OddsAPIURL)
	scraperSvc.StartCronJob(6)

	// Start AI predictions cron job (every 1 hour)
	predictionSvc.StartCronJob(1)

	port := cfg.AppPort
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s (env: %s)", port, cfg.AppEnv)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}
