package config233

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/neko233-com/config233-go/pkg/config233/dto"
	"github.com/neko233-com/config233-go/pkg/config233/excel"
	jsonhandler "github.com/neko233-com/config233-go/pkg/config233/json"
	"github.com/neko233-com/config233-go/pkg/config233/tsv"
)

// IKvConfig KvConfig接口，避免反射实现
type IKvConfig interface {
	GetValue() string
}

// IBusinessConfigManager 业务配置管理器接口
type IBusinessConfigManager interface {
	// OnConfigLoadComplete 配置加载完成回调
	OnConfigLoadComplete(configName string)

	// OnConfigHotUpdate 配置热更新回调
	OnConfigHotUpdate()
}

// ConfigManager233 全新的配置管理器，支持热重载
// 提供简化的配置管理接口，支持多种配置格式的自动加载和热重载
// 内部使用 Config233 进行文件监听和配置处理
type ConfigManager233 struct {
	mutex            sync.RWMutex                      // 读写锁，保证线程安全
	configs          map[string]interface{}            // 配置名 -> 配置数据映射
	configMaps       map[string]map[string]interface{} // 配置名 -> (ID -> 配置数据) 映射
	configDir        string                            // 配置目录路径
	reloadFuncs      []func()                          // 配置重载时的回调函数列表
	businessManagers []IBusinessConfigManager          // 业务配置管理器列表
	watcher          *fsnotify.Watcher                 // 文件监听器
}

// Instance 全局配置管理器实例
// 提供单例模式的全局配置管理器，方便快速访问
var Instance *ConfigManager233

// init 初始化全局配置管理器
// 在包初始化时创建全局配置管理器实例
// 配置目录优先从环境变量 CONFIG233_DIR 获取，默认为 "config"
func init() {
	// 默认配置目录，可以通过环境变量或参数覆盖
	configDir := os.Getenv("CONFIG233_DIR")
	if configDir == "" {
		configDir = "config"
	}
	Instance = NewConfigManager233(configDir)
	// 为全局实例自动加载配置
	if err := Instance.LoadAllConfigs(); err != nil {
		getLogger().Error(err, "加载配置失败")
	}
}

// NewConfigManager233 创建新的配置管理器
// 初始化配置管理器实例，设置配置目录
// 参数:
//
//	configDir: 配置文件的目录路径
//
// 返回值:
//
//	*ConfigManager233: 新创建的配置管理器实例
func NewConfigManager233(configDir string) *ConfigManager233 {
	manager := &ConfigManager233{
		configs:          make(map[string]interface{}),
		configMaps:       make(map[string]map[string]interface{}),
		configDir:        configDir,
		reloadFuncs:      make([]func(), 0),
		businessManagers: make([]IBusinessConfigManager, 0),
		watcher:          nil,
	}

	// 不自动加载配置，让用户手动调用 LoadAllConfigs

	return manager
}

// LoadAllConfigs 从目录加载所有配置
// 遍历配置目录，自动识别并加载所有支持格式的配置文件
// 支持的格式包括: Excel (.xlsx, .xls), JSON (.json), TSV (.tsv)
// 加载过程中出现的错误会被记录但不会中断整个加载过程
// 返回值:
//
//	error: 加载过程中的错误，如果遍历目录失败则返回错误
func (cm *ConfigManager233) LoadAllConfigs() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 遍历配置目录
	err := filepath.Walk(cm.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 处理不同类型的配置文件
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".xlsx", ".xls":
				if err := cm.loadExcelConfig(path); err != nil {
					getLogger().Error(err, "加载Excel配置失败", "path", path)
					return nil // 继续处理其他文件
				}
			case ".json":
				if err := cm.loadJsonConfig(path); err != nil {
					getLogger().Error(err, "加载JSON配置失败", "path", path)
					return nil
				}
			case ".tsv":
				if err := cm.loadTsvConfig(path); err != nil {
					getLogger().Error(err, "加载TSV配置失败", "path", path)
					return nil
				}
			}
		}

		return nil
	})

	if err == nil {
		// 加载完成后调用业务配置管理器的回调
		for configName := range cm.configs {
			for _, manager := range cm.businessManagers {
				manager.OnConfigLoadComplete(configName)
			}
		}
		// cm.buildGlobalCaches() // Removed: Incremental cache update handled in loadXxxConfig
	}

	return err
}

