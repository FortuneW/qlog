package xlog

import (
	"fmt"

	"github.com/FortuneW/qlog/internal"
)

// 只是一个日志通道不做真实存储和打印,在子进程中写入管道等待主进程获取即可
type eLogger struct {
	moduleName string
	traceId    string
}

type ELogItem struct {
	Level   uint32
	Content string
}

var eLogItems = make(chan *ELogItem, 1024)

func sendToELogItems(item *ELogItem) {
	select {
	case eLogItems <- item:
	default:
	}
}

func (e eLogger) Trace(args ...interface{}) {
	val := fmt.Sprint(args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.TraceLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelTrace, e.formatMessage(val)),
	})
}

func (e eLogger) Debug(args ...interface{}) {
	val := fmt.Sprint(args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.DebugLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelDebug, e.formatMessage(val)),
	})
}

func (e eLogger) Info(args ...interface{}) {
	val := fmt.Sprint(args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.InfoLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelInfo, e.formatMessage(val)),
	})
}

func (e eLogger) Warn(args ...interface{}) {
	val := fmt.Sprint(args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.WarnLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelWarn, e.formatMessage(val)),
	})
}

func (e eLogger) Error(args ...interface{}) {
	val := fmt.Sprint(args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.ErrorLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelError, e.formatMessage(val)),
	})
}

func (e eLogger) Tracef(format string, args ...interface{}) {
	val := fmt.Sprintf(format, args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.TraceLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelTrace, e.formatMessage(val)),
	})
}

func (e eLogger) Debugf(format string, args ...interface{}) {
	val := fmt.Sprintf(format, args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.DebugLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelDebug, e.formatMessage(val)),
	})
}

func (e eLogger) Infof(format string, args ...interface{}) {
	val := fmt.Sprintf(format, args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.InfoLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelInfo, e.formatMessage(val)),
	})
}

func (e eLogger) Warnf(format string, args ...interface{}) {
	val := fmt.Sprintf(format, args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.WarnLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelWarn, e.formatMessage(val)),
	})
}

func (e eLogger) Errorf(format string, args ...interface{}) {
	val := fmt.Sprintf(format, args...)
	if len(val) == 0 {
		return
	}
	sendToELogItems(&ELogItem{
		Level:   internal.ErrorLevel,
		Content: internal.GetOutputStringFormatted(internal.LevelError, e.formatMessage(val)),
	})
}

// WriteELogItem elog 本身不用实现写入函数,统一由主进程的ALog负责写入
func (e eLogger) WriteELogItem(item *ELogItem) {
	return
}

// GetPopELogItemChannel 获取回去的日志通道调用者不可以close,如果担心外部线程长期阻塞或需要中断时，外面可以通过select实现
func (e eLogger) GetPopELogItemChannel() chan *ELogItem {
	return eLogItems
}

func (e eLogger) WithTraceId(traceId string) RLogger {
	return &eLogger{moduleName: e.moduleName, traceId: traceId}
}

func (e eLogger) formatMessage(msg string) string {
	if e.traceId != "" {
		return fmt.Sprintf("[%s] [%s] %s", e.moduleName, e.traceId, msg)
	}
	return fmt.Sprintf("[%s] %s", e.moduleName, msg)
}
