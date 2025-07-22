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
	baseFilter := NewBaseFilter(RewriteFilterType, PostRouting, priority, true, name)
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

// parseRewriteConfig 解析路径重写配置
func parseRewriteConfig(config map[string]interface{}) (from, to, mode string, err error) {
	// 设置默认值
	mode = "simple"
	
	// 优先支持前端驼峰命名格式
	if rewriteConfig, ok := config["rewriteConfig"].(map[string]interface{}); ok {
		// 从rewriteConfig中提取参数
		if f, ok := rewriteConfig["from"].(string); ok {
			from = f
		}
		if t, ok := rewriteConfig["to"].(string); ok {
			to = t
		}
		if m, ok := rewriteConfig["mode"].(string); ok {
			mode = strings.ToLower(m)
		}
	} else {
		// 兼容旧的直接配置格式
		if f, ok := config["from"].(string); ok {
			from = f
		}
		if t, ok := config["to"].(string); ok {
			to = t
		}
		if m, ok := config["mode"].(string); ok {
			mode = strings.ToLower(m)
		}
	}
	
	// 参数验证
	if from == "" {
		return "", "", "", fmt.Errorf("重写规则的from参数不能为空")
	}
	if to == "" {
		return "", "", "", fmt.Errorf("重写规则的to参数不能为空")
	}
	
	// 验证模式的有效性
	if mode != "simple" && mode != "regex" {
		return "", "", "", fmt.Errorf("无效的重写模式: %s，支持的模式: simple, regex", mode)
	}
	
	return from, to, mode, nil
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

// PathRewriteFilterFromConfig 从配置创建路径重写过滤器
// 支持前端传递的驼峰命名配置格式：
// {
//   "rewriteConfig": {
//     "mode": "simple|regex",     // 重写模式
//     "from": "string",           // 查找内容
//     "to": "string"              // 替换内容
//   }
// }
func PathRewriteFilterFromConfig(config FilterConfig) (Filter, error) {
	// 解析重写配置
	from, to, mode, parseErr := parseRewriteConfig(config.Config)
	if parseErr != nil {
		return nil, parseErr
	}

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	var filter *PathRewriteFilter
	var createErr error

	if mode == "regex" {
		filter, createErr = NewRegexPathRewriteFilter(config.Name, from, to, order)
		if createErr != nil {
			return nil, fmt.Errorf("创建正则路径重写过滤器失败: %w", createErr)
		}
	} else {
		filter = NewSimplePathRewriteFilter(config.Name, from, to, order)
	}

	// 存储原始配置
	filter.BaseFilter.originalConfig = config

	return filter, nil
}
