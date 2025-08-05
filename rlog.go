package qlog

import "github.com/FortuneW/qlog/internal"

// 服务的运行日志
type rLogger struct {
	rlog internal.Logger
}

func (r rLogger) Trace(args ...interface{}) {
	r.rlog.Trace(args...)
}

func (r rLogger) Debug(args ...interface{}) {
	r.rlog.Debug(args...)
}

func (r rLogger) Info(args ...interface{}) {
	r.rlog.Info(args...)
}

func (r rLogger) Warn(args ...interface{}) {
	r.rlog.Warn(args...)
}

func (r rLogger) Error(args ...interface{}) {
	r.rlog.Error(args...)
}

func (r rLogger) Tracef(format string, args ...interface{}) {
	r.rlog.Tracef(format, args...)
}

func (r rLogger) Debugf(format string, args ...interface{}) {
	r.rlog.Debugf(format, args...)
}

func (r rLogger) Infof(format string, args ...interface{}) {
	r.rlog.Infof(format, args...)
}

func (r rLogger) Warnf(format string, args ...interface{}) {
	r.rlog.Warnf(format, args...)
}

func (r rLogger) Errorf(format string, args ...interface{}) {
	r.rlog.Errorf(format, args...)
}

func (r rLogger) WithTraceId(traceId string) RLogger {
	return &rLogger{rlog: r.rlog.WithTraceId(traceId)}
}

var elog = internal.WithModuleName("")

// WriteELogItem 负责写入来自子进程的日志条目
func WriteELogItem(item *ELogItem) {
	if item == nil {
		return
	}
	if internal.GetLevel() > item.Level {
		return
	}
	elog.WriteRawString(item.Content)
}
