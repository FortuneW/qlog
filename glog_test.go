package xlog

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogLevelOperations(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		isValidLevel  bool
		expectedLevel string
	}{
		{
			name:          "有效的DEBUG级别",
			level:         "DEB",
			isValidLevel:  true,
			expectedLevel: "DEB",
		},
		{
			name:          "有效的INFO级别(小写)",
			level:         "inf",
			isValidLevel:  true,
			expectedLevel: "INF",
		},
		{
			name:          "有效的WARNING级别",
			level:         "WAR",
			isValidLevel:  true,
			expectedLevel: "WAR",
		},
		{
			name:          "无效的日志级别",
			level:         "INVALID",
			isValidLevel:  false,
			expectedLevel: "ERR", // 因为默认是ERR级别
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试检查日志级别
			err := CheckLogLevelStr(tt.level)
			if tt.isValidLevel && err != nil {
				t.Errorf("CheckLogLevelStr(%s) 期望无错误，但得��错误: %v", tt.level, err)
			}
			if !tt.isValidLevel && err == nil {
				t.Errorf("CheckLogLevelStr(%s) 期望有错误，但没有得到错误", tt.level)
			}

			// 如果是有效的日志级别，测试设置和获取
			if tt.isValidLevel {
				SetLogLevelStr(tt.level)
				got := GetLogLevelStr()
				if got != tt.expectedLevel {
					t.Errorf("设置日志级别后获取值不匹配，期望 %s，得到 %s", tt.expectedLevel, got)
				}
			}
		})
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name          string
		serverLogDir  string
		managerLogDir string
		maxLogBackups int
		maxFileSize   int
		expectedError bool
	}{
		{
			name:          "基本初始化",
			serverLogDir:  "./logs/server",
			managerLogDir: "./logs/manager",
			maxLogBackups: 3,
			maxFileSize:   100,
			expectedError: false,
		},
		{
			name:          "无效的备份数量",
			serverLogDir:  "./logs/server",
			managerLogDir: "./logs/manager",
			maxLogBackups: -1,
			maxFileSize:   100,
			expectedError: true,
		},
		{
			name:          "无效的文件大小",
			serverLogDir:  "./logs/server",
			managerLogDir: "./logs/manager",
			maxLogBackups: 3,
			maxFileSize:   0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理之前的状态
			UnInit()

			// 执行初始化
			err := InitWithConfig(Config{
				ServerLogDir:  tt.serverLogDir,
				ManagerLogDir: tt.managerLogDir,
				MaxBackups:    tt.maxLogBackups,
				MaxSize:       tt.maxFileSize,
			})
			if (err != nil) != tt.expectedError {
				t.Errorf("Init() 错误 = %v, 期望错误 = %v", err, tt.expectedError)
			}

			// 如果初始化成功，验证日志级别和其他设置
			if err == nil {
				// 验证默认日志级别是 ERR
				if level := GetLogLevelStr(); level != "ERR" {
					t.Errorf("初始化后的默认日志级别错误，期望 ERR，得到 %s", level)
				}

				// 尝试写入日志
				logger := GetRLog("test")
				logger.Error("输出错误级别测试日志消息")

				// 验证日志目录是否创建
				if _, err := os.Stat(tt.serverLogDir); os.IsNotExist(err) {
					t.Errorf("服务器日志目录未创建: %s", tt.serverLogDir)
				}
				if _, err := os.Stat(tt.managerLogDir); os.IsNotExist(err) {
					t.Errorf("管理日志目录未创建: %s", tt.managerLogDir)
				}
			}

			// 测试完成后清理
			UnInit()

			// 清理测试创建的目录
			os.RemoveAll("./logs")
		})
	}
}

