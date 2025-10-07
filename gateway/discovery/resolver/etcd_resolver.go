package resolver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud-storage/discovery"
	"cloud-storage/utils"

	"google.golang.org/grpc/resolver"
)

// EtcdResolver 实现gRPC的名称解析器
type EtcdResolver struct {
	client  *discovery.EtcdClient
	service string
	cc      resolver.ClientConn
	ctx     context.Context
	cancel  context.CancelFunc
	freq    time.Duration
}

// schemeName 自定义scheme名称
const schemeName = "etcd"

// Build 实现resolver.Builder接口
func (r *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	service := target.URL.Host
	if service == "" {
		service = target.URL.Path
	}
	service = strings.TrimPrefix(service, "/")

	if service == "" {
		return nil, fmt.Errorf("service name is required in target: %s", target.URL.String())
	}

	// 创建etcd解析器
	etcdResolver := &EtcdResolver{
		client:  r.client,
		service: service,
		cc:      cc,
		freq:    30 * time.Second, // 默认30秒刷新一次
	}

	// 创建上下文
	etcdResolver.ctx, etcdResolver.cancel = context.WithCancel(context.Background())

	// 立即解析一次
	if err := etcdResolver.resolve(); err != nil {
		utils.Warn("initial resolve failed: %v", err)
	}

	// 启动后台监听
	go etcdResolver.watch()

	return etcdResolver, nil
}

// Scheme 实现resolver.Builder接口
func (r *etcdBuilder) Scheme() string {
	return schemeName
}

// resolve 解析服务地址
func (r *EtcdResolver) resolve() error {
	// 从etcd获取服务地址
	addrs, err := r.client.Discover(r.service)
	if err != nil {
		return fmt.Errorf("failed to discover service %s: %w", r.service, err)
	}

	// 转换为gRPC地址格式
	grpcAddrs := make([]resolver.Address, len(addrs))
	for i, addr := range addrs {
		grpcAddrs[i] = resolver.Address{Addr: addr}
	}

	// 更新地址
	r.cc.UpdateState(resolver.State{Addresses: grpcAddrs})
	utils.Info("resolved service %s with %d instances", r.service, len(addrs))
	return nil
}

// watch 监听服务变化
func (r *EtcdResolver) watch() {
	ticker := time.NewTicker(r.freq)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			utils.Info("stopped watching service %s", r.service)
			return
		case <-ticker.C:
			if err := r.resolve(); err != nil {
				utils.Error("failed to resolve service %s: %v", r.service, err)
			}
		}
	}
}

// ResolveNow 实现resolver.Resolver接口
func (r *EtcdResolver) ResolveNow(resolver.ResolveNowOptions) {
	if err := r.resolve(); err != nil {
		utils.Error("failed to resolve service %s: %v", r.service, err)
	}
}

// Close 实现resolver.Resolver接口
func (r *EtcdResolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
	utils.Info("closed resolver for service %s", r.service)
}

// etcdBuilder 实现resolver.Builder接口
type etcdBuilder struct {
	client *discovery.EtcdClient
}

// NewEtcdBuilder 创建etcd解析器构建器
func NewEtcdBuilder(client *discovery.EtcdClient) resolver.Builder {
	return &etcdBuilder{client: client}
}

// RegisterEtcdResolver 注册etcd解析器
func RegisterEtcdResolver(client *discovery.EtcdClient) {
	resolver.Register(NewEtcdBuilder(client))
	utils.Info("registered etcd resolver with scheme: %s", schemeName)
}
