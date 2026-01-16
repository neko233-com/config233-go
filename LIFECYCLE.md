# Config233 配置加载与生命周期

## 配置加载机制

### 字段映射策略（重要！）

config233 使用**直接字段映射**而非 JSON 序列化方式：

```
Excel 读取 → string 值 → 直接转换为字段类型 → 赋值
```

**不同于传统的 JSON 序列化方式：**
```
❌ 旧方式: Excel → map → JSON.Marshal → JSON.Unmarshal → struct
✅ 新方式: Excel → map → 直接类型转换 → struct
```

这样做的好处：
- ✅ Excel 列名直接映射到 `config233_column` 标签
- ✅ 避免 JSON 标签不匹配导致的字段丢失
- ✅ 类型转换错误直接在控制台输出，便于调试

### 类型转换

自动支持以下类型转换（string → 目标类型）：

| 字段类型 | 转换规则 | 空值处理 | 错误处理 |
|---------|---------|---------|---------|
| `string` | 直接赋值 | 保持空字符串 | N/A |
| `int/int64` 等 | `strconv.ParseInt` | 设为 0 | 红色输出 |
| `uint` 等 | `strconv.ParseUint` | 设为 0 | 红色输出 |
| `float32/float64` | `strconv.ParseFloat` | 设为 0.0 | 红色输出 |
| `bool` | `strconv.ParseBool` | 设为 false | 红色输出 |

**bool 类型特殊处理：**
- 支持：`true/false`, `1/0`, `yes/no`, `on/off`, `enabled/disabled`
- 不区分大小写

**转换失败时的输出示例：**
```bash
[ERROR] 字段类型转换失败: 无法将 'abc' 转换为 int, 错误: strconv.ParseInt: parsing "abc": invalid syntax
```

## config233_column 标签

### 标签优先级

字段映射按以下优先级查找：

1. **`config233_column` 标签**（最高优先级）
2. **字段名**（不区分大小写）
3. **首字母小写的字段名**（如 `Id` -> `id`）

### 正确使用示例

假设 Excel Server 行的列名为：`id`, `skillId`, `unlockCostGoldCount`

```go
// ✅ 正确 - 标签与 Excel 列名完全匹配
type WeaponConfig struct {
    Id                  int `config233_column:"id"`
    SkillId             int `config233_column:"skillId"`
    UnlockCostGoldCount int `config233_column:"unlockCostGoldCount"`
}

// ❌ 错误 - 标签与 Excel 列名不匹配，字段值会是 0
type WeaponConfig struct {
    Id           int `json:"id" config233_column:"weaponId"`           // Excel 没有 weaponId 列
    UnlockedCost int `json:"cost" config233_column:"unlockedCost"`     // Excel 是 unlockCostGoldCount
}
```

### 标签规则

- 如果设置了 `config233_column`，**必须与 Excel 列名完全匹配**（不区分大小写）
- JSON 标签只影响 JSON 序列化输出，**不影响 Excel 字段映射**
- 如果没有 `config233_column` 标签，会尝试用字段名匹配

## 生命周期方法

### 1. AfterLoad() - 加载后处理

```go
// IConfigLifecycle 配置生命周期接口
type IConfigLifecycle interface {
    AfterLoad()
}
```

**调用时机：** 所有字段赋值完成后

**使用场景：**
- 构建索引（如 `map[int]*Config`）
- 缓存计算结果
- 建立配置间的引用关系

**示例：**
```go
type WeaponConfig struct {
    Id       int
    SkillId  int
    // 缓存字段
    skillRef *SkillConfig // 引用另一个配置
}

func (c *WeaponConfig) AfterLoad() {
    // 根据 SkillId 查找并缓存 SkillConfig
    c.skillRef = config233.GetConfigById[SkillConfig](c.SkillId)
}
```

### 2. Check() error - 数据校验

```go
// IConfigValidator 配置校验接口
type IConfigValidator interface {
    Check() error
}
```

**调用时机：** AfterLoad() 之后

**返回值：**
- `nil` - 校验通过
- `error` - 校验失败，**输出红色错误但不中断加载**

**使用场景：**
- 验证字段值范围（如 `price >= 0`）
- 检查字段间逻辑关系
- 验证引用的有效性

**示例：**
```go
func (c *WeaponConfig) Check() error {
    if c.Id <= 0 {
        return fmt.Errorf("WeaponConfig.id=%d 必须大于0", c.Id)
    }
    
    if c.UnlockCost < 0 {
        return fmt.Errorf("WeaponConfig.id=%d 价格不能为负数: %d", c.Id, c.UnlockCost)
    }
    
    // 检查引用
    if c.SkillId > 0 && c.skillRef == nil {
        return fmt.Errorf("WeaponConfig.id=%d 找不到 skillId=%d", c.Id, c.SkillId)
    }
    
    return nil
}
```

