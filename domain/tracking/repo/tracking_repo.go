package repo

import "ns-tracking-go/domain/tracking/entity"

// TrackingRepo 轨迹仓储接口（领域层定义）
type TrackingRepo interface {
	Save(detail *entity.TrackingDetail) error
	FindByTrackingNumber(trackingNumber string) (*entity.TrackingDetail, error)
}

// TrackingLogRepo 轨迹日志仓储接口（领域层定义）
type TrackingLogRepo interface {
	SyncTrackingLog(trackingNumber string, trackStatus int32, syncedAt int64) error
}