package xlog

import (
	"strings"
	"testing"
	"time"

	"github.com/FortuneW/qlog/internal"
)

func TestELogger_LogLevels(t *testing.T) {
	eLog := &eLogger{moduleName: "elog_test_module"}

	tests := []struct {
		name    string
		logFunc func(args ...interface{})
		level   uint32
		message string
	}{
		{
			name:    "Debug日志测试",
			logFunc: eLog.Debug,
			level:   internal.DebugLevel,
			message: "debug message",
		},
		{
			name:    "Info日志测试",
			logFunc: eLog.Info,
			level:   internal.InfoLevel,
			message: "info message",
		},
		{
			name:    "Warn日志测试",
			logFunc: eLog.Warn,
			level:   internal.WarnLevel,
			message: "warn message",
		},
		{
			name:    "Error日志测试",
			logFunc: eLog.Error,
			level:   internal.ErrorLevel,
			message: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(tt.message)

			select {
			case logItem := <-eLog.GetPopELogItemChannel():
				assertLogItem(t, logItem, tt.level, "elog_test_module", tt.message, "")
				WriteELogItem(logItem)
			case <-time.After(time.Second):
				t.Error("获取日志超时")
			}
		})
	}
}

func TestELogger_FormatLogLevels(t *testing.T) {
	logger := eLogger{moduleName: "elog_test_module"}

	tests := []struct {
		name    string
		logFunc func(format string, args ...interface{})
		level   uint32
		format  string
		args    []interface{}
		want    string
	}{
		{
			name:    "Debugf日志测试",
			logFunc: logger.Debugf,
			level:   internal.DebugLevel,
			format:  "debug message %s",
			args:    []interface{}{"test"},
			want:    "debug message test",
		},
		{
			name:    "Infof日志测试",
			logFunc: logger.Infof,
			level:   internal.InfoLevel,
			format:  "info message %s",
			args:    []interface{}{"test"},
			want:    "info message test",
		},
		{
			name:    "Warnf日志测试",
			logFunc: logger.Warnf,
			level:   internal.WarnLevel,
			format:  "warn message %s",
			args:    []interface{}{"test"},
			want:    "warn message test",
		},
		{
			name:    "Errorf日志测试",
			logFunc: logger.Errorf,
			level:   internal.ErrorLevel,
			format:  "error message %s",
			args:    []interface{}{"test"},
			want:    "error message test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(tt.format, tt.args...)

			select {
			case logItem := <-logger.GetPopELogItemChannel():
				assertLogItem(t, logItem, tt.level, "elog_test_module", tt.want, "")
				WriteELogItem(logItem)
			case <-time.After(time.Second):
				t.Error("获取日志超时")
			}
		})
	}
}

func TestELogger_WithTraceId(t *testing.T) {
	baseLogger := eLogger{moduleName: "elog_test_module"}
	tracedLogger := baseLogger.WithTraceId("trace123")

	tests := []struct {
		name    string
		logFunc func(args ...interface{})
		level   uint32
		message string
	}{
		{
			name:    "带TraceId的Debug日志测试",
			logFunc: tracedLogger.Debug,
			level:   internal.DebugLevel,
			message: "debug message",
		},
		{
			name:    "带TraceId的Info日志测试",
			logFunc: tracedLogger.Info,
			level:   internal.InfoLevel,
			message: "info message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(tt.message)

			select {
			case logItem := <-baseLogger.GetPopELogItemChannel():
				assertLogItem(t, logItem, tt.level, "elog_test_module", tt.message, "trace123")
				WriteELogItem(logItem)
			case <-time.After(time.Second):
				t.Error("获取日志超时")
			}
		})
	}
}

func TestELogger_WithTraceIdFormat(t *testing.T) {
	baseLogger := eLogger{moduleName: "elog_test_module"}
	tracedLogger := baseLogger.WithTraceId("trace123")

	tests := []struct {
		name    string
		logFunc func(format string, args ...interface{})
		level   uint32
		format  string
		args    []interface{}
		want    string
	}{
		{
			name:    "带TraceId的Debugf日志测试",
			logFunc: tracedLogger.Debugf,
			level:   internal.DebugLevel,
			format:  "debug message %s",
			args:    []interface{}{"test"},
			want:    "debug message test",
		},
		{
			name:    "带TraceId的Infof日志测试",
			logFunc: tracedLogger.Infof,
			level:   internal.InfoLevel,
			format:  "info message %s",
			args:    []interface{}{"test"},
			want:    "info message test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(tt.format, tt.args...)

			select {
			case logItem := <-baseLogger.GetPopELogItemChannel():
				assertLogItem(t, logItem, tt.level, "elog_test_module", tt.want, "trace123")
				WriteELogItem(logItem)
			case <-time.After(time.Second):
				t.Error("获取日志超时")
			}
		})
	}
}

func assertLogItem(t *testing.T, logItem *ELogItem, expectedLevel uint32, expectedModule, expectedMessage, expectedTraceId string) {
	if logItem.Level != expectedLevel {
		t.Errorf("日志级别不匹配, 期望 %d, 得到 %d", expectedLevel, logItem.Level)
	}

	if !strings.Contains(logItem.Content, "["+expectedModule+"]") {
		t.Errorf("日志缺少模块名称\n得到: %s", logItem.Content)
	}

	if !strings.Contains(logItem.Content, expectedMessage) {
		t.Errorf("日志缺少预期消息\n预期包含: %s\n得到: %s", expectedMessage, logItem.Content)
	}

	if expectedTraceId != "" {
		if !strings.Contains(logItem.Content, "["+expectedTraceId+"]") {
			t.Errorf("日志缺少TraceId\n预期包含: [%s]\n得到: %s", expectedTraceId, logItem.Content)
		}
	}
}

func TestELogger_ChannelBuffer(t *testing.T) {
	eLog := GetELog("elog_test_module")

	for i := 0; i < 1025; i++ {
		eLog.Info("test message")
	}

	eLog.Info("additional message")

	count := 0
	for {
		select {
		case <-eLog.GetPopELogItemChannel():
			count++
		case <-time.After(100 * time.Millisecond):
			if count == 0 {
				t.Error("通道为空")
			}
			return
		}
	}
}