func TestLogRotation(t *testing.T) {

	nullFile, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	defer nullFile.Close()
	os.Stdout = nullFile
	os.Stderr = nullFile

	// 设置测试配置
	serverLogDir := "/tmp/logs/server"
	managerLogDir := "/tmp/logs/manager"
	maxLogBackups := 3 // 最多保留3个备份
	maxFileSize := 1   // 设置为1MB触发滚动

	os.MkdirAll(serverLogDir, 0755)
	os.MkdirAll(managerLogDir, 0755)

	// 清理之前的状态并初始化
	UnInit()
	err := InitWithConfig(Config{
		ServerLogDir:  serverLogDir,
		ManagerLogDir: managerLogDir,
		MaxBackups:    maxLogBackups,
		MaxSize:       maxFileSize,
	})
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	defer func() {
		UnInit()
		os.RemoveAll("/tmp/logs")
	}()

	// 获取日志实例
	alog := GetALog()
	rlog := GetRLog("rotation-test")

	// 使用可读的重复字符串替代二进制数据
	longMessage := strings.Repeat("这是一条测试日志消息", 50) // 每条消息大约300-400字节
	for i := 0; i < 1000; i++ {                     // 减少循环次数，但仍能触发日志滚动
		alog.Printf("ALog测试消息 #%d: %s", i, longMessage)
		rlog.Errorf("RLog测试消息 #%d: %s", i, longMessage)
	}

	// 验证日志文件是否按预期滚动
	checkRotatedLogs(t, serverLogDir, maxLogBackups)
	checkRotatedLogs(t, managerLogDir, maxLogBackups)
}

// 辅助函数：检查滚动日志文件
func checkRotatedLogs(t *testing.T, logDir string, expectedBackups int) {
	time.Sleep(time.Second)
	files, err := os.ReadDir(logDir)
	if err != nil {
		t.Errorf("无法读取日志目录 %s: %v", logDir, err)
		return
	}

	// 计算.log和.gz文件数量
	logFiles := 0
	gzFiles := 0
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".log") {
			logFiles++
		} else if strings.HasSuffix(name, ".gz") {
			gzFiles++
		}
	}

	// 验证文件数量
	totalFiles := logFiles + gzFiles
	if totalFiles > expectedBackups+1 { // +1是因为当前活动日志文件
		t.Errorf("日志文件数量超出预期，目录 %s 中有 %d 个文件，期望最多 %d 个",
			logDir, totalFiles, expectedBackups+1)
	}

	t.Logf("目录 %s 中有 %d 个当前日志文件和 %d 个压缩的备份文件",
		logDir, logFiles, gzFiles)
}

func TestSetAndGetOpenTime(t *testing.T) {
	// 初始化日志系统
	err := InitWithConfig(Config{
		ServerLogDir:  "./logs/server",
		ManagerLogDir: "./logs/manager",
		MaxBackups:    3,
		MaxSize:       100,
	})
	if err != nil {
		t.Fatalf("初始化日志系统失败: %v", err)
	}
	defer func() {
		UnInit()
		os.RemoveAll("./logs")
	}()

	tests := []struct {
		name          string
		openTime      time.Duration
		tempLogLevel  string
		expectedError bool
		wait          bool
	}{
		{
			name:          "设置3秒超时",
			openTime:      3 * time.Second,
			tempLogLevel:  "DEB",
			expectedError: false,
			wait:          true,
		},
		{
			name:          "设置0秒（取消定时）",
			openTime:      0,
			tempLogLevel:  "DEB",
			expectedError: false,
			wait:          false,
		},
		{
			name:          "设置1秒超时",
			openTime:      1 * time.Second,
			tempLogLevel:  "INF",
			expectedError: false,
			wait:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置临时日志级别
			SetLogLevelStr(tt.tempLogLevel)
			GetRLog("test").Errorf("测试日志级别修改为：%s", tt.tempLogLevel)

			// 测试设置开启时间
			err := SetOpenTime(tt.openTime, nil)
			if (err != nil) != tt.expectedError {
				t.Errorf("SetOpenTime() 错误 = %v, 期望错误 = %v", err, tt.expectedError)
			}

			// 验证GetOpenTime返回值
			if got := GetOpenTime(); got != tt.openTime {
				t.Errorf("GetOpenTime() = %v, 期望 %v", got, tt.openTime)
			}

			// 如果需要等待定时器触发
			if tt.wait && tt.openTime > 0 {
				// 等待定时器触发（多等待100ms确保定时器执行完成）
				time.Sleep(tt.openTime + 100*time.Millisecond)

				// 验证日志级别是否恢复到默认值
				if level := GetLogLevelStr(); level != defaultLogLevel {
					t.Errorf("日志级别未恢复到默认值，当前级别 = %v, 期望 = %v", level, defaultLogLevel)
				}

				// 验证openTime是否重置为0
				if got := GetOpenTime(); got != -1 {
					t.Errorf("定时器触发后 GetOpenTime() = %v, 期望 = 0", got)
				}
			}
		})
	}
}
