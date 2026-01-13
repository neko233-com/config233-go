package main

import (
	"fmt"
	"log"

	"github.com/neko233-com/config233-go/pkg/config233"

	"github.com/go-logr/logr"
)

// stdLogger 标准库日志适配器，实现 logr.LogSink 接口
type stdLogger struct{}

func (l *stdLogger) Init(info logr.RuntimeInfo) {}

func (l *stdLogger) Info(level int, msg string, keysAndValues ...interface{}) {
	if level <= 0 { // 只输出 Info 级别及以上的日志
		log.Printf("[INFO] %s %v", msg, keysAndValues)
	}
}

func (l *stdLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Printf("[ERROR] %s: %v %v", msg, err, keysAndValues)
}

func (l *stdLogger) Enabled(level int) bool {
	return level <= 0 // 启用 Info 级别及以上的日志
}

func (l *stdLogger) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return l // 简化实现
}

func (l *stdLogger) WithName(name string) logr.LogSink {
	return l // 简化实现
}

// Student 示例配置结构体
type Student struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	fmt.Println("Config233-Go logr 日志接口示例")

	// 注册配置类型
	config233.RegisterType[Student]()

	// 示例1: 使用默认日志（静默模式，不输出日志）
	fmt.Println("=== 使用默认日志（静默模式） ===")
	config, exists := config233.GetConfigById[Student](1)
	if exists {
		fmt.Printf("找到配置: %+v\n", config)
	} else {
		fmt.Println("配置不存在")
	}

	// 示例2: 设置标准库日志实现
	fmt.Println("\n=== 设置标准库日志实现 ===")
	logger := logr.New(&stdLogger{})
	config233.SetLogger(logger)

	// 现在日志调用会输出到控制台
	config, exists = config233.GetConfigById[Student](1)
	if exists {
		fmt.Printf("找到配置: %+v\n", config)
	} else {
		fmt.Println("配置不存在")
	}

	// 示例3: 使用其他日志库（如 zap）
	fmt.Println("\n=== 使用 zap 日志库示例（需要额外安装） ===")
	fmt.Println("// go get go.uber.org/zap")
	fmt.Println("// zapLogger := zap.New(...)")
	fmt.Println("// config233.SetLogger(zapr.NewLogger(zapLogger))")
}
