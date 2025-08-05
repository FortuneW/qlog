package xlog

import "testing"

func TestGetRLog(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
		traceID    string
		message    string
	}{
		{
			name:       "基础日志测试",
			moduleName: "test_module",
			message:    "test message",
		},
		{
			name:       "带追踪ID的日志测试",
			moduleName: "test_module",
			traceID:    "trace_123",
			message:    "traced message",
		},
		{
			name:       "空模块名测试",
			moduleName: "",
			message:    "empty module message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 获取日志实例
			rlog := GetRLog(tt.moduleName)
			if rlog == nil {
				t.Fatal("GetRLog返回了nil")
			}

			// 测试所有日志级别的基本方法
			rlog.Debug(tt.message)
			rlog.Debugf("%s - debug formatted", tt.message)

			rlog.Info(tt.message)
			rlog.Infof("%s - info formatted", tt.message)

			rlog.Warn(tt.message)
			rlog.Warnf("%s - warn formatted", tt.message)

			rlog.Error(tt.message)
			rlog.Errorf("%s - error formatted", tt.message)

			// 测试带TraceID的各个级别日志
			if tt.traceID != "" {
				tracedLogger := rlog.WithTraceId(tt.traceID)
				if tracedLogger == nil {
					t.Error("WithTraceId返回了nil")
				}

				tracedLogger.Debug(tt.message)
				tracedLogger.Debugf("%s - traced debug", tt.message)

				tracedLogger.Info(tt.message)
				tracedLogger.Infof("%s - traced info", tt.message)

				tracedLogger.Warn(tt.message)
				tracedLogger.Warnf("%s - traced warn", tt.message)

				tracedLogger.Error(tt.message)
				tracedLogger.Errorf("%s - traced error", tt.message)
			}
		})
	}
}
