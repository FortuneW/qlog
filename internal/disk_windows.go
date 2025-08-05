//go:build windows

package internal

import (
	"fmt"
	"syscall"
	"unsafe"
)

// GetDirOnDiskTotalSize 返回给定目录所在磁盘的总大小(字节)
func GetDirOnDiskTotalSize(path string) (int64, error) {
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return 0, fmt.Errorf("failed to load kernel32.dll: %v", err)
	}

	GetDiskFreeSpaceEx, err := kernel32.FindProc("GetDiskFreeSpaceExW")
	if err != nil {
		return 0, fmt.Errorf("failed to find GetDiskFreeSpaceExW procedure: %v", err)
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes int64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("failed to convert path: %v", err)
	}

	ret, _, err := GetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return 0, fmt.Errorf("failed to get disk space information: %v", err)
	}

	return totalBytes, nil
}

// GetDirOnDiskFreeSize 返回给定目录所在磁盘的可用空间大小(字节)
func GetDirOnDiskFreeSize(path string) (int64, error) {
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return 0, fmt.Errorf("failed to load kernel32.dll: %v", err)
	}

	GetDiskFreeSpaceEx, err := kernel32.FindProc("GetDiskFreeSpaceExW")
	if err != nil {
		return 0, fmt.Errorf("failed to find GetDiskFreeSpaceExW procedure: %v", err)
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes int64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("failed to convert path: %v", err)
	}

	ret, _, err := GetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return 0, fmt.Errorf("failed to get disk space information: %v", err)
	}

	return freeBytesAvailable, nil
}