// loadExcelConfig 从Excel文件加载配置
// 使用 Excel 处理器读取并解析 Excel 配置文件
// 参数:
//
//	filePath: Excel 配置文件的路径
//
// 返回值:
//
//	error: 加载过程中的错误
func (cm *ConfigManager233) loadExcelConfig(filePath string) error {
	// 创建 Excel 处理器
	handler := &excel.ExcelConfigHandler{}

	// 获取文件名（不含扩展名）作为配置名
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// 读取前端数据格式
	dto := handler.ReadToFrontEndDataList(fileName, filePath).(*dto.FrontEndConfigDto)
	if dto.DataList == nil {
		return nil // 空文件，跳过
	}

	// 转换为配置映射
	configMap := make(map[string]interface{})
	for _, item := range dto.DataList {
		// 使用第一列作为 ID（如果存在的话）
		var id string
		for _, v := range item {
			if id == "" {
				if str, ok := v.(string); ok {
					id = str
				} else {
					id = fmt.Sprintf("%v", v)
				}
			}
			break
		}
		if id != "" {
			configMap[id] = item
		}
	}

	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap

	// Convert to []interface{}
	slice := make([]interface{}, len(dto.DataList))
	for i, v := range dto.DataList {
		slice[i] = v
	}
	cm.setConfigCache(fileName, configMap, slice) // Update cache

	return nil
}

// loadJsonConfig 从JSON文件加载配置
// 使用 JSON 处理器读取并解析 JSON 配置文件
// 参数:
//
//	filePath: JSON 配置文件的路径
//
// 返回值:
//
//	error: 加载过程中的错误
func (cm *ConfigManager233) loadJsonConfig(filePath string) error {
	// 创建 JSON 处理器
	handler := &jsonhandler.JsonConfigHandler{}

	// 获取文件名（不含扩展名）作为配置名
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// 读取前端数据格式
	dto := handler.ReadToFrontEndDataList(fileName, filePath).(*dto.FrontEndConfigDto)
	if dto.DataList == nil {
		return nil // 空文件，跳过
	}

	// 转换为配置映射
	configMap := make(map[string]interface{})
	for _, item := range dto.DataList {
		// 尝试从 map 中提取 ID（支持 "id", "ID", "Id" 等字段）
		var id string
		if idVal, ok := item["id"]; ok {
			if str, ok := idVal.(string); ok {
				id = str
			} else {
				id = fmt.Sprintf("%v", idVal)
			}
		} else if idVal, ok := item["ID"]; ok {
			if str, ok := idVal.(string); ok {
				id = str
			} else {
				id = fmt.Sprintf("%v", idVal)
			}
		} else if idVal, ok := item["Id"]; ok {
			if str, ok := idVal.(string); ok {
				id = str
			} else {
				id = fmt.Sprintf("%v", idVal)
			}
		}

		if id != "" {
			configMap[id] = item
		}
	}

	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap

	// Convert to []interface{}
	slice := make([]interface{}, len(dto.DataList))
	for i, v := range dto.DataList {
		slice[i] = v
	}
	cm.setConfigCache(fileName, configMap, slice) // Update cache

	return nil
}

// loadTsvConfig 从TSV文件加载配置
// 使用 TSV 处理器读取并解析 TSV 配置文件
// 参数:
//
//	filePath: TSV 配置文件的路径
//
// 返回值:
//
//	error: 加载过程中的错误
func (cm *ConfigManager233) loadTsvConfig(filePath string) error {
	// 创建 TSV 处理器
	handler := &tsv.TsvConfigHandler{}

	// 获取文件名（不含扩展名）作为配置名
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// 读取前端数据格式
	dto := handler.ReadToFrontEndDataList(fileName, filePath).(*dto.FrontEndConfigDto)
	if dto.DataList == nil {
		return nil // 空文件，跳过
	}

	// 转换为配置映射
	configMap := make(map[string]interface{})
	for _, item := range dto.DataList {
		// 使用第一列作为 ID（如果存在的话）
		var id string
		for _, v := range item {
			if id == "" {
				if str, ok := v.(string); ok {
					id = str
				} else {
					id = fmt.Sprintf("%v", v)
				}
			}
			break
		}
		if id != "" {
			configMap[id] = item
		}
	}

	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap

	// Convert to []interface{}
	slice := make([]interface{}, len(dto.DataList))
	for i, v := range dto.DataList {
		slice[i] = v
	}
	cm.setConfigCache(fileName, configMap, slice) // Update cache

	return nil
}

