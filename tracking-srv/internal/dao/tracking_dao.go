package dao

import (
	"context"
	"time"

	"tracking-srv/internal/model"

	"gorm.io/gorm"
	"github.com/zeromicro/go-zero/core/logx"
)

// TrackingDAO 追踪数据访问层（GORM 实现）
type TrackingDAO struct {
	db *gorm.DB
}

// NewTrackingDAO 创建数据访问层实例
func NewTrackingDAO(db *gorm.DB) *TrackingDAO {
	return &TrackingDAO{db: db}
}

// Upsert 插入或更新追踪数据（幂等+并发控制）- PostgreSQL ON CONFLICT
// 对标：Ruby tracking_detail.rb并发控制逻辑
// 关键：WHERE auto_delivered_at IS NULL 防止覆盖已签收记录
func (d *TrackingDAO) Upsert(ctx context.Context, trackingNumber, detail string, status int32, serviceClass string) error {
	now := time.Now().Unix()

	// 并发控制：只有未签收的记录才更新（防止覆盖Ruby auto_sign的数据）
	// 技术方案要求：WHERE auto_delivered_at IS NULL
	err := d.db.WithContext(ctx).Exec(`
		INSERT INTO tracking_details (tracking_number, detail, status, service_class, synced_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (tracking_number) DO UPDATE SET
			last_detail = tracking_details.detail,
			detail = EXCLUDED.detail,
			status = EXCLUDED.status,
			service_class = EXCLUDED.service_class,
			synced_at = EXCLUDED.synced_at,
			updated_at = EXCLUDED.updated_at
		WHERE tracking_details.auto_delivered_at IS NULL
	`, trackingNumber, detail, status, serviceClass, now, now).Error

	if err != nil {
		logx.Errorf("Upsert failed for %s: %v", trackingNumber, err)
		return err
	}

	logx.Infof("Upsert success: %s, status=%d", trackingNumber, status)
	return nil
}

// UpsertWithTrackingLogId 插入或更新追踪数据（带 tracking_log_id 关联+并发控制）
func (d *TrackingDAO) UpsertWithTrackingLogId(ctx context.Context, trackingNumber, detail string, status int32, serviceClass string, trackingLogId int64) error {
	now := time.Now().Unix()

	// 并发控制：只有未签收的记录才更新
	err := d.db.WithContext(ctx).Exec(`
		INSERT INTO tracking_details (tracking_number, tracking_log_id, detail, status, service_class, synced_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (tracking_number) DO UPDATE SET
			tracking_log_id = COALESCE(tracking_details.tracking_log_id, EXCLUDED.tracking_log_id),
			last_detail = tracking_details.detail,
			detail = EXCLUDED.detail,
			status = EXCLUDED.status,
			service_class = EXCLUDED.service_class,
			synced_at = EXCLUDED.synced_at,
			updated_at = EXCLUDED.updated_at
		WHERE tracking_details.auto_delivered_at IS NULL
	`, trackingNumber, trackingLogId, detail, status, serviceClass, now, now).Error

	return err
}

// SyncTrackingLog 同步回写 tracking_log 状态（5个字段同步）
// 字段：track_status, synced_at, received_at, delivered_at, tracked_at
func (d *TrackingDAO) SyncTrackingLog(ctx context.Context, trackingNumber string, trackStatus int32, syncedAt int64, receivedAt, deliveredAt, trackedAt string) error {
	// 转换ISO8601字符串为Unix时间戳（如果不为空）
	var receivedAtUnix, deliveredAtUnix, trackedAtUnix int64
	if receivedAt != "" {
		t, err := time.Parse(time.RFC3339, receivedAt)
		if err == nil {
			receivedAtUnix = t.Unix()
		}
	}
	if deliveredAt != "" {
		t, err := time.Parse(time.RFC3339, deliveredAt)
		if err == nil {
			deliveredAtUnix = t.Unix()
		}
	}
	if trackedAt != "" {
		t, err := time.Parse(time.RFC3339, trackedAt)
		if err == nil {
			trackedAtUnix = t.Unix()
		}
	}

	// 更新5个字段
	result := d.db.WithContext(ctx).
		Model(&model.TrackingLog{}).
		Where("source_tracking_number = ?", trackingNumber).
		Updates(map[string]interface{}{
			"track_status": trackStatus,
			"synced_at":    syncedAt,
			"received_at":  receivedAtUnix,
			"delivered_at": deliveredAtUnix,
			"tracked_at":   trackedAtUnix,
		})

	err := result.Error
	if err != nil {
		logx.Errorf("SyncTrackingLog failed for %s: %v", trackingNumber, err)
		return err
	}

	logx.Infof("SyncTrackingLog success: %s, affected_rows=%d, received=%d, delivered=%d, tracked=%d",
		trackingNumber, result.RowsAffected, receivedAtUnix, deliveredAtUnix, trackedAtUnix)
	return nil
}

// UpsertTrackingCacheLog 写入 tracking_cache_logs 表
func (d *TrackingDAO) UpsertTrackingCacheLog(ctx context.Context, trackingNumber string, serviceClass string, trackStatus int32) error {
	now := time.Now().Unix()

	err := d.db.WithContext(ctx).Exec(`
		INSERT INTO tracking_cache_logs (tracking_number, service_class, track_status, synced_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (tracking_number) DO UPDATE SET
			service_class = EXCLUDED.service_class,
			track_status = EXCLUDED.track_status,
			synced_at = EXCLUDED.synced_at,
			updated_at = EXCLUDED.updated_at
	`, trackingNumber, serviceClass, trackStatus, now, now).Error

	return err
}

// GetLatest 获取最新追踪数据
func (d *TrackingDAO) GetLatest(ctx context.Context, trackingNumber string) (*model.TrackingDetail, error) {
	var detail model.TrackingDetail

	err := d.db.WithContext(ctx).
		Where("tracking_number = ?", trackingNumber).
		Order("updated_at DESC").
		First(&detail).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		logx.Errorf("GetLatest failed for %s: %v", trackingNumber, err)
		return nil, err
	}

	return &detail, nil
}

// MarkError 标记错误信息
func (d *TrackingDAO) MarkError(ctx context.Context, trackingNumber, errorMessage string) error {
	now := time.Now().Unix()

	err := d.db.WithContext(ctx).Exec(`
		INSERT INTO tracking_details (tracking_number, error_message, status, service_class, synced_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (tracking_number) DO UPDATE SET
			error_message = EXCLUDED.error_message,
			service_class = EXCLUDED.service_class,
			synced_at = EXCLUDED.synced_at,
			updated_at = EXCLUDED.updated_at
	`, trackingNumber, errorMessage, 0, "YunExpressService", now, now).Error

	return err
}
