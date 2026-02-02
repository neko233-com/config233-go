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
	// 1. 注册配置类型
	config233.RegisterType[ItemConfig]()

	// 2. 创建配置管理器并加载配置
	manager := config233.NewConfigManager233("../testdata")
	err := manager.LoadAllConfigs()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}

	// 3. 测试获取所有配置列表 - 这是修复后的核心功能
	itemConfigs := config233.GetConfigList[ItemConfig]()
	fmt.Printf("获取到 %d 个 ItemConfig 项\n", len(itemConfigs))

	for i, config := range itemConfigs {
		fmt.Printf("配置 %d: %+v\n", i, config)
	}

	// 4. 测试按 ID 获取配置
	itemConfig1, exists1 := config233.GetConfigById[ItemConfig]("1")
	if exists1 {
		fmt.Printf("ID 为 1 的配置: %+v\n", itemConfig1)
	} else {
		fmt.Println("未找到 ID 为 1 的配置")
	}

	itemConfig2, exists2 := config233.GetConfigById[ItemConfig]("2")
	if exists2 {
		fmt.Printf("ID 为 2 的配置: %+v\n", itemConfig2)
	} else {
		fmt.Println("未找到 ID 为 2 的配置")
	}

	// 5. 验证类型安全 - 现在可以直接访问字段
	for _, config := range itemConfigs {
		fmt.Printf("物品ID: %d, 名称: %s, 品质: %d\n",
			config.Itemid, config.Itemname, config.Quality)
	}

	fmt.Println("\n修复成功！GetConfigList 现在返回实际的 struct 类型，而不是 map[string]interface{}")
}
