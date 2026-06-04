package dao

import (
	"context"

	"ns-tracking-go/service/tracking/rpc/internal/model"
	"ns-tracking-go/service/tracking/rpc/internal/utils"

	"gorm.io/gorm"
	"github.com/zeromicro/go-zero/core/logx"
)

// OverseasPackageDAO overseas_package_tracking_logs 数据访问层
type OverseasPackageDAO struct {
	db *gorm.DB
}

// NewOverseasPackageDAO 创建数据访问层实例
func NewOverseasPackageDAO(db *gorm.DB) *OverseasPackageDAO {
	return &OverseasPackageDAO{db: db}
}

// HasOverseasPackage 检查运单是否为海外包裹
// 对标：Ruby tracking_detail.rb:75 (overseas_package_tracking_log)
// 用途：InfoReceived 状态判断 track_status（海外包裹 → in_transit, 否则 → to_receive）
func (d *OverseasPackageDAO) HasOverseasPackage(ctx context.Context, trackingNumber string) (bool, error) {
	var count int64

	err := d.db.WithContext(ctx).
		Model(&model.OverseasPackageTrackingLog{}).
		Where("tracking_number = ?", trackingNumber).
		Count(&count).Error

	if err != nil {
		logx.Errorf("HasOverseasPackage failed for %s: %v", utils.MaskTrackingNumber(trackingNumber), err)
		return false, err
	}

	return count > 0, nil
}