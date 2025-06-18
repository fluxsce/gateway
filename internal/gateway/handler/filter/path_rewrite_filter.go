package filter

import (
	"fmt"
	"regexp"
	"strings"

	"gohub/internal/gateway/core"
)

// PathRewriteMode 路径重写模式
type PathRewriteMode string

const (
	// SimpleReplace 简单替换模式
	// 直接替换路径中的特定部分
	SimpleReplace PathRewriteMode = "simple"

	// RegexReplace 正则替换模式
	// 使用正则表达式匹配和替换路径
	RegexReplace PathRewriteMode = "regex"
)

// PathRewriteFilter 路径重写过滤器
// 用于在请求转发前修改请求路径
type PathRewriteFilter struct {
	BaseFilter

	// 重写模式
	Mode PathRewriteMode

	// 查找内容（根据模式可以是字符串或正则表达式）
	From string

	// 替换内容
	To string

	// 编译后的正则表达式（当Mode为RegexReplace时使用）
	regex *regexp.Regexp
}

// NewPathRewriteFilter 创建路径重写过滤器
// 参数:
// - name: 过滤器名称
// - from: 查找内容（字符串或正则表达式）
// - to: 替换内容
// - mode: 重写模式
// - priority: 优先级
// 返回值:
// - *PathRewriteFilter: 创建的过滤器
// - error: 创建过程中的错误
func NewPathRewriteFilter(name, from, to string, mode PathRewriteMode, priority int) (*PathRewriteFilter, error) {
	baseFilter := NewBaseFilter(URLFilterType, PostRouting, priority, true, name)
	filter := &PathRewriteFilter{
		BaseFilter: *baseFilter,
		Mode:       mode,
		From:       from,
		To:         to,
	}

	// 如果是正则模式，预编译正则表达式
	if mode == RegexReplace {
		regex, err := regexp.Compile(from)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		filter.regex = regex
	}

	return filter, nil
}

// Apply 实现Filter接口
// 根据配置的规则重写请求路径
func (f *PathRewriteFilter) Apply(ctx *core.Context) error {
	if ctx.Request == nil {
		return fmt.Errorf("request is nil")
	}

	// 获取当前路径
	path := ctx.Request.URL.Path

	// 根据不同模式执行路径重写
	switch f.Mode {
	case SimpleReplace:
		// 简单字符串替换
		path = strings.Replace(path, f.From, f.To, -1)

	case RegexReplace:
		// 正则表达式替换
		if f.regex != nil {
			path = f.regex.ReplaceAllString(path, f.To)
		}
	}

	// 设置新路径
	ctx.Request.URL.Path = path

	// 记录重写操作到上下文（可选，用于调试和监控）
	ctx.Set("path_rewritten", true)
	ctx.Set("original_path", ctx.Request.URL.Path)
	ctx.Set("rewrite_filter", f.Name)

	return nil
}

// NewSimplePathRewriteFilter 创建简单路径重写过滤器
// 简单封装了 NewPathRewriteFilter 函数，使用 SimpleReplace 模式
func NewSimplePathRewriteFilter(name, from, to string, priority int) *PathRewriteFilter {
	filter, _ := NewPathRewriteFilter(name, from, to, SimpleReplace, priority)
	return filter
}

// NewRegexPathRewriteFilter 创建正则路径重写过滤器
// 简单封装了 NewPathRewriteFilter 函数，使用 RegexReplace 模式
func NewRegexPathRewriteFilter(name, pattern, replacement string, priority int) (*PathRewriteFilter, error) {
	return NewPathRewriteFilter(name, pattern, replacement, RegexReplace, priority)
}
