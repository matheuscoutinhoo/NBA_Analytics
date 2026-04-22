package bankroll

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/matheuscoutinhoo/better/internal/auth"
	"github.com/matheuscoutinhoo/better/internal/models"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type depositRequest struct {
	Amount float64 `json:"amount"`
}

type withdrawRequest struct {
	Amount float64 `json:"amount"`
}

func (h *Handler) GetBankroll(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	bankroll, err := h.svc.repo.GetBankroll(claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch bankroll"})
		return
	}

	if bankroll == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"bankroll":     nil,
			"transactions": []models.BankrollTransaction{},
		})
		return
	}

	transactions, err := h.svc.repo.GetTransactions(claims.UserID)
	if err != nil {
		transactions = []models.BankrollTransaction{}
	}
	if transactions == nil {
		transactions = []models.BankrollTransaction{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"bankroll":     bankroll,
		"transactions": transactions,
	})
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req depositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "amount must be positive"})
		return
	}

	bankroll, err := h.svc.Deposit(claims.UserID, req.Amount)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to deposit"})
		return
	}

	writeJSON(w, http.StatusOK, bankroll)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "amount must be positive"})
		return
	}

	bankroll, err := h.svc.Withdraw(claims.UserID, req.Amount)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, bankroll)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	stats, err := h.svc.GetStats(claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get stats"})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

type BankrollStats struct {
	TotalBets     int     `json:"total_bets"`
	WonBets       int     `json:"won_bets"`
	LostBets      int     `json:"lost_bets"`
	PendingBets   int     `json:"pending_bets"`
	WinRate       float64 `json:"win_rate"`
	TotalStaked   float64 `json:"total_staked"`
	TotalReturns  float64 `json:"total_returns"`
	ProfitLoss    float64 `json:"profit_loss"`
	ROI           float64 `json:"roi"`
	InitialAmount float64 `json:"initial_amount"`
	CurrentAmount float64 `json:"current_amount"`
}

func (s *Service) GetStats(userID int64) (*BankrollStats, error) {
	bankroll, _ := s.repo.GetBankroll(userID)

	var totalBets, wonBets, lostBets, pendingBets int
	var totalStaked, totalReturns float64

	rows, err := s.repo.db.Query(
		"SELECT result, stake, potential_return FROM user_bets WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return &BankrollStats{}, nil
	}
	defer rows.Close()

	for rows.Next() {
		var result string
		var stake, potentialReturn float64
		rows.Scan(&result, &stake, &potentialReturn)
		totalBets++
		totalStaked += stake
		switch result {
		case "won":
			wonBets++
			totalReturns += potentialReturn
		case "lost":
			lostBets++
		case "pending":
			pendingBets++
		}
	}

	stats := &BankrollStats{
		TotalBets:    totalBets,
		WonBets:      wonBets,
		LostBets:     lostBets,
		PendingBets:  pendingBets,
		TotalStaked:  totalStaked,
		TotalReturns: totalReturns,
		ProfitLoss:   totalReturns - totalStaked,
	}

	settled := wonBets + lostBets
	if settled > 0 {
		stats.WinRate = math.Round(float64(wonBets)/float64(settled)*10000) / 100
	}
	if totalStaked > 0 {
		stats.ROI = math.Round((totalReturns-totalStaked)/totalStaked*10000) / 100
	}

	if bankroll != nil {
		stats.InitialAmount = bankroll.InitialAmount
		stats.CurrentAmount = bankroll.CurrentAmount
	}

	return stats, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
