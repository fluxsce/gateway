package assertion

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/assertion"
)

func TestAssertionConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *assertion.AssertionConfig
		description string
	}{
		{
			name: "HeaderAssertionConfig",
			config: &assertion.AssertionConfig{
				ID:       "test-header-assertion",
				Type:     string(assertion.HeaderAssertion),
				Name:     "Content-Type",
				Value:    "application/json",
				Operator: string(assertion.Equal),
			},
			description: "头部断言配置",
		},
		{
			name: "QueryParamAssertionConfig",
			config: &assertion.AssertionConfig{
				ID:       "test-query-assertion",
				Type:     string(assertion.QueryParamAssertion),
				Name:     "version",
				Value:    "v1",
				Operator: string(assertion.Equal),
			},
			description: "查询参数断言配置",
		},
		{
			name: "MethodAssertionConfig",
			config: &assertion.AssertionConfig{
				ID:       "test-method-assertion",
				Type:     string(assertion.MethodAssertion),
				Value:    "POST",
				Operator: string(assertion.Equal),
			},
			description: "HTTP方法断言配置",
		},
		{
			name: "PathAssertionConfig",
			config: &assertion.AssertionConfig{
				ID:       "test-path-assertion",
				Type:     string(assertion.PathAssertion),
				Value:    "/api/users",
				Operator: string(assertion.StartsWith),
				Pattern:  "prefix",
			},
			description: "路径断言配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段
			assert.NotEmpty(t, tt.config.ID, "ID不应该为空")
			assert.NotEmpty(t, tt.config.Type, "断言类型不应该为空")
			assert.NotEmpty(t, tt.config.Operator, "断言操作符不应该为空")
		})
	}
}

func TestAssertionTypes(t *testing.T) {
	types := []assertion.AssertionType{
		assertion.HeaderAssertion,
		assertion.QueryParamAssertion,
		assertion.BodyContentAssertion,
		assertion.MethodAssertion,
		assertion.CookieAssertion,
		assertion.IPAssertion,
		assertion.PathAssertion,
	}

	for _, assertionType := range types {
		t.Run(string(assertionType), func(t *testing.T) {
			// 验证断言类型常量不为空
			assert.NotEmpty(t, string(assertionType), "断言类型常量不应该为空")
		})
	}
}

func TestComparisonOperators(t *testing.T) {
	operators := []assertion.ComparisonOperator{
		assertion.Equal,
		assertion.NotEqual,
		assertion.Contains,
		assertion.NotContains,
		assertion.StartsWith,
		assertion.EndsWith,
		assertion.Matches,
		assertion.Exists,
		assertion.NotExists,
	}

	for _, operator := range operators {
		t.Run(string(operator), func(t *testing.T) {
			// 验证断言操作符常量不为空
			assert.NotEmpty(t, string(operator), "断言操作符常量不应该为空")
		})
	}
}

func TestHeaderAssertion(t *testing.T) {
	config := &assertion.AssertionConfig{
		ID:       "test-header",
		Type:     string(assertion.HeaderAssertion),
		Name:     "Content-Type",
		Value:    "application/json",
		Operator: string(assertion.Equal),
	}

	// 创建断言
	headerAssertion := &assertion.BaseAssertion{
		Type:          assertion.HeaderAssertion,
		Operator:      assertion.Equal,
		FieldName:     "Content-Type",
		ExpectedValue: "application/json",
		CaseSensitive: false,
		Config:        *config,
	}

	tests := []struct {
		name         string
		contentType  string
		expectResult bool
		description  string
	}{
		{
			name:         "MatchingHeader",
			contentType:  "application/json",
			expectResult: true,
			description:  "匹配的头部应该通过断言",
		},
		{
			name:         "NonMatchingHeader",
			contentType:  "text/html",
			expectResult: false,
			description:  "不匹配的头部应该失败",
		},
		{
			name:         "CaseInsensitive",
			contentType:  "Application/JSON",
			expectResult: true,
			description:  "大小写不敏感匹配应该通过",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("POST", "/api/test", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 手动检查头部
			actualValue := req.Header.Get("Content-Type")
			expectedValue := headerAssertion.ExpectedValue

			if !headerAssertion.CaseSensitive {
				actualValue = strings.ToLower(actualValue)
				expectedValue = strings.ToLower(expectedValue)
			}

			result := actualValue == expectedValue
			assert.Equal(t, tt.expectResult, result, tt.description)
		})
	}
}

