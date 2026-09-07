package middleware

import (
	"github.com/dee25092005/go-core-pkg/apperrors"
	"github.com/dee25092005/go-core-pkg/utils"
	"github.com/labstack/echo/v4"
)

type TokenRepository interface {
	FindUserByToken(token string) (string, error)
}

func AuthMiddleware(tokenRepo TokenRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return apperrors.Unauthorized("Authorization header is missing")
			}
			const prefix = "Bearer"
			if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
				return apperrors.Unauthorized("Authorization header is invalid")
			}
			rawToken := authHeader[len(prefix):]

			tokenHash := utils.HashToken(rawToken)
			userID, err := tokenRepo.FindUserByToken(tokenHash)
			if err != nil {
				return err
			}
			c.Set("user_id", userID)
			return next(c)
		}
	}
}
