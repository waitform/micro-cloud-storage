package test

import (
	"context"
	"testing"
	"time"

	"cloud-storage-user-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestUserServiceClient_Register(t *testing.T) {
	// 连接到gRPC服务
	conn, err := grpc.Dial("localhost:35002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("无法连接到gRPC服务: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := proto.NewUserServiceClient(conn)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 测试注册
	req := &proto.RegisterRequest{
		Username: "testuser",
		Password: "testpassword",
		Email:    "test@example.com",
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		t.Fatalf("注册请求失败: %v", err)
	}

	if !resp.Success {
		t.Errorf("注册应该成功，但返回了失败: %s", resp.Message)
	}

	if resp.User == nil {
		t.Errorf("注册成功后应该返回用户信息")
	}
}

func TestUserServiceClient_Login(t *testing.T) {
	// 连接到gRPC服务
	conn, err := grpc.Dial("localhost:35002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("无法连接到gRPC服务: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := proto.NewUserServiceClient(conn)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 测试登录
	req := &proto.LoginRequest{
		Username: "testuser",
		Password: "testpassword",
	}

	resp, err := client.Login(ctx, req)
	if err != nil {
		t.Fatalf("登录请求失败: %v", err)
	}

	if !resp.Success {
		t.Errorf("登录应该成功，但返回了失败: %s", resp.Message)
	}

	if resp.Token == "" {
		t.Errorf("登录成功后应该返回token")
	}
}

func TestUserServiceClient_GetUserInfo(t *testing.T) {
	// 首先注册并登录以获取有效的用户ID和token
	conn, err := grpc.Dial("localhost:35002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("无法连接到gRPC服务: %v", err)
	}
	defer conn.Close()

	client := proto.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 注册用户
	registerReq := &proto.RegisterRequest{
		Username: "testuser_info",
		Password: "testpassword",
		Email:    "test_info@example.com",
	}

	registerResp, err := client.Register(ctx, registerReq)
	if err != nil {
		t.Fatalf("注册请求失败: %v", err)
	}

	if !registerResp.Success {
		t.Fatalf("注册应该成功，但返回了失败: %s", registerResp.Message)
	}

	// 测试获取用户信息
	getUserReq := &proto.GetUserInfoRequest{
		UserId: registerResp.User.Id,
	}

	getUserResp, err := client.GetUserInfo(ctx, getUserReq)
	if err != nil {
		t.Fatalf("获取用户信息失败: %v", err)
	}

	if getUserResp.User == nil {
		t.Errorf("应该返回用户信息")
	}

	if getUserResp.User.Id != getUserReq.UserId {
		t.Errorf("用户ID不匹配，期望: %d, 实际: %d", getUserReq.UserId, getUserResp.User.Id)
	}
}