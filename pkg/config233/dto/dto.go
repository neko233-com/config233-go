package dto

// FrontEndConfigDto 前端配置数据传输对象
// 用于向前端传输配置数据的标准化格式
// 包含配置数据列表、类型信息和元数据
type FrontEndConfigDto struct {
	// DataList 配置数据列表，每个元素是一个字段名到值的映射
	DataList []map[string]interface{} `json:"dataList"`
	// Type 配置文件的类型，如 "json", "excel", "tsv"
	Type string `json:"type"`
	// Suffix 文件扩展名，如 "json", "xlsx", "tsv"
	Suffix string `json:"suffix"`
	// ConfigNameSimple 配置的简单名称，不包含路径和扩展名
	ConfigNameSimple string `json:"configNameSimple"`
}
