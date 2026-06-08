package entity

// WebhookRequestLog Webhook请求日志实体
// 用途：两阶段写入，记录Webhook请求的完整信息
// 技术方案参考：tms-talk/app/services/logistics/webhook_request_logs.py
type WebhookRequestLog struct {
	Id int64 `json:"id"`

	// 基础信息
	ProviderCode    string `json:"provider_code"`    // 服务商代码（如：yunexpress）
	TrackingNumber  string `json:"tracking_number"`  // 运单号（可能多条轨迹）
	IdempotencyKey  string `json:"idempotency_key"`  // 幂等key（SHA256前16字节）

	// 请求信息（两阶段写入：接收时记录）
	Headers  string `json:"headers"`  // 请求头（JSON字符串）
	RawBody  string `json:"raw_body"` // 原始请求体（完整payload）
	ClientIp string `json:"client_ip"` // 客户端IP

	// 处理状态（两阶段写入：处理完成后更新）
	Status   string `json:"status"`   // 状态：received/processing/success/failed
	Error    string `json:"error"`    // 错误信息（失败时记录）
	Response string `json:"response"` // 响应内容（成功时记录）

	// 时间戳
	CreatedAt int64 `json:"created_at"` // 创建时间（接收请求时，Unix timestamp）
	UpdatedAt int64 `json:"updated_at"` // 更新时间（处理完成时，Unix timestamp）
}

// WebhookStatus Webhook处理状态枚举
const (
	WebhookStatusReceived    = "received"    // 接收请求（第一阶段）
	WebhookStatusProcessing  = "processing"  // 处理中
	WebhookStatusSuccess     = "success"     // 处理成功
	WebhookStatusFailed      = "failed"      // 处理失败
)