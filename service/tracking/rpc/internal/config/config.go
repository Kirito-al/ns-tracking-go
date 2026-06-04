package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// DatabaseConfig 数据库连接池配置
type DatabaseConfig struct {
	MaxOpenConns    int `json:",default=100"`
	MaxIdleConns    int `json:",default=10"`
	ConnMaxLifetime int `json:",default=300"` // 秒
	ConnMaxIdleTime int `json:",default=60"`  // 秒
}

// TimeoutConfig 服务超时配置
type TimeoutConfig struct {
	QueryTimeout int `json:",default=5"`  // SQL 查询超时 (秒)
	RpcTimeout   int `json:",default=10"` // RPC 调用超时 (秒)
}

// Config tracking-srv gRPC 服务配置
type Config struct {
	zrpc.RpcServerConf // gRPC 服务配置（内置：Host、Port、Timeout 等）

	// 数据库配置
	DataSource string `json:",optional"` // PostgreSQL 连接字符串

	// Redis 配置（使用 go-zero 内置 RedisConf）
	RedisConf redis.RedisConf `json:",optional"`

	// Asynq Redis 配置（队列专用，Db=1）
	AsynqRedisConf AsynqRedisConfig `json:",optional"`

	// 数据库连接池配置
	DatabaseConfig DatabaseConfig `json:",optional"`

	// 超时配置
	TimeoutConfig TimeoutConfig `json:",optional"`

	// 业务配置
	YunExpressToken   string `json:",optional"` // 云途普通账号 Token
	GEYunExpressToken string `json:",optional"` // 云途 GE 账号 Token
}

// AsynqRedisConfig Asynq队列Redis 配置
type AsynqRedisConfig struct {
	Addr       string `json:",default=localhost:6379"` // Redis 地址
	Password   string `json:",optional"`              // Redis 密码
	DB         int    `json:",default=1"`             // Redis DB（默认 1，与缓存隔离）
	Concurrency int    `json:",default=10"`            // Worker 并发数
}
