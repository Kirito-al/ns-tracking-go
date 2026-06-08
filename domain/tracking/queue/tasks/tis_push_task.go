package tasks

import (
	"encoding/json"
)

// TaskTisPushAsync 异步处理 TIS Push 的任务类型
// 用途：快返回模式下的后台异步处理
const TaskTisPushAsync = "task:tis_push_async"

// TisPushAsyncPayload 异步任务 Payload
// 用途：传递完整的 TIS Push 数据给后台异步处理
type TisPushAsyncPayload struct {
	// 原始 TIS Push 数据（完整）
	TisData json.RawMessage `json:"tis_data"`

	// 唯一事件列表（幂等去重后的结果）
	UniqueEventIndices []int `json:"unique_event_indices"` // 事件索引列表

	// Webhook日志ID（用于第二阶段更新）
	WebhookLogId int64 `json:"webhook_log_id"`

	// 提取的关键信息（避免重复解析）
	WaybillNumber string `json:"waybill_number"`
	ProviderCode  string `json:"provider_code"` // 默认 yunexpress
}

// NewTisPushAsyncPayload 创建异步任务 Payload
func NewTisPushAsyncPayload(tisData json.RawMessage, uniqueIndices []int, logId int64, waybill string) *TisPushAsyncPayload {
	return &TisPushAsyncPayload{
		TisData:            tisData,
		UniqueEventIndices: uniqueIndices,
		WebhookLogId:       logId,
		WaybillNumber:      waybill,
		ProviderCode:       "yunexpress",
	}
}