// Package consts 全局常量定义
// 包含：渠道枚举、状态码映射、错误定义
package consts

// ===================== 云途渠道常量 =====================

const (
	// ChannelNormal 普通云途渠道
	ChannelNormal = "normal"
	// ChannelGE GE专属渠道
	ChannelGE = "ge"
)

// 渠道名称映射
var ChannelNames = map[string]string{
	"normal": "云途普通渠道",
	"ge":     "云途GE专属渠道",
}

// ===================== 状态码映射（对标Ruby STATUS_CODES） =====================

// YunExpressStatus 云途状态码映射
// 来源: Ruby yun_express_track_formatter.rb
var YunExpressStatus = map[string]string{
	"10": "电子信息已收到",     // ELECTRONIC_INFO_RECEIVED
	"20": "已揽收",             // PICKED_UP
	"30": "运输中",             // IN_TRANSIT
	"40": "到达目的国",         // ARRIVED_AT_DESTINATION_COUNTRY
	"50": "海关查验中",         // CUSTOMS_INSPECTION
	"60": "海关放行",           // CUSTOMS_RELEASED
	"70": "派送中",             // OUT_FOR_DELIVERY
	"80": "已签收",             // DELIVERED
	"90": "异常",               // EXCEPTION
	"100": "退回",              // RETURNED
	"110": "拦截",              // INTERCEPTED
	"120": "已取消",            // CANCELLED
}

// PackageState 包裹状态映射
var PackageState = map[string]string{
	"1":  "正常",
	"2":  "已发出",
	"3":  "到达",
	"4":  "签收",
	"5":  "异常",
	"6":  "退回",
}

// ===================== 错误码定义 =====================

const (
	// ErrCodeSuccess 成功
	ErrCodeSuccess = 200
	// ErrCodeBadRequest 参数错误
	ErrCodeBadRequest = 400
	// ErrCodeUnauthorized 签名验证失败
	ErrCodeUnauthorized = 401
	// ErrCodeNotFound 数据不存在
	ErrCodeNotFound = 404
	// ErrCodeInternalError 内部错误
	ErrCodeInternalError = 500
	// ErrCodeServiceUnavailable 服务不可用
	ErrCodeServiceUnavailable = 503
)

// ErrorMessages 错误消息映射
var ErrorMessages = map[int]string{
	ErrCodeSuccess:            "success",
	ErrCodeBadRequest:         "invalid request parameters",
	ErrCodeUnauthorized:       "invalid HMAC signature",
	ErrCodeNotFound:           "tracking data not found",
	ErrCodeInternalError:      "internal server error",
	ErrCodeServiceUnavailable: "service temporarily unavailable",
}

// ===================== GE渠道正则匹配 =====================

// GEChannelPatterns GE渠道匹配模式（运单号前缀）
var GEChannelPatterns = []string{
	"YTGE",     // GE云途前缀
	"GEYT",     // GE云途前缀（反向）
	"GE",       // GE前缀
	"GLS",      // GLS快递
	"HKGE",     // 香港GE
}

// ===================== 缓存Key常量 =====================

const (
	// CacheKeyTrackingPrefix 轨迹缓存前缀
	CacheKeyTrackingPrefix = "tracking:detail:"
	// CacheKeyLogPrefix 轨迹日志缓存前缀
	CacheKeyLogPrefix = "tracking:log:"
	// CacheKeySystemConfig 系统配置缓存
	CacheKeySystemConfig = "system:config:"
	// CacheTTLDefault 默认缓存时长（4小时）
	CacheTTLDefault = 14400 // seconds
)

// ===================== 时间常量 =====================

const (
	// SyncInterval 同步间隔
	SyncInterval = 3600 // 1小时
	// RetryMax 最大重试次数
	RetryMax = 3
	// RetryDelay 重试延迟
	RetryDelay = 5 // 5秒
)