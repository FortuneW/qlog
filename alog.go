package qlog

import "github.com/FortuneW/qlog/internal"

var alog = internal.WithModuleName("")

// 服务的管理日志
type aLogger struct {
}

func (a aLogger) Print(args ...interface{}) {
	alog.Print(args...)
}

func (a aLogger) Printf(format string, args ...interface{}) {
	alog.Printf(format, args...)
}
