// auth.go contains application logic.
package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"

	"github.com/shrinit12-projects/reistta-common/utils"
)

var ErrUnauthorized = errors.New("unauthorized")

type ctxKey struct{}
type sessionIDKey struct{}

type Claims struct {
	Email      string `json:"email"`
	MerchantID string `json:"merchant_id"`
	jwt.RegisteredClaims
}

func Middleware(redisClient *redis.Client, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Authenticate request and inject session into context.
			session, sessionID, err := Authenticate(r.Context(), r.Header.Get("Authorization"), redisClient, jwtSecret, r)
			if err != nil {
				utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, session)
			ctx = context.WithValue(ctx, sessionIDKey{}, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		next.ServeHTTP(w, r)
	})
}

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(rps rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{limiter: rate.NewLimiter(rps, burst)}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.limiter.Allow() {
			utils.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SessionFromContext(ctx context.Context) (SecureSession, bool) {
	session, ok := ctx.Value(ctxKey{}).(SecureSession)
	return session, ok
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey{}).(string)
	return sessionID, ok
}

func Authenticate(ctx context.Context, authHeader string, redisClient *redis.Client, jwtSecret string, r *http.Request) (SecureSession, string, error) {
	token, ok := parseBearer(authHeader)
	if !ok {
		return SecureSession{}, "", ErrUnauthorized
	}

	// Verify JWT and extract session ID.
	claims, err := VerifyAccessToken(token, jwtSecret)
	if err != nil {
		return SecureSession{}, "", ErrUnauthorized
	}

	sessionID := claims.ID
	if sessionID == "" {
		return SecureSession{}, "", ErrUnauthorized
	}

	key := "session:" + sessionID
	// Load session payload from Redis.
	raw, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return SecureSession{}, "", ErrUnauthorized
	}

	var session SecureSession
	if err := utils.DecodeJSONBytes(raw, &session); err != nil {
		return SecureSession{}, "", ErrUnauthorized
	}

	// Optional IP/UA pinning if session stores them.
	if session.IPAddress != "" && session.IPAddress != clientIP(r) {
		return SecureSession{}, "", ErrUnauthorized
	}
	if session.UserAgent != "" && session.UserAgent != r.UserAgent() {
		return SecureSession{}, "", ErrUnauthorized
	}

	return session, sessionID, nil
}

func CreateAccessToken(claims Claims, secret string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(ttl))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyAccessToken(tokenStr, secret string) (Claims, error) {
	var claims Claims
	parsed, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrUnauthorized
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return Claims{}, ErrUnauthorized
	}
	return claims, nil
}

func parseBearer(header string) (string, bool) {
	if header == "" {
		return "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return "", false
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}

func clientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

