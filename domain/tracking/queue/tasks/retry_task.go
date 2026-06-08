package tasks

// 任务类型定义（Asynq Task Type）
// 用途：定义所有异步任务类型，用于 Enqueue 和 Handler

const (
	// TaskRetryFailed 入库失败重试任务
	// Payload: TrackingUpsertFailedEvent JSON
	// MaxRetry: 5，指数退避（Asynq 内置）
	TaskRetryFailed = "task:retry_failed"

	// 📌 下期预留任务类型
	// TaskLogUpserted    = "task:log_upserted"    // 入库日志落地
	// TaskSyncLog        = "task:sync_log"        // 同步 tracking_log
	// TaskKafkaProduce   = "task:kafka_produce"   // Kafka 投递
	// TaskDeadLetter     = "task:dead_letter"     // 死信队列处理
)

// MaxRetry 最大重试次数（Asynq 内置）
// 技术方案要求：最大5次，退避时间表 30s → 120s → 600s → 3600s
const MaxRetry = 5

// RetryDelay 退避时间表（技术方案要求）
// 4级退避：30s → 120s → 600s → 3600s
// 用途：配置Asynq的重试延迟策略
var RetryDelay = []int{
	30,   // 第1次重试：30秒后
	120,  // 第2次重试：120秒后
	600,  // 第3次重试：600秒后（10分钟）
	3600, // 第4次重试：3600秒后（1小时）
}