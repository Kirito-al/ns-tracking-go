package cache

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// SystemCache 系统级缓存（配置、渠道信息等�?
type SystemCache struct {
	redis *redis.Redis
	prefix string
}

// NewSystemCache 创建系统缓存
func NewSystemCache(redisClient *redis.Redis) *SystemCache {
	return &SystemCache{
		redis:  redisClient,
		prefix: "system:",
	}
}

// GetConfig 获取系统配置
func (c *SystemCache) GetConfig(ctx context.Context, key string) (string, error) {
	cacheKey := fmt.Sprintf("%sconfig:%s", c.prefix, key)
	return c.redis.GetCtx(ctx, cacheKey)
}

// SetConfig 设置系统配置
func (c *SystemCache) SetConfig(ctx context.Context, key, value string, ttl int) error {
	cacheKey := fmt.Sprintf("%sconfig:%s", c.prefix, key)
	return c.redis.SetexCtx(ctx, cacheKey, value, ttl)
}

// GetChannelInfo 获取渠道信息（用�?GE 判断�?
func (c *SystemCache) GetChannelInfo(ctx context.Context, trackingNumber string) (string, error) {
	cacheKey := fmt.Sprintf("%schannel:%s", c.prefix, trackingNumber)
	return c.redis.GetCtx(ctx, cacheKey)
}

// SetChannelInfo 设置渠道信息
func (c *SystemCache) SetChannelInfo(ctx context.Context, trackingNumber, channelInfo string, ttl int) error {
	cacheKey := fmt.Sprintf("%schannel:%s", c.prefix, trackingNumber)
	return c.redis.SetexCtx(ctx, cacheKey, channelInfo, ttl)
}

// InvalidateChannelInfo 失效渠道信息缓存
func (c *SystemCache) InvalidateChannelInfo(ctx context.Context, trackingNumber string) error {
	cacheKey := fmt.Sprintf("%schannel:%s", c.prefix, trackingNumber)
	_, err := c.redis.DelCtx(ctx, cacheKey)
	return err
}

// HealthCheck Redis 健康检�?
func (c *SystemCache) HealthCheck(ctx context.Context) bool {
	return c.redis.PingCtx(ctx)
}
