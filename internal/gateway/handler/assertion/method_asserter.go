package assertion

import (
	"gohub/internal/gateway/core"
	"strings"
)

// MethodAsserter HTTP方法断言器
// 根据HTTP请求方法进行断言
type MethodAsserter struct {
	BaseAssertion
}

// MethodAsserterFromConfig 从配置创建HTTP方法断言器
func MethodAsserterFromConfig(config AssertionConfig, operator ComparisonOperator) (Assertion, error) {
	return &MethodAsserter{
		BaseAssertion: BaseAssertion{
			Type:          MethodAssertion,
			FieldName:     "method",
			ExpectedValue: config.Value,
			Operator:      operator,
			CaseSensitive: false,
			Description:   config.Description,
			Config:        config,
		},
	}, nil
}

// Evaluate 实现Assertion接口
func (a *MethodAsserter) Evaluate(ctx *core.Context) (bool, error) {
	// 获取HTTP方法
	method := ctx.Request.Method

	// 应用比较规则，方法比较通常不区分大小写
	return a.compare(strings.ToUpper(method), strings.ToUpper(a.ExpectedValue)), nil
}
