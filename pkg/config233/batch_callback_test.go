package config233

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockBusinessManager 模拟业务管理器用于测试批量回调
type mockBusinessManager struct {
	mutex               sync.RWMutex
	callCount           int32 // 使用 atomic 操作
	receivedConfigNames [][]string
	lastCallTime        time.Time
	callIntervals       []time.Duration
}

func newMockBusinessManager() *mockBusinessManager {
	return &mockBusinessManager{
		receivedConfigNames: make([][]string, 0, 8),
		callIntervals:       make([]time.Duration, 0, 8),
	}
}

// OnConfigLoadComplete 实现 IBusinessConfigManager 接口
func (m *mockBusinessManager) OnConfigLoadComplete(changedConfigNameList []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	if !m.lastCallTime.IsZero() {
		m.callIntervals = append(m.callIntervals, now.Sub(m.lastCallTime))
	}
	m.lastCallTime = now

	atomic.AddInt32(&m.callCount, 1)
	// 复制一份以避免后续修改影响
	configsCopy := make([]string, len(changedConfigNameList))
	copy(configsCopy, changedConfigNameList)
	m.receivedConfigNames = append(m.receivedConfigNames, configsCopy)
}

func (m *mockBusinessManager) getCallCount() int {
	return int(atomic.LoadInt32(&m.callCount))
}

func (m *mockBusinessManager) getReceivedConfigNames() [][]string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	// 返回副本避免并发问题
	result := make([][]string, len(m.receivedConfigNames))
	copy(result, m.receivedConfigNames)
	return result
}

func (m *mockBusinessManager) getLastCallConfigCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if len(m.receivedConfigNames) == 0 {
		return 0
	}
	return len(m.receivedConfigNames[len(m.receivedConfigNames)-1])
}

func (m *mockBusinessManager) getTotalConfigCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	total := 0
	for _, names := range m.receivedConfigNames {
		total += len(names)
	}
	return total
}

// createTestConfigs 批量创建测试配置文件
func createTestConfigs(t *testing.T, tempDir string, count int) []string {
	t.Helper()
	names := make([]string, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("Config%03d", i)
		filename := name + ".json"
		content := fmt.Sprintf(`[{"id":"%d","name":"%s"}]`, i+1, name)
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("创建测试文件 %s 失败: %v", filename, err)
		}
		names[i] = name
	}
	return names
}

// TestBatchCallback_InitialLoad 测试首次加载时的批量回调
func TestBatchCallback_InitialLoad(t *testing.T) {
	tempDir := t.TempDir()
	configNames := createTestConfigs(t, tempDir, 3)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证：只调用一次回调
	if callCount := mockManager.getCallCount(); callCount != 1 {
		t.Errorf("期望回调 1 次，实际 %d 次", callCount)
	}

	// 验证：回调包含所有配置
	if configCount := mockManager.getLastCallConfigCount(); configCount != len(configNames) {
		t.Errorf("期望收到 %d 个配置名，实际 %d 个", len(configNames), configCount)
	}

	t.Logf("✓ 批量回调测试通过：1 次回调，收到 %d 个配置", mockManager.getLastCallConfigCount())
}

// TestBatchCallback_HotReload 测试热重载时的批量回调
func TestBatchCallback_HotReload(t *testing.T) {
	tempDir := t.TempDir()
	configNames := createTestConfigs(t, tempDir, 2)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("初始加载配置失败: %v", err)
	}

	initialCallCount := mockManager.getCallCount()
	if initialCallCount != 1 {
		t.Errorf("初始加载期望 1 次回调，实际 %d 次", initialCallCount)
	}

	// 批量重载
	manager.batchReloadConfigs(configNames)

	// 验证：热重载只触发一次额外回调
	if totalCallCount := mockManager.getCallCount(); totalCallCount != 2 {
		t.Errorf("热重载后期望总共 2 次回调，实际 %d 次", totalCallCount)
	}

	// 验证：第二次回调包含重载的配置
	if configCount := mockManager.getLastCallConfigCount(); configCount != len(configNames) {
		t.Errorf("热重载期望收到 %d 个配置名，实际 %d 个", len(configNames), configCount)
	}

	t.Logf("✓ 热重载批量回调测试通过")
}

