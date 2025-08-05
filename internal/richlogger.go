package internal

import (
	"fmt"
)

type richLogger struct {
	moduleName string
	traceId    string
}

func WithModuleName(moduleName string) Logger {
	return &richLogger{
		moduleName: moduleName,
	}
}

func (l *richLogger) Trace(v ...any) {
	if !shallLog(TraceLevel) {
		return
	}
	getWriter().Trace(l.formatMessage(fmt.Sprint(v...)))
}

func (l *richLogger) Tracef(format string, v ...any) {
	if !shallLog(TraceLevel) {
		return
	}
	getWriter().Trace(l.formatMessage(fmt.Sprintf(format, v...)))
}

func (l *richLogger) Debug(v ...any) {
	if !shallLog(DebugLevel) {
		return
	}
	getWriter().Debug(l.formatMessage(fmt.Sprint(v...)))
}

func (l *richLogger) Debugf(format string, v ...any) {
	if !shallLog(DebugLevel) {
		return
	}
	getWriter().Debug(l.formatMessage(fmt.Sprintf(format, v...)))
}

func (l *richLogger) Error(v ...any) {
	if !shallLog(ErrorLevel) {
		return
	}
	getWriter().Error(l.formatMessage(fmt.Sprint(v...)))
}

func (l *richLogger) Errorf(format string, v ...any) {
	if !shallLog(ErrorLevel) {
		return
	}
	getWriter().Error(l.formatMessage(fmt.Sprintf(format, v...)))
}

func (l *richLogger) Warn(v ...any) {
	if !shallLog(WarnLevel) {
		return
	}
	getWriter().Warn(l.formatMessage(fmt.Sprint(v...)))
}

func (l *richLogger) Warnf(format string, v ...any) {
	if !shallLog(WarnLevel) {
		return
	}
	getWriter().Warn(l.formatMessage(fmt.Sprintf(format, v...)))
}

func (l *richLogger) Info(v ...any) {
	if !shallLog(InfoLevel) {
		return
	}
	getWriter().Info(l.formatMessage(fmt.Sprint(v...)))
}

func (l *richLogger) Infof(format string, v ...any) {
	if !shallLog(InfoLevel) {
		return
	}
	getWriter().Info(l.formatMessage(fmt.Sprintf(format, v...)))
}

func (l *richLogger) Print(args ...any) {
	getWriter().AccessRecord(fmt.Sprint(args...))
}

func (l *richLogger) Printf(format string, args ...any) {
	getWriter().AccessRecord(fmt.Sprintf(format, args...))
}

func (l *richLogger) WithTraceId(traceId string) Logger {
	return &richLogger{
		moduleName: l.moduleName,
		traceId:    traceId,
	}
}

func (l *richLogger) WriteRawString(msg string) {
	getWriter().WriteRawString(msg)
}

func (l *richLogger) formatMessage(msg string) string {
	if l.traceId != "" {
		return fmt.Sprintf("[%s] [%s] %s", l.moduleName, l.traceId, msg)
	}
	return fmt.Sprintf("[%s] %s", l.moduleName, msg)
}
