# Config233-Go

Config233 的 Go 语言实现，用于配置文件的加载、热更新和数据管理。

## 文档

- [API 可见性说明](API_VISIBILITY.md) - 了解哪些内容对第三方用户可见
- [贡献指南](CONTRIBUTING.md) - 项目结构和开发规范
- [重构说明](REFACTORING_SUMMARY.md) - 代码重构和优化说明
- [更新日志](ChangeLog/) - 版本更新记录

## 功能特性

- ✅ 支持多种配置文件格式（JSON, TSV, Excel）
- ✅ **并行加载** - 多核 CPU 下加速 3-7x
- ✅ **智能热重载** - 批量重载 + 冷却机制，避免频繁刷新
- ✅ **批量回调** - 配置变更一次性通知，精确知道哪些配置变了
- ✅ 配置数据 ORM 到结构体
- ✅ 线程安全 - 无锁读取，支持高并发

## 安装

```bash
go get github.com/neko233-com/config233-go
```

## 快速开始

### 使用 ConfigManager233（推荐）

ConfigManager233 提供了更简单的全局配置管理接口：

```go
import "github.com/neko233-com/config233-go/pkg/config233"

// 1. 注册配置类型
config233.RegisterType[Student]()

// 2. 获取全局单例并设置配置目录
manager := config233.GetInstance()
manager.SetConfigDir("./myconfig")

// 3. 启动（自动并行加载 + 启动热重载监听）
manager.Start()

// 4. 使用配置
// 按 ID 获取配置（支持 string/int/int64）
config, exists := config233.GetConfigById[Student]("1")
config, exists := config233.GetConfigById[Student](1)    // int 也支持

// 获取所有配置列表
configs := config233.GetConfigList[Student]()

// 获取配置映射（ID -> Config）
configMap := config233.GetConfigMap[Student]()
```

### 业务配置管理器（接收配置变更通知）

```go
// 实现 IBusinessConfigManager 接口
type MyConfigManager struct {}

// OnConfigLoadComplete 配置加载/重载完成时调用（批量）
// changedConfigNameList: 本次变更的配置名称列表
func (m *MyConfigManager) OnConfigLoadComplete(changedConfigNameList []string) {
    for _, name := range changedConfigNameList {
        switch name {
        case "ItemConfig":
            m.refreshItemCache()
        case "PlayerConfig":
            m.refreshPlayerCache()
        }
    }
    log.Printf("配置已更新: %v", changedConfigNameList)
}

// 注册业务管理器
manager.RegisterBusinessManager(&MyConfigManager{})
```

### KV 配置使用

```go
// 定义 KV 配置结构体
type GameKvConfig struct {
    Id    string `json:"id"`
    Value string `json:"value"`
}

// 实现 IKvConfig 接口
func (c *GameKvConfig) GetValue() string { return c.Value }

// 注册并使用
config233.RegisterType[GameKvConfig]()

// 获取 KV 配置值
maxLevel := config233.GetKvToInt[GameKvConfig]("max_level", 100)
serverName := config233.GetKvToString[GameKvConfig]("server_name", "默认服务器")
isOpen := config233.GetKvToBoolean[GameKvConfig]("is_open", false)
```

## 性能特性

### 并行加载
首次启动时使用并行加载，充分利用多核 CPU：
```
测试环境: 7 个配置文件（Excel + JSON）
- 串行加载: ~50ms
- 并行加载: ~15ms
- 提升: 约 3.3x
```

### 智能热重载
文件变更时自动批量重载，避免频繁刷新：
- 收集 500ms 内的所有变更
- 批量重载所有变更的配置
- 两次重载之间至少间隔 300ms

### 批量回调
配置变更时只调用一次回调，传递所有变更的配置名：
```go
// 之前（多次回调）
OnConfigLoadComplete("Config1")  // 第1次
OnConfigLoadComplete("Config2")  // 第2次

// 现在（一次回调，知道哪些配置变了）
OnConfigLoadComplete([]string{"Config1", "Config2"})  // 只调用1次
```

## 测试

项目使用 Go 标准测试框架，测试覆盖：

运行所有测试：
```bash
go test ./... -v
```

运行性能基准测试：
```bash
go test ./pkg/config233 -bench=. -benchmem
```

测试覆盖的场景：
- ✅ 并行加载正确性
- ✅ 批量回调机制（19+ 测试用例）
- ✅ 热重载批量和冷却
- ✅ 并发访问安全性
- ✅ 内存效率
- ✅ 边界情况和异常处理

## 项目结构

```
config233-go/
├── pkg/config233/          # 公开 API
│   ├── manager.go          # 核心配置管理器
│   ├── loader_excel.go     # Excel 加载器
│   ├── loader_json.go      # JSON 加载器
│   ├── loader_tsv.go       # TSV 加载器
│   ├── hot_reload.go       # 热重载机制（批量 + 冷却）
│   ├── *_test.go           # 单元测试（30+ 测试用例）
���   ├── dto/                # 数据传输对象
│   ├── excel/              # Excel 处理器
│   ├── json/               # JSON 处理器
│   └── tsv/                # TSV 处理器
├── examples/               # 示例代码
├── tests/                  # 集成测试
├── testdata/               # 测试数据
└── GeneratedStruct/        # 生成的结构体代码
```

