package public

import (
	"net/http"

	"ns-tracking-go/service/tracking/api/internal/logic/public"
	"ns-tracking-go/service/tracking/api/internal/svc"
	"ns-tracking-go/service/tracking/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// PublicQueryHandler 公开轨迹查询接口
// URL: GET /api/v1/tracking/:number
// 用途：ns-tracking 前端页面查询轨迹
func PublicQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTrackingRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := public.NewPublicQueryLogic(r.Context(), svcCtx)
		resp, err := l.PublicQuery(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

// PublicBatchQueryHandler 公开批量轨迹查询
// URL: POST /api/v1/tracking/batch
// 用途：ns-tracking 前端页面批量查询轨迹
func PublicBatchQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublicBatchQueryRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := public.NewPublicBatchQueryLogic(r.Context(), svcCtx)
		resp, err := l.PublicBatchQuery(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}