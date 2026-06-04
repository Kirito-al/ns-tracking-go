package tasks

// 任务类型定义（Asynq Task Type）
// 用途：定义所有异步任务类型，用于 Enqueue 和 Handler

const (
	// TaskRetryFailed 入库失败重试任务
	// Payload: TrackingUpsertFailedEvent JSON
	// MaxRetry: 3，指数退避（Asynq 内置）
	TaskRetryFailed = "task:retry_failed"

	// 📌 下期预留任务类型
	// TaskLogUpserted    = "task:log_upserted"    // 入库日志落地
	// TaskSyncLog        = "task:sync_log"        // 同步 tracking_log
	// TaskKafkaProduce   = "task:kafka_produce"   // Kafka 投递
	// TaskDeadLetter     = "task:dead_letter"     // 死信队列处理
)

// MaxRetry 最大重试次数（Asynq 内置）
const MaxRetry = 3