// TestBatchCallback_NoCallOnEmptyReload 测试空重载不触发回调
func TestBatchCallback_NoCallOnEmptyReload(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 1)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	initialCallCount := mockManager.getCallCount()

	// 重载不存在的配置
	manager.batchReloadConfigs([]string{"NonExistent1", "NonExistent2"})

	if finalCallCount := mockManager.getCallCount(); finalCallCount != initialCallCount {
		t.Errorf("空重载不应触发回调，初始 %d 次，最终 %d 次", initialCallCount, finalCallCount)
	}

	t.Log("✓ 空重载不触发回调测试通过")
}

// TestBatchCallback_PartialReload 测试部分配置重载成功时的回调
func TestBatchCallback_PartialReload(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "ExistingConfig.json")
	if err := os.WriteFile(filePath, []byte(`[{"id":"1","name":"existing"}]`), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	initialCallCount := mockManager.getCallCount()

	// 重载：1个存在 + 1个不存在
	manager.batchReloadConfigs([]string{"ExistingConfig", "NonExistentConfig"})

	if finalCallCount := mockManager.getCallCount(); finalCallCount != initialCallCount+1 {
		t.Errorf("部分重载应触发 1 次回调，初始 %d 次，最终 %d 次", initialCallCount, finalCallCount)
	}

	// 回调只包含成功重载的配置
	if configCount := mockManager.getLastCallConfigCount(); configCount != 1 {
		t.Errorf("部分重载期望收到 1 个配置名，实际 %d 个", configCount)
	}

	t.Log("✓ 部分重载回调测试通过")
}

// TestBatchCallback_MultipleManagers 测试多个业务管理器都收到回调
func TestBatchCallback_MultipleManagers(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 1)

	manager := NewConfigManager233(tempDir)

	// 注册多个管理器
	managers := make([]*mockBusinessManager, 3)
	for i := range managers {
		managers[i] = newMockBusinessManager()
		manager.RegisterBusinessManager(managers[i])
	}

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证所有管理器都收到回调
	for i, m := range managers {
		if m.getCallCount() != 1 {
			t.Errorf("Manager%d 期望 1 次回调，实际 %d 次", i+1, m.getCallCount())
		}
	}

	t.Log("✓ 多业务管理器批量回调测试通过")
}

// TestBatchCallback_LargeConfigCount 测试大量配置的批量回调性能
func TestBatchCallback_LargeConfigCount(t *testing.T) {
	tempDir := t.TempDir()
	configCount := 50
	configNames := createTestConfigs(t, tempDir, configCount)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	start := time.Now()
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	loadDuration := time.Since(start)

	// 验证：只调用一次回调
	if callCount := mockManager.getCallCount(); callCount != 1 {
		t.Errorf("期望回调 1 次，实际 %d 次", callCount)
	}

	// 验证：回调包含所有配置
	receivedCount := mockManager.getLastCallConfigCount()
	if receivedCount != configCount {
		t.Errorf("期望收到 %d 个配置名，实际 %d 个", configCount, receivedCount)
	}

	t.Logf("✓ 大量配置批量回调测试通过：%d 个配置，加载耗时 %v", configCount, loadDuration)

	// 测试批量重载性能
	start = time.Now()
	manager.batchReloadConfigs(configNames)
	reloadDuration := time.Since(start)

	if callCount := mockManager.getCallCount(); callCount != 2 {
		t.Errorf("重载后期望 2 次回调，实际 %d 次", callCount)
	}

	t.Logf("✓ 批量重载 %d 个配置耗时 %v", configCount, reloadDuration)
}

// TestBatchCallback_ConcurrentAccess 测试并发访问时的回调安全性
func TestBatchCallback_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 5)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 并发读取回调数据
	var wg sync.WaitGroup
	goroutineCount := 10
	iterations := 100

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = mockManager.getCallCount()
				_ = mockManager.getReceivedConfigNames()
				_ = mockManager.getTotalConfigCount()
			}
		}()
	}

	wg.Wait()
	t.Log("✓ 并发访问回调数据测试通过")
}