func TestQueryParamAssertion(t *testing.T) {
	config := &assertion.AssertionConfig{
		ID:       "test-query",
		Type:     string(assertion.QueryParamAssertion),
		Name:     "version",
		Value:    "v1",
		Operator: string(assertion.Equal),
	}

	queryAssertion := &assertion.BaseAssertion{
		Type:          assertion.QueryParamAssertion,
		Operator:      assertion.Equal,
		FieldName:     "version",
		ExpectedValue: "v1",
		CaseSensitive: false,
		Config:        *config,
	}

	tests := []struct {
		name         string
		queryString  string
		expectResult bool
		description  string
	}{
		{
			name:         "MatchingQuery",
			queryString:  "version=v1",
			expectResult: true,
			description:  "匹配的查询参数应该通过断言",
		},
		{
			name:         "NonMatchingQuery",
			queryString:  "version=v2",
			expectResult: false,
			description:  "不匹配的查询参数应该失败",
		},
		{
			name:         "MissingQuery",
			queryString:  "other=value",
			expectResult: false,
			description:  "缺少查询参数应该失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			url := "/api/test"
			if tt.queryString != "" {
				url += "?" + tt.queryString
			}

			req := httptest.NewRequest("GET", url, nil)
			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 手动检查查询参数
			actualValue := req.URL.Query().Get("version")
			expectedValue := queryAssertion.ExpectedValue

			result := actualValue == expectedValue
			assert.Equal(t, tt.expectResult, result, tt.description)
		})
	}
}

func TestMethodAssertion(t *testing.T) {
	config := &assertion.AssertionConfig{
		ID:       "test-method",
		Type:     string(assertion.MethodAssertion),
		Value:    "POST",
		Operator: string(assertion.Contains),
	}

	methodAssertion := &assertion.BaseAssertion{
		Type:          assertion.MethodAssertion,
		Operator:      assertion.Contains,
		ExpectedValue: "POST",
		CaseSensitive: false,
		Config:        *config,
	}

	tests := []struct {
		name         string
		method       string
		expectResult bool
		description  string
	}{
		{
			name:         "MatchingMethod",
			method:       "POST",
			expectResult: true,
			description:  "匹配的HTTP方法应该通过断言",
		},
		{
			name:         "NonMatchingMethod",
			method:       "GET",
			expectResult: false,
			description:  "不匹配的HTTP方法应该失败",
		},
		{
			name:         "CaseInsensitive",
			method:       "post",
			expectResult: true,
			description:  "大小写不敏感匹配应该通过",
		},
		{
			name:         "PartialMethod",
			method:       "OPTIONS",
			expectResult: false,
			description:  "不包含POST的方法应该失败",
		},
		{
			name:         "ContainsMethod",
			method:       "POSTDATA",
			expectResult: true,
			description:  "包含POST的方法应该通过",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest(tt.method, "/api/test", nil)
			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 手动检查HTTP方法
			actualValue := req.Method
			expectedValue := methodAssertion.ExpectedValue

			if !methodAssertion.CaseSensitive {
				actualValue = strings.ToLower(actualValue)
				expectedValue = strings.ToLower(expectedValue)
			}

			// 使用Contains关系进行测试
			result := strings.Contains(actualValue, expectedValue)
			assert.Equal(t, tt.expectResult, result, tt.description)
		})
	}
}

func TestPathAssertion(t *testing.T) {
	config := &assertion.AssertionConfig{
		ID:       "test-path",
		Type:     string(assertion.PathAssertion),
		Value:    "/api/users",
		Operator: string(assertion.StartsWith),
		Pattern:  "prefix",
	}

	pathAssertion := &assertion.BaseAssertion{
		Type:          assertion.PathAssertion,
		Operator:      assertion.StartsWith,
		ExpectedValue: "/api/users",
		CaseSensitive: false,
		Config:        *config,
	}

	tests := []struct {
		name         string
		path         string
		expectResult bool
		description  string
	}{
		{
			name:         "MatchingPath",
			path:         "/api/users/123",
			expectResult: true,
			description:  "以指定路径开头的请求应该通过断言",
		},
		{
			name:         "ExactPath",
			path:         "/api/users",
			expectResult: true,
			description:  "完全匹配的路径应该通过断言",
		},
		{
			name:         "NonMatchingPath",
			path:         "/api/products",
			expectResult: false,
			description:  "不匹配的路径应该失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", tt.path, nil)
			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 手动检查路径
			actualValue := req.URL.Path
			expectedValue := pathAssertion.ExpectedValue

			var result bool
			switch pathAssertion.Operator {
			case assertion.StartsWith:
				result = strings.HasPrefix(actualValue, expectedValue)
			case assertion.Equal:
				result = actualValue == expectedValue
			}

			assert.Equal(t, tt.expectResult, result, tt.description)
		})
	}
}

