package config233

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/neko233-com/config233-go/pkg/config233/dto"
	"github.com/neko233-com/config233-go/pkg/config233/excel"
)

// loadExcelConfigThreadSafe 线程安全的 Excel 配置加载（用于并行加载）
func (cm *ConfigManager233) loadExcelConfigThreadSafe(filePath string) error {
	// 创建 Excel 处理器
	handler := &excel.ExcelConfigHandler{}

	// 获取文件名（不含扩展名）作为配置名
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// 读取前端数据格式（不需要锁）
	configDto := handler.ReadToFrontEndDataList(fileName, filePath).(*dto.FrontEndConfigDto)
	if configDto.DataList == nil {
		return nil // 空文件，跳过
	}

	// 转换为配置映射与切片（每条数据只转换一次，避免 AfterLoad 重复触发）
	configMap := make(map[string]interface{})
	slice := make([]interface{}, 0, len(configDto.DataList))
	for i, item := range configDto.DataList {
		converted := any(item)
		if c, err := cm.convertMapToRegisteredStruct(fileName, item); err == nil {
			converted = c
		} else {
			getLogger().Error(err, "转换配置项失败", "index", i, "configName", fileName, "data", item)
		}

		// 优先使用 id/ID/Id/itemId 字段作为配置 ID
		var id string
		if idVal, ok := item["id"]; ok && idVal != "" {
			id = fmt.Sprintf("%v", idVal)
		} else if idVal, ok := item["ID"]; ok && idVal != "" {
			id = fmt.Sprintf("%v", idVal)
		} else if idVal, ok := item["Id"]; ok && idVal != "" {
			id = fmt.Sprintf("%v", idVal)
		} else if idVal, ok := item["itemId"]; ok && idVal != "" {
			id = fmt.Sprintf("%v", idVal)
		}
		if id != "" {
			configMap[id] = converted
		}

		slice = append(slice, converted)
	}

	// 加锁更新共享数据
	cm.mutex.Lock()
	cm.configs[fileName] = configDto.DataList
	cm.configMaps[fileName] = configMap
	cm.mutex.Unlock()

	// 更新缓存（内部已有锁保护）
	cm.setConfigCache(fileName, configMap, slice)

	getLogger().Info("Excel配置加载完成", "configName", fileName, "count", len(slice))

	// 导出配置到文件（如果开启）
	cm.ExportConfigToJSON(fileName, slice)

	return nil
}

// loadExcelConfig 从Excel文件加载配置
// 使用 Excel 处理器读取并解析 Excel 配置文件
// 参数:
//
//	filePath: Excel 配置文件的路径
//
// 返回值:
//
//	error: 加载过程中的错误
func (cm *ConfigManager233) loadExcelConfig(filePath string) error {
	// 直接调用线程安全版本
	return cm.loadExcelConfigThreadSafe(filePath)
}
