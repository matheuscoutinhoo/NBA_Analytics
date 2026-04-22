package auth

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/matheuscoutinhoo/better/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(email, passwordHash, role string) (*models.User, error) {
	result, err := r.db.Exec(
		"INSERT INTO users (email, password_hash, role) VALUES (?, ?, ?)",
		email, passwordHash, role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetUserByID(id)
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, role, created_at, updated_at, deleted_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (r *Repository) GetUserByID(id int64) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, role, created_at, updated_at, deleted_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

func (r *Repository) UpdateUser(id int64, email string) error {
	_, err := r.db.Exec(
		"UPDATE users SET email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		email, id,
	)
	return err
}

func (r *Repository) UpdatePassword(id int64, passwordHash string) error {
	_, err := r.db.Exec(
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		passwordHash, id,
	)
	return err
}

func (r *Repository) SoftDeleteUser(id int64) error {
	_, err := r.db.Exec(
		"UPDATE users SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	return err
}

// Refresh token methods
func (r *Repository) StoreRefreshToken(userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)",
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *Repository) GetRefreshToken(tokenHash string) (*models.RefreshToken, error) {
	rt := &models.RefreshToken{}
	err := r.db.QueryRow(
		"SELECT id, user_id, token_hash, expires_at, revoked, created_at FROM refresh_tokens WHERE token_hash = ?",
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return rt, nil
}

func (r *Repository) RevokeRefreshToken(tokenHash string) error {
	_, err := r.db.Exec("UPDATE refresh_tokens SET revoked = 1 WHERE token_hash = ?", tokenHash)
	return err
}

func (r *Repository) RevokeAllUserTokens(userID int64) error {
	_, err := r.db.Exec("UPDATE refresh_tokens SET revoked = 1 WHERE user_id = ?", userID)
	return err
}

// Login attempts (brute force protection)
func (r *Repository) RecordLoginAttempt(email, ip string, success bool) error {
	successInt := 0
	if success {
		successInt = 1
	}
	_, err := r.db.Exec(
		"INSERT INTO login_attempts (email, ip_address, success) VALUES (?, ?, ?)",
		email, ip, successInt,
	)
	return err
}

func (r *Repository) GetRecentFailedAttempts(email string, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM login_attempts WHERE email = ? AND success = 0 AND attempted_at > ?",
		email, since,
	).Scan(&count)
	return count, err
}

// Audit logs
func (r *Repository) CreateAuditLog(log *models.AuditLog) error {
	_, err := r.db.Exec(
		"INSERT INTO audit_logs (user_id, action, entity, entity_id, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?)",
		log.UserID, log.Action, log.Entity, log.EntityID, log.IPAddress, log.UserAgent,
	)
	return err
}
