package repo

import "ns-tracking-go/domain/tracking/entity"

// OverseasPackageRepo 海外包裹仓储接口（领域层定义）
// 用途：判断运单是否为海外包裹（影响 InfoReceived → track_status 映射）
type OverseasPackageRepo interface {
	HasOverseasPackage(trackingNumber string) (bool, error)
}

// TrackingCacheLogRepo 缓存日志仓储接口（领域层定义）
type TrackingCacheLogRepo interface {
	Save(cacheLog *entity.TrackingCacheLog) error
	FindByTrackingNumber(trackingNumber string) (*entity.TrackingCacheLog, error)
}