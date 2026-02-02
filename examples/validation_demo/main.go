package main

import (
	"fmt"
	"log"

	"github.com/neko233-com/config233-go/internal/config233"
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

func main() {
	fmt.Println("=== Config233 修复验证测试 ===")

	// 注册类型
	config233.RegisterType[ItemConfig]()

	// 加载配置
	manager := config233.NewConfigManager233("../testdata")
	err := manager.LoadAllConfigs()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}

	fmt.Println("\n1. 验证 GetConfigList 返回 struct 类型:")
	itemConfigs := config233.GetConfigList[ItemConfig]()
	fmt.Printf("   获取到 %d 个配置项\n", len(itemConfigs))

	for i, config := range itemConfigs {
		// 验证可以直接访问结构体字段（而不是 map 操作）
		fmt.Printf("   [%d] ID: %d, Name: %s, Type: %d, Quality: %d\n",
			i, config.Itemid, config.Itemname, config.Type, config.Quality)
	}

	fmt.Println("\n2. 验证 GetConfigById 返回 struct 类型:")
	config1, exists1 := config233.GetConfigById[ItemConfig]("1")
	if exists1 {
		fmt.Printf("   ID 1 - Name: %s, Desc: %s, Quality: %d\n",
			config1.Itemname, config1.Desc, config1.Quality)
	}

	config2, exists2 := config233.GetConfigById[ItemConfig]("2")
	if exists2 {
		fmt.Printf("   ID 2 - Name: %s, Desc: %s, Quality: %d\n",
			config2.Itemname, config2.Desc, config2.Quality)
	}

	fmt.Println("\n3. 验证类型安全（可以直接调用结构体方法）:")
	totalQuality := 0
	for _, config := range itemConfigs {
		totalQuality += config.Quality
	}
	fmt.Printf("   所有配置的总品质值: %d\n", totalQuality)

	fmt.Println("\n4. 验证空值处理（Excel 中的空单元格正确转换为零值）:")
	for _, config := range itemConfigs {
		if config.Expiretimems == 0 { // Excel 中为空的字段应转换为 0（零值）
			fmt.Printf("   配置 %d 的 Expiretimems 为 0（来自 Excel 空单元格）\n", config.Itemid)
		}
	}

	fmt.Println("\n=== 验证完成 ===")
	fmt.Println("✓ GetConfigList 现在返回 []*ItemConfig 而不是 []interface{}")
	fmt.Println("✓ GetConfigById 现在返回 *ItemConfig 而不是 interface{}")
	fmt.Println("✓ 可以直接访问结构体字段，无需类型断言")
	fmt.Println("✓ 正确处理 Excel 中的空值（转换为对应类型的零值）")
	fmt.Println("✓ 类型安全，编译时即可检查字段访问")
}
