package logic

import (
	"context"
	"encoding/json"

	"ns-tracking-go/service/tracking/api/internal/svc"
	"ns-tracking-go/service/tracking/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"ns-tracking-go/service/tracking/rpc/tracking"
)

// GetTrackingLogic 查询追踪数据业务逻辑
type GetTrackingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTrackingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrackingLogic {
	return &GetTrackingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetTracking 查询追踪数据（调用 gRPC Client）
func (l *GetTrackingLogic) GetTracking(req *types.GetTrackingRequest) (resp *types.Response, err error) {
	// 1. 调用 gRPC Client 查询数据
	grpcReq := &tracking.GetTrackingRequest{
		TrackingNumber: req.TrackingNumber,
	}

	grpcRes, err := l.svcCtx.TrackingRpc.GetTracking(l.ctx, grpcReq)
	if err != nil {
		return &types.Response{
			Code:    500,
			Message: "查询失败: " + err.Error(),
			Data:    nil,
		}, nil
	}

	// 2. 数据不存在
	if grpcRes.Detail == "" {
		return &types.Response{
			Code:    404,
			Message: "运单不存在",
			Data:    nil,
		}, nil
	}

	// 3. 解析 JSON
	var detail types.TrackingDetailDTO
	if err := json.Unmarshal([]byte(grpcRes.Detail), &detail); err != nil {
		return &types.Response{
			Code:    500,
			Message: "解析失败: " + err.Error(),
			Data:    nil,
		}, nil
	}

	// 4. 返回数据
	return &types.Response{
		Code:    200,
		Message: "success",
		Data: types.GetTrackingResponse{
			TrackingNumber: grpcRes.TrackingNumber,
			Detail:         &detail,
			Status:         grpcRes.Status,
			ServiceClass:   grpcRes.ServiceClass,
			SyncedAt:       grpcRes.SyncedAt,
			ErrorMessage:   grpcRes.ErrorMessage,
		},
	}, nil
}