package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/better/backend/internal/auth"
	"github.com/better/backend/pkg/logger"
	"github.com/better/backend/pkg/response"
)

type Middleware struct {
	authService *auth.Service
	logger      *logger.Logger
	ipLimiters  map[string]*rate.Limiter
	mu          sync.RWMutex
	ipRate      int
}

func New(authService *auth.Service, log *logger.Logger, ipRate int) *Middleware {
	return &Middleware{
		authService: authService,
		logger:      log,
		ipLimiters:  make(map[string]*rate.Limiter),
		ipRate:      ipRate,
	}
}

// Logging middleware
func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		m.logger.Info("request completed", map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrapped.statusCode,
			"duration":   time.Since(start).String(),
			"ip":         getClientIP(r),
			"user_agent": r.UserAgent(),
		})
	})
}

// Recovery middleware
func (m *Middleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("panic recovered", map[string]interface{}{
					"error": err,
					"stack": string(debug.Stack()),
					"path":  r.URL.Path,
				})
				response.InternalError(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Authentication middleware
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(w, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(w, "invalid authorization header format")
			return
		}

		claims, err := m.authService.ValidateAccessToken(parts[1])
		if err != nil {
			response.Unauthorized(w, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), auth.ClaimsContextKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware
func (m *Middleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(auth.ClaimsContextKey{}).(*auth.Claims)
			if !ok {
				response.Unauthorized(w, "not authenticated")
				return
			}

			if claims.Role != role {
				response.Forbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Rate limiting middleware
func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := m.getIPLimiter(ip)

		if !limiter.Allow() {
			response.TooManyRequests(w, "rate limit exceeded, please try again later")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Security headers middleware
func (m *Middleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) getIPLimiter(ip string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, exists := m.ipLimiters[ip]
	if !exists {
		// Allow ipRate requests per minute with burst of ipRate/10
		limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(m.ipRate)), m.ipRate/10)
		m.ipLimiters[ip] = limiter

		// Cleanup old limiters periodically (every 1000 entries)
		if len(m.ipLimiters) > 1000 {
			go m.cleanupLimiters()
		}
	}

	return limiter
}

func (m *Middleware) cleanupLimiters() {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Simple cleanup: reset all limiters
	m.ipLimiters = make(map[string]*rate.Limiter)
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (trusted proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-Ip
	realIP := r.Header.Get("X-Real-Ip")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return r.RemoteAddr
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
