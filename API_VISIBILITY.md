# API 可见性说明

## 对第三方库用户可见的内容

当其他项目导入 `github.com/neko233-com/config233-go/pkg/config233` 时，他们可以访问：

### 核心类型和函数

```go
import "github.com/neko233-com/config233-go/pkg/config233"

// 配置管理器（推荐使用）
manager := config233.NewConfigManager233(configDir)
config233.GetInstance()
config233.RegisterType[T]()
config233.GetConfigById[T](id)
config233.GetAllConfigList[T]()
config233.GetConfigMap[T]()

// 完整 API
cfg := config233.NewConfig233()
config233.GetConfigList[T](cfg)
```

### 接口

```go
// 配置处理器接口
type IConfigHandler interface {
    ReadToMap(filePath string) (map[string]interface{}, error)
    ReadToFrontEndDataList(configName string, filePath string) interface{}
}

// 配置监听器接口
type IConfigListener interface {
    OnConfigLoaded(configName string, configs []interface{})
    OnConfigReloaded(configName string, configs []interface{})
}

// 字段监听器接口
type IFieldListener interface {
    OnFieldChanged(fieldName string, oldValue, newValue interface{})
}
```

### 处理器

```go
import "github.com/neko233-com/config233-go/pkg/config233/excel"
import "github.com/neko233-com/config233-go/pkg/config233/json"
import "github.com/neko233-com/config233-go/pkg/config233/tsv"

excelHandler := &excel.ExcelConfigHandler{}
jsonHandler := &json.JsonConfigHandler{}
tsvHandler := &tsv.TsvConfigHandler{}
```

### DTO

```go
import "github.com/neko233-com/config233-go/pkg/config233/dto"

type FrontEndConfigDto struct {
    ConfigName string
    FieldNames []string
    Rows       [][]interface{}
}
```

## 对第三方库用户不可见的内容

### examples/ 目录

第三方用户**无法导入** `examples/` 目录下的任何代码。

```go
// ❌ 这不会工作
import "github.com/neko233-com/config233-go/examples"
```

`examples/` 目录仅用于：
- 展示如何使用库
- 提供可运行的示例代码
- 作为开发者的参考

包含的文件：
- `example_usage.go` - 基本使用示例
- `manager_example.go` - 配置管理器示例
- `logr_example.go` - 日志集成示例
- `validation_demo.go` - 配置验证示例
- `excel_validation.go` - Excel 验证示例

### tests/ 目录

测试代码也**无法被导入**。

```go
// ❌ 这不会工作
import "github.com/neko233-com/config233-go/tests"
```

### testdata/ 目录

测试数据目录不包含任何可导入的代码。

### 临时输出目录

这些目录被 git 忽略：
- `CheckOutput/` - 测试输出
- `GeneratedStruct/` - 生成的结构体

## 如何验证

### 第三方用户的使用方式

```go
// 在你的项目中
package main

import (
    "fmt"
    "github.com/neko233-com/config233-go/pkg/config233"
)

type MyConfig struct {
    ID   string `json:"id" config233:"uid"`
    Name string `json:"name"`
}

func main() {
    // 注册配置类型
    config233.RegisterType[MyConfig]()
    
    // 创建配置管理器
    manager := config233.NewConfigManager233("./config")
    manager.LoadAllConfigs()
    
    // 获取配置
    configs := config233.GetAllConfigList[MyConfig]()
    fmt.Printf("Loaded %d configs\n", len(configs))
}
```

### 在本项目中查看示例

```bash
# 克隆仓库
git clone https://github.com/neko233-com/config233-go.git
cd config233-go

# 运行示例
cd examples
go run example_usage.go
```

## 最佳实践

### 对于库用户

1. 只导入 `pkg/config233` 及其子包
2. 参考 `examples/` 目录学习用法
3. 查看 `README.md` 了解 API 文档

### 对于贡献者

1. 公开 API 代码放在 `pkg/config233/`
2. 示例代码放在 `examples/`
3. 测试代码放在 `tests/`
4. 所有导出的类型和函数必须有文档注释
5. 查看 `CONTRIBUTING.md` 了解详细规范

## Go 模块系统说明

Go 的模块系统允许用户导入任何包含 `.go` 文件的目录，但通过项目结构约定：

- `pkg/` - 表示可供外部使用的包
- `internal/` - Go 会**强制阻止**外部导入
- `cmd/` - 通常包含可执行程序
- `examples/` - 按约定不应被导入
- `tests/` - 按约定不应被导入

我们使用 `pkg/` 和 `examples/` 的结构来清晰地区分公开 API 和示例代码。
