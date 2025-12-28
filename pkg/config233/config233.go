// Package config233 提供统一的配置文件加载、解析和管理功能
// 支持多种配置文件格式（JSON、XML、Excel、TSV）的热更新和ORM映射
package config233

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Config233 统一配置入口类
// 负责配置文件扫描、加载、监听和数据管理
type Config233 struct {
	configDirPath    string                   // 配置目录路径
	scanPackage      string                   // 要扫描的包名（Go中暂未使用）
	excludeFileNames map[string]bool          // 要排除的文件名集合
	fileHandlers     map[string]ConfigHandler // 文件扩展名到处理器的映射
	configRepository *ConfigDataRepository    // 配置数据仓库
	startCalled      bool                     // 是否已调用Start方法
	firstInitDone    bool                     // 是否已完成首次初始化
	classToHotUpdate map[string]bool          // 需要热更新的类映射
	configClasses    map[string]reflect.Type  // 配置名到类型的映射
	mu               sync.RWMutex             // 读写锁
}

// NewConfig233 创建新的 Config233 实例
// 返回初始化后的Config233对象，可以链式调用配置方法
func NewConfig233() *Config233 {
	return &Config233{
		excludeFileNames: make(map[string]bool),
		fileHandlers:     make(map[string]ConfigHandler),
		configRepository: NewConfigDataRepository(),
		classToHotUpdate: make(map[string]bool),
		configClasses:    make(map[string]reflect.Type),
	}
}

// AddConfigHandler 添加配置文件处理器
// ext: 文件扩展名（如"json", "xlsx"）
// handler: 对应的配置处理器实现
// 返回Config233实例支持链式调用
func (c *Config233) AddConfigHandler(ext string, handler ConfigHandler) *Config233 {
	c.fileHandlers[ext] = handler
	return c
}

// GetFileHandlers 获取所有已注册的文件处理器
// 返回文件扩展名到处理器的映射副本
func (c *Config233) GetFileHandlers() map[string]ConfigHandler {
	handlers := make(map[string]ConfigHandler)
	for k, v := range c.fileHandlers {
		handlers[k] = v
	}
	return handlers
}

// AddExcludeFileName 添加要排除的文件名
// 这些文件不会被当作配置文件处理，通常用于排除临时文件或日志文件
// fileNames: 要排除的文件名列表
// 返回Config233实例支持链式调用
func (c *Config233) AddExcludeFileName(fileNames ...string) *Config233 {
	for _, name := range fileNames {
		c.excludeFileNames[name] = true
	}
	return c
}

// Directory 设置配置目录路径
// dirPath: 配置文件的根目录路径
// 返回Config233实例支持链式调用
func (c *Config233) Directory(dirPath string) *Config233 {
	c.configDirPath = dirPath
	return c
}

// ScanPackage 设置要扫描的包名（Go中暂未使用，保持与Kotlin版本兼容）
// pkg: 包名
// 返回Config233实例支持链式调用
func (c *Config233) ScanPackage(pkg string) *Config233 {
	c.scanPackage = pkg
	return c
}

// RegisterConfigClass 注册配置类类型
// name: 配置名（通常是文件名去扩展名）
// typ: 配置对应的Go类型
// 返回Config233实例支持链式调用
func (c *Config233) RegisterConfigClass(name string, typ reflect.Type) *Config233 {
	c.configClasses[name] = typ
	return c
}

// Start 开始加载配置并启动监听
// 这是一个核心方法，会执行以下操作：
// 1. 扫描配置类
// 2. 建立文件名到文件路径的映射
// 3. 加载所有配置文件
// 4. 启动文件监听器进行热更新
// 返回Config233实例
func (c *Config233) Start() *Config233 {
	if c.startCalled {
		return c
	}
	c.startCalled = true

	// 扫描配置类
	configClasses := c.scanConfigClasses()

	// 获取文件映射
	fileMap := c.getFileNameToPathMap()

	// 初始加载配置
	c.loadConfigs(configClasses, fileMap)

	// 启动文件监听
	c.startFileWatcher(fileMap)

	c.firstInitDone = true
	return c
}

// scanConfigClasses 扫描配置类
// 在Go中，我们通过RegisterConfigClass方法手动注册配置类
// 返回配置名到类型的映射
func (c *Config233) scanConfigClasses() map[string]reflect.Type {
	// 返回已注册的配置类
	return c.configClasses
}

// getFileNameToPathMap 获取文件名到路径的映射
// 递归扫描配置目录，建立配置文件名（去扩展名）到完整路径的映射
// 只包含有对应处理器的文件类型
// 返回文件名到路径的映射表
func (c *Config233) getFileNameToPathMap() map[string]string {
	fileMap := make(map[string]string)

	// 递归扫描目录
	filepath.Walk(c.configDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		filename := info.Name()
		if c.excludeFileNames[filename] {
			return nil
		}

		ext := filepath.Ext(filename)
		if ext == "" {
			return nil
		}
		ext = ext[1:] // 移除点

		if _, ok := c.fileHandlers[ext]; !ok {
			return nil
		}

		nameWithoutExt := filename[:len(filename)-len(ext)-1]
		if existing, ok := fileMap[nameWithoutExt]; ok {
			panic(fmt.Sprintf("重复的配置文件名: %s 和 %s", existing, path))
		}
		fileMap[nameWithoutExt] = path
		return nil
	})

	return fileMap
}

