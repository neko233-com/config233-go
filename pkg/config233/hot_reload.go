package config233

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	// ReloadBatchDelay 批量重载延迟时间（收集变更文件）
	ReloadBatchDelay = 500 * time.Millisecond

	// ReloadCooldown 重载冷却时间（避免频繁重载）
	ReloadCooldown = 300 * time.Millisecond
)

// hotReloadState 热重载状态管理
type hotReloadState struct {
	mutex          sync.Mutex
	pendingReloads map[string]bool // 待重载的配置名集合
	timer          *time.Timer     // 批量重载定时器
	lastReloadTime time.Time       // 上次重载时间
	isReloading    bool            // 是否正在重载
}

func newHotReloadState() *hotReloadState {
	return &hotReloadState{
		pendingReloads: make(map[string]bool),
		lastReloadTime: time.Time{},
	}
}

// addPendingReload 添加待重载的配置
func (hrs *hotReloadState) addPendingReload(configName string) {
	hrs.mutex.Lock()
	defer hrs.mutex.Unlock()

	hrs.pendingReloads[configName] = true

	// 如果定时器已存在，停止它
	if hrs.timer != nil {
		hrs.timer.Stop()
	}

	// 创建新的批量重载定时器
	hrs.timer = time.AfterFunc(ReloadBatchDelay, func() {
		hrs.triggerBatchReload()
	})

	getLogger().Info("添加待重载配置", "configName", configName, "pendingCount", len(hrs.pendingReloads))
	fmt.Printf("[config233] 添加待重载配置: configName=%s, pendingCount=%d\n", configName, len(hrs.pendingReloads))
}

// triggerBatchReload 触发批量重载
func (hrs *hotReloadState) triggerBatchReload() {
	hrs.mutex.Lock()

	// 检查冷却时间
	timeSinceLastReload := time.Since(hrs.lastReloadTime)
	if timeSinceLastReload < ReloadCooldown {
		// 还在冷却期，延迟重载
		remainingCooldown := ReloadCooldown - timeSinceLastReload
		getLogger().Info("热重载冷却中，延迟重载", "remainingMs", remainingCooldown.Milliseconds())

		hrs.timer = time.AfterFunc(remainingCooldown, func() {
			hrs.triggerBatchReload()
		})
		hrs.mutex.Unlock()
		return
	}

	if hrs.isReloading {
		// 正在重载，等待完成后再重试
		getLogger().Info("热重载进行中，稍后重试")
		hrs.timer = time.AfterFunc(100*time.Millisecond, func() {
			hrs.triggerBatchReload()
		})
		hrs.mutex.Unlock()
		return
	}

	// 获取待重载列表
	configsToReload := make([]string, 0, len(hrs.pendingReloads))
	for configName := range hrs.pendingReloads {
		configsToReload = append(configsToReload, configName)
	}

	// 清空待重载列表
	hrs.pendingReloads = make(map[string]bool)
	hrs.isReloading = true

	hrs.mutex.Unlock()

	// 执行批量重载
	if len(configsToReload) > 0 {
		getLogger().Info("开始批量热重载", "configCount", len(configsToReload), "configs", configsToReload)
		fmt.Printf("[config233] 开始批量热重载: configCount=%d, configs=%v\n", len(configsToReload), configsToReload)
		startTime := time.Now()

		// 调用实际的重载逻辑
		manager := GetInstance()
		manager.batchReloadConfigs(configsToReload)

		elapsed := time.Since(startTime)
		getLogger().Info("批量热重载完成", "configCount", len(configsToReload), "elapsedMs", elapsed.Milliseconds())
		fmt.Printf("[config233] 批量热重载完成: configCount=%d, elapsedMs=%d\n", len(configsToReload), elapsed.Milliseconds())
	}

	// 更新重载状态
	hrs.mutex.Lock()
	hrs.lastReloadTime = time.Now()
	hrs.isReloading = false
	hrs.mutex.Unlock()
}

