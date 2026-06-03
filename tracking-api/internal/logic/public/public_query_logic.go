package public

import (
	"context"

	"tracking-api/internal/svc"
	"tracking-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// PublicQueryLogic 公开轨迹查询逻辑
type PublicQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPublicQueryLogic 创建公开轨迹查询逻辑
func NewPublicQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicQueryLogic {
	return &PublicQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PublicQuery 公开轨迹查询
func (l *PublicQueryLogic) PublicQuery(req *types.GetTrackingRequest) (resp *types.Response, err error) {
	// 参数验证
	if req.TrackingNumber == "" {
		return &types.Response{
			Code:    400,
			Message: "tracking number is required",
			Data:    "请提供运单号",
		}, nil
	}

	// TODO: 调用 gRPC 查询
	// 这里暂时返回示例数据
	return &types.Response{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"tracking_number": req.TrackingNumber,
			"status":          "20",
			"status_text":     "运输中",
			"package_state":   "2",
			"last_update":     "2025-05-20T12:00:00Z",
		},
	}, nil
}