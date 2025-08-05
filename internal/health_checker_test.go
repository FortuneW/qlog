package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHealthChecker_LogRecovery(t *testing.T) {
	// 设置测试目录和文件
	//tmpDir := t.TempDir()
	tmpDir := "/log/limitDir"
	logFile := filepath.Join(tmpDir, "test.log")

	// 创建 RotateLogger
	logger, err := NewLogger(logFile, DefaultRotateRule(logFile, ".", 1, false), false)
	if err != nil {
		t.Fatalf("创建日志器失败: %v", err)
	}
	defer logger.Close()

	logger.Write([]byte(fmt.Sprintf("故障前写一条错误日志:%s\n", time.Now().Format(time.RFC3339Nano))))

	if false {
		// 模拟磁盘满，外部linux做个虚拟磁盘挂载到日志目录从而模拟磁盘满
	}

	if true {
		// 模拟文件损坏：关闭并删除文件
		logger.fp.Close()
		logger.fp = nil
		if err := os.Remove(logFile); err != nil {
			t.Fatalf("删除日志文件失败: %v", err)
		}
	}

	time.Sleep(time.Second * 1)

	logger.Write([]byte(fmt.Sprintf("故障后写一条错误日志:%s\n", time.Now().Format(time.RFC3339Nano))))

	// 等待恢复
	time.Sleep(time.Second * 4) // 给系统一些时间进行恢复

	// 再次写入，应该成功
	logger.Write([]byte(fmt.Sprintf("恢复后写一条错误日志:%s\n", time.Now().Format(time.RFC3339Nano))))

	// 验证文件内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	t.Logf("日志文件内容: %s", string(content))

	if len(content) == 0 {
		t.Error("日志文件不应该为空")
	}
}

func TestHealthChecker_MultipleErrors(t *testing.T) {
	// 设置测试目录和文件
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test_multiple.log")

	// 创建 RotateLogger
	logger, err := NewLogger(logFile, DefaultRotateRule(logFile, ".", 1, false), false)
	if err != nil {
		t.Fatalf("创建日志器失败: %v", err)
	}
	defer logger.Close()

	// 测试连续报告多个错误
	testErr := errors.New("测试错误")
	for i := 0; i < 3; i++ {
		logger.health.ReportError(testErr)
		time.Sleep(time.Millisecond * 100)
	}

	// 验证最后一次错误被正确记录
	if logger.health.lastError != testErr {
		t.Errorf("期望错误 %v, 得到 %v", testErr, logger.health.lastError)
	}
}
