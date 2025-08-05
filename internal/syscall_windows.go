//go:build windows

package internal

import (
	"os"
	"syscall"
)

func CloseOnExec(file *os.File) {
	if file != nil {
		syscall.CloseOnExec(syscall.Handle(file.Fd()))
	}
}
