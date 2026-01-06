package assertion

import (
	"bytes"
	"fmt"
	"gateway/internal/gateway/core"
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
	// 获取请求体内容
	var bodyBytes []byte
	var err error

	// 优先从上下文获取已缓存的请求体（避免重复读取）
	if bodyData, exists := ctx.Get("request_body"); exists {
		if cachedBody, ok := bodyData.([]byte); ok {
			bodyBytes = cachedBody
		}
	}

	// 如果没有缓存，从 Request.Body 读取
	if bodyBytes == nil {
		if ctx.Request.Body == nil {
			// 请求体不存在，使用空字符串
			bodyBytes = []byte{}
		} else {
			// 读取请求体内容
			// 注意：io.ReadAll 会读取 Body 直到 EOF，但不会关闭 Body
			// 原始的 Body 的生命周期由 HTTP 服务器管理，我们不需要手动关闭
			bodyBytes, err = io.ReadAll(ctx.Request.Body)
			if err != nil {
				// 读取失败，根据操作符处理
				// 对于 Exists/NotExists，读取失败视为不存在
				if a.Operator == Exists {
					return false, nil
				}
				if a.Operator == NotExists {
					return true, nil
				}
				// 对于其他操作符，返回错误
				return false, fmt.Errorf("读取请求体失败: %w", err)
			}

			// 重置请求体，以便后续使用
			// io.NopCloser 包装的 bytes.Reader 不需要关闭，不会造成资源泄露
			// bytes.Reader 是内存中的读取器，不持有任何需要释放的资源
			ctx.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	// 将请求体转换为字符串
	bodyStr := string(bodyBytes)

	// 对于 Exists 和 NotExists 操作符，使用 compare 方法处理
	// compare 方法会根据操作符正确处理：Exists 检查 actual != ""，NotExists 检查 actual == ""
	// 对于其他操作符，也会正确应用比较规则
	return a.compare(bodyStr, a.ExpectedValue), nil
}
