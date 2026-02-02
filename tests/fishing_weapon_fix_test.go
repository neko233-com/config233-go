package tests

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	config233 "github.com/neko233-com/config233-go/pkg/config233"
)

// FishingWeaponConfigFix 对应 testdata/FishingWeaponConfig.xlsx
// 使用 config233_column 解决 Excel 列名与 JSON tag 不一致的问题
type FishingWeaponConfigFix struct {
	Id                  int `json:"id" config233_column:"id"`
	UnlockCostGoldCount int `json:"unlockedCost" config233_column:"unlockCostGoldCount"` // Excel 列名 unlockCostGoldCount，JSON tag unlockedCost
	SkillId             int `json:"skillId" config233_column:"skillId"`
}

// AfterLoad 配置加载后调用
// 可以在这里进行数据预处理、建立索引、缓存分组等
func (c *FishingWeaponConfigFix) AfterLoad() {
	// 实现配置预处理逻辑
	fmt.Printf("[FishingWeaponConfigFix] AfterLoad called for ID: %d\n", c.Id)
}

// Check 配置校验
// 返回 nil 表示校验通过，否则返回错误信息
func (c *FishingWeaponConfigFix) Check() error {
	if c.UnlockCostGoldCount <= 0 {
		// 这里直接打印日志，实际项目中可以使用 logger233.Error
		// 注意：1001 的价格为 0，所以会触发这个错误，这是预期行为
		return fmt.Errorf("FishingWeaponConfig.id=%d 解锁价格必须>0: unlockCostGoldCount=%d", c.Id, c.UnlockCostGoldCount)
	}
	return nil
}

func TestFishingWeaponConfig_Fix(t *testing.T) {
	// 设置配置目录
	configDir := "../testdata"
	exportDir := "../TempCheck/CheckConfig"

	// 获取管理器实例
	cm := config233.GetInstance()

	// 重置管理器状态（为了测试）
	// 注意：ConfigManager233 设计为单例，这里我们通过重新设置配置目录来模拟
	// 如果已经启动，SetConfigDir 会报错，但我们可以忽略，只要确保路径正确
	cm.SetConfigDir(configDir)

	// 开启导出功能
	cm.SetIsOpenWriteTempFileToSeeMemoryConfig(true)
	cm.SetLoadDoneWriteConfigFileDir(exportDir)

	// 注册类型
	config233.RegisterType[FishingWeaponConfigFix]()

	// 加载 Excel 配置 (FishingWeaponConfig.xlsx)
	// 我们手动调用 loadExcelConfigThreadSafe 或者 LoadAllConfigs
	// 这里为了测试方便，我们调用 LoadAllConfigs
	// 注意：LoadAllConfigs 会加载目录下所有文件，我们主要关注 FishingWeaponConfig
	// 由于我们的 Struct 名字是 FishingWeaponConfigFix，而文件名是 FishingWeaponConfig
	// 默认自动匹配是 FilesName -> StructName
	// 但是这里不匹配。

	// ConfigManager 默认根据文件名查找 registeredTypes。
	// FishingWeaponConfig.xlsx -> "FishingWeaponConfig"
	// 我们的 Struct 注册名为 "FishingWeaponConfigFix"
	// 所以需要让它们匹配。
	// 我们需手动注册： "FishingWeaponConfig" -> FishingWeaponConfigFix
	// 但是 config233.RegisterType[T] 使用 T 的 Name 作为 Key。

	// Hack: 我们需要让 ConfigManager 知道 "FishingWeaponConfig" 对应 FishingWeaponConfigFix
	// 但是目前的 RegisterType 只能注册 StructName。
	// 所以我们应该把 struct 命名为 FishingWeaponConfig。
	// 但是 tests 包下可能已经有 FishingWeaponConfig 定义了。
}

// 为了避免命名冲突，我们在单独的函数内定义 struct 或者使用别名？
// Go 不支持在函数内定义 struct 用于反射创建（reflect.New 需要全局类型 or visible type? No, it works but names are tricky)
// 我们可以临时修改 RegisterType 的逻辑吗？
// 或者我们覆盖 map?
// 我们可以手动注册: cm.RegisterType(reflect.TypeOf(FishingWeaponConfigFix{}))
// 但是 Name 是 "FishingWeaponConfigFix".
// 这意味着文件名必须是 "FishingWeaponConfigFix.xlsx"。
// 我们可以拷贝文件。

func TestFishingWeaponConfig_Fix_Run(t *testing.T) {
	// 准备：拷贝 FishingWeaponConfig.xlsx 为 FishingWeaponConfigFix.xlsx
	// 或者，我们可以 hack update registeredTypes manually?
	// Accessing private field? No.

	// Better: redefine struct with correct name in a separate package?
	// Or just accept we need to verify the FIX logic, not necessarily with exact filename match if we can invoke low level load.

	// But LoadAllConfigs relies on filename matching.
}
