package test

import (
	"testing"

	config233 "github.com/neko233-com/config233-go/internal/config233"
)

// ItemConfig 对应 ItemConfig.xlsx 文件的结构体
type ItemConfig struct {
	Itemid           int64  `json:"itemId"`
	Itemname         string `json:"itemName"`
	Desc             string `json:"desc"`
	Bagtype          string `json:"bagType"`
	Type             int    `json:"type"`
	Expiretimems     int64  `json:"expireTimeMs"`
	Quality          int    `json:"quality"`
	Icon             string `json:"icon"`
	Effect           string `json:"effect"`
	Useconditionlist string `json:"useConditionList"`
	Useitemcontext   string `json:"useItemContext"`
	Stacknumber      int    `json:"stackNumber"`
	Isautouse        bool   `json:"isAutoUse,string"`
	Jumpid           int    `json:"jumpId"`
	Sort             int    `json:"sort"`
}

// TestItemConfig_LoadAndQuery 测试加载 ItemConfig 后查询 id=1 和 getDataList
func TestItemConfig_LoadAndQuery(t *testing.T) {
	// 1. 创建配置管理器并指向测试数据目录
	testDir := getTestDataDir()
	manager := config233.NewConfigManager233(testDir)
	config233.Instance = manager

	// 2. 注册 ItemConfig 类型
	config233.RegisterType[ItemConfig]()

	// 3. 加载所有配置
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 4. 测试查询 id=1 的配置
	t.Run("查询 id=1", func(t *testing.T) {
		cfg, ok := config233.GetConfigById[ItemConfig](1)
		if !ok || cfg == nil {
			t.Fatalf("期望找到 id=1 的配置，但未找到")
		}

		// 验证配置数据
		if cfg.Itemid != 1 {
			t.Errorf("期望 Itemid=1，实际得到 %d", cfg.Itemid)
		}

		// 验证其他字段不为空（至少应该有名称）
		if cfg.Itemname == "" {
			t.Error("期望 Itemname 不为空")
		}

		t.Logf("成功查询到配置: Itemid=%d, Itemname=%s", cfg.Itemid, cfg.Itemname)
	})

	// 5. 测试获取所有配置列表
	t.Run("获取配置列表", func(t *testing.T) {
		list := config233.GetConfigList[ItemConfig]()
		if len(list) == 0 {
			t.Fatalf("期望配置列表不为空，但得到空列表")
		}

		// 验证列表中的配置
		foundId1 := false
		for _, cfg := range list {
			if cfg == nil {
				t.Error("配置列表中不应包含 nil")
				continue
			}
			if cfg.Itemid == 1 {
				foundId1 = true
			}
		}

		if !foundId1 {
			t.Error("配置列表中应包含 id=1 的配置")
		}

		t.Logf("成功获取配置列表，共 %d 条配置", len(list))
	})

	// 6. 测试使用字符串 ID 查询
	t.Run("使用字符串 ID 查询", func(t *testing.T) {
		cfg, ok := config233.GetConfigById[ItemConfig]("1")
		if !ok || cfg == nil {
			t.Fatalf("期望使用字符串 '1' 能找到配置，但未找到")
		}

		if cfg.Itemid != 1 {
			t.Errorf("期望 Itemid=1，实际得到 %d", cfg.Itemid)
		}

		t.Logf("成功使用字符串 ID 查询到配置: Itemid=%d", cfg.Itemid)
	})
}
