package types

// ===================== 云途 Webhook 推送（官方格式） =====================

// TrackingDetail 轨迹明细项（云途 Webhook 推送的单条轨迹事件）
// 对标 Ruby tracking_detail.rb 中的 OrderTrackingDetails 项
type TrackingDetail struct {
	ProcessDate     string `json:"processDate"`     // 扫描时间（ISO8601）
	ProcessLocation string `json:"processLocation"` // 扫描地点
	ProcessContent  string `json:"processContent"`  // 事件描述
	TrackingStatus  string `json:"trackingStatus"`  // 该节点状态码
}

// WebhookRequest 云途 Webhook 推送的请求结构（官方格式，camelCase）
type WebhookRequest struct {
	TrackingNumber       string          `json:"trackingNumber"`       // 云途运单号（主单号）
	WayBillNumber        string          `json:"wayBillNumber"`        // 运单号
	TrackingStatus       string          `json:"trackingStatus"`       // 当前最新状态码
	PackageState         string          `json:"packageState"`         // 包裹状态
	OrderTrackingDetails []TrackingDetail `json:"orderTrackingDetails"` // 轨迹明细数组
	ProviderName         string          `json:"providerName"`         // 物流服务商名称
	ProviderSite         string          `json:"providerSite"`         // 物流商查询网站
	ProvicerTelephone    string          `json:"provicerTelephone"`    // 物流商客服电话
	CountryCode          string          `json:"countryCode"`          // 目的国代码
	OriginCountryCode    string          `json:"originCountryCode"`    // 始发国代码
	TrackingNumber2      string          `json:"trackingNumber2"`      // 尾程单号
	LastMileCarrierName  string          `json:"lastMileCarrierName"`  // 尾程承运商
	CreatedBy            string          `json:"createdBy"`            // 运单创建时间
	POD                  string          `json:"pod"`                  // 签收证明
}

// ===================== TIS Push 推送（真实业务格式） =====================

// TisPushRequest TIS Push Data 格式（snake_case，带 data 包装层）
// 对标：真实业务中的 tisPushData 格式
type TisPushRequest struct {
	Data     *TisPushData `json:"data,omitempty"`      // 数据主体
	DataCode string       `json:"data_code,omitempty"` // 数据标识 "tisPushData"
}

// TisPushData TIS Push 数据主体
type TisPushData struct {
	TrackingNumber                string              `json:"tracking_number"`                // 轨迹单号
	WaybillNumber                 string              `json:"waybill_number"`                 // 运单号
	PackageStatus                 string              `json:"package_status"`                 // 包裹状态 (T=运输中)
	SignatureUrls                 []string            `json:"SignatureUrls"`                  // 签名 URL 数组
	IsSignature                   bool                `json:"IsSignature"`                    // 是否已签名
	IntervalWorkDay               float64             `json:"interval_work_day"`              // 工作日天数
	ProductCode                   string              `json:"product_code"`                   // 产品代码
	CustomerOrderNumber           string              `json:"customer_order_number"`          // 客户订单号
	ProductName                   string              `json:"product_name"`                   // 产品名称
	DestinationCode               string              `json:"destination_code"`               // 目的地代码
	IntervalDay                   float64             `json:"interval_day"`                   // 总天数
	OriginCode                    string              `json:"origin_code"`                    // 始发地代码
	PodUrls                       []string            `json:"pod_urls"`                       // POD URL 数组
	CustomerCode                  string              `json:"customer_code"`                  // 客户代码
	ChannelCode                   string              `json:"channel_code"`                   // 渠道代码
	CheckInTime                   string              `json:"check_in_time"`                  // 入库时间
	CheckOutTime                  string              `json:"check_out_time"`                 // 出库时间
	PickUpTime                    string              `json:"pick_up_time"`                   // 揽收时间
	LastMileSite                  string              `json:"last_mile_site"`                 // 尾程查询网址
	EstimatedDeliveryToDateZone   string              `json:"EstimatedDeliveryToDateZone"`    // 预计送达时间（结束）
	EstimatedDeliveryFromDateZone string              `json:"EstimatedDeliveryFromDateZone"`  // 预计送达时间（开始）
	PostalCode                    string              `json:"postal_code"`                    // 邮编
	ActualWeight                  float64             `json:"actual_weight"`                  // 实际重量
	LastMileName                  string              `json:"last_mile_name"`                 // 尾程承运商名称
	PhoneNumber                   string              `json:"phone_number"`                   // 联系电话
	TrackEvents                   []TisTrackEvent     `json:"track_events"`                   // 轨迹事件列表
}

// TisTrackEvent TIS 轨迹事件
type TisTrackEvent struct {
	TrackNodeDescription string `json:"track_node_description"` // 轨迹节点描述
	ProcessUTCTime       string `json:"process_utc_time"`       // UTC 时间
	TrackNodeCode        string `json:"track_node_code"`        // 轨迹节点代码
	ProcessTime          string `json:"process_time"`           // 当地时间
	ProcessCity          string `json:"process_city"`           // 城市
	ProcessCountry       string `json:"process_country"`        // 国家
	ProcessLocation      string `json:"process_location"`       // 地点
	ProcessProvince      string `json:"process_province"`       // 省份
	PodURL               string `json:"pod_url"`                // POD 链接
}

