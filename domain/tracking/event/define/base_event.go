package define

// Event 事件基础接口
// 所有事件必须实现此接口，用于事件分发器统一处理
type Event interface {
	// EventID 返回事件唯一标识（UUID�?
	EventID() string

	// EventType 返回事件类型（如�?tracking.upserted"�?
	EventType() string

	// EventTimestamp 返回事件触发时间戳（Unix�?
	EventTimestamp() int64
}
