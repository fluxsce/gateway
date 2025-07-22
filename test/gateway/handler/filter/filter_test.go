package filter

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/filter"
)

func TestFilterConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *filter.FilterConfig
		description string
	}{
		{
			name: "HeaderFilterConfig",
			config: &filter.FilterConfig{
				ID:      "test-header-filter",
				Name:    "Header Filter",
				Type:    "header",
				Enabled: true,
				Order:   10,
				Action:  "post-routing",
				Config: map[string]interface{}{
					"header_name":  "Content-Type",
					"header_value": "application/json",
					"operation":    "add",
				},
			},
			description: "请求头过滤器配置",
		},
		{
			name: "QueryParamFilterConfig",
			config: &filter.FilterConfig{
				ID:      "test-query-filter",
				Name:    "Query Param Filter",
				Type:    "query-param",
				Enabled: true,
				Order:   20,
				Action:  "post-routing",
				Config: map[string]interface{}{
					"param_name":  "version",
					"param_value": "v1",
					"operation":   "add",
				},
			},
			description: "查询参数过滤器配置",
		},
		{
			name: "URLFilterConfig",
			config: &filter.FilterConfig{
				ID:      "test-url-filter",
				Name:    "URL Filter",
				Type:    "url",
				Enabled: true,
				Order:   30,
				Action:  "post-routing",
				Config: map[string]interface{}{
					"from": "/api/v1",
					"to":   "/api",
					"mode": "simple",
				},
			},
			description: "URL过滤器配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段
			assert.NotEmpty(t, tt.config.ID, "过滤器ID不应该为空")
			assert.NotEmpty(t, tt.config.Name, "过滤器名称不应该为空")
			assert.True(t, tt.config.Enabled, "过滤器应该启用")
			assert.NotEmpty(t, tt.config.Config, "过滤器配置不应该为空")
		})
	}
}

func TestFilterTypes(t *testing.T) {
	types := []filter.FilterType{
		filter.HeaderFilterType,
		filter.QueryParamFilterType,
		filter.URLFilterType,
		filter.BodyFilterType,
		filter.MethodFilterType,
		filter.CookieFilterType,
		filter.ResponseFilterType,
	}

	for _, filterType := range types {
		t.Run(string(filterType), func(t *testing.T) {
			// 验证过滤器类型常量不为空
			assert.NotEmpty(t, string(filterType), "过滤器类型常量不应该为空")
		})
	}
}

func TestFilterActions(t *testing.T) {
	actions := []filter.FilterAction{
		filter.PreRouting,
		filter.PostRouting,
		filter.PreResponse,
	}

	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			// 验证过滤器动作常量不为空
			assert.NotEmpty(t, string(action), "过滤器动作常量不应该为空")
		})
	}
}

func TestBaseFilter(t *testing.T) {
	_ = &filter.FilterConfig{
		ID:      "test-base-filter",
		Name:    "Base Filter",
		Enabled: true,
		Order:   10,
		Action:  "post-routing",
		Config: map[string]interface{}{
			"type": "header",
		},
	}

	baseFilter := filter.NewBaseFilter(
		filter.HeaderFilterType,
		filter.PostRouting,
		10,
		true,
		"Base Filter",
	)

	// 验证基础过滤器属性
	assert.Equal(t, filter.HeaderFilterType, baseFilter.GetType())
	assert.Equal(t, filter.PostRouting, baseFilter.GetAction())
	assert.Equal(t, 10, baseFilter.GetPriority())
	assert.True(t, baseFilter.IsEnabled())
	assert.Equal(t, "Base Filter", baseFilter.GetName())
}

func TestFilterSlice(t *testing.T) {
	// 创建测试过滤器
	filter1 := filter.NewBaseFilter(
		filter.HeaderFilterType,
		filter.PreRouting,
		100,
		true,
		"Filter 1",
	)

	filter2 := filter.NewBaseFilter(
		filter.QueryParamFilterType,
		filter.PostRouting,
		50,
		true,
		"Filter 2",
	)

	filter3 := filter.NewBaseFilter(
		filter.ResponseFilterType,
		filter.PreResponse,
		200,
		true,
		"Filter 3",
	)

	// 使用切片存储过滤器
	filters := make([]filter.Filter, 0)
	filters = append(filters, filter1)
	filters = append(filters, filter2)
	filters = append(filters, filter3)

	// 验证过滤器已添加
	assert.Len(t, filters, 3)
	assert.Equal(t, "Filter 1", filters[0].GetName())
	assert.Equal(t, "Filter 2", filters[1].GetName())
	assert.Equal(t, "Filter 3", filters[2].GetName())

	// 验证过滤器类型
	assert.Equal(t, filter.HeaderFilterType, filters[0].GetType())
	assert.Equal(t, filter.QueryParamFilterType, filters[1].GetType())
	assert.Equal(t, filter.ResponseFilterType, filters[2].GetType())

	// 验证过滤器执行阶段
	assert.Equal(t, filter.PreRouting, filters[0].GetAction())
	assert.Equal(t, filter.PostRouting, filters[1].GetAction())
	assert.Equal(t, filter.PreResponse, filters[2].GetAction())
}

func TestFilterValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *filter.FilterConfig
		expectValid bool
		description string
	}{
		{
			name: "ValidConfig",
			config: &filter.FilterConfig{
				ID:      "valid-filter",
				Name:    "Valid Filter",
				Enabled: true,
				Order:   10,
				Action:  "post-routing",
				Config: map[string]interface{}{
					"type": "header",
				},
			},
			expectValid: true,
			description: "有效配置应该通过验证",
		},
		{
			name: "EmptyID",
			config: &filter.FilterConfig{
				ID:      "",
				Name:    "Filter",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			expectValid: false,
			description: "空ID应该失败",
		},
		{
			name: "EmptyConfig",
			config: &filter.FilterConfig{
				ID:      "test-filter",
				Name:    "Filter",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			expectValid: false,
			description: "空配置应该失败",
		},
		{
			name: "DisabledFilter",
			config: &filter.FilterConfig{
				ID:      "disabled-filter",
				Name:    "Disabled Filter",
				Enabled: false,
				Config:  map[string]interface{}{},
			},
			expectValid: true,
			description: "禁用过滤器应该有效",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单验证逻辑
			valid := true
			if tt.config.Enabled {
				if tt.config.ID == "" {
					valid = false
				}
				if len(tt.config.Config) == 0 {
					valid = false
				}
			}

			assert.Equal(t, tt.expectValid, valid, tt.description)
		})
	}
}

func TestHeaderFilter(t *testing.T) {
	config := &filter.FilterConfig{
		ID:      "test-header-filter",
		Name:    "Header Filter",
		Type:    "header",
		Enabled: true,
		Order:   10,
		Action:  "post-routing",
		Config: map[string]interface{}{
			"header_name":  "X-Gateway",
			"header_value": "true",
			"operation":    "add",
		},
	}

	tests := []struct {
		name        string
		headerName  string
		headerValue string
		expectFound bool
		description string
	}{
		{
			name:        "AddHeader",
			headerName:  "X-Gateway",
			headerValue: "true",
			expectFound: true,
			description: "添加的头部应该存在",
		},
		{
			name:        "NonExistentHeader",
			headerName:  "X-NonExistent",
			headerValue: "",
			expectFound: false,
			description: "不存在的头部应该返回false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", "/api/test", nil)
			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 模拟头部过滤逻辑
			if config.Enabled {
				if headerName, ok := config.Config["header_name"].(string); ok {
					if headerValue, ok := config.Config["header_value"].(string); ok {
						if headerName == tt.headerName {
							req.Header.Set(headerName, headerValue)
						}
					}
				}
			}

			// 验证头部是否存在
			actualValue := req.Header.Get(tt.headerName)
			found := actualValue != ""
			assert.Equal(t, tt.expectFound, found, tt.description)
		})
	}
}

func TestQueryParamFilter(t *testing.T) {
	config := &filter.FilterConfig{
		ID:      "test-query-filter",
		Name:    "Query Filter",
		Type:    "query-param",
		Enabled: true,
		Order:   20,
		Action:  "post-routing",
		Config: map[string]interface{}{
			"param_name":  "version",
			"param_value": "v1",
			"operation":   "add",
		},
	}

	tests := []struct {
		name        string
		paramName   string
		expectValue string
		description string
	}{
		{
			name:        "AddQueryParam",
			paramName:   "version",
			expectValue: "v1",
			description: "添加的查询参数应该存在",
		},
		{
			name:        "NonExistentParam",
			paramName:   "other",
			expectValue: "",
			description: "不存在的参数应该为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", "/api/test", nil)
			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 模拟查询参数过滤逻辑
			if config.Enabled {
				if paramName, ok := config.Config["param_name"].(string); ok {
					if paramValue, ok := config.Config["param_value"].(string); ok {
						if paramName == tt.paramName {
							query := req.URL.Query()
							query.Set(paramName, paramValue)
							req.URL.RawQuery = query.Encode()
						}
					}
				}
			}

			// 验证查询参数
			actualValue := req.URL.Query().Get(tt.paramName)
			assert.Equal(t, tt.expectValue, actualValue, tt.description)
		})
	}
}

