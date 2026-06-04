// Package formatter 云途轨迹格式化器
// 对标：Ruby yun_express_track_formatter.rb
package formatter

// ===================== 状态码映射（对标 Ruby STATUS_CODES）=====================

// STATUS_CODES 云途状态码 → 文本描述映射
// 对标：yun_express_track_formatter.rb:6-19（12 项完整映射）
var STATUS_CODES = map[string]string{
	"0":    "NotFound",
	"10":   "InfoReceived",
	"20":   "InTransit",
	"30":   "AvailableForPickup",
	"40":   "DeliveryFailure",
	"50":   "Delivered",
	"60":   "Exception",
	"70":   "Expired",
	"80":   "Exception",           // 海关查验（会被 FixCustomsInspectionStatus 改写为 20）
	"90":   "Exception_Returned",
	"100":  "Exception_Cancel",
	"1001": "InTransit_Arrival",    // 非云途状态（用于兼容其他物流商）
}

// STATUS_CODES_TO_PACKAGE_STATES 状态码 → 包裹状态映射
// 对标：yun_express_track_formatter.rb:38-50（11 项完整映射）
var STATUS_CODES_TO_PACKAGE_STATES = map[string]string{
	"0":    "0",  // Undefined
	"10":   "4",  // Received / InfoReceived
	"20":   "2",  // InTransit
	"30":   "2",  // AvailableForPickup → InTransit
	"40":   "6",  // DeliveryFailure
	"50":   "3",  // Delivered
	"60":   "6",  // Exception → Delivery Failed
	"70":   "6",  // Expired → Delivery Failed
	"80":   "6",  // Exception → Delivery Failed（改写前）
	"90":   "7",  // Exception_Returned → Returned
	"100":  "5",  // Exception_Cancel → Canceled
	"1001": "2",  // InTransit_Arrival → InTransit（非云途状态）
}

// PACKAGE_STATES 包裹状态文本描述（补充映射）
// 对标：yun_express_track_formatter.rb:27-36
var PACKAGE_STATES = map[string]string{
	"0": "Undefined",
	"1": "Submitted",
	"2": "In Transit",
	"3": "Delivered",
	"4": "Received",
	"5": "Canceled",
	"6": "Delivery Failed",
	"7": "Returned",
}

// GetStatusText 获取状态码对应的文本描述
// 参数：trackingStatus（如 "20"）
// 返回：文本描述（如 "InTransit"）
func GetStatusText(trackingStatus string) string {
	if text, ok := STATUS_CODES[trackingStatus]; ok {
		return text
	}
	return "Undefined"
}

// GetPackageState 获取状态码对应的包裹状态码
// 参数：trackingStatus（如 "20"）
// 返回：包裹状态码（如 "2"）
func GetPackageState(trackingStatus string) string {
	if state, ok := STATUS_CODES_TO_PACKAGE_STATES[trackingStatus]; ok {
		return state
	}
	return "0" // 默认 Undefined
}

// GetPackageStateText 获取包裹状态的文本描述
// 参数：packageState（如 "2"）
// 返回：文本描述（如 "In Transit"）
func GetPackageStateText(packageState string) string {
	if text, ok := PACKAGE_STATES[packageState]; ok {
		return text
	}
	return "Undefined"
}