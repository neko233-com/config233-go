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
	globalIdMaps     atomic.Value                      // 缓存 ID -> interface{} (存储 *map[string]map[string]interface{})
	globalSlices     atomic.Value                      // 缓存 slice []interface{} (存储 *map[string][]interface{})
	registeredTypes  map[string]reflect.Type           // 已注册的类型
	registerTypeMu   sync.RWMutex                      // 保护 registeredTypes
	started          atomic.Bool                       // 是否已启动，启动后不允许修改配置目录
}

var (
	instance     *ConfigManager233
	instanceOnce sync.Once
)

// GetInstance 获取全局单例配置管理器实例
// 首次调用时会创建实例，后续调用返回同一个实例
// 返回值:
//
//	*ConfigManager233: 全局单例配置管理器实例
func GetInstance() *ConfigManager233 {
	instanceOnce.Do(func() {
		// 默认配置目录，可以通过环境变量或 SetConfigDir 覆盖
		configDir := os.Getenv("CONFIG233_DIR")
		if configDir == "" {
			configDir = "config"
		}
		instance = &ConfigManager233{
			configs:          make(map[string]interface{}),
			configMaps:       make(map[string]map[string]interface{}),
			configDir:        configDir,
			reloadFuncs:      make([]func(), 0),
			businessManagers: make([]IBusinessConfigManager, 0),
			watcher:          nil,
			registeredTypes:  make(map[string]reflect.Type),
		}

		// 初始化 atomic.Value
		instance.globalIdMaps.Store(&map[string]map[string]interface{}{})
		instance.globalSlices.Store(&map[string][]interface{}{})
	})
	return instance
}

// Instance 全局配置管理器实例（已废弃，请使用 GetInstance()）
// 提供单例模式的全局配置管理器，方便快速访问
var Instance *ConfigManager233

// init 初始化全局配置管理器（向后兼容）
func init() {
	Instance = GetInstance()
}

// SetConfigDir 设置配置目录路径（链式调用）
// 只能在启动前调用，启动后调用会返回错误
// 参数:
//
//	configDir: 配置文件的目录路径
//
// 返回值:
//
//	*ConfigManager233: 返回自身，支持链式调用
//	error: 如果已启动则返回错误
func (cm *ConfigManager233) SetConfigDir(configDir string) (*ConfigManager233, error) {
	if cm.started.Load() {
		return cm, fmt.Errorf("配置管理器已启动，不允许修改配置目录")
	}
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.configDir = configDir
	return cm, nil
}

// Start 启动配置管理器（链式调用）
// 加载所有配置并启动文件监听，启动后不允许修改配置目录
// 返回值:
//
//	*ConfigManager233: 返回自身，支持链式调用
//	error: 启动过程中的错误
func (cm *ConfigManager233) Start() (*ConfigManager233, error) {
	if cm.started.Load() {
		return cm, nil // 已经启动，直接返回
	}

	// 加载所有配置
	if err := cm.LoadAllConfigs(); err != nil {
		return cm, fmt.Errorf("加载配置失败: %w", err)
	}

	// 启动文件监听
	if err := cm.StartWatching(); err != nil {
		return cm, fmt.Errorf("启动文件监听失败: %w", err)
	}

	// 标记为已启动
	cm.started.Store(true)

	return cm, nil
}

// NewConfigManager233 已废弃：请使用 GetInstance().SetConfigDir().Start() 代替
// 为了向后兼容保留此函数，但建议使用新的单例模式
// Deprecated: 使用 GetInstance().SetConfigDir(configDir) 代替
func NewConfigManager233(configDir string) *ConfigManager233 {
	manager := GetInstance()
	manager.SetConfigDir(configDir)
	return manager
}

// RegisterType 注册配置结构体类型，用于将加��的配置数据自动转换为指定类型
// 这个函数应该在加载配置之前调用
func (cm *ConfigManager233) RegisterType(typ reflect.Type) {
	if typ == nil {
		return
	}
	name := typ.Name()

	// 如果是指针类型，则获取元素类型
	if typ.Kind() == reflect.Ptr {
		name = typ.Elem().Name()
		typ = typ.Elem()
	}

	cm.registerTypeMu.Lock()
	defer cm.registerTypeMu.Unlock()

	cm.registeredTypes[name] = typ
	getLogger().Info("注册配置类型 (Object)", "name", name, "type", typ.String())
}

// RegisterType 注册配置结构体类型，用于将加载的配置数据自动转换为指定类型
// 这个函数应该在加载配置之前调用
func RegisterType[T any]() {
	var example T
	GetInstance().RegisterType(reflect.TypeOf(example))
}

