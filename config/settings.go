// settings.go contains application logic.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Settings struct {
	DatabaseURL        string
	RedisURL           string
	PGMaxConns         int32
	PGMinConns         int32
	PGMaxConnIdleTime  time.Duration
	PGHealthcheckPeriod time.Duration
	RedisPoolSize      int
	RedisMinIdleConns  int
	SessionSecret      string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	LoginMaxAttempts   int
	LoginLockoutTTL    time.Duration
	RateLimitRPS       float64
	RateLimitBurst     int
	MinioEndpoint      string
	MinioAccessKey     string
	MinioSecretKey     string
	MinioBucket        string
	MinioUseSSL        bool
}

func LoadSettings() (Settings, error) {
	_ = godotenv.Load()
	var missing []string

	dbURL, ok := lookupRequired("DATABASE_URL", &missing)
	redisURL, ok2 := lookupRequired("REDIS_URL", &missing)
	_ = ok
	_ = ok2

	pgMaxConns := mustInt32("PG_MAX_CONNS", &missing)
	pgMinConns := mustInt32("PG_MIN_CONNS", &missing)
	pgMaxIdle := mustDuration("PG_MAX_CONN_IDLE_TIME", &missing)
	pgHealth := mustDuration("PG_HEALTHCHECK_PERIOD", &missing)

	redisPool := mustInt("REDIS_POOL_SIZE", &missing)
	redisMinIdle := mustInt("REDIS_MIN_IDLE_CONNS", &missing)
	sessionSecret := mustString("SESSION_SECRET", &missing)
	accessTTL := mustDuration("ACCESS_TOKEN_TTL", &missing)
	refreshTTL := mustDuration("REFRESH_TOKEN_TTL", &missing)
	loginMax := mustInt("LOGIN_MAX_ATTEMPTS", &missing)
	loginLock := mustDuration("LOGIN_LOCKOUT_TTL", &missing)
	rateRPS := mustFloat64("RATE_LIMIT_RPS", &missing)
	rateBurst := mustInt("RATE_LIMIT_BURST", &missing)
	minioEndpoint := mustString("MINIO_ENDPOINT", &missing)
	minioAccessKey := mustString("MINIO_ACCESS_KEY", &missing)
	minioSecretKey := mustString("MINIO_SECRET_KEY", &missing)
	minioBucket := mustString("MINIO_BUCKET", &missing)
	minioUseSSL := mustBool("MINIO_USE_SSL", &missing)

	if len(missing) > 0 {
		return Settings{}, fmt.Errorf("missing or invalid env vars: %s", strings.Join(missing, ", "))
	}

	return Settings{
		DatabaseURL:        dbURL,
		RedisURL:           redisURL,
		PGMaxConns:         pgMaxConns,
		PGMinConns:         pgMinConns,
		PGMaxConnIdleTime:  pgMaxIdle,
		PGHealthcheckPeriod: pgHealth,
		RedisPoolSize:      redisPool,
		RedisMinIdleConns:  redisMinIdle,
		SessionSecret:      sessionSecret,
		AccessTokenTTL:     accessTTL,
		RefreshTokenTTL:    refreshTTL,
		LoginMaxAttempts:   loginMax,
		LoginLockoutTTL:    loginLock,
		RateLimitRPS:       rateRPS,
		RateLimitBurst:     rateBurst,
		MinioEndpoint:      minioEndpoint,
		MinioAccessKey:     minioAccessKey,
		MinioSecretKey:     minioSecretKey,
		MinioBucket:        minioBucket,
		MinioUseSSL:        minioUseSSL,
	}, nil
}

func lookupRequired(key string, missing *[]string) (string, bool) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		*missing = append(*missing, key)
		return "", false
	}
	return val, true
}

func mustString(key string, missing *[]string) string {
	val, ok := lookupRequired(key, missing)
	if !ok {
		return ""
	}
	return val
}

func mustInt(key string, missing *[]string) int {
	val, ok := lookupRequired(key, missing)
	if !ok {
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		*missing = append(*missing, key)
		return 0
	}
	return n
}

func mustInt32(key string, missing *[]string) int32 {
	val, ok := lookupRequired(key, missing)
	if !ok {
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		*missing = append(*missing, key)
		return 0
	}
	return int32(n)
}

func mustDuration(key string, missing *[]string) time.Duration {
	val, ok := lookupRequired(key, missing)
	if !ok {
		return 0
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		*missing = append(*missing, key)
		return 0
	}
	return d
}

func mustFloat64(key string, missing *[]string) float64 {
	val, ok := lookupRequired(key, missing)
	if !ok {
		return 0
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		*missing = append(*missing, key)
		return 0
	}
	return f
}

func mustBool(key string, missing *[]string) bool {
	val, ok := lookupRequired(key, missing)
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		*missing = append(*missing, key)
		return false
	}
}

