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
	
	// 5. 构建 response 包装层 + synced_at
	return &TrackingDetailDTO{
		Response: &TrackingResponseDTO{
			Item: TrackingItemDTO{
				TrackingNumber:       trackingNumber,
				WayBillNumber:        extractStringField(f.raw, "WayBillNumber"),
				TrackingStatus:       latestStatus,
				PackageState:         packageState,
				ProviderName:         extractStringField(f.raw, "ProviderName"),
				ProviderTelephone:    extractStringField(f.raw, "ProviderTelephone"),
				ProviderSite:         extractStringField(f.raw, "ProviderSite"),
				CountryCode:          extractStringField(f.raw, "CountryCode"),
				OriginCountryCode:    extractStringField(f.raw, "OriginCountryCode"),
				LastMileCarrierName:  extractStringField(f.raw, "LastMileCarrierName"),
				CarrierName:          extractStringField(f.raw, "CarrierName"),
				CheckOutTime:         extractStringField(f.raw, "CheckOutTime"),
				PickUpTime:           extractStringField(f.raw, "PickUpTime"),
				CustomerCode:         extractStringField(f.raw, "CustomerCode"),
				CustomerOrderNumber:  extractStringField(f.raw, "CustomerOrderNumber"),
				ProductCode:          extractStringField(f.raw, "ProductCode"),
				ChannelCode:          extractStringField(f.raw, "ChannelCode"),
				ActualWeight:         extractFloatField(f.raw, "ActualWeight"),
				IntervalDay:          extractFloatField(f.raw, "IntervalDay"),
				IntervalWorkDay:      extractFloatField(f.raw, "IntervalWorkDay"),
				CreatedBy:            extractStringField(f.raw, "CreatedBy"),
				POD:                  extractStringField(f.raw, "POD"),
				IsSignature:          extractBoolField(f.raw, "IsSignature"),
				SignatureUrls:        extractStringSliceField(f.raw, "SignatureUrls"),
				PodUrls:              extractStringSliceField(f.raw, "PodUrls"),
				OrderTrackingDetails: orderTrackingDetails,
			},
		},
		SyncedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// extractOrderTrackingDetails 提取轨迹明细列表（完整版）
func (f *YunExpressFormatter) extractOrderTrackingDetails() []TrackingDetailDTOItem {
	// 从 raw 中提取 OrderTrackingDetails
	if details, ok := f.raw["OrderTrackingDetails"]; ok {
		// 支持两种类型：[]interface{} 和 []map[string]interface{}
		// 类型1：[]interface{}（通用类型）
		if detailsArray, ok := details.([]interface{}); ok {
			items := make([]TrackingDetailDTOItem, len(detailsArray))
			for i, detail := range detailsArray {
				if detailMap, ok := detail.(map[string]interface{}); ok {
					items[i] = TrackingDetailDTOItem{
						ProcessDate:          extractStringField(detailMap, "ProcessDate"),
						ProcessUTCTime:       extractStringField(detailMap, "ProcessUTCTime"),
						ProcessLocation:      extractStringField(detailMap, "ProcessLocation"),
						ProcessContent:       extractStringField(detailMap, "ProcessContent"),
						ProcessCity:          extractStringField(detailMap, "ProcessCity"),
						ProcessCountry:       extractStringField(detailMap, "ProcessCountry"),
						ProcessProvince:      extractStringField(detailMap, "ProcessProvince"),
						TrackingStatus:       extractStringField(detailMap, "TrackingStatus"),
						TrackNodeCode:        extractStringField(detailMap, "TrackNodeCode"),
						TrackCodeDescription: extractStringField(detailMap, "TrackCodeDescription"),
						PodURL:               extractStringField(detailMap, "PodURL"),
						ProcessTimezone:      extractStringField(detailMap, "ProcessTimezone"),
					}
				}
			}
			return items
		}
		// 类型2：[]map[string]interface{}（直接类型）
		if detailsMapArray, ok := details.([]map[string]interface{}); ok {
			items := make([]TrackingDetailDTOItem, len(detailsMapArray))
			for i, detailMap := range detailsMapArray {
				items[i] = TrackingDetailDTOItem{
					ProcessDate:          extractStringField(detailMap, "ProcessDate"),
					ProcessUTCTime:       extractStringField(detailMap, "ProcessUTCTime"),
					ProcessLocation:      extractStringField(detailMap, "ProcessLocation"),
					ProcessContent:       extractStringField(detailMap, "ProcessContent"),
					ProcessCity:          extractStringField(detailMap, "ProcessCity"),
					ProcessCountry:       extractStringField(detailMap, "ProcessCountry"),
					ProcessProvince:      extractStringField(detailMap, "ProcessProvince"),
					TrackingStatus:       extractStringField(detailMap, "TrackingStatus"),
					TrackNodeCode:        extractStringField(detailMap, "TrackNodeCode"),
					TrackCodeDescription: extractStringField(detailMap, "TrackCodeDescription"),
					PodURL:               extractStringField(detailMap, "PodURL"),
					ProcessTimezone:      extractStringField(detailMap, "ProcessTimezone"),
				}
			}
			return items
		}
	}
	return []TrackingDetailDTOItem{}
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

// extractFloatField 提取 float64 字段（辅助函数）
func extractFloatField(raw map[string]interface{}, field string) float64 {
	if value, ok := raw[field]; ok {
		if num, ok := value.(float64); ok {
			return num
		}
	}
	return 0.0
}

// extractBoolField 提取 bool 字段（辅助函数）
func extractBoolField(raw map[string]interface{}, field string) bool {
	if value, ok := raw[field]; ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// extractStringSliceField 提取字符串数组字段（辅助函数）
func extractStringSliceField(raw map[string]interface{}, field string) []string {
	if value, ok := raw[field]; ok {
		if arr, ok := value.([]string); ok {
			return arr
		}
		if arr, ok := value.([]interface{}); ok {
			result := make([]string, len(arr))
			for i, item := range arr {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return []string{}
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
	ProviderTelephone    string                 `json:"ProviderTelephone,omitempty"`
	ProviderSite         string                 `json:"ProviderSite,omitempty"`
	CountryCode          string                 `json:"CountryCode,omitempty"`
	OriginCountryCode    string                 `json:"OriginCountryCode,omitempty"`
	LastMileCarrierName  string                 `json:"LastMileCarrierName,omitempty"`
	TrackingStatus       string                 `json:"TrackingStatus"`
	PackageState         string                 `json:"PackageState"`
	CheckOutTime         string                 `json:"CheckOutTime,omitempty"`
	PickUpTime           string                 `json:"PickUpTime,omitempty"`
	CustomerCode         string                 `json:"CustomerCode,omitempty"`
	CustomerOrderNumber  string                 `json:"CustomerOrderNumber,omitempty"`
	ProductCode          string                 `json:"ProductCode,omitempty"`
	ChannelCode          string                 `json:"ChannelCode,omitempty"`
	ActualWeight         float64                `json:"ActualWeight,omitempty"`
	IntervalDay          float64                `json:"IntervalDay,omitempty"`
	IntervalWorkDay      float64                `json:"IntervalWorkDay,omitempty"`
	CreatedBy            string                 `json:"CreatedBy,omitempty"`
	POD                  string                 `json:"POD,omitempty"`
	IsSignature          bool                   `json:"IsSignature"`
	SignatureUrls        []string               `json:"SignatureUrls"`
	PodUrls              []string               `json:"PodUrls"`
	OrderTrackingDetails []TrackingDetailDTOItem `json:"OrderTrackingDetails"`
}

// TrackingDetailDTOItem 轨迹明细项
type TrackingDetailDTOItem struct {
	ProcessDate          string `json:"ProcessDate"`
	ProcessUTCTime       string `json:"ProcessUTCTime,omitempty"`
	ProcessLocation      string `json:"ProcessLocation"`
	ProcessContent       string `json:"ProcessContent"`
	ProcessCity          string `json:"ProcessCity,omitempty"`
	ProcessCountry       string `json:"ProcessCountry,omitempty"`
	ProcessProvince      string `json:"ProcessProvince,omitempty"`
	TrackingStatus       string `json:"TrackingStatus"`
	TrackNodeCode        string `json:"TrackNodeCode,omitempty"`
	TrackCodeDescription string `json:"TrackCodeDescription,omitempty"`
	PodURL               string `json:"PodURL,omitempty"`
	ProcessTimezone      string `json:"ProcessTimezone,omitempty"`
}

// ToJSON 转换为JSON字符串（用于存入JSONB字段）
func (dto *TrackingDetailDTO) ToJSON() (string, error) {
	bytes, err := json.Marshal(dto)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}