package handler

import (
	"net/http"

	"ns-tracking-go/app/internal/logic/webhook"
	"ns-tracking-go/app/internal/normalize"
	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// TisPushHandler TIS Push Data 处理器
// 对标：真实业务中的 tisPushData 格式（snake_case，带 data 包装层）
// 关键：TIS Push 不走签名验证，调用独立的 TisPushLogic
func TisPushHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 解析请求（TIS Push 格式）
		var req types.TisPushRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 2. 验证 data 存在
		if req.Data == nil {
			httpx.OkJsonCtx(r.Context(), w, &types.Response{
				Code:    400,
				Message: "Invalid TIS Push data",
				Data:    "data 字段不能为空",
			})
			return
		}

		// 3. 调用独立的 TisPushLogic（不走签名验证）
		l := webhook.NewTisPushLogic(r.Context(), svcCtx)
		resp, err := l.TisPush(req.Data)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 4. 返回响应
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// convertTisToWebhook 将 TIS Push 转换为 WebhookRequest
func convertTisToWebhook(tisData *types.TisPushData) *types.WebhookRequest {
	// 转换 track_events → orderTrackingDetails
	details := make([]types.TrackingDetail, len(tisData.TrackEvents))
	for i, evt := range tisData.TrackEvents {
		// 处理可选指针字段
		processLocation := ""
		if evt.ProcessLocation != nil {
			processLocation = *evt.ProcessLocation
		}

		details[i] = types.TrackingDetail{
			ProcessDate:     evt.ProcessTime,       // 用当地时间
			ProcessLocation: processLocation,       // 地点（处理指针）
			ProcessContent:  evt.TrackNodeDescription, // 描述作为内容
			TrackingStatus:  getEventStatus(evt.TrackNodeCode), // 根据节点代码推断状态
		}
	}

	// 获取最新状态（最后一条事件）
	latestStatus := "20" // 默认运输中
	if len(details) > 0 {
		latestStatus = details[len(details)-1].TrackingStatus
	}

	// 映射 package_status → packageState
	packageState := mapPackageState(tisData.PackageStatus)

	// 处理可选指针字段
	destinationCode := ""
	if tisData.DestinationCode != nil {
		destinationCode = *tisData.DestinationCode
	}

	originCode := ""
	if tisData.OriginCode != nil {
		originCode = *tisData.OriginCode
	}

	lastMileName := ""
	if tisData.LastMileName != nil {
		lastMileName = *tisData.LastMileName
	}

	checkInTime := ""
	if tisData.CheckInTime != nil {
		checkInTime = *tisData.CheckInTime
	}

	return &types.WebhookRequest{
		TrackingNumber:       tisData.TrackingNumber,
		WayBillNumber:        tisData.WaybillNumber,
		TrackingStatus:       latestStatus,
		PackageState:         packageState,
		OrderTrackingDetails: details,
		ProviderName:         "TIS",
		CountryCode:          destinationCode,      // 处理指针
		OriginCountryCode:    originCode,           // 处理指针
		LastMileCarrierName:  lastMileName,         // 处理指针
		CreatedBy:            checkInTime,          // 处理指针
	}
}

// getEventStatus 根据节点代码推断状态码
// 使用完整的39个节点映射表
func getEventStatus(nodeCode string) string {
	return normalize.MapNodeCode(nodeCode)
}

// mapPackageState 映射 TIS package_status → packageState
func mapPackageState(tisStatus string) string {
	switch tisStatus {
	case "I", "InfoReceived":
		return "4" // Received
	case "T", "InTransit":
		return "2" // In Transit
	case "D", "Delivered":
		return "3" // Delivered
	case "E", "Exception":
		return "6" // Exception
	case "R", "Returned":
		return "7" // Returned
	default:
		return "2" // 默认运输中
	}
}
