package webhook

import (
	"context"
	"encoding/json"
	"strconv"

	"ns-tracking-go/domain/tracking/service"
	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"ns-tracking-go/domain/tracking/entity"
)

// TisPushLogic TIS Push 处理逻辑
// 与 YunExpress Webhook Logic 分离，不走签名验证
// 关键：TIS Push 接收云途 OpenAPI Webhook 推送，标准化为 OMS 格式入库
type TisPushLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewTisPushLogic 创建 TIS Push Logic 实例
func NewTisPushLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TisPushLogic {
	return &TisPushLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// TisPush 处理 TIS Push 请求
// 流程：转换格式 → YunExpressFormatter → gRPC Upsert
func (l *TisPushLogic) TisPush(tisData *types.TisPushData) (*types.Response, error) {
	// 1. 转换 TIS Push 格式 → WebhookRequest（OMS 格式）
	webhookReq := l.convertTisToWebhook(tisData)

	// 2. 提取时间戳（用于 tracking_logs 同步）
	// 提取时间戳（暂不使用）

	// 3. 格式化数据（复用 YunExpressFormatter）
	raw := l.convertWebhookRequestToMap(webhookReq)
	f := service.NewYunExpressFormatter(raw)
	formatted := f.Format()

	// 4. 转换为 JSON
	detailJSON, err := json.Marshal(formatted)
	if err != nil {
		l.Logger.Errorf("JSON marshal failed: %v", err)
		return &types.Response{
			Code:    500,
			Message: "Format error",
			Data:    "", // ← 修复：不泄露内部错误信息
		}, nil
	}

	// 5. 解析状态码
	statusCode := 0
	if formatted.Response.Item.TrackingStatus != "" {
		code, _ := strconv.Atoi(formatted.Response.Item.TrackingStatus)
		statusCode = code
	}

	// 6. Upsert 入库（传递3个时间戳）
	detail := &entity.TrackingDetail{
		TrackingNumber: webhookReq.WayBillNumber,
		Detail:         string(detailJSON),
		Status:         int32(statusCode),
		ServiceClass:   l.svcCtx.Config.ServiceClass, // 从配置读取（修复硬编码）
	}

	err = l.svcCtx.TrackingRepo.Save(l.ctx, detail)
	if err != nil {
		// 记录详细错误日志（内部调试）
		l.Logger.Errorf("gRPC Upsert failed: %v", err)
		
		// 返回通用错误（防止内部信息泄露）
		return &types.Response{
			Code:    500,
			Message: "Internal server error",
			Data:    "", // ← 不返回内部错误详情
		}, nil
	}

	// 6. 返回成功响应
	return &types.Response{
		Code:    200,
		Message: "success",
		Data: types.WebhookResponse{
			Success:        true,
			TrackingNumber: webhookReq.WayBillNumber,
			Status:         formatted.Response.Item.TrackingStatus,
			PackageState:   formatted.Response.Item.PackageState,
		},
	}, nil
}

// convertTisToWebhook 将 TIS Push 转换为 WebhookRequest
// 关键：track_events → OrderTrackingDetails，节点编码 → TrackingStatus
func (l *TisPushLogic) convertTisToWebhook(tisData *types.TisPushData) *types.WebhookRequest {
	// 1. 转换 track_events → orderTrackingDetails
	details := make([]types.TrackingDetail, len(tisData.TrackEvents))
	for i, evt := range tisData.TrackEvents {
		// 处理可选指针字段（如果为nil使用空字符串）
		processLocation := ""
		if evt.ProcessLocation != nil {
			processLocation = *evt.ProcessLocation
		}

		details[i] = types.TrackingDetail{
			ProcessDate:     evt.ProcessTime,                   // 用当地时间
			ProcessLocation: processLocation,                   // 地点（处理指针）
			ProcessContent:  evt.TrackNodeDescription,          // 描述作为内容
			TrackingStatus:  service.MapNodeCode(evt.TrackNodeCode), // 使用完整映射表
		}
	}

	// 2. 获取最新状态（最后一条事件）
	latestStatus := "20" // 默认运输中
	if len(details) > 0 {
		latestStatus = details[len(details)-1].TrackingStatus
	}

	// 3. 映射 package_status → packageState（必须是数字）
	packageState := l.mapPackageState(tisData.PackageStatus)

	// 4. 处理可选指针字段
	lastMileSite := ""
	if tisData.LastMileSite != nil {
		lastMileSite = *tisData.LastMileSite
	}

	phoneNumber := ""
	if tisData.PhoneNumber != nil {
		phoneNumber = *tisData.PhoneNumber
	}

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

	// 5. 构建 WebhookRequest（OMS 格式）
	return &types.WebhookRequest{
		TrackingNumber:       tisData.TrackingNumber,                     // 尾程单号
		WayBillNumber:        tisData.WaybillNumber,                      // 主单号
		TrackingStatus:       latestStatus,                               // 最新状态码
		PackageState:         packageState,                               // 包裹状态（数字）
		OrderTrackingDetails: details,                                    // 轨迹明细数组
		ProviderName:         l.svcCtx.Config.ProviderName,               // 服务商名称（从配置读取）
		ProviderSite:         lastMileSite,                               // 尾程网站（处理指针）
		ProvicerTelephone:    phoneNumber,                                // 尾程电话（处理指针）
		CountryCode:          destinationCode,                            // 目的国（处理指针）
		OriginCountryCode:    originCode,                                 // 始发国（处理指针）
		TrackingNumber2:      tisData.TrackingNumber,                     // 尾程单号（与 trackingNumber 相同）
		LastMileCarrierName:  lastMileName,                               // 尾程承运商（处理指针）
		CreatedBy:            checkInTime,                                // 入库时间（处理指针）
		POD:                  l.extractPodUrl(tisData.TrackEvents),       // POD URL
	}
}

// mapPackageState 映射 TIS package_status → PackageState（必须是数字）
// Ruby 要求：PackageState 必须是数字字符串（'0'-'7'）
func (l *TisPushLogic) mapPackageState(tisStatus string) string {
	switch tisStatus {
	case "I", "InfoReceived", "F":
		return "4" // Received
	case "T", "InTransit":
		return "2" // In Transit
	case "D", "Delivered":
		return "3" // Delivered
	case "E", "Exception":
		return "6" // Exception
	case "R", "Returned":
		return "7" // Returned
	case "C", "Canceled":
		return "5" // Canceled
	case "N", "NotFound":
		return "0" // Undefined
	default:
		return "2" // 默认运输中
	}
}

// extractPodUrl 提取 POD URL（从最后一条事件）
func (l *TisPushLogic) extractPodUrl(events []types.TisTrackEvent) string {
	// 从最后一条事件提取 pod_url
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].PodURL != nil && *events[i].PodURL != "" {
			return *events[i].PodURL
		}
	}
	return ""
}