// WebhookResponse Webhook 响应结构
type WebhookResponse struct {
	Success        bool   `json:"success"`        // 处理是否成功
	TrackingNumber string `json:"trackingNumber"` // 确认处理的运单号
	Status         string `json:"status"`         // 确认收到的状态码
	PackageState   string `json:"packageState"`   // 确认收到的包裹状态
}

// GetTrackingRequest 查询追踪数据请求
type GetTrackingRequest struct {
	TrackingNumber string `path:"number"` // 运单号
}

// GetTrackingResponse 查询追踪数据响应
type GetTrackingResponse struct {
	TrackingNumber string              `json:"trackingNumber"` // 运单号
	Detail         *TrackingDetailDTO  `json:"detail"`         // 轨迹详情
	Status         int32               `json:"status"`         // 状态码
	ServiceClass   string              `json:"serviceClass"`   // 服务类名
	SyncedAt       int64               `json:"syncedAt"`       // 同步时间戳
	ErrorMessage   string              `json:"errorMessage"`   // 错误信息
}

// TrackingDetailDTO 轨迹详情 DTO（格式化后的数据）
// 对标 Ruby tracking_detail.detail JSONB 格式
// 结构：{response: {Item: {...}}, synced_at: "..."}
type TrackingDetailDTO struct {
	Response  *TrackingResponseDTO `json:"response"`  // response 包装层
	SyncedAt  string               `json:"synced_at"` // 同步时间（ISO8601 格式）
}

// TrackingResponseDTO response 包装层结构
type TrackingResponseDTO struct {
	Item TrackingItemDTO `json:"Item"` // Item 字段
}

// TrackingItemDTO Item 字段结构
type TrackingItemDTO struct {
	TrackingNumber       string                   `json:"TrackingNumber"`       // 尾程单号
	WayBillNumber        string                   `json:"WayBillNumber"`        // 主单号
	CarrierName          string                   `json:"CarrierName"`          // 承运商名称
	ProviderName         string                   `json:"ProviderName"`         // 物流商名称
	ProvicerTelephone    string                   `json:"ProvicerTelephone"`    // 物流商电话
	ProviderSite         string                   `json:"ProviderSite"`         // 物流商网站
	CountryCode          string                   `json:"CountryCode"`          // 目的国
	OriginCountryCode    string                   `json:"OriginCountryCode"`    // 始发国
	TrackingStatus       string                   `json:"TrackingStatus"`       // 状态码
	PackageState         string                   `json:"PackageState"`         // 包裹状态
	IntervalDays         *int32                   `json:"IntervalDays"`         // 运输天数
	CreatedBy            string                   `json:"CreatedBy"`            // 创建时间
	POD                  string                   `json:"POD"`                  // 签收证明
	LastMileCarrierName  string                   `json:"LastMileCarrierName"`  // 尾程承运商
	OrderTrackingDetails []TrackingDetailDTOItem  `json:"OrderTrackingDetails"` // 轨迹明细列表
}

// TrackingDetailDTOItem 轨迹明细项（完整字段列表）
// 对标：文档 5.3节 TRACK_DETAIL_MAPPINGS 要求
type TrackingDetailDTOItem struct {
	ProcessDate          string      `json:"ProcessDate"`          // 扫描时间（必须）
	ProcessLocation      string      `json:"ProcessLocation"`      // 扫描地点（必须）
	ProcessContent       string      `json:"ProcessContent"`       // 事件描述（必须）
	TrackingStatus       string      `json:"TrackingStatus"`       // 状态码（必须）
	TrackNodeCode        string      `json:"TrackNodeCode"`        // 轨迹节点代码（如 ORDER_CREATION）
	TrackCodeDescription string      `json:"TrackCodeDescription"` // 轨迹节点英文描述
	ProcessTimezone      string      `json:"ProcessTimezone"`      // 时区（如 UTC+08:00）
	ProcessCountry       string      `json:"ProcessCountry"`       // 轨迹发生国家（可选）
	ProcessProvince      string      `json:"ProcessProvince"`      // 轨迹发生省州（可选）
	ProcessCity          string      `json:"ProcessCity"`          // 轨迹发生城市（可选）
	AbnormalReasons      interface{} `json:"AbnormalReasons"`      // 异常原因数组（可为 null）
	Node                 string      `json:"Node"`                 // 节点标记（如 Delivered）
}

// ===================== ns-admin 管理后台类型 =====================

// AdminBatchQueryRequest 管理员批量查询请求
type AdminBatchQueryRequest struct {
	TrackingNumbers []string `json:"tracking_numbers"`
	IncludeDetails  bool     `json:"include_details"`
	Limit           int      `json:"limit"`
}

// AdminHealthRequest 管理员健康检查请求
type AdminHealthRequest struct {
	CheckRedis    bool `json:"check_redis"`
	CheckDatabase bool `json:"check_database"`
}

// ===================== ns-tracking 公开 API 类型 =====================

// PublicBatchQueryRequest 公开批量查询请求
type PublicBatchQueryRequest struct {
	TrackingNumbers []string `json:"tracking_numbers"`
}

// PublicBatchQueryResponse 公开批量查询响应
type PublicBatchQueryResponse struct {
	Results []PublicQueryResult `json:"results"`
}

// PublicQueryResult 单个查询结果
type PublicQueryResult struct {
	TrackingNumber string `json:"tracking_number"`
	Status         string `json:"status"`
	PackageState   string `json:"package_state"`
	LastUpdate     string `json:"last_update"`
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 响应码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}