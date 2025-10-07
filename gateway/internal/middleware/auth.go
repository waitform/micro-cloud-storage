package middleware

import (
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里应该实现JWT验证逻辑
		// 为简化示例，暂时跳过验证
		c.Next()
	}
}