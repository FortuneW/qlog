package internal

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func gzipFile(file string, fsys fileSystem) (err error) {
	in, err := fsys.Open(file)
	if err != nil {
		return err
	}

	defer func() {
		if e := fsys.Close(in); e != nil {
			Errorf("failed to close file: %s, error: %v", file, e)
		}
		if err == nil {
			// only remove the original file when compression is successful
			err = fsys.Remove(file)
		}
	}()

	gzipFile := fmt.Sprintf("%s%s", file, gzipExt)
	out, err := fsys.Create(gzipFile)
	if err != nil {
		return err
	}
	if err = out.Chmod(preGzipFileMode); err != nil {
		Errorf("failed to change gzip file mode: %s, error: %v", file, err)
	}

	defer func() {
		e := fsys.Close(out)
		if err == nil {
			err = e
		}
	}()

	w := gzip.NewWriter(out)
	if _, err = fsys.Copy(w, in); err != nil {
		// failed to copy, no need to close w
		return err
	}

	_ = out.Chmod(gzipFileMode)
	return fsys.Close(w)
}

// 确保有足够的磁盘空间
func ensureEnoughSpace(file string) error {
	dir := filepath.Dir(file)
	fileInfo, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	requiredSpace := fileInfo.Size()

	// 检查可用空间
	available, err := GetDirOnDiskTotalSize(dir)
	if err != nil {
		return fmt.Errorf("failed to get available disk space: %w", err)
	}

	// 如果空间足够，直接返回
	if available >= requiredSpace {
		return nil
	}

	// 空间不足，尝试清理旧文件
	Warnf("insufficient disk space (available: %d, required: %d), trying to clean up", available, requiredSpace)

	// 获取目录下所有日志文件
	pattern := fmt.Sprintf("%s.*", strings.TrimSuffix(file, filepath.Ext(file)))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list log files: %w", err)
	}

	// 按修改时间排序
	fileInfos := make([]struct {
		path string
		info os.FileInfo
	}, 0, len(files))

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			Warnf("failed to get file info for %s: %v", f, err)
			continue
		}
		fileInfos = append(fileInfos, struct {
			path string
			info os.FileInfo
		}{f, info})
	}

	// 按时间从旧到新排序
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].info.ModTime().Before(fileInfos[j].info.ModTime())
	})

	// 逐个删除旧文件，直到有足够空间
	for _, fi := range fileInfos {
		// 不要删除当前要压缩的文件
		if fi.path == file {
			continue
		}

		Infof("removing old log file to free space: %s", fi.path)
		if err := os.Remove(fi.path); err != nil {
			Warnf("failed to remove file %s: %v", fi.path, err)
			continue
		}

		// 重新检查空间是否足够
		available, err = GetDirOnDiskFreeSize(dir)
		if err != nil {
			return fmt.Errorf("failed to get available disk space: %w", err)
		}

		if available >= requiredSpace {
			Infof("cleaned up enough space for compression")
			return nil
		}
	}

	// 如果清理后仍然空间不足
	available, _ = GetDirOnDiskFreeSize(dir)
	if available < requiredSpace {
		return fmt.Errorf("still insufficient disk space after cleanup (available: %d, required: %d)", available, requiredSpace)
	}

	return nil
}
