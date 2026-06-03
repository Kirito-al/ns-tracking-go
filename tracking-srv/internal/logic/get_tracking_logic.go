package logic

import (
	"context"

	"tracking-srv/internal/svc"
	"tracking-srv/tracking"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetTrackingLogic 查询追踪数据业务逻辑
type GetTrackingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrackingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrackingLogic {
	return &GetTrackingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTracking 查询追踪数据
func (l *GetTrackingLogic) GetTracking(in *tracking.GetTrackingRequest) (*tracking.GetTrackingResponse, error) {
	// 1. 调用 DAO 层查询数据
	detail, err := l.svcCtx.TrackingDAO.GetLatest(l.ctx, in.TrackingNumber)

	if err != nil {
		return nil, err
	}

	// 2. 数据不存在
	if detail == nil || detail.TrackingNumber == "" {
		return &tracking.GetTrackingResponse{
			TrackingNumber: in.TrackingNumber,
			Detail:         "",
			Status:         0,
		}, nil
	}

	// 3. 返回数据
	return &tracking.GetTrackingResponse{
		TrackingNumber: detail.TrackingNumber,
		Detail:         detail.Detail,
		Status:         detail.Status,
		ServiceClass:   detail.ServiceClass,
		SyncedAt:       detail.SyncedAt,
		ErrorMessage:   detail.ErrorMessage,
	}, nil
}