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
	"time"

	"github.com/fsnotify/fsnotify"
)

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
	isStarted        atomic.Bool                       // 是否已启动，启动后不允许修改配置目录
	isFirstLoadDone  atomic.Bool                       // 首次加载是否完成
	lastLoadTimeMs   atomic.Int64                      // 最后一次加载配置的时间戳（毫秒）

	// 导出配置相关
	loadDoneWriteConfigFileDir string // 导出配置文件的目录
	isOpenWriteTempFile        bool   // 是否开启导出功能
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
	if cm.isStarted.Load() {
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
	if cm.isStarted.Load() {
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
	cm.isStarted.Store(true)

	return cm, nil
}

// NewConfigManager233 已废弃：请使用 GetInstance().SetConfigDir().Start() 代替
// 为了向后兼容保留此函数，但建议使用新的单例模式
// Deprecated: 使用 GetInstance().SetConfigDir(configDir) 代替
func NewConfigManager233(configDir string) *ConfigManager233 {
	manager := GetInstance()
	// 如果未启动，清空之前的配置（用于测试场景）
	if !manager.isStarted.Load() {
		manager.mutex.Lock()
		manager.configs = make(map[string]interface{})
		manager.configMaps = make(map[string]map[string]interface{})
		manager.configDir = configDir
		// 清空缓存
		manager.globalIdMaps.Store(&map[string]map[string]interface{}{})
		manager.globalSlices.Store(&map[string][]interface{}{})
		// 重置首次加载标志（用于测试场景）
		manager.isFirstLoadDone.Store(false)
		// 清空业务管理器列表（用于测试场景）
		manager.businessManagers = nil
		manager.mutex.Unlock()

		manager.registerTypeMu.Lock()
		manager.registeredTypes = make(map[string]reflect.Type)
		manager.registerTypeMu.Unlock()
	} else {
		// 如果已启动，只更新配置目录（会返回错误，但保持向后兼容）
		manager.SetConfigDir(configDir)
	}
	return manager
}

// SetLoadDoneWriteConfigFileDir 设置加载完成后导出配置文件的目录
// 支持相对路径和绝对路径
func (cm *ConfigManager233) SetLoadDoneWriteConfigFileDir(dir string) *ConfigManager233 {
	cm.loadDoneWriteConfigFileDir = dir
	return cm
}

// GetLoadDoneWriteConfigFileDir 获取加载完成后导出配置文件的目录
func (cm *ConfigManager233) GetLoadDoneWriteConfigFileDir() string {
	return cm.loadDoneWriteConfigFileDir
}

// GetLastLoadTimeMs 获取最后一次加载配置的时间戳（毫秒）
// 返回值:
//
//	int64: Unix 时间戳（毫秒），如果从未加载过则返回 0
func (cm *ConfigManager233) GetLastLoadTimeMs() int64 {
	return cm.lastLoadTimeMs.Load()
}

// SetIsOpenWriteTempFileToSeeMemoryConfig 设置是否开启导出内存配置到文件的功能
// 开启后，每次加载或重载配置后，会将内存中的配置导出为 JSON 文件
func (cm *ConfigManager233) SetIsOpenWriteTempFileToSeeMemoryConfig(isOpen bool) *ConfigManager233 {
	cm.isOpenWriteTempFile = isOpen
	return cm
}

// ExportConfigToJSON 将指定配置导出为 JSON 文件
func (cm *ConfigManager233) ExportConfigToJSON(configName string, data interface{}) {
	if !cm.isOpenWriteTempFile || cm.loadDoneWriteConfigFileDir == "" {
		return
	}

	// 确保目录存在
	if err := os.MkdirAll(cm.loadDoneWriteConfigFileDir, 0755); err != nil {
		getLogger().Error(err, "创建导出目录失败", "dir", cm.loadDoneWriteConfigFileDir)
		return
	}

	// 序列化
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		getLogger().Error(err, "序列化配置失败", "configName", configName)
		return
	}

	// 写入文件
	filePath := filepath.Join(cm.loadDoneWriteConfigFileDir, configName+".json")
	if err := os.WriteFile(filePath, jsonBytes, 0644); err != nil {
		getLogger().Error(err, "写入配置文件失败", "path", filePath)
		return
	}

	getLogger().Info("已导出配置到文件", "configName", configName, "path", filePath)
}

