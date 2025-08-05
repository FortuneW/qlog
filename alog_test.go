package qlog

import (
	"testing"
)

func TestALogger_Print(t *testing.T) {
	// 获取 ALog 实例并测试
	alog := GetALog()
	// 测试 Print 方法
	t.Run("测试Print方法", func(t *testing.T) {
		alog.Print("测试消息")
	})
}

func TestALogger_Printf(t *testing.T) {
	// 获取 ALog 实例并测试
	alog := GetALog()
	// 测试 Printf 方法
	t.Run("测试Printf方法", func(t *testing.T) {
		alog.Printf("测试消息 %s", "格式化")
	})
}