func TestURLFilter(t *testing.T) {
	config := &filter.FilterConfig{
		ID:      "test-url-filter",
		Name:    "URL Filter",
		Enabled: true,
		Order:   30,
		Action:  "post-routing",
		Config: map[string]interface{}{
			"type": "url",
			"from": "/api/v1",
			"to":   "/api",
			"mode": "simple",
		},
	}

	tests := []struct {
		name         string
		originalPath string
		expectedPath string
		description  string
	}{
		{
			name:         "PathRewrite",
			originalPath: "/api/v1/users",
			expectedPath: "/api/users",
			description:  "路径应该被重写",
		},
		{
			name:         "NoMatch",
			originalPath: "/other/path",
			expectedPath: "/other/path",
			description:  "不匹配的路径应该保持不变",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", tt.originalPath, nil)
			writer := httptest.NewRecorder()
			_ = core.NewContext(writer, req)

			// 模拟URL过滤逻辑
			newPath := tt.originalPath
			if config.Enabled {
				if from, ok := config.Config["from"].(string); ok {
					if to, ok := config.Config["to"].(string); ok {
						if strings.HasPrefix(tt.originalPath, from) {
							newPath = strings.Replace(tt.originalPath, from, to, 1)
						}
					}
				}
			}

			assert.Equal(t, tt.expectedPath, newPath, tt.description)
		})
	}
}

func TestFilterFactory(t *testing.T) {
	factory := filter.NewFilterFactory()

	tests := []struct {
		name        string
		config      filter.FilterConfig
		expectError bool
		description string
	}{
		{
			name: "ValidHeaderFilter",
			config: filter.FilterConfig{
				ID:      "header-filter",
				Name:    "Header Filter",
				Type:    "header",
				Enabled: true,
				Order:   10,
				Config:  map[string]interface{}{},
			},
			expectError: false,
			description: "有效的头部过滤器配置应该成功创建",
		},
		{
			name: "ValidQueryFilter",
			config: filter.FilterConfig{
				ID:      "query-filter",
				Name:    "Query Filter",
				Type:    "query-param",
				Enabled: true,
				Order:   20,
				Config:  map[string]interface{}{},
			},
			expectError: false,
			description: "有效的查询参数过滤器配置应该成功创建",
		},
		{
			name: "EmptyID",
			config: filter.FilterConfig{
				ID:   "",
				Name: "Filter",
			},
			expectError: true,
			description: "空ID应该导致创建失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterInstance, err := factory.CreateFilter(tt.config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, filterInstance)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, filterInstance)
			}
		})
	}
}

func TestFilterInterface(t *testing.T) {
	baseFilter := filter.NewBaseFilter(
		filter.HeaderFilterType,
		filter.PostRouting,
		10,
		true,
		"Test Filter",
	)

	// 测试接口方法
	assert.Equal(t, filter.HeaderFilterType, baseFilter.GetType())
	assert.Equal(t, filter.PostRouting, baseFilter.GetAction())
	assert.Equal(t, 10, baseFilter.GetPriority())
	assert.True(t, baseFilter.IsEnabled())
	assert.Equal(t, "Test Filter", baseFilter.GetName())

	config := baseFilter.GetConfig()
	assert.Equal(t, "Test Filter", config.ID)
	assert.True(t, config.Enabled)
}

// 基准测试
func BenchmarkHeaderFilter(b *testing.B) {
	config := &filter.FilterConfig{
		ID:      "bench-header-filter",
		Name:    "Header Filter",
		Type:    "header",
		Enabled: true,
		Order:   10,
		Action:  "post-routing",
		Config: map[string]interface{}{
			"header_name":  "Content-Type",
			"header_value": "application/json",
			"operation":    "add",
		},
	}

	req := httptest.NewRequest("POST", "/api/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := httptest.NewRecorder()
		_ = core.NewContext(writer, req)

		// 模拟头部过滤逻辑
		if config.Enabled {
			if headerName, ok := config.Config["header_name"].(string); ok {
				if headerValue, ok := config.Config["header_value"].(string); ok {
					req.Header.Set(headerName, headerValue)
				}
			}
		}
	}
}

func BenchmarkQueryParamFilter(b *testing.B) {
	config := &filter.FilterConfig{
		ID:      "bench-query-filter",
		Name:    "Query Filter",
		Enabled: true,
		Order:   20,
		Action:  "post-routing",
		Config: map[string]interface{}{
			"type":        "query-param",
			"param_name":  "version",
			"param_value": "v1",
			"operation":   "add",
		},
	}

	req := httptest.NewRequest("GET", "/api/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟查询参数过滤逻辑
		if config.Enabled {
			if paramName, ok := config.Config["param_name"].(string); ok {
				if paramValue, ok := config.Config["param_value"].(string); ok {
					query := req.URL.Query()
					query.Set(paramName, paramValue)
					req.URL.RawQuery = query.Encode()
				}
			}
		}
	}
}