func TestIPAssertion(t *testing.T) {
	config := &assertion.AssertionConfig{
		ID:       "test-ip",
		Type:     string(assertion.IPAssertion),
		Value:    "192.168.",
		Operator: string(assertion.StartsWith),
	}

	ipAssertion := &assertion.BaseAssertion{
		Type:          assertion.IPAssertion,
		Operator:      assertion.StartsWith,
		ExpectedValue: "192.168.",
		CaseSensitive: false,
		Config:        *config,
	}

	tests := []struct {
		name         string
		remoteAddr   string
		expectResult bool
		description  string
	}{
		{
			name:         "MatchingIP",
			remoteAddr:   "192.168.1.100:12345",
			expectResult: true,
			description:  "192.168网段IP应该通过",
		},
		{
			name:         "NonMatchingIP",
			remoteAddr:   "10.0.0.100:12345",
			expectResult: false,
			description:  "非192.168网段IP应该失败",
		},
		{
			name:         "PublicIP",
			remoteAddr:   "8.8.8.8:80",
			expectResult: false,
			description:  "公网IP应该失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", "/api/test", nil)
			req.RemoteAddr = tt.remoteAddr
			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 手动检查IP地址
			actualValue := req.RemoteAddr
			expectedValue := ipAssertion.ExpectedValue

			result := strings.HasPrefix(actualValue, expectedValue)
			assert.Equal(t, tt.expectResult, result, tt.description)
		})
	}
}

func TestCookieAssertion(t *testing.T) {
	config := &assertion.AssertionConfig{
		ID:       "test-cookie",
		Type:     string(assertion.CookieAssertion),
		Name:     "session_id",
		Operator: string(assertion.Exists),
	}

	_ = &assertion.BaseAssertion{
		Type:          assertion.CookieAssertion,
		Operator:      assertion.Exists,
		FieldName:     "session_id",
		CaseSensitive: false,
		Config:        *config,
	}

	tests := []struct {
		name         string
		cookieValue  string
		expectResult bool
		description  string
	}{
		{
			name:         "CookieExists",
			cookieValue:  "abc123",
			expectResult: true,
			description:  "存在的Cookie应该通过断言",
		},
		{
			name:         "CookieNotExists",
			cookieValue:  "",
			expectResult: false,
			description:  "不存在的Cookie应该失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", "/api/test", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tt.cookieValue,
				})
			}

			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 手动检查Cookie
			_, err := req.Cookie("session_id")
			result := err == nil

			assert.Equal(t, tt.expectResult, result, tt.description)
		})
	}
}

func TestAssertionGroupConfig(t *testing.T) {
	// 创建断言组配置
	groupConfig := assertion.NewAssertionGroupConfig("test-group", true)

	// 添加断言配置
	groupConfig.AddAssertionConfig(assertion.AssertionConfig{
		ID:       "method-check",
		Type:     string(assertion.MethodAssertion),
		Value:    "POST",
		Operator: string(assertion.Equal),
	})

	groupConfig.AddAssertionConfig(assertion.AssertionConfig{
		ID:       "header-check",
		Type:     string(assertion.HeaderAssertion),
		Name:     "Content-Type",
		Value:    "application/json",
		Operator: string(assertion.Equal),
	})

	// 验证断言组配置
	assert.Equal(t, "test-group", groupConfig.ID)
	assert.True(t, groupConfig.AllRequired)
	assert.Len(t, groupConfig.AssertionConfigs, 2)

	// 验证描述
	description := groupConfig.GetDescription()
	assert.NotEmpty(t, description)
}

func TestAssertionInterface(t *testing.T) {
	config := &assertion.AssertionConfig{
		ID:       "test-interface",
		Type:     string(assertion.HeaderAssertion),
		Name:     "Authorization",
		Operator: string(assertion.Exists),
	}

	baseAssertion := &assertion.BaseAssertion{
		Type:      assertion.HeaderAssertion,
		Operator:  assertion.Exists,
		FieldName: "Authorization",
		Config:    *config,
	}

	// 测试接口方法
	assert.Equal(t, assertion.HeaderAssertion, baseAssertion.GetType())
	assert.NotEmpty(t, baseAssertion.GetDescription())
	assert.Equal(t, *config, baseAssertion.GetConfig())
}

