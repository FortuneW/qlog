package qlog

// ALogger 定义了基础管理访问日志接口，仅包含时间，不包含日志级别的输出
// 提供最基本的Print和Printf方法
type ALogger interface {
	// Print 打印无格式日志
	Print(args ...interface{})
	// Printf 打印格式化日志
	Printf(format string, args ...interface{})
}

// RLogger 定义了常用的日志级别接口
// 包含Debug、Info、Warn、Error四个日志级别
// 支持链式调用设置traceId
type RLogger interface {
	// Trace 打印追踪级别日志
	Trace(args ...interface{})
	// Debug 打印调试级别日志
	Debug(args ...interface{})
	// Info 打印信息级别日志
	Info(args ...interface{})
	// Warn 打印警告级别日志
	Warn(args ...interface{})
	// Error 打印错误级别日志
	Error(args ...interface{})

	// Tracef 打印格式化的追踪级别日志
	Tracef(format string, args ...interface{})
	// Debugf 打印格式化的调试级别日志
	Debugf(format string, args ...interface{})
	// Infof 打印格式化的信息级别日志
	Infof(format string, args ...interface{})
	// Warnf 打印格式化的警告级别日志
	Warnf(format string, args ...interface{})
	// Errorf 打印格式化的错误级别日志
	Errorf(format string, args ...interface{})

	// WithTraceId 设置日志追踪ID
	// 返回设置了traceId的新logger实例，支持链式调用
	WithTraceId(traceId string) RLogger
}

// ELogger 扩展了RLogger接口
// 添加了日志传输通道相关功能
type ELogger interface {
	RLogger

	// GetPopELogItemChannel 获取日志传输通道
	// 用于子进程将日志项发送到主进程
	// 返回一个channel用于传输ELogItem
	GetPopELogItemChannel() chan *ELogItem
}
