// Package config233 提供了配置文件的加载、热更新和数据管理功能。
//
// # 功能特性
//
//   - 支持多种配置文件格式（JSON, TSV, Excel）
//   - 热更新监听文件变化
//   - 配置数据 ORM 到结构体
//   - 字段注入和方法回调
//   - 类型安全的泛型API
//
// # 快速开始
//
// 使用 ConfigManager233（推荐的简化API）:
//
//	import "github.com/neko233-com/config233-go/internal/config233"
//
//	// 初始化全局配置管理器
//	manager := config233.GetInstance()
//	manager.LoadConfig(reflect.TypeOf(YourStruct{}), "YourConfig")
//	manager.StartWatching()
//
//	// 获取配置
//	config, exists := config233.GetConfigById[YourStruct]("configId")
//	if exists {
//	    // 使用配置
//	}
//
// 使用 Config233（完整功能API）:
//
//	import "github.com/neko233-com/config233-go/internal/config233"
//
//	// 创建配置实例
//	cfg := config233.NewConfig233().
//	    Directory("./config").
//	    RegisterConfigClass("Student", reflect.TypeOf(Student{})).
//	    Start()
//
//	// 获取配置列表
//	students := config233.GetConfigList[Student](cfg)
//
// # 配置文件格式
//
// 配置文件应放在指定目录中，文件名对应配置类名。
// 例如：config/Student.json 对应 Student 配置类
//
// # 结构体标签
//
//   - `config233:"uid"` - 标记唯一标识字段（必须）
//   - `config233:"inject"` - 标记需要注入配置映射的字段
//   - `config233:"hotupdate"` - 标记热更新时调用的方法
//
// # 热更新
//
// 注册热更新监听器：
//
//	type Listener struct {
//	    ConfigMap map[string]*YourStruct `config233:"inject"`
//	}
//
//	listener := &Listener{}
//	cfg.RegisterForHotUpdate(listener)
//
// 当配置文件变化时，ConfigMap 字段会自动更新。
//
// # 日志集成
//
// Config233 支持 logr 接口：
//
//	import "github.com/go-logr/logr"
//	config233.SetLogger(yourLogger)
package config233
