package handler

import (
	"io"
	"net/http"
	"strings"

	webhookLogic "tracking-api/internal/logic/webhook"
	"tracking-api/internal/svc"
	"tracking-api/internal/types"
	"tracking-api/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 签名Header常量（需与云途确认实际Header名称）
const (
	// YunExpressSignatureHeader 云途Webhook签名Header名称
	// 注意：需与云途官方文档确认实际Header名称（可能是 X-Signature、X-YunExpress-Signature等）
	YunExpressSignatureHeader = "X-YunExpress-Signature"
)

// WebhookNormalHandler 云途 Webhook 推送接口（普通账号）
// 作用：接收云途物流的 Webhook 推送，调用 Logic 层处理业务
func WebhookNormalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
// ========== Step 1: 提取签名验证所需数据 ==========
		// 读取原始 Body（用于签名验证）
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// 恢复 Body（供后续解析使用）
		r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		// 提取签名Header
		receivedSignature := r.Header.Get(YunExpressSignatureHeader)

		// ========== Step 2: 解析请求参数 ==========
		var req types.WebhookRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// ========== Step 3: 验证签名（安全第一）==========
		// 使用签名工具验证
		signer := utils.NewHMACSHA256Signer()
		secret := svcCtx.Config.YunExpressWebhookSecret // 普通账号Secret

		// 生产环境必须验证签名（skipSignature=false时）
		if !svcCtx.Config.YunExpressSkipSignature && receivedSignature != "" {
			if !signer.VerifySignatureSafe(secret, bodyBytes, receivedSignature) {
				// 签名验证失败 → 返回401
				httpx.OkJsonCtx(r.Context(), w, &types.Response{
					Code:    401,
					Message: "Invalid HMAC signature",
					Data:    "签名验证失败（HMAC-SHA256不匹配）",
				})
				return
			}
		}

		// ========== Step 4: 创建 Logic 实例并处理业务 ==========
		l := webhookLogic.NewWebhookLogic(r.Context(), svcCtx)
		resp, err := l.Webhook(&req, "normal")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// ========== Step 5: 返回响应 ==========
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// WebhookGEHandler 云途 Webhook 推送接口（GE账号）
// 区别：使用GE账号Secret验证签名
func WebhookGEHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ========== Step 1: 提取签名验证所需数据 ==========
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		receivedSignature := r.Header.Get(YunExpressSignatureHeader)

		// ========== Step 2: 解析请求参数 ==========
		var req types.WebhookRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// ========== Step 3: 验证签名（使用GE Secret）==========
		signer := utils.NewHMACSHA256Signer()
		secret := svcCtx.Config.GEYunExpressWebhookSecret // GE账号Secret

		if !svcCtx.Config.YunExpressSkipSignature && receivedSignature != "" {
			if !signer.VerifySignatureSafe(secret, bodyBytes, receivedSignature) {
				httpx.OkJsonCtx(r.Context(), w, &types.Response{
					Code:    401,
					Message: "Invalid HMAC signature (GE account)",
					Data:    "签名验证失败（GE账号HMAC-SHA256不匹配）",
				})
				return
			}
		}

		// ========== Step 4: 创建 Logic 实例并处理业务 ==========
		l := webhookLogic.NewWebhookLogic(r.Context(), svcCtx)
		resp, err := l.Webhook(&req, "ge")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// ========== Step 5: 返回响应 ==========
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}