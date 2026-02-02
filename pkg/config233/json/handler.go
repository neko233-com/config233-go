package json

import (
	"encoding/json"
	"io/ioutil"
	"reflect"

	"github.com/neko233-com/config233-go/pkg/config233/dto"
)

// JsonConfigHandler JSON 配置处理器
// 负责处理 JSON 格式的配置文件，读取并解析为配置对象
type JsonConfigHandler struct{}

// TypeName 返回处理器类型名
// 返回值:
//
//	string: "json"
func (h *JsonConfigHandler) TypeName() string {
	return "json"
}

// ReadToFrontEndDataList 读取配置并转为前端数据列表
// 读取 JSON 配置文件并转换为前端可用的数据传输对象
// 参数:
//
//	configName: 配置名称
//	configFileFullPath: JSON 配置文件的完整路径
//
// 返回值:
//
//	interface{}: 包含解析后数据的传输对象
func (h *JsonConfigHandler) ReadToFrontEndDataList(configName, configFileFullPath string) interface{} {
	data, err := ioutil.ReadFile(configFileFullPath)
	if err != nil {
		panic(err)
	}

	var dataList []map[string]interface{}
	err = json.Unmarshal(data, &dataList)
	if err != nil {
		panic(err)
	}

	return &dto.FrontEndConfigDto{
		DataList:         dataList,
		Type:             h.TypeName(),
		Suffix:           "json",
		ConfigNameSimple: configName,
	}
}

// ReadConfigAndORM 读取配置并转换为对象列表
// 读取 JSON 配置文件并使用反射转换为指定类型的对象列表
// 参数:
//
//	typ: 目标配置对象的类型
//	configName: 配置名称
//	configFileFullPath: JSON 配置文件的完整路径
//
// 返回值:
//
//	[]interface{}: 配置对象实例列表
func (h *JsonConfigHandler) ReadConfigAndORM(typ reflect.Type, configName, configFileFullPath string) []interface{} {
	data, err := ioutil.ReadFile(configFileFullPath)
	if err != nil {
		panic(err)
	}

	if len(data) == 0 {
		return nil
	}

	// 创建切片类型
	sliceType := reflect.SliceOf(typ)
	slicePtr := reflect.New(sliceType)
	sliceVal := slicePtr.Elem()

	err = json.Unmarshal(data, slicePtr.Interface())
	if err != nil {
		panic(err)
	}

	result := make([]interface{}, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		result[i] = sliceVal.Index(i).Interface()
	}

	return result
}
