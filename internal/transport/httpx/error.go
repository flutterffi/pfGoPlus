package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AppError struct {
	Status  int         `json:"-"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Err     error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewError(status, code int, message string, err error) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func BadRequest(message string, err error) *AppError {
	return NewError(http.StatusBadRequest, 400000, message, err)
}

func NotFound(message string, err error) *AppError {
	return NewError(http.StatusNotFound, 404000, message, err)
}

func Internal(message string, err error) *AppError {
	return NewError(http.StatusInternalServerError, 500000, message, err)
}

func WriteError(c *gin.Context, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = Internal("internal server error", err)
	}

	Logger(c).Error("request failed",
		zap.Error(err),
	)

	c.JSON(appErr.Status, gin.H{
		"code":     appErr.Code,
		"message":  appErr.Message,
		"details":  appErr.Details,
		"trace_id": TraceID(c),
	})
}
