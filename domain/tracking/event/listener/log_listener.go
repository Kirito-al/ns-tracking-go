package listener

import (
	"ns-tracking-go/domain/tracking/event/define"
	"ns-tracking-go/pkg/toolx"

	"github.com/zeromicro/go-zero/core/logx"
)

// LogListener 日志落地监听�?
// Demo阶段：使�?zap 打印事件日志
// 下期：写入日志文件（zap + lumberjack�?
type LogListener struct{}

// NewLogListener 创建日志监听�?
func NewLogListener() *LogListener {
	return &LogListener{}
}

// Handle 处理事件：打印日�?
func (l *LogListener) Handle(event define.Event) {
	// 根据事件类型选择日志级别
	switch event.EventType() {
	case "tracking.upserted":
		// 入库成功 �?Info 级别
		logx.Infof("Event: %s | EventID: %s | TrackingNumber: %s",
			event.EventType(),
			event.EventID(),
			l.extractTrackingNumber(event),
		)
		// 详细 JSON 日志（脱敏后�?
		l.logEventJSON("info", event)

	case "tracking.upsert_failed":
		// 入库失败 �?Error 级别
		logx.Errorf("Event: %s | EventID: %s | TrackingNumber: %s | Error: %s",
			event.EventType(),
			event.EventID(),
			l.extractTrackingNumber(event),
			l.extractErrorMessage(event),
		)
		// 详细 JSON 日志（脱敏后�?
		l.logEventJSON("error", event)

	default:
		// 其他事件 �?Info 级别
		logx.Infof("Event: %s | EventID: %s",
			event.EventType(),
			event.EventID(),
		)
		l.logEventJSON("info", event)
	}
}

// logEventJSON 打印事件 JSON 日志（辅助方法）
func (l *LogListener) logEventJSON(level string, event define.Event) {
	// 如需调试，可在此打印 JSON（生产环境建议关闭或降低级别�?
	// eventJSON, _ := json.Marshal(event)
	// logx.Debugf("Event Detail: %s", string(eventJSON))
}

// extractTrackingNumber 提取运单号（通用方法�?
func (l *LogListener) extractTrackingNumber(event define.Event) string {
	switch e := event.(type) {
	case *define.TrackingUpsertedEvent:
		return utils.MaskTrackingNumber(e.TrackingNumber)
	case *define.TrackingUpsertFailedEvent:
		return utils.MaskTrackingNumber(e.TrackingNumber)
	default:
		return "unknown"
	}
}

// extractErrorMessage 提取错误信息（仅失败事件�?
func (l *LogListener) extractErrorMessage(event define.Event) string {
	if e, ok := event.(*define.TrackingUpsertFailedEvent); ok {
		return e.ErrorMessage
	}
	return ""
}
