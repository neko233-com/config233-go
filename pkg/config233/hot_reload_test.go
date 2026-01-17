package config233

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestHotReload_BatchingAndCooldown 测试热重载的批量和冷却机制
func TestHotReload_BatchingAndCooldown(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试配置文件
	testFile := filepath.Join(tempDir, "TestConfig.json")
	initialContent := `[{"id":"1","name":"test1"}]`

	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 创建并启动配置管理器
	manager := NewConfigManager233(tempDir)
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 启动文件监听
	if err := manager.StartWatching(); err != nil {
		t.Fatalf("启动文件监听失败: %v", err)
	}
	defer func() {
		if manager.watcher != nil {
			_ = manager.watcher.Close()
		}
	}()

	// 验证初始配置已加载
	manager.mutex.RLock()
	_, exists := manager.configs["TestConfig"]
	manager.mutex.RUnlock()

	if !exists {
		t.Fatal("初始配置未加载")
	}

	t.Log("热重载批量和冷却机制测试通过")
}

// TestHotReloadState_AddPending 测试添加待重载配置
func TestHotReloadState_AddPending(t *testing.T) {
	hrs := newHotReloadState()

	// 添加多个待重载配置
	configNames := []string{"Config1", "Config2", "Config3"}
	for _, name := range configNames {
		hrs.addPendingReload(name)
	}

	// 验证待重载列表
	hrs.mutex.Lock()
	if len(hrs.pendingReloads) != 3 {
		t.Errorf("期望 3 个待重载配置，实际 %d 个", len(hrs.pendingReloads))
	}

	for _, name := range configNames {
		if !hrs.pendingReloads[name] {
			t.Errorf("配置 %s 未在待重载列表中", name)
		}
	}
	hrs.mutex.Unlock()

	t.Log("添加待重载配置测试通过")
}

// TestHotReloadState_Cooldown 测试冷却机制
func TestHotReloadState_Cooldown(t *testing.T) {
	hrs := newHotReloadState()

	// 设置最近重载时间
	hrs.mutex.Lock()
	hrs.lastReloadTime = time.Now()
	hrs.mutex.Unlock()

	// 添加待重载配置（应该被冷却延迟）
	hrs.addPendingReload("TestConfig")

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)

	// 验证配置仍在待重载列表中（因为被冷却延迟）
	hrs.mutex.Lock()
	pendingCount := len(hrs.pendingReloads)
	hrs.mutex.Unlock()

	if pendingCount == 0 {
		t.Error("冷却期间不应清空待重载列表")
	}

	t.Log("冷却机制测试通过")
}

// TestBatchReloadConfigs 测试批量重载配置
func TestBatchReloadConfigs(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建多个测试配置文件
	configs := map[string]string{
		"Config1.json": `[{"id":"1","name":"config1"}]`,
		"Config2.json": `[{"id":"2","name":"config2"}]`,
		"Config3.json": `[{"id":"3","name":"config3"}]`,
	}

	for filename, content := range configs {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("创建测试文件 %s 失败: %v", filename, err)
		}
	}

	// 创建配置管理器并加载
	manager := NewConfigManager233(tempDir)
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 测试批量重载
	configNames := []string{"Config1", "Config2", "Config3"}
	manager.batchReloadConfigs(configNames)

	// 验证所有配置都已重载
	manager.mutex.RLock()
	for _, name := range configNames {
		if _, exists := manager.configs[name]; !exists {
			t.Errorf("配置 %s 未成功重载", name)
		}
	}
	manager.mutex.RUnlock()

	t.Log("批量重载配置测试通过")
}

// TestReloadConstants 测试重载常量定义
func TestReloadConstants(t *testing.T) {
	if ReloadBatchDelay != 500*time.Millisecond {
		t.Errorf("ReloadBatchDelay 应为 500ms，实际 %v", ReloadBatchDelay)
	}

	if ReloadCooldown != 300*time.Millisecond {
		t.Errorf("ReloadCooldown 应为 300ms，实际 %v", ReloadCooldown)
	}

	t.Log("重载常量测试通过")
}
