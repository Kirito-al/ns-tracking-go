package repo

import (
	"context"
	"ns-tracking-go/domain/tracking/entity"
)

// OverseasPackageRepo 海外包裹仓储接口（领域层定义）
// 用途：判断运单是否为海外包裹（影响 InfoReceived → track_status 映射）
// 修复：添加 ctx 参数，与实现保持一致
type OverseasPackageRepo interface {
	HasOverseasPackage(ctx context.Context, trackingNumber string) (bool, error)
}

// TrackingCacheLogRepo 缓存日志仓储接口（领域层定义）
// 修复：添加 ctx 参数
type TrackingCacheLogRepo interface {
	Save(ctx context.Context, cacheLog *entity.TrackingCacheLog) error
	FindByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.TrackingCacheLog, error)
}