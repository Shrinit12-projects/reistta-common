package auth

import "time"

type SecureSession struct {
	UserID     string    `json:"user_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	MerchantID string    `json:"merchant_id"`
	IssuedAt   time.Time `json:"iat"`
	ExpiresAt  time.Time `json:"exp"`
	IPAddress  string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
}
