package internal

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	dateFormat           = "2006-01-02"
	hoursPerDay          = 24
	maxLogItemBufferSize = 1000 // 日志条目缓冲区大小
	defaultDirMode       = 0o755
	defaultFileMode      = 0o600
	gzipExt              = ".gz"
	megaBytes            = 1 << 20
	gzipFileMode         = 0o400
	preGzipFileMode      = 0o600
)

var (
	ErrLogFileClosed = errors.New("error: log file closed")
	fileTimeFormat   = "2006-01-02T15:04:05.000000000Z"
)

type (
	// RotateRule 接口用于定义日志轮转规则
	RotateRule interface {
		BackupFileName() string
		MarkRotated()
		OutdatedFiles() []string
		ShallRotate(size int64) bool
		FilePathPattern() string // 返回文件名pattern
	}

	// RotateLogger 是一个可以按照给定规则轮转日志文件的日志器
	RotateLogger struct {
		filename      string
		backup        string
		fp            *os.File
		channel       chan []byte
		done          chan PlaceholderType
		rule          RotateRule
		compress      bool
		retryCompress chan string
		// 不能使用 threading.RoutineGroup，因为会导致循环导入
		waitGroup   sync.WaitGroup
		closeOnce   sync.Once
		currentSize int64

		health *HealthChecker // 新增健康检查器
	}

	// DailyRotateRule 是一个按天轮转日志文件的规则
	DailyRotateRule struct {
		rotatedTime string
		filename    string
		delimiter   string
		days        int
		gzip        bool
	}

	// SizeLimitRotateRule 是一个基于文件大小的日志轮转规则
	SizeLimitRotateRule struct {
		DailyRotateRule
		maxSize    int64
		maxBackups int
	}
)

// DefaultRotateRule 返回默认的日志轮转规则，目前是 DailyRotateRule
func DefaultRotateRule(filename, delimiter string, days int, gzip bool) RotateRule {
	return &DailyRotateRule{
		rotatedTime: getNowDate(),
		filename:    filename,
		delimiter:   delimiter,
		days:        days,
		gzip:        gzip,
	}
}

// BackupFileName 返回轮转时的备份文件名
func (r *DailyRotateRule) BackupFileName() string {
	return fmt.Sprintf("%s%s%s", r.filename, r.delimiter, getNowDate())
}

// MarkRotated 将轮转时间标记为当前时间
func (r *DailyRotateRule) MarkRotated() {
	r.rotatedTime = getNowDate()
}

func (r *DailyRotateRule) FilePathPattern() string {
	return fmt.Sprintf("%s%s*", r.filename, r.delimiter)
}

// OutdatedFiles 返回超过保留天数的文件列表
func (r *DailyRotateRule) OutdatedFiles() []string {
	if r.days <= 0 {
		return nil
	}

	var pattern string
	if r.gzip {
		pattern = fmt.Sprintf("%s%s*%s", r.filename, r.delimiter, gzipExt)
	} else {
		pattern = fmt.Sprintf("%s%s*", r.filename, r.delimiter)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		Errorf("failed to delete outdated log files, error: %s", err)
		return nil
	}

	var buf strings.Builder
	boundary := time.Now().UTC().Add(-time.Hour * time.Duration(hoursPerDay*r.days)).Format(dateFormat)
	buf.WriteString(r.filename)
	buf.WriteString(r.delimiter)
	buf.WriteString(boundary)
	if r.gzip {
		buf.WriteString(gzipExt)
	}
	boundaryFile := buf.String()

	var outdates []string
	for _, file := range files {
		if file < boundaryFile {
			outdates = append(outdates, file)
		}
	}

	return outdates
}

// ShallRotate 检查文件是否应该进行轮转
func (r *DailyRotateRule) ShallRotate(_ int64) bool {
	return len(r.rotatedTime) > 0 && getNowDate() != r.rotatedTime
}

// NewSizeLimitRotateRule 返回一个基于大小限制的轮转规则
func NewSizeLimitRotateRule(filename, delimiter string, days, maxSize, maxBackups int, gzip bool) RotateRule {
	return &SizeLimitRotateRule{
		DailyRotateRule: DailyRotateRule{
			rotatedTime: getNowDateInRFC3339Format(),
			filename:    filename,
			delimiter:   delimiter,
			days:        days,
			gzip:        gzip,
		},
		maxSize:    int64(maxSize) * megaBytes,
		maxBackups: maxBackups,
	}
}

func (r *SizeLimitRotateRule) BackupFileName() string {
	dir := filepath.Dir(r.filename)
	prefix, ext := r.parseFilename()
	timestamp := getNowDateInRFC3339Format()
	timestamp = strings.Replace(timestamp, ":", ".", -1)
	return filepath.Join(dir, fmt.Sprintf("%s%s%s%s", prefix, r.delimiter, timestamp, ext))
}

func (r *SizeLimitRotateRule) MarkRotated() {
	r.rotatedTime = getNowDateInRFC3339Format()
}

