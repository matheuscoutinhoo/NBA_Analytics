package betting

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/better/backend/internal/auth"
	"github.com/better/backend/pkg/logger"
	"github.com/better/backend/pkg/response"
	"github.com/better/backend/pkg/validator"
)

type BankrollEntry struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	GameID       *uuid.UUID `json:"game_id,omitempty"`
	EntryType    string     `json:"entry_type"`
	BetType      *string    `json:"bet_type,omitempty"`
	Odd          *float64   `json:"odd,omitempty"`
	Stake        float64    `json:"stake"`
	Result       *float64   `json:"result,omitempty"`
	BalanceAfter float64    `json:"balance_after"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreateEntryRequest struct {
	GameID    *uuid.UUID `json:"game_id"`
	EntryType string     `json:"entry_type"`
	BetType   *string    `json:"bet_type"`
	Odd       *float64   `json:"odd"`
	Stake     float64    `json:"stake"`
	Result    *float64   `json:"result"`
	Notes     *string    `json:"notes"`
}

type DashboardStats struct {
	CurrentBalance   float64            `json:"current_balance"`
	TotalDeposits    float64            `json:"total_deposits"`
	TotalWithdrawals float64            `json:"total_withdrawals"`
	TotalBets        int                `json:"total_bets"`
	WinCount         int                `json:"win_count"`
	LossCount        int                `json:"loss_count"`
	WinRate          float64            `json:"win_rate"`
	ROI              float64            `json:"roi"`
	MaxDrawdown      float64            `json:"max_drawdown"`
	ProfitLoss       float64            `json:"profit_loss"`
	BalanceHistory   []BalancePoint     `json:"balance_history"`
	MonthlyStats     []MonthlyStatEntry `json:"monthly_stats"`
}

type BalancePoint struct {
	Date    string  `json:"date"`
	Balance float64 `json:"balance"`
}

type MonthlyStatEntry struct {
	Month   string  `json:"month"`
	Profit  float64 `json:"profit"`
	Bets    int     `json:"bets"`
	WinRate float64 `json:"win_rate"`
}

type Service struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func NewService(db *pgxpool.Pool, log *logger.Logger) *Service {
	return &Service{db: db, logger: log}
}

func (s *Service) CreateEntry(ctx context.Context, userID uuid.UUID, req CreateEntryRequest) (*BankrollEntry, error) {
	v := validator.New()
	v.InList(req.EntryType, "entry_type", []string{"DEPOSIT", "WITHDRAWAL", "BET", "WIN", "LOSS"})
	v.PositiveFloat(req.Stake, "stake")
	if !v.Valid() {
		return nil, fmt.Errorf("validation failed: %v", v.Errors)
	}

	// Get current balance
	currentBalance := s.getCurrentBalance(ctx, userID)

	var balanceAfter float64
	switch req.EntryType {
	case "DEPOSIT":
		balanceAfter = currentBalance + req.Stake
	case "WITHDRAWAL":
		balanceAfter = currentBalance - req.Stake
	case "BET":
		balanceAfter = currentBalance - req.Stake
	case "WIN":
		if req.Result != nil {
			balanceAfter = currentBalance + *req.Result
		} else if req.Odd != nil {
			result := req.Stake * (*req.Odd - 1)
			req.Result = &result
			balanceAfter = currentBalance + result
		}
	case "LOSS":
		balanceAfter = currentBalance
		zero := 0.0
		req.Result = &zero
	}

	entry := &BankrollEntry{
		ID:           uuid.New(),
		UserID:       userID,
		GameID:       req.GameID,
		EntryType:    req.EntryType,
		BetType:      req.BetType,
		Odd:          req.Odd,
		Stake:        req.Stake,
		Result:       req.Result,
		BalanceAfter: balanceAfter,
		Notes:        req.Notes,
		CreatedAt:    time.Now().UTC(),
	}

	_, err := s.db.Exec(ctx,
		`INSERT INTO bankroll_entries (id, user_id, game_id, entry_type, bet_type, odd, stake, result, balance_after, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		entry.ID, entry.UserID, entry.GameID, entry.EntryType, entry.BetType,
		entry.Odd, entry.Stake, entry.Result, entry.BalanceAfter, entry.Notes, entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Service) GetEntries(ctx context.Context, userID uuid.UUID, page, perPage int) ([]BankrollEntry, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM bankroll_entries WHERE user_id = $1`, userID).Scan(&total)

	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, game_id, entry_type, bet_type, odd, stake, result, balance_after, notes, created_at
		FROM bankroll_entries WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []BankrollEntry
	for rows.Next() {
		var e BankrollEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.GameID, &e.EntryType, &e.BetType,
			&e.Odd, &e.Stake, &e.Result, &e.BalanceAfter, &e.Notes, &e.CreatedAt); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	return entries, total, nil
}

func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Current balance
	stats.CurrentBalance = s.getCurrentBalance(ctx, userID)

	// Totals
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(stake), 0) FROM bankroll_entries WHERE user_id = $1 AND entry_type = 'DEPOSIT'`,
		userID,
	).Scan(&stats.TotalDeposits)

	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(stake), 0) FROM bankroll_entries WHERE user_id = $1 AND entry_type = 'WITHDRAWAL'`,
		userID,
	).Scan(&stats.TotalWithdrawals)

	_ = s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bankroll_entries WHERE user_id = $1 AND entry_type IN ('BET', 'WIN', 'LOSS')`,
		userID,
	).Scan(&stats.TotalBets)

	_ = s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bankroll_entries WHERE user_id = $1 AND entry_type = 'WIN'`,
		userID,
	).Scan(&stats.WinCount)

	_ = s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bankroll_entries WHERE user_id = $1 AND entry_type = 'LOSS'`,
		userID,
	).Scan(&stats.LossCount)

	if stats.TotalBets > 0 {
		stats.WinRate = math.Round(float64(stats.WinCount)/float64(stats.WinCount+stats.LossCount)*10000) / 100
	}

	// ROI
	if stats.TotalDeposits > 0 {
		stats.ProfitLoss = stats.CurrentBalance - stats.TotalDeposits + stats.TotalWithdrawals
		stats.ROI = math.Round(stats.ProfitLoss/stats.TotalDeposits*10000) / 100
	}

	// Balance history
	rows, err := s.db.Query(ctx,
		`SELECT DATE(created_at) as date, balance_after
		FROM bankroll_entries WHERE user_id = $1
		ORDER BY created_at ASC`,
		userID,
	)
	if err == nil {
		defer rows.Close()
		maxBalance := 0.0
		for rows.Next() {
			var bp BalancePoint
			if err := rows.Scan(&bp.Date, &bp.Balance); err != nil {
				continue
			}
			if bp.Balance > maxBalance {
				maxBalance = bp.Balance
			}
			drawdown := 0.0
			if maxBalance > 0 {
				drawdown = (maxBalance - bp.Balance) / maxBalance * 100
			}
			if drawdown > stats.MaxDrawdown {
				stats.MaxDrawdown = math.Round(drawdown*100) / 100
			}
			stats.BalanceHistory = append(stats.BalanceHistory, bp)
		}
	}

	return stats, nil
}

