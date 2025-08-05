package internal

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var Placeholder PlaceholderType

type PlaceholderType = struct{}

func getCaller(callDepth int) string {
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		return ""
	}

	return prettyCaller(file, line)
}

func getTimestamp() string {
	return time.Now().UTC().Format(timeFormat)
}

func prettyCaller(file string, line int) string {
	idx := strings.LastIndexByte(file, '/')
	if idx < 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}

	idx = strings.LastIndexByte(file[:idx], '/')
	if idx < 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}

	return fmt.Sprintf("%s:%d", file[idx+1:], line)
}
