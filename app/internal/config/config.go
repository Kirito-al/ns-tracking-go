package config

import (
	"os"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	DataSource string `yaml:"datasource"`

	RedisHost string `yaml:"redis_host"`
	RedisPass string `yaml:"redis_pass"`
	RedisType string `yaml:"redis_type"`

	// Asynq配置（Worker队列）
	AsynqRedisHost    string `yaml:"asynq_redis_host"`     // ← 新增：Asynq Redis地址（默认与缓存Redis相同）
	AsynqRedisPass    string `yaml:"asynq_redis_pass"`     // ← 新增：Asynq Redis密码
	AsynqRedisDB      int    `yaml:"asynq_redis_db"`       // ← 新增：Asynq Redis DB（默认=1，与缓存隔离）
	AsynqConcurrency  int    `yaml:"asynq_concurrency"`    // ← 新增：Worker并发数（默认=10）

	YunExpressWebhookSecret   string `yaml:"yun_express_webhook_secret"`
	GEYunExpressWebhookSecret string `yaml:"ge_yun_express_webhook_secret"`
	YunExpressEncryptKey      string `yaml:"yun_express_encrypt_key"`
	YunExpressSkipSignature   bool   `yaml:"yun_express_skip_signature"`

	ProviderName string `yaml:"provider_name"`
	ServiceClass string `yaml:"service_class"`
}

// Load 加载配置文件
func Load(file string) (*Config, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
