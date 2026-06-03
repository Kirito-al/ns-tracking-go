package admin

import (
	"context"

	"tracking-api/internal/svc"
	"tracking-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// AdminHealthLogic 管理员健康检查逻辑
type AdminHealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAdminHealthLogic 创建管理员健康检查逻辑
func NewAdminHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminHealthLogic {
	return &AdminHealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminHealth 管理员健康检查
func (l *AdminHealthLogic) AdminHealth(req *types.AdminHealthRequest) (*types.Response, error) {
	health := map[string]interface{}{
		"service": "tracking-api",
		"status":  "healthy",
	}

	// 检查 Redis (Ping() returns bool in go-zero v1.10)
	if req.CheckRedis {
		if l.svcCtx.Redis != nil && l.svcCtx.Redis.Ping() {
			health["redis"] = "connected"
		} else if l.svcCtx.Redis != nil {
			health["redis"] = "disconnected"
		} else {
			health["redis"] = "not_configured"
		}
	}

	// 检查数据库连接
	if req.CheckDatabase {
		health["database"] = "not_implemented"
	}

	return &types.Response{
		Code:    200,
		Message: "success",
		Data:    health,
	}, nil
}