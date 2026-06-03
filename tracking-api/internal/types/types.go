package types

// ===================== HTTP 请求/响应结构（对标 tracking.api） =====================

// WebhookRequest 云途 Webhook 推送的请求结构
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

// TrackingDetail 轨迹详情节点
type TrackingDetail struct {
	ProcessDate     string `json:"processDate"`     // 扫描时间
	ProcessLocation string `json:"processLocation"` // 扫描地点
	ProcessContent  string `json:"processContent"`  // 事件描述
	TrackingStatus  string `json:"trackingStatus"`  // 状态码
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
type TrackingDetailDTO struct {
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

// TrackingDetailDTOItem 轨迹明细项
type TrackingDetailDTOItem struct {
	ProcessDate     string `json:"ProcessDate"`     // 扫描时间
	ProcessLocation string `json:"ProcessLocation"` // 扫描地点
	ProcessContent  string `json:"ProcessContent"`  // 事件描述
	TrackingStatus  string `json:"TrackingStatus"`  // 状态码
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