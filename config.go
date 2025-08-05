package xlog

import "fmt"

// Config 日志配置
type Config struct {
	ServiceName   string // 服务名
	ServerLogDir  string // 服务日志目录
	ManagerLogDir string // 管理日志目录
	MaxBackups    int    // 最大备份数量
	MaxSize       int    // 单个日志文件最大尺寸(MB)
	KeepDays      int    // 日志保留天数
	Level         string // 日志级别 (DEB/INF/WAR/ERR/OFF)
	Compress      bool   // 是否压缩
	Rotation      string // 轮转方式 ("size"/"time")
	Mode          string // 日志模式 ("file"/"console")
	ToConsole     bool   // 是否输出到控制台,即使file模式
	ColorConsole  bool   // 仅console有效
}

const (
	// 日志大小限制
	minLogSize = 1    // 最小日志文件大小(MB)
	maxLogSize = 1024 // 最大日志文件大小(MB)

	// 备份数量限制
	minBackups = 1   // 最小备份数量
	maxBackups = 100 // 最大备份数量

	// 合法的日志级别
	validLogLevels = "DEB,INF,WAR,ERR,OFF"

	// 合法的轮转方式
	rotationSize = "size"
	rotationTime = "time"

	// 合法的日志模式
	modeFile    = "file"
	modeConsole = "console"
)

// ValidateConfig 验证日志配置是否合法
func (c *Config) ValidateConfig() error {
	// 验证日志目录
	if c.Mode == modeFile {
		if c.ServerLogDir == "" {
			return fmt.Errorf("server log directory cannot be empty in file mode")
		}
		if c.ManagerLogDir == "" {
			return fmt.Errorf("manager log directory cannot be empty in file mode")
		}
	}

	// 验证日志大小
	if c.MaxSize > 0 && (c.MaxSize < minLogSize || c.MaxSize > maxLogSize) {
		return fmt.Errorf("invalid max size: %d, should be between %d and %d MB",
			c.MaxSize, minLogSize, maxLogSize)
	}

	// 验证备份数量
	if c.MaxBackups > 0 && (c.MaxBackups < minBackups || c.MaxBackups > maxBackups) {
		return fmt.Errorf("invalid max backups: %d, should be between %d and %d",
			c.MaxBackups, minBackups, maxBackups)
	}

	// 验证日志级别
	if len(c.Level) > 0 {
		if err := CheckLogLevelStr(c.Level); err != nil {
			return fmt.Errorf("invalid log level: %s, valid levels are: %s",
				c.Level, validLogLevels)
		}
	}

	// 验证轮转方式
	if len(c.Rotation) > 0 {
		if c.Rotation != rotationSize && c.Rotation != rotationTime {
			return fmt.Errorf("invalid rotation: %s, should be either '%s' or '%s'",
				c.Rotation, rotationSize, rotationTime)
		}
	}

	// 验证日志模式
	if len(c.Mode) > 0 {
		if c.Mode != modeFile && c.Mode != modeConsole {
			return fmt.Errorf("invalid mode: %s, should be either '%s' or '%s'",
				c.Mode, modeFile, modeConsole)
		}
	}

	return nil
}
