package apperrors

import (
	"fmt"
	"net/http"
)

type FieldViolation struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
	Error  string `json:"error"`
}

type AppError struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Status  string        `json:"status"`
	Detail  []interface{} `json:"detail,omitempty"`
	Err     error         `json:"-"`
}

type ErrorResponse struct {
	Error *AppError `json:"error"`
}

var (
	ErrNotFound     = NotFound("Resource not found")
	ErrBadRequest   = BadRequest("Bad request")
	ErrUnauthorized = Unauthorized("Unauthorized access")
	ErrForbidden    = Forbidden("Forbidden access")
	ErrConflict     = Conflict("Resource already exists")
)

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithDetail(detail interface{}) *AppError {
	e.Detail = append(e.Detail, detail)
	return e
}

func NotFound(msg string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Status:  "NOT_FOUND",
		Message: msg,
	}
}

func BadRequest(msg string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Status:  "INVALID_ARGUMENT",
		Message: msg,
	}
}

func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Status:  "UNAUTHENTICATED",
		Message: msg,
	}
}

func Forbidden(msg string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Status:  "FORBIDDEN",
		Message: msg,
	}
}

func Internal(err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Status:  "INTERNAL_ERROR",
		Message: "Internal Server Error",
		Err:     err,
	}
}

func Conflict(msg string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Status:  "ALREADY_EXISTS",
		Message: msg,
	}
}
