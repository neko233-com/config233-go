# 如何注册热重载 (Hot Reload)

`config233-go` 提供了强大的配置热重载机制。当配置文件（Excel, JSON 等）发生变更时，系统会自动重新加载配置，并通知注册的监听者。

以下是三种注册热重载监听的方式：

## 1. 注册简单的回调函数

如果你只需要在配置重载后执行一些简单的逻辑（如清空缓存、打印日志），可以使用 `RegisterReloadFunc`。

```go
import (
    "fmt"
    "github.com/neko233-com/config233-go/internal/config233"
)

func init() {
    // 获取配置管理器实例
    cm := config233.GetInstance()

    // 注册重载回调
    cm.RegisterReloadFunc(func() {
        fmt.Println("配置已重载！可以在这里清理缓存或重新计算数据。")
    })
}
```

## 2. 注册业务配置管理器 (推荐)

对于复杂的业务模块，建议实现 `IBusinessConfigManager` 接口，并注册到配置管理器中。这样可以获得更精细的生命周期控制。

```go
package manager

import (
    "github.com/neko233-com/config233-go/pkg/config233"
)

// Define your manager
type MyGameManager struct {
}

// 1. 实现 OnFirstAllConfigDone (首次加载完成)
func (m *MyGameManager) OnFirstAllConfigDone() {
    // 首次启动时，所有配置加载完毕后调用
    // 初始化游戏数据...
}

// 2. 实现 OnConfigLoadComplete (配置加载/重载完成)
// 参数 names 是本次加载/重载涉及的配置表名称列表
func (m *MyGameManager) OnConfigLoadComplete(names []string) {
    // 这里不仅在首次加载时调用，热重载时也会调用
    for _, name := range names {
        if name == "ItemConfig" {
            // 重建道具索引
        }
    }
}

// 3. 实现 OnConfigHotUpdate (热重载通知) - 可选
// 如果实现了这个接口，会在热重载流程结束后额外通知
// 注意：config233 目前的主要通知机制是 OnConfigLoadComplete
// 如果需要专门监听热重载事件，可以配合 RegisterReloadFunc 使用
```

注册方式：

```go
func init() {
    myManager := &MyGameManager{}
    config233.GetInstance().RegisterBusinessManager(myManager)
}
```

## 3. 在配置结构体内处理 (AfterLoad)

如果你希望在某个特定的配置表加载完数据后立即处理（例如构建 ID 映射、解析复杂字段），可以直接在配置结构体中实现 `AfterLoad` 方法。

这属于 `IConfigLifecycle` 接口的一部分。

```go
type ItemConfig struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

// 配置加载后自动调用
func (c *ItemConfig) AfterLoad() {
    // 数据预处理
    // 例如：将 Name 字段转为大写
    // c.Name = strings.ToUpper(c.Name)
    fmt.Printf("ItemConfig %d loaded\n", c.Id)
}
```

## 总结

| 机制 | 适用场景 | 触发时机 |
| --- | --- | --- |
| `AfterLoad` | 单行数据处理、字段转换 | 解析每行数据并映射到 Struct 后立即调用 |
| `RegisterReloadFunc` | 全局通知、清理缓存 | 每次 `Reload()` 或热更新完成后调用 |
| `RegisterBusinessManager` | 业务模块初始化、依赖特定表的逻辑 | 首次加载完成、热更新完成时调用 |

