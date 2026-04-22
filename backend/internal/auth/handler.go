package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/better/backend/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Sanitize inputs
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	user, tokens, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch err {
		case ErrUserExists:
			response.Error(w, http.StatusConflict, "USER_EXISTS", "A user with this email or username already exists")
		case ErrAgeNotVerified:
			response.BadRequest(w, err.Error())
		case ErrConsentRequired:
			response.BadRequest(w, err.Error())
		default:
			if strings.Contains(err.Error(), "validation failed") {
				response.BadRequest(w, err.Error())
				return
			}
			response.InternalError(w)
		}
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, tokens, err := h.service.Login(r.Context(), req)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			response.Unauthorized(w, "invalid email or password")
		case ErrAccountInactive:
			response.Forbidden(w, "account is inactive")
		default:
			if strings.Contains(err.Error(), "validation failed") {
				response.BadRequest(w, err.Error())
				return
			}
			response.InternalError(w)
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	tokens, err := h.service.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case ErrTokenRevoked, ErrTokenExpired, ErrInvalidCredentials:
			response.Unauthorized(w, "invalid or expired refresh token")
		default:
			response.InternalError(w)
		}
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(ClaimsContextKey{}).(*Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if err := h.service.Logout(r.Context(), claims.UserID, token); err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(ClaimsContextKey{}).(*Claims)
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		response.NotFound(w, "user not found")
		return
	}

	response.JSON(w, http.StatusOK, user)
}

type ClaimsContextKey struct{}
