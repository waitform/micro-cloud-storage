package middleware

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	pack "github.com/waitform/micro-cloud-storage/internal/pack"
)

// CasbinMiddleware Casbin鉴权中间件
type CasbinMiddleware struct {
	enforcer *casbin.Enforcer
}

// NewCasbinMiddleware 创建Casbin中间件实例
func NewCasbinMiddleware(enforcer *casbin.Enforcer) (*CasbinMiddleware, error) {

	return &CasbinMiddleware{
		enforcer: enforcer,
	}, nil
}

// RequirePermission 权限检查中间件
func (cm *CasbinMiddleware) RequirePermission(objPrefix string, paramKey string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中获取用户ID
		userIDValue, exists := c.Get("user_id")
		if !exists {
			pack.WriteError(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		userID, ok := userIDValue.(string)
		if !ok {
			pack.WriteError(c, http.StatusUnauthorized, "Invalid user ID")
			c.Abort()
			return
		}

		// 获取资源ID
		resourceID := c.Param(paramKey)
		if resourceID == "" {
			pack.WriteError(c, http.StatusBadRequest, "Resource ID is required")
			c.Abort()
			return
		}

		// 构造资源标识
		obj := objPrefix + resourceID

		// 检查权限
		allowed, err := cm.enforcer.Enforce(userID, obj, action)
		if err != nil {
			pack.WriteError(c, http.StatusInternalServerError, "Error occurred when authorizing user")
			c.Abort()
			return
		}

		if !allowed {
			pack.WriteError(c, http.StatusForbidden, "You don't have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}
