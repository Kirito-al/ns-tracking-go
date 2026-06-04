package handler

import (
	"net/http"

	"ns-tracking-go/app/internal/logic/webhook"
	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// TisPushHandler TIS Push Data 处理器
// 对标：真实业务中的 tisPushData 格式（snake_case，带 data 包装层）
// 关键：TIS Push 不走签名验证，调用独立的 TisPushLogic
func TisPushHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 解析请求（TIS Push 格式）
		var req types.TisPushRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 2. 验证 data 存在
		if req.Data == nil {
			httpx.OkJsonCtx(r.Context(), w, &types.Response{
				Code:    400,
				Message: "Invalid TIS Push data",
				Data:    "data 字段不能为空",
			})
			return
		}

		// 3. 调用独立的 TisPushLogic（不走签名验证）
		// 修复：删除 handler 层的业务逻辑，统一在 logic 层处理
		l := webhook.NewTisPushLogic(r.Context(), svcCtx)
		resp, err := l.TisPush(req.Data)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 4. 返回响应
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
