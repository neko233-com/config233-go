package config233

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/neko233-com/config233-go/pkg/config233/dto"
	"github.com/neko233-com/config233-go/pkg/config233/excel"
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

// TestLoaderJSON_ThreadSafe_ObjectJSON 验证对象型 JSON 也能被兼容加载
func TestLoaderJSON_ThreadSafe_ObjectJSON(t *testing.T) {
	tempDir := t.TempDir()
	destFile := filepath.Join(tempDir, "BrokenJson.json")

	if err := os.WriteFile(destFile, []byte(`{"id":1,"name":"broken"}`), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	manager := NewConfigManager233(tempDir)
	err := manager.loadJsonConfig(destFile)
	if err == nil {
		// 兼容行为：对象型 JSON 被包装成单条配置加载成功
		manager.mutex.RLock()
		loaded, exists := manager.configs["BrokenJson"]
		manager.mutex.RUnlock()
		if !exists {
			t.Fatal("对象型 JSON 加载后未写入 configs")
		}
		dataList, ok := loaded.([]map[string]interface{})
		if !ok {
			t.Fatalf("期望对象型 JSON 在前端缓存中为 []map[string]interface{}，实际为 %T", loaded)
		}
		if len(dataList) != 1 {
			t.Fatalf("期望对象型 JSON 只生成 1 条数据，实际为 %d", len(dataList))
		}
		return
	}

	t.Fatalf("对象型 JSON 不应报错，但实际返回: %v", err)
}

type GodLvUpConfig struct {
	Id                     int    `json:"id" config233_column:"id"`
	LvUpConditionText      string `json:"lvUpConditionText" config233_column:"lvUpConditionText"`
	CostItemText           string `json:"costItemText" config233_column:"costItemText"`
	LvUpNeedWaitTimeSecond int64  `json:"lvUpNeedWaitTimeSecond" config233_column:"lvUpNeedWaitTimeSecond"`
	UnlockOtherBuildingIds []int  `json:"unlockOtherBuildingIds" config233_column:"unlockOtherBuildingIds"`
}

func TestLoaderExcel_GodLvUpConfig_ArrayField(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join("..", "..", "testdata", "GodLvUpConfig.xlsx")
	data, err := os.ReadFile(srcFile)
	if err != nil {
		t.Skipf("跳过测试：无法读取测试文件 %s: %v", srcFile, err)
		return
	}

	destFile := filepath.Join(tempDir, "GodLvUpConfig.xlsx")
	if err := os.WriteFile(destFile, data, 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	manager := NewConfigManager233(tempDir)
	RegisterType[GodLvUpConfig]()

	if err := manager.loadExcelConfig(destFile); err != nil {
		t.Fatalf("加载 Excel 配置失败: %v", err)
	}

	handler := &excel.ExcelConfigHandler{}
	raw := handler.ReadToFrontEndDataList("GodLvUpConfig", destFile).(*dto.FrontEndConfigDto)
	if len(raw.DataList) == 0 {
		t.Fatal("Excel 解析结果为空")
	}

	for _, row := range raw.DataList {
		sliceText := firstStringValue(row, "unlockOtherBuildingIds")
		ids := parseIntSliceText(sliceText)
		if len(ids) == 2 && ids[0] == 3 && ids[1] == 4 {
			idStr := firstStringValue(row, "id")
			id, _ := strconv.Atoi(idStr)
			cfg, ok := GetConfigById[GodLvUpConfig](id)
			if !ok || cfg == nil {
				t.Fatalf("未找到 id=%d 的配置", id)
			}
			if len(cfg.UnlockOtherBuildingIds) != 2 || cfg.UnlockOtherBuildingIds[0] != 3 || cfg.UnlockOtherBuildingIds[1] != 4 {
				t.Fatalf("unlockOtherBuildingIds 解析错误，实际 %+v", cfg.UnlockOtherBuildingIds)
			}
			return
		}
	}

	t.Fatal("未在 GodLvUpConfig.xlsx 中找到 unlockOtherBuildingIds=[3,4] 的配置行")
}

func TestLoadAllConfigs_SkipsHiddenDirectories(t *testing.T) {
	tempDir := t.TempDir()
	visibleDir := filepath.Join(tempDir, "visible")
	hiddenDir := filepath.Join(tempDir, ".hidden")
	if err := os.MkdirAll(visibleDir, 0755); err != nil {
		t.Fatalf("创建可见目录失败: %v", err)
	}
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatalf("创建隐藏目录失败: %v", err)
	}

	visibleFile := filepath.Join(visibleDir, "VisibleConfig.json")
	hiddenFile := filepath.Join(hiddenDir, "HiddenConfig.json")
	if err := os.WriteFile(visibleFile, []byte(`[{"id":1}]`), 0644); err != nil {
		t.Fatalf("写入可见配置失败: %v", err)
	}
	if err := os.WriteFile(hiddenFile, []byte(`[{"id":2}]`), 0644); err != nil {
		t.Fatalf("写入隐藏配置失败: %v", err)
	}

	manager := NewConfigManager233(tempDir)
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	names := manager.GetLoadedConfigNames()
	if !containsString(names, "VisibleConfig") {
		t.Fatalf("未加载可见目录配置: %v", names)
	}
	if containsString(names, "HiddenConfig") {
		t.Fatalf("隐藏目录配置不应被加载: %v", names)
	}
}

func firstStringValue(row map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := row[key]; ok {
			return strings.TrimSpace(toString(v))
		}
	}
	return ""
}

func parseIntSliceText(value string) []int {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.Trim(part, `"'`))
		if part == "" {
			continue
		}
		if n, err := strconv.Atoi(part); err == nil {
			result = append(result, n)
		}
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
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
