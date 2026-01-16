package config233

// IConfigLifecycle 配置生命周期接口
// 实现此接口的配置结构体可以在加载后执行自定��逻辑
type IConfigLifecycle interface {
	// AfterLoad 配置加载后调用
	// 可以在这里进行数据预处理、建立索引、缓存分组等
	AfterLoad()
}

// IConfigValidator 配置校验接口
// 实现此接口的配置结构体可以在加载后进行数据校验
type IConfigValidator interface {
	// Check 配置校验
	// 返回 nil 表示校验通过，否则返回错误信息
	Check() error
}
