package user

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/better/backend/internal/auth"
	"github.com/better/backend/pkg/logger"
	"github.com/better/backend/pkg/response"
)

type Preferences struct {
	ID                   uuid.UUID `json:"id"`
	UserID               uuid.UUID `json:"user_id"`
	DailyLimit           *float64  `json:"daily_limit"`
	WeeklyLimit          *float64  `json:"weekly_limit"`
	MonthlyLimit         *float64  `json:"monthly_limit"`
	AlertThreshold       float64   `json:"alert_threshold"`
	DarkMode             bool      `json:"dark_mode"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
}

type UpdatePreferencesRequest struct {
	DailyLimit           *float64 `json:"daily_limit"`
	WeeklyLimit          *float64 `json:"weekly_limit"`
	MonthlyLimit         *float64 `json:"monthly_limit"`
	AlertThreshold       *float64 `json:"alert_threshold"`
	DarkMode             *bool    `json:"dark_mode"`
	NotificationsEnabled *bool    `json:"notifications_enabled"`
}

type ExportData struct {
	User            *auth.User    `json:"user"`
	Preferences     *Preferences  `json:"preferences"`
	BankrollEntries []interface{} `json:"bankroll_entries"`
	AuditLogs       []interface{} `json:"audit_logs"`
	ExportedAt      time.Time     `json:"exported_at"`
}

type Service struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func NewService(db *pgxpool.Pool, log *logger.Logger) *Service {
	return &Service{db: db, logger: log}
}

func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) (*Preferences, error) {
	query := `SELECT id, user_id, daily_limit, weekly_limit, monthly_limit, alert_threshold, dark_mode, notifications_enabled
		FROM user_preferences WHERE user_id = $1`

	prefs := &Preferences{}
	err := s.db.QueryRow(ctx, query, userID).Scan(
		&prefs.ID, &prefs.UserID, &prefs.DailyLimit, &prefs.WeeklyLimit,
		&prefs.MonthlyLimit, &prefs.AlertThreshold, &prefs.DarkMode, &prefs.NotificationsEnabled,
	)
	if err != nil {
		// Create default preferences
		prefs = &Preferences{
			ID:                   uuid.New(),
			UserID:               userID,
			AlertThreshold:       80.0,
			DarkMode:             true,
			NotificationsEnabled: true,
		}
		_, err := s.db.Exec(ctx,
			`INSERT INTO user_preferences (id, user_id, alert_threshold, dark_mode, notifications_enabled) VALUES ($1, $2, $3, $4, $5)`,
			prefs.ID, prefs.UserID, prefs.AlertThreshold, prefs.DarkMode, prefs.NotificationsEnabled,
		)
		if err != nil {
			return nil, err
		}
	}
	return prefs, nil
}

func (s *Service) UpdatePreferences(ctx context.Context, userID uuid.UUID, req UpdatePreferencesRequest) (*Preferences, error) {
	prefs, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.DailyLimit != nil {
		prefs.DailyLimit = req.DailyLimit
	}
	if req.WeeklyLimit != nil {
		prefs.WeeklyLimit = req.WeeklyLimit
	}
	if req.MonthlyLimit != nil {
		prefs.MonthlyLimit = req.MonthlyLimit
	}
	if req.AlertThreshold != nil {
		prefs.AlertThreshold = *req.AlertThreshold
	}
	if req.DarkMode != nil {
		prefs.DarkMode = *req.DarkMode
	}
	if req.NotificationsEnabled != nil {
		prefs.NotificationsEnabled = *req.NotificationsEnabled
	}

	_, err = s.db.Exec(ctx,
		`UPDATE user_preferences SET daily_limit=$2, weekly_limit=$3, monthly_limit=$4, alert_threshold=$5, dark_mode=$6, notifications_enabled=$7, updated_at=NOW() WHERE user_id=$1`,
		userID, prefs.DailyLimit, prefs.WeeklyLimit, prefs.MonthlyLimit, prefs.AlertThreshold, prefs.DarkMode, prefs.NotificationsEnabled,
	)
	if err != nil {
		return nil, err
	}

	return prefs, nil
}

// LGPD: Export all user data
func (s *Service) ExportUserData(ctx context.Context, userID uuid.UUID) (*ExportData, error) {
	// Get user
	var user auth.User
	err := s.db.QueryRow(ctx,
		`SELECT id, username, email, role, is_active, age_verified, consent_given, consent_given_at, xp_points, created_at, updated_at FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.AgeVerified, &user.ConsentGiven, &user.ConsentGivenAt, &user.XPPoints, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	prefs, _ := s.GetPreferences(ctx, userID)

	// Get bankroll entries
	rows, err := s.db.Query(ctx,
		`SELECT id, entry_type, bet_type, odd, stake, result, balance_after, notes, created_at FROM bankroll_entries WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []interface{}
	for rows.Next() {
		var entry map[string]interface{}
		values, _ := rows.Values()
		entry = make(map[string]interface{})
		cols := []string{"id", "entry_type", "bet_type", "odd", "stake", "result", "balance_after", "notes", "created_at"}
		for i, col := range cols {
			if i < len(values) {
				entry[col] = values[i]
			}
		}
		entries = append(entries, entry)
	}

	return &ExportData{
		User:            &user,
		Preferences:     prefs,
		BankrollEntries: entries,
		ExportedAt:      time.Now().UTC(),
	}, nil
}

// LGPD: Delete all user data (right to be forgotten)
func (s *Service) DeleteAllUserData(ctx context.Context, userID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete in order respecting foreign keys
	tx.Exec(ctx, `DELETE FROM user_badges WHERE user_id = $1`, userID)
	tx.Exec(ctx, `DELETE FROM bankroll_entries WHERE user_id = $1`, userID)
	tx.Exec(ctx, `DELETE FROM user_preferences WHERE user_id = $1`, userID)
	tx.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	tx.Exec(ctx, `UPDATE audit_logs SET user_id = NULL, details = '{"anonymized": true}' WHERE user_id = $1`, userID)
	tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.logger.Info("user data completely deleted (LGPD)", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	prefs, err := h.service.GetPreferences(r.Context(), claims.UserID)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, prefs)
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	var req UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	prefs, err := h.service.UpdatePreferences(r.Context(), claims.UserID, req)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, prefs)
}

func (h *Handler) ExportData(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	data, err := h.service.ExportUserData(r.Context(), claims.UserID)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, data)
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	if err := h.service.DeleteAllUserData(r.Context(), claims.UserID); err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "All your data has been permanently deleted",
	})
}
