package tests

import (
	"testing"

	config233 "github.com/neko233-com/config233-go/internal/config233"
)

// InvalidTypeConfig 用于测试类型转换错误输出
type InvalidTypeConfig struct {
	Id       int     `config233_column:"id"`
	BadInt   int     `config233_column:"badInt"`   // Excel 中可能是非数字
	BadFloat float64 `config233_column:"badFloat"` // Excel 中可能是非数字
}

// AfterLoad 配置加载后调用
func (c *InvalidTypeConfig) AfterLoad() {
	// 不做任何事
}

// Check 配置校验
func (c *InvalidTypeConfig) Check() error {
	// 不做校验，只是为了测试
	return nil
}

// TestTypeConversionError 测试类型转换错误输出（需要手动创建测试数据）
// 这个测试需要创建一个包含无效数据的 Excel 文件来验证红色错误输出
func TestTypeConversionError(t *testing.T) {
	t.Skip("此测试需要手动创建包含无效数据的 Excel 文件")

	testDir := getTestDataDir()
	manager := config233.NewConfigManager233(testDir)
	config233.Instance = manager

	config233.RegisterType[InvalidTypeConfig]()

	// 如果存在 InvalidTypeConfig.xlsx 并且包含无效数据
	// 加载时应该会看到红色的错误输出
	if err := manager.LoadAllConfigs(); err != nil {
		t.Logf("加载配置时出错: %v", err)
	}
}