## 公开 API

当用户导入 `github.com/neko233-com/config233-go/pkg/config233` 时，他们可以访问：

### 核心类型
- `Config233` - 核心配置管理类
- `ConfigManager233` - 简化的配置管理器
- `IKvConfig` - KV 配置接口（用于键值对配置）
- `IConfigHandler` - 配置处理器接口
- `IConfigListener` - 配置监听器接口
- `dto` 包中的数据传输对象
- 各种处理器（excel, json, tsv）

### 泛型查询函数（推荐使用）
- `GetConfigById[T any](id interface{}) (*T, bool)` - 根据 ID 获取单个配置
- `GetConfigList[T any]() []*T` - 获取所有配置列表
- `GetConfigMap[T any]() map[string]*T` - 获取配置映射（ID -> Config）
- `GetKvToString[T IKvConfig](id string, defaultVal string) string` - 从 KV 配置获取字符串值
- `GetKvToInt[T IKvConfig](id string, defaultVal int) int` - 从 KV 配置获取整数值
- `GetKvToBoolean[T IKvConfig](id string, defaultVal bool) bool` - 从 KV 配置获取布尔值
- `GetKvToCsvStringList[T IKvConfig](id string, defaultVal []string) []string` - 从 KV 配置获取 CSV 字符串列表（按逗号分隔）

### 类型注册
- `RegisterType[T any]()` - 注册配置结构体类型
- `RegisterTypeByReflect(typ reflect.Type)` - 通过反射类型注册

### 配置管理器
- `GetInstance() *ConfigManager233` - 获取全局单例实例
- `NewConfigManager233(configDir string) *ConfigManager233` - 创建配置管理器（已废弃，建议使用 GetInstance）

## 示例代码

查看 `examples/` 目录获取完整的使用示例：
- `examples/example_usage.go` - 基本使用
- `examples/manager_example.go` - 配置管理器使用
- `examples/logr_example.go` - 日志集成
- `examples/validation_demo.go` - 配置验证

### 日志配置

Config233-Go 支持 logr 接口，可以集成各种日志库：

```go
import "config233-go/pkg/config233"

// 设置自定义日志器
config233.SetLogger(yourLogrLogger)
```

### 使用 Config233（完整功能）

```go
type Student struct {
    ID   int    `json:"id" config233:"uid"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}
```

### 2. 初始化配置

```go
cfg := config233.NewConfig233().
    Directory("./config").
    AddConfigHandler("json", &json.JsonConfigHandler{}).
    RegisterConfigClass("Student", reflect.TypeOf(Student{})).
    Start()
```

### 3. 获取配置数据

```go
// 使用 ConfigManager233（推荐）
config233.RegisterType[Student]()
manager := config233.NewConfigManager233("./config")
manager.LoadAllConfigs()

// 获取所有配置列表
students := config233.GetConfigList[Student]()

// 按 ID 获取
student, exists := config233.GetConfigById[Student]("1")

// 获取配置映射
studentMap := config233.GetConfigMap[Student]()

// 使用 Config233（完整功能）
students := config233.GetConfigList[Student](cfg)
```

### 4. 热更新注册

```go
type StudentUpdater struct {
    StudentMap map[int]*Student `config233:"inject"`
}

updater := &StudentUpdater{}
cfg.RegisterForHotUpdate(updater)
```

## 配置处理器

### JSON 处理器

```go
handler := &json.JsonConfigHandler{}
cfg.AddConfigHandler("json", handler)
```

### TSV 处理器

```go
handler := &tsv.TsvConfigHandler{}
cfg.AddConfigHandler("tsv", handler)
```

### Excel 处理器

```go
handler := &excel.ExcelConfigHandler{}
cfg.AddConfigHandler("xlsx", handler)
```

## 配置文件格式

配置文件应放在指定目录中，文件名对应配置类名。

例如：
- `config/Student.json` 对应 `Student` 配置类

## 注解说明

- `config233:"uid"` - 标记唯一标识字段
- `config233:"inject"` - 标记需要注入配置映射的字段
- `config233:"hotupdate"` - 标记热更新时调用的方法

## 发布

项目包含自动发布脚本，支持一键发布到 Go 模块代理。

### 使用批处理脚本 (Windows)

```cmd
.\release.cmd
```

CMD脚本会调用PowerShell脚本执行发布流程。

### 使用 PowerShell 脚本

```powershell
.\release.ps1
```

脚本会提示您输入版本标签。

发布脚本会：
1. 检查工作目录是否干净
2. 运行所有测试
3. 构建项目
4. 创建 Git tag
5. 推送 tag 和主分支

## 许可证

与原 Kotlin 版本相同。