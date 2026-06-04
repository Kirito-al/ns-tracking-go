package entity

import "time"

// TrackingDetail 运单主数据表实体
// 对标：demo1-gozero service/tracking/rpc/internal/model/models.go
type TrackingDetail struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	TrackingNumber  string `gorm:"column:tracking_number;type:varchar(100);not null;uniqueIndex"`
	TrackingLogID   int64  `gorm:"column:tracking_log_id"`

	// JSON 数据存储（JSONB 类型）
	Detail          string `gorm:"column:detail;type:jsonb;not null"`        // 当前轨迹详情
	LastDetail      string `gorm:"column:last_detail;type:jsonb"`            // 上次轨迹详情（备份）
	ReplaceDetail   string `gorm:"column:replace_detail;type:jsonb"`         // 假轨迹替换数据

	Status          int32  `gorm:"column:status;default:0"`
	ServiceClass    string `gorm:"column:service_class;type:varchar(100)"`
	SyncedAt        int64  `gorm:"column:synced_at"`
	AutoDeliveredAt int64  `gorm:"column:auto_delivered_at"`
	ErrorMessage    string `gorm:"column:error_message;type:text"`
	CreatedAt       int64  `gorm:"column:created_at"`
	UpdatedAt       int64  `gorm:"column:updated_at"`
}

// TableName 指定表名
func (TrackingDetail) TableName() string {
	return "tracking_details"
}

// TrackingLog 轨迹日志表实体
// 对标：demo1-gozero service/tracking/rpc/internal/model/models.go
type TrackingLog struct {
	ID                   int64     `gorm:"column:id;primaryKey;autoIncrement"`
	SourceTrackingNumber string    `gorm:"column:source_tracking_number;type:varchar(100);not null;index"`
	TrackingNumber       string    `gorm:"column:tracking_number;type:varchar(100)"`
	ClientOrderNumber    string    `gorm:"column:client_order_number;type:varchar(100)"`
	ChannelAlias         string    `gorm:"column:channel_alias;type:varchar(100)"`
	ShippingAgent        string    `gorm:"column:shipping_agent;type:varchar(100)"`
	ShippingChannel      string    `gorm:"column:shipping_channel;type:varchar(100)"`
	MabangLogisticsCode  string    `gorm:"column:mabang_logistics_code;type:varchar(100)"`
	PackageNumber        string    `gorm:"column:package_number;type:varchar(100)"`
	CountryCode          string    `gorm:"column:country_code;type:varchar(10)"`
	InnerNumber          string    `gorm:"column:inner_number;type:varchar(100)"`
	SourceNumber         string    `gorm:"column:source_number;type:varchar(100)"`
	Status               int64     `gorm:"column:status;default:0"`
	TrackStatus          int64     `gorm:"column:track_status;default:0"`
	SyncedAt             int64     `gorm:"column:synced_at"`
	FulfillAt            int64     `gorm:"column:fulfill_at"`
	ReceivedAt           int64     `gorm:"column:received_at"`
	DeliveredAt          int64     `gorm:"column:delivered_at"`
	TrackedAt            int64     `gorm:"column:tracked_at"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (TrackingLog) TableName() string {
	return "tracking_logs"
}

// TrackingCacheLog 缓存日志表实体
type TrackingCacheLog struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TrackingNumber string    `gorm:"column:tracking_number;type:varchar(100);not null;uniqueIndex"`
	TrackingLogID  int64     `gorm:"column:tracking_log_id"`
	Detail         string    `gorm:"column:detail;type:jsonb"`              // JSONB 类型
	Status         int32     `gorm:"column:status;default:0"`
	ServiceClass   string    `gorm:"column:service_class;type:varchar(100)"`
	SyncedAt       int64     `gorm:"column:synced_at"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 指定表名
func (TrackingCacheLog) TableName() string {
	return "tracking_cache_log"
}

// OverseasPackageTrackingLog 海外包裹与轨迹关联表实体
// 对标：Ruby overseas_package_tracking_log.rb
// 用途：判断运单是否为海外包裹（影响 track_status 映射）
type OverseasPackageTrackingLog struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TrackingNumber  string    `gorm:"column:tracking_number;type:varchar(100);index"`
	OverseasPackage int64     `gorm:"column:overseas_package"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (OverseasPackageTrackingLog) TableName() string {
	return "overseas_package_tracking_logs"
}