// extractTimestamps 提取3个关键时间戳（用于 tracking_logs 同步）
// received_at：揽收时间（FIRST_MILE_ARRIVE 或 XD/ZC 节点的 process_time）
// delivered_at：签收时间（DELIVERED 或 MD/TT 节点的 process_time）
// tracked_at：最后轨迹时间（最后一条事件的 process_time）
func (l *TisPushLogic) extractTimestamps(events []types.TisTrackEvent) (receivedAt, deliveredAt, trackedAt string) {
	// 提取最后轨迹时间（最后一条事件）
	if len(events) > 0 {
		trackedAt = events[len(events)-1].ProcessTime
	}

	// 提取揽收时间（FIRST_MILE_ARRIVE / XD/ZC / XD/CS）
	for _, evt := range events {
		if evt.TrackNodeCode == "FIRST_MILE_ARRIVE" || evt.TrackNodeCode == "XD/ZC" || evt.TrackNodeCode == "XD/CS" {
			receivedAt = evt.ProcessTime
			break // 找到第一个即停止
		}
	}

	// 提取签收时间（DELIVERED / MD/TT）
	for _, evt := range events {
		if evt.TrackNodeCode == "DELIVERED" || evt.TrackNodeCode == "MD/TT" {
			deliveredAt = evt.ProcessTime
			break // 找到第一个即停止
		}
	}

	return receivedAt, deliveredAt, trackedAt
}

// convertWebhookRequestToMap 转换 WebhookRequest 为 map（供 Formatter 使用）
func (l *TisPushLogic) convertWebhookRequestToMap(req *types.WebhookRequest) map[string]interface{} {
	// 转换 OrderTrackingDetails
	events := make([]map[string]interface{}, len(req.OrderTrackingDetails))
	for i, detail := range req.OrderTrackingDetails {
		events[i] = map[string]interface{}{
			"ProcessDate":     detail.ProcessDate,
			"ProcessLocation": detail.ProcessLocation,
			"ProcessContent":  detail.ProcessContent,
			"TrackingStatus":  detail.TrackingStatus,
		}
	}

	return map[string]interface{}{
		"TrackingNumber":       req.TrackingNumber,
		"WayBillNumber":        req.WayBillNumber,
		"TrackingStatus":       req.TrackingStatus,
		"PackageState":         req.PackageState,
		"ProviderName":         req.ProviderName,
		"ProviderSite":         req.ProviderSite,
		"ProvicerTelephone":    req.ProvicerTelephone,
		"CountryCode":          req.CountryCode,
		"OriginCountryCode":    req.OriginCountryCode,
		"LastMileCarrierName":  req.LastMileCarrierName,
		"CreatedBy":            req.CreatedBy,
		"POD":                  req.POD,
		"OrderTrackingDetails": events,
	}
}