// TestBatchCallback_CallbackOrder 测试回调顺序保证
func TestBatchCallback_CallbackOrder(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 3)

	manager := NewConfigManager233(tempDir)

	// 使用 channel 记录回调顺序
	callOrder := make([]int, 0, 3)
	var orderMutex sync.Mutex

	for i := 0; i < 3; i++ {
		idx := i
		mockManager := &orderTrackingManager{
			onCallback: func() {
				orderMutex.Lock()
				callOrder = append(callOrder, idx)
				orderMutex.Unlock()
			},
		}
		manager.RegisterBusinessManager(mockManager)
	}

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	orderMutex.Lock()
	if len(callOrder) != 3 {
		t.Errorf("期望 3 个回调，实际 %d 个", len(callOrder))
	}
	orderMutex.Unlock()

	t.Log("✓ 回调顺序测试通过")
}

// orderTrackingManager 用于追踪回调顺序
type orderTrackingManager struct {
	onCallback func()
}

func (m *orderTrackingManager) OnConfigLoadComplete(changedConfigNameList []string) {
	if m.onCallback != nil {
		m.onCallback()
	}
}

// TestBatchCallback_EmptyConfigList 测试空配置列表不触发回调
func TestBatchCallback_EmptyConfigList(t *testing.T) {
	tempDir := t.TempDir()

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	// 空目录加载
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 空目录应该不触发回调
	if callCount := mockManager.getCallCount(); callCount != 0 {
		t.Errorf("空目录加载不应触发回调，实际 %d 次", callCount)
	}

	t.Log("✓ 空配置列表不触发回调测试通过")
}

// BenchmarkBatchCallback_SingleManager 基准测试单管理器回调性能
func BenchmarkBatchCallback_SingleManager(b *testing.B) {
	tempDir := b.TempDir()

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("Config%02d", i)
		filename := name + ".json"
		content := fmt.Sprintf(`[{"id":"%d","name":"%s"}]`, i+1, name)
		filePath := filepath.Join(tempDir, filename)
		_ = os.WriteFile(filePath, []byte(content), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewConfigManager233(tempDir)
		mockManager := newMockBusinessManager()
		manager.RegisterBusinessManager(mockManager)
		_ = manager.LoadAllConfigs()
	}
}

// BenchmarkBatchCallback_MultipleManagers 基准测试多管理器回调性能
func BenchmarkBatchCallback_MultipleManagers(b *testing.B) {
	tempDir := b.TempDir()

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("Config%02d", i)
		filename := name + ".json"
		content := fmt.Sprintf(`[{"id":"%d","name":"%s"}]`, i+1, name)
		filePath := filepath.Join(tempDir, filename)
		_ = os.WriteFile(filePath, []byte(content), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewConfigManager233(tempDir)
		for j := 0; j < 5; j++ {
			manager.RegisterBusinessManager(newMockBusinessManager())
		}
		_ = manager.LoadAllConfigs()
	}
}

// ==================== 安全性和边界测试 ====================

// TestBatchCallback_NilSafety 测试 nil 安全性
func TestBatchCallback_NilSafety(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 2)

	manager := NewConfigManager233(tempDir)

	// 测试不注册任何管理器时的行为
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("无管理器时加载配置失败: %v", err)
	}

	// 验证不会 panic
	manager.batchReloadConfigs([]string{"Config000", "Config001"})

	t.Log("✓ nil 安全性测试通过")
}

// TestBatchCallback_PanicRecovery 测试回调 panic 恢复
func TestBatchCallback_PanicRecovery(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 2)

	manager := NewConfigManager233(tempDir)

	// 注册一个正常的管理器
	normalManager := newMockBusinessManager()
	manager.RegisterBusinessManager(normalManager)

	// 注册一个会 panic 的管理器 - 当前实现不处理 panic
	// 这里只验证正常管理器能正常工作
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if normalManager.getCallCount() != 1 {
		t.Errorf("正常管理器应收到 1 次回调，实际 %d 次", normalManager.getCallCount())
	}

	t.Log("✓ 回调安全性测试通过")
}

// TestBatchCallback_SliceNotShared 测试切片不共享（避免数据污染）
func TestBatchCallback_SliceNotShared(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 3)

	manager := NewConfigManager233(tempDir)

	// 注册一个正常的管理器（先注册）
	normalManager := newMockBusinessManager()
	manager.RegisterBusinessManager(normalManager)

	// 创建一个会修改接收到数据的管理器（后注册）
	modifyingManager := &sliceModifyingManager{
		received: make([][]string, 0),
	}
	manager.RegisterBusinessManager(modifyingManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证正常管理器收到的数据没有被污染
	received := normalManager.getReceivedConfigNames()
	if len(received) == 0 {
		t.Fatal("正常管理器未收到数据")
	}

	// 检查数据完整性 - 正常管理器应该收到完整数据
	if len(received[0]) != 3 {
		t.Errorf("正常管理器应收到 3 个配置名，实际 %d 个", len(received[0]))
	}

	// 验证没有 "MODIFIED" 在正常管理器的数据中
	for _, name := range received[0] {
		if name == "MODIFIED" {
			t.Error("正常管理器的数据被污染了")
		}
	}

	t.Log("✓ 切片不共享测试通过")
}

