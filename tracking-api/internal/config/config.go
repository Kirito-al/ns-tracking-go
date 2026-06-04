package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

// Config tracking-api HTTP 服务配置
type Config struct {
	rest.RestConf // go-zero REST 服务标准配置（包含 Name, Host, Port）

	// gRPC Client 配置（开发环境直连）
	TrackingRpc struct {
		Target string `json:",optional"`
		Timeout int64 `json:",optional"`
		NonBlock bool `json:",optional"`
	}

	// Redis 配置
	RedisConf redis.RedisConf

	// Webhook 安全配置
	YunExpressWebhookSecret      string
	GEYunExpressWebhookSecret     string
	YunExpressSkipSignature       bool

	// 灰度开关配置（新增）
	GrayScale GrayScaleConfig `yaml:"gray_scale"`
}
