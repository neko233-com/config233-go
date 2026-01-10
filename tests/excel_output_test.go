package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neko233-com/config233-go/pkg/config233"
	"github.com/xuri/excelize/v2"
)

// TestExcelReadAndOutputJSON 测试 Excel 读取并输出 JSON 到 CheckOutput 目录
// 用于人工检查 Excel 读取的数据是否正确转换类型
func TestExcelReadAndOutputJSON(t *testing.T) {
	// 创建 CheckOutput 目录
	outputDir := "../CheckOutput"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("创建 CheckOutput 目录失败: %v", err)
	}

	// 创建配置管理器并加载配置
	manager := config233.NewConfigManager233("../testdata")
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 获取所有加载的配置名
	configNames := manager.GetLoadedConfigNames()
	t.Logf("加载了 %d 个配置: %v", len(configNames), configNames)

	// 遍历每个配置，输出为 JSON 文件
	for _, configName := range configNames {
		allConfigs, exists := manager.GetAllConfigs(configName)
		if !exists {
			t.Logf("配置 %s 不存在", configName)
			continue
		}

		// 转为 JSON
		jsonBytes, err := json.MarshalIndent(allConfigs, "", "  ")
		if err != nil {
			t.Errorf("配置 %s 转 JSON 失败: %v", configName, err)
			continue
		}

		// 写入文件
		outputPath := filepath.Join(outputDir, configName+".json")
		if err := os.WriteFile(outputPath, jsonBytes, 0644); err != nil {
			t.Errorf("写入文件 %s 失败: %v", outputPath, err)
			continue
		}

		t.Logf("已输出 %s 到 %s (共 %d 条记录)", configName, outputPath, len(allConfigs))
	}

	t.Log("所有配置已输出到 CheckOutput/ 目录，请人工检查类型转换是否正确")
}

// TestExcelGenerateStruct 测试从 Excel 生成 Go struct 代码
func TestExcelGenerateStruct(t *testing.T) {
	outputDir := "../GeneratedStruct"

	// 从 testdata 目录生成 struct
	err := config233.GenerateStructsFromExcelDir("../testdata", outputDir)
	if err != nil {
		t.Fatalf("生成 struct 失败: %v", err)
	}

	// 检查生成的文件
	files, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("读取输出目录失败: %v", err)
	}

	t.Logf("生成了 %d 个文件:", len(files))
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			t.Logf("  - %s", file.Name())

			// 读取并打印文件内容
			content, err := os.ReadFile(filepath.Join(outputDir, file.Name()))
			if err != nil {
				t.Errorf("读取文件失败: %v", err)
				continue
			}
			t.Logf("=== %s ===\n%s", file.Name(), string(content))
		}
	}

	t.Log("所有 struct 已生成到 GeneratedStruct/ 目录")
}

// TestExcelGenerateSingleStruct 测试从单个 Excel 文件生成 struct
func TestExcelGenerateSingleStruct(t *testing.T) {
	outputDir := "../GeneratedStruct"

	// 从单个文件生成
	err := config233.GenerateStructFromExcel("../testdata/ItemConfig.xlsx", outputDir)
	if err != nil {
		t.Fatalf("生成 struct 失败: %v", err)
	}

	// 检查生成的文件
	outputPath := filepath.Join(outputDir, "itemconfig.go")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("读取生成的文件失败: %v", err)
	}

	t.Logf("生成的 ItemConfig struct:\n%s", string(content))

	// 验证内容
	contentStr := string(content)
	if !strings.Contains(contentStr, "type ItemConfig struct") {
		t.Error("生成的代码中没有找到 struct 定义")
	}
	if !strings.Contains(contentStr, "ItemId") {
		t.Error("生成的代码中没有找到 ItemId 字段")
	}
}

// TestExcelGenerateKvConfig 测试 KV 配置生成 IKvConfig 接口
func TestExcelGenerateKvConfig(t *testing.T) {
	outputDir := "../GeneratedStruct"

	// 创建一个临时的 KV 配置 Excel 文件来测试
	// 使用 excelize 创建
	f := createTestKvExcel(t)
	defer func() {
		if err := f.Close(); err != nil {
			t.Logf("关闭 Excel 文件失败: %v", err)
		}
	}()

	// 保存到临时文件
	kvExcelPath := filepath.Join("../testdata", "TestKvConfig.xlsx")
	if err := f.SaveAs(kvExcelPath); err != nil {
		t.Fatalf("保存测试 Excel 失败: %v", err)
	}
	defer os.Remove(kvExcelPath) // 测试后删除

	// 生成 struct
	err := config233.GenerateStructFromExcel(kvExcelPath, outputDir)
	if err != nil {
		t.Fatalf("生成 KV struct 失败: %v", err)
	}

	// 检查生成的文件
	outputPath := filepath.Join(outputDir, "testkvconfig.go")
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("读取生成的 KV 文件失败: %v", err)
	}

	contentStr := string(content)
	t.Logf("生成的 TestKvConfig struct:\n%s", contentStr)

	// 验证是否包含 IKvConfig 接口实现
	if !strings.Contains(contentStr, "GetValue()") {
		t.Error("KV 配置没有生成 GetValue 方法")
	}
	if !strings.Contains(contentStr, "IKvConfig") {
		t.Error("KV 配置没有标注 IKvConfig 接口实现")
	}
}

