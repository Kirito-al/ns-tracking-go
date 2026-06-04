package database

import (
	"context"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/pkg/toolx"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// OverseasPackageRepoImpl 海外包裹仓储实现（基础设施层）
// 对标：demo1-gozero service/tracking/rpc/internal/dao/overseas_package_dao.go
type OverseasPackageRepoImpl struct {
	db *DB
}

func NewOverseasPackageRepoImpl(db *DB) *OverseasPackageRepoImpl {
	return &OverseasPackageRepoImpl{db: db}
}

// HasOverseasPackage 检查运单是否为海外包裹
// 对标：Ruby tracking_detail.rb:75 (overseas_package_tracking_log)
// 用途：InfoReceived 状态判断 track_status（海外包裹 → in_transit, 否则 → to_receive）
// 修复：添加 ctx 参数，传入请求 context
func (r *OverseasPackageRepoImpl) HasOverseasPackage(ctx context.Context, trackingNumber string) (bool, error) {
	var count int64

	// 真实查询实现（GORM）
	err := r.db.WithContext(ctx).
		Model(&entity.OverseasPackageTrackingLog{}).
		Where("tracking_number = ?", trackingNumber).
		Count(&count).Error

	if err != nil {
		logx.Errorf("HasOverseasPackage failed for %s: %v", toolx.MaskTrackingNumber(trackingNumber), err)
		return false, err
	}

	return count > 0, nil
}

// Save 保存缓存日志（可选功能）
// 修复：添加 ctx 参数
func (r *OverseasPackageRepoImpl) Save(ctx context.Context, cacheLog *entity.TrackingCacheLog) error {
	// GORM Upsert 实现
	err := r.db.WithContext(ctx).
		Model(&entity.TrackingCacheLog{}).
		Where("tracking_number = ?", cacheLog.TrackingNumber).
		Assign(map[string]interface{}{
			"tracking_log_id": cacheLog.TrackingLogID,
			"detail":          cacheLog.Detail,
			"status":          cacheLog.Status,
			"service_class":   cacheLog.ServiceClass,
			"synced_at":       cacheLog.SyncedAt,
		}).
		FirstOrCreate(cacheLog).Error

	if err != nil {
		logx.Errorf("Save TrackingCacheLog failed: %v", err)
		return err
	}

	return nil
}

// FindByTrackingNumber 根据运单号查询缓存日志
// 修复：添加 ctx 参数
func (r *OverseasPackageRepoImpl) FindByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.TrackingCacheLog, error) {
	var cacheLog entity.TrackingCacheLog

	err := r.db.WithContext(ctx).
		Where("tracking_number = ?", trackingNumber).
		First(&cacheLog).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		logx.Errorf("FindByTrackingNumber failed for %s: %v", toolx.MaskTrackingNumber(trackingNumber), err)
		return nil, err
	}

	return &cacheLog, nil
}