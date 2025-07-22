// Package executor 提供任务执行器功能
package executor

import (
	"context"

	"gateway/pkg/timer"
)

// Executor 任务执行器接口的别名
type Executor = timer.TaskExecutor

// BaseExecutor 基础执行器，提供通用功能
type BaseExecutor struct {
	name        string
	description string
}

// NewBaseExecutor 创建基础执行器实例
// 为其他具体执行器提供基础功能和通用属性
// 参数:
//
//	name: 执行器名称，用于标识和日志记录
//	description: 执行器描述，用于说明执行器的功能
//
// 返回:
//
//	*BaseExecutor: 初始化的基础执行器实例
func NewBaseExecutor(name, description string) *BaseExecutor {
	return &BaseExecutor{
		name:        name,        // 执行器名称
		description: description, // 执行器描述
	}
}

// GetName 获取执行器名称
// 返回执行器的唯一标识名称，用于日志记录和监控
// 返回:
//
//	string: 执行器名称
func (e *BaseExecutor) GetName() string {
	return e.name
}

// GetDescription 获取执行器描述
// 返回执行器的功能描述，用于说明执行器的用途和特性
// 返回:
//
//	string: 执行器描述信息
func (e *BaseExecutor) GetDescription() string {
	return e.description
}

// FunctionExecutor 函数执行器
// 允许直接执行自定义函数
type FunctionExecutor struct {
	*BaseExecutor
	fn func(ctx context.Context, params interface{}) error
}

// NewFunctionExecutor 创建函数执行器实例
// 允许用户直接传入自定义函数作为任务执行逻辑，提供最大的灵活性
// 参数:
//
//	name: 执行器名称，用于标识和日志记录
//	fn: 要执行的函数，接收上下文和参数，返回执行结果
//
// 返回:
//
//	*FunctionExecutor: 初始化的函数执行器实例
func NewFunctionExecutor(name string, fn func(ctx context.Context, params interface{}) error) *FunctionExecutor {
	return &FunctionExecutor{
		BaseExecutor: NewBaseExecutor(name, "Function executor"), // 基础执行器
		fn:           fn,                                         // 用户自定义函数
	}
}

// Execute 执行用户自定义函数
// 调用用户提供的函数来完成具体的任务执行逻辑
// 参数:
//
//	ctx: 上下文，用于控制超时和取消操作
//	params: 任务参数，传递给用户函数
//
// 返回:
//
//	error: 执行错误，nil表示成功
func (e *FunctionExecutor) Execute(ctx context.Context, params interface{}) error {
	return e.fn(ctx, params)
}
