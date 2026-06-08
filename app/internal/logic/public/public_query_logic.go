package public

import (
	"context"
	"encoding/json"
	"time"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/infrastructure/cache"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublicQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicQueryLogic {
	return &PublicQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublicQueryLogic) PublicQuery(req *types.GetTrackingRequest) (resp *types.Response, err error) {
	if req.TrackingNumber == "" {
		return &types.Response{
			Code:    400,
			Message: "tracking number is required",
			Data:    "请提供运单号",
		}, nil
	}

	cacheDetail, hit, err := l.svcCtx.TrackingCache.Get(l.ctx, req.TrackingNumber)
	if err != nil {
		l.Logger.Errorf("Redis cache read failed: %v", err)
	}

	if hit && cacheDetail != nil {
		l.Logger.Infof("Cache hit: tracking_number=%s", req.TrackingNumber)
		var storedMap map[string]interface{}
		if err := json.Unmarshal([]byte(cacheDetail.Detail.(string)), &storedMap); err == nil {
			return buildSuccessResponseFromMap(storedMap)
		}
	}

	l.Logger.Infof("Cache miss: tracking_number=%s", req.TrackingNumber)
	
	detail, err := l.svcCtx.TrackingRepo.FindByTrackingNumber(l.ctx, req.TrackingNumber)
	if err != nil || detail == nil {
		return &types.Response{
			Code:    404,
			Message: "tracking number not found",
			Data:    "运单号不存在",
		}, nil
	}

	var storedMap map[string]interface{}
	if err := json.Unmarshal([]byte(detail.Detail), &storedMap); err != nil {
		return &types.Response{
			Code:    500,
			Message: "data parse error",
			Data:    "数据解析失败",
		}, nil
	}

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

	return buildSuccessResponseFromMap(storedMap)
}

func buildSuccessResponseFromMap(storedMap map[string]interface{}) (*types.Response, error) {
	response, ok := storedMap["response"].(map[string]interface{})
	if !ok {
		return &types.Response{Code: 500, Message: "Invalid data format"}, nil
	}
	
	item, ok := response["Item"].(map[string]interface{})
	if !ok {
		return &types.Response{Code: 500, Message: "Invalid Item format"}, nil
	}
	
	detailsRaw, ok := item["OrderTrackingDetails"]
	if !ok {
		detailsRaw = []interface{}{}
	}
	
	var events []map[string]interface{}
	switch d := detailsRaw.(type) {
	case []interface{}:
		events = convertDetailsArrayToEvents(d)
	case []map[string]interface{}:
		events = convertDetailsMapArrayToEvents(d)
	default:
		events = []map[string]interface{}{}
	}
	
	trackInfo := map[string]interface{}{
		"waybill_number":        getString(item, "WayBillNumber"),
		"tracking_number":       getString(item, "TrackingNumber"),
		"customer_order_number": getString(item, "CustomerOrderNumber"),
		"product_code":          getString(item, "ProductCode"),
		"product_name":          getString(item, "ProviderName"),
		"channel_code":          getString(item, "ChannelCode"),
		"check_in_time":         getString(item, "CreatedBy"),
		"check_out_time":        getString(item, "CheckOutTime"),
		"pick_up_time":          getString(item, "PickUpTime"),
		"customer_code":         getString(item, "CustomerCode"),
		"origin_code":           getString(item, "OriginCountryCode"),
		"destination_code":      getString(item, "CountryCode"),
		"postal_code":           "",
		"actual_weight":         getFloat(item, "ActualWeight"),
		"interval_day":          getFloat(item, "IntervalDay"),
		"interval_work_day":     getFloat(item, "IntervalWorkDay"),
		"last_mile_site":        getString(item, "ProviderSite"),
		"last_mile_name":        getString(item, "LastMileCarrierName"),
		"phone_number":          getString(item, "ProviderTelephone"),
		"track_events":          events,
		"pod_url":               getString(item, "POD"),
		"pod_urls":              getSlice(item, "PodUrls"),
		"IsSignature":           getBool(item, "IsSignature"),
		"SignatureUrls":         getSlice(item, "SignatureUrls"),
		"EstimatedDeliveryToDateZone":   "",
		"EstimatedDeliveryFromDateZone": "",
	}
	
	return &types.Response{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"success": true,
			"result": []map[string]interface{}{
				{
					"order_number":   getString(item, "WayBillNumber"),
					"package_status": mapPackageStateToStatus(getString(item, "PackageState")),
					"track_Info":     trackInfo,
				},
			},
			"t": time.Now().UnixMilli(),
		},
	}, nil
}

func convertDetailsArrayToEvents(details []interface{}) []map[string]interface{} {
	events := make([]map[string]interface{}, len(details))
	for i, detail := range details {
		if d, ok := detail.(map[string]interface{}); ok {
			events[i] = map[string]interface{}{
				"process_time":           getString(d, "ProcessDate"),
				"process_utc_time":       getString(d, "ProcessUTCTime"),
				"process_content":        getString(d, "ProcessContent"),
				"process_country":        getString(d, "ProcessCountry"),
				"process_province":       getString(d, "ProcessProvince"),
				"process_city":           getString(d, "ProcessCity"),
				"process_location":       getString(d, "ProcessLocation"),
				"track_node_code":        getString(d, "TrackNodeCode"),
				"track_node_description": getString(d, "TrackCodeDescription"),
				"pod_url":                getString(d, "PodURL"),
			}
		}
	}
	return events
}

func convertDetailsMapArrayToEvents(details []map[string]interface{}) []map[string]interface{} {
	events := make([]map[string]interface{}, len(details))
	for i, d := range details {
		events[i] = map[string]interface{}{
			"process_time":           getString(d, "ProcessDate"),
			"process_utc_time":       getString(d, "ProcessUTCTime"),
			"process_content":        getString(d, "ProcessContent"),
			"process_country":        getString(d, "ProcessCountry"),
			"process_province":       getString(d, "ProcessProvince"),
			"process_city":           getString(d, "ProcessCity"),
			"process_location":       getString(d, "ProcessLocation"),
			"track_node_code":        getString(d, "TrackNodeCode"),
			"track_node_description": getString(d, "TrackCodeDescription"),
			"pod_url":                getString(d, "PodURL"),
		}
	}
	return events
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if s, ok := v.([]string); ok {
			return s
		}
		if s, ok := v.([]interface{}); ok {
			result := make([]string, len(s))
			for i, item := range s {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return nil
}

func mapPackageStateToStatus(packageState string) string {
	switch packageState {
	case "0":
		return "N"
	case "2":
		return "T"
	case "3":
		return "D"
	case "4":
		return "F"
	case "5":
		return "C"
	case "6":
		return "E"
	case "7":
		return "R"
	default:
		return "T"
	}
}