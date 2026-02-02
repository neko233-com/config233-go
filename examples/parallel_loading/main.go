package main

import (
	"fmt"
	"log"
	"time"

	"github.com/neko233-com/config233-go/pkg/config233"
)

// ItemConfig 示例配置结构
type ItemConfig struct {
	ID       string `json:"id"`
	ItemName string `json:"itemName"`
	Quality  int    `json:"quality"`
}

func main() {
	fmt.Println("=== Config233 并行加载示例 ===")
	fmt.Println()

	// 1. 获取全局单例管理器
	manager := config233.GetInstance()

	// 2. 设置配置目录
	if _, err := manager.SetConfigDir("../../testdata"); err != nil {
		log.Fatal("设置配置目录失败:", err)
	}

	// 3. 注册配置类型（可选，用于类型转换）
	config233.RegisterType[ItemConfig]()

	// 4. 启动管理器（自动使用并行加载）
	fmt.Println("⏱️  开始加载配置...")
	startTime := time.Now()

	if _, err := manager.Start(); err != nil {
		log.Fatal("启动配置管理器失败:", err)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("✅ 配置加载完成，耗时: %v\n\n", elapsed)

	// 5. 查看加载的配置
	configNames := manager.GetLoadedConfigNames()
	fmt.Printf("📦 已加载 %d 个配置文件:\n", len(configNames))
	for i, name := range configNames {
		fmt.Printf("  %d. %s\n", i+1, name)
	}

	// 6. 使用配置数据
	fmt.Println("\n📖 配置使用示例:")

	// 6.1 获取单个配置
	item, exists := config233.GetConfigById[ItemConfig]("1001")
	if exists {
		fmt.Printf("  - 物品 1001: %s (品质: %d)\n", item.ItemName, item.Quality)
	}

	// 6.2 获取配置列表
	items := config233.GetConfigList[ItemConfig]()
	fmt.Printf("  - 总共有 %d 个物品配置\n", len(items))

	// 6.3 获取配置映射
	itemMap := config233.GetConfigMap[ItemConfig]()
	fmt.Printf("  - 配置映射大小: %d\n", len(itemMap))

	fmt.Println("\n🎉 并行加载示例运行完成！")
	fmt.Println("\n💡 性能提示:")
	fmt.Println("  - 配置文件越多，并行加载的性能提升越明显")
	fmt.Println("  - 多核 CPU 环境下可获得 3-7x 的加速")
	fmt.Println("  - 热重载会自动监听文件变化，无需手动重启")
}