// RegisterType 注册配置结构体类型，用于将加载的配置数据自动转换为指定类型
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
// 使用 config233_column tag 来映射 Excel 列名到 struct 字段
// 如果没有 config233_column tag，则使用字段名匹配（不区分大小写）
func (cm *ConfigManager233) convertMapToRegisteredStruct(configName string, data map[string]interface{}) (interface{}, error) {
	typ, exists := cm.getRegisteredType(configName)
	if !exists {
		// 如果类型未注册，返回原始 map
		return data, nil
	}

	// 创建新实例
	instance := reflect.New(typ).Elem()

	// 构建 map key 到 struct 字段名的映射
	// 优先使用 config233_column tag，否则使用字段名匹配
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name

		// 获取 config233_column tag
		columnTag := field.Tag.Get("config233_column")
		jsonTag := field.Tag.Get("json")
		// parse json tag to get name (e.g. `json:"id,omitempty"`)
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			jsonTag = parts[0]
		}

		// 确定要查找的 key
		var keyToFind string
		if columnTag != "" {
			keyToFind = columnTag
		} else if jsonTag != "" {
			keyToFind = jsonTag
		} else {
			// 使用字段名（尝试小写首字母）
			keyToFind = lowerFirst(fieldName)
		}

		// 在 data 中查找对应的值
		var value interface{}
		var found bool

		// 优先精确匹配
		if v, ok := data[keyToFind]; ok {
			value = v
			found = true
		} else if columnTag == "" && jsonTag == "" {
			// 如果没有 config233_column 或 json tag，尝试不区分大小写匹配
			for k, v := range data {
				if strings.EqualFold(k, fieldName) || strings.EqualFold(k, keyToFind) {
					value = v
					found = true
					break
				}
			}
		}

		if !found {
			continue
		}

		// 设置字段值
		fieldValue := instance.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		if err := setFieldValueFromInterface(fieldValue, value, configName, fieldName); err != nil {
			fmt.Printf("\033[31m[config233] 字段类型转换失败 [%s.%s]: %v\033[0m\n", configName, fieldName, err)
		}
	}

	// 获取指针以便调用方法
	instancePtr := instance.Addr().Interface()

	// lifecycle/AfterLoad 生命周期回调
	if lifecycle, ok := instancePtr.(IConfigLifecycle); ok {
		lifecycle.AfterLoad()
	}

	// lifecycle/Check 校验配置
	if validator, ok := instancePtr.(IConfigValidator); ok {
		if err := validator.Check(); err != nil {
			fmt.Printf("\033[31m[config233] 配置校验失败 [%s]: %v\033[0m\n", configName, err)
			getLogger().Error(err, "配置校验失败", "configName", configName, "data", data)
			// 注意：校验失败仍然返回实例，只是输出错误信息
		}
	}

	return instancePtr, nil
}

