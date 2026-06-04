package svc

import (
	"ns-tracking-go/service/tracking/api/internal/config"
	"ns-tracking-go/service/tracking/rpc/tracking"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// ServiceContext 服务上下文（依赖注入容器）
type ServiceContext struct {
	Config config.Config

	// Redis 连接（缓存 + Session）
	Redis *redis.Redis

	// gRPC Client（连接 tracking-srv）
	TrackingRpc tracking.TrackingServiceClient

	// 内部 gRPC 连接（用于关闭）
	grpcConn *grpc.ClientConn
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 1. 创建 Redis 连接
	var redisClient *redis.Redis
	if c.RedisConf.Host != "" {
		redisClient = redis.MustNewRedis(c.RedisConf)
	}

	// 2. 创建 gRPC Client（直连模式，不使用 etcd）
	// 开发环境直接连接 127.0.0.1:50051
	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{c.TrackingRpc.Target},
		Timeout:   c.TrackingRpc.Timeout,
		NonBlock:  c.TrackingRpc.NonBlock,
	})

	return &ServiceContext{
		Config:      c,
		Redis:       redisClient,
		TrackingRpc: tracking.NewTrackingServiceClient(client.Conn()),
		grpcConn:    client.Conn(), // 保存 gRPC 连接用于关闭
	}
}

// Close 关闭所有资源连接
func (ctx *ServiceContext) Close() {
	// 关闭 gRPC Client 连接
	if ctx.grpcConn != nil {
		ctx.grpcConn.Close()
	}
}