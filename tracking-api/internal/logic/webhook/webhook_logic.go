package logic

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"

	"tracking-api/internal/svc"
	"tracking-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"tracking-srv/tracking"
)

// WebhookLogic Webhook 业务逻辑
// 作用：编排业务流程（签名验证 + 格式化 + 调用 gRPC Client）
type WebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewWebhookLogic 创建 Webhook Logic 实例
func NewWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WebhookLogic {
	return &WebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GE_CHANNEL_PREFIX_REGEX GE渠道正则匹配
// 对标：Ruby YunExpressService::GE_CHANNEL_PREFIX_REGEX (yun_express_service.rb:6)
var GE_CHANNEL_PREFIX_REGEX = regexp.MustCompile(`(?i)^(GE-云途|GE云途|GE-YT|GEYT)`)

// IsGeYunChannel 判断是否是GE专属云途渠道
// 对标：Ruby YunExpressService#ge_yun_channel?
func IsGeYunChannel(channelAlias, shippingAgent, shippingChannel string) bool {
	values := []string{channelAlias, shippingAgent, shippingChannel}
	for _, value := range values {
		if GE_CHANNEL_PREFIX_REGEX.MatchString(value) {
			return true
		}
	}
	return false
}

// Webhook 处理云途 Webhook 推送的完整业务流程
// 流程：验证签名 → 格式化数据 → 调用 gRPC 写入数据库
// accountType: "normal"（普通云途）或 "ge"（GE云途）
func (l *WebhookLogic) Webhook(req *types.WebhookRequest, accountType string) (resp *types.Response, err error) {
	// 1. 验证签名（安全第一）
	// 签名验证逻辑：accountType 参数决定使用哪个 Secret
	// - accountType="normal" → 使用 YunExpressWebhookSecret
	// - accountType="ge" → 使用 GEYunExpressWebhookSecret
	if !l.verifySignature(req, accountType) {
		return &types.Response{
			Code:    401,
			Message: "Invalid signature",
			Data:    "签名验证失败",
		}, nil
	}

	// 2. 格式化轨迹数据（对标 Ruby YunExpressTrackFormatter）
	formatted := l.formatPayload(req)

	// 3. 转换为 JSON
	detailJSON, err := json.Marshal(formatted)
	if err != nil {
		return &types.Response{
			Code:    500,
			Message: "Format error",
			Data:    err.Error(),
		}, nil
	}

	// 4. 解析状态码
	statusCode := 0
	if formatted.Item.TrackingStatus != "" {
		code, _ := strconv.Atoi(formatted.Item.TrackingStatus)
		statusCode = code
	}

	// 5. 调用 gRPC Client 写入数据库
	grpcReq := &tracking.UpsertRequest{
		TrackingNumber: req.WayBillNumber,
		Detail:         string(detailJSON),
		Status:         int32(statusCode),
		ServiceClass:   "YunExpressService",
	}

	_, err = l.svcCtx.TrackingRpc.UpsertTracking(l.ctx, grpcReq)
	if err != nil {
		return &types.Response{
			Code:    500,
			Message: "Database error",
			Data:    err.Error(),
		}, nil
	}

	// 6. 返回成功响应
	return &types.Response{
		Code:    200,
		Message: "success",
		Data: types.WebhookResponse{
			Success:        true,
			TrackingNumber: req.WayBillNumber,
			Status:         formatted.Item.TrackingStatus,
			PackageState:   formatted.Item.PackageState,
		},
	}, nil
}

// verifySignature 验证云途 Webhook 签名（HMAC-SHA256）
// 作用：安全验证，防止伪造请求
// accountType: "normal"（普通云途）或 "ge"（GE云途），决定使用哪个Secret
func (l *WebhookLogic) verifySignature(req *types.WebhookRequest, accountType string) bool {
	// 开发环境可跳过签名验证
	if l.svcCtx.Config.YunExpressSkipSignature {
		l.Logger.Info("Skipping signature verification (sandbox mode)")
		return true
	}

	// 根据账号类型选择正确的Secret（对标 Ruby 双账号逻辑）
	var secret string
	if accountType == "ge" {
		secret = l.svcCtx.Config.GEYunExpressWebhookSecret
		l.Logger.Info("Using GE Webhook Secret for signature verification")
	} else {
		secret = l.svcCtx.Config.YunExpressWebhookSecret
		l.Logger.Info("Using Normal Webhook Secret for signature verification")
	}

	// 生产环境必须验证签名（这里简化实现，实际需从 Header 获取签名）
	// TODO: 从 Header 获取签名，计算 HMAC-SHA256，比对签名
	// 参考文档：云途 Webhook 签名规范（需与云途确认）
	
	// 临时实现：Secret配置了就返回true（实际生产必须完整实现HMAC验证）
	if secret != "" {
		l.Logger.Infof("Signature verification configured (secret length: %d)", len(secret))
		return true
	}

	// Secret未配置 → 签名验证失败
	l.Logger.Error("Webhook Secret not configured")
	return false
}

// formatPayload 格式化云途推送的 Payload（对标 Ruby YunExpressTrackFormatter）
func (l *WebhookLogic) formatPayload(req *types.WebhookRequest) *types.TrackingDetailDTO {
	// 对标 Ruby YunExpressTrackFormatter.Format()
	// 1. 计算 PackageState
	packageState := l.calculatePackageState(req.OrderTrackingDetails)

	// 2. 获取最新状态码
	trackingStatus := l.getLatestTrackingStatus(req.OrderTrackingDetails)

	// 3. 排序轨迹详情（按时间升序）
	sortedDetails := l.sortTrackingDetails(req.OrderTrackingDetails)

	// 4. 构建格式化结构（添加 Item 包装层，匹配查询接口期望格式）
	return &types.TrackingDetailDTO{
		Item: types.TrackingItemDTO{
			TrackingNumber:       req.TrackingNumber2,
			WayBillNumber:        req.WayBillNumber,
			CarrierName:          "云途",
			ProviderName:         req.ProviderName,
			ProvicerTelephone:    req.ProvicerTelephone,
			ProviderSite:         req.ProviderSite,
			CountryCode:          req.CountryCode,
			OriginCountryCode:    req.OriginCountryCode,
			TrackingStatus:       trackingStatus,
			PackageState:         packageState,
			IntervalDays:         nil,
			CreatedBy:            req.CreatedBy,
			POD:                  req.POD,
			LastMileCarrierName:  req.LastMileCarrierName,
			OrderTrackingDetails: sortedDetails,
		},
	}
}

// calculatePackageState 计算包裹状态（对标 Ruby STATUS_CODES_TO_PACKAGE_STATES）
func (l *WebhookLogic) calculatePackageState(details []types.TrackingDetail) string {
	if len(details) == 0 {
		return "0"
	}
	latestStatus := l.getLatestTrackingStatus(details)
	return l.getPackageState(latestStatus)
}

// getLatestTrackingStatus 获取最新状态码
func (l *WebhookLogic) getLatestTrackingStatus(details []types.TrackingDetail) string {
	if len(details) == 0 {
		return ""
	}
	sorted := l.sortTrackingDetails(details)
	return sorted[len(sorted)-1].TrackingStatus
}

// sortTrackingDetails 排序轨迹详情（按时间升序）
func (l *WebhookLogic) sortTrackingDetails(details []types.TrackingDetail) []types.TrackingDetailDTOItem {
	if len(details) == 0 {
		return nil
	}

	// 转换为 DTO 格式
	items := make([]types.TrackingDetailDTOItem, len(details))
	for i, d := range details {
		items[i] = types.TrackingDetailDTOItem{
			ProcessDate:     d.ProcessDate,
			ProcessLocation: d.ProcessLocation,
			ProcessContent:  d.ProcessContent,
			TrackingStatus:  d.TrackingStatus,
		}
	}

	// 按时间排序（简化实现：冒泡排序）
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].ProcessDate > items[j].ProcessDate {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	return items
}

// getStatusText 获取状态码对应的文本描述（对标 Ruby YunExpressTrackFormatter::STATUS_CODES）
// 完整版本：对标 yun_express_track_formatter.rb:6-19
func (l *WebhookLogic) getStatusText(trackingStatus string) string {
	STATUS_CODES := map[string]string{
		"0":    "NotFound",
		"10":   "InfoReceived",
		"20":   "InTransit",
		"30":   "AvailableForPickup",
		"40":   "DeliveryFailure",
		"50":   "Delivered",
		"60":   "Exception",
		"70":   "Expired",
		"80":   "Exception",
		"90":   "Exception_Returned",
		"100":  "Exception_Cancel",
		"1001": "InTransit_Arrival", // 非云途状态（用于兼容其他物流商）
	}
	
	if text, ok := STATUS_CODES[trackingStatus]; ok {
		return text
	}
	return "Undefined"
}

// getPackageState 状态码映射到包裹状态（对标 Ruby YunExpressTrackFormatter）
// 完整版本：对标 yun_express_track_formatter.rb:38-50
func (l *WebhookLogic) getPackageState(trackingStatus string) string {
	// 完整的状态码映射（对标 Ruby STATUS_CODES_TO_PACKAGE_STATES）
	STATUS_CODES_TO_PACKAGE_STATES := map[string]string{
		"0":   "0", // Undefined
		"10":  "4", // Received / InfoReceived
		"20":  "2", // InTransit
		"30":  "2", // AvailableForPickup → InTransit
		"40":  "6", // DeliveryFailure
		"50":  "3", // Delivered
		"60":  "6", // Exception → Delivery Failed
		"70":  "6", // Expired → Delivery Failed
		"80":  "6", // Exception → Delivery Failed
		"90":  "7", // Exception_Returned → Returned
		"100": "5", // Exception_Cancel → Canceled
		"1001": "2", // InTransit_Arrival → InTransit（非云途状态）
	}
	
	if state, ok := STATUS_CODES_TO_PACKAGE_STATES[trackingStatus]; ok {
		return state
	}
	return "0" // 默认 Undefined
}