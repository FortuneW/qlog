package internal

// LogConf 是日志配置结构体
type LogConf struct {
	// ServiceName 表示服务名称
	ServiceName string `json:",optional"`
	// Mode 表示日志模式，默认为 `console`
	// console: 输出到控制台
	// file: 输出到文件
	Mode string `json:",default=console,options=[console,file]"`
	// ServerLogDir 表示服务日志目录路径，默认为 `logs`
	ServerLogDir string `json:",default=logs"`
	// ManagerLogDir 表示管理日志目录路径，默认为 `logs`
	ManagerLogDir string `json:",default=logs"`
	// Level 表示日志级别，默认为 `ERR`
	Level string `json:",default=ERR,options=[DEB,INF,WAR,ERR]"`
	// MaxContentLength 表示最大内容字节数，默认无限制
	MaxContentLength uint32 `json:",optional"`
	// Compress 表示是否压缩日志文件，默认为 `false`
	Compress bool `json:",optional"`
	// KeepDays 表示日志文件保留天数，默认保留所有文件
	// 仅在 Mode 为 `file` 时生效，对 Rotation 为 `daily` 或 `size` 都有效
	KeepDays int `json:",optional"`
	// MaxBackups 表示要保留的备份日志文件数量，0表示永久保留所有文件
	// 仅在 RotationRuleType 为 `size` 时生效
	// 即使 MaxBackups 设置为0，如果达到 KeepDays 限制，日志文件仍会被删除
	MaxBackups int `json:",default=0"`
	// MaxSize 表示正在写入的日志文件可占用的最大空间，0表示无限制，单位为MB
	// 仅在 RotationRuleType 为 `size` 时生效
	MaxSize int `json:",default=0"`
	// Rotation 表示日志轮转规则类型，默认为 `daily`
	// daily: 按天轮转
	// size: 按大小轮转
	Rotation string `json:",default=daily,options=[daily,size]"`
	// colorConsole 表示是否在控制台输出彩色日志，默认为 `false`
	ColorConsole bool `json:",default=false"`
}
