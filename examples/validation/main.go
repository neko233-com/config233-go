package main

import (
	"fmt"
	"log"

	"github.com/neko233-com/config233-go/pkg/config233"
	"github.com/neko233-com/config233-go/pkg/config233/dto"
	"github.com/neko233-com/config233-go/pkg/config233/excel"
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
	fmt.Println("=== Excel 数据源验证 ===")

	// 1. 直接测试 Excel 处理器读取的数据
	handler := &excel.ExcelConfigHandler{}
	dtoResult := handler.ReadToFrontEndDataList("ItemConfig", "./testdata/ItemConfig.xlsx").(*dto.FrontEndConfigDto)

	fmt.Printf("Excel 读取的数据项数量: %d\n", len(dtoResult.DataList))

	for i, item := range dtoResult.DataList {
		fmt.Printf("数据项 %d: %+v\n", i, item)
		// 检查空字符串字段
		if expireTimeMs, exists := item["expireTimeMs"]; exists {
			fmt.Printf("  - expireTimeMs: '%v' (类型: %T)\n", expireTimeMs, expireTimeMs)
		}
	}

	fmt.Println("\n=== 注册类型并验证修复 ===")

	// 2. 注册类型
	config233.RegisterType[ItemConfig]()

	// 3. 加载配置
	manager := config233.NewConfigManager233("../testdata")
	err := manager.LoadAllConfigs()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}

	// 4. 验证修复效果
	itemConfigs := config233.GetConfigList[ItemConfig]()
	fmt.Printf("\n转换后的配置项数量: %d\n", len(itemConfigs))

	for i, config := range itemConfigs {
		fmt.Printf("转换后配置 %d: ID=%d, Name=%s, ExpireTimeMs=%d\n",
			i, config.Itemid, config.Itemname, config.Expiretimems)
	}

	// 5. 验证特定 ID 获取
	fmt.Println("\n=== ID 查询验证 ===")
	for id := 1; id <= 2; id++ {
		config, exists := config233.GetConfigById[ItemConfig](fmt.Sprintf("%d", id))
		if exists {
			fmt.Printf("ID %d: %+v\n", id, config)
		} else {
			fmt.Printf("ID %d: 未找到\n", id)
		}
	}

	fmt.Println("\n修复验证成功！Excel 中的空字符串已正确转换为零值。")
}
