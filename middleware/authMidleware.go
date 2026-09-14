package middleware

import (
	"cloud_storage-backend/internal/auth"
	"cloud_storage-backend/pkg/apperrors"
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return apperrors.Unauthorized("Authorization header is missing")
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				return apperrors.Unauthorized("Authorization header is invalid")
			}
			rawToken := strings.TrimPrefix(authHeader, prefix)

			claims, err := auth.VerifyJWT(rawToken, jwtSecret)

			if err != nil {
				return apperrors.Unauthorized("Invalid token or token is expired")
			}

			c.Set("user_id", claims.UserID)
			c.Set("tier", claims.Tier)

			return next(c)
		}
	}
}
