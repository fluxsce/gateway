package assertion

import (
	"fmt"
	"gateway/internal/gateway/core"
)

// QueryParamAsserter 查询参数断言器
// 根据URL查询参数进行断言
type QueryParamAsserter struct {
	BaseAssertion
}

// QueryParamAsserterFromConfig 从配置创建查询参数断言器
func QueryParamAsserterFromConfig(config AssertionConfig, operator ComparisonOperator) (Assertion, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("查询参数断言必须指定参数名称")
	}

	return &QueryParamAsserter{
		BaseAssertion: BaseAssertion{
			Type:          QueryParamAssertion,
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
func (a *QueryParamAsserter) Evaluate(ctx *core.Context) (bool, error) {
	// 获取查询参数值
	paramValue := ctx.Request.URL.Query().Get(a.FieldName)

	// 检查参数是否存在
	if paramValue == "" {
		// 参数不存在
		if a.Operator == NotExists {
			return true, nil
		}
		return false, nil
	}

	// 应用比较规则
	return a.compare(paramValue, a.ExpectedValue), nil
}