// RegisterTypeByReflect 传入 reflect.Type 来注册
func RegisterTypeByReflect(typ reflect.Type) {
	GetInstance().RegisterType(typ)
}

// getRegisteredType 获取已注册的类型
func (cm *ConfigManager233) getRegisteredType(configName string) (reflect.Type, bool) {
	cm.registerTypeMu.RLock()
	defer cm.registerTypeMu.RUnlock()

	typ, exists := cm.registeredTypes[configName]
	return typ, exists
}

// convertMapToRegisteredStruct 将 map 转换为已注册的结构体类型
func (cm *ConfigManager233) convertMapToRegisteredStruct(configName string, data map[string]interface{}) (interface{}, error) {
	typ, exists := cm.getRegisteredType(configName)
	if !exists {
		// 如果类型未注册，返回原始 map
		return data, nil
	}

	// 对于所有情况，使用预处理数据转换
	processedData := preprocessMapData(data)
	if jsonBytes, err := json.Marshal(processedData); err == nil {
		instance := reflect.New(typ).Interface()
		if err := json.Unmarshal(jsonBytes, instance); err == nil {
			// 返回解引用的指针
			return reflect.ValueOf(instance).Elem().Addr().Interface(), nil
		} else {
			// 转换失败，返回原始错误
			return nil, err
		}
	} else {
		return nil, err
	}
}

// convertSliceToRegisteredStructSlice 将 []interface{} 转换为已注册结构体类型的切片
func (cm *ConfigManager233) convertSliceToRegisteredStructSlice(configName string, data []interface{}) ([]interface{}, error) {
	typ, exists := cm.getRegisteredType(configName)
	if !exists {
		// 如果类型未注册，返回原始数据
		return data, nil
	}

	result := make([]interface{}, 0, len(data))
	for _, item := range data {
		if reflect.TypeOf(item) == reflect.PtrTo(typ) {
			// 如果已经是正确类型，直接添加
			result = append(result, item)
		} else if mapItem, ok := item.(map[string]interface{}); ok {
			// 转换 map 为注册的结构体类型
			if converted, err := cm.convertMapToRegisteredStruct(configName, mapItem); err == nil {
				result = append(result, converted)
			} else {
				// 转换失败，跳过此项
				continue
			}
		} else {
			// 其他类型保持不变
			result = append(result, item)
		}
	}
	return result, nil
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
			// 跳过临时文件和特殊文件
			baseName := filepath.Base(path)
			// 跳过 Excel 临时文件 (以 ~$ 开头) 和包含 ~ 或 # 的文件
			if strings.HasPrefix(baseName, "~$") ||
				strings.Contains(baseName, "~") ||
				strings.Contains(baseName, "#") {
				return nil
			}

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
		// 优先使用 id/ID/Id 字段作为配置 ID
		var id string
		if idVal, ok := item["id"]; ok && idVal != "" {
			id = fmt.Sprintf("%v", idVal)
		} else if idVal, ok := item["ID"]; ok && idVal != "" {
			id = fmt.Sprintf("%v", idVal)
		} else if idVal, ok := item["Id"]; ok && idVal != "" {
			id = fmt.Sprintf("%v", idVal)
		} else if idVal, ok := item["itemId"]; ok && idVal != "" {
			// 兼容 itemId 字段
			id = fmt.Sprintf("%v", idVal)
		}

		if id != "" {
			// 如果有注册的类型，转换为具体结构体
			if converted, err := cm.convertMapToRegisteredStruct(fileName, item); err == nil {
				configMap[id] = converted
				getLogger().Info("成功转换配置项", "index", -1, "configName", fileName, "itemId", item["itemId"])
			} else {
				// 转换失败则使用原始 map
				configMap[id] = item
				getLogger().Error(err, "转换配置项失败", "index", -1, "configName", fileName, "data", item)
			}
		}
	}

	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap

	// Convert to []interface{}
	slice := make([]interface{}, len(dto.DataList))
	for i, v := range dto.DataList {
		// 尝试转换为注册的结构体类型
		if converted, err := cm.convertMapToRegisteredStruct(fileName, v); err == nil {
			slice[i] = converted
			getLogger().Info("成功转换配置项", "index", i, "configName", fileName, "itemId", v["itemId"])
		} else {
			// 转换失败则使用原始 map
			slice[i] = v
			getLogger().Error(err, "转换配置项失败", "index", i, "configName", fileName, "data", v)
		}
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
			// 如果有注册的类型，转换为具体结构体
			if converted, err := cm.convertMapToRegisteredStruct(fileName, item); err == nil {
				configMap[id] = converted
			} else {
				// 转换失败则使用原始 map
				configMap[id] = item
			}
		}
	}

	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap

	// Convert to []interface{}
	slice := make([]interface{}, len(dto.DataList))
	for i, v := range dto.DataList {
		// 尝试转换为注册的结构体类型
		if converted, err := cm.convertMapToRegisteredStruct(fileName, v); err == nil {
			slice[i] = converted
			getLogger().Info("成功转换JSON配置项", "index", i, "configName", fileName, "itemId", v["itemId"])
		} else {
			// 转换失败则使用原始 map
			slice[i] = v
			getLogger().Error(err, "转换JSON配置项失败", "index", i, "configName", fileName, "data", v)
		}
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
			// 如果有注册的类型，转换为具体结构体
			if converted, err := cm.convertMapToRegisteredStruct(fileName, item); err == nil {
				configMap[id] = converted
			} else {
				// 转换失败则使用原始 map
				configMap[id] = item
			}
		}
	}

	cm.configs[fileName] = dto.DataList
	cm.configMaps[fileName] = configMap

	// Convert to []interface{}
	slice := make([]interface{}, len(dto.DataList))
	for i, v := range dto.DataList {
		// 尝试转换为注册的结构体类型
		if converted, err := cm.convertMapToRegisteredStruct(fileName, v); err == nil {
			slice[i] = converted
			getLogger().Info("成功转换TSV配置项", "index", i, "configName", fileName, "itemId", v["itemId"])
		} else {
			// 转换失败则使用原始 map
			slice[i] = v
			getLogger().Error(err, "转换TSV配置项失败", "index", i, "configName", fileName, "data", v)
		}
	}
	cm.setConfigCache(fileName, configMap, slice) // Update cache

	return nil
}

