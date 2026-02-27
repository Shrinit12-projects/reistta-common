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
	AppEnv              string
	IsProduction        bool
	DatabaseURL         string
	RedisURL            string
	PGMaxConns          int32
	PGMinConns          int32
	PGMaxConnIdleTime   time.Duration
	PGHealthcheckPeriod time.Duration
	RedisPoolSize       int
	RedisMinIdleConns   int
	SessionSecret       string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	LoginMaxAttempts    int
	LoginLockoutTTL     time.Duration
	RateLimitRPS        float64
	RateLimitBurst      int
	S3Region            string
	S3Bucket            string
	S3Endpoint          string
	S3AccessKeyID       string
	S3SecretAccessKey   string
	S3UseSSL            bool
	MinioEndpoint       string
	MinioAccessKey      string
	MinioSecretKey      string
	MinioBucket         string
	MinioUseSSL         bool
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
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if appEnv == "" {
		appEnv = "development"
	}
	isProduction := appEnv == "production"

	var (
		s3Region          string
		s3Bucket          string
		s3Endpoint        string
		s3AccessKeyID     string
		s3SecretAccessKey string
		s3UseSSL          bool
		minioEndpoint     string
		minioAccessKey    string
		minioSecretKey    string
		minioBucket       string
		minioUseSSL       bool
	)

	// Stage-agnostic storage selection:
	// prefer S3 when complete S3 config is present, otherwise fallback to MinIO.
	s3Region = strings.TrimSpace(os.Getenv("S3_REGION"))
	s3Bucket = strings.TrimSpace(os.Getenv("S3_BUCKET"))
	s3AccessKeyID = strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID"))
	s3SecretAccessKey = strings.TrimSpace(os.Getenv("S3_SECRET_ACCESS_KEY"))
	s3Endpoint = strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	if s3Endpoint == "" && s3Region != "" {
		s3Endpoint = fmt.Sprintf("s3.%s.amazonaws.com", s3Region)
	}
	s3UseSSL = true
	if val, ok := os.LookupEnv("S3_USE_SSL"); ok && strings.TrimSpace(val) != "" {
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "true", "1", "yes":
			s3UseSSL = true
		case "false", "0", "no":
			s3UseSSL = false
		default:
			missing = append(missing, "S3_USE_SSL")
		}
	}

	minioEndpoint = strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	minioAccessKey = strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY"))
	minioSecretKey = strings.TrimSpace(os.Getenv("MINIO_SECRET_KEY"))
	minioBucket = strings.TrimSpace(os.Getenv("MINIO_BUCKET"))
	if val, ok := os.LookupEnv("MINIO_USE_SSL"); ok && strings.TrimSpace(val) != "" {
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "true", "1", "yes":
			minioUseSSL = true
		case "false", "0", "no":
			minioUseSSL = false
		default:
			missing = append(missing, "MINIO_USE_SSL")
		}
	}

	s3Configured := s3Region != "" && s3Bucket != "" && s3AccessKeyID != "" && s3SecretAccessKey != ""
	minioConfigured := minioEndpoint != "" && minioAccessKey != "" && minioSecretKey != "" && minioBucket != ""

	if s3Configured {
		// Keep legacy field names populated so existing services keep working.
		minioEndpoint = s3Endpoint
		minioAccessKey = s3AccessKeyID
		minioSecretKey = s3SecretAccessKey
		minioBucket = s3Bucket
		minioUseSSL = s3UseSSL
	} else if !minioConfigured {
		missing = append(missing,
			"S3_REGION|MINIO_ENDPOINT",
			"S3_BUCKET|MINIO_BUCKET",
			"S3_ACCESS_KEY_ID|MINIO_ACCESS_KEY",
			"S3_SECRET_ACCESS_KEY|MINIO_SECRET_KEY",
		)
	}

	if len(missing) > 0 {
		return Settings{}, fmt.Errorf("missing or invalid env vars: %s", strings.Join(missing, ", "))
	}

	return Settings{
		AppEnv:              appEnv,
		IsProduction:        isProduction,
		DatabaseURL:         dbURL,
		RedisURL:            redisURL,
		PGMaxConns:          pgMaxConns,
		PGMinConns:          pgMinConns,
		PGMaxConnIdleTime:   pgMaxIdle,
		PGHealthcheckPeriod: pgHealth,
		RedisPoolSize:       redisPool,
		RedisMinIdleConns:   redisMinIdle,
		SessionSecret:       sessionSecret,
		AccessTokenTTL:      accessTTL,
		RefreshTokenTTL:     refreshTTL,
		LoginMaxAttempts:    loginMax,
		LoginLockoutTTL:     loginLock,
		RateLimitRPS:        rateRPS,
		RateLimitBurst:      rateBurst,
		S3Region:            s3Region,
		S3Bucket:            s3Bucket,
		S3Endpoint:          s3Endpoint,
		S3AccessKeyID:       s3AccessKeyID,
		S3SecretAccessKey:   s3SecretAccessKey,
		S3UseSSL:            s3UseSSL,
		MinioEndpoint:       minioEndpoint,
		MinioAccessKey:      minioAccessKey,
		MinioSecretKey:      minioSecretKey,
		MinioBucket:         minioBucket,
		MinioUseSSL:         minioUseSSL,
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
