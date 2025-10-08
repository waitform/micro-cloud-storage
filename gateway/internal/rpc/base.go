package rpc

import (
	"context"
	"time"

	"github.com/waitform/micro-cloud-storage/discovery"
	"github.com/waitform/micro-cloud-storage/global"
	"github.com/waitform/micro-cloud-storage/utils"
)

// ServiceClient 封装服务客户端的基础结构
type ServiceClient struct {
	EtcdClient *discovery.EtcdClient
	GlobalCfg  *global.GlobalConfig
}

// NewServiceClient 创建服务客户端
func NewServiceClient(etcdClient *discovery.EtcdClient, globalCfg *global.GlobalConfig) *ServiceClient {
	return &ServiceClient{
		EtcdClient: etcdClient,
		GlobalCfg:  globalCfg,
	}
}

// GetContextWithTimeout 创建带超时的上下文
func (s *ServiceClient) GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// Close 关闭服务客户端，清理资源
func (s *ServiceClient) Close() {
	// 由于使用了gRPC内置的服务发现，这里不需要额外清理
	utils.Info("service client closed")
}
