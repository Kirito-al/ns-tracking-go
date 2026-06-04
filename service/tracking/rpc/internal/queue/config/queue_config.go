package config

import (
	"github.com/hibiken/asynq"
)

// AsynqConfig Asynq 队列配置
// 用途：配置 Redis 连接（Db=1，与缓存隔离）
type AsynqConfig struct {
	// Redis 连接配置
	RedisAddr     string // Redis 地址（如：localhost:6379）
	RedisPassword string // Redis 密码（空字符串表示无密码）
	RedisDB       int    // Redis DB（默认 1，与缓存隔离）

	// Worker 配置
	Concurrency int // 并发数（默认 10）
}

// NewAsynqClient 创建 Asynq Client（用于 Enqueue 任务）
func NewAsynqClient(cfg AsynqConfig) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}

// NewAsynqServer 创建 Asynq Server（用于 Worker 消费）
func NewAsynqServer(cfg AsynqConfig) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: cfg.Concurrency, // 并发数
			Queues: map[string]int{
				"default":  6,  // 默认队列优先级
				"critical": 10, // 关键任务优先级
				"low":      1,  // 低优先级任务
			},
			// ErrorHandler: 下期实现（投递死信队列或发送告警）
		},
	)
}

// DefaultAsynqConfig 默认配置（开发环境）
func DefaultAsynqConfig() AsynqConfig {
	return AsynqConfig{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       1, // Db=1，与缓存隔离
		Concurrency:   10,
	}
}