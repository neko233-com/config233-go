package config233

import (
	"github.com/go-logr/logr"
)

// Logger 日志接口，基于 logr
// logr 是 Go 生态中类似 slf4j 的日志抽象接口
// 用户可以注入不同的日志实现（如 zap, zerolog 等）
type Logger = logr.Logger

// SetLogger 设置全局日志实现
// 用户可以调用此函数设置自定义的日志实现
// 例如: SetLogger(zap.New(...)) 或 SetLogger(zerolog.New(...))
var globalLogger Logger

func SetLogger(logger Logger) {
	globalLogger = logger
}

// getLogger 获取当前日志实现
// 如果未设置，使用默认的 logr.Discard()（不输出日志）
func getLogger() Logger {
	if globalLogger.IsZero() {
		// 返回一个不输出日志的 logger，避免 nil 指针
		return logr.Discard()
	}
	return globalLogger
}
