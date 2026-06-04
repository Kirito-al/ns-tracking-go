package logic

import (
	"context"
	"encoding/json"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

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

func (l *GetTrackingLogic) GetTracking(req *types.GetTrackingRequest) (resp *types.Response, err error) {
	// DDD架构：直接调用仓储层查询（不经过gRPC）
	detail, err := l.svcCtx.TrackingRepo.FindByTrackingNumber(l.ctx, req.TrackingNumber)
	if err != nil {
		l.Logger.Errorf("Query failed: %v", err)
		return &types.Response{
			Code:    500,
			Message: "查询失败",
			Data:    nil,
		}, nil
	}

	// 数据不存在
	if detail == nil {
		return &types.Response{
			Code:    404,
			Message: "运单不存在",
			Data:    nil,
		}, nil
	}

	// 解析 JSON
	var detailDTO types.TrackingDetailDTO
	if err := json.Unmarshal([]byte(detail.Detail), &detailDTO); err != nil {
		return &types.Response{
			Code:    500,
			Message: "解析失败",
			Data:    nil,
		}, nil
	}

	// 返回数据
	return &types.Response{
		Code:    200,
		Message: "success",
		Data: types.GetTrackingResponse{
			TrackingNumber: detail.TrackingNumber,
			Detail:         &detailDTO,
			Status:         detail.Status,
			ServiceClass:   detail.ServiceClass,
			SyncedAt:       detail.SyncedAt,
			ErrorMessage:   detail.ErrorMessage,
		},
	}, nil
}