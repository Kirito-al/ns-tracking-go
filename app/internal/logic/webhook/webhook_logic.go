package webhook

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/domain/tracking/entity"
	"ns-tracking-go/domain/tracking/service"
	"ns-tracking-go/pkg/toolx"

	"github.com/zeromicro/go-zero/core/logx"
)

type WebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WebhookLogic {
	return &WebhookLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

var GE_CHANNEL_PREFIX_REGEX = regexp.MustCompile(`(?i)^(GE-云途|GE云途|GE-YT|GEYT)`)

func IsGeYunChannel(channelAlias, shippingAgent, shippingChannel string) bool {
	values := []string{channelAlias, shippingAgent, shippingChannel}
	for _, value := range values {
		if GE_CHANNEL_PREFIX_REGEX.MatchString(value) {
			return true
		}
	}
	return false
}

func (l *WebhookLogic) Webhook(req *types.WebhookRequest, accountType string) (resp *types.Response, err error) {
	// 1. 格式化轨迹数据（DDD架构：直接调用domain层）
	raw := convertWebhookRequestToMap(req)
	formatter := service.NewYunExpressFormatter(raw)
	formatted := formatter.Format()

	// 2. 转换为 JSON
	detailJSON, err := json.Marshal(formatted)
	if err != nil {
		return &types.Response{
			Code:    500,
			Message: "Format error",
			Data:    "",
		}, nil
	}

	// 3. 解析状态码
	statusCode := 0
	if formatted.Response.Item.TrackingStatus != "" {
		code, _ := strconv.Atoi(formatted.Response.Item.TrackingStatus)
		statusCode = code
	}

	// 4. DDD架构：直接调用仓储层入库（不经过gRPC）
	detail := &entity.TrackingDetail{
		TrackingNumber: req.WayBillNumber,
		Detail:         string(detailJSON),
		Status:         int32(statusCode),
		ServiceClass:   l.svcCtx.Config.ServiceClass, // 从配置读取（修复硬编码）
	}

	err = l.svcCtx.TrackingRepo.Save(l.ctx, detail)
	if err != nil {
		l.Logger.Errorf("Upsert tracking_details failed: %v", err)
		return &types.Response{
			Code:    500,
			Message: "Database error",
			Data:    "",
		}, nil
	}

	l.Logger.Infof("Webhook processed successfully: %s", toolx.MaskTrackingNumber(req.WayBillNumber))

	return &types.Response{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"trackingNumber": req.WayBillNumber,
			"status":         statusCode,
		},
	}, nil
}

func convertWebhookRequestToMap(req *types.WebhookRequest) map[string]interface{} {
	raw := map[string]interface{}{
		"TrackingNumber":    req.TrackingNumber,
		"WayBillNumber":     req.WayBillNumber,
		"TrackingStatus":    req.TrackingStatus,
		"PackageState":      req.PackageState,
		"ProviderName":      req.ProviderName,
		"ProviderSite":      req.ProviderSite,
		"ProviderTelephone": req.ProviderTelephone,
		"CountryCode":       req.CountryCode,
		"OriginCountryCode": req.OriginCountryCode,
		"LastMileCarrierName": req.LastMileCarrierName,
		"CreatedBy":         req.CreatedBy,
		"POD":               req.POD,
	}

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