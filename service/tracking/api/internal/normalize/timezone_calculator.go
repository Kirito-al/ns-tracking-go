package normalize

import (
	"time"
)

// CalculateProcessTimezone 计算ProcessTimezone（从时间差值推算）
// 对标：技术方案要求从process_utc_time与process_time差值推算时区
// 输入：
//   - processUTC: UTC时间字符串（如"2025-05-30T08:00:00Z"）
//   - processLocal: 当地时间字符串（如"2025-05-30T16:00:00Z"）
// 输出：
//   - ProcessTimezone字符串（如"UTC+08:00"）
func CalculateProcessTimezone(processUTC, processLocal string) string {
	// 1. 解析UTC时间
	utcTime, err := time.Parse(time.RFC3339, processUTC)
	if err != nil {
		return "UTC+00:00" // 默认UTC
	}

	// 2. 解析当地时间（假设也是RFC3339格式）
	localTime, err := time.Parse(time.RFC3339, processLocal)
	if err != nil {
		return "UTC+00:00" // 默认UTC
	}

	// 3. 计算时间差值（小时）
	duration := localTime.Sub(utcTime)
	hours := int(duration.Hours())

	// 4. 转换为UTC+XX:00格式
	if hours >= 0 {
		return formatTimezonePositive(hours)
	} else {
		return formatTimezoneNegative(hours)
	}
}

// formatTimezonePositive 格式化正时区（UTC+XX:00）
func formatTimezonePositive(hours int) string {
	return "UTC+" + formatHours(hours)
}

// formatTimezoneNegative 格式化负时区（UTC-XX:00）
func formatTimezoneNegative(hours int) string {
	return "UTC-" + formatHours(-hours)
}

// formatHours 格式化小时数（补零）
func formatHours(hours int) string {
	if hours < 10 {
		return "0" + string(rune('0'+hours)) + ":00"
	}
	return string(rune('0'+hours/10)) + string(rune('0'+hours%10)) + ":00"
}