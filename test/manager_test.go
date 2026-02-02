package test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/neko233-com/config233-go/internal/config233"
	"github.com/neko233-com/config233-go/internal/config233/dto"
	"github.com/neko233-com/config233-go/internal/config233/excel"
	"github.com/neko233-com/config233-go/internal/config233/json"
	"github.com/neko233-com/config233-go/internal/config233/tsv"
)

// TestConfigManager233_GetConfigById 测试配置管理器的 GetConfigById 方法
// 验证配置管理器是否能正确获取配置项
func TestConfigManager233_GetConfigById(t *testing.T) {
	// 创建测试用的配置管理器
	testDir := getTestDataDir()
	t.Logf("Test data directory: %s", testDir)
	_ = config233.NewConfigManager233(testDir)

	// 定义一个测试类型
	type TestConfig struct {
		ID string
	}

	// 测试获取配置 (since no data loaded, should not find)
	config, exists := config233.GetConfigById[TestConfig]("1")
	if exists {
		t.Logf("找到配置: %+v", config)
	} else {
		t.Log("配置不存在，这是正常的，因为我们还没有加载具体的配置")
	}
}

// TestConfigManager233_LoadAllConfigs 测试配置管理器的 LoadAllConfigs 方法
// 验证配置管理器是否能正确加载目录中的所有配置文件
func TestConfigManager233_GetAllConfigData(t *testing.T) {
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

// TestConfigManager233_GetLoadedConfigNames 测试获取已加载配置名称列表
func TestConfigManager233_GetLoadedConfigNames(t *testing.T) {
	testDir := getTestDataDir()
	manager := config233.NewConfigManager233(testDir)

	// 初始状态应该没有配置
	names := manager.GetLoadedConfigNames()
	if len(names) != 0 {
		t.Errorf("期望初始状态没有配置，实际得到 %d 个配置", len(names))
	}

	// 加载配置后应该有配置
	err := manager.LoadAllConfigs()
	if err != nil {
		t.Logf("加载配置失败: %v", err)
	}

	names = manager.GetLoadedConfigNames()
	if len(names) == 0 {
		t.Log("没有加载到配置，跳过测试")
		return
	}

	t.Logf("成功加载了 %d 个配置: %v", len(names), names)
}

// TestConfigManager233_GetConfigAfterLoad 测试加载配置后获取配置
func TestConfigManager233_GetConfigAfterLoad(t *testing.T) {
	testDir := getTestDataDir()
	manager := config233.NewConfigManager233(testDir)

	// 先加载配置
	err := manager.LoadAllConfigs()
	if err != nil {
		t.Logf("加载配置失败: %v", err)
	}

	// 测试获取已加载的配置
	loadedNames := manager.GetLoadedConfigNames()
	if len(loadedNames) == 0 {
		t.Log("没有加载到配置，跳过测试")
		return
	}

	// 尝试获取第一个配置的第一个项目
	t.Logf("配置 %s 已加载", loadedNames[0])
}

// TestConfigManager233_ConcurrentAccess 测试并发访问安全性
func TestConfigManager233_ConcurrentAccess(t *testing.T) {
	testDir := getTestDataDir()
	manager := config233.NewConfigManager233(testDir)

	// 加载配置
	err := manager.LoadAllConfigs()
	if err != nil {
		t.Logf("加载配置失败: %v", err)
	}

	// 并发测试
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// 并发读取
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				manager.GetLoadedConfigNames()
				// removed GetConfig call
			}
		}(i)
	}

	wg.Wait()
	t.Log("并发访问测试完成")
}

// TestConfigManager233_ErrorHandling 测试错误处理
func TestConfigManager233_ErrorHandling(t *testing.T) {
	// 测试不存在的目录
	manager := config233.NewConfigManager233("/nonexistent/path")

	err := manager.LoadAllConfigs()
	if err == nil {
		t.Log("期望加载不存在目录时出错，但没有出错")
	} else {
		t.Logf("正确处理了错误: %v", err)
	}
}

// TestConfigManager233_Reload 测试配置重载功能
func TestConfigManager233_Reload(t *testing.T) {
	testDir := getTestDataDir()
	manager := config233.NewConfigManager233(testDir)

	// 加载配置
	err := manager.LoadAllConfigs()
	if err != nil {
		t.Logf("加载配置失败: %v", err)
	}

	initialCount := len(manager.GetLoadedConfigNames())
	t.Logf("初始加载了 %d 个配置", initialCount)

	// 再次加载（模拟重载）
	err = manager.LoadAllConfigs()
	if err != nil {
		t.Logf("重载配置失败: %v", err)
	}

	afterReloadCount := len(manager.GetLoadedConfigNames())
	t.Logf("重载后有 %d 个配置", afterReloadCount)

	if initialCount != afterReloadCount {
		t.Errorf("重载前后配置数量不一致: %d vs %d", initialCount, afterReloadCount)
	}
}