// =====================================================
// ID 索引缓存（O(1) 查找）- 完全无锁化设计
// =====================================================

var (
	// globalIdMaps atomic.Value stores *map[string]map[string]interface{}
	// 使用指针类型以支持 CompareAndSwap（map 不可比较）
	// 读操作：Load() 完全无锁
	// 写操作：CAS 重试机制实现无锁更新
	globalIdMaps atomic.Value

	// globalSlices atomic.Value stores *map[string][]interface{}
	// 使用指针类型以支持 CompareAndSwap（map 不可比较）
	// 读操作：Load() 完全无锁
	// 写操作：CAS 重试机制实现无锁更新
	globalSlices atomic.Value
)

func init() {
	// 初始化 atomic.Value，避免 Load 返回 nil
	// 存储指针类型
	globalIdMaps.Store(&map[string]map[string]interface{}{})
	globalSlices.Store(&map[string][]interface{}{})
}

// BuildIdIndex 构建指定配置的 ID 索引（加载配置后调用）
// 采用 Copy-On-Write + CAS 重试策略实现无锁更新
func BuildIdIndex(configName string) {
	allConfigs, ok := Instance.GetAllConfigs(configName)
	if !ok {
		return
	}

	idMap := make(map[string]interface{})
	for id, cfg := range allConfigs {
		idMap[id] = cfg
	}

	// 无锁 CAS 重试更新
	for {
		currentMapPtr := globalIdMaps.Load().(*map[string]map[string]interface{})
		currentMap := *currentMapPtr

		// Copy-On-Write: 创建新的 map
		newMap := make(map[string]map[string]interface{}, len(currentMap)+1)
		for k, v := range currentMap {
			newMap[k] = v
		}

		// 更新目标配置
		newMap[configName] = idMap

		// CAS 尝试更新，如果失败则重试
		if globalIdMaps.CompareAndSwap(currentMapPtr, &newMap) {
			break
		}
		// CAS 失败，有其他 goroutine 更新了，重试
	}

	getLogger().Info("构建 ID 索引", "config", configName, "count", len(idMap))
}

// BuildAllIdIndexes 构建所有已加载配置的 ID 索引
func BuildAllIdIndexes() {
	// 直接调用 buildGlobalCaches 更高效
	Instance.buildGlobalCaches()
	getLogger().Info("所有配置 ID 索引构建完成")
}

// =====================================================
// 泛型查询方法
// =====================================================

// GetConfigById 根据 ID 获取单个配置（O(1) 查找）
// 自动根据类型名推断 configName
// ID 支持 string | int | int64 类型
func GetConfigById[T any](id interface{}) (*T, bool) {
	var zero T
	configName := reflect.TypeOf(zero).Name()

	// 统一转成 string
	var idStr string
	switch v := id.(type) {
	case string:
		idStr = v
	case int:
		idStr = strconv.Itoa(v)
	case int64:
		idStr = strconv.FormatInt(v, 10)
	default:
		getLogger().Error(nil, "GetConfigById 不支持的 ID 类型", "type", fmt.Sprintf("%T", id))
		return nil, false
	}

	return GetConfigByIdWithName[T](configName, idStr)
}

