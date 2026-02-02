package config233

import "reflect"

// ConfigDataChangeListener 配置数据变更监听器
// 定义配置数据发生变化时的回调接口
// 当配置数据被重新加载或更新时，会触发此监听器的回调方法
// 实现此接口的对象可以监听特定类型的配置变化
type ConfigDataChangeListener interface {
	// OnConfigDataChange 当配置数据发生变化时被调用
	// 参数:
	//   typ: 发生变化的配置数据类型
	//   dataList: 新的配置数据列表
	OnConfigDataChange(typ reflect.Type, dataList []interface{})
}
