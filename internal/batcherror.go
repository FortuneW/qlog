package internal

import (
	"errors"
	"sync"
)

// BatchError 是一个可以存储多个错误的错误类型
type BatchError struct {
	errs []error
	lock sync.RWMutex
}

// Add 添加一个或多个非空错误到 BatchError 实例中
func (be *BatchError) Add(errs ...error) {
	be.lock.Lock()
	defer be.lock.Unlock()

	for _, err := range errs {
		if err != nil {
			be.errs = append(be.errs, err)
		}
	}
}

// Err 返回一个代表所有累积错误的错误对象
// 如果没有错误则返回 nil
func (be *BatchError) Err() error {
	be.lock.RLock()
	defer be.lock.RUnlock()

	// 如果没有非空错误，errors.Join(...) 会返回 nil
	return errors.Join(be.errs...)
}

// NotNil 检查 BatchError 中是否至少存在一个错误
func (be *BatchError) NotNil() bool {
	be.lock.RLock()
	defer be.lock.RUnlock()

	return len(be.errs) > 0
}
