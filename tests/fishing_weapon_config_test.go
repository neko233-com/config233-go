package tests

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	config233 "github.com/neko233-com/config233-go/pkg/config233"
	"github.com/neko233-com/config233-go/pkg/config233/dto"
	"github.com/neko233-com/config233-go/pkg/config233/excel"
)

// FishingWeaponConfig 对应 testdata/FishingWeaponConfig.xlsx（使用 Server 行作为字段名）
// 注意: Excel 的 Server 行字段为: id, skillId, unlockCostGoldCount
type FishingWeaponConfig struct {
	Id                  int `json:"id" config233_column:"id"`
	SkillId             int `json:"skillId" config233_column:"skillId"`
	UnlockCostGoldCount int `json:"unlockCostGoldCount" config233_column:"unlockCostGoldCount"`
}

// TestFishingWeaponConfig_Parse 测试 FishingWeaponConfig.xlsx 能正确加载并映射到结构体
func TestFishingWeaponConfig_Parse(t *testing.T) {
	testDir := getTestDataDir()

	// 原始解析按 id 的映射（key 为字符串 id）
	rawById := make(map[string]map[string]interface{})

	// 直接读取 Excel 并打印解析到的字���名/所有行（JSON），便于调试字段对应关系和空列情况
	t.Run("debug_read_headers_and_dump_json", func(t *testing.T) {
		h := &excel.ExcelConfigHandler{}
		full := filepath.Join(testDir, "FishingWeaponConfig.xlsx")
		res := h.ReadToFrontEndDataList("FishingWeaponConfig", full)
		d, ok := res.(*dto.FrontEndConfigDto)
		if !ok {
			t.Fatalf("无法将结果断言为 FrontEndConfigDto，可能解析失败: %#v", res)
		}

		// 打印原始 DTO 的 keys 和全部数据的 JSON（���于查看有无空列）
		if len(d.DataList) == 0 {
			t.Log("解析到的数据为空")
			return
		}

		// 遍历所有行，记录到 rawById
		for _, row := range d.DataList {
			idStr := idToString(row["id"])
			if idStr == "" {
				// 尝试其他可能的 id 字段名
				idStr = idToString(row["Id"])
			}
			if idStr == "" {
				// 跳过没有 id 的行
				continue
			}
			rawById[idStr] = row
		}

		// 查看第一行的键以确认 Server 行字段
		first := d.DataList[0]
		t.Logf("解析到的第一行字段: %v", getMapKeys(first))
		for k, v := range first {
			t.Logf("  %s = %v", k, v)
		}

		// JSON dump of the parsed rows (indent)
		if b, err := json.MarshalIndent(d.DataList, "", "  "); err == nil {
			t.Logf("解析后的 DataList JSON:\n%s", string(b))
		} else {
			t.Logf("解析 JSON 失败: %v", err)
		}
	})

	manager := config233.NewConfigManager233(testDir)
	config233.Instance = manager

	// 注册测试中的结构体类型
	config233.RegisterType[FishingWeaponConfig]()

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 要验证的 ID 列表
	ids := []int{1001, 1002, 1003}

	// 打印转换后结构体列表的 JSON（便于对比）
	if list := config233.GetConfigList[FishingWeaponConfig](); len(list) > 0 {
		if b, err := json.MarshalIndent(list, "", "  "); err == nil {
			t.Logf("转换后的结构体列表 JSON:\n%s", string(b))
		} else {
			t.Logf("序列化转换后结构体失败: %v", err)
		}
	}

	// 逐个断言三个配置
	for _, id := range ids {
		idStr := fmt.Sprintf("%d", id)

		cfg, ok := config233.GetConfigById[FishingWeaponConfig](id)
		if !ok || cfg == nil {
			t.Errorf("期望找到 id=%d 的配置，但未找到", id)
			continue
		}

		if cfg.Id != id {
			t.Errorf("id=%d: 期望 Id=%d，实际 %d", id, id, cfg.Id)
		}

		// 如果原始解析包含该 id，则按原始字段有选择地断言
		if raw, exists := rawById[idStr]; exists {
			// skillId - 只有当原始数据中存在且不为 0 时才断言
			if v, ok := raw["skillId"]; ok {
				if rawSkill := toInt(v); rawSkill != 0 {
					if cfg.SkillId != rawSkill {
						t.Errorf("id=%d: 期望 SkillId=%d，实际 %d", id, rawSkill, cfg.SkillId)
					} else {
						t.Logf("id=%d: SkillId=%d ✓", id, cfg.SkillId)
					}
				}
			}

			// unlockCostGoldCount - 允许为 0
			if v, ok := raw["unlockCostGoldCount"]; ok {
				rawCost := toInt(v)
				if cfg.UnlockCostGoldCount != rawCost {
					t.Errorf("id=%d: 期望 UnlockCostGoldCount=%d，实际 %d", id, rawCost, cfg.UnlockCostGoldCount)
				} else {
					t.Logf("id=%d: UnlockCostGoldCount=%d ✓", id, cfg.UnlockCostGoldCount)
				}
			}
		} else {
			t.Logf("��始解析未包含 id=%d 的行，跳过基于原始值的断言", id)
		}
	}
}

// idToString 将可能的 id 值转换为字符串
func idToString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case int:
		return fmt.Sprintf("%d", t)
	case int32:
		return fmt.Sprintf("%d", int(t))
	case int64:
		return fmt.Sprintf("%d", int(t))
	case float32:
		return fmt.Sprintf("%d", int(t))
	case float64:
		return fmt.Sprintf("%d", int(t))
	default:
		return fmt.Sprintf("%v", t)
	}
}

// getMapKeys returns keys of a map in a slice (order not guaranteed)
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// toInt tries to convert interface{} to int (handles int, int64, float64, string)
func toInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float32:
		return int(t)
	case float64:
		return int(t)
	case string:
		// try parse
		var i int
		if err := json.Unmarshal([]byte(t), &i); err == nil {
			return i
		}
		// fallback: try strconv
		return 0
	default:
		return 0
	}
}
