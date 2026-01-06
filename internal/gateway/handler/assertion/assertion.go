package assertion

import (
	"fmt"
	"gateway/internal/gateway/core"
	"regexp"
	"strings"
)

// AssertionType 断言类型
// 定义了不同的断言类型，用于对请求进行更精细的匹配
type AssertionType string

const (
	// HeaderAssertion HTTP头部断言
	// 根据请求头部字段进行断言
	HeaderAssertion AssertionType = "header"

	// QueryParamAssertion 查询参数断言
	// 根据URL查询参数进行断言
	QueryParamAssertion AssertionType = "query-param"

	// BodyContentAssertion 报文内容断言
	// 根据请求体内容进行断言
	BodyContentAssertion AssertionType = "body-content"

	// MethodAssertion HTTP方法断言
	// 根据HTTP方法进行断言
	MethodAssertion AssertionType = "method"

	// CookieAssertion Cookie断言
	// 根据Cookie值进行断言
	CookieAssertion AssertionType = "cookie"

	// IPAssertion IP地址断言
	// 根据客户端IP地址进行断言
	IPAssertion AssertionType = "ip"

	// PathAssertion 路径断言
	// 根据请求路径进行断言
	PathAssertion AssertionType = "path"
)

// ComparisonOperator 比较操作符
// 定义了不同的比较操作，用于断言规则
type ComparisonOperator string

const (
	// Equal 等于
	Equal ComparisonOperator = "equal"

	// NotEqual 不等于
	NotEqual ComparisonOperator = "not-equal"

	// Contains 包含
	Contains ComparisonOperator = "contains"

	// NotContains 不包含
	NotContains ComparisonOperator = "not-contains"

	// StartsWith 以...开头
	StartsWith ComparisonOperator = "starts-with"

	// EndsWith 以...结尾
	EndsWith ComparisonOperator = "ends-with"

	// Matches 正则匹配
	Matches ComparisonOperator = "matches"

	// Exists 存在（字段存在即可，不关心值）
	Exists ComparisonOperator = "exists"

	// NotExists 不存在
	NotExists ComparisonOperator = "not-exists"
)

// Assertion 断言规则接口
// 所有类型的断言规则都实现此接口
type Assertion interface {
	// Evaluate 评估请求是否符合断言规则
	// 参数:
	// - ctx: 请求上下文（包含HTTP请求和其他上下文信息）
	// 返回值:
	// - bool: 是否通过断言
	// - error: 评估过程中的错误
	Evaluate(ctx *core.Context) (bool, error)

	// GetType 获取断言类型
	// 返回值:
	// - AssertionType: 断言类型
	GetType() AssertionType

	// GetDescription 获取断言描述
	// 返回值:
	// - string: 人类可读的断言规则描述
	GetDescription() string

	// GetConfig 获取断言配置
	// 返回值:
	// - AssertionConfig: 断言的配置信息
	GetConfig() AssertionConfig
}

// BaseAssertion 断言基础结构
// 包含所有断言类型共有的属性
type BaseAssertion struct {
	// 断言类型
	Type AssertionType

	// 比较操作符
	Operator ComparisonOperator

	// 字段名称
	FieldName string

	// 期望值
	ExpectedValue string

	// 是否区分大小写
	CaseSensitive bool

	// 断言描述
	Description string

	// 原始配置（用于GetConfig方法）
	Config AssertionConfig
}

// GetType 获取断言类型
func (b *BaseAssertion) GetType() AssertionType {
	return b.Type
}

// GetDescription 获取断言描述
func (b *BaseAssertion) GetDescription() string {
	if b.Description != "" {
		return b.Description
	}

	var op string
	switch b.Operator {
	case Equal:
		op = "等于"
	case NotEqual:
		op = "不等于"
	case Contains:
		op = "包含"
	case NotContains:
		op = "不包含"
	case StartsWith:
		op = "以...开头"
	case EndsWith:
		op = "以...结尾"
	case Matches:
		op = "匹配正则"
	case Exists:
		op = "存在"
	case NotExists:
		op = "不存在"
	}

	typeStr := ""
	switch b.Type {
	case HeaderAssertion:
		typeStr = "头部"
	case QueryParamAssertion:
		typeStr = "查询参数"
	case BodyContentAssertion:
		typeStr = "报文内容"
	case MethodAssertion:
		typeStr = "HTTP方法"
	case CookieAssertion:
		typeStr = "Cookie"
	case IPAssertion:
		typeStr = "IP地址"
	case PathAssertion:
		typeStr = "路径"
	}

	if b.Operator == Exists || b.Operator == NotExists {
		return typeStr + "字段 " + b.FieldName + " " + op
	}
	return typeStr + "字段 " + b.FieldName + " " + op + " " + b.ExpectedValue
}

