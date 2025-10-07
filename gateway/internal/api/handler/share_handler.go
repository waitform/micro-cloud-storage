package handler

import (
	"context"
	"net/http"

	"cloud-storage/internal/pack"
	"cloud-storage/internal/rpc"
	sharepb "cloud-storage/protos/share/proto"
	"cloud-storage/utils"

	"github.com/gin-gonic/gin"
)

type ShareHandler struct {
	shareClient *rpc.ShareServiceClient
}

func NewShareHandler(shareClient *rpc.ShareServiceClient) *ShareHandler {
	return &ShareHandler{
		shareClient: shareClient,
	}
}

// HandleCreateShare 处理创建分享请求
func (h *ShareHandler) HandleCreateShare(c *gin.Context) {
	var req sharepb.CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.shareClient.CreateShare(ctx, &req)
	if err != nil {
		utils.Error("Failed to create share: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to create share")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Share created successfully", resp)
}

// HandleGetShareInfo 处理获取分享信息请求
func (h *ShareHandler) HandleGetShareInfo(c *gin.Context) {
	// 从查询参数获取分享ID
	shareID := c.Query("share_id")
	if shareID == "" {
		pack.WriteError(c, http.StatusBadRequest, "Missing share_id parameter")
		return
	}

	req := &sharepb.GetShareInfoRequest{
		ShareId: shareID,
	}

	ctx := context.Background()
	resp, err := h.shareClient.GetShareInfo(ctx, req)
	if err != nil {
		utils.Error("Failed to get share info: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to get share info")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Share info retrieved successfully", resp.GetInfo())
}

// HandleValidateAccess 处理验证访问权限请求
func (h *ShareHandler) HandleValidateAccess(c *gin.Context) {
	var req sharepb.ValidateAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.shareClient.ValidateAccess(ctx, &req)
	if err != nil {
		utils.Error("Failed to validate access: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to validate access")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Access validation completed", resp)
}
