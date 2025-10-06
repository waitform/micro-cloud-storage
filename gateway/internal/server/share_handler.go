package server

import (
	sharepb "cloud-storage/protos/share/proto"
	"context"
	"net/http"

	"cloud-storage/utils"
	"github.com/gin-gonic/gin"
)

// handleCreateShare 处理创建分享请求
func (s *GatewayServer) handleCreateShare(c *gin.Context) {
	var req sharepb.CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := s.ShareClient.CreateShare(ctx, &req)
	if err != nil {
		utils.Error("Failed to create share: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to create share")
		return
	}

	s.writeJSON(c, http.StatusOK, "Share created successfully", resp)
}

// handleGetShareInfo 处理获取分享信息请求
func (s *GatewayServer) handleGetShareInfo(c *gin.Context) {
	// 从查询参数获取分享ID
	shareID := c.Query("share_id")
	if shareID == "" {
		s.writeError(c, http.StatusBadRequest, "Missing share_id parameter")
		return
	}

	req := &sharepb.GetShareInfoRequest{
		ShareId: shareID,
	}

	ctx := context.Background()
	resp, err := s.ShareClient.GetShareInfo(ctx, req)
	if err != nil {
		utils.Error("Failed to get share info: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to get share info")
		return
	}

	s.writeJSON(c, http.StatusOK, "Share info retrieved successfully", resp.GetInfo())
}

// handleValidateAccess 处理验证访问权限请求
func (s *GatewayServer) handleValidateAccess(c *gin.Context) {
	var req sharepb.ValidateAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := s.ShareClient.ValidateAccess(ctx, &req)
	if err != nil {
		utils.Error("Failed to validate access: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to validate access")
		return
	}

	s.writeJSON(c, http.StatusOK, "Access validation completed", resp)
}