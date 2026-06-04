package define

import (
	"time"

	"github.com/google/uuid"
)

// TrackingUpsertedEvent 入库成功事件（Demo 阶段 - 10 核心字段�?
// 触发时机：Upsert tracking_details 成功�?
// 用途：日志落地、Kafka 投递（下期）、缓存更新（下期�?
type TrackingUpsertedEvent struct {
	// ========== 基础标识字段 ==========
	ID        string `json:"event_id"`        // 事件唯一 ID（UUID�?
	Type      string `json:"event_type"`      // 固定值："tracking.upserted"
	Timestamp int64  `json:"event_timestamp"` // 事件触发时间戳（Unix�?

	// ========== 运单基础信息 ==========
	TrackingNumber string `json:"tracking_number"` // 运单号（主键�?
	Status         int32  `json:"status"`          // 状态码�?-1001�?
	ServiceClass   string `json:"service_class"`   // 服务类名（YunExpressService�?

	// ========== 轨迹数据（全�?JSON�?=========
	Detail string `json:"detail"` // 轨迹原始 JSON（完整数据都在这里）

	// ========== 其他核心字段 ==========
	CountryCode string `json:"country_code"` // 目的国（US�?
	SyncedAt    int64  `json:"synced_at"`    // 同步时间�?
	Source      string `json:"source"`       // 数据来源（webhook/manual�?

	// ========== 下期扩展字段（预留）==========
	// WayBillNumber      string `json:"way_bill_number"`
	// TrackingNumber2    string `json:"tracking_number2"`
	// ProviderName       string `json:"provider_name"`
	// ChannelAlias       string `json:"channel_alias"`
	// ... 全量 30 字段版本
}

// NewTrackingUpsertedEvent 创建入库成功事件
func NewTrackingUpsertedEvent(trackingNumber string, status int32, serviceClass, detail, countryCode string, syncedAt int64, source string) *TrackingUpsertedEvent {
	return &TrackingUpsertedEvent{
		ID:             uuid.New().String(),
		Type:          "tracking.upserted",
		Timestamp:      time.Now().Unix(),
		TrackingNumber: trackingNumber,
		Status:         status,
		ServiceClass:   serviceClass,
		Detail:         detail,
		CountryCode:    countryCode,
		SyncedAt:       syncedAt,
		Source:         source,
	}
}

// EventID 实现 Event 接口
func (e *TrackingUpsertedEvent) EventID() string {
	return e.ID
}

// EventType 实现 Event 接口
func (e *TrackingUpsertedEvent) EventType() string {
	return e.Type
}

// EventTimestamp 实现 Event 接口
func (e *TrackingUpsertedEvent) EventTimestamp() int64 {
	return e.Timestamp
}
