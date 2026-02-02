package tests

import (
	"testing"

	"github.com/neko233-com/config233-go/internal/config233"
)

// BenchmarkParallelLoading 测试并行加载性能
func BenchmarkParallelLoading(b *testing.B) {
	testDir := getTestDataDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := config233.NewConfigManager233(testDir)
		if err := manager.LoadAllConfigs(); err != nil {
			b.Fatalf("加载配置失败: %v", err)
		}
	}
}

// TestParallelLoadingCorrectness 测试并行加载的正确性
func TestParallelLoadingCorrectness(t *testing.T) {
	testDir := getTestDataDir()

	// 运行多次以确保并行加载的稳定性
	for i := 0; i < 10; i++ {
		manager := config233.NewConfigManager233(testDir)
		if err := manager.LoadAllConfigs(); err != nil {
			t.Fatalf("第 %d 次加载配置失败: %v", i+1, err)
		}

		names := manager.GetLoadedConfigNames()
		if len(names) == 0 {
			t.Errorf("第 %d 次加载：没有加载到任何配置", i+1)
		}

		// 验证关键配置存在
		expectedConfigs := []string{"ItemConfig", "StudentExcel", "FishingKvConfig"}
		loadedMap := make(map[string]bool)
		for _, name := range names {
			loadedMap[name] = true
		}

		for _, expected := range expectedConfigs {
			if !loadedMap[expected] {
				t.Errorf("第 %d 次加载：缺少配置 %s", i+1, expected)
			}
		}
	}

	t.Log("并行加载正确性测试通过：10 次加载全部成功")
}
