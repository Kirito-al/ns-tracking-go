package formatter

import (
	"time"

	"ns-tracking-go/service/tracking/api/internal/types"
)

// ===================== 云途格式化器（对标 Ruby YunExpressTrackFormatter）=====================

// YunExpressFormatter 云途轨迹格式化器
// 对标：yun_express_track_formatter.rb
type YunExpressFormatter struct {
	raw map[string]interface{} // 云途原始数据（Webhook Payload）
}

// NewYunExpressFormatter 创建云途格式化器
// 参数：raw（云途 Webhook Payload，map 格式）
func NewYunExpressFormatter(raw map[string]interface{}) *YunExpressFormatter {
	return &YunExpressFormatter{raw: raw}
}

// Format 执行完整格式化流程
// 对标：yun_express_track_formatter.rb:58-71（execute 方法）
//
// 流程：
// 1. 提取轨迹明细列表（OrderTrackingDetails）
// 2. 事件过滤（移除无效事件）
// 3. 字段映射（含 TrackNodeCode 等完整字段）
// 4. 按时间排序（升序）
// 5. 海关查验状态改写（80 → 20）
// 6. 标记 Delivered 节点
// 7. 获取最新状态码
// 8. 计算 PackageState
// 9. 构建返回结构（Item 包装层）
//
// 返回：格式化后的轨迹详情 DTO
func (f *YunExpressFormatter) Format() *types.TrackingDetailDTO {
	// 1. 提取轨迹明细列表
	rawDetails := ExtractOrderTrackingDetails(f.raw)
	if len(rawDetails) == 0 {
		// 如果没有轨迹明细，返回空结构
		return f.buildEmptyDTO()
	}

	// 2. 事件过滤（基础过滤）
	filtered := FilterOrderTrackingDetails(rawDetails)

	// 3. 字段映射（含 TrackNodeCode 等完整字段）
	mapped := MapAllEvents(filtered)

	// 4. 按时间排序（升序）
	sorted := SortEventsByTime(mapped)

	// 5. 海关查验状态改写（80 → 20）
	sorted = FixCustomsInspectionStatusInEvents(sorted)

	// 6. 标记 Delivered 节点
	sorted = MarkDeliveredNode(sorted)

	// 7. 获取最新状态码
	latestStatus := GetLatestStatus(sorted)

	// 8. 海关查验状态改写（确保最新状态也被改写）
	latestStatus = FixCustomsInspectionStatus(latestStatus)

	// 9. 计算 PackageState（基于改写后的状态码）
	packageState := GetPackageState(latestStatus)

	// 10. 构建返回结构（Item 包装层）
	return f.buildDTO(latestStatus, packageState, sorted)
}

// buildDTO 构建返回 DTO 结构（添加 response 包装层 + synced_at）
// 对标 Ruby tracking_detail.rb:282-288 (sync! 方法写入格式)
func (f *YunExpressFormatter) buildDTO(
	latestStatus string,
	packageState string,
	events []map[string]interface{},
) *types.TrackingDetailDTO {
	// 提取基础信息
	trackingNumber := ExtractTrackingNumber(f.raw)
	wayBillNumber := ExtractWayBillNumber(f.raw)
	providerName := ExtractProviderName(f.raw)

	// 提取其他字段（从原始数据）
	countryCode := extractStringField(f.raw, "CountryCode")
	originCountryCode := extractStringField(f.raw, "OriginCountryCode")
	providerSite := extractStringField(f.raw, "ProviderSite")
	providerTelephone := extractStringField(f.raw, "ProviderTelephone") // 正确拼写
	lastMileCarrierName := extractStringField(f.raw, "LastMileCarrierName")
	createdBy := extractStringField(f.raw, "CreatedBy")
	pod := extractStringField(f.raw, "POD")

	// 构建 DTO（对标 Ruby 返回结构，添加 response 包装层）
	return &types.TrackingDetailDTO{
		Response: &types.TrackingResponseDTO{
			Item: types.TrackingItemDTO{
				TrackingNumber:       trackingNumber,       // 尾程单号
				WayBillNumber:        wayBillNumber,        // 主单号
				CarrierName:          "云途",               // 固定值
				ProviderName:         providerName,         // 物流商名称
				ProvicerTelephone:    providerTelephone,    // 物流商电话（注：types.go 拼写有误）
				ProviderSite:         providerSite,         // 物流商网站
				CountryCode:          countryCode,          // 目的国
				OriginCountryCode:    originCountryCode,    // 始发国
				TrackingStatus:       latestStatus,         // 最新状态码
				PackageState:         packageState,         // 包裹状态
				IntervalDays:         nil,                  // 运输天数（未计算）
				CreatedBy:            createdBy,            // 创建时间
				POD:                  pod,                  // 签收证明
				LastMileCarrierName:  lastMileCarrierName,  // 尾程承运商
				OrderTrackingDetails: convertEventsToDTOItems(events), // 轨迹明细列表
			},
		},
		SyncedAt: time.Now().UTC().Format(time.RFC3339), // 当前 UTC 时间（ISO8601）
	}
}

// buildEmptyDTO 构建空 DTO（当没有轨迹明细时，添加 response 包装层 + synced_at）
func (f *YunExpressFormatter) buildEmptyDTO() *types.TrackingDetailDTO {
	trackingNumber := ExtractTrackingNumber(f.raw)
	wayBillNumber := ExtractWayBillNumber(f.raw)

	return &types.TrackingDetailDTO{
		Response: &types.TrackingResponseDTO{
			Item: types.TrackingItemDTO{
				TrackingNumber:       trackingNumber,
				WayBillNumber:        wayBillNumber,
				CarrierName:          "云途",
				TrackingStatus:       "0",   // 默认 NotFound
				PackageState:         "0",   // 默认 Undefined
				OrderTrackingDetails: nil,   // 空列表
			},
		},
		SyncedAt: time.Now().UTC().Format(time.RFC3339), // 当前 UTC 时间（ISO8601）
	}
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

// convertEventsToDTOItems 将 map 格式的事件转换为 DTO 格式（完整字段）
func convertEventsToDTOItems(events []map[string]interface{}) []types.TrackingDetailDTOItem {
	dtoItems := make([]types.TrackingDetailDTOItem, len(events))

	for i, event := range events {
		dtoItems[i] = types.TrackingDetailDTOItem{
			ProcessDate:          extractStringField(event, "time"),               // 映射后的字段名
			ProcessLocation:      extractStringField(event, "location"),           // 映射后的字段名
			ProcessContent:       extractStringField(event, "content"),            // 映射后的字段名
			TrackingStatus:       extractStringField(event, "TrackingStatus"),     // 透传字段
			TrackNodeCode:        extractStringField(event, "TrackNodeCode"),      // 透传字段（新增）
			TrackCodeDescription: extractStringField(event, "TrackCodeDescription"), // 透传字段（新增）
			ProcessTimezone:      extractStringField(event, "ProcessTimezone"),    // 透传字段（新增）
			ProcessCountry:       extractStringField(event, "ProcessCountry"),     // 可选字段（新增）
			ProcessProvince:      extractStringField(event, "ProcessProvince"),    // 可选字段（新增）
			ProcessCity:          extractStringField(event, "ProcessCity"),        // 可选字段（新增）
			AbnormalReasons:      event["AbnormalReasons"],                       // 异常原因（可为 null）
			Node:                 extractStringField(event, "node"),               // 节点标记（Delivered）
		}
	}

	return dtoItems
}