package config233

import "reflect"

// =============================================================================
// 配置管理器核心接口定义
// =============================================================================

// IKvConfig 键值配置接口，用于避免反射实现
// 实现此接口的结构体可以通过 GetKvToXxx 系列函数进行类型化访问
//
// 使用示例:
//
//	type GameConfig struct {
//	    Id    string `json:"id"`
//	    Value string `json:"value"`
//	}
//	func (c *GameConfig) GetValue() string { return c.Value }
//
//	// 使用时：
//	maxLevel := config233.GetKvToInt[GameConfig]("max_level", 100)
type IKvConfig interface {
	// GetValue 获取配置值的字符串表示
	// 返回配置项的值，用于后续的类型转换
	// 返回值:
	//   string: 配置值的字符串形式，将被转换为目标类型
	GetValue() string
}

// IBusinessConfigManager 业务配置管理器接口
// 实现此接口的结构体可以接收配置加载和变更的通知
// 用于在配置变更时执行业务逻辑，如缓存刷新、服务重启等
//
// 典型实现模式:
//
//	type MyBusinessManager struct {
//	    itemCache    map[string]*ItemConfig
//	    playerCache  map[string]*PlayerConfig
//	}
//
//	func (m *MyBusinessManager) OnConfigLoadComplete(changedConfigs []string) {
//	    for _, name := range changedConfigs {
//	        switch name {
//	        case "ItemConfig":
//	            m.refreshItemCache()
//	        case "PlayerConfig":
//	            m.refreshPlayerCache()
//	        }
//	    }
//	}
//
//	func (m *MyBusinessManager) OnFirstAllConfigDone() {
//	    m.initCrossConfigRelations()  // 初始化配置间关联
//	    m.startBusinessServices()     // 启动业务服务
//	}
type IBusinessConfigManager interface {
	// OnConfigLoadComplete 配置加载完成回调（批量）
	//
	// 调用时机:
	//   - 首次加载所有配置完成后
	//   - 热重载检测到配置文件变更后
	//   - 手动调用 LoadAllConfigs() 或 Reload() 后
	//
	// 参数:
	//   changedConfigNameList: 本次发生变更的配置名称列表
	//     - 首次加载时包含所有成功加载的配置名
	//     - 热重载时只包含实际发生变更的配置名
	//     - 每个管理器收到的是独立的切片副本，避免并发修改问题
	//
	// 性能说明:
	//   - 使用批量回调避免频繁调用（N个配置变更只调用1次而非N次）
	//   - 可以精确知道哪些配置发生了变更，避免无关的缓存刷新
	//   - 支持多个业务管理器同时注册，按注册顺序依次调用
	OnConfigLoadComplete(changedConfigNameList []string)

	// OnFirstAllConfigDone 首次所有配置加载完成后调用
	//
	// 调用时机:
	//   - 仅在首次启动且所有配置加载完成后调用一次
	//   - 热重载时不会重复调用
	//   - 使用 atomic.CompareAndSwap 确保全局只调用一次
	//
	// 适用场景:
	//   - 构建配置间的关联关系（如物品配置关联到商店配置）
	//   - 初始化全局缓存或索引结构
	//   - 启动需要依赖配置数据的后台服务
	//   - 执行配置完整性校验
	//   - 预计算复杂的业务数据
	//
	// 注意事项:
	//   - 此方法应该是幂等的（可重复执行不会产生副作用）
	//   - 避免执行耗时过长的操作，以免阻塞启动流程
	//   - 如果初始化失败，应该记录日志但不要 panic
	OnFirstAllConfigDone()
}

// =============================================================================
// 配置处理器接口（用于扩展支持新的配置文件格式）
// =============================================================================

// ConfigHandler 配置处理器接口（完整版）
// 定义不同格式配置文件（如 JSON、XML、Excel 等）的读取和解析接口
// 每个处理器负责处理特定格式的配置文件，并将其转换为统一的配置对象列表
//
// 内置处理器:
//   - ExcelConfigHandler: 处理 .xlsx/.xls 文件
//   - JsonConfigHandler: 处理 .json 文件
//   - TsvConfigHandler: 处理 .tsv 文件
//
// 自定义处理器示例:
//
//	type YamlConfigHandler struct {}
//
//	func (h *YamlConfigHandler) TypeName() string { return "yaml" }
//
//	func (h *YamlConfigHandler) ReadToFrontEndDataList(configName, filePath string) interface{} {
//	    data := h.parseYamlFile(filePath)
//	    return &dto.FrontEndConfigDto{DataList: data}
//	}
//
//	func (h *YamlConfigHandler) ReadConfigAndORM(typ reflect.Type, configName, filePath string) []interface{} {
//	    data := h.parseYamlFile(filePath)
//	    return h.convertToType(typ, data)
//	}
type ConfigHandler interface {
	// TypeName 处理器类型名
	// 返回处理器支持的文件类型名称，用于注册和查找处理器
	// 返回值:
	//   string: 处理器类型名称，如 "json", "xml", "excel" 等
	TypeName() string

	// ReadToFrontEndDataList 读取配置并转为前端数据列表
	// 读取配置文件并转换为前端可用的数据传输对象
	// 主要用于配置管理界面或API输出
	// 参数:
	//   configName: 配置名称（通常是文件名去掉扩展名）
	//   configFileFullPath: 配置文件的完整路径
	// 返回值:
	//   interface{}: 前端配置数据传输对象（实际类型为*dto.FrontEndConfigDto）
	ReadToFrontEndDataList(configName, configFileFullPath string) interface{}

	// ReadConfigAndORM 读取配置并转换为对象列表
	// 读取配置文件并使用反射将其转换为指定类型的对象列表
	// 这是配置系统的核心方法，用于将文件数据转换为Go结构体
	// 参数:
	//   typ: 目标配置对象的反射类型
	//   configName: 配置名称
	//   configFileFullPath: 配置文件的完整路径
	// 返回值:
	//   []interface{}: 配置对象实例列表，每个元素都是typ类型的实例
	ReadConfigAndORM(typ reflect.Type, configName, configFileFullPath string) []interface{}
}

// IConfigHandler 配置处理器接口（简化版）
// 提供最基础的配置读取功能，用于简单的配置处理场景
// 如果需要完整功能，建议使用 ConfigHandler 接口
type IConfigHandler interface {
	// ReadToFrontEndDataList 读取配置文件并转换为前端数据格式
	//
	// 参数:
	//   configName: 配置名称（通常是文件名去掉扩展名）
	//   filePath: 配置文件的完整路径
	//
	// 返回值:
	//   interface{}: 应该返回 *dto.FrontEndConfigDto 类型
	//     其中 DataList 字段包含解析后的配置数据数组
	//     每个数组元素是 map[string]interface{} 格式
	ReadToFrontEndDataList(configName string, filePath string) interface{}
}

// =============================================================================
// 配置监听器接口（用于配置变更事件处理）
// =============================================================================

// IConfigListener 配置监听器接口
// 实现此接口可以监听特定类型配置的变更事件
// 提供比 IBusinessConfigManager 更细粒度的控制
//
// 使用场景:
//   - 监听特定配置类型的变更
//   - 实现配置级别的验证和后处理
//   - 与第三方系统集成配置变更通知
type IConfigListener interface {
	// OnConfigDataChange 配置数据变更时调用
	//
	// 参数:
	//   configType: 配置的反射类型信息
	//   dataList: 变更后的配置数据列表
	//
	// 调用时机:
	//   - 配置文件重新加载并解析完成后
	//   - 数据已经转换为目标结构体类型
	OnConfigDataChange(configType reflect.Type, dataList []interface{})
}
