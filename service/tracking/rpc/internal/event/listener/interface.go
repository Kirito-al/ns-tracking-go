package listener

import "ns-tracking-go/service/tracking/rpc/internal/event/define"

// Listener 监听器接口（统一定义）
// 作用：处理事件，执行具体业务逻辑
type Listener interface {
	// Handle 处理事件
	// event: 事件实例（TrackingUpsertedEvent / TrackingUpsertFailedEvent）
	Handle(event define.Event)
}