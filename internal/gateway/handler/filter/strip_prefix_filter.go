package filter

import (
	"fmt"
	"path"
	"strings"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
)

// StripPrefixFilter 前缀剥离过滤器
// 用于移除请求路径中的指定前缀
type StripPrefixFilter struct {
	BaseFilter

	// 要剥离的前缀，已规范成以 / 开头且不是根路径
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
	baseFilter := NewBaseFilter(StripFilterType, PostRouting, priority, true, name)
	return &StripPrefixFilter{
		BaseFilter: *baseFilter,
		Prefix:     prefix,
	}
}

// Apply 实现Filter接口
// 从请求路径中按路径段边界剥离指定前缀
func (f *StripPrefixFilter) Apply(ctx *core.Context) error {
	if ctx.Request == nil || ctx.Request.URL == nil {
		return fmt.Errorf("request is nil")
	}

	requestPath := ctx.Request.URL.Path
	newPath, stripped := stripConfiguredPrefix(requestPath, f.Prefix)
	if !stripped {
		return nil
	}

	ctx.Request.URL.Path = newPath
	// Path 已改写，清掉原始转义路径，避免 EscapedPath 仍返回旧值
	ctx.Request.URL.RawPath = ""

	ctx.Set(constants.ContextKeyPrefixStripped, true)
	ctx.Set("original_path", requestPath)
	ctx.Set("strip_prefix_filter", f.Name)

	return nil
}

// stripConfiguredPrefix 按路径段边界去掉前缀。
// prefix 须已规范化（以 / 开头且不是根路径）。/api 可剥 /api/users，不会误剥 /apix。
func stripConfiguredPrefix(requestPath, prefix string) (string, bool) {
	if prefix == "" || prefix == "/" {
		return requestPath, false
	}
	if requestPath == "" {
		requestPath = "/"
	}
	if !strings.HasPrefix(requestPath, prefix) {
		return requestPath, false
	}
	if len(requestPath) == len(prefix) {
		return "/", true
	}
	if requestPath[len(prefix)] != '/' {
		return requestPath, false
	}
	return requestPath[len(prefix):], true
}

// StripPrefixFilterFromConfig 从配置创建前缀剥离过滤器。
// 与 header、query-param 相同，在过滤器内读取嵌套配置，不依赖加载器摊平。
func StripPrefixFilterFromConfig(config FilterConfig) (Filter, error) {
	prefix, err := parseStripPrefix(config.Config)
	if err != nil {
		return nil, err
	}

	order := config.Order
	if order <= 0 {
		order = 100
	}

	filter := NewStripPrefixFilter(config.Name, prefix, order)
	filter.Action = getFilterActionFromConfig(config)
	filter.BaseFilter.originalConfig = config

	return filter, nil
}

// parseStripPrefix 读取并规范要剥离的前缀。
// 页面保存为 stripConfig.prefix；没有该对象时兼容顶层 prefix。
func parseStripPrefix(config map[string]interface{}) (string, error) {
	stripConfig := config
	if config != nil {
		if nested, ok := config["stripConfig"].(map[string]interface{}); ok {
			stripConfig = nested
		}
	}

	prefix := ""
	if stripConfig != nil {
		prefix, _ = stripConfig["prefix"].(string)
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "", fmt.Errorf("前缀剥离过滤器的 prefix 不能为空")
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = path.Clean(prefix)
	if prefix == "/" || prefix == "." {
		return "", fmt.Errorf("前缀剥离过滤器的 prefix 不能为根路径")
	}
	return prefix, nil
}
