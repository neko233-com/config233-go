package config233

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"config233-go/pkg/config233/dto"
	"config233-go/pkg/config233/excel"
	"config233-go/pkg/config233/json"
	"config233-go/pkg/config233/tsv"
)

// ConfigManager233 全新的配置管理器，支持热重载
// 提供简化的配置管理接口，支持多种配置格式的自动加载和热重载
// 内部使用 Config233 进行文件监听和配置处理
type ConfigManager233 struct {
	mutex       sync.RWMutex                      // 读写锁，保证线程安全
	configs     map[string]interface{}            // 配置名 -> 配置数据映射
	configMaps  map[string]map[string]interface{} // 配置名 -> (ID -> 配置数据) 映射
	configDir   string                            // 配置目录路径
	reloadFuncs []func()                          // 配置重载时的回调函数列表
	watcher     *Config233                        // 内部使用的 Config233 实例，用于文件监听
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
}

// NewConfigManager233 创建新的配置管理器
// 初始化配置管理器实例，设置配置目录并自动加载所有配置
// 参数:
//
//	configDir: 配置文件的目录路径
//
// 返回值:
//
//	*ConfigManager233: 新创建的配置管理器实例
func NewConfigManager233(configDir string) *ConfigManager233 {
	manager := &ConfigManager233{
		configs:     make(map[string]interface{}),
		configMaps:  make(map[string]map[string]interface{}),
		configDir:   configDir,
		reloadFuncs: make([]func(), 0),
		watcher:     NewConfig233(),
	}

	// 初始化配置
	if err := manager.LoadAllConfigs(); err != nil {
		getLogger().Errorf("加载配置失败: %v", err)
	}

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
					getLogger().Errorf("加载Excel配置失败 %s: %v", path, err)
					return nil // 继续处理其他文件
				}
			case ".json":
				if err := cm.loadJsonConfig(path); err != nil {
					getLogger().Errorf("加载JSON配置失败 %s: %v", path, err)
					return nil
				}
			case ".tsv":
				if err := cm.loadTsvConfig(path); err != nil {
					getLogger().Errorf("加载TSV配置失败 %s: %v", path, err)
					return nil
				}
			}
		}

		return nil
	})

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
				id = v
			}
			break
		}
		if id != "" {
			configMap[id] = item
		}
	}

	cm.mutex.Lock()
	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap
	cm.mutex.Unlock()

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
	handler := &json.JsonConfigHandler{}

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
				id = v
			}
			break
		}
		if id != "" {
			configMap[id] = item
		}
	}

	cm.mutex.Lock()
	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap
	cm.mutex.Unlock()

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
				id = v
			}
			break
		}
		if id != "" {
			configMap[id] = item
		}
	}

	cm.mutex.Lock()
	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap
	cm.mutex.Unlock()

	return nil
}

// GetConfig 获取指定配置项
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

// GetConfigAsStruct 将配置转换为指定类型的struct
// 获取配置项并将其转换为指定的结构体类型
// 参数:
//
//	configName: 配置名称
//	id: 配置项ID
//	target: 目标结构体指针，用于接收转换后的数据
//
// 返回值:
//
//	error: 转换过程中的错误
func (cm *ConfigManager233) GetConfigAsStruct(configName, id string, target interface{}) error {
	config, exists := cm.GetConfig(configName, id)
	if !exists {
		return fmt.Errorf("配置项 %s/%s 不存在", configName, id)
	}

	configMap, ok := config.(map[string]interface{})
	if !ok {
		return fmt.Errorf("配置项格式不正确")
	}

	return cm.mapToStruct(configMap, target)
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

// mapToStruct 将map转换为struct
func (cm *ConfigManager233) mapToStruct(data map[string]interface{}, target interface{}) error {
	// 使用简单的类型转换
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("target 必须是指针")
	}

	targetValue = targetValue.Elem()
	targetType := targetValue.Type()

	for i := 0; i < targetValue.NumField(); i++ {
		field := targetValue.Field(i)
		fieldType := targetType.Field(i)
		fieldName := fieldType.Name

		if value, exists := data[fieldName]; exists && field.CanSet() {
			// 简单类型转换
			if value != nil {
				val := reflect.ValueOf(value)
				if val.Type().AssignableTo(field.Type()) {
					field.Set(val)
				}
			}
		}
	}

	return nil
}

// StartWatching 启动文件监听
// 启动对配置目录的文件监听，当配置文件发生变化时自动重载配置
// 注意: 当前版本暂未实现此功能，避免循环导入问题
// 返回值:
//
//	error: 启动监听过程中的错误
func (cm *ConfigManager233) StartWatching() error {
	// 暂时不启动监听，避免循环导入
	// TODO: 实现文件监听功能
	getLogger().Info("ConfigManager233 文件监听暂未实现")
	return nil
}

// ConfigManagerReloadListener 配置管理器重载监听器
type ConfigManagerReloadListener struct {
	manager *ConfigManager233
}

// OnConfigDataChange 配置数据变更时调用
func (l *ConfigManagerReloadListener) OnConfigDataChange(typ reflect.Type, dataList []interface{}) {
	getLogger().Infof("检测到配置变更，类型: %s, 数据项数: %d", typ.String(), len(dataList))
	l.manager.Reload()
}
