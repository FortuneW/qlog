//go:build !windows

package internal

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// GetDirOnDiskTotalSize 返回给定目录所在磁盘的总大小（单位：字节）
func GetDirOnDiskTotalSize(path string) (int64, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, fmt.Errorf("failed to get disk status: %v", err)
	}

	// 检查值是否为负数
	if stat.Blocks < 0 || stat.Bsize < 0 {
		return 0, fmt.Errorf("invalid disk size: blocks=%d, bsize=%d", stat.Blocks, stat.Bsize)
	}

	// 总空间 = 块大小 * 总块数
	total := int64(stat.Blocks) * int64(stat.Bsize)
	return total, nil
}

// GetDirOnDiskFreeSize 返回给定目录所在磁盘的可用空间大小（单位：字节）
func GetDirOnDiskFreeSize(path string) (int64, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, fmt.Errorf("failed to get disk status: %v", err)
	}

	// 检查值是否为负数
	if stat.Bavail < 0 || stat.Bsize < 0 {
		return 0, fmt.Errorf("invalid disk size: bavail=%d, bsize=%d", stat.Bavail, stat.Bsize)
	}

	// 可用空间 = 块大小 * 可用块数
	free := int64(stat.Bavail) * int64(stat.Bsize)
	return free, nil
}
