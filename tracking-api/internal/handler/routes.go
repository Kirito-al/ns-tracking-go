package routes

import (
	"net/http"

	webhookHandler "tracking-api/internal/handler/webhook"
	trackingHandler "tracking-api/internal/handler/tracking"
	healthHandler "tracking-api/internal/handler/health"
	adminHandler "tracking-api/internal/handler/admin"
	publicHandler "tracking-api/internal/handler/public"
	"tracking-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册所有 HTTP 处理器
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// ========== Webhook 端点（云途推送） ==========
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/webhook/yunexpress/tracking/normal",
				Handler: webhookHandler.WebhookNormalHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/webhook/yunexpress/tracking/ge",
				Handler: webhookHandler.WebhookGEHandler(serverCtx),
			},
		},
	)

	// ========== 内部 API（供 Ruby Rails 调用） ==========
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/internal/tracking/:number",
				Handler: trackingHandler.GetTrackingHandler(serverCtx),
			},
		},
	)

	// ========== ns-admin 管理后台 API ==========
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/admin/tracking/batch",
				Handler: adminHandler.AdminBatchQueryHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/admin/tracking/:number",
				Handler: adminHandler.AdminGetTrackingHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/admin/health",
				Handler: adminHandler.AdminHealthHandler(serverCtx),
			},
		},
	)

	// ========== ns-tracking 公开 API ==========
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/tracking/:number",
				Handler: publicHandler.PublicQueryHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/tracking/batch",
				Handler: publicHandler.PublicBatchQueryHandler(serverCtx),
			},
		},
	)

	// ========== 健康检查 ==========
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/health",
				Handler: healthHandler.HealthHandler(serverCtx),
			},
		},
	)
}