// GetConfigByIdWithName 根据配置名和 ID 获取单个配置
func GetConfigByIdWithName[T any](configName string, id string) (*T, bool) {
	// 优先从缓存获取 (Lock-Free)
	idMapsPtr := globalIdMaps.Load().(*map[string]map[string]interface{})
	idMaps := *idMapsPtr
	if idMap, exists := idMaps[configName]; exists {
		if item, ok := idMap[id]; ok {
			// 尝试直接类型断言
			if result, ok := item.(*T); ok {
				return result, true
			}
			// 如果是 map，尝试转换为 T
			if mapItem, ok := item.(map[string]interface{}); ok {
				var result T
				// 使用 json 进行转换（简单但有效）
				if jsonBytes, err := json.Marshal(mapItem); err == nil {
					if err := json.Unmarshal(jsonBytes, &result); err == nil {
						return &result, true
					}
				}
			}
		}
		return nil, false
	}

	// 缓存未命中，从 Instance 获取 (Need Lock)
	data, ok := Instance.GetConfig(configName, id)
	if !ok {
		return nil, false
	}

	// 尝试直接类型断言
	if result, ok := data.(*T); ok {
		return result, true
	}

	// 如果是 map，尝试转换为 T
	if mapData, ok := data.(map[string]interface{}); ok {
		var result T
		if jsonBytes, err := json.Marshal(mapData); err == nil {
			if err := json.Unmarshal(jsonBytes, &result); err == nil {
				return &result, true
			}
		}
	}

	return nil, false
}

// GetAllConfigList 获取某类型的所有配置列表（纯泛型）
// 返回 []*T，相当于 map.values() 转 slice
func GetAllConfigList[T any]() []*T {
	var zero T
	typ := reflect.TypeOf(zero)
	configName := typ.Name()
	if typ.Kind() == reflect.Ptr {
		configName = typ.Elem().Name()
	}

	// Lock-Free
	slicesPtr := globalSlices.Load().(*map[string][]interface{})
	slices := *slicesPtr
	slice, exists := slices[configName]

	if !exists {
		return nil
	}

	result := make([]*T, 0, len(slice))
	for _, cfg := range slice {
		if item, ok := cfg.(*T); ok {
			result = append(result, item)
		}
	}
	return result
}

// =====================================================
// KvConfig 快捷方法（一行搞定）
// =====================================================

// GetKvInt 从 KvConfig 类型配置中获取 int 值
func GetKvInt[T IKvConfig](id string, defaultVal int) int {
	cfg, _ := GetConfigById[T](id)
	if cfg == nil {
		return defaultVal
	}
	result, err := strconv.Atoi((*cfg).GetValue())
	if err != nil {
		getLogger().Error(err, "GetKvInt 解析失败", "id", id, "value", (*cfg).GetValue())
		return defaultVal
	}
	return result
}

// GetKvString 从 KvConfig 类型配置中获取 string 值
func GetKvString[T IKvConfig](id string, defaultVal string) string {
	cfg, _ := GetConfigById[T](id)
	if cfg == nil {
		return defaultVal
	}
	return (*cfg).GetValue()
}

// GetKvFloat 从 KvConfig 类型配置中获取 float64 值
func GetKvFloat[T IKvConfig](id string, defaultVal float64) float64 {
	cfg, _ := GetConfigById[T](id)
	if cfg == nil {
		return defaultVal
	}
	result, err := strconv.ParseFloat((*cfg).GetValue(), 64)
	if err != nil {
		getLogger().Error(err, "GetKvFloat 解析失败", "id", id, "value", (*cfg).GetValue())
		return defaultVal
	}
	return result
}

// GetKvBool 从 KvConfig 类型配置中获取 bool 值
func GetKvBool[T IKvConfig](id string, defaultVal bool) bool {
	cfg, _ := GetConfigById[T](id)
	if cfg == nil {
		return defaultVal
	}
	s := strings.ToLower((*cfg).GetValue())
	return s == "true" || s == "1" || s == "yes"
}