func (r *SizeLimitRotateRule) FilePathPattern() string {
	dir := filepath.Dir(r.filename)
	prefix, ext := r.parseFilename()
	return fmt.Sprintf("%s%s%s%s*%s", dir, string(filepath.Separator), prefix, r.delimiter, ext)
}

func (r *SizeLimitRotateRule) OutdatedFiles() []string {
	dir := filepath.Dir(r.filename)
	prefix, ext := r.parseFilename()

	var pattern string
	if r.gzip {
		pattern = fmt.Sprintf("%s%s%s%s*%s%s", dir, string(filepath.Separator),
			prefix, r.delimiter, ext, gzipExt)
	} else {
		pattern = fmt.Sprintf("%s%s%s%s*%s", dir, string(filepath.Separator),
			prefix, r.delimiter, ext)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		Errorf("failed to delete outdated log files, error: %s", err)
		return nil
	}

	sort.Strings(files)

	outdated := make(map[string]PlaceholderType)

	// 1. 检查备份数量是否超过限制
	if r.maxBackups > 0 && len(files) > r.maxBackups {
		for _, f := range files[:len(files)-r.maxBackups] {
			outdated[f] = Placeholder
		}
		files = files[len(files)-r.maxBackups:] // 更新files列表为保留的文件
	}

	// 2. 检查总大小是否超过限制
	var (
		totalSize   int64
		guessGzSize int64 // 可能压缩后的文件大小
	)
	fileSizes := make(map[string]int64)
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			Warnf("failed to get file size: %s, error: %s", f, err)
			continue
		}
		totalSize += info.Size()
		fileSizes[f] = info.Size()
		if info.Size() > guessGzSize {
			guessGzSize = info.Size()
		}
	}

	freeSize, err := GetDirOnDiskFreeSize(dir)
	if err == nil && len(files) > 0 {
		if r.gzip {
			// 剩余空间不足后大小也补充进去希望释多放出一些空间
			if guessGzSize > freeSize {
				totalSize += guessGzSize * 2
			}
		}
	} else {
		freeSize = math.MaxInt64
	}

	maxTotalSize := int64(math.MaxInt64)
	if diskTotalSize, err := GetDirOnDiskTotalSize(dir); err != nil || diskTotalSize == 0 {
		// 获取不到磁盘实际大小，也用配置的个数和大小兜底磁盘容量限制
		maxTotalSize = int64(float64(r.maxSize*int64(r.maxBackups)) * 0.8)
	} else {
		// 滚卷时保证归档文件占用磁盘容量不超过磁盘大小80%阈值
		maxTotalSize = int64(float64(diskTotalSize) * 0.8)
	}

	if r.maxSize > 0 && r.maxBackups > 0 {
		if totalSize > maxTotalSize || freeSize < guessGzSize {
			for _, f := range files {
				if _, exists := outdated[f]; !exists {
					outdated[f] = Placeholder
					totalSize -= fileSizes[f]
				}
				if totalSize <= maxTotalSize {
					break
				}
			}
		}
	}

	// 3. 检查是否有过期的文件（按天数）
	if r.days > 0 {
		boundary := time.Now().UTC().Add(-time.Hour * time.Duration(hoursPerDay*r.days)).Format(fileTimeFormat)
		boundary = strings.Replace(boundary, ":", ".", -1)
		boundaryFile := filepath.Join(dir, fmt.Sprintf("%s%s%s%s", prefix, r.delimiter, boundary, ext))
		if r.gzip {
			boundaryFile += gzipExt
		}
		for _, f := range files {
			if f >= boundaryFile {
				break
			}
			outdated[f] = Placeholder
		}
	}

	var result []string
	for k := range outdated {
		result = append(result, k)
	}
	return result
}

func (r *SizeLimitRotateRule) ShallRotate(size int64) bool {
	return r.maxSize > 0 && r.maxSize < size
}

func (r *SizeLimitRotateRule) parseFilename() (prefix, ext string) {
	logName := filepath.Base(r.filename)
	ext = filepath.Ext(r.filename)
	prefix = logName[:len(logName)-len(ext)]
	return
}

// NewLogger 返回一个 RotateLogger 实例，给定文件名和规则等
func NewLogger(filename string, rule RotateRule, compress bool) (*RotateLogger, error) {
	l := &RotateLogger{
		filename:      filename,
		channel:       make(chan []byte, maxLogItemBufferSize),
		retryCompress: make(chan string, 100),
		done:          make(chan PlaceholderType),
		rule:          rule,
		compress:      compress,
	}
	if err := l.initialize(); err != nil {
		return nil, err
	}

	l.health = NewHealthChecker(l)
	l.health.Start()
	l.startWorker()
	return l, nil
}

// Close 关闭 RotateLogger
func (l *RotateLogger) Close() error {
	var err error

	l.closeOnce.Do(func() {
		close(l.done)
		l.waitGroup.Wait()

		if err = l.fp.Sync(); err != nil {
			return
		}

		err = l.fp.Close()
	})

	return err
}

func (l *RotateLogger) Write(data []byte) (int, error) {
	select {
	case l.channel <- data:
		return len(data), nil
	case <-l.done:
		log.Println(string(data))
		return 0, ErrLogFileClosed
	default:
		// 故障的时候channel满之后的日志丢弃
		return 0, nil
	}
}

