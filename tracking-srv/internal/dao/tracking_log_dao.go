package dao

import (
	"context"

	"tracking-srv/internal/model"

	"gorm.io/gorm"
	"github.com/zeromicro/go-zero/core/logx"
)

// TrackingLogDAO tracking_logs 数据访问层（GORM 实现）
type TrackingLogDAO struct {
	db *gorm.DB
}

// NewTrackingLogDAO 创建数据访问层实例
func NewTrackingLogDAO(db *gorm.DB) *TrackingLogDAO {
	return &TrackingLogDAO{db: db}
}

// GetBySourceTrackingNumber 根据主单号查询 tracking_log 记录
func (d *TrackingLogDAO) GetBySourceTrackingNumber(ctx context.Context, trackingNumber string) (*model.TrackingLog, error) {
	var log model.TrackingLog

	err := d.db.WithContext(ctx).
		Where("source_tracking_number = ?", trackingNumber).
		Order("updated_at DESC").
		First(&log).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		logx.Errorf("GetBySourceTrackingNumber failed for %s: %v", trackingNumber, err)
		return nil, err
	}

	return &log, nil
}

// GetTrackingContext 获取渠道上下文信息
func (d *TrackingLogDAO) GetTrackingContext(ctx context.Context, trackingNumber string) (channelAlias, shippingAgent, shippingChannel string, err error) {
	log, err := d.GetBySourceTrackingNumber(ctx, trackingNumber)
	if err != nil {
		return "", "", "", err
	}
	if log == nil {
		return "", "", "", nil
	}

	return log.ChannelAlias, log.ShippingAgent, log.ShippingChannel, nil
}
