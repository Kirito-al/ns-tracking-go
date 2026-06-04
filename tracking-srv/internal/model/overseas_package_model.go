package model

import "time"

// OverseasPackageTrackingLog 海外包裹与轨迹关联表 model
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