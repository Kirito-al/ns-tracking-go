package admin

import (
	"context"

	"tracking-api/internal/svc"
	"tracking-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// AdminGetTrackingLogic 管理员查询轨迹详情逻辑
type AdminGetTrackingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAdminGetTrackingLogic 创建管理员查询轨迹详情逻辑
func NewAdminGetTrackingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetTrackingLogic {
	return &AdminGetTrackingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminGetTracking 管理员查询轨迹详情
func (l *AdminGetTrackingLogic) AdminGetTracking(req *types.GetTrackingRequest) (resp *types.Response, err error) {
	// 参数验证
	if req.TrackingNumber == "" {
		return &types.Response{
			Code:    400,
			Message: "tracking number is required",
			Data:    "请提供运单号",
		}, nil
	}

	// TODO: 调用 gRPC 查询详情
	// 这里暂时返回示例数据
	return &types.Response{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"tracking_number": req.TrackingNumber,
			"status":          "20",
			"status_text":     "InTransit",
			"package_state":   "2",
			"details": []map[string]string{
				{
					"ProcessDate":     "2025-05-20T08:00:00Z",
					"ProcessLocation": "深圳",
					"ProcessContent":  "快件电子信息已收到",
					"TrackingStatus":  "10",
				},
				{
					"ProcessDate":     "2025-05-20T12:00:00Z",
					"ProcessLocation": "广州",
					"ProcessContent":  "快件已发出",
					"TrackingStatus":  "20",
				},
			},
		},
	}, nil
}