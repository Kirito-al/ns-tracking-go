package database

import (
	"context"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/pkg/toolx"

	"github.com/zeromicro/go-zero/core/logx"
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
	// TODO: 补充GORM查询逻辑
	// SELECT * FROM tracking_logs WHERE source_tracking_number = ? ORDER BY updated_at DESC LIMIT 1

	logx.Infof("Find tracking_log by source_tracking_number: %s", toolx.MaskTrackingNumber(trackingNumber))
	return nil, nil
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
func (r *TrackingLogRepoImpl) SyncTrackingLog(
	trackingNumber string,
	trackStatus int32,
	syncedAt int64,
	receivedAt string,
	deliveredAt string,
	trackedAt string,
) error {
	// 防止全表更新：检查 trackingNumber 是否为空
	if trackingNumber == "" {
		return nil
	}

	// TODO: 补充GORM DB连接后实现
	logx.Infof("Sync tracking_logs: %s, track_status=%d", toolx.MaskTrackingNumber(trackingNumber), trackStatus)
	return nil
}