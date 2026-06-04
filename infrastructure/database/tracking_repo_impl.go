package database

import (
	"context"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/pkg/toolx"

	"github.com/zeromicro/go-zero/core/logx"
)

// TrackingRepoImpl 轨迹仓储实现（基础设施层）
// 对标：demo1-gozero service/tracking/rpc/internal/dao/tracking_dao.go
type TrackingRepoImpl struct {
	db *DB
}

func NewTrackingRepoImpl(db *DB) *TrackingRepoImpl {
	return &TrackingRepoImpl{db: db}
}

// Save Upsert插入或更新追踪数据（幂等+并发控制）
// 对标：demo1-gozero tracking_dao.go:28-53 (Upsert方法)
// 关键：WHERE auto_delivered_at IS NULL 防止覆盖已签收记录
func (r *TrackingRepoImpl) Save(detail *entity.TrackingDetail) error {
	ctx := context.Background()
	now := time.Now().Unix()

	// 并发控制：只有未签收的记录才更新（防止覆盖Ruby auto_sign的数据）
	// 技术方案要求：WHERE auto_delivered_at IS NULL
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO tracking_details (tracking_number, detail, status, service_class, synced_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tracking_number) DO UPDATE SET
			last_detail = tracking_details.detail,
			detail = EXCLUDED.detail,
			status = EXCLUDED.status,
			service_class = EXCLUDED.service_class,
			synced_at = EXCLUDED.synced_at,
			updated_at = EXCLUDED.updated_at
		WHERE tracking_details.auto_delivered_at IS NULL
	`, detail.TrackingNumber, detail.Detail, detail.Status, detail.ServiceClass, now, now, now).Error

	if err != nil {
		logx.Errorf("Upsert failed for %s: %v", toolx.MaskTrackingNumber(detail.TrackingNumber), err)
		return err
	}

	logx.Infof("Upsert success: %s, status=%d", toolx.MaskTrackingNumber(detail.TrackingNumber), detail.Status)
	return nil
}

// SaveWithTrackingLogId Upsert插入或更新追踪数据（带 tracking_log_id 关联）
func (r *TrackingRepoImpl) SaveWithTrackingLogId(detail *entity.TrackingDetail, trackingLogId int64) error {
	ctx := context.Background()
	now := time.Now().Unix()

	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO tracking_details (tracking_number, tracking_log_id, detail, status, service_class, synced_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (tracking_number) DO UPDATE SET
			tracking_log_id = COALESCE(tracking_details.tracking_log_id, EXCLUDED.tracking_log_id),
			last_detail = tracking_details.detail,
			detail = EXCLUDED.detail,
			status = EXCLUDED.status,
			service_class = EXCLUDED.service_class,
			synced_at = EXCLUDED.synced_at,
			updated_at = EXCLUDED.updated_at
		WHERE tracking_details.auto_delivered_at IS NULL
	`, detail.TrackingNumber, trackingLogId, detail.Detail, detail.Status, detail.ServiceClass, now, now, now).Error

	return err
}

// FindByTrackingNumber 根据运单号查询轨迹详情
func (r *TrackingRepoImpl) FindByTrackingNumber(trackingNumber string) (*entity.TrackingDetail, error) {
	ctx := context.Background()
	var detail entity.TrackingDetail

	err := r.db.WithContext(ctx).
		Where("tracking_number = ?", trackingNumber).
		First(&detail).Error

	if err != nil {
		return nil, err
	}

	return &detail, nil
}

// SyncTrackingLog 同步回写 tracking_log 状态（5个字段）
// 对标：demo1-gozero tracking_dao.go:78-132 (SyncTrackingLog方法)
func (r *TrackingRepoImpl) SyncTrackingLog(
	ctx context.Context,
	trackingNumber string,
	trackStatus int32,
	syncedAt int64,
	receivedAt string,
	deliveredAt string,
	trackedAt string,
) error {
	// 防止全表更新：检查 trackingNumber 是否为空
	if trackingNumber == "" {
		return fmt.Errorf("tracking_number cannot be empty")
	}

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

	// 更新 tracking_logs 表（5个字段同步）
	err := r.db.WithContext(ctx).Exec(`
		UPDATE tracking_logs
		SET track_status = ?, synced_at = ?, received_at = ?, delivered_at = ?, tracked_at = ?
		WHERE source_tracking_number = ?
	`, trackStatus, syncedAt, receivedAtUnix, deliveredAtUnix, trackedAtUnix, trackingNumber).Error

	if err != nil {
		logx.Errorf("SyncTrackingLog failed for %s: %v", toolx.MaskTrackingNumber(trackingNumber), err)
		return err
	}

	logx.Infof("Sync tracking_logs: %s, track_status=%d", toolx.MaskTrackingNumber(trackingNumber), trackStatus)
	return nil
}