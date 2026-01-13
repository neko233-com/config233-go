# 快速参考

## 安装

```bash
go get github.com/neko233-com/config233-go
```

## 基本使用（3 步）

```go
package main

import (
    "fmt"
    "github.com/neko233-com/config233-go/pkg/config233"
)

// 1. 定义配置结构体
type MyConfig struct {
    ID   string `json:"id" config233:"uid"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    // 2. 注册配置类型
    config233.RegisterType[MyConfig]()
    
    // 3. 加载配置
    manager := config233.NewConfigManager233("./config")
    manager.LoadAllConfigs()
    
    // 使用配置
    configs := config233.GetConfigList[MyConfig]()
    fmt.Printf("加载了 %d 个配置\n", len(configs))
    
    // 按 ID 获取
    cfg, exists := config233.GetConfigById[MyConfig]("1")
    if exists {
        fmt.Printf("配置: %+v\n", cfg)
    }
}
```

## 支持的文件格式

### JSON
```json
[
  {
    "id": "1",
    "name": "配置1",
    "age": 10
  }
]
```

### Excel (.xlsx)
第 1 行：注释
第 2 行：字段名称 
第 3 行：字段描述（client）    
第 4 行：字段类型
第 5 行：字段描述（server）
第 6 行及以后：数据

----
Excel 第一列, 也是空列, 完全无意义

### TSV
制表符分隔的值文件。

## 热更新

```go
// 启动文件监听
manager.StartWatching()

// 注册热更新监听器
type MyListener struct {
    ConfigMap map[string]*MyConfig `config233:"inject"`
}

listener := &MyListener{}
manager.GetConfig233().RegisterForHotUpdate(listener)
```

## 常用 API

### 配置管理器（推荐）

```go
// 创建管理器
manager := config233.NewConfigManager233(configDir)

// 获取全局实例
manager := config233.GetInstance()

// 设置配置目录（链式调用）
manager, err := config233.GetInstance().SetConfigDir(configDir).Start()

// 加载所有配置
manager.LoadAllConfigs()

// 启动热更新监听
manager.StartWatching()
```

### 配置查询（泛型 API）

```go
// 注册类型（必须在使用泛型查询前注册）
config233.RegisterType[MyConfig]()

// 获取所有配置列表（返回 []*MyConfig）
configs := config233.GetConfigList[MyConfig]()

// 按 ID 获取配置（支持 string, int, int64 等类型）
cfg, exists := config233.GetConfigById[MyConfig]("id")
cfg, exists := config233.GetConfigById[MyConfig](1)

// 获取配置 Map（ID -> Config，返回 map[string]*MyConfig）
configMap := config233.GetConfigMap[MyConfig]()
for id, config := range configMap {
    fmt.Printf("ID: %s, Config: %+v\n", id, config)
}
```

### KV 配置查询

```go
// 定义 KV 配置结构体（需要实现 IKvConfig 接口）
type MyKvConfig struct {
    Id    string `json:"id"`
    Value string `json:"value"`
}

// 实现 IKvConfig 接口
func (c *MyKvConfig) GetValue() string {
    return c.Value
}

// 注册类型
config233.RegisterType[MyKvConfig]()

// 从 KV 配置获取字符串列表（按逗号分隔）
list := config233.GetKvStringList[MyKvConfig]("list1", []string{"default"})
// 如果配置值为 "a,b,c"，返回 ["a", "b", "c"]
// 如果配置不存在或为空，返回默认值 ["default"]
```

### 完整 API（Config233 - 高级用法）

```go
// 创建 Config233 实例（完整功能，支持热更新回调等）
cfg := config233.NewConfig233().
    Directory("./config").
    RegisterConfigClass("Student", reflect.TypeOf(Student{})).
    Start()

// 获取配置列表
students := config233.GetConfigList[Student](cfg)
```

### 代码生成（自动生成结构体）

```go
// 从 Excel 文件生成 Go 结构体代码
err := config233.GenerateStructFromExcel("./config/ItemConfig.xlsx", "./GeneratedStruct")

// 从目录下所有 Excel 文件生成结构体
err := config233.GenerateStructsFromExcelDir("./config", "./GeneratedStruct")

// 生成的代码会自动：
// - 包含正确的 JSON 标签
// - 对于 KV 配置（文件名包含 "kv"），自动实现 IKvConfig 接口
// - 自动添加必要的导入语句
```

## 结构体标签

```go
type Config struct {
    // 唯一标识字段（必须有一个）
    ID string `json:"id" config233:"uid"`
    
    // 普通字段
    Name string `json:"name"`
    
    // 注入配置 Map
    Configs map[string]*OtherConfig `config233:"inject"`
}
```

## 日志集成

```go
import "github.com/go-logr/logr"

// 设置自定义日志器
config233.SetLogger(yourLogrLogger)
```

## 文件命名规则

配置文件名应与配置结构体名称对应（不区分大小写）：

```
config/
  ├── MyConfig.json      -> MyConfig struct
  ├── UserInfo.xlsx      -> UserInfo struct
  └── ItemData.json      -> ItemData struct
```

## 示例代码

查看 `examples/` 目录获取更多示例：

```bash
cd examples
go run example_usage.go
```

## 常见问题

### Q: 如何指定唯一标识字段？
A: 配置会自动从 `id`、`ID`、`Id` 或 `itemId` 字段提取唯一标识。无需特殊标签。

### Q: 支持热更新吗？
A: 支持。使用 `manager.StartWatching()` 启动文件监听，配置文件变化时自动重载。

### Q: 如何处理嵌套配置？
A: 使用 JSON 字符串或自定义解析。也可以使用多个配置文件，通过 ID 关联。

### Q: 配置文件必须放在同一目录吗？
A: 是的，当前版本要求所有配置文件在同一目录下。

### Q: 如何生成配置结构体代码？
A: 使用 `GenerateStructFromExcel` 或 `GenerateStructsFromExcelDir` 函数，可以从 Excel 文件自动生成 Go 结构体代码。

### Q: KV 配置如何使用？
A: 定义实现 `IKvConfig` 接口的结构体，使用 `GetKvStringList` 获取按逗号分隔的字符串列表。代码生成器会自动为 KV 配置生成接口实现。

### Q: 为什么需要注册类型？
A: 注册类型后，配置数据会自动转换为对应的结构体类型，而不是 `map[string]interface{}`，提供类型安全和更好的 IDE 支持。

## 更多信息

- [完整文档](README.md)
- [API 可见性](API_VISIBILITY.md)
- [贡献指南](CONTRIBUTING.md)
- [示例代码](examples/)
