package service

import "strings"

// ===================== 事件过滤（对标 Ruby TrackingEventFilter）=====================

// RECEIVE_PATTERNS receive 事件识别模式（16个）
// 对标：tracking_detail.rb:17-41（RECEIVE_LOGS 常量）
// 注：当前暂不在 Go 侧实现 received_at 计算，保留在 Ruby 侧
var RECEIVE_PATTERNS = []string{
	"shipping label created",
	"shipment info received",
	"shipment information received",
	"shipping information receive",
	"shipping info received",
	"shipment processing at facility",
	"shipment arrived at facility",
	"order processed by shipper",
	"parcel information received",
	"infomation receive",
	"pre-shipment info sent to",
	"parcel data receive",
	"shipment departed from facility",
	"departed from sorting center",
	"last mile=>",
	"shipment information sent to",
}

// FilterOrderTrackingDetails 过滤重复/无效事件
// 对标：base_express_service.rb:52（TrackingEventFilter.filter_order_tracking_details）
func FilterOrderTrackingDetails(events []map[string]interface{}) []map[string]interface{} {
	if len(events) == 0 {
		return events
	}

	var filtered []map[string]interface{}

	for _, event := range events {
		// 基础过滤：只保留有效事件
		if isValidEvent(event) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// isValidEvent 判断事件是否有效
// 基础规则：必须有内容和时间
func isValidEvent(event map[string]interface{}) bool {
	// 1. 检查内容字段（content 或 ProcessContent）
	content := extractContent(event)
	if content == "" {
		return false // 无内容的事件无效
	}

	// 2. 检查时间字段（time 或 ProcessDate）
	timeStr := extractTimeStr(event)
	if timeStr == "" {
		return false // 无时间的事件无效
	}

	return true
}

// extractContent 提取事件内容（兼容大小写）
func extractContent(event map[string]interface{}) string {
	// 尝试多种字段名格式（兼容大小写）
	contentFields := []string{"content", "ProcessContent", "processContent", "ProcessContent"}

	for _, field := range contentFields {
		if content, ok := event[field]; ok {
			if str, ok := content.(string); ok {
				return strings.TrimSpace(str)
			}
		}
	}

	return ""
}

// extractTimeStr 提取时间字符串（兼容大小写）
func extractTimeStr(event map[string]interface{}) string {
	// 尝试多种字段名格式（兼容大小写）
	timeFields := []string{"time", "ProcessDate", "processDate", "ProcessDate"}

	for _, field := range timeFields {
		if timeStr, ok := event[field]; ok {
			if str, ok := timeStr.(string); ok {
				return strings.TrimSpace(str)
			}
		}
	}

	return ""
}

// IsReceiveEvent 判断是否为 receive 事件
// 对标：tracking_detail.rb:17-41（16 个模式匹配）
// 注：当前保留在 Ruby 侧实现 received_at 计算
func IsReceiveEvent(content string) bool {
	contentLower := strings.ToLower(content)

	for _, pattern := range RECEIVE_PATTERNS {
		if strings.Contains(contentLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// MarkDeliveredNode 标记 Delivered 状态节点
// 对标：yun_express_track_formatter.rb:83-85
func MarkDeliveredNode(events []map[string]interface{}) []map[string]interface{} {
	for _, event := range events {
		// 检查 TrackingStatus 是否为 "50" 或 "Delivered"
		status := extractStatus(event)
		if status == "50" || status == "Delivered" {
			event["node"] = "Delivered" // 添加 node 字段标记
		}
	}

	return events
}

// extractStatus 提取状态码
func extractStatus(event map[string]interface{}) string {
	statusFields := []string{"TrackingStatus", "trackingStatus", "status"}

	for _, field := range statusFields {
		if status, ok := event[field]; ok {
			if str, ok := status.(string); ok {
				return str
			}
		}
	}

	return ""
}

// RemoveDuplicateEvents 移除重复事件（可选功能）
// 后续迭代补充：根据时间+地点+内容判断重复
func RemoveDuplicateEvents(events []map[string]interface{}) []map[string]interface{} {
	// TODO: 实现重复事件移除逻辑
	// 当前：直接返回原列表
	return events
}