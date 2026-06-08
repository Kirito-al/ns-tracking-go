package repo

import (
	"context"

	"ns-tracking-go/domain/tracking/entity"
)

// WebhookRequestLogRepo Webhook请求日志仓储接口
// 用途：两阶段写入，记录Webhook请求的完整信息
type WebhookRequestLogRepo interface {
	// Create 创建日志（第一阶段：接收请求时）
	// 记录：headers、raw_body、client_ip、status=received
	Create(ctx context.Context, log *entity.WebhookRequestLog) error

	// Update 更新日志（第二阶段：处理完成后）
	// 更新：status、error、response
	Update(ctx context.Context, id int64, status string, error string, response string) error

	// FindById 根据ID查询日志
	FindById(ctx context.Context, id int64) (*entity.WebhookRequestLog, error)

	// FindByIdempotencyKey 根据幂等key查询日志（用于去重）
	FindByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entity.WebhookRequestLog, error)
}