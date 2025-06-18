package filter

import (
	"fmt"
	"strings"

	"gohub/internal/gateway/core"
)

// StripPrefixFilter 前缀剥离过滤器
// 用于移除请求路径中的指定前缀
type StripPrefixFilter struct {
	BaseFilter

	// 要剥离的前缀
	Prefix string
}

// NewStripPrefixFilter 创建前缀剥离过滤器
// 参数:
// - name: 过滤器名称
// - prefix: 要剥离的前缀
// - priority: 优先级
// 返回值:
// - *StripPrefixFilter: 创建的过滤器
func NewStripPrefixFilter(name, prefix string, priority int) *StripPrefixFilter {
	baseFilter := NewBaseFilter(URLFilterType, PostRouting, priority, true, name)
	return &StripPrefixFilter{
		BaseFilter: *baseFilter,
		Prefix:     prefix,
	}
}

// Apply 实现Filter接口
// 从请求路径中剥离指定前缀
func (f *StripPrefixFilter) Apply(ctx *core.Context) error {
	if ctx.Request == nil {
		return fmt.Errorf("request is nil")
	}

	// 获取当前路径
	path := ctx.Request.URL.Path

	// 检查路径是否以前缀开始
	if strings.HasPrefix(path, f.Prefix) {
		// 剥离前缀
		newPath := path[len(f.Prefix):]
		if len(newPath) == 0 || newPath[0] != '/' {
			newPath = "/" + newPath
		}

		// 设置新路径
		ctx.Request.URL.Path = newPath

		// 记录操作到上下文（可选，用于调试和监控）
		ctx.Set("prefix_stripped", true)
		ctx.Set("original_path", path)
		ctx.Set("strip_prefix_filter", f.Name)
	}

	return nil
}
