package internal

import (
	"fmt"
	"log"
	"os"
	"time"
)

// HealthChecker 负责日志系统的健康检查和恢复
type HealthChecker struct {
	healthChan  chan error    // 健康状态通道
	recoverChan chan struct{} // 恢复信号通道
	lastError   error         // 最后一次错误
	errorTime   time.Time     // 错误发生时间
	done        chan PlaceholderType
	logger      *RotateLogger
}

// NewHealthChecker 创建新的健康检查器
func NewHealthChecker(logger *RotateLogger) *HealthChecker {
	return &HealthChecker{
		healthChan:  make(chan error, 1),
		recoverChan: make(chan struct{}, 1),
		done:        logger.done,
		logger:      logger,
	}
}

// Start 启动健康检查器
func (h *HealthChecker) Start() {
	h.logger.waitGroup.Add(1)
	go func() {
		defer h.logger.waitGroup.Done()

		// 添加定时器，每秒检查一次文件状态
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case err := <-h.healthChan:
				h.lastError = err
				h.errorTime = time.Now()
				log.Println("health check error:", err, h.errorTime)
				h.tryRecover()
			case <-ticker.C:
				// 定期检查文件是否存在
				if _, err := os.Stat(h.logger.filename); os.IsNotExist(err) {
					h.ReportError(fmt.Errorf("log file does not exist: %v", err))
				}
			case <-h.done:
				return
			}
		}
	}()
}

// tryRecover 尝试恢复日志系统
func (h *HealthChecker) tryRecover() {
	retryCount := 0

	for {
		recoverMsg := GetOutputStringFormatted(LevelError, fmt.Sprintf("[qlog] system recovered from(%s:%q), outage duration: %v",
			h.errorTime.Format(timeFormat),
			h.lastError,
			time.Since(h.errorTime)),
		)
		if err := h.testWrite(recoverMsg); err == nil {
			// 恢复成功，发送恢复信号
			select {
			case h.recoverChan <- struct{}{}:
			default:
			}
			return
		}

		retryCount++
		time.Sleep(time.Second)

		select {
		case <-h.done:
			return
		default:
		}
	}
}

// testWrite 测试写入功能
func (h *HealthChecker) testWrite(msg string) error {
	if h.logger.fp == nil {
		if err := h.logger.initialize(); err != nil {
			return fmt.Errorf("initialization failed: %v", err)
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(h.logger.filename); os.IsNotExist(err) {
		// 只在文件指针非空时才尝试关闭
		if h.logger.fp != nil {
			if err := h.logger.fp.Close(); err != nil {
				return fmt.Errorf("failed to close old file: %v", err)
			}
			h.logger.fp = nil
		}
		if err := h.logger.initialize(); err != nil {
			return fmt.Errorf("re-initialization failed: %v", err)
		}
	}

	_, err := h.logger.fp.Write([]byte(msg))
	return err
}

// ReportError 报告错误
func (h *HealthChecker) ReportError(err error) {
	select {
	case h.healthChan <- err:
	default:
	}
}

// WaitRecover 等待恢复信号
func (h *HealthChecker) WaitRecover() bool {
	select {
	case <-h.recoverChan:
		return true
	case <-h.done:
		return false
	}
}