// batchReloadConfigs 批量重载指定的配置文件
func (cm *ConfigManager233) batchReloadConfigs(configNames []string) {
	if len(configNames) == 0 {
		return
	}

	// 构建配置名到文件路径的映射
	configFiles := make(map[string]string)

	// 遍历配置目录，查找对应的配置文件
	_ = filepath.Walk(cm.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil || info.IsDir() {
			return nil
		}

		fileName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		for _, configName := range configNames {
			if fileName == configName {
				configFiles[configName] = path
				break
			}
		}

		return nil
	})

	// 串行重载每个配置文件（避免并发冲突）
	successCount := 0
	successConfigs := make([]string, 0, len(configFiles))
	for configName, filePath := range configFiles {
		ext := strings.ToLower(filepath.Ext(filePath))
		var err error

		switch ext {
		case ".xlsx", ".xls":
			err = cm.loadExcelConfig(filePath)
		case ".json":
			err = cm.loadJsonConfig(filePath)
		case ".tsv":
			err = cm.loadTsvConfig(filePath)
		default:
			continue
		}

		if err != nil {
			getLogger().Error(err, "重载配置失败", "configName", configName, "path", filePath)
			fmt.Printf("\033[31m[config233] 重载配置失败: configName=%s, path=%s, error=%v\033[0m\n", configName, filePath, err)
		} else {
			successCount++
			successConfigs = append(successConfigs, configName)
			getLogger().Info("重载配置成功", "configName", configName, "path", filePath)
			fmt.Printf("[config233] 重载配置成功: configName=%s, path=%s\n", configName, filePath)
		}
	}

	// 通知业务管理器（批量，每个管理器收到独立副本）
	if len(successConfigs) > 0 {
		for _, manager := range cm.businessManagers {
			// 为每个管理器创建独立副本，防止数据污染
			configsCopy := make([]string, len(successConfigs))
			copy(configsCopy, successConfigs)
			manager.OnConfigLoadComplete(configsCopy)
		}
		// 更新最后一次加载配置的时间戳
		cm.lastLoadTimeMs.Store(time.Now().UnixMilli())
	}

	getLogger().Info("批量重载完成", "total", len(configNames), "success", successCount, "failed", len(configNames)-successCount)
	fmt.Printf("[config233] 批量重载完成: total=%d, success=%d, failed=%d\n", len(configNames), successCount, len(configNames)-successCount)
}

// StartWatching 启动文件监听（带批量重载和冷却机制）
// 启动对配置目录的文件监听，当配置文件发生变化时自动批量重载配置
// 特性：
// - 批量重载：收集 500ms 内的所有变更，一次性重载
// - 冷却机制：两次重载之间至少间隔 300ms
// - 智能过滤：只监听已加载的配置文件，忽略临时文件
// - 递归监听：自动监听所有子目录
// 返回值:
//
//	error: 启动监听过程中的错误
func (cm *ConfigManager233) StartWatching() error {
	if cm.watcher != nil {
		getLogger().Info("文件监听已启动")
		fmt.Printf("\033[33m[config233] 文件监听已启动\033[0m\n")
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监听器失败: %w", err)
	}

	// 递归添加所有目录到监听器（包括子目录）
	watchedDirs := []string{}
	err = filepath.Walk(cm.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if addErr := watcher.Add(path); addErr != nil {
				getLogger().Error(addErr, "添加监听目录失败", "path", path)
				fmt.Printf("\033[31m[config233] 添加监听目录失败: %s, 错误: %v\033[0m\n", path, addErr)
				return addErr
			}
			watchedDirs = append(watchedDirs, path)
		}
		return nil
	})
	if err != nil {
		_ = watcher.Close()
		return fmt.Errorf("添加监听目录失败: %w", err)
	}

	cm.watcher = watcher

	// 初始化热重载状态
	hotReload := newHotReloadState()

	go func() {
		defer func() {
			_ = watcher.Close()
		}()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// 只处理写和创建事件
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					baseName := filepath.Base(event.Name)

					// 跳过临时文件
					if strings.HasPrefix(baseName, "~$") ||
						strings.Contains(baseName, "~") ||
						strings.Contains(baseName, "#") {
						continue
					}

					ext := strings.ToLower(filepath.Ext(event.Name))
					if ext == ".json" || ext == ".xlsx" || ext == ".xls" || ext == ".tsv" {
						// 检查是否是已加载的配置
						configName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

						cm.mutex.RLock()
						_, exists := cm.configs[configName]
						cm.mutex.RUnlock()

						if exists {
							getLogger().Info("检测到已加载配置变化", "file", event.Name, "configName", configName)
							fmt.Printf("[config233] 检测到已加载配置变化: file=%s, configName=%s\n", event.Name, configName)

							// 添加到待重载队列（触发批量重载）
							hotReload.addPendingReload(configName)
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				getLogger().Error(err, "文件监听错误")
				fmt.Printf("\033[31m[config233] 文件监听错误: %v\033[0m\n", err)
			}
		}
	}()

	getLogger().Info("文件监听已启动（批量重载模式）",
		"dir", cm.configDir,
		"batchDelay", ReloadBatchDelay.Milliseconds(),
		"cooldown", ReloadCooldown.Milliseconds())
	fmt.Printf("[config233] 文件监听已启动（批量重载模式）: dir=%s, batchDelay=%dms, cooldown=%dms, watchedDirs=%d\n",
		cm.configDir, ReloadBatchDelay.Milliseconds(), ReloadCooldown.Milliseconds(), len(watchedDirs))
	return nil
}