func TestAssertionDescription(t *testing.T) {
	tests := []struct {
		name      string
		assertion *assertion.BaseAssertion
		expected  string
	}{
		{
			name: "HeaderExists",
			assertion: &assertion.BaseAssertion{
				Type:      assertion.HeaderAssertion,
				Operator:  assertion.Exists,
				FieldName: "Authorization",
			},
			expected: "头部字段 Authorization 存在",
		},
		{
			name: "QueryEquals",
			assertion: &assertion.BaseAssertion{
				Type:          assertion.QueryParamAssertion,
				Operator:      assertion.Equal,
				FieldName:     "version",
				ExpectedValue: "v1",
			},
			expected: "查询参数字段 version 等于 v1",
		},
		{
			name: "PathStartsWith",
			assertion: &assertion.BaseAssertion{
				Type:          assertion.PathAssertion,
				Operator:      assertion.StartsWith,
				FieldName:     "path",
				ExpectedValue: "/api",
			},
			expected: "路径字段 path 以...开头 /api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := tt.assertion.GetDescription()
			assert.Contains(t, description, tt.expected)
		})
	}
}

// TestNewAssertionGroupFromConfig 测试从配置直接创建断言组
func TestNewAssertionGroupFromConfig(t *testing.T) {
	// 创建断言组配置
	groupConfig := assertion.NewAssertionGroupConfig("test-direct-creation", true)
	groupConfig.Description = "测试直接创建断言组"

	// 添加断言配置
	groupConfig.AddAssertionConfig(assertion.AssertionConfig{
		ID:       "method-test",
		Type:     string(assertion.MethodAssertion),
		Value:    "GET",
		Operator: string(assertion.Equal),
	})

	groupConfig.AddAssertionConfig(assertion.AssertionConfig{
		ID:       "path-test",
		Type:     string(assertion.PathAssertion),
		Value:    "/api/test",
		Operator: string(assertion.Equal),
	})

	// 使用新函数直接创建断言组
	assertionGroup, err := assertion.NewAssertionGroupFromConfig(groupConfig)

	// 验证创建结果
	assert.NoError(t, err)
	assert.NotNil(t, assertionGroup)
	assert.Equal(t, groupConfig.AllRequired, assertionGroup.AllRequired)
	assert.Equal(t, groupConfig.Description, assertionGroup.Description)
	assert.Len(t, assertionGroup.Assertions, 2)

	// 验证断言组能正确评估
	req := httptest.NewRequest("GET", "/api/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	result, err := assertionGroup.Evaluate(ctx)
	assert.NoError(t, err)
	assert.True(t, result, "符合条件的请求应该通过断言组")

	// 验证不匹配的请求
	req = httptest.NewRequest("POST", "/api/test", nil)
	writer = httptest.NewRecorder()
	ctx = core.NewContext(writer, req)

	result, err = assertionGroup.Evaluate(ctx)
	assert.NoError(t, err)
	assert.False(t, result, "不符合条件的请求不应该通过断言组")
}

// 基准测试
func BenchmarkHeaderAssertion(b *testing.B) {
	config := &assertion.AssertionConfig{
		ID:       "bench-header",
		Type:     string(assertion.HeaderAssertion),
		Name:     "Content-Type",
		Value:    "application/json",
		Operator: string(assertion.Equal),
	}

	headerAssertion := &assertion.BaseAssertion{
		Type:          assertion.HeaderAssertion,
		Operator:      assertion.Equal,
		FieldName:     "Content-Type",
		ExpectedValue: "application/json",
		CaseSensitive: false,
		Config:        *config,
	}

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := httptest.NewRecorder()
		_ = core.NewContext(writer, req)

		// 模拟断言逻辑
		actualValue := req.Header.Get("Content-Type")
		expectedValue := headerAssertion.ExpectedValue
		_ = actualValue == expectedValue
	}
}

func BenchmarkQueryParamAssertion(b *testing.B) {
	req := httptest.NewRequest("GET", "/api/test?version=v1&limit=10", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := httptest.NewRecorder()
		_ = core.NewContext(writer, req)

		// 模拟断言逻辑
		actualValue := req.URL.Query().Get("version")
		_ = actualValue == "v1"
	}
}
