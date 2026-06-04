package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config tracking-api HTTP 服务配置
type Config struct {
	Name      string
	Host      string
	Port      int

	// gRPC Client配置
	TrackingRpc struct {
		Target string
		Timeout int64
	}

	// Redis配置
	RedisConf struct {
		Host string
		Pass string
		Db   int
	}

	// Webhook安全配置
	YunExpressWebhookSecret      string
	GEYunExpressWebhookSecret     string
	YunExpressSkipSignature       bool

	// 灰度开关配置（新增）
	GrayScale GrayScaleConfig `yaml:"gray_scale"`
}
