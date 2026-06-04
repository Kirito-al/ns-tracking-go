package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	DataSource string // PostgreSQL 杩炴帴瀛楃涓?
	RedisConf struct {
		Host string
		Pass string
		Db   int
	}

	YunExpressWebhookSecret      string
	GEYunExpressWebhookSecret    string
	YunExpressSkipSignature      bool

	CORS struct {
		AllowedOrigins []string
		AllowedMethods []string
	}
}
