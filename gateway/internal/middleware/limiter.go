package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waitform/micro-cloud-storage/utils"
)

func IPRateLimitMiddleware(ipratelimiter *utils.IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := ipratelimiter.GetLimiter(ip)
		if limiter.Allow() {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests from your IP, please slow down",
			})
		}

	}
}

// TransferRateLimiterMiddleware 下载限流中间件
// 使用公平传输管理器来限制下载速率，确保带宽公平分配
func TransferRateLimiterMiddleware(transferManager *utils.FairTransferManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID（如果已登录）
		var userID uint
		if userIDVal, exists := c.Get("user_id"); exists {
			userID = uint(userIDVal.(int64))
		}

		// 获取全局限速器
		globalLimiter := transferManager.GetGlobalLimiter()

		// 获取用户限速器
		userLimiter := transferManager.GetUserLimiter(userID)

		// 将限速器存储到上下文中，供后续处理使用
		c.Set("global_limiter", globalLimiter)
		c.Set("user_limiter", userLimiter)

		// 继续处理请求
		c.Next()
	}
}
