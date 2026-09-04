package middleware

import (
	"errors"
	"log/slog"
	"time"
	"todo-backend/pkg/apperrors"

	"github.com/labstack/echo/v4"
)

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			req := c.Request()
			res := c.Response()

			status := res.Status
			if err != nil {
				var appErr *apperrors.AppError
				var echoErr *echo.HTTPError
				switch {
				case errors.As(err, &appErr):
					status = appErr.Code
				case errors.As(err, &echoErr):
					status = echoErr.Code
				default:
					status = 500
				}
			}

			slog.Info("http_request",
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.Int("status", status),
				slog.Int64("bytes_out", res.Size),
				slog.Duration("latency", time.Since(start)),
				slog.String("remote_ip", c.RealIP()),
			)
			return err

		}
	}
}