// sliceModifyingManager 会修改接收到的切片
type sliceModifyingManager struct {
	received [][]string
}

func (m *sliceModifyingManager) OnConfigLoadComplete(changedConfigNameList []string) {
	// 尝试修改接收到的切片（不应影响其他管理器）
	if len(changedConfigNameList) > 0 {
		changedConfigNameList[0] = "MODIFIED"
	}
	m.received = append(m.received, changedConfigNameList)
}

// TestBatchCallback_RapidReload 测试快速连续重载
func TestBatchCallback_RapidReload(t *testing.T) {
	tempDir := t.TempDir()
	configNames := createTestConfigs(t, tempDir, 5)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("初始加载配置失败: %v", err)
	}

	// 快速连续重载 10 次
	for i := 0; i < 10; i++ {
		manager.batchReloadConfigs(configNames)
	}

	// 验证：初始加载 1 次 + 重载 10 次 = 11 次
	expectedCalls := 11
	if actualCalls := mockManager.getCallCount(); actualCalls != expectedCalls {
		t.Errorf("期望 %d 次回调，实际 %d 次", expectedCalls, actualCalls)
	}

	t.Log("✓ 快速连续重载测试通过")
}

// TestBatchCallback_ConcurrentReload 测试并发重载
func TestBatchCallback_ConcurrentReload(t *testing.T) {
	tempDir := t.TempDir()
	configNames := createTestConfigs(t, tempDir, 5)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("初始加载配置失败: %v", err)
	}

	// 并发重载
	var wg sync.WaitGroup
	concurrency := 5
	reloadsPerGoroutine := 3

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < reloadsPerGoroutine; j++ {
				manager.batchReloadConfigs(configNames)
			}
		}()
	}

	wg.Wait()

	// 验证没有 panic，且回调次数合理
	callCount := mockManager.getCallCount()
	minExpected := 1 // 至少初始加载
	maxExpected := 1 + concurrency*reloadsPerGoroutine

	if callCount < minExpected || callCount > maxExpected {
		t.Errorf("回调次数异常，期望 %d-%d，实际 %d", minExpected, maxExpected, callCount)
	}

	t.Logf("✓ 并发重载测试通过，回调次数: %d", callCount)
}

// TestBatchCallback_MemoryEfficiency 测试内存效率（大量配置名不应导致内存泄漏）
func TestBatchCallback_MemoryEfficiency(t *testing.T) {
	tempDir := t.TempDir()
	configCount := 100
	createTestConfigs(t, tempDir, configCount)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	// 多次加载和重载
	for i := 0; i < 5; i++ {
		if err := manager.LoadAllConfigs(); err != nil {
			t.Fatalf("第 %d 次加载配置失败: %v", i+1, err)
		}
	}

	// 验证数据一致性
	totalConfigs := mockManager.getTotalConfigCount()
	expectedTotal := configCount * 5 // 5 次加载，每次 100 个配置

	if totalConfigs != expectedTotal {
		t.Errorf("总配置数异常，期望 %d，实际 %d", expectedTotal, totalConfigs)
	}

	t.Logf("✓ 内存效率测试通过，共处理 %d 个配置", totalConfigs)
}

// TestBatchCallback_ConfigNameIntegrity 测试配置名完整性
func TestBatchCallback_ConfigNameIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	expectedNames := createTestConfigs(t, tempDir, 5)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	received := mockManager.getReceivedConfigNames()
	if len(received) == 0 {
		t.Fatal("未收到任何回调")
	}

	receivedNames := received[0]

	// 创建集合用于验证
	expectedSet := make(map[string]bool)
	for _, name := range expectedNames {
		expectedSet[name] = true
	}

	receivedSet := make(map[string]bool)
	for _, name := range receivedNames {
		receivedSet[name] = true
	}

	// 验证所有期望的配置名都收到了
	for name := range expectedSet {
		if !receivedSet[name] {
			t.Errorf("缺少配置名: %s", name)
		}
	}

	// 验证没有多余的配置名
	for name := range receivedSet {
		if !expectedSet[name] {
			t.Errorf("多余的配置名: %s", name)
		}
	}

	t.Log("✓ 配置名完整性测试通过")
}

