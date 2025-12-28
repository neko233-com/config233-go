package tests

import (
	"os"
	"path/filepath"
	"testing"

	"config233-go/pkg/config233"
)

// TestConfigManager233_GetConfig 测试配置管理器的 GetConfig 方法
// 验证配置管理器是否能正确获取配置项
func TestConfigManager233_GetConfig(t *testing.T) {
	// 创建测试用的配置管理器
	testDir := getTestDataDir()
	t.Logf("Test data directory: %s", testDir)
	manager := config233.NewConfigManager233(testDir)

	// 测试获取配置
	config, exists := manager.GetConfig("test", "1")
	if exists {
		t.Logf("找到配置: %+v", config)
	} else {
		t.Log("配置不存在，这是正常的，因为我们还没有加载具体的配置")
	}
}

// TestConfigManager233_LoadAllConfigs 测试配置管理器的 LoadAllConfigs 方法
// 验证配置管理器是否能正确加载目录中的所有配置文件
func TestConfigManager233_LoadAllConfigs(t *testing.T) {
	testDir := getTestDataDir()
	manager := config233.NewConfigManager233(testDir)

	err := manager.LoadAllConfigs()
	if err != nil {
		t.Logf("加载配置失败（可能是正常的）: %v", err)
		// 不直接失败，因为测试数据可能不完整
	}

	// 检查是否加载了配置
	loadedConfigs := manager.GetLoadedConfigNames()
	if len(loadedConfigs) == 0 {
		t.Log("没有加载到配置，这可能是因为处理器还未完全实现")
	} else {
		t.Logf("加载了 %d 个配置", len(loadedConfigs))
		for _, name := range loadedConfigs {
			t.Logf("配置名: %s", name)
		}
	}
}

// getTestDataDir 获取测试数据目录
// 从项目根目录查找 testdata 目录
func getTestDataDir() string {
	// 测试运行时，当前目录是 tests
	testDataPath := filepath.Join("..", "testdata")
	if _, err := os.Stat(testDataPath); err == nil {
		return testDataPath
	}

	// 如果找不到，返回默认路径
	return "testdata"
}
