package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/handler/limiter"
	"gateway/web/views/hubcommon002/models"
)

func TestRateLimitConfig(t *testing.T) {
	config := &models.RateLimitConfig{
		TenantId:            "tenant-001",
		RateLimitConfigId:   "RATE20240615143022A1B2",
		LimitName:           "API限流规则",
		Algorithm:           "token-bucket",
		KeyStrategy:         "ip",
		LimitRate:           100,
		BurstCapacity:       50,
		TimeWindowSeconds:   1,
		RejectionStatusCode: 429,
		RejectionMessage:    "请求过于频繁，请稍后再试",
		ConfigPriority:      0,
		CustomConfig:        `{"enabled": true, "description": "测试配置"}`,
		ActiveFlag:          "Y",
	}

	// 测试表名
	assert.Equal(t, "HUB_GW_RATE_LIMIT_CONFIG", config.TableName())

	// 测试字段值
	assert.Equal(t, "tenant-001", config.TenantId)
	assert.Equal(t, "token-bucket", config.Algorithm)
	assert.Equal(t, "ip", config.KeyStrategy)
	assert.Equal(t, 100, config.LimitRate)
	assert.Equal(t, 50, config.BurstCapacity)
	assert.Equal(t, "Y", config.ActiveFlag)
}

func TestRateLimitConfigConverter_ToLimiterConfig(t *testing.T) {
	converter := models.NewRateLimitConfigConverter()

	tests := []struct {
		name     string
		dbConfig *models.RateLimitConfig
		expected *limiter.RateLimitConfig
		wantErr  bool
	}{
		{
			name: "正确转换基本配置",
			dbConfig: &models.RateLimitConfig{
				RateLimitConfigId:   "test-001",
				LimitName:           "测试限流",
				Algorithm:           "token-bucket",
				KeyStrategy:         "ip",
				LimitRate:           100,
				BurstCapacity:       50,
				TimeWindowSeconds:   1,
				RejectionStatusCode: 429,
				RejectionMessage:    "限流中",
				CustomConfig:        `{"test": true}`,
				ActiveFlag:          "Y",
			},
			expected: &limiter.RateLimitConfig{
				ID:              "test-001",
				Name:            "测试限流",
				Enabled:         true,
				Algorithm:       "token-bucket",
				Rate:            100,
				Burst:           50,
				WindowSize:      1,
				KeyStrategy:     "ip",
				ErrorStatusCode: 429,
				ErrorMessage:    "限流中",
				CustomConfig:    map[string]interface{}{"test": true},
			},
			wantErr: false,
		},
		{
			name: "转换旧格式算法",
			dbConfig: &models.RateLimitConfig{
				RateLimitConfigId: "test-002",
				LimitName:         "旧格式测试",
				Algorithm:         "TOKEN_BUCKET", // 旧格式
				KeyStrategy:       "user",
				LimitRate:         200,
				BurstCapacity:     100,
				TimeWindowSeconds: 5,
				CustomConfig:      "{}",
				ActiveFlag:        "N",
			},
			expected: &limiter.RateLimitConfig{
				ID:           "test-002",
				Name:         "旧格式测试",
				Enabled:      false,
				Algorithm:    "token-bucket", // 转换后的新格式
				Rate:         200,
				Burst:        100,
				WindowSize:   5,
				KeyStrategy:  "user",
				CustomConfig: map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name:     "处理空配置",
			dbConfig: nil,
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ToLimiterConfig(tt.dbConfig)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Enabled, result.Enabled)
			assert.Equal(t, tt.expected.Algorithm, result.Algorithm)
			assert.Equal(t, tt.expected.Rate, result.Rate)
			assert.Equal(t, tt.expected.Burst, result.Burst)
			assert.Equal(t, tt.expected.WindowSize, result.WindowSize)
			assert.Equal(t, tt.expected.KeyStrategy, result.KeyStrategy)
			assert.Equal(t, tt.expected.ErrorStatusCode, result.ErrorStatusCode)
			assert.Equal(t, tt.expected.ErrorMessage, result.ErrorMessage)
			assert.Equal(t, tt.expected.CustomConfig, result.CustomConfig)
		})
	}
}

