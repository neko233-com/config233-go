package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/neko233-com/config233-go/pkg/config233/dto"
)

// JsonConfigHandler JSON 配置处理器
// 负责处理 JSON 格式的配置文件，读取并解析为配置对象
type JsonConfigHandler struct{}

func jsonTopLevelKind(data []byte) byte {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return 0
	}
	return data[0]
}

func jsonContentPreview(data []byte, limit int) string {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	text := strings.TrimSpace(string(data))
	if limit > 0 && len(text) > limit {
		return text[:limit] + "...(truncated)"
	}
	return text
}

func unmarshalJSONDataList(configName, configFileFullPath string, data []byte) ([]map[string]interface{}, string, error) {
	switch jsonTopLevelKind(data) {
	case '{':
		var item map[string]interface{}
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, "object", fmt.Errorf("parse json config %q (%s) as object failed: %w", configName, configFileFullPath, err)
		}
		return []map[string]interface{}{item}, "object", nil
	case '[':
		var dataList []map[string]interface{}
		if err := json.Unmarshal(data, &dataList); err != nil {
			return nil, "array", fmt.Errorf("parse json config %q (%s) as array failed: %w", configName, configFileFullPath, err)
		}
		return dataList, "array", nil
	case 0:
		return nil, "empty", nil
	default:
		return nil, fmt.Sprintf("%q", jsonTopLevelKind(data)), fmt.Errorf("json config %q (%s) must start with object or array, got %q", configName, configFileFullPath, jsonTopLevelKind(data))
	}
}

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
	data, err := os.ReadFile(configFileFullPath)
	if err != nil {
		err = fmt.Errorf("read json config %q (%s) failed: %w", configName, configFileFullPath, err)
		slog.Error("读取JSON配置文件失败", "configName", configName, "path", configFileFullPath, "error", err)
		panic(err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return &dto.FrontEndConfigDto{
			DataList:         nil,
			Type:             h.TypeName(),
			Suffix:           "json",
			ConfigNameSimple: configName,
		}
	}

	dataList, topLevelKind, err := unmarshalJSONDataList(configName, configFileFullPath, data)
	if err != nil {
		err = fmt.Errorf("parse json config %q (%s) into data list failed: %w", configName, configFileFullPath, err)
		slog.Error("解析JSON配置失败", "configName", configName, "path", configFileFullPath, "error", err, "topLevelKind", topLevelKind, "contentPreview", jsonContentPreview(data, 4096))
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
	data, err := os.ReadFile(configFileFullPath)
	if err != nil {
		err = fmt.Errorf("read json config %q (%s) failed: %w", configName, configFileFullPath, err)
		slog.Error("读取JSON配置文件失败", "configName", configName, "path", configFileFullPath, "error", err)
		panic(err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}

	switch jsonTopLevelKind(data) {
	case '{':
		instancePtr := reflect.New(typ)
		if err := json.Unmarshal(data, instancePtr.Interface()); err != nil {
			err = fmt.Errorf("parse json config %q (%s) into %s failed: %w", configName, configFileFullPath, typ.String(), err)
			slog.Error("解析JSON配置失败", "configName", configName, "path", configFileFullPath, "error", err, "targetType", typ.String(), "topLevelKind", "object", "contentPreview", jsonContentPreview(data, 4096))
			panic(err)
		}
		return []interface{}{instancePtr.Elem().Interface()}
	case '[':
		// 创建切片类型
		sliceType := reflect.SliceOf(typ)
		slicePtr := reflect.New(sliceType)
		sliceVal := slicePtr.Elem()

		if err := json.Unmarshal(data, slicePtr.Interface()); err != nil {
			err = fmt.Errorf("parse json config %q (%s) into []%s failed: %w", configName, configFileFullPath, typ.String(), err)
			slog.Error("解析JSON配置失败", "configName", configName, "path", configFileFullPath, "error", err, "targetType", typ.String(), "topLevelKind", "array", "contentPreview", jsonContentPreview(data, 4096))
			panic(err)
		}

		result := make([]interface{}, sliceVal.Len())
		for i := 0; i < sliceVal.Len(); i++ {
			result[i] = sliceVal.Index(i).Interface()
		}

		return result
	default:
		err = fmt.Errorf("json config %q (%s) must start with object or array, got %q", configName, configFileFullPath, jsonTopLevelKind(data))
		slog.Error("JSON配置格式不正确", "configName", configName, "path", configFileFullPath, "error", err, "contentPreview", jsonContentPreview(data, 4096))
		panic(err)
	}
}
