package server

import (
	userpb "cloud-storage/protos/user/proto"
	"context"
	"net/http"
	"strconv"

	"cloud-storage/utils"
	"github.com/gin-gonic/gin"
)

// handleUserRegister 处理用户注册请求
func (s *GatewayServer) handleUserRegister(c *gin.Context) {
	var req userpb.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := s.UserClient.Register(ctx, &req)
	if err != nil {
		utils.Error("Failed to register user: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	s.writeJSON(c, http.StatusOK, "User registered successfully", resp)
}

// handleUserLogin 处理用户登录请求
func (s *GatewayServer) handleUserLogin(c *gin.Context) {
	var req userpb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := s.UserClient.Login(ctx, &req)
	if err != nil {
		utils.Error("Failed to login user: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to login user")
		return
	}

	s.writeJSON(c, http.StatusOK, "User logged in successfully", resp)
}

// handleGetUserInfo 处理获取用户信息请求
func (s *GatewayServer) handleGetUserInfo(c *gin.Context) {
	// 从查询参数获取用户ID
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		s.writeError(c, http.StatusBadRequest, "Missing user_id parameter")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid user_id parameter")
		return
	}

	req := &userpb.GetUserInfoRequest{
		UserId: userID,
	}

	ctx := context.Background()
	resp, err := s.UserClient.GetUserInfo(ctx, req)
	if err != nil {
		utils.Error("Failed to get user info: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	s.writeJSON(c, http.StatusOK, "User info retrieved successfully", resp.GetUser())
}