func TestRateLimitConfigConverter_FromLimiterConfig(t *testing.T) {
	converter := models.NewRateLimitConfigConverter()
	tenantId := "tenant-001"

	tests := []struct {
		name           string
		limiterConfig  *limiter.RateLimitConfig
		expectedFields map[string]interface{}
		wantErr        bool
	}{
		{
			name: "正确转换限流器配置",
			limiterConfig: &limiter.RateLimitConfig{
				ID:              "limiter-001",
				Name:            "限流器测试",
				Enabled:         true,
				Algorithm:       "leaky-bucket",
				Rate:            150,
				Burst:           75,
				WindowSize:      2,
				KeyStrategy:     "path",
				ErrorStatusCode: 503,
				ErrorMessage:    "服务不可用",
				CustomConfig:    map[string]interface{}{"priority": "high"},
			},
			expectedFields: map[string]interface{}{
				"RateLimitConfigId":   "limiter-001",
				"LimitName":           "限流器测试",
				"Algorithm":           "leaky-bucket",
				"KeyStrategy":         "path",
				"LimitRate":           150,
				"BurstCapacity":       75,
				"TimeWindowSeconds":   2,
				"RejectionStatusCode": 503,
				"RejectionMessage":    "服务不可用",
				"ActiveFlag":          "Y",
			},
			wantErr: false,
		},
		{
			name: "处理禁用的配置",
			limiterConfig: &limiter.RateLimitConfig{
				ID:          "limiter-002",
				Name:        "禁用的限流器",
				Enabled:     false,
				Algorithm:   "sliding-window",
				Rate:        50,
				KeyStrategy: "service",
			},
			expectedFields: map[string]interface{}{
				"RateLimitConfigId": "limiter-002",
				"LimitName":         "禁用的限流器",
				"Algorithm":         "sliding-window",
				"KeyStrategy":       "service",
				"LimitRate":         50,
				"ActiveFlag":        "N",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.FromLimiterConfig(tt.limiterConfig, tenantId)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tenantId, result.TenantId)

			for field, expected := range tt.expectedFields {
				switch field {
				case "RateLimitConfigId":
					assert.Equal(t, expected, result.RateLimitConfigId, "字段 %s 不匹配", field)
				case "LimitName":
					assert.Equal(t, expected, result.LimitName, "字段 %s 不匹配", field)
				case "Algorithm":
					assert.Equal(t, expected, result.Algorithm, "字段 %s 不匹配", field)
				case "KeyStrategy":
					assert.Equal(t, expected, result.KeyStrategy, "字段 %s 不匹配", field)
				case "LimitRate":
					assert.Equal(t, expected, result.LimitRate, "字段 %s 不匹配", field)
				case "BurstCapacity":
					assert.Equal(t, expected, result.BurstCapacity, "字段 %s 不匹配", field)
				case "TimeWindowSeconds":
					assert.Equal(t, expected, result.TimeWindowSeconds, "字段 %s 不匹配", field)
				case "RejectionStatusCode":
					assert.Equal(t, expected, result.RejectionStatusCode, "字段 %s 不匹配", field)
				case "RejectionMessage":
					assert.Equal(t, expected, result.RejectionMessage, "字段 %s 不匹配", field)
				case "ActiveFlag":
					assert.Equal(t, expected, result.ActiveFlag, "字段 %s 不匹配", field)
				}
			}

			// 验证CustomConfig是有效的JSON
			var customConfig map[string]interface{}
			err = json.Unmarshal([]byte(result.CustomConfig), &customConfig)
			assert.NoError(t, err, "CustomConfig应该是有效的JSON")
		})
	}
}

