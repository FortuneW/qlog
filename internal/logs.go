package internal

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	timeFormat       = "2006-01-02T15:04:05.000Z"
	maxContentLength uint32
	logLevel         uint32
	options          logOptions
	writer           = new(atomicWriter)
	setupOnce        sync.Once
)

type (
	// LogOption 定义了自定义日志配置的方法
	LogOption func(options *logOptions)

	logOptions struct {
		gzipEnabled  bool
		keepDays     int
		maxBackups   int
		maxSize      int
		rotationRule string
	}
)

// AddWriter 添加一个新的日志写入器
// 如果已经存在写入器，新的写入器将被添加到写入器链中
// 例如，要同时将日志写入文件和控制台，如果已经存在文件写入器：
// xlog.AddWriter(xlog.NewWriter(os.Stdout))
func AddWriter(w Writer) {
	ow := Reset()
	if ow == nil {
		SetWriter(w)
	} else {
		// no need to check if the existing writer is a comboWriter,
		// because it is not common to add more than one writer.
		// even more than one writer, the behavior is the same.
		SetWriter(comboWriter{
			writers: []Writer{ow, w},
		})
	}
}

// Close 关闭日志系统
func Close() error {
	if w := writer.Swap(nil); w != nil {
		return w.(io.Closer).Close()
	}

	return nil
}

// Trace 将参数写入调试日志
func Trace(v ...any) {
	if shallLog(TraceLevel) {
		getWriter().Debug(fmt.Sprint(v...))
	}
}

// Tracef 将参数写入调试日志
func Tracef(format string, v ...any) {
	if shallLog(TraceLevel) {
		getWriter().Debug(fmt.Sprintf(format, v...))
	}
}

// Debug 将参数写入调试日志
func Debug(v ...any) {
	if shallLog(DebugLevel) {
		getWriter().Debug(fmt.Sprint(v...))
	}
}

// Debugf 将参数写入调试日志
func Debugf(format string, v ...any) {
	if shallLog(DebugLevel) {
		getWriter().Debug(fmt.Sprintf(format, v...))
	}
}

// Error 将参数写入错误日志
func Error(v ...any) {
	if shallLog(ErrorLevel) {
		getWriter().Error(fmt.Sprint(v...))
	}
}

// Errorf 将参数写入错误日志
func Errorf(format string, v ...any) {
	if shallLog(ErrorLevel) {
		getWriter().Error(fmt.Errorf(format, v...).Error())
	}
}

// Warn 将参数写入警告日志
func Warn(v ...any) {
	if shallLog(WarnLevel) {
		getWriter().Warn(fmt.Sprint(v...))
	}
}

// Warnf 将参数写入警告日志
func Warnf(format string, v ...any) {
	if shallLog(WarnLevel) {
		getWriter().Warn(fmt.Errorf(format, v...).Error())
	}
}

// Info 将参数写入访问日志
func Info(v ...any) {
	if shallLog(InfoLevel) {
		getWriter().Info(fmt.Sprint(v...))
	}
}

// Infof 将参数写入访问日志
func Infof(format string, v ...any) {
	if shallLog(InfoLevel) {
		getWriter().Info(fmt.Sprintf(format, v...))
	}
}

// Reset 清除写入器并重置日志级别
func Reset() Writer {
	return writer.Swap(nil)
}

// SetLevel 设置日志级别，可用于抑制某些日志的输出
func SetLevel(level uint32) {
	atomic.StoreUint32(&logLevel, level)
}

// GetLevel 获取日志级别
func GetLevel() uint32 {
	return atomic.LoadUint32(&logLevel)
}

// SetWriter 设置日志写入器，可用于自定义日志配置
func SetWriter(w Writer) {
	if atomic.LoadUint32(&logLevel) != DisableLevel {
		writer.Store(w)
	}
}

// SetUp 初始化日志系统
// 如果已经初始化过，返回 nil
func SetUp(c LogConf) (err error) {
	// Ignore the later SetUp calls.
	// Because multiple services in one process might call SetUp respectively.
	// Need to wait for the first caller to complete the execution.
	setupOnce.Do(func() {
		setupLogLevel(c)

		atomic.StoreUint32(&maxContentLength, c.MaxContentLength)

		switch c.Mode {
		case fileMode:
			err = setupWithFiles(c)
		default:
			setupWithConsole(&c)
		}
	})

	return
}

// WithKeepDays 自定义日志保留天数
func WithKeepDays(days int) LogOption {
	return func(opts *logOptions) {
		opts.keepDays = days
	}
}

// WithGzip 自定义日志文件自动使用 gzip 压缩
func WithGzip() LogOption {
	return func(opts *logOptions) {
		opts.gzipEnabled = true
	}
}

// WithMaxBackups 自定义保留的日志文件备份数量
func WithMaxBackups(count int) LogOption {
	return func(opts *logOptions) {
		opts.maxBackups = count
	}
}

// WithMaxSize 自定义单个日志文件的最大大小（MB）
func WithMaxSize(size int) LogOption {
	return func(opts *logOptions) {
		opts.maxSize = size
	}
}

// WithRotation 自定义使用的日志轮转规则
func WithRotation(r string) LogOption {
	return func(opts *logOptions) {
		opts.rotationRule = r
	}
}

func createOutput(path string) (io.WriteCloser, error) {
	if len(path) == 0 {
		return nil, ErrLogPathNotSet
	}

	var rule RotateRule
	switch options.rotationRule {
	case sizeRotationRule:
		rule = NewSizeLimitRotateRule(path, backupFileDelimiter, options.keepDays, options.maxSize,
			options.maxBackups, options.gzipEnabled)
	default:
		rule = DefaultRotateRule(path, backupFileDelimiter, options.keepDays, options.gzipEnabled)
	}

	return NewLogger(path, rule, options.gzipEnabled)
}

func encodeError(err error) (ret string) {
	return encodeWithRecover(err, func() string {
		return err.Error()
	})
}

func encodeStringer(v fmt.Stringer) (ret string) {
	return encodeWithRecover(v, func() string {
		return v.String()
	})
}

func encodeWithRecover(arg any, fn func() string) (ret string) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(arg); v.Kind() == reflect.Ptr && v.IsNil() {
				ret = nilAngleString
			} else {
				ret = fmt.Sprintf("panic: %v", err)
			}
		}
	}()

	return fn()
}

func getWriter() Writer {
	w := writer.Load()
	if w == nil {
		w = writer.StoreIfNil(newEmptyWriter())
	}

	return w
}

func handleOptions(opts []LogOption) {
	for _, opt := range opts {
		opt(&options)
	}
}

func setupLogLevel(c LogConf) {
	switch strings.ToUpper(c.Level) {
	case LevelTrace:
		SetLevel(TraceLevel)
	case LevelDebug:
		SetLevel(DebugLevel)
	case LevelInfo:
		SetLevel(InfoLevel)
	case LevelWarn:
		SetLevel(WarnLevel)
	case LevelError:
		SetLevel(ErrorLevel)
	case LevelDisable:
		SetLevel(DisableLevel)
	}
}

func setupWithConsole(c *LogConf) {
	if c.ColorConsole {
		SetWriter(newColorConsoleWriter())
	} else {
		SetWriter(newConsoleWriter())
	}
}

func setupWithFiles(c LogConf) error {
	w, err := newFileWriter(c)
	if err != nil {
		return err
	}

	SetWriter(w)
	return nil
}

func shallLog(level uint32) bool {
	return atomic.LoadUint32(&logLevel) <= level
}

func writeError(val any) {
	getWriter().Error(val)
}
