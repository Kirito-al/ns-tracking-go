package admin

import (
	"context"
	"encoding/json"
	"time"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/domain/tracking/service"
	"ns-tracking-go/infrastructure/cache"

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

// AdminGetTracking 管理员查询轨迹详情（Redis缓存优先）
func (l *AdminGetTrackingLogic) AdminGetTracking(req *types.GetTrackingRequest) (resp *types.Response, err error) {
	// 参数验证
	if req.TrackingNumber == "" {
		return &types.Response{
			Code:    400,
			Message: "tracking number is required",
			Data:    "请提供运单号",
		}, nil
	}

	// ===== Redis缓存优先读取 =====
	
	// 1. 先查Redis缓存
	cacheDetail, hit, err := l.svcCtx.TrackingCache.Get(l.ctx, req.TrackingNumber)
	if err != nil {
		l.Logger.Errorf("Redis cache read failed: %v", err)
	}
	
	// 2. 缓存命中：直接返回完整数据
	if hit && cacheDetail != nil {
		l.Logger.Infof("Cache hit: tracking_number=%s", req.TrackingNumber)
		
		var storedDetail types.TrackingDetailDTO
		if err := json.Unmarshal([]byte(cacheDetail.Detail.(string)), &storedDetail); err == nil {
			// 转换OrderTrackingDetails为events
			events := convertDetailsToEvents(storedDetail.Response.Item.OrderTrackingDetails)
			
			return &types.Response{
				Code:    200,
				Message: "success",
				Data: map[string]interface{}{
					"tracking_number": cacheDetail.TrackingNumber,
					"status":          cacheDetail.Status,
					"status_text":     mapStatusToText(cacheDetail.Status),
					"package_state":   storedDetail.Response.Item.PackageState,
					"details":         events,
					"last_update":     time.Unix(cacheDetail.SyncedAt, 0).Format(time.RFC3339),
				},
			}, nil
		}
	}
	
	// 3. 缓存未命中：查DB
	l.Logger.Infof("Cache miss: tracking_number=%s", req.TrackingNumber)
	
	detail, err := l.svcCtx.TrackingRepo.FindByTrackingNumber(l.ctx, req.TrackingNumber)
	if err != nil || detail == nil {
		return &types.Response{
			Code:    404,
			Message: "tracking number not found",
			Data:    "运单号不存在",
		}, nil
	}

	// 4. 解析JSONB
	var storedDetail types.TrackingDetailDTO
	if err := json.Unmarshal([]byte(detail.Detail), &storedDetail); err != nil {
		return &types.Response{
			Code:    500,
			Message: "data parse error",
			Data:    "数据解析失败",
		}, nil
	}

	// 5. 回写Redis缓存
	if l.svcCtx.TrackingCache != nil {
		cacheDTO := &cache.TrackingDetailDTO{
			TrackingNumber: detail.TrackingNumber,
			Detail:         detail.Detail,
			Status:         detail.Status,
			ServiceClass:   detail.ServiceClass,
			SyncedAt:       detail.SyncedAt,
			CachedAt:       time.Now().Unix(),
		}
		l.svcCtx.TrackingCache.Set(l.ctx, req.TrackingNumber, cacheDTO)
	}

	// 6. 转换并返回
	events := convertDetailsToEvents(storedDetail.Response.Item.OrderTrackingDetails)
	
	return &types.Response{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"tracking_number": detail.TrackingNumber,
			"status":          detail.Status,
			"status_text":     mapStatusToText(detail.Status),
			"package_state":   storedDetail.Response.Item.PackageState,
			"details":         events,
			"last_update":     time.Unix(detail.SyncedAt, 0).Format(time.RFC3339),
		},
	}, nil
}

// convertDetailsToEvents 转换OrderTrackingDetails为events格式
func convertDetailsToEvents(details []types.TrackingDetailDTOItem) []map[string]string {
	events := make([]map[string]string, len(details))
	for i, detail := range details {
		// 只在异常节点时填充node_labels
		var nodeLabelText string
		if detail.TrackNodeCode == "DELIVERY_FAILURE" || detail.TrackingStatus == "40" {
			zh, en := service.GetNodeLabels(detail.TrackNodeCode)
			nodeLabelText = zh + "/" + en
		}
		
		events[i] = map[string]string{
			"ProcessDate":          detail.ProcessDate,
			"ProcessLocation":      detail.ProcessLocation,
			"ProcessContent":       detail.ProcessContent,
			"TrackingStatus":       detail.TrackingStatus,
			"TrackNodeCode":        detail.TrackNodeCode,
			"TrackNodeDescription": detail.TrackCodeDescription,
		}
		
		if nodeLabelText != "" {
			events[i]["node_labels"] = nodeLabelText
		}
	}
	return events
}

// mapStatusToText 状态码 → 状态文本
func mapStatusToText(status int32) string {
	switch status {
	case 0:
		return "NotFound"
	case 10:
		return "InfoReceived"
	case 20:
		return "InTransit"
	case 30:
		return "AvailableForPickup"
	case 40:
		return "DeliveryFailure"
	case 50:
		return "Delivered"
	case 60:
		return "Exception"
	case 70:
		return "Expired"
	case 80:
		return "CustomsInspection"
	case 90:
		return "Returned"
	case 100:
		return "Canceled"
	default:
		return "InTransit"
	}
}