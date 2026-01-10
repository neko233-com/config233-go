# Excel 配置格式说明

本项目支持的 Excel 配置格式如下（基于 1-based 行号）：

| 行号 | 内容类型 | 说明 |
| :--- | :--- | :--- |
| 第 1 行 | 注释 (Comment) | 字段的中文描述或说明信息 |
| 第 2 行 | 客户端字段名 (Client) | 客户端使用的字段名（本项目目前主要参考第 4 行） |
| 第 3 行 | 类型 (Type) | 字段数据类型（如 long, string, int, bool 等） |
| 第 4 行 | 服务端字段名 (Server) | **核心：** 映射到 Go Struct 的字段名或 Map 的 Key |
| 第 5+ 行 | 数据 (Data) | 具体的配置数据内容 |

## 示例

| 注释 | 物品id | 物品名称 | 备注 | 物品描述信息 | 背包分组 | 道具类型 | 过期毫秒数 | 道具品质 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Client** | itemId | itemName | | desc | bagType | type | expireTimeMs | quality |
| **type** | long | string | string | string | string | int | long | int |
| **Server** | itemId | itemName | | desc | bagType | type | expireTimeMs | quality |
| **数据** | 1 | demo | | xxx | item | 1 | | 1 |
| **数据** | 2 | demo232345 | | | | | | |

## 注意事项

1. 程序会自动读取第一个 Working Sheet（工作表）。
2. 第 4 行（索引为 3）作为反射映射的字段名来源。
3. 第 5 行（索引为 4）开始读取实际数据。
4. 如果字段名为空，该列将被跳过。

