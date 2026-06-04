package service

import (
	"encoding/json"
	"time"
)

// YunExpressFormatter 云途轨迹格式化器
// 对标：Ruby yun_express_track_formatter.rb
type YunExpressFormatter struct {
	raw map[string]interface{} // 云途原始数据（Webhook Payload）
}

// NewYunExpressFormatter 创建云途格式化器
func NewYunExpressFormatter(raw map[string]interface{}) *YunExpressFormatter {
	return &YunExpressFormatter{raw: raw}
}

// Format 执行完整格式化流程
// 返回：格式化后的轨迹详情 DTO（包含 response 包装层 + synced_at）
func (f *YunExpressFormatter) Format() *TrackingDetailDTO {
	// 1. 提取基础信息
	// 关键：优先使用 WayBillNumber 作为 TrackingNumber（主单号优先）
	trackingNumber := extractStringField(f.raw, "WayBillNumber")
	if trackingNumber == "" {
		trackingNumber = extractStringField(f.raw, "TrackingNumber")
	}
	latestStatus := extractStringField(f.raw, "TrackingStatus")
	
	// 2. 海关查验状态改写（80 → 20）
	latestStatus = FixCustomsInspectionStatus(latestStatus)
	
	// 3. 计算 PackageState（基于改写后的状态码）
	packageState := GetPackageState(latestStatus)
	
	// 4. 构建 OrderTrackingDetails（简化版，后续补充完整逻辑）
	orderTrackingDetails := f.extractOrderTrackingDetails()
	
	// 5. 构建 response 包装层 + synced_at（关键：Ruby 缓存命中条件）
	return &TrackingDetailDTO{
		Response: &TrackingResponseDTO{
			Item: TrackingItemDTO{
				TrackingNumber:       trackingNumber,      // 使用 WayBillNumber
				WayBillNumber:        extractStringField(f.raw, "WayBillNumber"),
				TrackingStatus:       latestStatus,
				PackageState:         packageState,
				OrderTrackingDetails: orderTrackingDetails,
			},
		},
		SyncedAt: time.Now().UTC().Format(time.RFC3339), // 当前 UTC 时间（ISO8601）
	}
}

// extractOrderTrackingDetails 提取轨迹明细列表（简化版）
func (f *YunExpressFormatter) extractOrderTrackingDetails() []TrackingDetailDTOItem {
	// 从 raw 中提取 OrderTrackingDetails
	if details, ok := f.raw["OrderTrackingDetails"]; ok {
		if detailsArray, ok := details.([]interface{}); ok {
			items := make([]TrackingDetailDTOItem, len(detailsArray))
			for i, detail := range detailsArray {
				if detailMap, ok := detail.(map[string]interface{}); ok {
					items[i] = TrackingDetailDTOItem{
						ProcessDate:     extractStringField(detailMap, "ProcessDate"),
						ProcessLocation: extractStringField(detailMap, "ProcessLocation"),
						ProcessContent:  extractStringField(detailMap, "ProcessContent"),
						TrackingStatus:  extractStringField(detailMap, "TrackingStatus"),
					}
				}
			}
			return items
		}
	}
	return nil
}

// extractStringField 提取字符串字段（辅助函数）
func extractStringField(raw map[string]interface{}, field string) string {
	if value, ok := raw[field]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// TrackingDetailDTO 格式化后的轨迹详情DTO（关键：Ruby缓存命中要求）
type TrackingDetailDTO struct {
	Response *TrackingResponseDTO `json:"response"` // response 包装层（Ruby: detail['response']['Item'])
	SyncedAt string               `json:"synced_at"` // synced_at（ISO8601格式）
}

// TrackingResponseDTO response包装层
type TrackingResponseDTO struct {
	Item TrackingItemDTO `json:"Item"` // Item结构（Ruby: detail['response']['Item'])
}

// TrackingItemDTO Item结构
type TrackingItemDTO struct {
	TrackingNumber       string                 `json:"TrackingNumber"`
	WayBillNumber        string                 `json:"WayBillNumber"`
	CarrierName          string                 `json:"CarrierName,omitempty"`
	ProviderName         string                 `json:"ProviderName,omitempty"`
	TrackingStatus       string                 `json:"TrackingStatus"`
	PackageState         string                 `json:"PackageState"`
	OrderTrackingDetails []TrackingDetailDTOItem `json:"OrderTrackingDetails,omitempty"`
}

// TrackingDetailDTOItem 轨迹明细项
type TrackingDetailDTOItem struct {
	ProcessDate     string `json:"ProcessDate"`
	ProcessLocation string `json:"ProcessLocation"`
	ProcessContent  string `json:"ProcessContent"`
	TrackingStatus  string `json:"TrackingStatus"`
}

// ToJSON 转换为JSON字符串（用于存入JSONB字段）
func (dto *TrackingDetailDTO) ToJSON() string {
	bytes, _ := json.Marshal(dto)
	return string(bytes)
}