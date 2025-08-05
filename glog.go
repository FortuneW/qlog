package qlog

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/FortuneW/qlog/internal"
)

// levelMap 用于日志级别字符串映射
var levelMap = map[string]uint32{
	internal.LevelTrace:   internal.TraceLevel,
	internal.LevelDebug:   internal.DebugLevel,
	internal.LevelInfo:    internal.InfoLevel,
	internal.LevelWarn:    internal.WarnLevel,
	internal.LevelError:   internal.ErrorLevel,
	internal.LevelDisable: internal.DisableLevel,
}

// levelStrMap 用于日志级别数值到字符串的映射
var levelStrMap = map[uint32]string{
	internal.TraceLevel:   internal.LevelTrace,
	internal.DebugLevel:   internal.LevelDebug,
	internal.InfoLevel:    internal.LevelInfo,
	internal.WarnLevel:    internal.LevelWarn,
	internal.ErrorLevel:   internal.LevelError,
	internal.DisableLevel: internal.LevelDisable,
}

var (
	defaultLogLevel = "TRA"       // 保存默认日志级别
	openTimer       *time.Timer   // 定时器
	openTimeMutex   sync.Mutex    // 保护定时相关操作的互斥锁
	currentOpenTime time.Duration // 当前设置的超时时间
)

func InitWithConfig(config Config) error {
	// 首先验证配置
	if err := config.ValidateConfig(); err != nil {
		return fmt.Errorf("invalid config: %v", err)
	}

	// 转换为内部配置结构
	internalConfig := internal.LogConf{
		ServiceName:   config.ServiceName,
		ServerLogDir:  config.ServerLogDir,
		ManagerLogDir: config.ManagerLogDir,
		MaxBackups:    config.MaxBackups,
		MaxSize:       config.MaxSize,
		Level:         config.Level,
		Compress:      config.Compress,
		Rotation:      config.Rotation,
		Mode:          config.Mode,
		ColorConsole:  config.ColorConsole,
	}

	defaultLogLevel = config.Level

	defer func() {
		// 文件模式下也希望输出到控制台
		if config.ToConsole && config.Mode == modeFile {
			internal.AddWriter(internal.NewWriter(os.Stdout))
		}
	}()
	return internal.SetUp(internalConfig)
}

func UnInit() {
	_ = internal.Close()
}

// CheckLogLevelStr 检查日志级别字符串是否有效
func CheckLogLevelStr(level string) error {
	upperLevel := strings.ToUpper(level)
	if _, ok := levelMap[upperLevel]; !ok {
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

// GetLogLevelStr 获取日志级别对应的字符串
func GetLogLevelStr() string {
	if level, ok := levelStrMap[internal.GetLevel()]; ok {
		return level
	} else {
		return "INVALID"
	}
}

// SetLogLevelStr 设置日志级别
func SetLogLevelStr(level string) {
	if logLevel, ok := levelMap[strings.ToUpper(level)]; ok {
		internal.SetLevel(logLevel)
		mlog.Infof("set log level to %s", level)
	} else {
		mlog.Warnf("invalid log level: %s", level)
	}
}

var mlog = GetRLog("xlog")

// SetOpenTime 设置日志级别的临时提升时间
func SetOpenTime(duration time.Duration, callback func(level string)) error {
	openTimeMutex.Lock()
	defer openTimeMutex.Unlock()

	// 如果已经存在定时器，先停止它
	if openTimer != nil {
		openTimer.Stop()
		openTimer = nil
	}

	currentOpenTime = duration

	// 如果设置为0，保持当前日志级别永久生效
	if duration == 0 {
		mlog.Infof("log level (%s) set permanently", GetLogLevelStr())
		return nil
	}

	// 启动新的定时器，直接使用duration
	openTimer = time.AfterFunc(duration, func() {
		mlog.Infof("log level temporary elevation timeout, restoring (%s) to default level: %s", GetLogLevelStr(), defaultLogLevel)

		openTimeMutex.Lock()
		defer openTimeMutex.Unlock()

		SetLogLevelStr(defaultLogLevel)
		currentOpenTime = -1
		openTimer = nil

		if callback != nil {
			callback(defaultLogLevel)
		}
	})

	return nil
}

// GetOpenTime 获取当前设置的超时时间
// 返回值：
// - 如果设置了超时时间，返回设置的超时时间
// - 如果未设置超时时间或永久生效，返回0
// - 如果超时时间已过，返回-1
func GetOpenTime() time.Duration {
	openTimeMutex.Lock()
	defer openTimeMutex.Unlock()
	return currentOpenTime
}

// GetRLog 获取运日志实例
func GetRLog(moduleName string) RLogger {
	return &rLogger{rlog: internal.WithModuleName(moduleName)}
}

// GetALog 获取access日志实例，管理接口使用
func GetALog() ALogger {
	return alog
}

// GetELog 获取子进程日志实例，子进程使用
func GetELog(moduleName string) ELogger {
	return &eLogger{moduleName: moduleName}
}

// TimeTrackWithDebug 便于打印时间消耗
// 使用示例：defer TimeTrackWithDebug(mlog, "test")()
func TimeTrackWithDebug(logger RLogger, msg string) func() {
	start := time.Now()
	return func() {
		if logger != nil {
			elapsed := time.Since(start)
			logger.Debugf("Leave:%s,cost:%v", msg, elapsed)
		}
	}
}

func TimeTrackWithTrace(logger RLogger, msg string) func() {
	start := time.Now()
	return func() {
		if logger != nil {
			elapsed := time.Since(start)
			logger.Tracef("Leave:%s,cost:%v", msg, elapsed)
		}
	}
}

func TimeTrackWithInfo(logger RLogger, msg string) func() {
	start := time.Now()
	return func() {
		if logger != nil {
			elapsed := time.Since(start)
			logger.Infof("Leave:%s,cost:%v", msg, elapsed)
		}
	}
}
