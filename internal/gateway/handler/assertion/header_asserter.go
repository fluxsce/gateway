package assertion

import (
	"fmt"
	"gateway/internal/gateway/core"
)

// HeaderAsserter HTTP头部断言器
// 根据HTTP请求头部字段进行断言
type HeaderAsserter struct {
	BaseAssertion
}

// HeaderAsserterFromConfig 从配置创建HTTP头部断言器
func HeaderAsserterFromConfig(config AssertionConfig, operator ComparisonOperator) (Assertion, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("头部断言必须指定头部名称")
	}

	return &HeaderAsserter{
		BaseAssertion: BaseAssertion{
			Type:          HeaderAssertion,
			FieldName:     config.Name,
			ExpectedValue: config.Value,
			Operator:      operator,
			CaseSensitive: config.CaseSensitive,
			Description:   config.Description,
			Config:        config,
		},
	}, nil
}

// Evaluate 实现Assertion接口
func (a *HeaderAsserter) Evaluate(ctx *core.Context) (bool, error) {
	// 获取头部值
	headerValue := ctx.Request.Header.Get(a.FieldName)

	// 应用比较规则
	return a.compare(headerValue, a.ExpectedValue), nil
}
