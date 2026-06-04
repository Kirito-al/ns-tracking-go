package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	DataSource string // PostgreSQL 连接字符串
	RedisConf struct {
		Host string
		Pass string
		Db   int
	}

	YunExpressWebhookSecret      string
	GEYunExpressWebhookSecret    string
	YunExpressSkipSignature      bool

	// 业务配置
	ProviderName string // 服务商名称（如：云途物流）
	ServiceClass string // 服务类型（如：YunExpressService）

	CORS struct {
		AllowedOrigins []string
		AllowedMethods []string
	}
}
