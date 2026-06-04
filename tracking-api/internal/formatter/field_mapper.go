package formatter

import "strings"

// ===================== 字段映射（对标 Ruby TRACK_DETAIL_MAPPINGS）=====================

// TRACK_DETAIL_MAPPINGS 轨迹明细字段映射
// 对标：yun_express_track_formatter.rb:21-25
// 格式：{Go字段名: 云途原始字段名}
var TRACK_DETAIL_MAPPINGS = map[string]string{
	"time":     "ProcessDate",       // 事件时间
	"location": "ProcessLocation",   // 事件地点
	"content":  "ProcessContent",    // 事件内容
}

// EXTRA_FIELDS 需要透传的额外字段（完整字段列表）
// 对标：文档 5.3节 TRACK_DETAIL_MAPPINGS 要求
var EXTRA_FIELDS = []string{
	"TrackingStatus",         // 轨迹状态码（必须）
	"TrackNodeCode",          // 轨迹节点代码（如 ORDER_CREATION）
	"TrackCodeDescription",   // 轨迹节点英文描述
	"ProcessTimezone",        // 时区（如 UTC+08:00）
	"ProcessCountry",         // 轨迹发生国家（可选）
	"ProcessProvince",        // 轨迹发生省州（可选）
	"ProcessCity",            // 轨迹发生城市（可选）
	"AbnormalReasons",        // 异常原因数组（可为 null）
}

// MapEventFields 映射单个事件字段
// 参数：rawEvent（云途原始事件，map格式）
// 返回：映射后的事件（Go格式）
func MapEventFields(rawEvent map[string]interface{}) map[string]interface{} {
	mapped := make(map[string]interface{})

	// 1. 基础字段映射（time/location/content）
	// 尝试多种字段名格式（兼容大小写）
	for goField, rawField := range TRACK_DETAIL_MAPPINGS {
		// 尝试原始字段名和首字母小写版本
		fieldVariants := []string{
			rawField,                         // 原始（如 ProcessDate）
			lowercaseFirst(rawField),         // 首字母小写（如 processDate）
		}

		for _, variant := range fieldVariants {
			if value, ok := rawEvent[variant]; ok {
				mapped[goField] = value
				break
			}
		}
	}

	// 2. 透传额外字段（完整字段列表）
	// 尝试多种字段名格式
	for _, extraField := range EXTRA_FIELDS {
		fieldVariants := []string{
			extraField,                       // 原始（如 TrackingStatus）
			lowercaseFirst(extraField),       // 首字母小写（如 trackingStatus）
		}

		for _, variant := range fieldVariants {
			if value, ok := rawEvent[variant]; ok {
				mapped[extraField] = value
				break
			}
		}
	}

	return mapped
}

// lowercaseFirst 将字符串首字母小写（辅助函数）
func lowercaseFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// MapAllEvents 映射所有事件字段
// 参数：rawEvents（云途原始事件列表）
// 返回：映射后的事件列表
func MapAllEvents(rawEvents []map[string]interface{}) []map[string]interface{} {
mapped := make([]map[string]interface{}, len(rawEvents))

	for i, rawEvent := range rawEvents {
		mapped[i] = MapEventFields(rawEvent)
	}

	return mapped
}

// ExtractTrackingNumber 提取运单号（兼容大小写）
// 对标：yun_express_track_formatter.rb:60-62
func ExtractTrackingNumber(raw map[string]interface{}) string {
	// 尝试多种字段名格式（兼容大小写）
	fieldVariants := []string{"TrackingNumber", "trackingNumber"}
	
	for _, variant := range fieldVariants {
		if trackingNumber, ok := raw[variant]; ok {
			if str, ok := trackingNumber.(string); ok {
				return str
			}
		}
	}
	return ""
}

// ExtractWayBillNumber 提取运单号（主单号，兼容大小写）
func ExtractWayBillNumber(raw map[string]interface{}) string {
	// 尝试多种字段名格式（兼容大小写）
	fieldVariants := []string{"WayBillNumber", "wayBillNumber"}
	
	for _, variant := range fieldVariants {
		if wayBillNumber, ok := raw[variant]; ok {
			if str, ok := wayBillNumber.(string); ok {
				return str
			}
		}
	}
	return ""
}

// ExtractProviderName 提取物流商名称（兼容大小写）
func ExtractProviderName(raw map[string]interface{}) string {
	// 尝试多种字段名格式（兼容大小写）
	fieldVariants := []string{"ProviderName", "providerName"}
	
	for _, variant := range fieldVariants {
		if providerName, ok := raw[variant]; ok {
			if str, ok := providerName.(string); ok {
				return str
			}
		}
	}
	return ""
}

// ExtractOrderTrackingDetails 提取轨迹明细列表
func ExtractOrderTrackingDetails(raw map[string]interface{}) []map[string]interface{} {
	// 尝试多个字段名（兼容大小写）
	fieldNames := []string{"OrderTrackingDetails", "orderTrackingDetails"}

	for _, fieldName := range fieldNames {
		if details, ok := raw[fieldName]; ok {
			if arr, ok := details.([]map[string]interface{}); ok {
				return arr
			}
			// 备选：尝试解析为数组
			if arr, ok := details.([]interface{}); ok {
				result := make([]map[string]interface{}, len(arr))
				for i, item := range arr {
					if m, ok := item.(map[string]interface{}); ok {
						result[i] = m
					}
				}
				return result
			}
		}
	}
	return nil
}