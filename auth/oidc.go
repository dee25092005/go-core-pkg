package auth

import (
	"errors"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type OIDValidator struct {
	jwks *keyfunc.JWKS
}

func NewOIDCValidator(jwksURL string) (*OIDValidator, error) {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour * 24,
	})
	if err != nil {
		return nil, err
	}
	return &OIDValidator{
		jwks: jwks,
	}, nil
}

type GooglClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	jwt.RegisteredClaims
}

func (v *OIDValidator) VerifyGoogleIDToken(tokenString, googleClientID string) (*GooglClaims, error) {
	Claims := &GooglClaims{}

	token, err := jwt.ParseWithClaims(tokenString, Claims, v.jwks.Keyfunc)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if len(Claims.Audience) == 0 || Claims.Audience[0] != googleClientID {
		return nil, errors.New("invalid audience")
	}

	return Claims, nil
}
