# Config233-Go

Config233 的 Go 语言实现，用于配置文件的加载、热更新和数据管理。

## 文档

- [API 可见性说明](API_VISIBILITY.md) - 了解哪些内容对第三方用户可见
- [贡献指南](CONTRIBUTING.md) - 项目结构和开发规范
- [更新日志](ChangeLog/) - 版本更新记录

## 功能特性

- 支持多种配置文件格式（JSON, TSV, Excel）
- 热更新监听文件变化
- 配置数据 ORM 到结构体
- 字段注入和方法回调
- 前端数据导出

## 安装

```bash
go get config233-go
```

## 快速开始

### 使用 ConfigManager233（推荐）

ConfigManager233 提供了更简单的全局配置管理接口：

```go
import "github.com/neko233-com/config233-go/pkg/config233"

// 1. 注册配置类型
config233.RegisterType[Student]()

// 2. 创建管理器并加载配置
manager := config233.NewConfigManager233("./myconfig")
manager.LoadAllConfigs()

// 3. 使用配置
// 按 ID 获取配置
config, exists := config233.GetConfigById[Student]("1")
if exists {
    fmt.Printf("学生信息: %+v", config)
}

// 获取所有配置列表
configs := config233.GetConfigList[Student]()
fmt.Printf("共有 %d 个配置", len(configs))

// 获取配置映射（ID -> Config）
configMap := config233.GetConfigMap[Student]()

// 启动热更新监听
manager.StartWatching()
```

## 测试

项目使用 Go 标准测试框架，测试文件位于 `tests/` 目录下，与主代码分离。

运行所有测试：

```bash
go test ./tests -v
```

运行覆盖率测试：

```bash
go test ./tests -cover
```

## 项目结构

```
config233-go/
├── pkg/config233/          # 公开 API - 这是第三方库用户可以导入的包
│   ├── config233.go        # 核心配置管理
│   ├── manager.go          # 配置管理器
│   ├── listener.go         # 监听器接口
│   ├── handler.go          # 处理器接口
│   ├── dto/                # 数据传输对象
│   ├── excel/              # Excel 处理器
│   ├── json/               # JSON 处理器
│   └── tsv/                # TSV 处理器
├── examples/               # 示例代码（不会被第三方导入）
│   ├── example_usage.go
│   ├── manager_example.go
│   ├── logr_example.go
│   └── validation_demo.go
├── tests/                  # 测试代码
├── testdata/               # 测试数据
├── CheckOutput/            # 临时输出目录（被 git 忽略）
└── GeneratedStruct/        # 生成的结构体代码（被 git 忽略）
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
- `GetKvStringList[T IKvConfig](id string, defaultVal []string) []string` - 从 KV 配置获取字符串列表

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