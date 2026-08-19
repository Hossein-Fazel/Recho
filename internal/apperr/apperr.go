package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorType string

const (
	ErrInternal     ErrorType = "INTERNAL_ERROR"
	ErrInvalidInput ErrorType = "INVALID_INPUT"
	ErrNotFound     ErrorType = "NOT_FOUND"
	ErrConflict     ErrorType = "DUPLICATE_ERROR"
)

type AppError struct {
	Status  int
	Type    ErrorType
	Message string
	Module  string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s/%s] %s: %v", e.Module, e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s/%s] %s", e.Module, e.Type, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

func New(status int, Type ErrorType, module, message string, cause error) *AppError {
	return &AppError{
		Status:  status,
		Type:    Type,
		Module:  module,
		Message: message,
		Err:     cause,
	}
}

func InvalidInput(module, msg string, cause error) *AppError {
	return New(http.StatusBadRequest, ErrInvalidInput, module, msg, cause)
}

func NotFound(module, msg string, cause error) *AppError {
	return New(http.StatusNotFound, ErrNotFound, module, msg, cause)
}

func Conflict(module, msg string, cause error) *AppError {
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
