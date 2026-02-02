package tests

import (
	"sync"
	"testing"

	"github.com/neko233-com/config233-go/internal/config233"
)

// TestLockFreeConcurrentWrites 测试无锁并发写入
func TestLockFreeConcurrentWrites(t *testing.T) {
	// 创建一个新的配置管理器
	manager := config233.NewConfigManager233("../testdata")

	// 模拟并发加载配置
	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 并发调用 LoadAllConfigs
			err := manager.LoadAllConfigs()
			if err != nil {
				t.Errorf("goroutine %d: 加载配置失败: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
	t.Log("并发加载配置完成")
}

// TestLockFreeReadWrite 测试无锁读写并发
func TestLockFreeReadWrite(t *testing.T) {
	// 创建新的实例用于测试
	testManager := config233.NewConfigManager233("../testdata")
	err := testManager.LoadAllConfigs()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	var wg sync.WaitGroup
	readerCount := 50
	writerCount := 5

	// 启动多个读 goroutine
	for i := 0; i < readerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 持续读取配置
			for j := 0; j < 1000; j++ {
				names := testManager.GetLoadedConfigNames()
				_ = names
			}
		}(i)
	}

	// 启动多个写 goroutine
	for i := 0; i < writerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 持续重新加载配置
			for j := 0; j < 10; j++ {
				err := testManager.LoadAllConfigs()
				if err != nil {
					t.Errorf("writer %d: 重载配置失败: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()
	t.Log("并发读写测试完成")
}

// BenchmarkLockFreeRead 基准测试：无锁读性能
func BenchmarkLockFreeRead(b *testing.B) {
	// 加载配置
	testManager := config233.NewConfigManager233("../testdata")
	testManager.LoadAllConfigs()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			names := testManager.GetLoadedConfigNames()
			_ = names
		}
	})
}
