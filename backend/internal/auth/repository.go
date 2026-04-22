package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, role, is_active, age_verified, consent_given, consent_given_at, xp_points, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.Role, user.IsActive, user.AgeVerified, user.ConsentGiven,
		user.ConsentGivenAt, user.XPPoints, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, role, is_active, age_verified, consent_given, consent_given_at, xp_points, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`

	user := &User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.AgeVerified, &user.ConsentGiven,
		&user.ConsentGivenAt, &user.XPPoints, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, role, is_active, age_verified, consent_given, consent_given_at, xp_points, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`

	user := &User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.AgeVerified, &user.ConsentGiven,
		&user.ConsentGivenAt, &user.XPPoints, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, role, is_active, age_verified, consent_given, consent_given_at, xp_points, created_at, updated_at
		FROM users WHERE username = $1 AND deleted_at IS NULL`

	user := &User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.AgeVerified, &user.ConsentGiven,
		&user.ConsentGivenAt, &user.XPPoints, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user by username: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users SET username = $2, email = $3, password_hash = $4, role = $5,
		is_active = $6, age_verified = $7, consent_given = $8, consent_given_at = $9,
		xp_points = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.Role, user.IsActive, user.AgeVerified, user.ConsentGiven,
		user.ConsentGivenAt, user.XPPoints, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft deleting user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("storing refresh token: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, expires_at, revoked FROM refresh_tokens WHERE token_hash = $1`
	token := &RefreshToken{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.Revoked,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting refresh token: %w", err)
	}
	return token, nil
}

func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked = true, revoked_at = NOW() WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	return nil
}

func (r *PostgresRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true, revoked_at = NOW() WHERE user_id = $1 AND revoked = false`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoking all user tokens: %w", err)
	}
	return nil
}
