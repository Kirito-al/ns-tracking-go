package main

// 任务路由注册（简化版）
// Demo阶段：只注册 task:retry_failed
// 下期：注册更多任务类型

import (
	"tracking-srv/internal/queue/handler"
	"tracking-srv/internal/svc"

	"github.com/hibiken/asynq"
)

// RegisterAllHandlers 注册所有任务处理器
// 用途：启动 Worker 时统一注册
func RegisterAllHandlers(mux *asynq.ServeMux, svcCtx *svc.ServiceContext) {
	// ========== Demo阶段：只注册重试任务 ==========
	handler.RegisterRetryHandler(mux, svcCtx)

	// ========== 下期预留：其他任务类型 ==========
	// handler.RegisterLogHandler(mux, svcCtx)        // 日志落地任务
	// handler.RegisterSyncLogHandler(mux, svcCtx)    // 同步 tracking_log
	// handler.RegisterKafkaHandler(mux, svcCtx)      // Kafka 投递任务
	// handler.RegisterDeadLetterHandler(mux, svcCtx) // 死信队列处理
}