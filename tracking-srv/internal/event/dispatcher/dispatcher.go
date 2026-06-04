package dispatcher

import (
	"sync"

	"tracking-srv/internal/event/define"

	"github.com/zeromicro/go-zero/core/logx"
)

// InMemoryDispatcher 内存事件分发器（同步调用，简化版）
// Demo阶段：同步调用监听器，避免并发复杂性
// 下期：改为异步调用（goroutine）
type InMemoryDispatcher struct {
	// listeners 按事件类型注册的监听器列表
	// key: eventType（如："tracking.upserted"）
	// value: []Listener（监听器数组）
	listeners map[string][]Listener

	// mu 读写锁（保护并发访问）
	mu sync.RWMutex
}

// NewInMemoryDispatcher 创建内存事件分发器
func NewInMemoryDispatcher() *InMemoryDispatcher {
	return &InMemoryDispatcher{
		listeners: make(map[string][]Listener),
	}
}

// Dispatch 分发事件给所有注册的监听器
// Demo阶段：同步调用（简化实现）
func (d *InMemoryDispatcher) Dispatch(event define.Event) {
	d.mu.RLock()
	listeners := d.listeners[event.EventType()]
	d.mu.RUnlock()

	// 获取事件类型对应的监听器列表
	if len(listeners) == 0 {
		logx.Infof("No listeners registered for event: %s", event.EventType())
		return
	}

	// 同步调用所有监听器（Demo阶段简化）
	// 下期：改为异步调用（goroutine）
	// 优化：添加 panic recover，避免单个监听器崩溃影响其他监听器
	for _, listener := range listeners {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("listener panic recovered, event=%s, panic=%v", event.EventType(), r)
				}
			}()
			listener.Handle(event)
		}()
	}

	logx.Infof("Event dispatched: %s, listeners_count=%d", event.EventType(), len(listeners))
}

// Register 注册监听器到指定事件类型
func (d *InMemoryDispatcher) Register(eventType string, listener Listener) {
	d.mu.Lock()
	d.listeners[eventType] = append(d.listeners[eventType], listener)
	d.mu.Unlock()

	logx.Infof("Listener registered: event_type=%s", eventType)
}

// RegisterMultiple 注册多个监听器到同一事件类型
func (d *InMemoryDispatcher) RegisterMultiple(eventType string, listeners []Listener) {
	d.mu.Lock()
	d.listeners[eventType] = append(d.listeners[eventType], listeners...)
	d.mu.Unlock()

	logx.Infof("Multiple listeners registered: event_type=%s, count=%d", eventType, len(listeners))
}