// =====================================================
// 泛型查询方法
// =====================================================

// idToString 将多种类型的 ID 统一转换为 string
func (cm *ConfigManager233) idToString(id interface{}) (string, bool) {
	switch v := id.(type) {
	case string:
		return v, true
	case int:
		return strconv.Itoa(v), true
	case int64:
		return strconv.FormatInt(v, 10), true
	default:
		getLogger().Error(nil, "不支持的 ID 类型", "type", fmt.Sprintf("%T", id))
		return "", false
	}
}

// GetConfigById 根据 ID 获取单个配置（O(1) 查找）- 指定管理器
func GetConfigById[T any](id interface{}) (*T, bool) {
	cm := GetInstance()
	configName := typeNameOf[T]()
	return getConfigByIdWithNameForManager[T](cm, configName, id)
}

// getConfigByIdWithNameForManager 根据配置名和 ID 获取单个配置 - 指定管理器（内部使用）
func getConfigByIdWithNameForManager[T any](cm *ConfigManager233, configName string, configId interface{}) (*T, bool) {
	idStr, ok := cm.idToString(configId)
	if !ok {
		return nil, false
	}
	// 优先从缓存获取 (Lock-Free)
	idMapsPtr := cm.globalIdMaps.Load().(*map[string]map[string]interface{})

	idMaps := *idMapsPtr
	if idMap, exists := idMaps[configName]; exists {
		if item, ok := idMap[idStr]; ok {
			// 尝试直接类型断言
			if result, ok := item.(*T); ok {
				return result, true
			}
			// 如果是 map，尝试转换为 T
			if mapItem, ok := item.(map[string]interface{}); ok {
				if result, err := convertMapToStruct[T](mapItem); err == nil {
					return result, true
				}
			}
		}
		return nil, false
	}

	// 缓存未命中，从实例获取 (Need Lock)
	data, ok := cm.getConfig(configName, configId)
	if !ok {
		return nil, false
	}

	// 尝试直接类型断言
	if result, ok := data.(*T); ok {
		return result, true
	}

	// 如果是 map，尝试转换为 T
	if mapData, ok := data.(map[string]interface{}); ok {
		if result, err := convertMapToStruct[T](mapData); err == nil {
			return result, true
		}
	}

	return nil, false
}

// typeNameOf 获取类型名称，指针时取元素名
func typeNameOf[T any]() string {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ == nil {
		return ""
	}
	if typ.Kind() == reflect.Ptr {
		return typ.Elem().Name()
	}
	return typ.Name()
}

