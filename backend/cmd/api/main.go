package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/better/backend/internal/analysis"
	"github.com/better/backend/internal/auth"
	"github.com/better/backend/internal/betting"
	"github.com/better/backend/internal/config"
	"github.com/better/backend/internal/database"
	"github.com/better/backend/internal/middleware"
	"github.com/better/backend/internal/scraper"
	"github.com/better/backend/internal/user"
	"github.com/better/backend/pkg/logger"
)

func main() {
	// Load .env in development
	godotenv.Load()

	log := logger.Default()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", map[string]interface{}{"error": err.Error()})
	}

	// Database connection
	db, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		log.Fatal("failed to connect to database", map[string]interface{}{"error": err.Error()})
	}
	defer db.Close()
	log.Info("connected to PostgreSQL")

	// Redis connection
	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal("failed to connect to redis", map[string]interface{}{"error": err.Error()})
	}
	defer redisClient.Close()
	log.Info("connected to Redis")

	// Initialize repositories and services
	authRepo := auth.NewPostgresRepository(db)
	authService := auth.NewService(authRepo, redisClient, cfg.JWT, log)
	authHandler := auth.NewHandler(authService)

	userService := user.NewService(db, log)
	userHandler := user.NewHandler(userService)

	scraperService := scraper.NewService(db, log, cfg.API.NBAStatsURL, cfg.API.OddsAPIKey)
	scraperHandler := scraper.NewHandler(scraperService)

	analysisService := analysis.NewService(db, redisClient, log, cfg.API.AbacusAPIKey, cfg.API.AbacusAPIURL)
	analysisHandler := analysis.NewHandler(analysisService)

	bettingService := betting.NewService(db, log)
	bettingHandler := betting.NewHandler(bettingService)

	// Middleware
	mw := middleware.New(authService, log, cfg.API.RateLimitIP)

	// Router
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(mw.SecurityHeaders)
	r.Use(mw.Recovery)
	r.Use(mw.Logging)
	r.Use(mw.RateLimit)

	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	}).Methods("GET")

	// Auth routes (public)
	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	authRouter.HandleFunc("/refresh", authHandler.Refresh).Methods("POST")

	// Protected auth routes
	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(mw.Authenticate)
	authProtected.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	authProtected.HandleFunc("/me", authHandler.Me).Methods("GET")

	// User routes (protected)
	userRouter := api.PathPrefix("/user").Subrouter()
	userRouter.Use(mw.Authenticate)
	userRouter.HandleFunc("/preferences", userHandler.GetPreferences).Methods("GET")
	userRouter.HandleFunc("/preferences", userHandler.UpdatePreferences).Methods("PUT")
	userRouter.HandleFunc("/export", userHandler.ExportData).Methods("GET")
	userRouter.HandleFunc("/delete", userHandler.DeleteAccount).Methods("DELETE")

	// Analysis routes (protected)
	analysisRouter := api.PathPrefix("/analysis").Subrouter()
	analysisRouter.Use(mw.Authenticate)
	analysisRouter.HandleFunc("/games/today", analysisHandler.GetTodaysGames).Methods("GET")
	analysisRouter.HandleFunc("/games/{gameId}", analysisHandler.GetGameAnalysis).Methods("GET")
	analysisRouter.HandleFunc("/teams", analysisHandler.GetTeams).Methods("GET")

	// Bankroll routes (protected)
	bankrollRouter := api.PathPrefix("/bankroll").Subrouter()
	bankrollRouter.Use(mw.Authenticate)
	bankrollRouter.HandleFunc("/entries", bettingHandler.CreateEntry).Methods("POST")
	bankrollRouter.HandleFunc("/entries", bettingHandler.GetEntries).Methods("GET")
	bankrollRouter.HandleFunc("/entries/{entryId}", bettingHandler.DeleteEntry).Methods("DELETE")
	bankrollRouter.HandleFunc("/dashboard", bettingHandler.GetDashboard).Methods("GET")

	// Admin routes
	adminRouter := api.PathPrefix("/admin").Subrouter()
	adminRouter.Use(mw.Authenticate)
	adminRouter.Use(mw.RequireRole("ADMIN"))
	adminRouter.HandleFunc("/scraper/daily", scraperHandler.TriggerDailyScrape).Methods("POST")
	adminRouter.HandleFunc("/scraper/results", scraperHandler.TriggerResultsUpdate).Methods("POST")
	adminRouter.HandleFunc("/scraper/jobs", scraperHandler.GetScrapingJobs).Methods("GET")

	// CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://better.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      corsHandler.Handler(r),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start cron jobs for scraping
	go startCronJobs(scraperService, log)

	// Graceful shutdown
	go func() {
		log.Info("server starting", map[string]interface{}{
			"port":        cfg.Server.Port,
			"environment": cfg.Server.Environment,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", map[string]interface{}{"error": err.Error()})
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", map[string]interface{}{"error": err.Error()})
	}

	log.Info("server stopped gracefully")
}

func startCronJobs(scraperSvc *scraper.Service, log *logger.Logger) {
	// Simple ticker-based scheduler
	dailyTicker := time.NewTicker(1 * time.Hour)
	defer dailyTicker.Stop()

	for range dailyTicker.C {
		now := time.Now().UTC()

		// Run daily scrape at 6 AM UTC
		if now.Hour() == 6 {
			log.Info("running daily scrape cron job")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			if err := scraperSvc.RunDailyScrape(ctx); err != nil {
				log.Error("daily scrape cron failed", map[string]interface{}{"error": err.Error()})
			}
			cancel()
		}

		// Run results update at 8 AM UTC
		if now.Hour() == 8 {
			log.Info("running results update cron job")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			if err := scraperSvc.RunResultsUpdate(ctx); err != nil {
				log.Error("results update cron failed", map[string]interface{}{"error": err.Error()})
			}
			cancel()
		}
	}
}
