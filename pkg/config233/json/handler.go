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

	if jsonTopLevelKind(data) == '{' {
		err = fmt.Errorf("json config %q (%s) must be an array of objects, but the top-level value is an object", configName, configFileFullPath)
		slog.Error("JSON配置格式不正确", "configName", configName, "path", configFileFullPath, "error", err, "topLevelKind", "object", "contentPreview", jsonContentPreview(data, 2048))
		panic(err)
	}

	var dataList []map[string]interface{}
	err = json.Unmarshal(data, &dataList)
	if err != nil {
		err = fmt.Errorf("parse json config %q (%s) into []map[string]interface{} failed: %w", configName, configFileFullPath, err)
		slog.Error("解析JSON配置失败", "configName", configName, "path", configFileFullPath, "error", err, "contentPreview", jsonContentPreview(data, 2048))
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

	if jsonTopLevelKind(data) == '{' {
		err = fmt.Errorf("json config %q (%s) must be an array of objects, but the top-level value is an object", configName, configFileFullPath)
		slog.Error("JSON配置格式不正确", "configName", configName, "path", configFileFullPath, "error", err, "topLevelKind", "object", "contentPreview", jsonContentPreview(data, 2048))
		panic(err)
	}

	// 创建切片类型
	sliceType := reflect.SliceOf(typ)
	slicePtr := reflect.New(sliceType)
	sliceVal := slicePtr.Elem()

	err = json.Unmarshal(data, slicePtr.Interface())
	if err != nil {
		err = fmt.Errorf("parse json config %q (%s) into []%s failed: %w", configName, configFileFullPath, typ.String(), err)
		slog.Error("解析JSON配置失败", "configName", configName, "path", configFileFullPath, "error", err, "targetType", typ.String(), "contentPreview", jsonContentPreview(data, 2048))
		panic(err)
	}

	result := make([]interface{}, sliceVal.Len())
	for i := 0; i < sliceVal.Len(); i++ {
		result[i] = sliceVal.Index(i).Interface()
	}

	return result
}
