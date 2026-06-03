package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config tracking-api HTTP 服务配置
type Config struct {
	rest.RestConf // HTTP 服务配置（内置：Host、Port、Timeout、Mode 等）

	// Redis 配置（缓存 + Session）
	RedisConf redis.RedisConf `json:",optional"`

	// gRPC Client 配置
	TrackingRpc zrpc.RpcClientConf `json:"TrackingRpc"` // tracking-srv 的 gRPC 服务地址

	// 业务配置
	YunExpressWebhookSecret   string `json:",optional"` // 云途 Webhook 签名密钥（普通账号）
	GEYunExpressWebhookSecret string `json:",optional"` // 云途 Webhook 签名密钥（GE 账号）
	YunExpressSkipSignature   bool   `json:",optional"` // 是否跳过签名验证（开发环境）
}
