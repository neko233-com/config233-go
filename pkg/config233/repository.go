package config233

import (
	"reflect"
	"sync"
)

// ConfigDataRepository 配置数据仓库
// 负责存储和管理所有配置数据的中央仓库
// 支持数据存储、检索和变更监听功能
// 线程安全，支持并发读写操作
type ConfigDataRepository struct {
	typeToDataList        map[reflect.Type][]interface{}              // 类型到数据列表的映射
	typeToChangeListeners map[reflect.Type][]ConfigDataChangeListener // 类型到变更监听器的映射
	mu                    sync.RWMutex                                // 读写锁，保证线程安全
}

// NewConfigDataRepository 创建新的仓库
// 初始化配置数据仓库，创建空的映射表
// 返回值:
//
//	*ConfigDataRepository: 新创建的仓库实例
func NewConfigDataRepository() *ConfigDataRepository {
	return &ConfigDataRepository{
		typeToDataList:        make(map[reflect.Type][]interface{}),
		typeToChangeListeners: make(map[reflect.Type][]ConfigDataChangeListener),
	}
}

// Put 存储数据
// 将配置数据列表存储到仓库中，并触发所有相关的变更监听器
// 参数:
//
//	typ: 配置数据的类型
//	dataList: 配置数据列表
func (r *ConfigDataRepository) Put(typ reflect.Type, dataList []interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.typeToDataList[typ] = dataList

	// 触发变更监听
	listeners := r.typeToChangeListeners[typ]
	for _, listener := range listeners {
		listener.OnConfigDataChange(typ, dataList)
	}
}

// Get 获取数据
// 根据类型获取对应的配置数据列表
// 参数:
//
//	typ: 配置数据的类型
//
// 返回值:
//
//	[]interface{}: 配置数据列表，如果不存在则返回 nil
func (r *ConfigDataRepository) Get(typ reflect.Type) []interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.typeToDataList[typ]
}

// AddChangeListener 添加变更监听器
// 为指定的配置类型添加变更监听器，当该类型的数据发生变化时会触发监听器
// 参数:
//
//	typ: 配置数据的类型
//	listener: 变更监听器实例
func (r *ConfigDataRepository) AddChangeListener(typ reflect.Type, listener ConfigDataChangeListener) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.typeToChangeListeners[typ] = append(r.typeToChangeListeners[typ], listener)
}

// GetUIDMap 获取 UID 映射
// 根据配置类中带有 "config233":"uid" 标签的字段，创建 UID 到对象实例的映射
// 用于快速通过唯一标识符查找配置对象
// 参数:
//
//	typ: 配置数据的类型
//
// 返回值:
//
//	map[interface{}]interface{}: UID 到配置对象实例的映射
func (r *ConfigDataRepository) GetUIDMap(typ reflect.Type) map[interface{}]interface{} {
	dataList := r.Get(typ)
	if len(dataList) == 0 {
		return nil
	}

	uidMap := make(map[interface{}]interface{})
	for _, item := range dataList {
		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		// 查找带有 uid tag 的字段
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			if field.Tag.Get("config233") == "uid" {
				fieldVal := val.Field(i)
				uidMap[fieldVal.Interface()] = item
				break
			}
		}
	}

	return uidMap
}
