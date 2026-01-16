package excel

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

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

	// 第 1 列 column 完全没有跳过
	// 固定行结构（0-based 索引）：
	// 第 1 行 (index 0): 空行或注释
	// 第 2 行 (index 1): 中文字段名
	// 第 3 行 (index 2): Client 字段名
	// 第 4 行 (index 3): 类型 (type)
	// 第 5 行 (index 4): Server 字段名 (服务端使用这一行作为字段名)
	// 第 6 行 (index 5): 数据开始
	const (
		clientRowIndex = 2 // Client 字段名行
		typeRowIndex   = 3 // 类型行
		serverRowIndex = 4 // Server 字段名行 (使用这个)
		dataStartIndex = 5 // 数据开始行
	)

	// 检查行数是否足够
	if len(rows) <= dataStartIndex {
		return &dto.FrontEndConfigDto{
			DataList:         nil,
			Type:             h.TypeName(),
			Suffix:           "xlsx",
			ConfigNameSimple: configName,
		}
	}

	// Server 行作为字段名，从第 2 列开始（跳过第 1 列的标识）
	headers := rows[serverRowIndex]

	// 获取类型信息
	var types []string
	if typeRowIndex < len(rows) {
		types = rows[typeRowIndex]
	}

	var dataList []map[string]interface{}

	// 从数据行开始读取
	for _, row := range rows[dataStartIndex:] {
		item := make(map[string]interface{})
		// 从第二列开始（跳过第一列的标识符）
		for i := 1; i < len(row); i++ {
			if i < len(headers) {
				fieldName := headers[i]
				if fieldName == "" {
					continue
				}

				cellValue := row[i]

				// 如果有类型信息，进行类型转换
				if types != nil && i < len(types) {
					item[fieldName] = h.convertValue(cellValue, types[i])
				} else {
					item[fieldName] = cellValue
				}
			}
		}
		if len(item) > 0 {
			// 只添加非空行
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

	// 固定行结构（0-based 索引）：
	// 第 5 行 (index 4): Server 字段名
	// 第 6 行 (index 5): 数据开始
	const (
		serverRowIndex = 4 // Server 字段名行
		dataStartIndex = 5 // 数据开始行
	)

	// 检查行数是否足够
	if len(rows) <= dataStartIndex {
		return nil
	}

	headers := rows[serverRowIndex]
	var result []interface{}

	// 构建 header 名称到 struct 字段名的映射，使用灵活匹配策略
	headerToField := make(map[string]string)
	for _, hdr := range headers {
		h := strings.TrimSpace(hdr)
		if h == "" {
			continue
		}
		// 在 struct 字段中查找与 header 匹配的字段名或 tag（不区分大小写）
		// 优先匹配 `config233_column` 标签，如果没有标签则使用字段名或首字母小写的字段名
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			// 候选名称：config233_column 标签��字段名
			columnTag := f.Tag.Get("config233_column")
			fieldName := f.Name

			// 优先使用 config233_column 标签
			if columnTag != "" && strings.EqualFold(columnTag, h) {
				headerToField[h] = fieldName
				break
			}

			// 如果没有 config233_column 标签，则使用字段名匹配（不区分大小写）
			if columnTag == "" && (strings.EqualFold(fieldName, h) || strings.EqualFold(lowerFirst(fieldName), h)) {
				headerToField[h] = fieldName
				break
			}
		}
	}

	// 从数据行开始读取
	for _, row := range rows[dataStartIndex:] {
		obj := reflect.New(typ).Elem()

		// 从第二列开始（跳过第一列的标识符）
		for i := 1; i < len(row); i++ {
			if i >= len(headers) {
				continue
			}

			header := strings.TrimSpace(headers[i])
			if header == "" {
				continue
			}

			goFieldName, ok := headerToField[header]
			if !ok {
				// 没有映射到 struct 字段，跳过
				continue
			}

			field := obj.FieldByName(goFieldName)
			if !field.IsValid() || !field.CanSet() {
				continue
			}

			h.setFieldValue(field, row[i])
		}

		result = append(result, obj.Interface())
	}

	// 配置加载完成后，执行生命周期方法
	for i, item := range result {
		// 获取指针实际类型，以便调用方法
		itemPtr := reflect.ValueOf(item).Addr().Interface()

		// 1. 调用 AfterLoad() 方法（如果实现了 IConfigLifecycle 接口）
		if lifecycle, ok := itemPtr.(interface{ AfterLoad() }); ok {
			lifecycle.AfterLoad()
		}

		// 2. 调用 Check() 方法（如果实现了 IConfigValidator 接口）
		if validator, ok := itemPtr.(interface{ Check() error }); ok {
			if err := validator.Check(); err != nil {
				// 控制台红色输出校验错误
				fmt.Printf("\033[31m[ERROR] 配置校验失败 [%s] index=%d: %v\033[0m\n", configName, i, err)
			}
		}
	}

	return result
}

// lowerFirst 将字符串首字母转为小写（用于将 Go 字段名如 "Id" 对应到 header 的 "id"）
func lowerFirst(s string) string {
	// 将首字母小写（用于将 Go 字段名如 "Id" 对应到 header 的 "id"）
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// setFieldValue 设置字段值，自动转换 string 到目标类型
func (h *ExcelConfigHandler) setFieldValue(field reflect.Value, value string) {
	// 空字符串处理：对于数值类型设置为 0，字符串保持空，布尔���型为 false
	if value == "" {
		switch field.Kind() {
		case reflect.String:
			field.SetString("")
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(0)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field.SetUint(0)
		case reflect.Float32, reflect.Float64:
			field.SetFloat(0)
		case reflect.Bool:
			field.SetBool(false)
		}
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			fmt.Printf("\033[31m[ERROR] 字段类型转换失败: 无法将 '%s' 转换为 int, 错误: %v\033[0m\n", value, err)
			return
		}
		field.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			fmt.Printf("\033[31m[ERROR] 字段类型转换失败: 无法将 '%s' 转换为 uint, 错误: %v\033[0m\n", value, err)
			return
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Printf("\033[31m[ERROR] 字段类型转换失败: 无法将 '%s' 转换为 float, 错误: %v\033[0m\n", value, err)
			return
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			// 尝试其他布尔值格式
			lower := strings.ToLower(strings.TrimSpace(value))
			switch lower {
			case "true", "1", "yes", "on", "enabled":
				field.SetBool(true)
			case "false", "0", "no", "off", "disabled", "":
				field.SetBool(false)
			default:
				fmt.Printf("\033[31m[ERROR] 字段类型转换失败: 无法将 '%s' 转换为 bool\033[0m\n", value)
			}
			return
		}
		field.SetBool(boolVal)

	default:
		fmt.Printf("\033[31m[ERROR] 不支持的字段类型: %v\033[0m\n", field.Kind())
	}
}

// convertValue 根据类型字符串转换值
func (h *ExcelConfigHandler) convertValue(value string, typeStr string) interface{} {
	// 空值直接返回
	if value == "" {
		return value
	}

	switch typeStr {
	case "int", "int32":
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	case "long", "int64":
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	case "float", "float32":
		if floatVal, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatVal)
		}
	case "double", "float64":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	case "bool", "boolean":
		// 支持多种 bool 格式
		lower := strings.ToLower(value)
		return lower == "true" || lower == "1" || lower == "yes"
	case "json":
		// JSON 类型保持字符串，由调用方自行解析
		return value
	}

	// 默认返回字符串
	return value
}
