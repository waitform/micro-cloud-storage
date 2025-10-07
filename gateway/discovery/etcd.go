package discovery

import (
	"cloud-storage/utils"
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type EtcdClient struct {
	client *clientv3.Client
}

// NewEtcdClient 创建 etcd 客户端实例
func NewEtcdClient(endpoints []string) (*EtcdClient, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = c.Get(ctx, "health")
	if err != nil {
		// 不是致命错误，仅记录日志
		utils.Warn("etcd health check failed: %v", err)
	}

	utils.Info("etcd client initialized")
	return &EtcdClient{client: c}, nil
}

// Register 注册服务并自动续约
func (e *EtcdClient) Register(serviceName, addr string, ttl int64) (context.CancelFunc, error) {
	// 创建租约
	leaseResp, err := e.client.Grant(context.Background(), ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to create lease: %w", err)
	}

	key := fmt.Sprintf("/services/%s/%s", serviceName, addr)
	value := addr

	// 尝试注册服务，带重试机制
	maxRetries := 3
	var putErr error
	for i := 0; i < maxRetries; i++ {
		_, putErr = e.client.Put(context.Background(), key, value, clientv3.WithLease(leaseResp.ID))
		if putErr == nil {
			break
		}
		utils.Warn("failed to register service (attempt %d/%d): %v", i+1, maxRetries, putErr)
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1)) // 逐步增加延迟
		}
	}

	if putErr != nil {
		// 清理租约
		e.client.Revoke(context.Background(), leaseResp.ID)
		return nil, fmt.Errorf("failed to register service after %d attempts: %w", maxRetries, putErr)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 启动租约续约
	ch, err := e.client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		cancel()
		e.client.Revoke(context.Background(), leaseResp.ID)
		return nil, fmt.Errorf("failed to start keepalive: %w", err)
	}

	// 启动续约监控协程
	go func() {
		defer func() {
			// 确保在函数退出时撤销租约
			e.client.Revoke(context.Background(), leaseResp.ID)
			utils.Info("revoked lease for service %s at %s", serviceName, addr)
		}()

		for {
			select {
			case _, ok := <-ch:
				if !ok {
					utils.Info("keepalive channel closed for service %s at %s", serviceName, addr)
					return
				}
				// 续约成功，继续循环
			case <-ctx.Done():
				utils.Info("context cancelled for service %s at %s", serviceName, addr)
				return
			}
		}
	}()

	utils.Info("registered %s -> %s with TTL %ds", serviceName, addr, ttl)
	return cancel, nil
}

// Unregister 注销服务
func (e *EtcdClient) Unregister(serviceName, addr string) error {
	key := fmt.Sprintf("/services/%s/%s", serviceName, addr)
	_, err := e.client.Delete(context.Background(), key)
	if err != nil {
		return fmt.Errorf("failed to unregister service %s at %s: %w", serviceName, addr, err)
	}
	utils.Info("unregistered %s -> %s", serviceName, addr)
	return nil
}

// Discover 查询某个服务的可用地址
func (e *EtcdClient) Discover(serviceName string) ([]string, error) {
	resp, err := e.client.Get(context.Background(), "/services/"+serviceName, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}

	addrs := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}

	utils.Info("discovered %d instances for service %s", len(addrs), serviceName)
	return addrs, nil
}

// DiscoverWithWatch 监听服务变化
func (e *EtcdClient) DiscoverWithWatch(ctx context.Context, serviceName string) (<-chan []string, error) {
	watchChan := make(chan []string, 1)

	// 先获取当前的服务列表
	addrs, err := e.Discover(serviceName)
	if err != nil {
		close(watchChan)
		return nil, err
	}

	// 发送当前服务列表
	watchChan <- addrs

	// 启动监听协程
	go func() {
		defer close(watchChan)

		rch := e.client.Watch(ctx, "/services/"+serviceName, clientv3.WithPrefix())
		for {
			select {
			case wresp := <-rch:
				if wresp.Err() != nil {
					utils.Warn("watch error for service %s: %v", serviceName, wresp.Err())
					return
				}

				// 重新获取服务列表
				newAddrs, err := e.Discover(serviceName)
				if err != nil {
					utils.Warn("failed to rediscover service %s: %v", serviceName, err)
					continue
				}

				// 发送更新后的服务列表
				select {
				case watchChan <- newAddrs:
				case <-ctx.Done():
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return watchChan, nil
}

// GetClient 获取原始etcd客户端（谨慎使用）
func (e *EtcdClient) GetClient() *clientv3.Client {
	return e.client
}

// Close 关闭etcd客户端连接
func (e *EtcdClient) Close() error {
	if e.client != nil {
		err := e.client.Close()
		if err != nil {
			return fmt.Errorf("failed to close etcd client: %w", err)
		}
		utils.Info("etcd client closed")
	}
	return nil
}

// CreateMutex 创建分布式锁
func (e *EtcdClient) CreateMutex(key string) (*concurrency.Mutex, error) {
	session, err := concurrency.NewSession(e.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	mutex := concurrency.NewMutex(session, key)
	return mutex, nil
}
