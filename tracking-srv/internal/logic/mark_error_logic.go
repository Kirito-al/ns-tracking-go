package logic

import (
	"context"

	"tracking-srv/internal/svc"
	"tracking-srv/tracking"

	"github.com/zeromicro/go-zero/core/logx"
)

// MarkErrorLogic 标记错误业务逻辑
type MarkErrorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkErrorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkErrorLogic {
	return &MarkErrorLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkError 标记错误信息
func (l *MarkErrorLogic) MarkError(in *tracking.MarkErrorRequest) (*tracking.MarkErrorResponse, error) {
	// 1. 调用 DAO 层标记错误
	err := l.svcCtx.TrackingDAO.MarkError(l.ctx, in.TrackingNumber, in.ErrorMessage)

	if err != nil {
		return &tracking.MarkErrorResponse{
			Success: false,
			Message: "标记错误失败: " + err.Error(),
		}, nil
	}

	// 2. 返回成功响应
	return &tracking.MarkErrorResponse{
		Success: true,
		Message: "错误信息已标记",
	}, nil
}