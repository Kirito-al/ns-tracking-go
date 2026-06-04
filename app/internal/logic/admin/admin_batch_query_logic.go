package admin

import (
	"context"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// AdminBatchQueryLogic 管理员批量查询轨迹逻辑
type AdminBatchQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAdminBatchQueryLogic 创建管理员批量查询逻辑
func NewAdminBatchQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminBatchQueryLogic {
	return &AdminBatchQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AdminBatchQuery 管理员批量查询轨迹
func (l *AdminBatchQueryLogic) AdminBatchQuery(req *types.AdminBatchQueryRequest) (resp *types.Response, err error) {
	// 参数验证
	if len(req.TrackingNumbers) == 0 {
		return &types.Response{
			Code:    400,
			Message: "tracking_numbers is required",
			Data:    "请提供运单号列表",
		}, nil
	}

	// 限制查询数量（防止滥用）
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}

	// TODO: 调用 gRPC 批量查询
	// 这里暂时返回示例数据
	results := make([]map[string]interface{}, 0, len(req.TrackingNumbers))
	for _, tn := range req.TrackingNumbers[:limit] {
		// TODO: 实际调用 tracking-srv
		results = append(results, map[string]interface{}{
			"tracking_number": tn,
			"status":          "20",
			"status_text":     "InTransit",
			"package_state":   "2",
			"details":         []interface{}{},
		})
	}

	return &types.Response{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"total":   len(results),
			"results": results,
		},
	}, nil
}