// convertMapToStruct 将 map[string]interface{} 转换为指定的 struct 类型
func convertMapToStruct[T any](data map[string]interface{}) (*T, error) {
	// 先尝试直接 JSON 序列化/反序列化
	if jsonBytes, err := json.Marshal(data); err == nil {
		var result T
		if err := json.Unmarshal(jsonBytes, &result); err == nil {
			return &result, nil
		} else {
			// 如果直接转换失败，尝试预处理数据
			processedData := preprocessMapData(data)
			if processedJsonBytes, processedErr := json.Marshal(processedData); processedErr == nil {
				if err := json.Unmarshal(processedJsonBytes, &result); err == nil {
					return &result, nil
				}
			}
			// 返回原始错误
			return nil, err
		}
	} else {
		return nil, err
	}
}

// preprocessMapData 预处理 map 数据，处理类型不匹配问题（如空字符串转数字）
func preprocessMapData(data map[string]interface{}) map[string]interface{} {
	processed := make(map[string]interface{})

	for key, value := range data {
		processed[key] = preprocessValue(value)
	}

	return processed
}

// preprocessValue 预处理单个值，处理类型不匹配问题
func preprocessValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// 如果是空字符串，返回 nil，让 JSON 反序列化使用零值
		if v == "" {
			return nil
		}
		return v
	case nil:
		return v
	default:
		return v
	}
}

// convertSliceToStructSlice 将 []interface{} 转换为 []*T 类型
func convertSliceToStructSlice[T any](data []interface{}) ([]*T, error) {
	if len(data) == 0 {
		return make([]*T, 0), nil
	}

	result := make([]*T, 0, len(data))
	for _, item := range data {
		if typedItem, ok := item.(*T); ok {
			result = append(result, typedItem)
		} else if mapItem, ok := item.(map[string]interface{}); ok {
			if converted, err := convertMapToStruct[T](mapItem); err == nil {
				result = append(result, converted)
			} else {
				// 如果转换失败，记录错误并跳过
				getLogger().Error(err, "转换切片元素失败", "data", mapItem)
				continue
			}
		} else {
			// 如果不是 map 也不是目标类型，跳过
			continue
		}
	}
	return result, nil
}

// GetConfigList 获取某类型的所有配置列表（纯泛型）- 指定管理器
// 返回 []*T，相当于 map.values() 转 slice
func GetConfigList[T any]() []*T {
	cm := GetInstance()
	configName := typeNameOf[T]()

	// Lock-Free
	slicesPtr := cm.globalSlices.Load().(*map[string][]interface{})
	slices := *slicesPtr
	slice, exists := slices[configName]

	if !exists {
		return nil
	}

	result, err := convertSliceToStructSlice[T](slice)
	if err != nil {
		return nil
	}

	return result
}

// GetAllConfigList 获取某类型的所有配置列表（GetConfigList 的别名）
// 返回 []*T，相当于 map.values() 转 slice
func GetAllConfigList[T any]() []*T {
	return GetConfigList[T]()
}

// GetConfigMap 获取某类型的配置映射（纯泛型）
// 返回 map[string]*T，其中 key 是配置的 ID
func GetConfigMap[T any]() map[string]*T {
	cm := GetInstance()
	configName := typeNameOf[T]()

	// Lock-Free
	idMapsPtr := cm.globalIdMaps.Load().(*map[string]map[string]interface{})
	idMaps := *idMapsPtr
	idMap, exists := idMaps[configName]

	if !exists {
		return nil
	}

	result := make(map[string]*T, len(idMap))
	for id, item := range idMap {
		// 尝试直接类型断言
		if typedItem, ok := item.(*T); ok {
			result[id] = typedItem
		} else if mapItem, ok := item.(map[string]interface{}); ok {
			// 如果是 map，尝试转换为 T
			if converted, err := convertMapToStruct[T](mapItem); err == nil {
				result[id] = converted
			}
		}
	}

	return result
}

// GetKvStringList 从 KV 配置中获取字符串列表
// 参数:
//
//	id: 配置项的 ID
//	defaultVal: 如果配置不存在或值为空时的默认值
//
// 返回值:
//
//	[]string: 解析后的字符串列表（按逗号分隔）
func GetKvStringList[T IKvConfig](id string, defaultVal []string) []string {
	cm := GetInstance()
	configName := typeNameOf[T]()

	// 获取配置项
	config, exists := getConfigByIdWithNameForManager[T](cm, configName, id)
	if !exists {
		return defaultVal
	}

	// 使用类型断言将 *T 转换为 IKvConfig 接口
	kvConfig, ok := any(config).(IKvConfig)
	if !ok {
		return defaultVal
	}

	// 调用 GetValue 方法获取值
	value := kvConfig.GetValue()
	if value == "" {
		return defaultVal
	}

	// 按逗号分割并去除空格
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultVal
	}

	return result
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
						// 检查是否是已加载的配置
						configName := strings.TrimSuffix(filepath.Base(event.Name), filepath.Ext(event.Name))

						cm.mutex.RLock()
						_, exists := cm.configs[configName]
						cm.mutex.RUnlock()

						if exists {
							getLogger().Info("检测到已加载配置变化", "file", event.Name)

							// 重新加载配置
							cm.Reload()
						}
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

