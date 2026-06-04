package handler

import (
	"context"
	"encoding/json"

	"ns-tracking-go/service/tracking/rpc/internal/event/define"
	"ns-tracking-go/service/tracking/rpc/internal/queue/tasks"
	"ns-tracking-go/service/tracking/rpc/internal/svc"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logx"
)

// RetryHandler 入库失败重试任务处理器
// 用途：Worker 消费 task:retry_failed，重新执行 Upsert 操作
type RetryHandler struct {
	svcCtx *svc.ServiceContext
}

// NewRetryHandler 创建重试处理器
func NewRetryHandler(svcCtx *svc.ServiceContext) *RetryHandler {
	return &RetryHandler{
		svcCtx: svcCtx,
	}
}

// ProcessTask 处理重试任务（asynq.Handler 接口）
func (h *RetryHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// 1. 解析 Payload（TrackingUpsertFailedEvent JSON）
	var event define.TrackingUpsertFailedEvent
	if err := json.Unmarshal(task.Payload(), &event); err != nil {
		logx.Errorf("RetryHandler unmarshal payload failed: %v", err)
		return err
	}

	logx.Infof("RetryHandler processing: tracking_number=%s, retry_count=%d, failed_stage=%s",
		event.TrackingNumber, event.RetryCount, event.FailedStage)

	// 2. 提取原始数据（用于重新 Upsert）
	// OriginalPayload 包含：TrackingNumber, Detail, Status, ServiceClass 等字段
	var originalData struct {
		TrackingNumber string `json:"tracking_number"`
		Detail         string `json:"detail"`
		Status         int32  `json:"status"`
		ServiceClass   string `json:"service_class"`
	}

	if err := json.Unmarshal([]byte(event.OriginalPayload), &originalData); err != nil {
		logx.Errorf("RetryHandler unmarshal original payload failed: %v", err)
		return err
	}

	// 3. 调用 DAO 重新 Upsert（重试入库）
	err := h.svcCtx.TrackingDAO.Upsert(ctx,
		originalData.TrackingNumber,
		originalData.Detail,
		originalData.Status,
		originalData.ServiceClass,
	)

	if err != nil {
		logx.Errorf("RetryHandler upsert failed (retry %d): %v", event.RetryCount, err)
		// 返回错误 → Asynq 自动重试（MaxRetry 3）
		return err
	}

	// 4. 成功 → 打印日志
	logx.Infof("RetryHandler success: tracking_number=%s, retry_count=%d",
		originalData.TrackingNumber, event.RetryCount)

	// 5. 触发入库成功事件（可选）
	// TODO: h.svcCtx.EventDispatcher.Dispatch(ctx, define.NewTrackingUpsertedEvent(...))

	return nil
}

// RegisterRetryHandler 注册重试处理器到 Asynq Server
func RegisterRetryHandler(mux *asynq.ServeMux, svcCtx *svc.ServiceContext) {
	mux.HandleFunc(tasks.TaskRetryFailed, NewRetryHandler(svcCtx).ProcessTask)
	logx.Info("Asynq handler registered: task:retry_failed")
}