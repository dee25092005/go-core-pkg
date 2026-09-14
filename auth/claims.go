package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Tier   string `json:"tier"`
	jwt.RegisteredClaims
}

func NewClaims(userID, tier string, ttl time.Duration) Claims {
	now := time.Now()
	return Claims{
		UserID: userID,
		Tier:   tier,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
}
