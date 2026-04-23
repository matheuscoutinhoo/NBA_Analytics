package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matheuscoutinhoo/better/internal/auth"
	"github.com/matheuscoutinhoo/better/internal/database"
	_ "modernc.org/sqlite"
)

func setupIntegrationDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	if err := database.InitWithDB(db); err != nil {
		t.Fatalf("Failed to init test database: %v", err)
	}
	return db
}

func TestRegisterEndpoint(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	repo := auth.NewRepository(db)
	handler := auth.NewHandler(repo, "test-jwt-secret-at-least-32-chars!", "test-refresh-secret-at-least-32chars!")

	body := `{"email":"test@example.com","password":"TestPass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "user created successfully" {
		t.Errorf("unexpected message: %v", resp["message"])
	}
}

func TestRegisterEndpoint_DuplicateEmail(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	repo := auth.NewRepository(db)
	handler := auth.NewHandler(repo, "test-jwt-secret-at-least-32-chars!", "test-refresh-secret-at-least-32chars!")

	body := `{"email":"test@example.com","password":"TestPass123"}`

	// First register
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	// Second register with same email
	req = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
}

func TestRegisterEndpoint_InvalidEmail(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	repo := auth.NewRepository(db)
	handler := auth.NewHandler(repo, "test-jwt-secret-at-least-32-chars!", "test-refresh-secret-at-least-32chars!")

	body := `{"email":"invalid-email","password":"TestPass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestRegisterEndpoint_WeakPassword(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	repo := auth.NewRepository(db)
	handler := auth.NewHandler(repo, "test-jwt-secret-at-least-32-chars!", "test-refresh-secret-at-least-32chars!")

	body := `{"email":"test@example.com","password":"weak"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLoginEndpoint(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	repo := auth.NewRepository(db)
	handler := auth.NewHandler(repo, "test-jwt-secret-at-least-32-chars!", "test-refresh-secret-at-least-32chars!")

	// Register
	regBody := `{"email":"login@example.com","password":"TestPass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	// Login
	loginBody := `{"email":"login@example.com","password":"TestPass123"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["access_token"] == nil || resp["access_token"] == "" {
		t.Error("should return access token")
	}

	// Check refresh token cookie
	cookies := w.Result().Cookies()
	var foundCookie bool
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			foundCookie = true
			if !c.HttpOnly {
				t.Error("refresh token cookie should be HttpOnly")
			}
			if !c.Secure {
				t.Error("refresh token cookie should be Secure")
			}
		}
	}
	if !foundCookie {
		t.Error("should set refresh_token cookie")
	}
}

func TestLoginEndpoint_WrongPassword(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	repo := auth.NewRepository(db)
	handler := auth.NewHandler(repo, "test-jwt-secret-at-least-32-chars!", "test-refresh-secret-at-least-32chars!")

	// Register
	regBody := `{"email":"login2@example.com","password":"TestPass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	// Login with wrong password
	loginBody := `{"email":"login2@example.com","password":"WrongPass123"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestLoginEndpoint_BruteForce(t *testing.T) {
	db := setupIntegrationDB(t)
	defer db.Close()

	repo := auth.NewRepository(db)
	handler := auth.NewHandler(repo, "test-jwt-secret-at-least-32-chars!", "test-refresh-secret-at-least-32chars!")

	// Register
	regBody := `{"email":"brute@example.com","password":"TestPass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	// 5 failed login attempts
	for i := 0; i < 5; i++ {
		loginBody := `{"email":"brute@example.com","password":"WrongPass123"}`
		req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		handler.Login(w, req)
	}

	// 6th attempt should be rate limited
	loginBody := `{"email":"brute@example.com","password":"TestPass123"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.Login(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}
