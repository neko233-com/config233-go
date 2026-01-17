package config233

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/neko233-com/config233-go/pkg/config233/dto"
	jsonhandler "github.com/neko233-com/config233-go/pkg/config233/json"
)

// loadJsonConfigThreadSafe 线程安全的 JSON 配置加载（用于并行加载）
func (cm *ConfigManager233) loadJsonConfigThreadSafe(filePath string) error {
	// 创建 JSON 处理器
	handler := &jsonhandler.JsonConfigHandler{}

	// 获取文件名（不含扩展名）作为配置名
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	// 读取前端数据格式（不需要锁）
	configDto := handler.ReadToFrontEndDataList(fileName, filePath).(*dto.FrontEndConfigDto)
	if configDto.DataList == nil {
		return nil // 空文件，跳过
	}

	// 转换为配置映射
	configMap := make(map[string]interface{})
	for _, item := range configDto.DataList {
		// 尝试从 map 中提取 ID（支持 "id", "ID", "Id" 等字段）
		var id string
		if idVal, ok := item["id"]; ok {
			if str, ok := idVal.(string); ok {
				id = str
			} else {
				id = fmt.Sprintf("%v", idVal)
			}
		} else if idVal, ok := item["ID"]; ok {
			if str, ok := idVal.(string); ok {
				id = str
			} else {
				id = fmt.Sprintf("%v", idVal)
			}
		} else if idVal, ok := item["Id"]; ok {
			if str, ok := idVal.(string); ok {
				id = str
			} else {
				id = fmt.Sprintf("%v", idVal)
			}
		}

		if id != "" {
			// 如果有注册的类型，转换为具体结构体
			if converted, err := cm.convertMapToRegisteredStruct(fileName, item); err == nil {
				configMap[id] = converted
			} else {
				// 转换失败则使用原始 map
				configMap[id] = item
			}
		}
	}

	// Convert to []interface{}
	slice := make([]interface{}, len(configDto.DataList))
	for i, v := range configDto.DataList {
		// 尝试转换为注册的结构体类型
		if converted, err := cm.convertMapToRegisteredStruct(fileName, v); err == nil {
			slice[i] = converted
			getLogger().Info("成功转换JSON配置项", "index", i, "configName", fileName, "itemId", v["itemId"])
		} else {
			// 转换失败则使用原始 map
			slice[i] = v
			getLogger().Error(err, "转换JSON配置项失败", "index", i, "configName", fileName, "data", v)
		}
	}

	// 加锁更新共享数据
	cm.mutex.Lock()
	cm.configs[fileName] = configDto.DataList
	cm.configMaps[fileName] = configMap
	cm.mutex.Unlock()

	// 更新缓存（内部已有锁保护）
	cm.setConfigCache(fileName, configMap, slice)

	return nil
}

// loadJsonConfig 从JSON文件加载配置
// 使用 JSON 处理器读取并解析 JSON 配置文件
// 参数:
//
//	filePath: JSON 配置文件的路径
//
// 返回值:
//
//	error: 加载过程中的错误
func (cm *ConfigManager233) loadJsonConfig(filePath string) error {
	// 直接调用线程安全版本
	return cm.loadJsonConfigThreadSafe(filePath)
}
