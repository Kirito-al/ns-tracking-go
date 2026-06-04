package event

import (
	"ns-tracking-go/service/tracking/rpc/internal/event/dispatcher"
	"ns-tracking-go/service/tracking/rpc/internal/event/listener"

	"github.com/zeromicro/go-zero/core/logx"
)

// RegisterListeners 统一注册监听器
// 作用：启动时调用，将所有监听器注册到 Dispatcher
// Demo阶段：只注册 LogListener
func RegisterListeners(d dispatcher.Dispatcher) {
	// ========== 入库成功事件监听器 ==========
	// 注册 LogListener（zap 打印日志）
	d.Register("tracking.upserted", listener.NewLogListener())

	// 📌 下期预留：
	// d.Register("tracking.upserted", listener.NewKafkaListener())  // 投递 Kafka
	// d.Register("tracking.upserted", listener.NewCacheListener())   // 更新缓存

	logx.Info("Listeners registered: tracking.upserted -> LogListener")

	// ========== 入库失败事件监听器 ==========
	// 注册 LogListener（zap 打印日志）
	d.Register("tracking.upsert_failed", listener.NewLogListener())

	// 📌 下期预留：
	// d.Register("tracking.upsert_failed", listener.NewAlertListener())  // 告警通知
	// d.Register("tracking.upsert_failed", listener.NewRetryListener())  // 业务级补偿重试

	logx.Info("Listeners registered: tracking.upsert_failed -> LogListener")

	// ========== 下期扩展：其他事件 ==========
	// 订单创建事件：
	// d.Register("order.created", listener.NewLogListener())
	// d.Register("order.created", listener.NewKafkaListener())

	// 支付成功事件：
	// d.Register("payment.succeeded", listener.NewLogListener())
	// d.Register("payment.succeeded", listener.NewAlertListener())

	// 签收事件：
	// d.Register("tracking.delivered", listener.NewLogListener())
	// d.Register("tracking.delivered", listener.NewEmailListener())
}