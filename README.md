# Config233-Go

Config233 的 Go 语言实现，用于配置文件的加载、热更新和数据管理。

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
import "config233-go/pkg/config233"

// 使用全局实例
config, exists := config233.Instance.GetConfig("StudentConfig", "1")
if exists {
    fmt.Printf("学生信息: %+v", config)
}

// 或者创建自定义实例
manager := config233.NewConfigManager233("./myconfig")
manager.StartWatching() // 启动热更新监听
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

## 示例

查看 `logr_example.go` 和 `manager_example.go` 获取使用示例。

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
students := cfg.GetConfigList(reflect.TypeOf(Student{}))
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
release.cmd v1.0.0
```

### 使用 PowerShell 脚本

```powershell
.\release.ps1 -Version v1.0.0
```

发布脚本会：
1. 检查工作目录是否干净
2. 运行所有测试
3. 构建项目
4. 创建 Git tag
5. 推送 tag 和主分支

## 许可证

与原 Kotlin 版本相同。