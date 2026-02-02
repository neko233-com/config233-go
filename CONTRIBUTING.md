# 贡献指南

## 项目结构规范

### 公开 API (`pkg/config233/`)

这是第三方用户可以导入使用的包。所有放在 `pkg/config233/` 目录下的代码都是公开 API 的一部分。

**规则：**
- 导出的类型、函数、常量必须有清晰的文档注释
- 保持 API 稳定性，避免破坏性更改
- 遵循 Go 语言规范和最佳实践
- 所有公开的类型和函数必须以大写字母开头

**目录结构：**
```
pkg/config233/
├── config233.go          # 核心配置管理
├── manager.go            # 配置管理器（简化 API）
├── handler.go            # 配置处理器接口
├── listener.go           # 配置监听器接口
├── field_listener.go     # 字段监听器
├── logger.go             # 日志接口
├── repository.go         # 配置仓库
├── struct_generator.go   # 结构体生成器
├── doc.go                # 包级文档
├── dto/                  # 数据传输对象
│   └── dto.go
├── excel/                # Excel 处理器
│   └── handler.go
├── json/                 # JSON 处理器
│   └── handler.go
└── tsv/                  # TSV 处理器
    └── handler.go
```

### 示例代码 (`examples/`)

示例代码展示如何使用库，但**不会**被第三方用户导入。

**规则：**
- 所有示例都应该是可运行的完整程序（`package main`）
- 提供清晰的注释说明示例的用途
- 不依赖项目内部的测试数据（或明确说明如何设置）
- 每个示例应该专注于一个特定功能

**文件：**
- `example_usage.go` - 基本使用示例
- `manager_example.go` - 配置管理器示例
- `logr_example.go` - 日志集成示例
- `validation_demo.go` - 配置验证示例
- `excel_validation.go` - Excel 验证示例

### 测试代码 (`tests/`)

单元测试和集成测试。

**规则：**
- 使用 Go 标准测试框架
- 测试文件以 `_test.go` 结尾
- 测试函数以 `Test` 开头
- 提供足够的测试覆盖率
- 使用 `testdata/` 目录存放测试数据

### 测试数据 (`testdata/`)

测试使用的配置文件和数据。

**规则：**
- 只包含测试所需的数据
- 不包含敏感信息
- 文件应该尽可能小和简单

### 临时输出目录

以下目录被 `.gitignore` 忽略，不会提交到代码库：

- `CheckOutput/` - 测试输出的 JSON 文件
- `GeneratedStruct/` - 自动生成的结构体代码

## 开发工作流

### 1. 克隆仓库

```bash
git clone https://github.com/neko233-com/config233-go.git
cd config233-go
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 运行测试

```bash
go test ./tests -v
```

### 4. 运行示例

```bash
cd examples
go run example_usage.go
```

### 5. 构建

```bash
go build ./internal/config233
```

## 提交代码

### 提交信息格式

```
<type>: <subject>

<body>

<footer>
```

**Type:**
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整（不影响功能）
- `refactor`: 重构（既不是新功能也不是修复）
- `test`: 测试相关
- `chore`: 构建或辅助工具的变动

**示例:**

```
feat: 添加 Excel 配置文件热更新支持

- 实现 Excel 文件监听
- 添加文件变更检测
- 更新文档

Closes #123
```

## 发布流程

使用自动发布脚本：

### Windows

```cmd
.\release.cmd
```

或

```powershell
.\release.ps1
```

脚本会：
1. 检查工作目录是否干净
2. 运行所有测试
3. 构建项目
4. 创建 Git tag
5. 推送到远程仓库

## 代码规范

- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 使用 `go fmt` 格式化代码
- 使用 `go vet` 检查代码
- 导出的标识符必须有文档注释
- 保持函数简短和单一职责
- 优先使用组合而不是继承

## 问题反馈

如果发现问题或有改进建议，请：

1. 在 GitHub Issues 中创建问题
2. 提供详细的问题描述
3. 如果可能，提供最小可复现示例
4. 说明你的环境（Go 版本、操作系统等）