// loadConfigs 加载所有配置
// 遍历所有注册的配置类，为每个配置类加载对应的配置文件
// configClasses: 配置名到类型的映射
// fileMap: 配置名到文件路径的映射
func (c *Config233) loadConfigs(configClasses map[string]reflect.Type, fileMap map[string]string) {
	for name, typ := range configClasses {
		path, ok := fileMap[name]
		if !ok {
			continue
		}
		c.loadConfig(typ, name, path)
	}
}

// loadConfig 加载单个配置
// 根据文件扩展名选择对应的处理器，读取并解析配置文件，然后存储到仓库中
// typ: 配置数据类型
// name: 配置名
// path: 配置文件路径
func (c *Config233) loadConfig(typ reflect.Type, name, path string) {
	ext := filepath.Ext(path)[1:]
	handler := c.fileHandlers[ext]
	if handler == nil {
		return
	}

	dataList := handler.ReadConfigAndORM(typ, name, path)
	c.configRepository.Put(typ, dataList)
}

// startFileWatcher 启动文件监听器
// 使用fsnotify监听配置文件的变化，实现热更新功能
// fileMap: 要监听的文件映射
func (c *Config233) startFileWatcher(fileMap map[string]string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		getLogger().Error(err, "创建文件监听器失败")
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					c.handleFileChange(event.Name, fileMap)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				getLogger().Error(err, "文件监听器错误")
			}
		}
	}()

	// 添加所有配置文件到监听器
	for _, path := range fileMap {
		err = watcher.Add(path)
		if err != nil {
			getLogger().Error(err, "添加文件到监听器失败", "path", path)
		}
	}
}

// handleFileChange 处理文件变化事件
// 当配置文件被修改时，重新加载对应的配置数据
// path: 发生变化的文件路径
// fileMap: 文件路径映射表
func (c *Config233) handleFileChange(path string, fileMap map[string]string) {
	// 找到对应的配置名
	var configName string
	var configType reflect.Type
	for name, p := range fileMap {
		if p == path {
			configName = name
			configType = c.configClasses[name]
			break
		}
	}

	if configType != nil {
		c.loadConfig(configType, configName, path)
	}
}

// GetConfigList 获取配置列表
// 根据给定的类型获取所有配置数据的列表
// 参数:
//
//	typ: 配置类的反射类型
//
// 返回值:
//
//	interface{}: 配置数据列表，类型为 []interface{}
func (c *Config233) GetConfigList(typ reflect.Type) interface{} {
	return c.configRepository.Get(typ)
}

// AddConfigChangeListener 添加配置变更监听器
// 为指定的配置类型添加变更监听器，当配置发生变化时会触发监听器
// 参数:
//
//	typ: 配置类的反射类型
//	listener: 配置变更监听器接口实现
//
// 返回值:
//
//	*Config233: 返回自身，支持链式调用
func (c *Config233) AddConfigChangeListener(typ reflect.Type, listener ConfigDataChangeListener) *Config233 {
	c.configRepository.AddChangeListener(typ, listener)
	return c
}

// RegisterForHotUpdate 注册对象用于热更新
// 将对象注册到热更新系统中，支持字段注入和方法监听
// 当配置发生变化时，会自动更新对象的字段或调用指定的方法
// 参数:
//
//	obj: 需要注册热更新的对象指针
func (c *Config233) RegisterForHotUpdate(obj interface{}) {
	typ := reflect.TypeOf(obj)
	typeName := typ.String()
	if c.classToHotUpdate[typeName] {
		return
	}
	c.classToHotUpdate[typeName] = true

	// 处理字段注入
	c.injectFields(obj)

	// 处理方法监听
	c.registerMethods(obj)
}

// injectFields 注入字段
// 扫描对象的字段，查找带有 "config233":"inject" 标签的字段
// 并将对应的配置数据注入到这些字段中
// 当前支持 map[int]*T 类型的字段注入
// 参数:
//
//	obj: 需要注入字段的对象指针
func (c *Config233) injectFields(obj interface{}) {
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if tag := fieldType.Tag.Get("config233"); tag == "inject" {
			// 假设是 map 类型
			if fieldType.Type.Kind() == reflect.Map {
				// 获取配置类型
				elemType := fieldType.Type.Elem()
				uidMap := c.configRepository.GetUIDMap(elemType)

				// 创建正确类型的 map
				mapType := reflect.MapOf(reflect.TypeOf(0), elemType) // int -> *Student
				concreteMap := reflect.MakeMap(mapType)

				for k, v := range uidMap {
					if intKey, ok := k.(int); ok {
						concreteMap.SetMapIndex(reflect.ValueOf(intKey), reflect.ValueOf(v))
					}
				}

				field.Set(concreteMap)
			}
		}
	}
}

// registerMethods 注册方法监听
// 扫描对象的方法，查找需要监听配置变更的方法
// 当配置发生变化时，会调用这些方法
// 注意: Go 中方法没有标签，这里暂时跳过实现，
// 后续可以通过其他方式（如方法名约定）来识别监听方法
// 参数:
//
//	obj: 需要注册方法监听的对象指针
func (c *Config233) registerMethods(obj interface{}) {
	typ := reflect.TypeOf(obj)
	// val := reflect.ValueOf(obj)

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		// Go 中方法没有 tag，这里需要其他方式
		// 暂时跳过
		_ = method
	}
}
