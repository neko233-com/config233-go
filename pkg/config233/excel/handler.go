package excel

import (
	"reflect"
	"strconv"

	"github.com/neko233-com/config233-go/pkg/config233/dto"

	"github.com/xuri/excelize/v2"
)

// ExcelConfigHandler Excel 配置处理器
// 负责处理 Excel 格式的配置文件，读取并解析为配置对象
type ExcelConfigHandler struct{}

// TypeName 返回处理器类型名
// 返回值:
//
//	string: "excel"
func (h *ExcelConfigHandler) TypeName() string {
	return "excel"
}

// ReadToFrontEndDataList 读取配置并转为前端数据列表
// 读取 Excel 配置文件并转换为前端可用的数据传输对象
// 默认读取第一个工作表，第一行为表头
// 参数:
//
//	configName: 配置名称
//	configFileFullPath: Excel 配置文件的完整路径
//
// 返回值:
//
//	interface{}: 包含解析后数据的传输对象
func (h *ExcelConfigHandler) ReadToFrontEndDataList(configName, configFileFullPath string) interface{} {
	f, err := excelize.OpenFile(configFileFullPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 获取第一个工作表的名称
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return &dto.FrontEndConfigDto{
			DataList:         nil,
			Type:             h.TypeName(),
			Suffix:           "xlsx",
			ConfigNameSimple: configName,
		}
	}
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		panic(err)
	}

	// fieldNameRow = 4 (index 3)
	// dataStartRow = 5 (index 4)
	if len(rows) < 4 {
		return &dto.FrontEndConfigDto{
			DataList:         nil,
			Type:             h.TypeName(),
			Suffix:           "xlsx",
			ConfigNameSimple: configName,
		}
	}

	headers := rows[3]
	var dataList []map[string]interface{}

	if len(rows) >= 5 {
		for _, row := range rows[4:] {
			item := make(map[string]interface{})
			for i, cell := range row {
				if i < len(headers) {
					fieldName := headers[i]
					if fieldName == "" {
						continue
					}
					item[fieldName] = cell
				}
			}
			dataList = append(dataList, item)
		}
	}

	return &dto.FrontEndConfigDto{
		DataList:         dataList,
		Type:             h.TypeName(),
		Suffix:           "xlsx",
		ConfigNameSimple: configName,
	}
}

// ReadConfigAndORM 读取配置并转换为对象列表
func (h *ExcelConfigHandler) ReadConfigAndORM(typ reflect.Type, configName, configFileFullPath string) []interface{} {
	f, err := excelize.OpenFile(configFileFullPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 获取第一个工作表的名称
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil
	}
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		panic(err)
	}

	// fieldNameRow = 4 (index 3)
	// dataStartRow = 5 (index 4)
	if len(rows) < 4 {
		return nil
	}

	headers := rows[3]
	var result []interface{}

	if len(rows) >= 5 {
		for _, row := range rows[4:] {
			obj := reflect.New(typ).Elem()

			for i, cell := range row {
				if i >= len(headers) {
					continue
				}

				fieldName := headers[i]
				if fieldName == "" {
					continue
				}
				field := obj.FieldByName(fieldName)
				if !field.IsValid() || !field.CanSet() {
					continue
				}

				h.setFieldValue(field, cell)
			}

			result = append(result, obj.Interface())
		}
	}

	return result
}

// setFieldValue 设置字段值
func (h *ExcelConfigHandler) setFieldValue(field reflect.Value, value string) {
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
