package main

import (
	"fmt"
	"log"

	"github.com/neko233-com/config233-go/internal/config233"
)

// StudentConfig 示例配置结构体
type StudentConfig struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	fmt.Println("ConfigManager233 示例")

	// 使用全局实例
	fmt.Println("=== 使用全局实例 ===")
	// 注册配置类型
	config233.RegisterType[StudentConfig]()

	// 使用泛型方法获取配置
	config, exists := config233.GetConfigById[StudentConfig](1)
	if exists {
		fmt.Printf("找到配置: %+v\n", config)
	} else {
		fmt.Println("配置不存在")
	}

	// 注意：ConfigManager233 主要用于 Excel 配置管理
	// 对于其他格式的配置，建议使用 Config233

	// 创建自定义实例
	fmt.Println("\n=== 创建自定义实例 ===")
	manager := config233.NewConfigManager233("../testdata")

	err := manager.LoadAllConfigs()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	} else {
		configNames := manager.GetLoadedConfigNames()
		fmt.Printf("成功加载配置，配置数量: %d\n", len(configNames))
		for _, name := range configNames {
			fmt.Printf("配置名: %s\n", name)
		}
	}

	// 注册重载回调
	manager.RegisterReloadFunc(func() {
		fmt.Println("配置已重载！")
	})

	fmt.Println("\nConfigManager233 示例完成")
}