func (l *RotateLogger) getBackupFilename() string {
	if len(l.backup) == 0 {
		return l.rule.BackupFileName()
	}

	return l.backup
}

func (l *RotateLogger) initialize() error {
	l.backup = l.rule.BackupFileName()

	if fileInfo, err := os.Stat(l.filename); err != nil {
		basePath := path.Dir(l.filename)
		if _, err = os.Stat(basePath); err != nil {
			if err = os.MkdirAll(basePath, defaultDirMode); err != nil {
				return err
			}
		}

		if l.fp, err = os.Create(l.filename); err != nil {
			return err
		}
		if err = l.fp.Chmod(defaultFileMode); err != nil {
			return err
		}
	} else {
		if l.fp, err = os.OpenFile(l.filename, os.O_APPEND|os.O_WRONLY, defaultFileMode); err != nil {
			return err
		}

		l.currentSize = fileInfo.Size()
	}

	CloseOnExec(l.fp)

	return nil
}

func (l *RotateLogger) maybeCompressFile(file string) (retry bool) {
	if !l.compress {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			writeError(r)
		}
	}()

	if _, err := os.Stat(file); err != nil {
		// file doesn't exist or another error, ignore compression
		return
	}

	if ok := l.compressLogFile(file); !ok {
		retry = true
	}
	return
}

func (l *RotateLogger) maybeDeleteOutdatedFiles() {
	defer func() { recover() }()
	files := l.rule.OutdatedFiles()
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			Errorf("failed to remove outdated or limited file: %s", file)
		} else {
			Infof("delete outdated or limited file: %s", file)
		}
	}
}

func (l *RotateLogger) postRotate(file string) {
	go func() {
		if retry := l.maybeCompressFile(file); retry {
			l.retryCompress <- file
		}
		l.maybeDeleteOutdatedFiles()
	}()
}

func (l *RotateLogger) rotate() error {
	if l.fp != nil {
		err := l.fp.Close()
		l.fp = nil
		if err != nil {
			return err
		}
	}

	_, err := os.Stat(l.filename)
	if err == nil && len(l.backup) > 0 {
		backupFilename := l.getBackupFilename()
		err = os.Rename(l.filename, backupFilename)
		if err != nil {
			return err
		}

		l.postRotate(backupFilename)
	}

	l.backup = l.rule.BackupFileName()
	if l.fp, err = os.Create(l.filename); err == nil {
		CloseOnExec(l.fp)
		err = l.fp.Chmod(defaultFileMode)
	}

	return err
}

func (l *RotateLogger) startWorker() {
	l.waitGroup.Add(2)

	go func() {
		defer l.waitGroup.Done()

		for {
			select {
			case event := <-l.channel:
				l.write(event)
			case <-l.done:
				// avoid losing logs before closing.
				for {
					select {
					case event := <-l.channel:
						l.write(event)
					default:
						return
					}
				}
			}
		}
	}()

	go func() {
		defer l.waitGroup.Done()

		timer := time.NewTicker(time.Minute)

		l.compressUncompressedFiles()

		for {
			select {
			case filePath := <-l.retryCompress:
				Warnf("clean some old files and retry compress file: %s", filePath)
				l.maybeDeleteOutdatedFiles()
				l.compressLogFile(filePath)
			case <-timer.C:
				// 兜底一下每分钟也清理过期文件
				l.maybeDeleteOutdatedFiles()
			case <-l.done:
				return
			}
		}
	}()
}

func (l *RotateLogger) write(v []byte) {
	for {
		if l.rule.ShallRotate(l.currentSize + int64(len(v))) {
			if err := l.rotate(); err != nil {
				log.Println(err)
			} else {
				l.rule.MarkRotated()
				l.currentSize = 0
			}
		}

		if l.fp != nil {
			_, err := l.fp.Write(v)
			if err == nil {
				l.currentSize += int64(len(v))
				return
			}
			l.health.ReportError(err)
		}

		if !l.health.WaitRecover() {
			return
		}
	}
}

func (l *RotateLogger) compressLogFile(file string) bool {
	start := time.Now()
	Infof("compressing log file: %s", file)

	if err := gzipFile(file, fileSys); err != nil {
		Errorf("compress error: %s", err)
		return false
	} else {
		Infof("compressed log file: %s, took %s", file, time.Since(start))
		return true
	}
}

func getNowDate() string {
	return time.Now().UTC().Format(dateFormat)
}

func getNowDateInRFC3339Format() string {
	return time.Now().UTC().Format(fileTimeFormat)
}

// 处理一下遗留的未压缩文件
func (l *RotateLogger) compressUncompressedFiles() {
	if !l.compress {
		return
	}
	defer func() { recover() }()

	pattern := l.rule.FilePathPattern()
	// 查找所有可能的日志文件
	files, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	// 检查每个文件是否需要压缩
	for _, file := range files {
		if file == l.filename || strings.HasSuffix(file, gzipExt) {
			continue
		}
		go func(f string) {
			l.retryCompress <- f
		}(file)
	}
}
