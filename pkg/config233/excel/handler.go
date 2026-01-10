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

	// Excel 实际格式：
	// 人类第 1 行 (index 0): 注释
	// 人类第 2 行 (index 1): 中文字段名
	// 人类第 3 行 (index 2): Client 字段名
	// 人类第 4 行 (index 3): 类型 (type)
	// 人类第 5 行 (index 4): Server 字段名 ← 使用这一行
	// 人类第 6 行 (index 5): 数据开始
	if len(rows) < 5 {
		return &dto.FrontEndConfigDto{
			DataList:         nil,
			Type:             h.TypeName(),
			Suffix:           "xlsx",
			ConfigNameSimple: configName,
		}
	}

	// Server 字段名在第 5 行（index 4）
	headers := rows[4] // Server 字段名在第 5 行（index 4）
	var dataList []map[string]interface{}

	if len(rows) >= 6 { // 数据从第 6 行开始（index 5）
		for _, row := range rows[5:] {
			item := make(map[string]interface{})
			// 从第二列开始（跳过第一列的标识符，如 "Server", "Client" 等）
			for i := 1; i < len(row); i++ {
				if i < len(headers) {
					fieldName := headers[i]
					if fieldName == "" {
						continue
					}
					item[fieldName] = row[i]
				}
			}
			if len(item) > 0 {
				// 只添加非空行
				dataList = append(dataList, item)
			}
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

	// Excel 实际格式：
	// 人类第 1 行 (index 0): 注释
	// 人类第 2 行 (index 1): 中文字段名
	// 人类第 3 行 (index 2): Client 字段名
	// 人类第 4 行 (index 3): 类型 (type)
	// 人类第 5 行 (index 4): Server 字段名 ← 使用这一行
	// 人类第 6 行 (index 5): 数据开始
	if len(rows) < 5 {
		return nil
	}

	headers := rows[4] // Server 字段名在第 5 行（index 4）
	var result []interface{}

	if len(rows) >= 6 { // 数据从第 6 行开始（index 5）
		for _, row := range rows[5:] {
			obj := reflect.New(typ).Elem()

			// 从第二列开始（跳过第一列的标识符）
			for i := 1; i < len(row); i++ {
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

				h.setFieldValue(field, row[i])
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
