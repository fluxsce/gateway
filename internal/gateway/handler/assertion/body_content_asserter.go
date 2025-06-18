package assertion

import (
	"gohub/internal/gateway/core"
	"io"
)

// BodyContentAsserter 请求体内容断言器
// 根据HTTP请求体内容进行断言
type BodyContentAsserter struct {
	BaseAssertion
}

// BodyContentAsserterFromConfig 从配置创建请求体内容断言器
func BodyContentAsserterFromConfig(config AssertionConfig, operator ComparisonOperator) (Assertion, error) {
	return &BodyContentAsserter{
		BaseAssertion: BaseAssertion{
			Type:          BodyContentAssertion,
			FieldName:     "body",
			ExpectedValue: config.Value,
			Operator:      operator,
			CaseSensitive: config.CaseSensitive,
			Description:   config.Description,
			Config:        config,
		},
	}, nil
}

// Evaluate 实现Assertion接口
func (a *BodyContentAsserter) Evaluate(ctx *core.Context) (bool, error) {
	// 读取请求体内容
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return false, err
	}

	// 应用比较规则
	return a.compare(string(body), a.ExpectedValue), nil
}
