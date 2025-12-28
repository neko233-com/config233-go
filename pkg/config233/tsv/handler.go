package tsv

import (
	"bufio"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"config233-go/pkg/config233/dto"
)

// TsvConfigHandler TSV 配置处理器
// 负责处理 TSV (Tab-Separated Values) 格式的配置文件，读取并解析为配置对象
type TsvConfigHandler struct{}

// TypeName 返回处理器类型名
// 返回值:
//
//	string: "tsv"
func (h *TsvConfigHandler) TypeName() string {
	return "tsv"
}

// ReadToFrontEndDataList 读取配置并转为前端数据列表
// 读取 TSV 配置文件并转换为前端可用的数据传输对象
// 第一行为表头，后续行为数据行，以制表符(\t)分隔
// 参数:
//
//	configName: 配置名称
//	configFileFullPath: TSV 配置文件的完整路径
//
// 返回值:
//
//	interface{}: 包含解析后数据的传输对象
func (h *TsvConfigHandler) ReadToFrontEndDataList(configName, configFileFullPath string) interface{} {
	data, err := ioutil.ReadFile(configFileFullPath)
	if err != nil {
		panic(err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return &dto.FrontEndConfigDto{
			DataList:         nil,
			Type:             h.TypeName(),
			Suffix:           "tsv",
			ConfigNameSimple: configName,
		}
	}

	headers := strings.Split(strings.TrimSpace(lines[0]), "\t")
	var dataList []map[string]string

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		values := strings.Split(line, "\t")
		item := make(map[string]string)
		for i, value := range values {
			if i < len(headers) {
				item[headers[i]] = value
			}
		}
		dataList = append(dataList, item)
	}

	return &dto.FrontEndConfigDto{
		DataList:         dataList,
		Type:             h.TypeName(),
		Suffix:           "tsv",
		ConfigNameSimple: configName,
	}
}

// ReadConfigAndORM 读取配置并转换为对象列表
func (h *TsvConfigHandler) ReadConfigAndORM(typ reflect.Type, configName, configFileFullPath string) []interface{} {
	data, err := ioutil.ReadFile(configFileFullPath)
	if err != nil {
		panic(err)
	}

	content := string(data)
	scanner := bufio.NewScanner(strings.NewReader(content))

	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) < 2 {
		return nil
	}

	headers := strings.Split(lines[0], "\t")
	var result []interface{}

	for _, line := range lines[1:] {
		values := strings.Split(line, "\t")
		obj := reflect.New(typ).Elem()

		for i, value := range values {
			if i >= len(headers) {
				continue
			}

			fieldName := headers[i]
			field := obj.FieldByName(fieldName)
			if !field.IsValid() || !field.CanSet() {
				continue
			}

			h.setFieldValue(field, value)
		}

		result = append(result, obj.Interface())
	}

	return result
}

// setFieldValue 设置字段值
func (h *TsvConfigHandler) setFieldValue(field reflect.Value, value string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(uintVal)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(floatVal)
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolVal)
		}
	}
}
