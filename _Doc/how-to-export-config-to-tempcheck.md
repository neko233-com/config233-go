# 如何将内存中的配置导出到文件

`config233-go` 提供了将加载到内存中的配置导出为 JSON 文件的功能。这个功能在以下场景非常有用：

- **调试配置**：查看 Excel/JSON 配置解析后在内存中的实际数据结构
- **验证数据**：确认配置加载是否正确，字段映射是否符合预期
- **配置检查**：便于策划或 QA 人员检查配置数据

## 基本用法

```go
import (
    "github.com/neko233-com/config233-go/internal/config233"
)

func main() {
    // 1. 创建配置管理器
    cm := config233.NewConfigManager("./testdata")
    
    // 2. 设置导出目录（支持相对路径和绝对路径）
    cm.SetLoadDoneWriteConfigFileDir("./TempCheck/CheckConfig")
    
    // 3. 开启导出功能
    cm.SetIsOpenWriteTempFileToSeeMemoryConfig(true)
    
    // 4. 注册配置类型
    config233.RegisterType[YourConfigStruct]()
    
    // 5. 加载配置（加载完成后会自动导出到指定目录）
    err := cm.LoadAllConfigs()
    if err != nil {
        panic(err)
    }
}
```

## 详细说明

### SetLoadDoneWriteConfigFileDir

设置配置导出的目标目录。

```go
// 使用相对路径
cm.SetLoadDoneWriteConfigFileDir("./TempCheck/CheckConfig")

// 使用绝对路径
cm.SetLoadDoneWriteConfigFileDir("D:\\Code\\Go-Projects\\config233-go\\TempCheck\\CheckConfig")
```

### SetIsOpenWriteTempFileToSeeMemoryConfig

控制是否开启导出功能：

```go
// 开启导出
cm.SetIsOpenWriteTempFileToSeeMemoryConfig(true)

// 关闭导出（默认）
cm.SetIsOpenWriteTempFileToSeeMemoryConfig(false)
```

> **注意**：必须同时设置导出目录和开启导出功能，配置才会被导出。

## 完整示例

```go
package main

import (
    "fmt"
    "path/filepath"
    
    "github.com/neko233-com/config233-go/internal/config233"
)

// ItemConfig 物品配置
type ItemConfig struct {
    Id    int    `json:"id" config233_column:"id"`
    Name  string `json:"name" config233_column:"name"`
    Price int    `json:"price" config233_column:"price"`
}

func main() {
    // 配置目录
    configDir := "./testdata"
    
    // 导出目录
    exportDir := "./TempCheck/CheckConfig"
    
    // 创建配置管理器
    cm := config233.NewConfigManager(configDir)
    
    // 设置导出目录并开启导出功能
    cm.SetLoadDoneWriteConfigFileDir(exportDir)
    cm.SetIsOpenWriteTempFileToSeeMemoryConfig(true)
    
    // 注册配置类型
    config233.RegisterType[ItemConfig]()
    
    // 加载所有配置
    err := cm.LoadAllConfigs()
    if err != nil {
        fmt.Printf("加载配置失败: %v\n", err)
        return
    }
    
    fmt.Printf("配置已导出到: %s\n", exportDir)
    
    // 导出的文件名为: ItemConfig.json
    exportedFile := filepath.Join(exportDir, "ItemConfig.json")
    fmt.Printf("物品配置文件: %s\n", exportedFile)
}
```

## 导出文件格式

导出的 JSON 文件采用格式化的缩进格式，便于阅读：

```json
[
  {
    "id": 1001,
    "name": "铁剑",
    "price": 100
  },
  {
    "id": 1002,
    "name": "钢剑",
    "price": 500
  }
]
```

## 配置热重载时的导出

当配置发生热重载时，如果开启了导出功能，新的配置也会自动导出到指定目录，覆盖原有文件。

```go
// 启动文件监听（热重载）
cm.StartFileWatcher()

// 当配置文件发生变化时：
// 1. 自动重新加载配置
// 2. 自动导出更新后的配置到 TempCheck 目录
```

## 注意事项

1. **性能影响**：生产环境建议关闭此功能，避免不必要的 IO 操作
2. **磁盘空间**：导出的 JSON 文件会占用磁盘空间，请定期清理
3. **目录权限**：确保程序有写入导出目录的权限
4. **文件命名**：导出文件名与配置类型名相同，如 `ItemConfig` 导出为 `ItemConfig.json`

## 相关 API

| 方法 | 说明 |
|------|------|
| `SetLoadDoneWriteConfigFileDir(dir string)` | 设置导出目录 |
| `GetLoadDoneWriteConfigFileDir() string` | 获取导出目录 |
| `SetIsOpenWriteTempFileToSeeMemoryConfig(isOpen bool)` | 开启/关闭导出功能 |
| `ExportConfigToJSON(configName string, data interface{})` | 手动导出指定配置 |

