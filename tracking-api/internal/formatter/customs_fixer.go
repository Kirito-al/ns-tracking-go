package formatter

// ===================== 海关查验状态改写（对标 Ruby fix_customs_inspection_status!）=====================

// FixCustomsInspectionStatus 海关查验状态改写（80 → 20）
// 对标：tracking_detail.rb:107-130（fix_customs_inspection_status! 方法）
//
// 业务规则：
// - 云途状态码 80（海关查验）在业务上等同于 20（InTransit）
// - 线上 Ruby 有独立定时任务将 80 改写为 20
// - Go 在标准化阶段直接改写，不依赖 Ruby 侧定时任务
//
// 参数：trackingStatus（如 "80"）
// 返回：改写后的状态码（如 "20"）
func FixCustomsInspectionStatus(trackingStatus string) string {
	if trackingStatus == "80" {
		return "20" // 海关查验改写为 InTransit
	}
	return trackingStatus // 其他状态不变
}

// FixCustomsInspectionStatusInEvents 批量改写事件中的海关查验状态
// 参数：events（轨迹事件列表）
// 返回：改写后的事件列表
func FixCustomsInspectionStatusInEvents(events []map[string]interface{}) []map[string]interface{} {
	for _, event := range events {
		// 提取状态码
		status := extractStatusFromEvent(event)

		// 改写海关查验状态
		if status == "80" {
			// 改写为 20（InTransit）
			event["TrackingStatus"] = "20"

			// 同步改写 PackageState（可选，Format流程会重新计算）
			// 注：根据反馈，FixCustomsInspectionPackageState 函数是多余的
			// 因为状态改写后，GetPackageState("20") 会自动返回 "2"
			// 所以不需要显式改写 PackageState
		}
	}

	return events
}

// extractStatusFromEvent 从事件中提取状态码（辅助函数）
func extractStatusFromEvent(event map[string]interface{}) string {
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

// 注：根据反馈删除 FixCustomsInspectionPackageState 空函数
// 因为状态改写后，GetPackageState("20") 会自动返回 "2"（InTransit）
// 不需要额外的 PackageState 改写逻辑