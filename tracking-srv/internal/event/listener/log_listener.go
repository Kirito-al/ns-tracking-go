package listener

import (
	"encoding/json"

	"tracking-srv/internal/event/define"

	"github.com/zeromicro/go-zero/core/logx"
)

// LogListener 日志落地监听器
// Demo阶段：使用 zap 打印事件日志
// 下期：写入日志文件（zap + lumberjack）
type LogListener struct{}

// NewLogListener 创建日志监听器
func NewLogListener() *LogListener {
	return &LogListener{}
}

// Handle 处理事件：打印日志
func (l *LogListener) Handle(event define.Event) {
	// 将事件转为 JSON（用于日志输出）
	eventJSON, err := json.Marshal(event)
	if err != nil {
		logx.Errorf("LogListener marshal event failed: %v", err)
		return
	}

	// 根据事件类型选择日志级别
	switch event.EventType() {
	case "tracking.upserted":
		// 入库成功 → Info 级别
		logx.Infof("Event: %s | EventID: %s | TrackingNumber: %s | Detail: %s",
			event.EventType(),
			event.EventID(),
			l.extractTrackingNumber(event),
			string(eventJSON),
		)

	case "tracking.upsert_failed":
		// 入库失败 → Error 级别
		logx.Errorf("Event: %s | EventID: %s | TrackingNumber: %s | Error: %s | Detail: %s",
			event.EventType(),
			event.EventID(),
			l.extractTrackingNumber(event),
			l.extractErrorMessage(event),
			string(eventJSON),
		)

	default:
		// 其他事件 → Info 级别
		logx.Infof("Event: %s | EventID: %s | Detail: %s",
			event.EventType(),
			event.EventID(),
			string(eventJSON),
		)
	}
}

// extractTrackingNumber 提取运单号（通用方法）
func (l *LogListener) extractTrackingNumber(event define.Event) string {
	switch e := event.(type) {
	case *define.TrackingUpsertedEvent:
		return e.TrackingNumber
	case *define.TrackingUpsertFailedEvent:
		return e.TrackingNumber
	default:
		return "unknown"
	}
}

// extractErrorMessage 提取错误信息（仅失败事件）
func (l *LogListener) extractErrorMessage(event define.Event) string {
	if e, ok := event.(*define.TrackingUpsertFailedEvent); ok {
		return e.ErrorMessage
	}
	return ""
}