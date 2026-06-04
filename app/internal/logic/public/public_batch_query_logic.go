package public

import (
	"context"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// PublicBatchQueryLogic 公开批量轨迹查询逻辑
type PublicBatchQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPublicBatchQueryLogic 创建公开批量轨迹查询逻辑
func NewPublicBatchQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicBatchQueryLogic {
	return &PublicBatchQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PublicBatchQuery 公开批量轨迹查询
func (l *PublicBatchQueryLogic) PublicBatchQuery(req *types.PublicBatchQueryRequest) (resp *types.Response, err error) {
	// 参数验证
	if len(req.TrackingNumbers) == 0 {
		return &types.Response{
			Code:    400,
			Message: "tracking_numbers is required",
			Data:    "请提供运单号列表",
		}, nil
	}

	// 限制查询数量
	if len(req.TrackingNumbers) > 50 {
		return &types.Response{
			Code:    400,
			Message: "maximum 50 tracking numbers allowed",
			Data:    "最多查询 50 个运单号",
		}, nil
	}

	// TODO: 调用 gRPC 批量查询
	results := make([]types.PublicQueryResult, 0, len(req.TrackingNumbers))
	for _, tn := range req.TrackingNumbers {
		results = append(results, types.PublicQueryResult{
			TrackingNumber: tn,
			Status:         "运输中",
			PackageState:   "2",
			LastUpdate:     "2025-05-20T12:00:00Z",
		})
	}

	return &types.Response{
		Code:    200,
		Message: "success",
		Data: types.PublicBatchQueryResponse{
			Results: results,
		},
	}, nil
}