// createTestKvExcel 创建测试用的 KV Excel 文件
func createTestKvExcel(t *testing.T) *excelize.File {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// 设置行数据
	// 第 1 行：注释
	f.SetCellValue(sheet, "A1", "注释")
	f.SetCellValue(sheet, "B1", "配置ID")
	f.SetCellValue(sheet, "C1", "配置值")

	// 第 2 行：中文字段名
	f.SetCellValue(sheet, "A2", "中文")
	f.SetCellValue(sheet, "B2", "ID")
	f.SetCellValue(sheet, "C2", "值")

	// 第 3 行：Client
	f.SetCellValue(sheet, "A3", "Client")
	f.SetCellValue(sheet, "B3", "id")
	f.SetCellValue(sheet, "C3", "value")

	// 第 4 行：type
	f.SetCellValue(sheet, "A4", "type")
	f.SetCellValue(sheet, "B4", "string")
	f.SetCellValue(sheet, "C4", "string")

	// 第 5 行：Server
	f.SetCellValue(sheet, "A5", "Server")
	f.SetCellValue(sheet, "B5", "id")
	f.SetCellValue(sheet, "C5", "value")

	// 第 6+ 行：数据
	f.SetCellValue(sheet, "A6", "")
	f.SetCellValue(sheet, "B6", "config1")
	f.SetCellValue(sheet, "C6", "value1")

	f.SetCellValue(sheet, "A7", "")
	f.SetCellValue(sheet, "B7", "config2")
	f.SetCellValue(sheet, "C7", "value2")

	return f
}

// TestExcelTypeConversion 测试 Excel 类型转换是否正确
func TestExcelTypeConversion(t *testing.T) {
	// 创建配置管理器并加载配置
	manager := config233.NewConfigManager233("../testdata")
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 测试 ItemConfig
	itemConfig, exists := manager.GetConfig("ItemConfig", "1")
	if !exists {
		t.Fatal("ItemConfig 中 ID=1 的配置不存在")
	}

	item, ok := itemConfig.(map[string]interface{})
	if !ok {
		t.Fatalf("配置类型不是 map[string]interface{}, 实际类型: %T", itemConfig)
	}

	// 检查字段类型
	t.Logf("ItemConfig[1] = %+v", item)

	// 检查 itemId 字段（应该是 int64 或 int）
	if itemId, ok := item["itemId"]; ok {
		switch v := itemId.(type) {
		case int64:
			t.Logf("✓ itemId 是 int64 类型: %d", v)
		case int:
			t.Logf("✓ itemId 是 int 类型: %d", v)
		case string:
			t.Errorf("✗ itemId 是 string 类型（未正确转换）: %s", v)
		default:
			t.Errorf("✗ itemId 是未知类型: %T = %v", itemId, itemId)
		}
	}

	// 检查 quality 字段（应该是 int）
	if quality, ok := item["quality"]; ok {
		switch v := quality.(type) {
		case int64:
			t.Logf("✓ quality 是 int64 类型: %d", v)
		case int:
			t.Logf("✓ quality 是 int 类型: %d", v)
		case string:
			t.Errorf("✗ quality 是 string 类型（未正确转换）: %s", v)
		default:
			t.Errorf("✗ quality 是未知类型: %T = %v", quality, quality)
		}
	}

	// 检查 type 字段（根据 Excel 定义应该是 int）
	if typeVal, ok := item["type"]; ok {
		switch v := typeVal.(type) {
		case int64:
			t.Logf("✓ type 是 int64 类型: %d", v)
		case int:
			t.Logf("✓ type 是 int 类型: %d", v)
		case string:
			t.Errorf("✗ type 是 string 类型（未正确转换）: %s", v)
		default:
			t.Errorf("✗ type 是未知类型: %T = %v", typeVal, typeVal)
		}
	}
}
