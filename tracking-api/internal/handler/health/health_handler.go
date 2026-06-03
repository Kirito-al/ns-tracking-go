package handler

import (
	"net/http"

	"tracking-api/internal/svc"
	"tracking-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// HealthHandler 健康检查接口
// 作用：用于 K8s / 服务器监控，检查服务是否正常运行
func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 直接返回健康状态
		httpx.OkJsonCtx(r.Context(), w, types.Response{
			Code:    200,
			Message: "healthy",
			Data: map[string]string{
				"status":  "ok",
				"service": "yunexpress-webhook",
				"version": "1.0.0",
			},
		})
	}
}