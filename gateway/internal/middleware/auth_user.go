package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/waitform/micro-cloud-storage/utils"
)

// AuthUserMiddleware 鉴权中间件
func AuthUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取Authorization字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查Authorization格式是否正确 (Bearer <token>)
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Authorization header format must be Bearer <token>",
			})
			c.Abort()
			return
		}

		// 解析并验证token
		tokenString := parts[1]
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中，供后续处理使用
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		
		// 继续处理请求
		c.Next()
	}
}