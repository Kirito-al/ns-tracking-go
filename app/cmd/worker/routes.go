package main

// 任务路由注册（简化版）
// Demo阶段：注册 task:retry_failed + task:tis_push_async

import (
	"ns-tracking-go/app/internal/queue/handler"
	"ns-tracking-go/app/internal/svc"

	"github.com/hibiken/asynq"
)

// RegisterAllHandlers 注册所有任务处理器
// 用途：启动 Worker 时统一注册
func RegisterAllHandlers(mux *asynq.ServeMux, svcCtx *svc.ServiceContext) {
	// ========== Demo阶段：注册重试任务 + TIS Push异步处理 ==========
	handler.RegisterRetryHandler(mux, svcCtx)
	handler.RegisterTisPushAsyncHandler(mux, svcCtx) // ← 新增：快返回模式的异步处理

	// ========== 下期预留：其他任务类型 ==========
	// handler.RegisterLogHandler(mux, svcCtx)        // 日志落地任务
	// handler.RegisterSyncLogHandler(mux, svcCtx)    // 同步 tracking_log
	// handler.RegisterKafkaHandler(mux, svcCtx)      // Kafka 投递任务
	// handler.RegisterDeadLetterHandler(mux, svcCtx) // 死信队列处理
}