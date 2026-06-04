package service

// ===================== 节点代码映射（对标 Ruby NodeCodeMapper）=====================

// NodeCodeToStatus 节点代码 → TrackingStatus 映射表
// 来源：技术方案文档【技术方案】ns-tracking 云途 Webhook 化与缓存主路改造方案.md
// 包含：英文标准化代码（ORDER_CREATION等）+ 中文节点代码（YB/XD等）
// 共39个节点，按 TrackingStatus 分组

var NodeCodeToStatus = map[string]string{
	// ========== TrackingStatus = 10 (InfoReceived) ==========
	"ORDER_CREATION":      "10", // 下单
	"YB/XD":               "10", // 信息已收到（对应ORDER_CREATION）

	// ========== TrackingStatus = 20 (InTransit) ==========
	"PICKED_UP":           "20", // 揽收
	"XD/ZC":               "20", // 揽收（对应PICKED_UP）
	"XD/CS":               "20", // 揽收/发出
	"PRE_ADVICING":        "20", // 服务商电子预报信息
	"PRE_INFO":            "20", // 预上网
	"FIRST_MILE_ARRIVE":   "20", // 签入
	"FIRST_MILE_DEPART":   "20", // 筕出
	"TRANSIT_OUT":         "20", // 拨出仓
	"TRANSIT_IN":          "20", // 调拨入仓
	"DEPART_CONFIRM_OC":   "20", // 发货
	"ARRIVE_CONFIRM_OC":   "20", // 到港
	"ARRIVE_ORIN_CUSTOMS": "20", // 报关中
	"EXPORT_CUSTOMS_COMPLETE": "20", // 出口报关完成
	"AIRPORT_RELEASE":     "20", // 出口机场安检放行
	"MAIN_LINE_DEPART":    "20", // 港
	"MAIN_LINE_ARRIVE":    "20", // 到达目的地机场
	"CUSTOMS_PROCESSING":  "20", // 清关中
	"CUSTOMS_COMPLETE":    "20", // 清关完成
	"CUSTOMS_RELEASE":     "20", // 海关放行
	"READY_FOR_PICKUP":    "20", // 待提取
	"PICKUP_CARGO_TERMINAL": "20", // 货站提货
	"IN_TRANSIT":          "20", // 正常中转
	"READY_FOR_OUTBOUND":  "20", // 待发出
	"TRANSITHUB_ARRIVE":   "20", // 交货给末端派送商
	"CHANGE_WAYBILL_INFO": "20", // 跟踪号变更
	"EDD":                 "20", // 预计派送时间
	"CARRIER_PICKUP":      "20", // 干线服务商提取
	"IN_TRANSIT_CARRIER":  "20", // 承运商转运

	// ========== TrackingStatus = 30 (AvailableForPickup) ==========
	"MD/DD":               "30", // 到达目的国（真实payload确认）

	// ========== TrackingStatus = 40 (DeliveryFailure) ==========
	"DELIVERY_ATTEMPT":    "40", // 派送尝试（OMS样本确认）

	// ========== TrackingStatus = 50 (Delivered) ==========
	"DELIVERED":           "50", // 妥投
	"MD/TT":               "50", // 妥投（对应DELIVERED，真实payload确认）

	// ========== TrackingStatus = 60 (Exception) ==========
	"DELIVERY_FAILURE":    "60", // 末端派送失败
	"CUSTOMS_DELAY":       "60", // 清关延误
	"CUSTOMS_HOLD":        "60", // 海关查没
	"CUSTOMS_INSPCTION":   "60", // 海关查验
	"AIRPORT_HOLD":        "60", // 出口机场安检没收
	"AIRPORT_INSPECTION":  "60", // 出口机场安检查验
	"PACKAGE_EXCEPTION":   "60", // 异常/弃件/销毁/理赔
	"PACKAGE_LOST":        "60", // 确认丢件
	"OTHER":               "60", // 空运延误/平台异常

	// ========== TrackingStatus = 90 (Exception_Returned) ==========
	"RETURNED":            "90", // 退件
	"RETURNED_BACK":       "90", // 服务商退回中
	"RETURNED_TO_SENDER":  "90", // 退件出仓(国内)

	// ========== 默认值（未命中节点）==========
	// 默认返回 20 (InTransit)
}

// MapNodeCode 节点代码 → TrackingStatus 转换
// 输入：英文标准化代码 或 中文节点代码
// 输出：TrackingStatus 数字字符串（"10"/"20"/"30"/"40"/"50"/"60"/"90"）
// 未命中节点默认返回 "20"（InTransit）
func MapNodeCode(code string) string {
	if status, ok := NodeCodeToStatus[code]; ok {
		return status
	}
	return "20" // 默认运输中
}