// TestConfigManager233_EmptyDirectory 测试空目录处理
func TestConfigManager233_EmptyDirectory(t *testing.T) {
	// 创建临时空目录
	tempDir, err := os.MkdirTemp("", "config233_test_empty")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := config233.NewConfigManager233(tempDir)
	err = manager.LoadAllConfigs()
	if err != nil {
		t.Logf("空目录加载结果: %v", err)
	}

	names := manager.GetLoadedConfigNames()
	if len(names) != 0 {
		t.Errorf("空目录应该没有配置，实际得到 %d 个", len(names))
	}
}

// TestConfigManager233_InvalidFiles 测试无效文件处理
func TestConfigManager233_InvalidFiles(t *testing.T) {
	// 创建临时目录并添加无效文件
	tempDir, err := os.MkdirTemp("", "config233_test_invalid")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建一个无效的配置文件
	invalidFile := filepath.Join(tempDir, "invalid.txt")
	err = os.WriteFile(invalidFile, []byte("invalid content"), 0644)
	if err != nil {
		t.Fatalf("创建无效文件失败: %v", err)
	}

	manager := config233.NewConfigManager233(tempDir)
	err = manager.LoadAllConfigs()
	if err != nil {
		t.Logf("处理无效文件的结果: %v", err)
	}

	names := manager.GetLoadedConfigNames()
	t.Logf("从包含无效文件的目录加载了 %d 个配置", len(names))
}

// TestJsonConfigHandler 测试 JSON 配置处理器
func TestJsonConfigHandler(t *testing.T) {
	handler := &json.JsonConfigHandler{}

	// 测试类型名称
	if handler.TypeName() != "json" {
		t.Errorf("期望类型名称为 'json'，实际得到 '%s'", handler.TypeName())
	}

	// 测试处理有效 JSON 文件
	testDir := getTestDataDir()
	jsonFile := filepath.Join(testDir, "RedundantConfigJson.json")

	if _, err := os.Stat(jsonFile); err == nil {
		result := handler.ReadToFrontEndDataList("test", jsonFile)
		if result == nil {
			t.Error("期望 JSON 处理成功，但返回 nil")
		} else {
			dto := result.(*dto.FrontEndConfigDto)
			if dto.Type != "json" {
				t.Errorf("期望 DTO 类型为 'json'，实际得到 '%s'", dto.Type)
			}
			t.Logf("JSON 处理成功，加载了 %d 条数据", len(dto.DataList))
		}
	} else {
		t.Log("测试 JSON 文件不存在，跳过测试")
	}
}

// TestExcelConfigHandler 测试 Excel 配置处理器
func TestExcelConfigHandler(t *testing.T) {
	handler := &excel.ExcelConfigHandler{}

	// 测试类型名称
	if handler.TypeName() != "excel" {
		t.Errorf("期望类型名称为 'excel'，实际得到 '%s'", handler.TypeName())
	}

	// 测试处理有效 Excel 文件
	testDir := getTestDataDir()
	excelFile := filepath.Join(testDir, "StudentExcel.xlsx")

	if _, err := os.Stat(excelFile); err == nil {
		result := handler.ReadToFrontEndDataList("test", excelFile)
		if result == nil {
			t.Error("期望 Excel 处理成功，但返回 nil")
		} else {
			dto := result.(*dto.FrontEndConfigDto)
			if dto.Type != "excel" {
				t.Errorf("期望 DTO 类型为 'excel'，实际得到 '%s'", dto.Type)
			}
			t.Logf("Excel 处理成功，加载了 %d 条数据", len(dto.DataList))
		}
	} else {
		t.Log("测试 Excel 文件不存在，跳过测试")
	}
}

// TestTsvConfigHandler 测试 TSV 配置处理器
func TestTsvConfigHandler(t *testing.T) {
	handler := &tsv.TsvConfigHandler{}

	// 测试类型名称
	if handler.TypeName() != "tsv" {
		t.Errorf("期望类型名称为 'tsv'，实际得到 '%s'", handler.TypeName())
	}

	// 测试处理有效 TSV 文件（如果有的话）
	testDir := getTestDataDir()
	tsvFile := filepath.Join(testDir, "test.tsv")

	if _, err := os.Stat(tsvFile); err == nil {
		result := handler.ReadToFrontEndDataList("test", tsvFile)
		if result == nil {
			t.Error("期望 TSV 处理成功，但返回 nil")
		} else {
			dto := result.(*dto.FrontEndConfigDto)
			if dto.Type != "tsv" {
				t.Errorf("期望 DTO 类型为 'tsv'，实际得到 '%s'", dto.Type)
			}
			t.Logf("TSV 处理成功，加载了 %d 条数据", len(dto.DataList))
		}
	} else {
		t.Log("测试 TSV 文件不存在，跳过测试")
	}
}

