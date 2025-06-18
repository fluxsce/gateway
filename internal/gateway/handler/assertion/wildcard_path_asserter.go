package assertion

import (
	"fmt"
	"gohub/internal/gateway/core"
	"strings"
)

// WildcardPathAsserter 通配符路径断言器
// 用于匹配包含通配符(*)的路径模式
type WildcardPathAsserter struct {
	BaseAssertion
	pattern string // 完整的通配符模式
	parts   []part // 解析后的路径部分
}

// part 表示一个路径段
type part struct {
	isWildcard bool   // 是否为通配符
	value      string // 段的值（非通配符时）
}

// parseWildcardPattern 解析通配符模式
// 将路径模式解析为一系列通配符和非通配符部分
func parseWildcardPattern(pattern string) []part {
	// 标准化路径
	pattern = cleanPath(pattern)
	segments := strings.Split(strings.Trim(pattern, "/"), "/")
	parts := make([]part, len(segments))

	for i, segment := range segments {
		if segment == "*" {
			parts[i] = part{isWildcard: true}
		} else {
			parts[i] = part{isWildcard: false, value: segment}
		}
	}

	return parts
}

// Evaluate 实现Assertion接口
func (w *WildcardPathAsserter) Evaluate(ctx *core.Context) (bool, error) {
	// 获取请求路径
	reqPath := cleanPath(ctx.Request.URL.Path)
	segments := strings.Split(strings.Trim(reqPath, "/"), "/")

	// 如果请求路径段数小于模式中的非通配符段数，肯定不匹配
	if len(segments) < countNonWildcards(w.parts) {
		return false, nil
	}

	// 如果没有通配符，且段数不同，则不匹配
	if !hasWildcard(w.parts) && len(segments) != len(w.parts) {
		return false, nil
	}

	return matchWildcardPattern(segments, w.parts), nil
}

// countNonWildcards 计算非通配符段的数量
func countNonWildcards(parts []part) int {
	count := 0
	for _, p := range parts {
		if !p.isWildcard {
			count++
		}
	}
	return count
}

// hasWildcard 检查是否包含通配符
func hasWildcard(parts []part) bool {
	for _, p := range parts {
		if p.isWildcard {
			return true
		}
	}
	return false
}

// matchWildcardPattern 匹配通配符模式
// 判断请求路径段是否匹配通配符模式
func matchWildcardPattern(segments []string, parts []part) bool {
	// 特殊情况：如果模式只有一个通配符，匹配所有路径
	if len(parts) == 1 && parts[0].isWildcard {
		return true
	}

	// 动态规划算法匹配通配符
	// dp[i][j] 表示路径的前i段是否能匹配模式的前j段
	dp := make([][]bool, len(segments)+1)
	for i := range dp {
		dp[i] = make([]bool, len(parts)+1)
	}

	// 空路径匹配空模式
	dp[0][0] = true

	// 空路径匹配只有通配符的模式
	for j := 1; j <= len(parts); j++ {
		if parts[j-1].isWildcard {
			dp[0][j] = dp[0][j-1]
		}
	}

	// 构建DP表
	for i := 1; i <= len(segments); i++ {
		for j := 1; j <= len(parts); j++ {
			if parts[j-1].isWildcard {
				// 通配符可以匹配任何段，也可以跳过
				dp[i][j] = dp[i-1][j] || dp[i][j-1]
			} else if segments[i-1] == parts[j-1].value {
				// 段值相等
				dp[i][j] = dp[i-1][j-1]
			}
		}
	}

	return dp[len(segments)][len(parts)]
}

// GetDescription 获取断言描述
func (w *WildcardPathAsserter) GetDescription() string {
	if w.Description != "" {
		return w.Description
	}
	return fmt.Sprintf("请求路径 通配符匹配: %s", w.pattern)
}

// GetPathParams 实现路径参数获取接口
// 由于通配符匹配不提取命名参数，所以返回空map
func (w *WildcardPathAsserter) GetPathParams(path string) map[string]string {
	return nil
}
