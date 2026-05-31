package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id"`
}

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Code:    0,
		Message: message,
		Data:    data,
		TraceID: TraceID(c),
	})
}

func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, "success", data)
}
