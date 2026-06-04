package service

import "strconv"

// TrackingLogService 轨迹日志同步服务
// 对标：Ruby tracking_detail.rb:316-328 (sync_tracking_log)
type TrackingLogService struct {
	// 后续注入 TrackingLogRepo
}

// MapTrackStatus 状态码映射 → TrackStatus（对标 Ruby tracking_detail.rb:70-89）
// 参数：
//   - status：云途状态码（如 10, 20, 50）
// 返回：
//   - TrackStatus 值（0=待揽收, 1=运输中, 2=已签收, 3=异常）
//
// 映射规则（对标 Ruby tracking_detail.rb:70-89）：
//   - InfoReceived → 0 (to_receive) 或 1 (in_transit，取决于 overseas_package)
//   - InTransit / InTransit_Arrival → 1 (in_transit)
//   - Delivered / AvailableForPickup → 2 (delivered)
//   - Exception / DeliveryFailure / Expired / Exception_Returned / Exception_Cancel → 3 (track_exception)
//   - 默认 → 0 (to_receive)
func MapTrackStatus(status int32) int32 {
	statusStr := strconv.Itoa(int(status))

	// Step 1: 状态码 → 状态文本映射
	statusTxt := GetStatusText(statusStr)

	// Step 2: 状态文本映射
	// 对标：Ruby tracking_detail.rb:80-88
	switch statusTxt {
	case "InfoReceived":
		// TODO: 检查 overseas_package（后续实现）
		// 如果 overseas_package=true → return 1 (in_transit)
		// 如果 overseas_package=false → return 0 (to_receive)
		return 0 // 默认待揽收（保守策略）

	case "InTransit", "InTransit_Arrival":
		return 1 // in_transit

	case "Delivered", "AvailableForPickup":
		return 2 // delivered

	case "DeliveryFailure", "Exception", "Expired", "Exception_Returned", "Exception_Cancel":
		return 3 // track_exception

	default:
		// NotFound / Undefined → to_receive（保守策略）
		return 0
	}
}