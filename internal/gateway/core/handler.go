package core

// Handler 请求处理器接口
// 网关中所有的处理逻辑都通过实现这个接口来完成
// 如CORS处理、限流、熔断、认证、路由匹配等都实现了这个接口
type Handler interface {
	// Handle 处理请求，返回是否继续执行后续处理器
	// 参数ctx包含当前请求的上下文信息和状态
	// 返回true表示继续处理，返回false表示中断处理链
	// 处理器可以通过返回false来阻止后续处理器执行，例如认证失败时
	Handle(ctx *Context) bool
}

// HandlerFunc 是Handler接口的函数类型实现
// 这是一种函数适配器模式，允许将普通函数转换为Handler接口
// 便于快速创建简单的处理器而无需定义完整的结构体
type HandlerFunc func(ctx *Context) bool

// Handle 实现Handler接口
// 将函数类型转换为Handler接口类型，这样普通函数也能作为处理器使用
// 这种模式在Go语言中很常见，例如http.HandlerFunc
func (f HandlerFunc) Handle(ctx *Context) bool {
	return f(ctx)
}

// HandlerChain 处理器链，包含多个按顺序执行的处理器
// 实现了职责链模式，将多个处理器组合起来依次执行
// 处理器链是网关处理请求的核心机制，所有请求都会经过处理器链处理
type HandlerChain struct {
	handlers []Handler // 按顺序存储的处理器列表
}

// NewHandlerChain 创建新的处理器链
// 初始化一个空的处理器链，准备添加处理器
// 处理器链在网关初始化时创建，贯穿整个请求生命周期
func NewHandlerChain() *HandlerChain {
	return &HandlerChain{
		handlers: make([]Handler, 0),
	}
}

// Add 添加处理器到链末尾
// 参数handler是实现了Handler接口的处理器
// 返回处理器链自身，支持链式调用，如chain.Add(h1).Add(h2)
// 处理器的添加顺序决定了执行顺序，先添加的先执行
func (c *HandlerChain) Add(handler Handler) *HandlerChain {
	c.handlers = append(c.handlers, handler)
	return c
}

// AddFunc 添加处理器函数到链末尾
// 参数handlerFunc是一个函数，会被转换为HandlerFunc适配器
// 这是一个便捷方法，用于快速添加匿名函数作为处理器
// 例如：chain.AddFunc(func(ctx *Context) bool { ... })
func (c *HandlerChain) AddFunc(handlerFunc HandlerFunc) *HandlerChain {
	return c.Add(handlerFunc)
}

// Execute 执行处理器链
// 按照添加顺序依次执行每个处理器
// 如果某个处理器返回false或上下文已响应，则中断后续处理
// 这是处理器链的核心方法，网关引擎会调用它来处理请求
func (c *HandlerChain) Execute(ctx *Context) {
	for _, handler := range c.handlers {
		if !handler.Handle(ctx) || ctx.IsResponded() {
			// 处理器返回false或上下文已响应，中断处理链
			// 例如：认证失败、请求被限流、已发送响应等情况
			break
		}
	}
}
