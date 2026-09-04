package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"todo-backend/pkg/apperrors"

	"github.com/labstack/echo/v4"
)

func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed || err == nil {
		return
	}

	var appErr *apperrors.AppError
	var echoErr *echo.HTTPError

	switch {
	case errors.As(err, &appErr):
		if appErr.Code >= 500 {
			slog.Error("internal server error",
				"path", c.Path(),
				"method", c.Request().Method,
				"cause", appErr.Err,
			)
		}
		_ = c.JSON(appErr.Code, echo.Map{"error": appErr.Message})
		return
	case errors.As(err, &echoErr):
		msg := "request error"
		if strMsg, ok := echoErr.Message.(string); ok {
			msg = strMsg
		}
		_ = c.JSON(echoErr.Code, echo.Map{"error": msg})
		return
	}
	slog.Error("unhandled error",
		slog.String("path", c.Path()),
		slog.String("method", c.Request().Method),
		slog.Any("error", err),
	)
	_ = c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})

}