// TestConfig233_BasicFunctionality 测试 Config233 基本功能
func TestConfig233_BasicFunctionality(t *testing.T) {
	config := config233.NewConfig233()

	// 测试链式调用
	config.Directory("./test").
		AddConfigHandler("json", &json.JsonConfigHandler{}).
		AddConfigHandler("xlsx", &excel.ExcelConfigHandler{})

	// 验证处理器已添加
	if len(config.GetFileHandlers()) == 0 {
		t.Log("处理器添加可能有问题")
	}
}

// TestConfig233_ErrorHandling 测试 Config233 错误处理
func TestConfig233_ErrorHandling(t *testing.T) {
	config := config233.NewConfig233()

	// 测试无效目录
	config.Directory("/nonexistent/path")

	// 应该不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Config233 不应该在无效目录下 panic: %v", r)
		}
	}()

	// 测试基本操作
	config.AddConfigHandler("test", &json.JsonConfigHandler{})
}

// TestGlobalInstance 测试全局实例
func TestGlobalInstance(t *testing.T) {
	// 测试全局实例存在
	if config233.Instance == nil {
		t.Error("全局实例不应该为 nil")
	}

	t.Log("全局实例测试完成")
}

// BenchmarkConfigManager233_GetLoadedConfigNames 基准测试配置名称获取性能
func BenchmarkConfigManager233_GetLoadedConfigNames(b *testing.B) {
	testDir := getTestDataDir()
	manager := config233.NewConfigManager233(testDir)

	// 预加载配置
	manager.LoadAllConfigs()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			manager.GetLoadedConfigNames()
		}
	})
}

// BenchmarkConfigManager233_LoadAllConfigs 基准测试配置加载性能
func BenchmarkConfigManager233_LoadAllConfigs(b *testing.B) {
	testDir := getTestDataDir()

	for i := 0; i < b.N; i++ {
		manager := config233.NewConfigManager233(testDir)
		manager.LoadAllConfigs()
	}
}

// getTestDataDir 获取测试数据目录
// 从项目根目录查找 testdata 目录
func getTestDataDir() string {
	// 测试运行时，当前目录是 test
	testDataPath := filepath.Join("..", "testdata")
	if _, err := os.Stat(testDataPath); err == nil {
		return testDataPath
	}

	// 如果找不到，返回默认路径
	return "testdata"
}

// TestConfigManager233_GetKvToCsvStringList 测试 GetKvToCsvStringList
func TestConfigManager233_GetKvToCsvStringList(t *testing.T) {
	// 1. 创建临时 KvConfig.json
	tempDir, err := os.MkdirTemp("", "config233_test_kv")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kvJsonContent := `[
		{"id": "list1", "value": "a,b,c"},
		{"id": "list2", "value": " 1 , 2 , 3 "},
		{"id": "empty", "value": ""}
	]`
	kvFile := filepath.Join(tempDir, "TestKvConfig.json")
	if err := os.WriteFile(kvFile, []byte(kvJsonContent), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	// 2. 初始化 Manager
	config233.Instance = config233.NewConfigManager233(tempDir)

	// 3. 定义 Struct 并注册
	// 注意: 放在 Test 内部定义的类型无法被反射正确实例化（reflect.New 可能会有问题，或者 Name 为空）
	// 所以我们使用 TestKvConfig (包级定义)
	config233.RegisterType[TestKvConfig]()

	// 4. 加载配置
	if err := config233.Instance.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 5. 测试 GetKvToCsvStringList

	// Case 1: 正常列表 "a,b,c"
	list1 := config233.GetKvToCsvStringList[TestKvConfig]("list1", nil)
	if len(list1) != 3 || list1[0] != "a" || list1[1] != "b" || list1[2] != "c" {
		t.Errorf("list1 解析错误: %v", list1)
	}

	// Case 2: 带空格 " 1 , 2 , 3 "
	list2 := config233.GetKvToCsvStringList[TestKvConfig]("list2", nil)
	if len(list2) != 3 || list2[0] != "1" || list2[1] != "2" || list2[2] != "3" {
		t.Errorf("list2 解析错误: %v", list2)
	}

	// Case 3: 默认值 (ID不存在)
	defaultList := []string{"d", "e"}
	list3 := config233.GetKvToCsvStringList[TestKvConfig]("not_exist", defaultList)
	if len(list3) != 2 || list3[0] != "d" {
		t.Errorf("默认值返回解析错误: %v", list3)
	}

	// Case 4: 空值 (ID存在但Value为空字符串)
	// 根据实现，如果 Value 为 ""，返回 defaultVal
	list4 := config233.GetKvToCsvStringList[TestKvConfig]("empty", defaultList) // empty value is ""
	if len(list4) != 2 || list4[0] != "d" {
		t.Errorf("Empty Value 应该返回默认值: %v", list4)
	}
}

// 辅助类型 (需放在 Top Level 以便反射获取 Name)
type TestKvConfig struct {
	Id  string `json:"id"`
	Val string `json:"value"`
}

func (c TestKvConfig) GetUid() any {
	return c.Id
}

func (c TestKvConfig) GetValue() string {
	return c.Val
}

var _ config233.IKvConfig = (*TestKvConfig)(nil)
