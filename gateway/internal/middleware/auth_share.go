package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	pack "github.com/waitform/micro-cloud-storage/internal/pack"
	"github.com/waitform/micro-cloud-storage/internal/rpc"
	sharepb "github.com/waitform/micro-cloud-storage/protos/share/proto"
)

// AuthShareMiddleware 分享鉴权中间件
func AuthShareMiddleware(shareClient *rpc.ShareServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取分享ID
		shareID := c.Query("share_id")
		if shareID == "" {
			pack.WriteError(c, http.StatusBadRequest, "share_id is required")
			return
		}

		// 获取密码（可选）
		password := c.Query("password")

		// 构造验证请求
		req := &sharepb.ValidateAccessRequest{
			ShareId:  shareID,
			Password: password,
		}

		// 调用分享服务验证访问权限
		ctx := context.Background()
		resp, err := shareClient.ValidateAccess(ctx, req)
		if err != nil {
			pack.WriteError(c, http.StatusForbidden, "Failed to validate share access")
			return
		}

		// 检查访问是否有效
		if !resp.GetValid() {
			pack.WriteError(c, http.StatusForbidden, "Invalid share access")
			return
		}

		// 将文件ID和所有者ID存储到上下文中供后续使用
		c.Set("file_id", resp.GetFileId())
		c.Set("owner_id", resp.GetOwnerId())

		// 继续处理请求
		c.Next()
	}
}
