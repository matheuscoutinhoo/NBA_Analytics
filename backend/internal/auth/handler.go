package auth

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/matheuscoutinhoo/better/internal/models"
)

type Handler struct {
	repo          *Repository
	jwtSecret     string
	refreshSecret string
}

func NewHandler(repo *Repository, jwtSecret, refreshSecret string) *Handler {
	return &Handler{repo: repo, jwtSecret: jwtSecret, refreshSecret: refreshSecret}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateProfileRequest struct {
	Email string `json:"email"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if !validateEmail(req.Email) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email format"})
		return
	}
	if !validatePassword(req.Password) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters with uppercase, lowercase, and digit"})
		return
	}

	existing, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if existing != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	user, err := h.repo.CreateUser(req.Email, hash, "user")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	h.repo.CreateAuditLog(&models.AuditLog{
		UserID:    &user.ID,
		Action:    "register",
		Entity:    "user",
		EntityID:  &user.ID,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "user created successfully",
		"user":    user,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	clientIP := getClientIP(r)

	// Check brute force
	failedAttempts, err := h.repo.GetRecentFailedAttempts(req.Email, time.Now().Add(-15*time.Minute))
	if err == nil && failedAttempts >= 5 {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "account temporarily locked, try again in 15 minutes"})
		return
	}

	user, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if user == nil || !CheckPassword(req.Password, user.PasswordHash) {
		h.repo.RecordLoginAttempt(req.Email, clientIP, false)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if user.DeletedAt != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "account has been deleted"})
		return
	}

	h.repo.RecordLoginAttempt(req.Email, clientIP, true)

	accessToken, err := GenerateAccessToken(user.ID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	refreshToken, refreshHash, err := GenerateRefreshToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.repo.StoreRefreshToken(user.ID, refreshHash, expiresAt); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	setRefreshTokenCookie(w, refreshToken, expiresAt)

	h.repo.CreateAuditLog(&models.AuditLog{
		UserID:    &user.ID,
		Action:    "login",
		Entity:    "user",
		EntityID:  &user.ID,
		IPAddress: clientIP,
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"user":         user,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		tokenHash := HashToken(cookie.Value)
		h.repo.RevokeRefreshToken(tokenHash)
	}

	clearRefreshTokenCookie(w)

	claims := GetClaimsFromContext(r.Context())
	if claims != nil {
		h.repo.CreateAuditLog(&models.AuditLog{
			UserID:    &claims.UserID,
			Action:    "logout",
			Entity:    "user",
			EntityID:  &claims.UserID,
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *Handler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "no refresh token"})
		return
	}

	tokenHash := HashToken(cookie.Value)
	storedToken, err := h.repo.GetRefreshToken(tokenHash)
	if err != nil || storedToken == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		return
	}

	if storedToken.Revoked || storedToken.ExpiresAt.Before(time.Now()) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "refresh token expired or revoked"})
		return
	}

	// Revoke old token
	h.repo.RevokeRefreshToken(tokenHash)

	user, err := h.repo.GetUserByID(storedToken.UserID)
	if err != nil || user == nil || user.DeletedAt != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}

	accessToken, err := GenerateAccessToken(user.ID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	newRefreshToken, newRefreshHash, err := GenerateRefreshToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.repo.StoreRefreshToken(user.ID, newRefreshHash, expiresAt); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	setRefreshTokenCookie(w, newRefreshToken, expiresAt)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
	})
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	claims := GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.repo.SoftDeleteUser(claims.UserID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	h.repo.RevokeAllUserTokens(claims.UserID)
	clearRefreshTokenCookie(w)

	h.repo.CreateAuditLog(&models.AuditLog{
		UserID:    &claims.UserID,
		Action:    "delete_account",
		Entity:    "user",
		EntityID:  &claims.UserID,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "account deleted successfully"})
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !validateEmail(req.Email) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email format"})
		return
	}

	existing, _ := h.repo.GetUserByEmail(req.Email)
	if existing != nil && existing.ID != claims.UserID {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already in use"})
		return
	}

	if err := h.repo.UpdateUser(claims.UserID, req.Email); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	user, _ := h.repo.GetUserByID(claims.UserID)
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if !validatePassword(req.NewPassword) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "new password must be at least 8 characters with uppercase, lowercase, and digit"})
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if !CheckPassword(req.CurrentPassword, user.PasswordHash) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "current password is incorrect"})
		return
	}

	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if err := h.repo.UpdatePassword(claims.UserID, hash); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	// Revoke all refresh tokens on password change
	h.repo.RevokeAllUserTokens(claims.UserID)
	clearRefreshTokenCookie(w)

	writeJSON(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}

// Helper functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

func setRefreshTokenCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
