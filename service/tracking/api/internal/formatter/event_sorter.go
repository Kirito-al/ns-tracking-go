package formatter

import (
	"sort"
	"time"
)

// ===================== 事件排序（对标 Ruby sort_by ProcessDate）=====================

// SortEventsByTime 按时间升序排序轨迹事件
// 对标：yun_express_track_formatter.rb:68（sort_by ProcessDate）
func SortEventsByTime(events []map[string]interface{}) []map[string]interface{} {
	if len(events) == 0 {
		return events
	}

	// 创建副本避免修改原数组
	sorted := make([]map[string]interface{}, len(events))
	copy(sorted, events)

	// 使用 Go 标准库排序（比冒泡排序效率高）
	sort.Slice(sorted, func(i, j int) bool {
		timeI := parseTime(sorted[i])
		timeJ := parseTime(sorted[j])
		return timeI.Before(timeJ) // 升序排序
	})

	return sorted
}

// parseTime 从事件中解析时间
// 参数：event（包含 time 或 ProcessDate 字段）
// 返回：解析后的时间（解析失败返回零值）
func parseTime(event map[string]interface{}) time.Time {
	// 尝试多个时间字段（兼容不同输入格式）
	timeFields := []string{"time", "ProcessDate", "ProcessDate"}

	for _, field := range timeFields {
		if timeStr, ok := event[field]; ok {
			if str, ok := timeStr.(string); ok {
				t := parseTimeFormats(str)
				if !t.IsZero() {
					return t
				}
			}
		}
	}

	return time.Time{} // 解析失败返回零值
}

// parseTimeFormats 尝试多种时间格式解析
// 根据反馈：需要考虑时区偏移（如 +08:00）
func parseTimeFormats(timeStr string) time.Time {
	// 时间格式列表（优先级从高到低）
	// 注意：RFC3339 和 "2006-01-02T15:04:05Z" 基本等价，保留一个即可
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339（带时区偏移，如 +08:00）
		"2006-01-02T15:04:05Z",       // UTC 时间（无偏移）
		"2006-01-02 15:04:05",        // 空格分隔（无时区）
		"2006-01-02T15:04:05",        // ISO8601（无时区）
		time.RFC3339,                  // Go 标准格式
	}

	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t
		}
	}

	return time.Time{} // 所有格式都解析失败
}

// GetLatestStatus 从排序后的事件中获取最新状态码
// 参数：sortedEvents（已排序的事件列表）
// 返回：最新状态码（如 "20"）
func GetLatestStatus(sortedEvents []map[string]interface{}) string {
	if len(sortedEvents) == 0 {
		return ""
	}

	// 最后一条事件是最新的
	lastEvent := sortedEvents[len(sortedEvents)-1]

	// 尝试多个字段名（兼容性）
	statusFields := []string{"TrackingStatus", "trackingStatus", "status"}

	for _, field := range statusFields {
		if status, ok := lastEvent[field]; ok {
			if str, ok := status.(string); ok {
				return str
			}
		}
	}

	return ""
}

// GetFirstStatus 从排序后的事件中获取第一条状态码
func GetFirstStatus(sortedEvents []map[string]interface{}) string {
	if len(sortedEvents) == 0 {
		return ""
	}

	firstEvent := sortedEvents[0]

	statusFields := []string{"TrackingStatus", "trackingStatus", "status"}

	for _, field := range statusFields {
		if status, ok := firstEvent[field]; ok {
			if str, ok := status.(string); ok {
				return str
			}
		}
	}

	return ""
}

// GetEarliestTime 获取最早时间（第一条事件）
func GetEarliestTime(sortedEvents []map[string]interface{}) time.Time {
	if len(sortedEvents) == 0 {
		return time.Time{}
	}

	return parseTime(sortedEvents[0])
}

// GetLatestTime 获取最新时间（最后一条事件）
func GetLatestTime(sortedEvents []map[string]interface{}) time.Time {
	if len(sortedEvents) == 0 {
		return time.Time{}
	}

	return parseTime(sortedEvents[len(sortedEvents)-1])
}