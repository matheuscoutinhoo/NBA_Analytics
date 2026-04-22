package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/better/backend/internal/config"
	"github.com/better/backend/pkg/logger"
	"github.com/better/backend/pkg/validator"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrTokenExpired       = errors.New("token has expired")
	ErrAgeNotVerified     = errors.New("age verification required (18+)")
	ErrConsentRequired    = errors.New("consent is required")
	ErrAccountInactive    = errors.New("account is inactive")
)

type Service struct {
	repo   Repository
	redis  *redis.Client
	config config.JWTConfig
	logger *logger.Logger
}

func NewService(repo Repository, redisClient *redis.Client, cfg config.JWTConfig, log *logger.Logger) *Service {
	return &Service{
		repo:   repo,
		redis:  redisClient,
		config: cfg,
		logger: log,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, *TokenPair, error) {
	v := validator.New()
	v.ValidUsername(req.Username, "username")
	v.ValidEmail(req.Email, "email")
	v.ValidPassword(req.Password, "password")
	v.Check(req.AgeVerified, "age_verified", "you must verify you are 18 or older")
	v.Check(req.Consent, "consent", "you must consent to data processing")

	if !v.Valid() {
		return nil, nil, fmt.Errorf("validation failed: %v", v.Errors)
	}

	existing, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, nil, ErrUserExists
	}

	existingUsername, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, nil, fmt.Errorf("checking existing username: %w", err)
	}
	if existingUsername != nil {
		return nil, nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()
	user := &User{
		ID:             uuid.New(),
		Username:       req.Username,
		Email:          req.Email,
		PasswordHash:   string(hashedPassword),
		Role:           "USER",
		IsActive:       true,
		AgeVerified:    req.AgeVerified,
		ConsentGiven:   req.Consent,
		ConsentGivenAt: &now,
		XPPoints:       0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("creating user: %w", err)
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("generating tokens: %w", err)
	}

	s.logger.Info("user registered", map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
	})

	return user, tokens, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*User, *TokenPair, error) {
	v := validator.New()
	v.ValidEmail(req.Email, "email")
	v.RequiredString(req.Password, "password")

	if !v.Valid() {
		return nil, nil, fmt.Errorf("validation failed: %v", v.Errors)
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("getting user: %w", err)
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, nil, ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("generating tokens: %w", err)
	}

	s.logger.Info("user logged in", map[string]interface{}{
		"user_id": user.ID,
	})

	return user, tokens, nil
}

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	storedToken, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("getting refresh token: %w", err)
	}
	if storedToken == nil {
		return nil, ErrInvalidCredentials
	}

	if storedToken.Revoked {
		// Potential token reuse attack - revoke all user tokens
		s.repo.RevokeAllUserTokens(ctx, storedToken.UserID)
		s.logger.Warn("possible token reuse attack detected", map[string]interface{}{
			"user_id": storedToken.UserID,
		})
		return nil, ErrTokenRevoked
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Revoke the old refresh token (rotation)
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("revoking old token: %w", err)
	}

	// Also blacklist in Redis for faster checking
	s.redis.Set(ctx, "blacklist:"+tokenHash, "1", s.config.RefreshTTL)

	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("generating new tokens: %w", err)
	}

	return tokens, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID, accessToken string) error {
	// Blacklist the access token in Redis
	s.redis.Set(ctx, "blacklist:"+hashToken(accessToken), "1", s.config.AccessTTL)

	// Revoke all refresh tokens for this user
	if err := s.repo.RevokeAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("revoking user tokens: %w", err)
	}

	s.logger.Info("user logged out", map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.AccessSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidCredentials
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}

	// Check if token is blacklisted
	if s.redis != nil {
		blacklisted, _ := s.redis.Get(context.Background(), "blacklist:"+hashToken(tokenString)).Result()
		if blacklisted != "" {
			return nil, ErrTokenRevoked
		}
	}

	return &Claims{
		UserID: userID,
		Role:   claims["role"].(string),
		Email:  claims["email"].(string),
	}, nil
}

func (s *Service) generateTokenPair(ctx context.Context, user *User) (*TokenPair, error) {
	now := time.Now().UTC()
	accessExpires := now.Add(s.config.AccessTTL)

	accessClaims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"role":  user.Role,
		"email": user.Email,
		"iss":   s.config.Issuer,
		"iat":   now.Unix(),
		"exp":   accessExpires.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.AccessSecret))
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshTokenID := uuid.New().String()
	refreshExpires := now.Add(s.config.RefreshTTL)

	refreshClaims := jwt.MapClaims{
		"sub": user.ID.String(),
		"jti": refreshTokenID,
		"iss": s.config.Issuer,
		"iat": now.Unix(),
		"exp": refreshExpires.Unix(),
	}

	refreshTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshTokenJWT.SignedString([]byte(s.config.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	// Store refresh token hash in DB
	tokenHash := hashToken(refreshTokenString)
	if err := s.repo.StoreRefreshToken(ctx, user.ID, tokenHash, refreshExpires); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpires.Unix(),
	}, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