// GetKvIntList 从 KvConfig 类型配置中获取 []int 值（逗号分隔）
func GetKvIntList[T IKvConfig](id string, defaultVal []int) []int {
	cfg, _ := GetConfigById[T](id)
	if cfg == nil {
		return defaultVal
	}
	str := (*cfg).GetValue()
	if str == "" {
		return defaultVal
	}
	parts := strings.Split(str, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if v, err := strconv.Atoi(p); err == nil {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

// GetKvStringList 从 KvConfig 类型配置中获取 []string 值（逗号分隔）
func GetKvStringList[T IKvConfig](id string, defaultVal []string) []string {
	cfg, _ := GetConfigById[T](id)
	if cfg == nil {
		return defaultVal
	}
	str := (*cfg).GetValue()
	if str == "" {
		return defaultVal
	}
	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" { // 可选：是否过滤空字符串？通常 Kv 配置不应该有空的部分，或者空部分被忽略
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

// =====================================================
// 调试方法
// =====================================================

// PrintLoadedConfigs 打印已加载的配置
func PrintLoadedConfigs() {
	names := Instance.GetLoadedConfigNames()
	fmt.Printf("已加载配置列表 (%d):\n", len(names))
	for _, name := range names {
		count := Instance.GetConfigCount(name)
		fmt.Printf("  - %s: %d 条\n", name, count)
	}
}

// Reload 热重载所有配置
// 重新从配置目录加载所有配置文件，并执行所有注册的重载回调函数
// 通常在检测到配置文件变化时自动调用，也可手动触发
// 返回值:
//
//	error: 重载过程中的错误
func (cm *ConfigManager233) Reload() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 重新加载所有配置
	if err := cm.LoadAllConfigs(); err != nil {
		return err
	}

	// 调用所有重载回调
	for _, fn := range cm.reloadFuncs {
		fn()
	}

	// 调用业务配置管理器的热更新回调
	for _, manager := range cm.businessManagers {
		manager.OnConfigHotUpdate()
	}

	getLogger().Info("配置重载成功")
	return nil
}

// RegisterReloadFunc 注册重载回调函数
// 注册一个在配置重载时被调用的回调函数
// 当配置发生热重载时，所有注册的回调函数都会被执行
// 参数:
//
//	fn: 重载时要执行的回调函数
func (cm *ConfigManager233) RegisterReloadFunc(fn func()) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.reloadFuncs = append(cm.reloadFuncs, fn)
}

// RegisterBusinessManager 注册业务配置管理器
// 注册一个业务配置管理器，用于接收配置加载和热更新的回调
// 参数:
//
//	manager: 业务配置管理器实例
func (cm *ConfigManager233) RegisterBusinessManager(manager IBusinessConfigManager) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.businessManagers = append(cm.businessManagers, manager)
	getLogger().Info("注册业务配置管理器", "type", fmt.Sprintf("%T", manager))
}

// GetLoadedConfigNames 获取已加载的配置名列表
// 返回所有已成功加载的配置名称列表
// 返回值:
//
//	[]string: 已加载配置的名称数组
func (cm *ConfigManager233) GetLoadedConfigNames() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	names := make([]string, 0, len(cm.configMaps))
	for name := range cm.configMaps {
		names = append(names, name)
	}
	return names
}

// GetConfigCount 获取配置数量
// 获取指定配置名称下的配置项数量
// 参数:
//
//	configName: 配置名称
//
// 返回值:
//
//	int: 配置项数量
func (cm *ConfigManager233) GetConfigCount(configName string) int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if configMap, exists := cm.configMaps[configName]; exists {
		return len(configMap)
	}
	return 0
}

// StartWatching 启动文件监听
// 启动对配置目录的文件监听，当配置文件发生变化时自动重载配置
// 返回值:
//
//	error: 启动监听过程中的错误
func (cm *ConfigManager233) StartWatching() error {
	if cm.watcher != nil {
		getLogger().Info("文件监听已启动")
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监听器失败: %w", err)
	}

	err = watcher.Add(cm.configDir)
	if err != nil {
		watcher.Close()
		return fmt.Errorf("添加监听目录失败: %w", err)
	}

	cm.watcher = watcher

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// 只处理写和创建事件
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					ext := strings.ToLower(filepath.Ext(event.Name))
					if ext == ".json" || ext == ".xlsx" || ext == ".xls" || ext == ".tsv" {
						getLogger().Info("检测到配置文件变化", "file", event.Name)
						cm.Reload()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				getLogger().Error(err, "文件监听错误")
			}
		}
	}()

	getLogger().Info("文件监听已启动", "dir", cm.configDir)
	return nil
}

// ConfigManagerReloadListener 配置管理器重载监听器
type ConfigManagerReloadListener struct {
	manager *ConfigManager233
}

