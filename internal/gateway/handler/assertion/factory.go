package assertion

import (
	"fmt"
	"gohub/internal/gateway/core"
	"strings"
)

// AssertionGroup 运行时断言组
// 包含多个断言，可以设置逻辑关系（与/或）
type AssertionGroup struct {
	// 断言列表 - 运行时对象
	Assertions []Assertion

	// 逻辑关系: true=AND（所有断言都必须满足）, false=OR（任一断言满足即可）
	AllRequired bool

	// 断言组描述
	Description string
}

// NewAssertionGroup 创建运行时断言组
func NewAssertionGroup(allRequired bool) *AssertionGroup {
	return &AssertionGroup{
		Assertions:  make([]Assertion, 0),
		AllRequired: allRequired,
	}
}

// NewAssertionGroupFromConfig 从配置对象创建运行时断言组
// 这是一个便捷方法，自动创建工厂并处理从配置到运行时对象的转换
// 参数:
// - config: 断言组配置对象
// 返回值:
// - *AssertionGroup: 创建的运行时断言组
// - error: 创建过程中的错误
func NewAssertionGroupFromConfig(config *AssertionGroupConfig) (*AssertionGroup, error) {
	factory := NewAssertionFactory()
	return factory.CreateAssertionGroup(config)
}

// AddAssertion 添加断言到组
func (g *AssertionGroup) AddAssertion(assertion Assertion) {
	g.Assertions = append(g.Assertions, assertion)
}

// Evaluate 评估断言组
func (g *AssertionGroup) Evaluate(ctx *core.Context) (bool, error) {
	if len(g.Assertions) == 0 {
		// 没有断言，默认通过
		return true, nil
	}

	for _, assertion := range g.Assertions {
		result, err := assertion.Evaluate(ctx)
		if err != nil {
			return false, err
		}

		if g.AllRequired {
			// AND逻辑，任一断言失败则整组失败
			if !result {
				return false, nil
			}
		} else {
			// OR逻辑，任一断言成功则整组成功
			if result {
				return true, nil
			}
		}
	}

	// 对于AND逻辑，所有断言都通过才返回true
	// 对于OR逻辑，所有断言都失败才返回false
	return g.AllRequired, nil
}

// GetDescription 获取断言组描述
func (g *AssertionGroup) GetDescription() string {
	if g.Description != "" {
		return g.Description
	}

	assertionCount := len(g.Assertions)
	if assertionCount == 0 {
		return "空断言组（默认通过）"
	}

	logic := "任一满足"
	if g.AllRequired {
		logic = "全部满足"
	}

	return fmt.Sprintf("断言组（%s）: %d个断言", logic, assertionCount)
}

// AssertionFactory 断言工厂
// 用于根据配置创建各种类型的断言和断言组
type AssertionFactory struct{}

// NewAssertionFactory 创建断言工厂
func NewAssertionFactory() *AssertionFactory {
	return &AssertionFactory{}
}

// CreateAssertion 根据配置创建断言
// 参数:
// - config: 断言配置
// 返回值:
// - Assertion: 创建的断言
// - error: 创建过程中的错误
func (f *AssertionFactory) CreateAssertion(config AssertionConfig) (Assertion, error) {
	if config.Type == "" {
		return nil, fmt.Errorf("断言类型不能为空")
	}

	// 标准化断言类型
	assertionType := f.normalizeAssertionType(config.Type)

	// 解析比较操作符
	operator, err := f.parseOperator(config.Operator)
	if err != nil {
		return nil, fmt.Errorf("无效的比较操作符 '%s': %w", config.Operator, err)
	}

	// 根据断言类型委托给具体的实现类创建
	switch assertionType {
	case PathAssertion:
		return PathAsserterFromConfig(config, operator)
	case HeaderAssertion:
		return HeaderAsserterFromConfig(config, operator)
	case QueryParamAssertion:
		return QueryParamAsserterFromConfig(config, operator)
	case MethodAssertion:
		return MethodAsserterFromConfig(config, operator)
	case CookieAssertion:
		return CookieAsserterFromConfig(config, operator)
	case IPAssertion:
		return IPAsserterFromConfig(config, operator)
	case BodyContentAssertion:
		return BodyContentAsserterFromConfig(config, operator)
	default:
		return nil, fmt.Errorf("不支持的断言类型: %s", config.Type)
	}
}

// CreateAssertionGroup 根据配置创建断言组
// 参数:
// - config: 断言组配置
// 返回值:
// - *AssertionGroup: 创建的运行时断言组（已填充断言对象）
// - error: 创建过程中的错误
func (f *AssertionFactory) CreateAssertionGroup(config *AssertionGroupConfig) (*AssertionGroup, error) {
	// 创建运行时断言组
	group := NewAssertionGroup(config.AllRequired)
	group.Description = config.Description

	// 创建每个断言并添加到组中
	for i, assertionConfig := range config.AssertionConfigs {
		assertion, err := f.CreateAssertion(assertionConfig)
		if err != nil {
			return nil, fmt.Errorf("创建第 %d 个断言失败: %w", i+1, err)
		}
		group.AddAssertion(assertion)
	}

	return group, nil
}

// CreateAssertionGroupFromConfig 从配置参数创建断言组
// 参数:
// - id: 断言组ID
// - allRequired: 是否要求所有断言都满足
// - configs: 断言配置列表
// - description: 断言组描述
// 返回值:
// - *AssertionGroup: 创建的断言组
// - error: 创建过程中的错误
func (f *AssertionFactory) CreateAssertionGroupFromConfig(id string, allRequired bool, configs []AssertionConfig, description string) (*AssertionGroup, error) {
	config := NewAssertionGroupConfigFromConfig(id, allRequired, configs, description)
	return f.CreateAssertionGroup(config)
}

// normalizeAssertionType 标准化断言类型
func (f *AssertionFactory) normalizeAssertionType(assertionType string) AssertionType {
	switch strings.ToLower(strings.TrimSpace(assertionType)) {
	case "path":
		return PathAssertion
	case "header", "headers":
		return HeaderAssertion
	case "query", "query-param", "query_param":
		return QueryParamAssertion
	case "method", "http-method", "http_method":
		return MethodAssertion
	case "cookie", "cookies":
		return CookieAssertion
	case "ip", "client-ip", "client_ip":
		return IPAssertion
	case "body", "body-content", "body_content":
		return BodyContentAssertion
	default:
		return AssertionType(strings.ToLower(assertionType))
	}
}

// parseOperator 解析比较操作符
func (f *AssertionFactory) parseOperator(operator string) (ComparisonOperator, error) {
	switch strings.ToLower(strings.TrimSpace(operator)) {
	case "equal", "eq", "==":
		return Equal, nil
	case "not-equal", "not_equal", "ne", "!=":
		return NotEqual, nil
	case "contains", "contain":
		return Contains, nil
	case "not-contains", "not_contains", "not-contain", "not_contain":
		return NotContains, nil
	case "starts-with", "starts_with", "prefix":
		return StartsWith, nil
	case "ends-with", "ends_with", "suffix":
		return EndsWith, nil
	case "matches", "match", "regex":
		return Matches, nil
	case "exists", "exist":
		return Exists, nil
	case "not-exists", "not_exists", "not-exist", "not_exist":
		return NotExists, nil
	default:
		return "", fmt.Errorf("未知的比较操作符: %s", operator)
	}
}
