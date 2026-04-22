package bets

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matheuscoutinhoo/better/internal/auth"
	"github.com/matheuscoutinhoo/better/internal/models"
)

type Handler struct {
	repo        *Repository
	bankrollSvc BankrollService
}

type BankrollService interface {
	DeductBet(userID int64, betID int64, amount float64) error
	CreditWin(userID int64, betID int64, amount float64) error
	RefundBet(userID int64, betID int64, amount float64) error
}

func NewHandler(repo *Repository, bankrollSvc BankrollService) *Handler {
	return &Handler{repo: repo, bankrollSvc: bankrollSvc}
}

type createBetRequest struct {
	GameID    int64   `json:"game_id"`
	BetType   string  `json:"bet_type"`
	Selection string  `json:"selection"`
	OddValue  float64 `json:"odd_value"`
	Stake     float64 `json:"stake"`
}

type updateResultRequest struct {
	Result string `json:"result"`
}

func (h *Handler) GetBets(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	betsList, err := h.repo.GetUserBets(claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch bets"})
		return
	}
	if betsList == nil {
		betsList = []models.UserBet{}
	}
	writeJSON(w, http.StatusOK, betsList)
}

func (h *Handler) CreateBet(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req createBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.GameID <= 0 || req.BetType == "" || req.Selection == "" || req.OddValue <= 0 || req.Stake <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "all fields are required and must be positive"})
		return
	}

	potentialReturn := req.Stake * req.OddValue

	bet := &models.UserBet{
		UserID:          claims.UserID,
		GameID:          req.GameID,
		BetType:         req.BetType,
		Selection:       req.Selection,
		OddValue:        req.OddValue,
		Stake:           req.Stake,
		PotentialReturn: potentialReturn,
	}

	betID, err := h.repo.CreateBet(bet)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create bet"})
		return
	}

	// Deduct from bankroll
	if err := h.bankrollSvc.DeductBet(claims.UserID, betID, req.Stake); err != nil {
		// Don't fail the bet creation if bankroll deduction fails
		// (user might not have bankroll set up yet)
	}

	bet.ID = betID
	writeJSON(w, http.StatusCreated, bet)
}

func (h *Handler) UpdateResult(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	betID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bet id"})
		return
	}

	var req updateResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Result != "won" && req.Result != "lost" && req.Result != "void" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "result must be 'won', 'lost', or 'void'"})
		return
	}

	bet, err := h.repo.GetBetByID(betID, claims.UserID)
	if err != nil || bet == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "bet not found"})
		return
	}

	if err := h.repo.UpdateBetResult(betID, claims.UserID, req.Result); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update result"})
		return
	}

	// Update bankroll based on result
	switch req.Result {
	case "won":
		h.bankrollSvc.CreditWin(claims.UserID, betID, bet.PotentialReturn)
	case "void":
		h.bankrollSvc.RefundBet(claims.UserID, betID, bet.Stake)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "result updated"})
}

func (h *Handler) DeleteBet(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	betID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid bet id"})
		return
	}

	bet, err := h.repo.GetBetByID(betID, claims.UserID)
	if err != nil || bet == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "bet not found"})
		return
	}

	if err := h.repo.DeleteBet(betID, claims.UserID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete bet"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bet deleted"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
