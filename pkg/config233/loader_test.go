package config233

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoaderExcel_ThreadSafe 测试 Excel 加载器的线程安全性
func TestLoaderExcel_ThreadSafe(t *testing.T) {
	// 创建临时配置目录
	tempDir := t.TempDir()

	// 复制测试 Excel 文件
	testFile := filepath.Join("..", "..", "testdata", "ItemConfig.xlsx")
	destFile := filepath.Join(tempDir, "ItemConfig.xlsx")

	// 读取源文件
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Skipf("跳过测试：无法读取测试文件 %s: %v", testFile, err)
		return
	}

	// 写入目标文件
	if err := os.WriteFile(destFile, data, 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 创建配置管理器
	manager := NewConfigManager233(tempDir)

	// 注册类型（如果需要）
	// RegisterType[ItemConfig]()

	// 测试加载
	err = manager.loadExcelConfig(destFile)
	if err != nil {
		t.Fatalf("加载 Excel 配置失败: %v", err)
	}

	// 验证配置已加载
	manager.mutex.RLock()
	_, exists := manager.configs["ItemConfig"]
	manager.mutex.RUnlock()

	if !exists {
		t.Error("Excel 配置未成功加载")
	}

	t.Log("Excel 加载器线程安全测试通过")
}

// TestLoaderJSON_ThreadSafe 测试 JSON 加载器的线程安全性
func TestLoaderJSON_ThreadSafe(t *testing.T) {
	// 创建临时配置目录
	tempDir := t.TempDir()

	// 复制测试 JSON 文件
	testFile := filepath.Join("..", "..", "testdata", "configDir", "json", "StudentJson.json")
	destFile := filepath.Join(tempDir, "StudentJson.json")

	// 读取源文件
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Skipf("跳过测试：无法读取测试文件 %s: %v", testFile, err)
		return
	}

	// 写入目标文件
	if err := os.WriteFile(destFile, data, 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 创建配置管理器
	manager := NewConfigManager233(tempDir)

	// 测试加载
	err = manager.loadJsonConfig(destFile)
	if err != nil {
		t.Fatalf("加载 JSON 配置失败: %v", err)
	}

	// 验证配置已加载
	manager.mutex.RLock()
	_, exists := manager.configs["StudentJson"]
	manager.mutex.RUnlock()

	if !exists {
		t.Error("JSON 配置未成功加载")
	}

	t.Log("JSON 加载器线程安全测试通过")
}

// TestLoaderTSV_ThreadSafe 测试 TSV 加载器的线程安全性
func TestLoaderTSV_ThreadSafe(t *testing.T) {
	// 创建临时配置目录和 TSV 测试文件
	tempDir := t.TempDir()
	destFile := filepath.Join(tempDir, "TestConfig.tsv")

	// 创建简单的 TSV 内容
	tsvContent := "id\tname\tvalue\n1\ttest1\t100\n2\ttest2\t200\n"
	if err := os.WriteFile(destFile, []byte(tsvContent), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 创建配置管理器
	manager := NewConfigManager233(tempDir)

	// 测试加载
	err := manager.loadTsvConfig(destFile)
	if err != nil {
		t.Fatalf("加载 TSV 配置失败: %v", err)
	}

	// 验证配置已加载
	manager.mutex.RLock()
	_, exists := manager.configs["TestConfig"]
	manager.mutex.RUnlock()

	if !exists {
		t.Error("TSV 配置未成功加载")
	}

	t.Log("TSV 加载器线程安全测试通过")
}