func (s *Service) getCurrentBalance(ctx context.Context, userID uuid.UUID) float64 {
	var balance float64
	_ = s.db.QueryRow(ctx,
		`SELECT COALESCE(balance_after, 0) FROM bankroll_entries WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(&balance)
	return balance
}

// Handler

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	var req CreateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	entry, err := h.service.CreateEntry(r.Context(), claims.UserID, req)
	if err != nil {
		if err.Error()[:10] == "validation" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusCreated, entry)
}

func (h *Handler) GetEntries(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	entries, total, err := h.service.GetEntries(r.Context(), claims.UserID, page, perPage)
	if err != nil {
		response.InternalError(w)
		return
	}

	if perPage < 1 {
		perPage = 20
	}
	totalPages := (total + perPage - 1) / perPage

	response.JSONWithMeta(w, http.StatusOK, entries, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	stats, err := h.service.GetDashboard(r.Context(), claims.UserID)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, stats)
}

func (h *Handler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	vars := mux.Vars(r)
	entryID, err := uuid.Parse(vars["entryId"])
	if err != nil {
		response.BadRequest(w, "invalid entry ID")
		return
	}

	// Verify ownership
	var ownerID uuid.UUID
	err = h.service.db.QueryRow(r.Context(),
		`SELECT user_id FROM bankroll_entries WHERE id = $1`, entryID,
	).Scan(&ownerID)
	if err != nil {
		response.NotFound(w, "entry not found")
		return
	}
	if ownerID != claims.UserID {
		response.Forbidden(w, "you can only delete your own entries")
		return
	}

	_, err = h.service.db.Exec(r.Context(),
		`DELETE FROM bankroll_entries WHERE id = $1 AND user_id = $2`, entryID, claims.UserID,
	)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "entry deleted"})
}
