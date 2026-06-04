// Package formatter 云途轨迹格式化器
package formatter

// ===================== TrackStatus 映射（对标 Ruby tracking_detail.rb:70-89）=====================

// TrackStatus 常量定义（对标 Ruby tracking_status 枚举）
// 用途：tracking_logs.track_status 字段值
const (
	TrackStatusToReceive      int32 = 0 // 待揽收
	TrackStatusInTransit      int32 = 1 // 运输中
	TrackStatusDelivered      int32 = 2 // 已签收
	TrackStatusException      int32 = 3 // 异常
)

// MapTrackStatus 状态码映射 → TrackStatus（对标 Ruby tracking_detail.rb:70-89）
// 参数：
//   - trackingStatus：云途状态码（如 "10", "20", "50"）
//   - hasOverseasPackage：是否有海外包裹记录（InfoReceived 特殊判断）
//   - hasError：是否有错误信息（error_message 非空）
// 返回：
//   - TrackStatus 值（0=待揽收, 1=运输中, 2=已签收, 3=异常）
//
// 映射规则（对标 Ruby tracking_detail.rb:70-89）：
//   - error_message 存在 + 不是 Delivered → :track_exception (3)
//   - InfoReceived + overseas_package → :in_transit (1)
//   - InfoReceived + NO overseas_package → :to_receive (0)
//   - InTransit / InTransit_Arrival → :in_transit (1)
//   - Delivered / AvailableForPickup → :delivered (2)
//   - Exception / DeliveryFailure / Expired / Exception_Returned / Exception_Cancel → :track_exception (3)
//   - 默认 → :to_receive (0)
func MapTrackStatus(trackingStatus string, hasOverseasPackage bool, hasError bool) int32 {
	// Step 1: 获取状态文本描述
	statusTxt := GetStatusText(trackingStatus)

	// Step 2: 错误信息优先判断
	// 对标：Ruby tracking_detail.rb:71
	if hasError && statusTxt != "Delivered" {
		return TrackStatusException
	}

	// Step 3: 状态文本映射
	// 对标：Ruby tracking_detail.rb:73-88
	switch statusTxt {
	case "InfoReceived":
		// 特殊判断：海外包裹 → 运输中，否则 → 待揽收
		// 对标：Ruby tracking_detail.rb:75-79
		if hasOverseasPackage {
			return TrackStatusInTransit
		}
		return TrackStatusToReceive

	case "InTransit", "InTransit_Arrival":
		return TrackStatusInTransit

	case "Delivered", "AvailableForPickup":
		return TrackStatusDelivered

	case "DeliveryFailure", "Exception", "Expired", "Exception_Returned", "Exception_Cancel":
		return TrackStatusException

	default:
		// 默认：待揽收（保守策略）
		return TrackStatusToReceive
	}
}