package config233

import (
	"fmt"
)

// Logger 日志接口
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(err error, msg string, keysAndValues ...interface{})
}

// ConsoleLogger 默认的控制台日志实现
type ConsoleLogger struct{}

func (l *ConsoleLogger) Info(msg string, keysAndValues ...interface{}) {
	kvStr := ""
	if len(keysAndValues) > 0 {
		kvStr = fmt.Sprintf(" %v", keysAndValues)
	}
	fmt.Printf("[INFO] %s%s\n", msg, kvStr)
}

func (l *ConsoleLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	kvStr := ""
	if len(keysAndValues) > 0 {
		kvStr = fmt.Sprintf(" %v", keysAndValues)
	}
	if err != nil {
		fmt.Printf("\033[31m[ERROR] %s: %v%s\033[0m\n", msg, err, kvStr)
	} else {
		fmt.Printf("\033[31m[ERROR] %s%s\033[0m\n", msg, kvStr)
	}
}

// SetLogger 设置全局日志实现
var globalLogger Logger = &ConsoleLogger{}

func SetLogger(logger Logger) {
	globalLogger = logger
}

// getLogger 获取当前日志实现
func getLogger() Logger {
	if globalLogger == nil {
		return &ConsoleLogger{}
	}
	return globalLogger
}
