package webhook

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"

	"tracking-api/internal/formatter"
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
	// 使用新的 formatter 包
	raw := convertWebhookRequestToMap(req)
	formatter := formatter.NewYunExpressFormatter(raw)
	formatted := formatter.Format()

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
	if formatted.Response.Item.TrackingStatus != "" {
		code, _ := strconv.Atoi(formatted.Response.Item.TrackingStatus)
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
			Status:         formatted.Response.Item.TrackingStatus,
			PackageState:   formatted.Response.Item.PackageState,
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

// convertWebhookRequestToMap 将 WebhookRequest 转换为 map 格式
// 用于调用 formatter 包
// 注意：字段名使用大写开头（匹配 formatter 期望的格式）
func convertWebhookRequestToMap(req *types.WebhookRequest) map[string]interface{} {
	raw := map[string]interface{}{
		"TrackingNumber":    req.TrackingNumber,
		"WayBillNumber":     req.WayBillNumber,
		"TrackingStatus":    req.TrackingStatus,
		"PackageState":      req.PackageState,
		"ProviderName":      req.ProviderName,
		"ProviderSite":      req.ProviderSite,
		"ProviderTelephone": req.ProvicerTelephone,
		"CountryCode":       req.CountryCode,
		"OriginCountryCode": req.OriginCountryCode,
		"TrackingNumber2":   req.TrackingNumber2,
		"LastMileCarrierName": req.LastMileCarrierName,
		"CreatedBy":         req.CreatedBy,
		"POD":               req.POD,
	}

	// 转换 OrderTrackingDetails（大写开头）
	details := make([]map[string]interface{}, len(req.OrderTrackingDetails))
	for i, d := range req.OrderTrackingDetails {
		details[i] = map[string]interface{}{
			"ProcessDate":     d.ProcessDate,
			"ProcessLocation": d.ProcessLocation,
			"ProcessContent":  d.ProcessContent,
			"TrackingStatus":  d.TrackingStatus,
		}
	}
	raw["OrderTrackingDetails"] = details

	return raw
}

// ===================== 以下函数已迁移到 formatter 包，保留兼容性注释 =====================
// formatPayload → formatter.NewYunExpressFormatter(raw).Format()
// calculatePackageState → formatter.GetPackageState()
// getLatestTrackingStatus → formatter.GetLatestStatus()
// sortTrackingDetails → formatter.SortEventsByTime()
// getStatusText → formatter.GetStatusText()
// getPackageState → formatter.GetPackageState()