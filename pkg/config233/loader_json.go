package config233

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/neko233-com/config233-go/pkg/config233/dto"
	jsonhandler "github.com/neko233-com/config233-go/pkg/config233/json"
)

// loadJsonConfigThreadSafe 线程安全的 JSON 配置加载（用于并行加载）
func (cm *ConfigManager233) loadJsonConfigThreadSafe(filePath string) (err error) {
	// 创建 JSON 处理器
	handler := &jsonhandler.JsonConfigHandler{}

	// 获取文件名（不含扩展名）作为配置名
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	slog.Info("开始加载JSON配置", "configName", fileName, "path", filePath)

	defer func() {
		if r := recover(); r != nil {
			panicErr, ok := r.(error)
			if !ok {
				panicErr = fmt.Errorf("%v", r)
			}
			contentPreview := ""
			topLevelKind := ""
			if raw, readErr := os.ReadFile(filePath); readErr == nil {
				text := strings.TrimSpace(string(raw))
				if len(text) > 4096 {
					contentPreview = text[:4096] + "...(truncated)"
				} else {
					contentPreview = text
				}
				if len(text) > 0 {
					topLevelKind = string(text[0])
				}
			} else {
				contentPreview = fmt.Sprintf("<failed to read content: %v>", readErr)
			}
			err = fmt.Errorf("load json config %q (%s) failed: %w", fileName, filePath, panicErr)
			slog.Error("JSON配置加载失败",
				"configName", fileName,
				"path", filePath,
				"error", err,
				"panic", panicErr,
				"topLevelKind", topLevelKind,
				"contentPreview", contentPreview,
				"stack", string(debug.Stack()),
			)
		}
	}()

	// 读取前端数据格式（不需要锁）
	configDto := handler.ReadToFrontEndDataList(fileName, filePath).(*dto.FrontEndConfigDto)
	if configDto.DataList == nil {
		slog.Info("JSON配置为空，已跳过", "configName", fileName, "path", filePath)
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
			getLogger().Error(err, "转换JSON配置项失败", "index", i, "configName", fileName, "data", item)
		}

		// 尝试从原始 map 中提取 ID（支持 "id", "ID", "Id" 等字段）
		var id string
		if idVal, ok := item["id"]; ok {
			id = fmt.Sprintf("%v", idVal)
		} else if idVal, ok := item["ID"]; ok {
			id = fmt.Sprintf("%v", idVal)
		} else if idVal, ok := item["Id"]; ok {
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

	getLogger().Info("JSON配置加载完成", "configName", fileName, "count", len(slice))

	// 导出配置到文件（如果开启）
	cm.ExportConfigToJSON(fileName, slice)

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
