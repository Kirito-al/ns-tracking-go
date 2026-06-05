package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	DataSource string `yaml:"DataSource"` // PostgreSQL 连接字符串
	RedisConf struct {
		Host string `yaml:"Host"`
		Pass string `yaml:"Pass"`
		Db   int    `yaml:"Db"`
	} `yaml:"RedisConf"`

	YunExpressWebhookSecret      string `yaml:"YunExpressWebhookSecret"`
	GEYunExpressWebhookSecret    string `yaml:"GEYunExpressWebhookSecret"`
	YunExpressSkipSignature      bool   `yaml:"YunExpressSkipSignature"`

	// 业务配置
	ProviderName string `yaml:"ProviderName"` // 服务商名称（如：云途物流）
	ServiceClass string `yaml:"ServiceClass"` // 服务类型（如：YunExpressService）

	CORS struct {
		AllowedOrigins []string `yaml:"AllowedOrigins"`
		AllowedMethods []string `yaml:"AllowedMethods"`
	} `yaml:"CORS"`
}
