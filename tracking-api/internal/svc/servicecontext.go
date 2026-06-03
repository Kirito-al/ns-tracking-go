package svc

import (
	"tracking-api/internal/config"
	"tracking-srv/tracking"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 服务上下文（依赖注入容器）
type ServiceContext struct {
	Config config.Config

	// Redis 连接（缓存 + Session）
	Redis *redis.Redis

	// gRPC Client（连接 tracking-srv）
	TrackingRpc tracking.TrackingServiceClient
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 1. 创建 Redis 连接
	var redisClient *redis.Redis
	if c.RedisConf.Host != "" {
		redisClient = redis.MustNewRedis(c.RedisConf)
	}

	// 2. 创建 gRPC Client（服务发现:Etcd/Consul/Nacos）
	client := zrpc.MustNewClient(c.TrackingRpc)

	return &ServiceContext{
		Config:      c,
		Redis:       redisClient,
		TrackingRpc: tracking.NewTrackingServiceClient(client.Conn()),
	}
}