package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// TrackingCache 轨迹缓存管理
// 用途：缓存热点轨迹数据，减少数据库查询
type TrackingCache struct {
	redis  *redis.Redis
	prefix string
	ttl    time.Duration
}

// NewTrackingCache 创建缓存实例
func NewTrackingCache(redisClient *redis.Redis) *TrackingCache {
	return &TrackingCache{
		redis:  redisClient,
		prefix: "tracking:detail:",
		ttl:    4 * time.Hour, // 对标 Ruby CACHE_HOURS = 4
	}
}

// TrackingDetailDTO 缓存数据结构
type TrackingDetailDTO struct {
	TrackingNumber string      `json:"tracking_number"`
	Detail         interface{} `json:"detail"`
	Status         int32       `json:"status"`
	ServiceClass   string      `json:"service_class"`
	SyncedAt       int64       `json:"synced_at"`
	ErrorMessage   string      `json:"error_message"`
	CachedAt       int64       `json:"cached_at"`
}

// Get 从缓存获取轨迹数�?
// 返回：缓存数据，命中=true，错�?
func (c *TrackingCache) Get(ctx context.Context, trackingNumber string) (*TrackingDetailDTO, bool, error) {
	key := fmt.Sprintf("%s%s", c.prefix, trackingNumber)

	// �?Redis 获取
	data, err := c.redis.GetCtx(ctx, key)
	if err != nil {
		// KeyError 表示缓存不存�?
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	// 解析 JSON
	var detail TrackingDetailDTO
	if err := json.Unmarshal([]byte(data), &detail); err != nil {
		return nil, false, err
	}

	return &detail, true, nil
}

// Set 写入缓存
func (c *TrackingCache) Set(ctx context.Context, trackingNumber string, detail *TrackingDetailDTO) error {
	key := fmt.Sprintf("%s%s", c.prefix, trackingNumber)

	// 序列化为 JSON
	data, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	// 写入 Redis（带 TTL�?
	err = c.redis.SetexCtx(ctx, key, string(data), int(c.ttl.Seconds()))
	return err
}

// Delete 删除缓存
func (c *TrackingCache) Delete(ctx context.Context, trackingNumber string) error {
	key := fmt.Sprintf("%s%s", c.prefix, trackingNumber)
	_, err := c.redis.DelCtx(ctx, key)
	return err
}

// BatchGet 批量获取缓存
func (c *TrackingCache) BatchGet(ctx context.Context, trackingNumbers []string) (map[string]*TrackingDetailDTO, error) {
	result := make(map[string]*TrackingDetailDTO)

	for _, tn := range trackingNumbers {
		detail, hit, err := c.Get(ctx, tn)
		if err != nil {
			continue // 跳过错误�?
		}
		if hit {
			result[tn] = detail
		}
	}

	return result, nil
}

// BatchSet 批量写入缓存
func (c *TrackingCache) BatchSet(ctx context.Context, details []*TrackingDetailDTO) error {
	for _, detail := range details {
		if err := c.Set(ctx, detail.TrackingNumber, detail); err != nil {
			// 单个失败不影响其�?
			continue
		}
	}
	return nil
}

// Stats 获取缓存统计信息
func (c *TrackingCache) Stats(ctx context.Context) (map[string]interface{}, error) {
	// 获取缓存 key 数量（近似值）
	keyPattern := fmt.Sprintf("%s*", c.prefix)

	// 注意：Redis Keys 命令在生产环境慎用（性能问题�?
	// 这里是示例，实际应该�?SCAN 命令
	keys, err := c.redis.KeysCtx(ctx, keyPattern)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"cache_size": len(keys),
		"prefix":     c.prefix,
		"ttl":        c.ttl.String(),
	}, nil
}
