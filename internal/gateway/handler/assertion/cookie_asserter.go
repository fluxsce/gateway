package assertion

import (
	"fmt"
	"gateway/internal/gateway/core"
)

// CookieAsserter Cookie断言器
// 根据HTTP请求的Cookie值进行断言
type CookieAsserter struct {
	BaseAssertion
}

// CookieAsserterFromConfig 从配置创建Cookie断言器
func CookieAsserterFromConfig(config AssertionConfig, operator ComparisonOperator) (Assertion, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("Cookie断言必须指定Cookie名称")
	}

	return &CookieAsserter{
		BaseAssertion: BaseAssertion{
			Type:          CookieAssertion,
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
func (a *CookieAsserter) Evaluate(ctx *core.Context) (bool, error) {
	// 获取指定名称的Cookie值
	cookie, err := ctx.Request.Cookie(a.FieldName)
	if err != nil {
		// 如果Cookie不存在，根据操作符决定结果
		if a.Operator == NotExists {
			return true, nil
		}
		return false, nil
	}

	// 应用比较规则
	return a.compare(cookie.Value, a.ExpectedValue), nil
}