// setFieldValueFromInterface 从 interface{} 设置字段值，自动类型转换
func setFieldValueFromInterface(field reflect.Value, value interface{}, _, _ string) error {
	if value == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprintf("%v", value))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := toInt64(value)
		if err != nil {
			return fmt.Errorf("无法将 '%v' 转换为 int: %w", value, err)
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := toUint64(value)
		if err != nil {
			return fmt.Errorf("无法将 '%v' 转换为 uint: %w", value, err)
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := toFloat64(value)
		if err != nil {
			return fmt.Errorf("无法将 '%v' 转换为 float: %w", value, err)
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := toBool(value)
		if err != nil {
			return fmt.Errorf("无法将 '%v' 转换为 bool: %w", value, err)
		}
		field.SetBool(boolVal)

	default:
		return fmt.Errorf("不支持的字段类型: %v", field.Kind())
	}
	return nil
}

// toInt64 将 interface{} 转换为 int64
func toInt64(v interface{}) (int64, error) {
	switch t := v.(type) {
	case int:
		return int64(t), nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	case uint:
		return int64(t), nil
	case uint8:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		return int64(t), nil
	case float32:
		return int64(t), nil
	case float64:
		return int64(t), nil
	case string:
		if t == "" {
			return 0, nil
		}
		return strconv.ParseInt(t, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

// toUint64 将 interface{} 转换为 uint64
func toUint64(v interface{}) (uint64, error) {
	switch t := v.(type) {
	case int:
		return uint64(t), nil
	case int8:
		return uint64(t), nil
	case int16:
		return uint64(t), nil
	case int32:
		return uint64(t), nil
	case int64:
		return uint64(t), nil
	case uint:
		return uint64(t), nil
	case uint8:
		return uint64(t), nil
	case uint16:
		return uint64(t), nil
	case uint32:
		return uint64(t), nil
	case uint64:
		return t, nil
	case float32:
		return uint64(t), nil
	case float64:
		return uint64(t), nil
	case string:
		if t == "" {
			return 0, nil
		}
		return strconv.ParseUint(t, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

// toFloat64 将 interface{} 转换为 float64
func toFloat64(v interface{}) (float64, error) {
	switch t := v.(type) {
	case int:
		return float64(t), nil
	case int8:
		return float64(t), nil
	case int16:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint:
		return float64(t), nil
	case uint8:
		return float64(t), nil
	case uint16:
		return float64(t), nil
	case uint32:
		return float64(t), nil
	case uint64:
		return float64(t), nil
	case float32:
		return float64(t), nil
	case float64:
		return t, nil
	case string:
		if t == "" {
			return 0, nil
		}
		return strconv.ParseFloat(t, 64)
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

// toBool 将 interface{} 转换为 bool
func toBool(v interface{}) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(t).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(t).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(t).Float() != 0, nil
	case string:
		if t == "" || t == "0" || strings.EqualFold(t, "false") {
			return false, nil
		}
		if t == "1" || strings.EqualFold(t, "true") {
			return true, nil
		}
		return false, fmt.Errorf("cannot parse '%s' as bool", t)
	default:
		return false, fmt.Errorf("unsupported type %T", v)
	}
}

// lowerFirst 将字符串首字母转为小写
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
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

// LoadAllConfigs 从目录加载所有配置（并行加载以提升性能）
// 遍历配置目录，自动识别并加载所有支持格式的配置文件
// 支持的格式包括: Excel (.xlsx, .xls), JSON (.json), TSV (.tsv)
//
// 性能优化：使用并行加载大幅提升首次启动速度
// - 文件扫描阶段：快速收集所有需要加载的配置文件
// - 并行加载阶段：每个配置文件在独立 goroutine 中加载，充分利用多核 CPU
// - 线程安全保证：使用细粒度锁保护共享数据结构，缓存使用无锁 CAS 更新
//
// 加载过程中出现的错误会被记录但不会中断整个加载过程
// 返回值:
//
//	error: 加载过程中的错误，如果遍历目录失败则返回错误
func (cm *ConfigManager233) LoadAllConfigs() error {
	// 首先收集所有需要加载的配置文件（不持有锁）
	type configFile struct {
		path string
		ext  string
	}

	var filesToLoad []configFile
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
			case ".xlsx", ".xls", ".json", ".tsv":
				filesToLoad = append(filesToLoad, configFile{path: path, ext: ext})
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 并行加载所有配置文件
	var wg sync.WaitGroup
	loadErrors := make(chan error, len(filesToLoad))

	for _, file := range filesToLoad {
		wg.Add(1)
		go func(f configFile) {
			defer wg.Done()

			var loadErr error
			switch f.ext {
			case ".xlsx", ".xls":
				loadErr = cm.loadExcelConfigThreadSafe(f.path)
				if loadErr != nil {
					getLogger().Error(loadErr, "加载Excel配置失败", "path", f.path)
				}
			case ".json":
				loadErr = cm.loadJsonConfigThreadSafe(f.path)
				if loadErr != nil {
					getLogger().Error(loadErr, "加载JSON配置失败", "path", f.path)
				}
			case ".tsv":
				loadErr = cm.loadTsvConfigThreadSafe(f.path)
				if loadErr != nil {
					getLogger().Error(loadErr, "加载TSV配置失败", "path", f.path)
				}
			}

			if loadErr != nil {
				select {
				case loadErrors <- loadErr:
				default:
				}
			}
		}(file)
	}

	// 等待所有加载完成
	wg.Wait()
	close(loadErrors)

	// 加载完成后调用业务配置管理器的回调（批量）
	cm.mutex.RLock()
	configNames := make([]string, 0, len(cm.configs))
	for configName := range cm.configs {
		configNames = append(configNames, configName)
	}
	cm.mutex.RUnlock()

	// 批量通知所有业务管理器（每个管理器收到独立的切片副本，防止数据污染）
	if len(configNames) > 0 {
		for _, manager := range cm.businessManagers {
			// 为每个管理器创建独立副本，防止某个管理器修改影响其他管理器
			configNamesCopy := make([]string, len(configNames))
			copy(configNamesCopy, configNames)
			manager.OnConfigLoadComplete(configNamesCopy)
		}
	}

	// 首次加载完成后，调用 OnFirstAllConfigDone 回调
	// 使用 CAS 确保只调用一次
	if cm.isFirstLoadDone.CompareAndSwap(false, true) {
		for _, manager := range cm.businessManagers {
			manager.OnFirstAllConfigDone()
		}
		getLogger().Info("首次配置加载完成，已通知所有业务管理器")
	}

	// 更新最后一次加载配置的时间戳
	cm.lastLoadTimeMs.Store(time.Now().UnixMilli())

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

	// 获取 T 的类型信息，检查是否是接口类型
	var zero T
	tType := reflect.TypeOf(zero)
	isInterface := tType != nil && tType.Kind() == reflect.Interface

	// 优先从缓存获取 (Lock-Free)
	idMapsPtr := cm.globalIdMaps.Load().(*map[string]map[string]interface{})
	idMaps := *idMapsPtr

	// 如果 T 是接口类型，需要遍历所有配置查找实现了接口的配置
	if isInterface {
		for _, idMap := range idMaps {
			if item, ok := idMap[idStr]; ok {
				// 尝试转换为 *T
				if result := convertToType[T](item); result != nil {
					return result, true
				}
			}
		}
		return nil, false
	}

	// T 是具体类型，直接通过配置名查找
	if idMap, exists := idMaps[configName]; exists {
		if item, ok := idMap[idStr]; ok {
			// 尝试转换为 *T
			if result := convertToType[T](item); result != nil {
				return result, true
			}
		}
		return nil, false
	}

	// 缓存未命中，从实例获取 (Need Lock)
	// 如果 T 是接口类型，需要遍历所有配置
	if isInterface {
		cm.mutex.RLock()
		defer cm.mutex.RUnlock()

		for _, configMap := range cm.configMaps {
			if data, ok := configMap[idStr]; ok {
				if result := convertToType[T](data); result != nil {
					return result, true
				}
			}
		}
		return nil, false
	}

	// T 是具体类型，直接通过配置名查找
	data, ok := cm.getConfig(configName, configId)
	if !ok {
		return nil, false
	}

	// 尝试转换为 *T
	if result := convertToType[T](data); result != nil {
		return result, true
	}

	return nil, false
}

// convertToType 将 interface{} 转换为 *T，支持接口类型和具体类型
func convertToType[T any](data interface{}) *T {
	// 如果 data 是 nil，直接返回
	if data == nil {
		return nil
	}

	// 首先尝试直接类型断言（适用于具体类型）
	if result, ok := data.(*T); ok {
		return result
	}

	// 获取 T 的类型信息
	var zero T
	tType := reflect.TypeOf(zero)

	// 检查 T 是否是接口类型
	if tType != nil && tType.Kind() == reflect.Interface {
		// T 是接口类型（如 IKvConfig）
		// 当 T 受接口约束时（如 T IKvConfig），实际调用时 T 是具体类型（如 FishingKvConfig）
		// 但 reflect.TypeOf(zero) 会返回接口类型，而不是具体类型
		// 我们需要检查 data 是否实现了接口，如果实现了，就返回它

		// 如果 data 是指针类型
		if val := reflect.ValueOf(data); val.Kind() == reflect.Ptr && !val.IsNil() {
			elemType := val.Elem().Type()

			// 检查元素类型是否实现了接口
			if elemType.Implements(tType) {
				// 元素实现了接口，data 是指向实现了接口的具体类型的指针
				// 由于 T 受接口约束，实际调用时 T 是具体类型（如 FishingKvConfig）
				// 我们需要将 *FishingKvConfig 转换为 *FishingKvConfig（通过 T）
				// 但由于类型系统限制，我们需要使用反射
				// 实际上，当 T 受接口约束时，T 在编译时是具体类型，所以类型断言应该能成功
				// 但如果失败了，说明类型不匹配，我们需要通过反射来检查
				// 尝试通过 any 类型断言
				if result, ok := any(data).(*T); ok {
					return result
				}
				// 如果类型断言失败，可能是因为类型系统无法识别
				// 我们尝试使用反射来构造
				// 但由于 *T 在 Go 中不合法（如果 T 是接口），我们需要特殊处理
				// 实际上，当 T 受接口约束时，T 应该是具体类型，所以这里应该不会到达
				// 但如果到达了，说明类型系统有问题
			}
		}

		return nil
	}

	// T 是具体类型，如果是 map，尝试转换为 T
	if mapData, ok := data.(map[string]interface{}); ok {
		if result, err := convertMapToStruct[T](mapData); err == nil {
			return result
		}
	}

	return nil
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
// 使用 config233_column tag 来映射字段，如果没有则使用字段名匹配
func convertMapToStruct[T any](data map[string]interface{}) (*T, error) {
	var result T
	typ := reflect.TypeOf(result)
	val := reflect.ValueOf(&result).Elem()

	// 构建 map key 到 struct 字段的映射并设置值
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name
		columnTag := field.Tag.Get("config233_column")
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			jsonTag = parts[0]
		}

		// 确定要查找的 key
		var keyToFind string
		if columnTag != "" {
			keyToFind = columnTag
		} else if jsonTag != "" {
			keyToFind = jsonTag
		} else {
			keyToFind = lowerFirst(fieldName)
		}

		// 在 data 中查找对应的值
		var value interface{}
		var found bool

		if v, ok := data[keyToFind]; ok {
			value = v
			found = true
		} else if columnTag == "" && jsonTag == "" {
			// 如果没有 config233_column 或 json tag，尝试不区分大小写匹配
			for k, v := range data {
				if strings.EqualFold(k, fieldName) || strings.EqualFold(k, keyToFind) {
					value = v
					found = true
					break
				}
			}
		}

		if !found {
			continue
		}

		fieldValue := val.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		if err := setFieldValueFromInterface(fieldValue, value, "", ""); err != nil {
			fmt.Printf("\033[31m[config233] 字段类型转换失败 [%s.%s]: %v\033[0m\n", typ.Name(), fieldName, err)
		}
	}

	// 校验
	if validator, ok := any(&result).(IConfigValidator); ok {
		if err := validator.Check(); err != nil {
			fmt.Printf("\033[31m[config233] 配置校验失败 [%s]: %v\033[0m\n", typ.Name(), err)
		}
	}

	// 生命周期
	if lifecycle, ok := any(&result).(IConfigLifecycle); ok {
		lifecycle.AfterLoad()
	}

	return &result, nil
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

// getKvConfigInternal 获取 KV 配置对应的泛型配置对象
// 对于未找到的配置，会打错误日志
func getKvConfigInternal[T any](id string) (*T, bool) {
	cm := GetInstance()
	configName := typeNameOf[T]()

	// 获取配置项
	config, exists := getConfigByIdWithNameForManager[T](cm, configName, id)
	if !exists {
		// KV 配置未找到时打错误日志，方便排查
		getLogger().Error(nil, "KV 配置未找到", "configName", configName, "id", id)
		return nil, false
	}
	return config, true
}

// GetKvToString 从 KV 配置中获取字符串值
// 参数:
//
//	id: 配置项的 ID
//	defaultVal: 如果配置不存在或值为空时的默认值
//
// 返回值:
//
//	string: 配置的字符串值
//
// 注意: 这里的 T 不再受 IKvConfig 约束，实际约束由运行时的类型断言保证（*T 或 T 需要实现 IKvConfig）
func GetKvToString[T any](id string, defaultVal string) string {
	// 获取配置项
	config, exists := getKvConfigInternal[T](id)
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

	return value
}

// GetKvToInt 从 KV 配置中获取整数值
// 参数:
//
//	id: 配置项的 ID
//	defaultVal: 如果配置不存在或值无效时的默认值
//
// 返回值:
//
//	int: 配置的整数值
//
// 注意: 这里的 T 不再受 IKvConfig 约束，实际约束由运行时的类型断言保证（*T 或 T 需要实现 IKvConfig）
func GetKvToInt[T any](id string, defaultVal int) int {
	// 获取配置项
	config, exists := getKvConfigInternal[T](id)
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

	// 尝试转换为整数
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	return defaultVal
}

// GetKvToBoolean 从 KV 配置中获取布尔值
// 参数:
//
//	id: 配置项的 ID
//	defaultVal: 如果配置不存在或值无效时的默认值
//
// 返回值:
//
//	bool: 配置的布尔值
//
// 注意: 这里的 T 不再受 IKvConfig 约束，实际约束由运行时的类型断言保证（*T 或 T 需要实现 IKvConfig）
func GetKvToBoolean[T any](id string, defaultVal bool) bool {
	// 获取配置项
	config, exists := getKvConfigInternal[T](id)
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

	// 尝试转换为布尔值（支持 true/false, 1/0, yes/no, on/off）
	lowerValue := strings.ToLower(strings.TrimSpace(value))
	switch lowerValue {
	case "true", "1", "yes", "on", "enabled":
		return true
	case "false", "0", "no", "off", "disabled":
		return false
	default:
		// 如果无法解析，尝试使用 strconv.ParseBool
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
		return defaultVal
	}
}

// GetKvToCsvStringList 从 KV 配置中获取 CSV 字符串列表（按逗号分隔）
// 参数:
//
//	id: 配置项的 ID
//	defaultVal: 如果配置不存在或值为空时的默认值
//
// 返回值:
//
//	[]string: 解析后的字符串列表（按逗号分隔）
//
// 注意: 这里的 T 不再受 IKvConfig 约束，实际约束由运行时的类型断言保证（*T 或 T 需要实现 IKvConfig）
func GetKvToCsvStringList[T any](id string, defaultVal []string) []string {
	// 获取配置项
	config, exists := getKvConfigInternal[T](id)
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

	// 重新加载所有配置（内部会调用 OnConfigLoadComplete 批量回调）
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