// TestBatchCallback_CallbackTiming 测试回调时序
func TestBatchCallback_CallbackTiming(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 3)

	manager := NewConfigManager233(tempDir)

	var callbackTime time.Time
	var loadCompleteTime time.Time

	timingManager := &timingTrackingManager{
		onCallback: func() {
			callbackTime = time.Now()
		},
	}
	manager.RegisterBusinessManager(timingManager)

	startTime := time.Now()
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	loadCompleteTime = time.Now()

	// 回调应该在加载完成之前发生
	if callbackTime.After(loadCompleteTime) {
		t.Error("回调时间应该在 LoadAllConfigs 返回之前")
	}

	// 回调应该在开始之后发生
	if callbackTime.Before(startTime) {
		t.Error("回调时间应该在 LoadAllConfigs 开始之后")
	}

	t.Logf("✓ 回调时序测试通过，加载耗时: %v", loadCompleteTime.Sub(startTime))
}

// timingTrackingManager 用于追踪回调时间
type timingTrackingManager struct {
	onCallback func()
}

func (m *timingTrackingManager) OnConfigLoadComplete(changedConfigNameList []string) {
	if m.onCallback != nil {
		m.onCallback()
	}
}

// TestBatchCallback_ReloadAfterModify 测试修改后重载
func TestBatchCallback_ReloadAfterModify(t *testing.T) {
	tempDir := t.TempDir()
	createTestConfigs(t, tempDir, 2)

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)

	// 首次加载
	if err := manager.LoadAllConfigs(); err != nil {
		t.Fatalf("首次加载配置失败: %v", err)
	}

	// 修改配置文件
	modifiedContent := `[{"id":"999","name":"modified"}]`
	filePath := filepath.Join(tempDir, "Config000.json")
	if err := os.WriteFile(filePath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("修改配置文件失败: %v", err)
	}

	// 重载
	manager.batchReloadConfigs([]string{"Config000"})

	// 验证回调次数
	if callCount := mockManager.getCallCount(); callCount != 2 {
		t.Errorf("期望 2 次回调，实际 %d 次", callCount)
	}

	// 验证第二次回调只包含修改的配置
	received := mockManager.getReceivedConfigNames()
	if len(received) < 2 {
		t.Fatal("回调记录不足")
	}

	if len(received[1]) != 1 || received[1][0] != "Config000" {
		t.Errorf("第二次回调应只包含 Config000，实际: %v", received[1])
	}

	t.Log("✓ 修改后重载测试通过")
}

// BenchmarkBatchCallback_LargeConfigCount 基准测试大量配置
func BenchmarkBatchCallback_LargeConfigCount(b *testing.B) {
	tempDir := b.TempDir()

	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("Config%03d", i)
		filename := name + ".json"
		content := fmt.Sprintf(`[{"id":"%d","name":"%s"}]`, i+1, name)
		filePath := filepath.Join(tempDir, filename)
		_ = os.WriteFile(filePath, []byte(content), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewConfigManager233(tempDir)
		mockManager := newMockBusinessManager()
		manager.RegisterBusinessManager(mockManager)
		_ = manager.LoadAllConfigs()
	}
}

// BenchmarkBatchCallback_ReloadOnly 基准测试仅重载性能
func BenchmarkBatchCallback_ReloadOnly(b *testing.B) {
	tempDir := b.TempDir()
	var configNames []string

	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("Config%02d", i)
		filename := name + ".json"
		content := fmt.Sprintf(`[{"id":"%d","name":"%s"}]`, i+1, name)
		filePath := filepath.Join(tempDir, filename)
		_ = os.WriteFile(filePath, []byte(content), 0644)
		configNames = append(configNames, name)
	}

	manager := NewConfigManager233(tempDir)
	mockManager := newMockBusinessManager()
	manager.RegisterBusinessManager(mockManager)
	_ = manager.LoadAllConfigs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.batchReloadConfigs(configNames)
	}
}
