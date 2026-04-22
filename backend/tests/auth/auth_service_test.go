package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/better/backend/internal/auth"
	"github.com/better/backend/internal/config"
	"github.com/better/backend/pkg/logger"
)

// MockRepository implements auth.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, tokenHash, expiresAt)
	return args.Error(0)
}

func (m *MockRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*auth.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.RefreshToken), args.Error(1)
}

func (m *MockRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *MockRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func newTestService(repo *MockRepository) *auth.Service {
	log := logger.New(nil, logger.LevelError)
	cfg := config.JWTConfig{
		AccessSecret:  "test_access_secret_32_chars_long__",
		RefreshSecret: "test_refresh_secret_32_chars_long_",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    168 * time.Hour,
		Issuer:        "test",
	}
	return auth.NewService(repo, nil, cfg, log)
}

func TestRegister_Success(t *testing.T) {
	repo := new(MockRepository)
	service := newTestService(repo)

	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	repo.On("GetUserByUsername", mock.Anything, "testuser").Return(nil, nil)
	repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	repo.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	req := auth.RegisterRequest{
		Username:    "testuser",
		Email:       "test@example.com",
		Password:    "securepassword",
		AgeVerified: true,
		Consent:     true,
	}

	user, tokens, err := service.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, tokens)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "USER", user.Role)
	assert.True(t, user.AgeVerified)
	assert.True(t, user.ConsentGiven)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	repo.AssertExpectations(t)
}

func TestRegister_UserExists(t *testing.T) {
	repo := new(MockRepository)
	service := newTestService(repo)

	existingUser := &auth.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	req := auth.RegisterRequest{
		Username:    "newuser",
		Email:       "test@example.com",
		Password:    "securepassword",
		AgeVerified: true,
		Consent:     true,
	}

	user, tokens, err := service.Register(context.Background(), req)

	assert.ErrorIs(t, err, auth.ErrUserExists)
	assert.Nil(t, user)
	assert.Nil(t, tokens)
}

func TestRegister_AgeNotVerified(t *testing.T) {
	repo := new(MockRepository)
	service := newTestService(repo)

	req := auth.RegisterRequest{
		Username:    "testuser",
		Email:       "test@example.com",
		Password:    "securepassword",
		AgeVerified: false,
		Consent:     true,
	}

	user, tokens, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Nil(t, tokens)
}

func TestRegister_InvalidEmail(t *testing.T) {
	repo := new(MockRepository)
	service := newTestService(repo)

	req := auth.RegisterRequest{
		Username:    "testuser",
		Email:       "invalid-email",
		Password:    "securepassword",
		AgeVerified: true,
		Consent:     true,
	}

	user, tokens, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Nil(t, tokens)
}

func TestRegister_WeakPassword(t *testing.T) {
	repo := new(MockRepository)
	service := newTestService(repo)

	req := auth.RegisterRequest{
		Username:    "testuser",
		Email:       "test@example.com",
		Password:    "short",
		AgeVerified: true,
		Consent:     true,
	}

	user, tokens, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Nil(t, tokens)
}

func TestValidateAccessToken_Valid(t *testing.T) {
	repo := new(MockRepository)
	service := newTestService(repo)

	repo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	repo.On("GetUserByUsername", mock.Anything, "testuser").Return(nil, nil)
	repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(nil)
	repo.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	req := auth.RegisterRequest{
		Username:    "testuser",
		Email:       "test@example.com",
		Password:    "securepassword",
		AgeVerified: true,
		Consent:     true,
	}

	_, tokens, _ := service.Register(context.Background(), req)

	claims, err := service.ValidateAccessToken(tokens.AccessToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "USER", claims.Role)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	repo := new(MockRepository)
	service := newTestService(repo)

	claims, err := service.ValidateAccessToken("invalid.token.here")

	assert.Error(t, err)
	assert.Nil(t, claims)
}
