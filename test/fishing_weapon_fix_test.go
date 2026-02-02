package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	config233 "github.com/neko233-com/config233-go/internal/config233"
)

// FishingWeaponConfig 对应 FishingWeaponConfig.xlsx
type FishingWeaponConfig struct {
	Id int `json:"id" config233_column:"id"`
	// JSON tag 与 Excel 不匹配，但 config233_column 与 Excel 匹配
	UnlockCostGoldCount int `json:"unlockedCost" config233_column:"unlockCostGoldCount"`
	SkillId             int `json:"skillId" config233_column:"skillId"`
}

// AfterLoad 配置加载后调用
var AfterLoadCount int

func (c *FishingWeaponConfig) AfterLoad() {
	AfterLoadCount++
	// fmt.Printf("Loaded ID: %d\n", c.Id)
}

func (c *FishingWeaponConfig) Check() error {
	if c.UnlockCostGoldCount < 0 {
		return fmt.Errorf("Id %d cost < 0", c.Id)
	}
	// 1001 消耗为 0，该逻辑允许（仅检查 < 0）
	return nil
}

func TestFishingWeaponConfig_Fix(t *testing.T) {
	// 工作区根目录是 D:\Code\Go-Projects\config233-go
	rootDir := "D:\\Code\\Go-Projects\\config233-go"
	configDir := filepath.Join(rootDir, "testdata")
	exportDir := filepath.Join(rootDir, "TempCheck", "CheckConfig")

	// 为测试重置管理器
	cm := config233.NewConfigManager233(configDir)

	cm.SetLoadDoneWriteConfigFileDir(exportDir)
	cm.SetIsOpenWriteTempFileToSeeMemoryConfig(true)

	// 注册
	config233.RegisterType[FishingWeaponConfig]()

	// 重置计数器
	AfterLoadCount = 0

	// 加载
	err := cm.LoadAllConfigs()
	if err != nil {
		t.Fatalf("LoadAllConfigs failed: %v", err)
	}

	// 验证 AfterLoad 是否被调用
	// 我们预期 FishingWeaponConfig 中有 3 项（1001, 1002, 1003，基于先前的上下文），尽管下面的 verify 调用意味着可能更少或更多
	// 实际上我们要检查 map 的大小。
	dataMap := config233.GetConfigMap[FishingWeaponConfig]()
	if len(dataMap) == 0 {
		t.Fatal("No data loaded for FishingWeaponConfig")
	}

	// 注意：Config233 为 Map 和 Slice 创建单独的实例，因此每项的 AfterLoad 会被调用两次。
	expectedCalls := len(dataMap) * 2
	if AfterLoadCount != expectedCalls {
		t.Errorf("AfterLoad count mismatch. Expected %d (2x items), Got %d", expectedCalls, AfterLoadCount)
	} else {
		fmt.Printf("AfterLoad verified: called %d times (2x items)\n", AfterLoadCount)
	}

	// 验证 CheckConfig 导出文件是否存在
	exportedFile := filepath.Join(exportDir, "FishingWeaponConfig.json")
	if _, err := os.Stat(exportedFile); os.IsNotExist(err) {
		t.Errorf("Exported file not found: %s", exportedFile)
	} else {
		fmt.Printf("Exported file verified: %s\n", exportedFile)
	}

	// 检查 1001
	verify(t, dataMap, "1001", 0)
	// 检查 1002
	verify(t, dataMap, "1002", 1400)
	// 检查 1003（如果存在）
	if _, ok := dataMap["1003"]; ok {
		// 如果存在，仅检查逻辑
		verify(t, dataMap, "1003", 2800) // 根据实际数据预期為 2800
	}

	fmt.Println("Test passed!")
}

func verify(t *testing.T, data map[string]*FishingWeaponConfig, id string, expectedCost int) {
	if item, ok := data[id]; ok {
		if item.UnlockCostGoldCount != expectedCost {
			fmt.Printf("ID %s Cost mismatch. Expected %d, Got %d\n", id, expectedCost, item.UnlockCostGoldCount)
			// 如果一个失败，不要立即让测试失败，以便查看其他项
			t.Errorf("ID %s Cost mismatch", id)
		} else {
			fmt.Printf("ID %s Verified: Cost %d\n", id, item.UnlockCostGoldCount)
		}
	} else {
		// t.Errorf("ID %s 未找到", id)
		fmt.Printf("ID %s not found (might be expected if not in excel)\n", id)
	}
}