// OnConfigDataChange 配置数据变更时调用
func (l *ConfigManagerReloadListener) OnConfigDataChange(typ reflect.Type, dataList []interface{}) {
	getLogger().Info("检测到配置变更", "type", typ.String(), "dataCount", len(dataList))
	l.manager.Reload()
}

// GetAllConfigs 获取指定配置的所有项
// 获取指定配置名称下的所有配置项，返回ID到配置数据的映射
// 参数:
//
//	configName: 配置名称
//
// 返回值:
//
//	map[string]interface{}: ID到配置数据的映射
//	bool: 配置是否存在
func (cm *ConfigManager233) GetAllConfigs(configName string) (map[string]interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	configMap, exists := cm.configMaps[configName]
	return configMap, exists
}

// GetConfig 获取指定配置项
// 根据配置名称和ID获取单个配置项
// 参数:
//
//	configName: 配置名称
//	id: 配置项的唯一标识符
//
// 返回值:
//
//	interface{}: 配置项数据
//	bool: 是否找到该配置项
func (cm *ConfigManager233) GetConfig(configName, id string) (interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	configMap, exists := cm.configMaps[configName]
	if !exists {
		return nil, false
	}

	config, exists := configMap[id]
	return config, exists
}

// buildGlobalCaches 构建全局缓存（ID 索引和切片映射）- 完全无锁
func (cm *ConfigManager233) buildGlobalCaches() {
	// 重建 ID 索引
	newIdMaps := make(map[string]map[string]interface{})
	for configName, configMap := range cm.configMaps {
		newIdMaps[configName] = configMap
	}

	// 重建切片映射
	newSlices := make(map[string][]interface{})
	for configName, rawConfig := range cm.configs {
		// 尝试转为 []interface{}
		// 1. 直接是 []interface{}
		if slice, ok := rawConfig.([]interface{}); ok {
			newSlices[configName] = slice
			continue
		}

		// 2. 是 []map[string]interface{} (前端模式)
		if sliceOfMap, ok := rawConfig.([]map[string]interface{}); ok {
			slice := make([]interface{}, len(sliceOfMap))
			for i, v := range sliceOfMap {
				slice[i] = v
			}
			newSlices[configName] = slice
			continue
		}

		// 3. 反射处理其他切片类型
		v := reflect.ValueOf(rawConfig)
		if v.Kind() == reflect.Slice {
			len := v.Len()
			slice := make([]interface{}, len)
			for i := 0; i < len; i++ {
				slice[i] = v.Index(i).Interface()
			}
			newSlices[configName] = slice
		}
	}

	// 直接 Store 覆盖（全量构建时不需要 CAS）
	// 使用指针类型
	globalIdMaps.Store(&newIdMaps)
	globalSlices.Store(&newSlices)

	getLogger().Info("全局缓存构建完成(Lock-Free)")
}

// setConfigCache 更新单个配置的全局缓存 - 完全无锁
func (cm *ConfigManager233) setConfigCache(configName string, idMap map[string]interface{}, slice []interface{}) {
	// 1. 无锁更新 ID Maps (CAS 重试)
	for {
		currentIdMapsPtr := globalIdMaps.Load().(*map[string]map[string]interface{})
		currentIdMaps := *currentIdMapsPtr

		// Copy-On-Write
		newIdMaps := make(map[string]map[string]interface{}, len(currentIdMaps)+1)
		for k, v := range currentIdMaps {
			newIdMaps[k] = v
		}
		newIdMaps[configName] = idMap

		// CAS 尝试更新
		if globalIdMaps.CompareAndSwap(currentIdMapsPtr, &newIdMaps) {
			break
		}
	}

	// 2. 无锁更新 Slices (CAS 重试)
	for {
		currentSlicesPtr := globalSlices.Load().(*map[string][]interface{})
		currentSlices := *currentSlicesPtr

		// Copy-On-Write
		newSlices := make(map[string][]interface{}, len(currentSlices)+1)
		for k, v := range currentSlices {
			newSlices[k] = v
		}
		newSlices[configName] = slice

		// CAS 尝试更新
		if globalSlices.CompareAndSwap(currentSlicesPtr, &newSlices) {
			break
		}
	}
}
