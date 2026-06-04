package database

import (
	"context"
	"fmt"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/pkg/toolx"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// TrackingLogRepoImpl 轨迹日志仓储实现（基础设施层）
// 对标：demo1-gozero service/tracking/rpc/internal/dao/tracking_log_dao.go
type TrackingLogRepoImpl struct {
	db *DB
}

func NewTrackingLogRepoImpl(db *DB) *TrackingLogRepoImpl {
	return &TrackingLogRepoImpl{db: db}
}

// FindBySourceTrackingNumber 根据主单号查询 tracking_log 记录
// 对标：demo1-gozero tracking_log_dao.go:24-42
func (r *TrackingLogRepoImpl) FindBySourceTrackingNumber(ctx context.Context, trackingNumber string) (*entity.TrackingLog, error) {
	// 修复：实现真实GORM查询逻辑
	var log entity.TrackingLog
	
	err := r.db.WithContext(ctx).
		Where("source_tracking_number = ?", trackingNumber).
		Order("updated_at DESC").
		First(&log).Error
	
	if err == gorm.ErrRecordNotFound {
		logx.Infof("No tracking_log found for %s", toolx.MaskTrackingNumber(trackingNumber))
		return nil, nil
	}
	
	if err != nil {
		logx.Errorf("Find tracking_log failed for %s: %v", toolx.MaskTrackingNumber(trackingNumber), err)
		return nil, err
	}
	
	logx.Infof("Found tracking_log for %s", toolx.MaskTrackingNumber(trackingNumber))
	return &log, nil
}

// GetTrackingContext 获取渠道上下文信息
// 对标：demo1-gozero tracking_log_dao.go:45-55
func (r *TrackingLogRepoImpl) GetTrackingContext(ctx context.Context, trackingNumber string) (channelAlias, shippingAgent, shippingChannel string, err error) {
	log, err := r.FindBySourceTrackingNumber(ctx, trackingNumber)
	if err != nil {
		return "", "", "", err
	}
	if log == nil {
		return "", "", "", nil
	}

	return log.ChannelAlias, log.ShippingAgent, log.ShippingChannel, nil
}

// SyncTrackingLog 同步 tracking_logs 表（5 个字段）
// 对标：Ruby tracking_detail.rb:316-328 (sync_tracking_log)
// 修复：实现真实GORM更新逻辑
func (r *TrackingLogRepoImpl) SyncTrackingLog(
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

	// GORM 更新 tracking_logs 表（5个字段同步）
	err := r.db.WithContext(ctx).
		Model(&entity.TrackingLog{}).
		Where("source_tracking_number = ?", trackingNumber).
		Updates(map[string]interface{}{
			"track_status":  trackStatus,
			"synced_at":     syncedAt,
			"received_at":   receivedAtUnix,
			"delivered_at":  deliveredAtUnix,
			"tracked_at":    trackedAtUnix,
		}).Error

	if err != nil {
		logx.Errorf("Sync tracking_logs failed for %s: %v", toolx.MaskTrackingNumber(trackingNumber), err)
		return err
	}

	logx.Infof("Sync tracking_logs: %s, track_status=%d", toolx.MaskTrackingNumber(trackingNumber), trackStatus)
	return nil
}