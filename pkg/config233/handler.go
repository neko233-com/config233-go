package config233

import "reflect"

// ConfigHandler 配置处理器接口
// 定义不同格式配置文件（如 JSON、XML、Excel 等）的读取和解析接口
// 每个处理器负责处理特定格式的配置文件，并将其转换为统一的配置对象列表
type ConfigHandler interface {
	// TypeName 处理器类型名
	// 返回处理器支持的文件类型名称，如 "json", "xml", "excel" 等
	// 返回值:
	//   string: 处理器类型名称
	TypeName() string

	// ReadToFrontEndDataList 读取配置并转为前端数据列表
	// 读取配置文件并转换为前端可用的数据传输对象
	// 参数:
	//   configName: 配置名称
	//   configFileFullPath: 配置文件的完整路径
	// 返回值:
	//   interface{}: 前端配置数据传输对象（实际类型为*dto.FrontEndConfigDto）
	ReadToFrontEndDataList(configName, configFileFullPath string) interface{}

	// ReadConfigAndORM 读取配置并转换为对象列表
	// 读取配置文件并使用反射将其转换为指定类型的对象列表
	// 参数:
	//   typ: 目标配置对象的类型
	//   configName: 配置名称
	//   configFileFullPath: 配置文件的完整路径
	// 返回值:
	//   []interface{}: 配置对象实例列表
	ReadConfigAndORM(typ reflect.Type, configName, configFileFullPath string) []interface{}
}
