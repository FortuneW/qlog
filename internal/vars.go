package internal

import (
	"errors"
)

const (
	TraceLevel uint32 = iota
	// DebugLevel logs everything
	DebugLevel
	// InfoLevel includes info, warnings,errors
	InfoLevel
	// WarnLevel includes warnings,errors
	WarnLevel
	// ErrorLevel includes errors
	ErrorLevel
	// DisableLevel doesn't log any messages
	DisableLevel = 0xff

	LevelInfo    = "INF"
	LevelError   = "ERR"
	LevelWarn    = "WAR"
	LevelDebug   = "DEB"
	LevelTrace   = "TRA"
	LevelDisable = "OFF"

	levelAccessRecord = "access"
)

const (
	plainEncodingSep = ' '
	sizeRotationRule = "size"

	managerFilename = "manager.log"
	serverFilename  = "server.log"

	fileMode = "file"

	backupFileDelimiter = "-"
	nilAngleString      = "<nil>"
	flags               = 0x0
)

var (
	ErrLogPathNotSet        = errors.New("log path must be set")
	ErrLogServiceNameNotSet = errors.New("log service name must be set")
)
