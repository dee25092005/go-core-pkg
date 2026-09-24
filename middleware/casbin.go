package middleware

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/dee25092005/go-core-pkg/apperrors"
	"github.com/labstack/echo/v4"
)

func CasbinAuthZ(enforcer *casbin.Enforcer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Request().Header.Get("X-user")
			if user == "" {
				user = "anonymous"
			}

			path := c.Request().URL.Path
			method := c.Request().Method

			ok, err := enforcer.Enforce(user, path, method)
			if err != nil {
				return fmt.Errorf("failed to enforce: %w", err)
			}

			if !ok {
				return apperrors.Forbidden("Access denied")
			}

			return next(c)
		}
	}
}