**校验失败时的输出示例：**
```bash
[ERROR] 配置校验失败 [WeaponConfig] index=0: WeaponConfig.id=1001 价格不能为负数: -100
```

## 完整示例

```go
package gameconfig

import (
    "fmt"
    config233 "github.com/neko233-com/config233-go/pkg/config233"
)

// WeaponConfig 武器配置
type WeaponConfig struct {
    Id                  int    `config233_column:"id"`
    Name                string `config233_column:"name"`
    SkillId             int    `config233_column:"skillId"`
    UnlockCostGoldCount int    `config233_column:"unlockCostGoldCount"`
    
    // 缓存字段（不在 Excel 中）
    skillRef *SkillConfig
}

// AfterLoad 配置加载后调用
func (c *WeaponConfig) AfterLoad() {
    // 缓存技能引用
    if c.SkillId > 0 {
        skill, _ := config233.GetConfigById[SkillConfig](c.SkillId)
        c.skillRef = skill
    }
}

// Check 配置校验
func (c *WeaponConfig) Check() error {
    // 基础字段校验
    if c.Id <= 0 {
        return fmt.Errorf("WeaponConfig.id=%d 必须大于0", c.Id)
    }
    
    if c.UnlockCostGoldCount < 0 {
        return fmt.Errorf("WeaponConfig.id=%d 解锁价格不能为负数: %d", 
            c.Id, c.UnlockCostGoldCount)
    }
    
    // 引用校验
    if c.SkillId > 0 && c.skillRef == nil {
        return fmt.Errorf("WeaponConfig.id=%d 找不到技能配置 skillId=%d", 
            c.Id, c.SkillId)
    }
    
    return nil
}

// 使用示例
func init() {
    // 注册配置类型
    config233.RegisterType[WeaponConfig]()
    config233.RegisterType[SkillConfig]()
}
```

## 执行流程

```
1. Excel 读取
   └─> 读取每一行数据（string 值）

2. 字段映射与赋值
   └─> for each 行:
       └─> for each 列:
           ├─> 根据 config233_column 或字段名找到目标字段
           ├─> 将 string 转换为字段类型
           └─> 赋值（失败则输出红色错误）

3. 生命周期调用
   └─> for each 配置对象:
       ├─> 调用 AfterLoad()     // 如果实现了
       └─> 调用 Check()          // 如果实现了
           └─> 失败则输出红色错误
```

## 错误输出格式

所有错误使用 ANSI 红色输出到控制台：

```bash
\033[31m[ERROR] 错误信息\033[0m
```

### 错误类型

| 错误类型 | 示例 |
|---------|------|
| 类型转换失败 | `[ERROR] 字段类型转换失败: 无法将 'abc' 转换为 int, 错误: ...` |
| 配置校验失败 | `[ERROR] 配置校验失败 [WeaponConfig] index=0: ...` |
| 不支持的类型 | `[ERROR] 不支持的字段类型: slice` |

## 注意事项

1. ⚠️ **Excel 列名必须与 config233_column 标签匹配**
   - 列名：`unlockCostGoldCount`
   - 标签：`config233_column:"unlockCostGoldCount"` ✅
   - 标签：`config233_column:"unlockedCost"` ❌ (字段值会是 0)

2. ⚠️ **JSON 标签不影响 Excel 映射**
   ```go
   // JSON 标签只影响输出，不影响 Excel 读取
   Id int `json:"weaponId" config233_column:"id"`
   ```

3. ⚠️ **生命周期方法必须使用指针接收者**
   ```go
   func (c *WeaponConfig) AfterLoad()    // ✅ 正确
   func (c WeaponConfig) AfterLoad()     // ❌ 错误，不会被调用
   ```

4. ✅ **接口是可选的**
   - 可以只实现 `AfterLoad()`
   - 可以只实现 `Check()`
   - 可以两个都实现
   - 可以都不实现

5. ✅ **校验失败不会中断加载**
   - `Check()` 返回错误时，会输出错误信息
   - 但其他配置会继续加载

## 测试

### 测试文件
- `tests/fishing_weapon_config_test.go` - 完整示例
- `tests/type_conversion_test.go` - 类型转换测试

### 运行测试
```bash
# 测试配置加载
go test ./tests -run TestFishingWeaponConfig_Parse -v

# 测试生命周期
go test ./tests -run TestFishingWeaponConfig_Lifecycle -v
```

## 相关文件

- `pkg/config233/api_config233_lifecycle.go` - 生命周期接口定义
- `pkg/config233/excel/handler.go` - Excel 处理器实现（字段映射 + 生命周期）
- `tests/fishing_weapon_config_test.go` - 测试示例
