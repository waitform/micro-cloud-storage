package rpc

import (
	"cloud-storage/discovery"
	"cloud-storage/global"
	"cloud-storage/utils"
	"context"
	"fmt"
	"time"
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

// GetServiceAddr 获取服务地址
func (s *ServiceClient) GetServiceAddr(serviceName string) (string, error) {
	// 从etcd获取服务地址
	addrs, err := s.EtcdClient.Discover(serviceName)
	if err != nil {
		utils.Error("discover service %s failed: %v", serviceName, err)
		return "", err
	}

	if len(addrs) == 0 {
		err := fmt.Errorf("no available instances for service %s", serviceName)
		utils.Error("discover service %s failed: %v", serviceName, err)
		return "", err
	}

	// 简单的负载均衡：随机选择一个地址
	// 在实际项目中可以实现更复杂的负载均衡算法
	addr := addrs[0]
	utils.Info("discovered service %s at %s", serviceName, addr)
	return addr, nil
}

// GetContextWithTimeout 创建带超时的上下文
func (s *ServiceClient) GetContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}