// getAllConfigs 获取指定配置的所有项（内部方法）
// 获取指定配置名称下的所有配置项，返回ID到配置数据的映射
// 参数:
//
//	configName: 配置名称
//
// 返回值:
//
//	map[string]interface{}: ID到配置数据的映射
//	bool: 配置是否存在
func (cm *ConfigManager233) getAllConfigs(configName string) (map[string]interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	configMap, exists := cm.configMaps[configName]
	return configMap, exists
}

// getConfig 获取指定配置项（内部方法）
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
func (cm *ConfigManager233) getConfig(configName string, id interface{}) (interface{}, bool) {
	idStr, ok := cm.idToString(id)
	if !ok {
		return nil, false
	}
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	configMap, exists := cm.configMaps[configName]
	if !exists {
		return nil, false
	}

	config, exists := configMap[idStr]
	return config, exists
}

// buildGlobalCaches 已移除 - 索引现在在加载时自动构建（通过 setConfigCache）

// setConfigCache 更新单个配置的全局缓存 - 完全无锁
func (cm *ConfigManager233) setConfigCache(configName string, idMap map[string]interface{}, slice []interface{}) {
	// 转换为注册的结构体类型
	convertedIdMap := make(map[string]interface{}, len(idMap))
	for id, value := range idMap {
		if mapValue, ok := value.(map[string]interface{}); ok {
			if converted, err := cm.convertMapToRegisteredStruct(configName, mapValue); err == nil {
				convertedIdMap[id] = converted
			} else {
				// 转换失败则使用原始值
				// 记录转换错误，便于调试
				getLogger().Error(err, "转换配置失败", "configName", configName, "id", id, "data", mapValue)
				convertedIdMap[id] = value
			}
		} else {
			// 如果不是 map，直接使用原始值
			convertedIdMap[id] = value
		}
	}

	convertedSlice, err := cm.convertSliceToRegisteredStructSlice(configName, slice)
	if err != nil {
		getLogger().Error(err, "转换配置切片失败", "configName", configName)
		convertedSlice = slice
	}

	// 1. 无锁更新 ID Maps (CAS 重试)
	for {
		currentIdMapsPtr := cm.globalIdMaps.Load().(*map[string]map[string]interface{})
		currentIdMaps := *currentIdMapsPtr

		// Copy-On-Write
		newIdMaps := make(map[string]map[string]interface{}, len(currentIdMaps)+1)
		for k, v := range currentIdMaps {
			newIdMaps[k] = v
		}
		newIdMaps[configName] = convertedIdMap

		// CAS 尝试更新
		if cm.globalIdMaps.CompareAndSwap(currentIdMapsPtr, &newIdMaps) {
			break
		}
	}

	// 2. 无锁更新 Slices (CAS 重试)
	for {
		currentSlicesPtr := cm.globalSlices.Load().(*map[string][]interface{})
		currentSlices := *currentSlicesPtr

		// Copy-On-Write
		newSlices := make(map[string][]interface{}, len(currentSlices)+1)
		for k, v := range currentSlices {
			newSlices[k] = v
		}
		newSlices[configName] = convertedSlice

		// CAS 尝试更新
		if cm.globalSlices.CompareAndSwap(currentSlicesPtr, &newSlices) {
			break
		}
	}
}

// getConfigMap 获取配置映射（内部方法）
// 获取某个配置的 ID -> 配置 数据映射
// 参数:
//
//	configName: 配置名称
//
// 返回值:
//
//	map[string]interface{}: ID -> 配置 数据映射
//	bool: 配置是否存在
func (cm *ConfigManager233) getConfigMap(configName string) (map[string]interface{}, bool) {
	idMapsPtr := cm.globalIdMaps.Load().(*map[string]map[string]interface{})
	idMaps := *idMapsPtr
	idMap, exists := idMaps[configName]
	return idMap, exists
}
