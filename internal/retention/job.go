package retention

import (
	"context"
)

// Job 一类租户级数据的清理执行器。调度器只负责租约和周期，删除或以后转冷都实现此接口。
type Job interface {
	// Name 返回任务名，用于日志。
	Name() string
	// Run 执行一轮清理。应幂等，多节点偶发重入不得误删未到期数据。
	Run(ctx context.Context) error
}

// FuncJob 把函数适配为 Job。
type FuncJob struct {
	name string
	fn   func(context.Context) error
}

// NewFuncJob 创建函数型执行器。fn 为 nil 时 Run 直接返回。
func NewFuncJob(name string, fn func(context.Context) error) *FuncJob {
	return &FuncJob{name: name, fn: fn}
}

// Name 返回任务名。
func (j *FuncJob) Name() string {
	return j.name
}

// Run 调用包装的函数。
func (j *FuncJob) Run(ctx context.Context) error {
	if j == nil || j.fn == nil {
		return nil
	}
	return j.fn(ctx)
}
