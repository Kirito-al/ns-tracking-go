package handler

import (
	"net/http"

	trackingLogic "ns-tracking-go/app/internal/logic/tracking"
	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetTrackingHandler 查询追踪数据接口（内部 API）
// 作用：供 Ruby Rails 调用，查询缓存数据
func GetTrackingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 解析请求参数
		var req types.GetTrackingRequest
		if err := httpx.ParsePath(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 2. 创建 Logic 实例
		l := trackingLogic.NewGetTrackingLogic(r.Context(), svcCtx)

		// 3. 调用 Logic 层查询数据
		resp, err := l.GetTracking(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 4. 返回响应
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