func TestRateLimitConfigConverter_ValidateKeyStrategy(t *testing.T) {
	converter := models.NewRateLimitConfigConverter()

	validStrategies := []string{"ip", "user", "path", "service", "route"}
	invalidStrategies := []string{"invalid", "unknown", "", "IP", "User"}

	for _, strategy := range validStrategies {
		assert.True(t, converter.ValidateKeyStrategy(strategy), "应该接受有效的键策略: %s", strategy)
	}

	for _, strategy := range invalidStrategies {
		assert.False(t, converter.ValidateKeyStrategy(strategy), "应该拒绝无效的键策略: %s", strategy)
	}
}

func TestRateLimitConfigConverter_ValidateAlgorithm(t *testing.T) {
	converter := models.NewRateLimitConfigConverter()

	validAlgorithms := []string{"token-bucket", "leaky-bucket", "sliding-window", "fixed-window", "none"}
	invalidAlgorithms := []string{"TOKEN_BUCKET", "invalid", "", "unknown"}

	for _, algorithm := range validAlgorithms {
		assert.True(t, converter.ValidateAlgorithm(algorithm), "应该接受有效的算法: %s", algorithm)
	}

	for _, algorithm := range invalidAlgorithms {
		assert.False(t, converter.ValidateAlgorithm(algorithm), "应该拒绝无效的算法: %s", algorithm)
	}
}

func TestRateLimitConfigConverter_AlgorithmNormalization(t *testing.T) {
	converter := models.NewRateLimitConfigConverter()

	tests := []struct {
		name           string
		inputAlgorithm string
		expected       string
	}{
		{
			name:           "旧格式 TOKEN_BUCKET",
			inputAlgorithm: "TOKEN_BUCKET",
			expected:       "token-bucket",
		},
		{
			name:           "旧格式 LEAKY_BUCKET",
			inputAlgorithm: "LEAKY_BUCKET",
			expected:       "leaky-bucket",
		},
		{
			name:           "旧格式 SLIDING_WINDOW",
			inputAlgorithm: "SLIDING_WINDOW",
			expected:       "sliding-window",
		},
		{
			name:           "旧格式 FIXED_WINDOW",
			inputAlgorithm: "FIXED_WINDOW",
			expected:       "fixed-window",
		},
		{
			name:           "旧格式 NONE",
			inputAlgorithm: "NONE",
			expected:       "none",
		},
		{
			name:           "新格式保持不变",
			inputAlgorithm: "token-bucket",
			expected:       "token-bucket",
		},
		{
			name:           "未知格式保持不变",
			inputAlgorithm: "unknown-algorithm",
			expected:       "unknown-algorithm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConfig := &models.RateLimitConfig{
				RateLimitConfigId: "test-id",
				Algorithm:         tt.inputAlgorithm,
				ActiveFlag:        "Y",
			}

			result, err := converter.ToLimiterConfig(dbConfig)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result.Algorithm), "算法格式转换结果不符合预期")
		})
	}
}

func TestRateLimitConfigConverter_MergeCustomConfig(t *testing.T) {
	converter := models.NewRateLimitConfigConverter()

	base := map[string]interface{}{
		"timeout": 30,
		"retry":   3,
		"metadata": map[string]interface{}{
			"version": "1.0",
		},
	}

	override := map[string]interface{}{
		"timeout": 60,   // 覆盖现有值
		"debug":   true, // 新增值
	}

	result := converter.MergeCustomConfig(base, override)

	assert.Equal(t, 60, result["timeout"], "应该覆盖已存在的值")
	assert.Equal(t, 3, result["retry"], "应该保留基础配置中的值")
	assert.Equal(t, true, result["debug"], "应该添加新的值")
	assert.NotNil(t, result["metadata"], "应该保留嵌套对象")

	// 测试空值处理
	nilResult := converter.MergeCustomConfig(nil, override)
	assert.Equal(t, override, nilResult, "基础配置为nil时应该返回覆盖配置")

	baseResult := converter.MergeCustomConfig(base, nil)
	assert.Equal(t, base, baseResult, "覆盖配置为nil时应该返回基础配置")
}
