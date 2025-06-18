package assertion

import (
	"fmt"
	"gohub/internal/gateway/core"
	"path"
	"regexp"
	"strings"
)

// PathMatchType 路径匹配类型
type PathMatchType string

const (
	ExactPathMatch  PathMatchType = "exact"  // 精确匹配
	PrefixPathMatch PathMatchType = "prefix" // 前缀匹配
	RegexPathMatch  PathMatchType = "regex"  // 正则匹配
	ParamPathMatch  PathMatchType = "param"  // 参数匹配
)

// PathAsserter 路径断言器
// 根据HTTP请求路径进行断言，支持多种匹配模式
type PathAsserter struct {
	BaseAssertion

	// 匹配类型
	MatchType PathMatchType

	// 正则表达式对象（用于RegexPathMatch）
	Regexp *regexp.Regexp

	// 路径参数定义（用于ParamPathMatch）
	ParamNames   []string // 参数名称列表
	PatternParts []string // 模式分段
	IsParamPart  []bool   // 标记每个分段是否为参数
}

// PathAsserterFromConfig 从配置创建路径断言器
func PathAsserterFromConfig(config AssertionConfig, operator ComparisonOperator) (Assertion, error) {
	// 获取路径匹配模式
	pattern := strings.ToLower(strings.TrimSpace(config.Pattern))
	if pattern == "" {
		// 根据操作符推断匹配模式
		switch operator {
		case Equal:
			pattern = "exact"
		case StartsWith:
			pattern = "prefix"
		case Matches:
			pattern = "regex"
		default:
			pattern = "exact"
		}
	}

	// 根据匹配模式创建不同的路径断言
	switch pattern {
	case "exact":
		return &PathAsserter{
			BaseAssertion: BaseAssertion{
				Type:          PathAssertion,
				FieldName:     "path",
				ExpectedValue: config.Value,
				Operator:      Equal,
				CaseSensitive: true,
				Description:   config.Description,
				Config:        config,
			},
			MatchType: ExactPathMatch,
		}, nil

	case "prefix":
		return &PathAsserter{
			BaseAssertion: BaseAssertion{
				Type:          PathAssertion,
				FieldName:     "path",
				ExpectedValue: config.Value,
				Operator:      StartsWith,
				CaseSensitive: true,
				Description:   config.Description,
				Config:        config,
			},
			MatchType: PrefixPathMatch,
		}, nil

	case "regex":
		re, err := regexp.Compile(config.Value)
		if err != nil {
			return nil, fmt.Errorf("创建正则路径断言失败: %w", err)
		}
		return &PathAsserter{
			BaseAssertion: BaseAssertion{
				Type:          PathAssertion,
				FieldName:     "path",
				ExpectedValue: config.Value,
				Operator:      Matches,
				CaseSensitive: true,
				Description:   config.Description,
				Config:        config,
			},
			MatchType: RegexPathMatch,
			Regexp:    re,
		}, nil

	case "param", "parameter":
		// 解析路径模式，提取参数名
		patternParts := strings.Split(strings.Trim(config.Value, "/"), "/")
		paramNames := make([]string, 0)
		isParamPart := make([]bool, len(patternParts))

		for i, part := range patternParts {
			if strings.HasPrefix(part, ":") {
				paramName := strings.TrimPrefix(part, ":")
				paramNames = append(paramNames, paramName)
				isParamPart[i] = true
			}
		}

		return &PathAsserter{
			BaseAssertion: BaseAssertion{
				Type:          PathAssertion,
				FieldName:     "path",
				ExpectedValue: config.Value,
				Operator:      Equal, // 这里不是真正用于比较的操作符
				CaseSensitive: true,
				Description:   config.Description,
				Config:        config,
			},
			MatchType:    ParamPathMatch,
			ParamNames:   paramNames,
			PatternParts: patternParts,
			IsParamPart:  isParamPart,
		}, nil

	default:
		return nil, fmt.Errorf("不支持的路径匹配模式: %s", pattern)
	}
}

// GetPathParams 提取路径参数值
// 从请求路径中提取参数值，用于ParamPathMatch
func (a *PathAsserter) GetPathParams(reqPath string) map[string]string {
	if a.MatchType != ParamPathMatch {
		return nil
	}

	pathParts := strings.Split(strings.Trim(reqPath, "/"), "/")
	// 段数必须相同
	if len(pathParts) != len(a.PatternParts) {
		return nil
	}

	// 逐段匹配
	pathParams := make(map[string]string)
	paramIndex := 0

	for i, patternPart := range a.PatternParts {
		if a.IsParamPart[i] {
			// 参数段，存储值
			paramName := a.ParamNames[paramIndex]
			pathParams[paramName] = pathParts[i]
			paramIndex++
		} else if patternPart != pathParts[i] {
			// 非参数段，必须完全匹配
			return nil
		}
	}

	return pathParams
}

// cleanPath 清理路径
// 标准化URL路径，解析相对路径段
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

