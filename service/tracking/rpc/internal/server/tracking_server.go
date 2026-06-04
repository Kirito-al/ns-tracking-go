package server

import (
	"context"

	"ns-tracking-go/service/tracking/rpc/internal/logic"
	"ns-tracking-go/service/tracking/rpc/internal/svc"
"ns-tracking-go/service/tracking/rpc/tracking"
)

// TrackingServer gRPC 服务实现
type TrackingServer struct {
	svcCtx *svc.ServiceContext
	tracking.UnimplementedTrackingServiceServer
}

// NewTrackingServer 创建 gRPC 服务实例
func NewTrackingServer(svcCtx *svc.ServiceContext) *TrackingServer {
	return &TrackingServer{
		svcCtx: svcCtx,
	}
}

// UpsertTracking gRPC Upsert 接口实现
func (s *TrackingServer) UpsertTracking(ctx context.Context, in *tracking.UpsertRequest) (*tracking.UpsertResponse, error) {
	l := logic.NewUpsertLogic(ctx, s.svcCtx)
	return l.Upsert(in)
}

// GetTracking gRPC GetTracking 接口实现
func (s *TrackingServer) GetTracking(ctx context.Context, in *tracking.GetTrackingRequest) (*tracking.GetTrackingResponse, error) {
	l := logic.NewGetTrackingLogic(ctx, s.svcCtx)
	return l.GetTracking(in)
}

// MarkError gRPC MarkError 接口实现
func (s *TrackingServer) MarkError(ctx context.Context, in *tracking.MarkErrorRequest) (*tracking.MarkErrorResponse, error) {
	l := logic.NewMarkErrorLogic(ctx, s.svcCtx)
	return l.MarkError(in)
}
