package service

// NODE_CODE_TO_TRACKING_STATUS 节点代码 → TrackingStatus 映射表
// 数据来源：三组真实 YunExpress API 数据样本
// 共23个节点代码，覆盖从揽收到签收/失败的完整流程
var NODE_CODE_TO_TRACKING_STATUS = map[string]string{
	// 揽收阶段（10: InfoReceived）
	"ORDER_CREATION":      "10", // 信息接收
	"FIRST_MILE_ARRIVE":   "10", // 到达头程仓库（已揽收）

	// 运输阶段（20: InTransit）
	"FIRST_MILE_DEPART":       "20", // 头程发出
	"TRANSIT_OUT":             "20", // 中转发出
	"PRE_INFO":                "20", // 预报信息
	"TRANSIT_IN":              "20", // 中转到达
	"DEPART_CONFIRM_OC":       "20", // 出库确认
	"ARRIVE_CONFIRM_OC":       "20", // 到达国际机场
	"ARRIVE_ORIN_CUSTOMS":     "20", // 始发国报关
	"MAIN_LINE_DEPART":        "20", // 国际航班起飞
	"MAIN_LINE_ARRIVE":        "20", // 国际航班到达
	"CUSTOMS_PROCESSING":      "20", // 目的国清关中
	"CUSTOMS_COMPLETE":        "20", // 清关完成
	"READY_FOR_PICKUP":        "20", // 待提货
	"PICKUP_CARGO_TERMINAL":   "20", // 提货完成
	"IN_TRANSIT":              "20", // 运输中
	"READY_FOR_OUTBOUND":      "20", // 待出库
	"TRANSITHUB_ARRIVE":       "20", // 交本地承运商
	"CARRIER_PICKUP":          "20", // 本地承运商揽收
	"IN_TRANSIT_CARRIER":      "20", // 本地运输中
	"DELIVERY_ATTEMPT":        "20", // 投递尝试

	// 异常阶段（40: DeliveryFailure）
	"DELIVERY_FAILURE": "40", // 投递失败

	// 签收阶段（50: Delivered）
	"DELIVERED": "50", // 已签收
}

// MapNodeCode 映射节点代码 → TrackingStatus
// 如果节点代码不在映射表中，默认返回 "20"（运输中）
func MapNodeCode(nodeCode string) string {
	if status, ok := NODE_CODE_TO_TRACKING_STATUS[nodeCode]; ok {
		return status
	}
	// 未找到节点代码，默认运输中
	return "20"
}

// PACKAGE_STATUS_TO_TRACKING_STATUS 包裹状态 → TrackingStatus 映射表
// YunExpress API 返回的 package_status (T/D/E/N/R/C/F) → 内部状态码
var PACKAGE_STATUS_TO_TRACKING_STATUS = map[string]string{
	"N": "0",  // NotFound → 未找到订单
	"F": "10", // InfoReceived → 电子预报信息接收
	"T": "20", // InTransit → 运输中
	"D": "50", // Delivered → 成功投递
	"E": "60", // Exception → 可能异常
	"R": "70", // Returned → 包裹退回
	"C": "80", // Canceled → 订单取消
}

// MapPackageStatus 映射包裹状态 → TrackingStatus
func MapPackageStatus(packageStatus string) string {
	if status, ok := PACKAGE_STATUS_TO_TRACKING_STATUS[packageStatus]; ok {
		return status
	}
	// 未找到状态码，默认运输中
	return "20"
}

// NODE_CODE_LABELS 节点代码 → 多语言标签映射表
// 用于查询响应中的 node_labels 字段（zh/en双语）
var NODE_CODE_LABELS = map[string]struct {
	Zh string
	En string
}{
	// 揽收阶段
	"ORDER_CREATION":      {Zh: "订单创建", En: "Order Creation"},
	"FIRST_MILE_ARRIVE":   {Zh: "到达头程仓库", En: "Arrived at origin facility"},
	
	// 运输阶段
	"FIRST_MILE_DEPART":       {Zh: "头程发出", En: "Departed from origin"},
	"TRANSIT_OUT":             {Zh: "中转发出", En: "Transit out"},
	"PRE_INFO":                {Zh: "预报信息", En: "Pre-alert information"},
	"TRANSIT_IN":              {Zh: "中转到达", En: "Transit in"},
	"DEPART_CONFIRM_OC":       {Zh: "出库确认", En: "Departure confirmed"},
	"ARRIVE_CONFIRM_OC":       {Zh: "到达国际机场", En: "Arrived at airport"},
	"ARRIVE_ORIN_CUSTOMS":     {Zh: "始发国报关", En: "Origin customs clearance"},
	"MAIN_LINE_DEPART":        {Zh: "国际航班起飞", En: "International flight departed"},
	"MAIN_LINE_ARRIVE":        {Zh: "国际航班到达", En: "Arrived at destination country"},
	"CUSTOMS_PROCESSING":      {Zh: "目的国清关中", En: "Customs processing"},
	"CUSTOMS_COMPLETE":        {Zh: "清关完成", En: "Customs clearance completed"},
	"READY_FOR_PICKUP":        {Zh: "待提货", En: "Ready for pickup"},
	"PICKUP_CARGO_TERMINAL":   {Zh: "提货完成", En: "Cargo picked up"},
	"IN_TRANSIT":              {Zh: "运输中", En: "In transit"},
	"READY_FOR_OUTBOUND":      {Zh: "待出库", En: "Ready for outbound"},
	"TRANSITHUB_ARRIVE":       {Zh: "交本地承运商", En: "Handed to local carrier"},
	"CARRIER_PICKUP":          {Zh: "本地承运商揽收", En: "Local carrier pickup"},
	"IN_TRANSIT_CARRIER":      {Zh: "本地运输中", En: "In transit with local carrier"},
	"DELIVERY_ATTEMPT":        {Zh: "投递尝试", En: "Delivery attempt"},
	
	// 异常阶段
	"DELIVERY_FAILURE": {Zh: "投递失败", En: "Delivery failed"},
	
	// 签收阶段
	"DELIVERED": {Zh: "已签收", En: "Delivered"},
}

// GetNodeLabels 获取节点代码的多语言标签
func GetNodeLabels(nodeCode string) (zh string, en string) {
	if labels, ok := NODE_CODE_LABELS[nodeCode]; ok {
		return labels.Zh, labels.En
	}
	// 未找到节点代码，返回默认值
	return "未知节点", "Unknown node"
}