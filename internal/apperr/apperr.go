package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrInternal     ErrorCode = "INTERNAL_ERROR"
	ErrInvalidInput ErrorCode = "INVALID_INPUT"
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrConflict 	ErrorCode = "DUPLICATE_ERROR"
)

type AppError struct {
	Status    int
	Code      ErrorCode
	Message   string
	Module    string
	Err       error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s/%s] %s: %v", e.Module, e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s/%s] %s", e.Module, e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

func New(status int, code ErrorCode, module, message string, cause error) *AppError {
	return &AppError{
		Status:    status,
		Code:      code,
		Module:    module,
		Message:   message,
		Err:       cause,
	}
}

func InvalidInput(module, msg string, cause error) *AppError {
	return New(http.StatusBadRequest, ErrInvalidInput, module, msg, cause)
}

func NotFound(module, msg string, cause error) *AppError {
	return New(http.StatusNotFound, ErrNotFound, module, msg, cause)
}

func Conflict(module, msg string, cause error) *AppError{
	return New(http.StatusConflict, ErrConflict, module, msg, cause)
}

func Internal(module string, cause error) *AppError {
	return New(http.StatusInternalServerError, ErrInternal, module, "internal server error", cause)
}

// As unwraps err into *AppError if possible.
func As(err error) (*AppError, bool) {
	var ae *AppError
	return ae, errors.As(err, &ae)
}
