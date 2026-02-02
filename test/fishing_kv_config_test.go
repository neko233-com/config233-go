package test

import (
	"testing"

	generated "github.com/neko233-com/config233-go/GeneratedStruct"
	"github.com/neko233-com/config233-go/pkg/config233"
)

// TestFishingKvConfig_GetKvApis
// 使用实际的 FishingKvConfig.xlsx 做 KV 配置测试
//
// Excel 内容参考：
//
//	id                           value
//	autoAttackIntervalTimeSecond 1
//	test_int                     1
//	test_string                  TestValue
//	test_bool                    TRUE
func TestFishingKvConfig_GetKvApis(t *testing.T) {
	testDir := getTestDataDir()

	// 1. 初始化 Manager 并加载 ../testdata 下的 Excel 配置（包含 FishingKvConfig.xlsx）
	manager := config233.NewConfigManager233(testDir)
	config233.Instance = manager

	// 注册 FishingKvConfig 类型，确保可以映射到 struct
	config233.RegisterType[generated.FishingKvConfig]()

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 2. 逐个验证每个 key

	// 2.1 autoAttackIntervalTimeSecond -> "1"
	str := config233.GetKvToString[generated.FishingKvConfig]("autoAttackIntervalTimeSecond", "default")
	if str != "1" {
		t.Errorf("GetKvToString autoAttackIntervalTimeSecond 错误, got=%q, want=%q", str, "1")
	}

	intVal := config233.GetKvToInt[generated.FishingKvConfig]("autoAttackIntervalTimeSecond", 0)
	if intVal != 1 {
		t.Errorf("GetKvToInt autoAttackIntervalTimeSecond 错误, got=%d, want=%d", intVal, 1)
	}

	csv := config233.GetKvToCsvStringList[generated.FishingKvConfig]("autoAttackIntervalTimeSecond", nil)
	if len(csv) != 1 || csv[0] != "1" {
		t.Errorf("GetKvToCsvStringList autoAttackIntervalTimeSecond 错误, got=%v, want=[1]", csv)
	}

	// 2.2 test_int -> "1"
	int2 := config233.GetKvToInt[generated.FishingKvConfig]("test_int", 0)
	if int2 != 1 {
		t.Errorf("GetKvToInt test_int 错误, got=%d, want=%d", int2, 1)
	}

	// 2.3 test_string -> "TestValue"
	str2 := config233.GetKvToString[generated.FishingKvConfig]("test_string", "")
	if str2 != "TestValue" {
		t.Errorf("GetKvToString test_string 错误, got=%q, want=%q", str2, "TestValue")
	}

	csv2 := config233.GetKvToCsvStringList[generated.FishingKvConfig]("test_string", nil)
	if len(csv2) != 1 || csv2[0] != "TestValue" {
		t.Errorf("GetKvToCsvStringList test_string 错误, got=%v, want=[TestValue]", csv2)
	}

	// 2.4 test_bool -> "TRUE"
	boolVal := config233.GetKvToBoolean[generated.FishingKvConfig]("test_bool", false)
	if !boolVal {
		t.Errorf("GetKvToBoolean test_bool 错误, got=%v, want=%v", boolVal, true)
	}

	// 3. 未找到 ID 的默认值验证
	if v := config233.GetKvToString[generated.FishingKvConfig]("not_exist", "default"); v != "default" {
		t.Errorf("GetKvToString 默认值错误, got=%q, want=%q", v, "default")
	}

	if v := config233.GetKvToInt[generated.FishingKvConfig]("not_exist", 42); v != 42 {
		t.Errorf("GetKvToInt 默认值错误, got=%d, want=%d", v, 42)
	}

	if v := config233.GetKvToBoolean[generated.FishingKvConfig]("not_exist", true); !v {
		t.Errorf("GetKvToBoolean 默认值错误, got=%v, want=%v", v, true)
	}

	if v := config233.GetKvToCsvStringList[generated.FishingKvConfig]("not_exist", []string{"d"}); len(v) != 1 || v[0] != "d" {
		t.Errorf("GetKvToCsvStringList 默认值错误, got=%v, want=[d]", v)
	}
}
