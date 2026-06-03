package handler

import (
	"net/http"

	"tracking-api/internal/logic/admin"
	"tracking-api/internal/svc"
	"tracking-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AdminBatchQueryHandler 管理员批量查询轨迹
// URL: POST /api/v1/admin/tracking/batch
// 用途：ns-admin 管理后台批量查询轨迹数据
func AdminBatchQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminBatchQueryRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewAdminBatchQueryLogic(r.Context(), svcCtx)
		resp, err := l.AdminBatchQuery(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminGetTrackingHandler 管理员查询轨迹详情
// URL: GET /api/v1/admin/tracking/:number
// 用途：ns-admin 管理后台查询单个轨迹详情
func AdminGetTrackingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTrackingRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewAdminGetTrackingLogic(r.Context(), svcCtx)
		resp, err := l.AdminGetTracking(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// AdminHealthHandler 管理员健康检查
// URL: GET /api/v1/admin/health
// 用途：ns-admin 管理后台检查服务健康状态
func AdminHealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminHealthRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewAdminHealthLogic(r.Context(), svcCtx)
		resp, err := l.AdminHealth(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}