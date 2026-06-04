package handler

import (
	"net/http"

	webhookLogic "tracking-api/internal/logic/webhook"
	"tracking-api/internal/svc"
	"tracking-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// TisPushHandler TIS Push Data 处理器
// 对标：真实业务中的 tisPushData 格式（snake_case，带 data 包装层）
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

		// 3. 转换为 WebhookRequest（复用现有 YunExpressFormatter 逻辑）
		webhookReq := convertTisToWebhook(req.Data)

		// 4. 调用 WebhookLogic 处理
		l := webhookLogic.NewWebhookLogic(r.Context(), svcCtx)
		resp, err := l.Webhook(webhookReq, "normal")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 5. 返回响应
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// convertTisToWebhook 将 TIS Push 转换为 WebhookRequest
func convertTisToWebhook(tisData *types.TisPushData) *types.WebhookRequest {
	// 转换 track_events → orderTrackingDetails
	details := make([]types.TrackingDetail, len(tisData.TrackEvents))
	for i, evt := range tisData.TrackEvents {
		details[i] = types.TrackingDetail{
			ProcessDate:     evt.ProcessTime,       // 用当地时间
			ProcessLocation: evt.ProcessLocation,
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

	return &types.WebhookRequest{
		TrackingNumber:       tisData.TrackingNumber,
		WayBillNumber:        tisData.WaybillNumber,
		TrackingStatus:       latestStatus,
		PackageState:         packageState,
		OrderTrackingDetails: details,
		ProviderName:         "TIS",
		CountryCode:          tisData.DestinationCode,
		OriginCountryCode:    tisData.OriginCode,
		LastMileCarrierName:  tisData.LastMileName,
		CreatedBy:            tisData.CheckInTime,
	}
}

// getEventStatus 根据节点代码推断状态码
func getEventStatus(nodeCode string) string {
	// 简单映射（实际可能需要更复杂的逻辑）
	switch {
	case nodeCode == "YB/XD":
		return "10" // InfoReceived
	case nodeCode == "XD/ZC", nodeCode == "XD/CS":
		return "20" // InTransit (揽收/发出)
	case nodeCode == "MD/DD", nodeCode == "MD/TT":
		return "50" // Delivered (到达/投递)
	default:
		return "20" // 默认运输中
	}
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
