package pack

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JSONResponse 标准JSON响应格式
type JSONResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WriteJSON 写入JSON响应
func WriteJSON(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, JSONResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// WriteError 写入错误响应
func WriteError(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, JSONResponse{
		Code:    httpCode,
		Message: message,
	})
}
