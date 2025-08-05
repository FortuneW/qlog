package internal

import (
	"fmt"
	"os"
	"testing"
)

// 添加格式化函数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := unit, 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func TestGetDirOnDiskTotalSize(t *testing.T) {
	// 创建临时测试目录
	testDir, err := os.MkdirTemp("", "disk_test_*")
	if err != nil {
		t.Fatalf("无法创建临时测试目录: %v", err)
	}
	defer os.RemoveAll(testDir) // 测试结束后清理

	// 测试获取磁盘大小
	size, err := GetDirOnDiskTotalSize(testDir)
	if err != nil {
		t.Errorf("GetDirOnDiskTotalSize 失败: %v", err)
	} else {
		t.Logf("testDir: %s, 磁盘大小: %s", testDir, formatBytes(int64(size)))
	}

	// 验证返回的大小是否合理
	if size == 0 {
		t.Error("磁盘大小不应该为 0")
	}

	// 测试无效路径
	_, err = GetDirOnDiskTotalSize("/path/does/not/exist")
	if err == nil {
		t.Error("对于无效路径应该返回错误")
	}
}
