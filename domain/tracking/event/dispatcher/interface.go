package dispatcher

import "ns-tracking-go/domain/tracking/event/define"

// Dispatcher 事件分发器接口（统一定义）
// 作用：统一事件分发器支持注册监听器、触发事件
type Dispatcher interface {
	// Dispatch 分发事件给所有注册的监听器
	// 异步调用监听器，不阻塞主流程
	Dispatch(event define.Event)

	// Register 注册监听器到指定事件类型
	Register(eventType string, listener Listener)

	// RegisterMultiple 注册多个监听器到同一事件类型
	RegisterMultiple(eventType string, listeners []Listener)
}

// Listener 监听器接口（简化版）
// 作用：处理事件，执行具体业务逻辑
type Listener interface {
	// Handle 处理事件
	Handle(event define.Event)
}