// compare 通用比较函数
// 根据比较操作符比较两个字符串
func (b *BaseAssertion) compare(actual string, expected string) bool {
	// 如果不区分大小写，转换为小写比较
	if !b.CaseSensitive {
		actual = strings.ToLower(actual)
		expected = strings.ToLower(expected)
	}

	switch b.Operator {
	case Equal:
		return actual == expected
	case NotEqual:
		return actual != expected
	case Contains:
		// 检查 actual 是否包含 expected
		return strings.Contains(actual, expected)
	case NotContains:
		// 检查 actual 是否不包含 expected
		return !strings.Contains(actual, expected)
	case StartsWith:
		return strings.HasPrefix(actual, expected)
	case EndsWith:
		return strings.HasSuffix(actual, expected)
	case Matches:
		// expected 是正则表达式模式，actual 是要匹配的字符串
		matched, _ := regexp.MatchString(expected, actual)
		return matched
	case Exists:
		return actual != ""
	case NotExists:
		return actual == ""
	default:
		return false
	}
}

// AssertionGroupConfig 断言组配置
// 用于配置文件中的断言组定义
type AssertionGroupConfig struct {
	// 断言组ID
	ID string `json:"id" yaml:"id" mapstructure:"id"`

	// 断言配置列表 - 用于配置文件和序列化
	AssertionConfigs []AssertionConfig `json:"assertions" yaml:"assertions" mapstructure:"assertions"`

	// 逻辑关系: true=AND（所有断言都必须满足）, false=OR（任一断言满足即可）
	AllRequired bool `json:"all_required" yaml:"all_required" mapstructure:"all_required"`

	// 断言组描述
	Description string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`
}

// NewAssertionGroupConfig 创建断言组配置
func NewAssertionGroupConfig(id string, allRequired bool) *AssertionGroupConfig {
	return &AssertionGroupConfig{
		ID:               id,
		AssertionConfigs: make([]AssertionConfig, 0),
		AllRequired:      allRequired,
	}
}

// NewAssertionGroupConfigFromConfig 从配置参数创建断言组配置
func NewAssertionGroupConfigFromConfig(id string, allRequired bool, configs []AssertionConfig, description string) *AssertionGroupConfig {
	return &AssertionGroupConfig{
		ID:               id,
		AssertionConfigs: configs,
		AllRequired:      allRequired,
		Description:      description,
	}
}

// AddAssertionConfig 添加断言配置到组
func (g *AssertionGroupConfig) AddAssertionConfig(config AssertionConfig) {
	g.AssertionConfigs = append(g.AssertionConfigs, config)
}

// GetDescription 获取断言组描述
func (g *AssertionGroupConfig) GetDescription() string {
	if g.Description != "" {
		return g.Description
	}

	assertionCount := len(g.AssertionConfigs)
	if assertionCount == 0 {
		return "空断言组（默认通过）"
	}

	logic := "任一满足"
	if g.AllRequired {
		logic = "全部满足"
	}

	return fmt.Sprintf("断言组（%s）: %d个断言", logic, assertionCount)
}

// AssertionConfig 断言配置结构
// 用于从配置文件或API中创建断言
type AssertionConfig struct {
	// 断言ID
	ID string `yaml:"id" json:"id" mapstructure:"id"`

	// 断言类型：path, header, query, method, cookie, ip, body-content
	Type string `yaml:"type" json:"type" mapstructure:"type"`

	// 断言字段名（如header名、query参数名等）
	Name string `yaml:"name,omitempty" json:"name,omitempty" mapstructure:"name,omitempty"`

	// 期望值
	Value string `yaml:"value,omitempty" json:"value,omitempty" mapstructure:"value,omitempty"`

	// 比较操作符：equal, not-equal, contains, not-contains, starts-with, ends-with, matches, exists, not-exists
	Operator string `yaml:"operator" json:"operator" mapstructure:"operator"`

	// 是否区分大小写
	CaseSensitive bool `yaml:"case_sensitive,omitempty" json:"case_sensitive,omitempty" mapstructure:"case_sensitive,omitempty"`

	// 断言描述
	Description string `yaml:"description,omitempty" json:"description,omitempty" mapstructure:"description,omitempty"`

	// 路径匹配模式（仅用于path类型）：exact, prefix, regex, param
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty" mapstructure:"pattern,omitempty"`

	// 扩展配置，用于存储特定类型断言的额外参数
	Config map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty" mapstructure:"config,omitempty"`
}

// GetConfig 获取断言配置
func (b *BaseAssertion) GetConfig() AssertionConfig {
	// 如果有原始配置，直接返回
	if b.Config.Type != "" {
		return b.Config
	}

	// 否则根据当前字段构建配置
	return AssertionConfig{
		ID:            b.FieldName + "-" + string(b.Type), // 默认使用字段名+类型作为ID
		Type:          string(b.Type),
		Name:          b.FieldName,
		Value:         b.ExpectedValue,
		Operator:      string(b.Operator),
		CaseSensitive: b.CaseSensitive,
		Description:   b.Description,
	}
}
