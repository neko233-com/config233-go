package config233

import "log"

// Logger 日志接口，类似于 slf4j
// 提供统一的日志输出接口，允许用户注入不同的日志实现
type Logger interface {
	// Debug 输出调试级别日志
	Debug(args ...interface{})
	// Debugf 输出格式化的调试级别日志
	Debugf(format string, args ...interface{})

	// Info 输出信息级别日志
	Info(args ...interface{})
	// Infof 输出格式化的信息级别日志
	Infof(format string, args ...interface{})

	// Warn 输出警告级别日志
	Warn(args ...interface{})
	// Warnf 输出格式化的警告级别日志
	Warnf(format string, args ...interface{})

	// Error 输出错误级别日志
	Error(args ...interface{})
	// Errorf 输出格式化的错误级别日志
	Errorf(format string, args ...interface{})
}

// defaultLogger 默认日志实现，使用标准库 log
type defaultLogger struct{}

func (l *defaultLogger) Debug(args ...interface{}) {
	// 默认实现不输出调试日志
}

func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	// 默认实现不输出调试日志
}

func (l *defaultLogger) Info(args ...interface{}) {
	log.Print(args...)
}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *defaultLogger) Warn(args ...interface{}) {
	log.Print(args...)
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *defaultLogger) Error(args ...interface{}) {
	log.Print(args...)
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// SetLogger 设置全局日志实现
// 用户可以调用此函数设置自定义的日志实现
var globalLogger Logger = &defaultLogger{}

func SetLogger(logger Logger) {
	globalLogger = logger
}

// getLogger 获取当前日志实现
func getLogger() Logger {
	return globalLogger
}
