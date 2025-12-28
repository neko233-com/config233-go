package config233

import "reflect"

// FieldUpdateListener 字段更新监听器
// 当配置数据发生变化时，自动更新指定对象的字段
// 通常用于实现配置的热更新功能
type FieldUpdateListener struct {
	obj       interface{}   // 要更新的对象
	field     reflect.Value // 要更新的字段反射值
	fieldType reflect.Type  // 字段对应的配置类型
}

// OnConfigDataChange 配置数据变更时调用
// 当配置数据发生变化时，此方法会被触发来更新对应的字段
// 它会重新构建 UID 到配置对象的映射，并更新对象的字段
// 参数:
//   typ: 发生变化的配置数据类型
//   dataList: 新的配置数据列表
func (l *FieldUpdateListener) OnConfigDataChange(typ reflect.Type, dataList []interface{}) {
	if typ != l.fieldType {
		return
	}

	// 更新字段
	uidMap := make(map[interface{}]interface{})
	for _, item := range dataList {
		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		// 查找 UID 字段
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			if field.Tag.Get("config233") == "uid" {
				fieldVal := val.Field(i)
				uidMap[fieldVal.Interface()] = item
				break
			}
		}
	}

	l.field.Set(reflect.ValueOf(uidMap))
}
