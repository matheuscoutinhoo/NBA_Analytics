package games

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matheuscoutinhoo/better/internal/auth"
	"github.com/matheuscoutinhoo/better/internal/models"
)

type Handler struct {
	repo       *Repository
	insightSvc *InsightService
}

func NewHandler(repo *Repository, insightSvc *InsightService) *Handler {
	return &Handler{repo: repo, insightSvc: insightSvc}
}

func (h *Handler) GetRecentGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.repo.GetRecentGames(30)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch games"})
		return
	}
	if games == nil {
		games = []models.NBAGame{}
	}
	writeJSON(w, http.StatusOK, games)
}

func (h *Handler) GetUpcomingGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.repo.GetUpcomingGames(7)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch games"})
		return
	}
	if games == nil {
		games = []models.NBAGame{}
	}
	writeJSON(w, http.StatusOK, games)
}

func (h *Handler) GetGameStats(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid game id"})
		return
	}

	game, err := h.repo.GetGameByID(gameID)
	if err != nil || game == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "game not found"})
		return
	}

	stats, err := h.repo.GetGameStats(gameID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch stats"})
		return
	}

	playerStats, err := h.repo.GetPlayerStats(gameID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch player stats"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"game":         game,
		"team_stats":   stats,
		"player_stats": playerStats,
	})
}

func (h *Handler) GetGameOdds(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid game id"})
		return
	}

	odds, err := h.insightSvc.oddsRepo.GetOddsByGameID(gameID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch odds"})
		return
	}
	if odds == nil {
		odds = []models.GameOdds{}
	}
	writeJSON(w, http.StatusOK, odds)
}

func (h *Handler) GenerateInsights(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid game id"})
		return
	}

	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	insight, err := h.insightSvc.GenerateInsight(gameID, claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate insight: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, insight)
}

func parseGameID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	return strconv.ParseInt(idStr, 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
