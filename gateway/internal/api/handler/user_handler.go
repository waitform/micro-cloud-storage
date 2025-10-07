package handler

import (
	"context"
	"net/http"
	"strconv"

	"cloud-storage/internal/pack"
	"cloud-storage/internal/rpc"
	userpb "cloud-storage/protos/user/proto"
	"cloud-storage/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userClient *rpc.UserServiceClient
}

func NewUserHandler(userClient *rpc.UserServiceClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
	}
}

// HandleUserRegister 处理用户注册请求
func (h *UserHandler) HandleUserRegister(c *gin.Context) {
	var req userpb.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.userClient.Register(ctx, &req)
	if err != nil {
		utils.Error("Failed to register user: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "User registered successfully", resp)
}

// HandleUserLogin 处理用户登录请求
func (h *UserHandler) HandleUserLogin(c *gin.Context) {
	var req userpb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.userClient.Login(ctx, &req)
	if err != nil {
		utils.Error("Failed to login user: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to login user")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "User logged in successfully", resp)
}

// HandleGetUserInfo 处理获取用户信息请求
func (h *UserHandler) HandleGetUserInfo(c *gin.Context) {
	// 从查询参数获取用户ID
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		pack.WriteError(c, http.StatusBadRequest, "Missing user_id parameter")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid user_id parameter")
		return
	}

	req := &userpb.GetUserInfoRequest{
		UserId: userID,
	}

	ctx := context.Background()
	resp, err := h.userClient.GetUserInfo(ctx, req)
	if err != nil {
		utils.Error("Failed to get user info: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "User info retrieved successfully", resp.GetUser())
}