// Evaluate 实现Assertion接口
func (a *PathAsserter) Evaluate(ctx *core.Context) (bool, error) {
	// 获取请求路径
	reqPath := ctx.Request.URL.Path

	switch a.MatchType {
	case ExactPathMatch:
		// 精确匹配：路径必须完全相同
		return reqPath == a.ExpectedValue, nil

	case PrefixPathMatch:
		// 前缀匹配：路径必须以前缀开头
		prefix := cleanPath(a.ExpectedValue)

		// 处理特殊的 /** 通配符（匹配任意深度的路径）
		if strings.HasSuffix(prefix, "/**") {
			// 去掉末尾的 /**
			basePrefix := strings.TrimSuffix(prefix, "/**")
			// 去掉末尾的 /（如果有）
			basePrefix = strings.TrimSuffix(basePrefix, "/")
			// 检查请求路径是否以基础前缀开头
			return strings.HasPrefix(cleanPath(reqPath), basePrefix), nil
		}

		// 简单前缀匹配（无通配符）
		if !strings.Contains(prefix, "*") {
			return strings.HasPrefix(cleanPath(reqPath), prefix), nil
		}

		// 处理通配符情况
		// 1. 如果只是末尾的通配符 (如 /api/v1/*)
		if strings.HasSuffix(prefix, "*") && strings.Count(prefix, "*") == 1 {
			// 去掉末尾的 *
			prefix = strings.TrimSuffix(prefix, "*")
			// 去掉末尾的 /（如果有）
			prefix = strings.TrimSuffix(prefix, "/")
			return strings.HasPrefix(cleanPath(reqPath), prefix), nil
		}

		// 2. 处理复杂的通配符模式
		// 支持的模式示例:
		// - /api/*/users     匹配 /api/v1/users, /api/v2/users 等
		// - /api/*/users/*   匹配 /api/v1/users/123, /api/v2/users/profile 等
		// - /api/v*/users    匹配 /api/v1/users, /api/v2/users 等
		// - /*/users/*       匹配 /admin/users/123, /api/users/profile 等
		// - /api/**          匹配 /api 及其下所有子路径 (特殊处理)
		// - /api/**/users    匹配 /api/v1/users, /api/v1/v2/users 等 (** 被处理为 *)

		// 将模式转换为正则表达式
		regexPattern := "^"

		// 处理路径中的连续双星号 (**) 为单个星号 (*)
		processedPrefix := prefix
		if strings.Contains(processedPrefix, "**") && !strings.HasSuffix(processedPrefix, "/**") {
			// 将中间的 ** 替换为 *
			processedPrefix = strings.Replace(processedPrefix, "**", "*", -1)
		}

		parts := strings.Split(strings.Trim(processedPrefix, "/"), "/")
		for _, part := range parts {
			if part == "*" {
				// 完整通配符：匹配整个路径段
				regexPattern += "/[^/]*"
			} else if strings.Contains(part, "*") {
				// 部分通配符：将 * 转换为正则表达式的 .*
				regexEscaped := regexp.QuoteMeta(part)
				regexPattern += "/" + strings.Replace(regexEscaped, "\\*", ".*", -1)
			} else {
				// 普通路径段：精确匹配
				regexPattern += "/" + regexp.QuoteMeta(part)
			}
		}

		// 如果原始模式以 * 结尾，则添加匹配后续所有内容的模式
		// 否则也允许匹配后续内容，但需要以 / 开头
		regexPattern += "(/.*)?$"

		// 编译并匹配正则表达式
		re, err := regexp.Compile(regexPattern)
		if err != nil {
			return false, fmt.Errorf("编译路径匹配正则表达式失败: %w", err)
		}

		return re.MatchString(cleanPath(reqPath)), nil

	case RegexPathMatch:
		// 正则匹配：路径必须匹配正则表达式
		if a.Regexp == nil {
			var err error
			a.Regexp, err = regexp.Compile(a.ExpectedValue)
			if err != nil {
				return false, err
			}
		}
		return a.Regexp.MatchString(reqPath), nil

	case ParamPathMatch:
		// 参数匹配：提取路径参数
		params := a.GetPathParams(reqPath)
		// 如果能提取参数（非nil），则匹配成功
		matched := params != nil && len(params) > 0

		// 存储提取的参数到请求上下文
		if matched {
			// 将参数存储到core.Context中
			ctx.Set("path_params", params)
		}

		return matched, nil

	default:
		return false, nil
	}
}

// GetDescription 获取断言描述
func (a *PathAsserter) GetDescription() string {
	if a.Description != "" {
		return a.Description
	}

	var matchTypeStr string
	switch a.MatchType {
	case ExactPathMatch:
		matchTypeStr = "精确匹配"
	case PrefixPathMatch:
		matchTypeStr = "前缀匹配"
	case RegexPathMatch:
		matchTypeStr = "正则匹配"
	case ParamPathMatch:
		matchTypeStr = "参数匹配"
	}

	return "路径" + matchTypeStr + ": " + a.ExpectedValue
}
