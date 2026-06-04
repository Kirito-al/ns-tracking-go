package define

import (
	"time"

	"github.com/google/uuid"
)

// TrackingUpsertFailedEvent 入库失败事件（Demo 阶段 - 10 核心字段）
// 触发时机：Upsert tracking_details 失败后
// 用途：日志落地、Asynq 重试任务投递
type TrackingUpsertFailedEvent struct {
	// ========== 基础标识字段 ==========
	ID        string `json:"event_id"`        // 事件唯一 ID（UUID）
	Type      string `json:"event_type"`      // 固定值："tracking.upsert_failed"
	Timestamp int64  `json:"event_timestamp"` // 事件触发时间戳（Unix）

	// ========== 失败信息 ==========
	TrackingNumber string `json:"tracking_number"` // 运单号
	ErrorMessage   string `json:"error_message"`   // 错误详情（如：数据库连接失败）
	ErrorCode      int    `json:"error_code"`      // 错误码（500=数据库错误）

	// ========== 失败场景 ==========
	FailedStage    string `json:"failed_stage"`    // 失败阶段（Upsert/SyncLog/CacheLog）
	OriginalPayload string `json:"original_payload"` // 原始 JSON Payload（用于 Asynq 重试）
	RetryCount     int    `json:"retry_count"`     // 当前重试次数（Asynq 内置管理）
	FailedAt       int64  `json:"failed_at"`       // 失败时间戳

	// ========== 下期扩展字段（预留）==========
	// MaxRetry       int    `json:"max_retry"`       // 最大重试次数（默认 3）
	// NextRetryAt    int64  `json:"next_retry_at"`   // 下次重试时间（延迟策略）
}

// NewTrackingUpsertFailedEvent 创建入库失败事件
func NewTrackingUpsertFailedEvent(trackingNumber, errorMessage string, errorCode int, failedStage, originalPayload string) *TrackingUpsertFailedEvent {
	return &TrackingUpsertFailedEvent{
		ID:             uuid.New().String(),
		Type:           "tracking.upsert_failed",
		Timestamp:      time.Now().Unix(),
		TrackingNumber: trackingNumber,
		ErrorMessage:   errorMessage,
		ErrorCode:      errorCode,
		FailedStage:    failedStage,
		OriginalPayload: originalPayload,
		RetryCount:     0, // Asynq 内置管理
		FailedAt:       time.Now().Unix(),
	}
}

// EventID 实现 Event 接口
func (e *TrackingUpsertFailedEvent) EventID() string {
	return e.ID
}

// EventType 实现 Event 接口
func (e *TrackingUpsertFailedEvent) EventType() string {
	return e.Type
}

// EventTimestamp 实现 Event 接口
func (e *TrackingUpsertFailedEvent) EventTimestamp() int64 {
	return e.Timestamp
}