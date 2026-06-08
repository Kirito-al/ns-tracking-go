package database

import (
	"context"
	"time"

	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/repo"

	"github.com/zeromicro/go-zero/core/logx"
)

// WebhookRequestLogRepoImpl Webhook请求日志仓储实现
type WebhookRequestLogRepoImpl struct {
	db *DB
}

func NewWebhookRequestLogRepoImpl(db *DB) repo.WebhookRequestLogRepo {
	return &WebhookRequestLogRepoImpl{db: db}
}

// Create 创建日志（第一阶段：接收请求时）
func (r *WebhookRequestLogRepoImpl) Create(ctx context.Context, log *entity.WebhookRequestLog) error {
	now := time.Now().Unix()
	log.CreatedAt = now
	log.UpdatedAt = now
	log.Status = entity.WebhookStatusReceived // 默认状态：received

	err := r.db.WithContext(ctx).Create(log).Error
	if err != nil {
		logx.Errorf("WebhookRequestLogRepo.Create failed: %v", err)
		return err
	}

	logx.Infof("WebhookRequestLog created: id=%d, provider=%s, tracking_number=%s, idempotency_key=%s",
		log.Id, log.ProviderCode, log.TrackingNumber, log.IdempotencyKey)
	return nil
}

// Update 更新日志（第二阶段：处理完成后）
func (r *WebhookRequestLogRepoImpl) Update(ctx context.Context, id int64, status string, error string, response string) error {
	now := time.Now().Unix()

	err := r.db.WithContext(ctx).Model(&entity.WebhookRequestLog{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"error":      error,
			"response":   response,
			"updated_at": now,
		}).Error

	if err != nil {
		logx.Errorf("WebhookRequestLogRepo.Update failed: id=%d, err=%v", id, err)
		return err
	}

	logx.Infof("WebhookRequestLog updated: id=%d, status=%s", id, status)
	return nil
}

// FindById 根据ID查询日志
func (r *WebhookRequestLogRepoImpl) FindById(ctx context.Context, id int64) (*entity.WebhookRequestLog, error) {
	var log entity.WebhookRequestLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindByIdempotencyKey 根据幂等key查询日志（用于去重）
func (r *WebhookRequestLogRepoImpl) FindByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entity.WebhookRequestLog, error) {
	var log entity.WebhookRequestLog
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", idempotencyKey).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}