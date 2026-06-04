package repo

import (
	"context"
	"ns-tracking-go/domain/tracking/entity"
)

// TrackingRepo 轨迹仓储接口（领域层定义）
type TrackingRepo interface {
	Save(ctx context.Context, detail *entity.TrackingDetail) error
	FindByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.TrackingDetail, error)
}

// TrackingLogRepo 轨迹日志仓储接口（领域层定义）
type TrackingLogRepo interface {
	// 修复：接口与实现参数一致（添加 receivedAt, deliveredAt, trackedAt）
	SyncTrackingLog(ctx context.Context, trackingNumber string, trackStatus int32, syncedAt int64, receivedAt string, deliveredAt string, trackedAt string) error
}