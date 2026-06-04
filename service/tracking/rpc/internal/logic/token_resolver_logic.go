package logic

import (
	"context"
	"regexp"

	"ns-tracking-go/service/tracking/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// TokenResolverLogic Token 路由逻辑
// 作用：根据渠道信息判断是否是GE渠道，选择正确的Token
// 对标：Ruby YunExpressService#authorization_token (yun_express_service.rb:59-67)
type TokenResolverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewTokenResolverLogic 创建 Token Resolver 实例
func NewTokenResolverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TokenResolverLogic {
	return &TokenResolverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GE_CHANNEL_PREFIX_REGEX GE渠道正则匹配
// 对标：Ruby YunExpressService::GE_CHANNEL_PREFIX_REGEX (yun_express_service.rb:6)
// 匹配规则：GE-云途、GE云途、GE-YT、GEYT（不区分大小写）
var GE_CHANNEL_PREFIX_REGEX = regexp.MustCompile(`(?i)^(GE-云途|GE云途|GE-YT|GEYT)`)

// IsGeYunChannel 判断是否是GE专属云途渠道
// 对标：Ruby YunExpressService#ge_yun_channel? (yun_express_service.rb:65-67)
// 参数：channelAlias（渠道别名）、shippingAgent（物流代理商）、shippingChannel（物流渠道）
// 返回：true = GE渠道（需要使用GE Token），false = 普通云途渠道
func IsGeYunChannel(channelAlias, shippingAgent, shippingChannel string) bool {
	// 检查三个字段中是否有任何一个匹配GE渠道前缀
	values := []string{channelAlias, shippingAgent, shippingChannel}

	for _, value := range values {
		if GE_CHANNEL_PREFIX_REGEX.MatchString(value) {
			return true
		}
	}

	return false
}

// ResolveToken 根据运单号自动选择正确的Token（普通/GE）
// 对标：Ruby YunExpressService#authorization_token (yun_express_service.rb:59-63)
// 流程：
//  1. 查询 tracking_logs 获取渠道信息（channel_alias, shipping_agent, shipping_channel）
//  2. 判断是否是GE渠道（使用 IsGeYunChannel）
//  3. GE渠道 → 返回 GE_YUN_EXPRESS_TOKEN
//  4. 普通渠道 → 返回 YUN_EXPRESS_TOKEN
func (l *TokenResolverLogic) ResolveToken(trackingNumber string) string {
	// 1. 查询 tracking_logs 获取渠道信息（对标 Ruby tracking_context）
	channelAlias, shippingAgent, shippingChannel, err := l.svcCtx.TrackingLogDAO.GetTrackingContext(l.ctx, trackingNumber)
	if err != nil {
		l.Logger.Errorf("Failed to query tracking_log for %s: %v", trackingNumber, err)
		// 查询失败 → 降级使用普通Token（保守策略）
		return l.svcCtx.Config.YunExpressToken
	}

	// 2. 判断是否是GE渠道
	if IsGeYunChannel(channelAlias, shippingAgent, shippingChannel) {
		l.Logger.Infof("GE channel detected for %s: alias=%s, agent=%s, channel=%s",
			trackingNumber, channelAlias, shippingAgent, shippingChannel)

		// GE渠道 → 优先使用GE Token，降级使用普通Token
		geToken := l.svcCtx.Config.GEYunExpressToken
		if geToken != "" {
			return geToken
		}
		logx.Info("GE_YUN_EXPRESS_TOKEN not configured, fallback to normal token")
	}

	// 3. 普通渠道 → 返回普通Token
	return l.svcCtx.Config.YunExpressToken
}

// ResolveWebhookSecret 根据运单号自动选择正确的Webhook签名密钥（普通/GE）
// 对标：Ruby YunExpressService 的双账号逻辑（Webhook场景）
// 用途：供 webhook_handler.go 调用，验证签名时选择正确的Secret
func (l *TokenResolverLogic) ResolveWebhookSecret(trackingNumber string) string {
	// 1. 查询 tracking_logs 获取渠道信息
	channelAlias, shippingAgent, shippingChannel, err := l.svcCtx.TrackingLogDAO.GetTrackingContext(l.ctx, trackingNumber)
	if err != nil {
		l.Logger.Errorf("Failed to query tracking_log for %s: %v", trackingNumber, err)
		// 查询失败 → 降级使用普通Secret
		return l.svcCtx.Config.YunExpressToken // 注意：Webhook Secret 在 API 层配置
	}

	// 2. 判断是否是GE渠道
	if IsGeYunChannel(channelAlias, shippingAgent, shippingChannel) {
		// GE渠道 → 返回GE Secret（需要从API层Config获取）
		// 注意：这里返回空字符串，实际逻辑在API层的 webhook_logic.go 实现
		return "GE_WEBHOOK_SECRET"
	}

	// 3. 普通渠道 → 返回普通Secret
	return "NORMAL_WEBHOOK_SECRET"
}