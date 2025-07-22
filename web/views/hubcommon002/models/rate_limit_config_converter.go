package models

import (
	"encoding/json"
	"gateway/internal/gateway/handler/limiter"
)

// RateLimitConfigConverter 限流配置转换器
type RateLimitConfigConverter struct{}

// NewRateLimitConfigConverter 创建转换器
func NewRateLimitConfigConverter() *RateLimitConfigConverter {
	return &RateLimitConfigConverter{}
}

// ToLimiterConfig 将数据库模型转换为限流器配置
func (c *RateLimitConfigConverter) ToLimiterConfig(dbConfig *RateLimitConfig) (*limiter.RateLimitConfig, error) {
	if dbConfig == nil {
		return nil, nil
	}

	// 解析自定义配置
	customConfig := make(map[string]interface{})
	if dbConfig.CustomConfig != "" {
		if err := json.Unmarshal([]byte(dbConfig.CustomConfig), &customConfig); err != nil {
			// 如果解析失败，使用空配置
			customConfig = make(map[string]interface{})
		}
	}

	// 转换算法格式
	algorithm := c.normalizeAlgorithm(dbConfig.Algorithm)

	// 创建限流器配置
	config := &limiter.RateLimitConfig{
		ID:              dbConfig.RateLimitConfigId,
		Name:            dbConfig.LimitName,
		Enabled:         dbConfig.ActiveFlag == "Y",
		Algorithm:       limiter.RateLimitAlgorithm(algorithm),
		Rate:            dbConfig.LimitRate,
		Burst:           dbConfig.BurstCapacity,
		WindowSize:      dbConfig.TimeWindowSeconds,
		KeyStrategy:     dbConfig.KeyStrategy,
		ErrorStatusCode: dbConfig.RejectionStatusCode,
		ErrorMessage:    dbConfig.RejectionMessage,
		CustomConfig:    customConfig,
	}

	return config, nil
}

// FromLimiterConfig 将限流器配置转换为数据库模型
func (c *RateLimitConfigConverter) FromLimiterConfig(limiterConfig *limiter.RateLimitConfig, tenantId string) (*RateLimitConfig, error) {
	if limiterConfig == nil {
		return nil, nil
	}

	// 序列化自定义配置
	customConfigBytes, err := json.Marshal(limiterConfig.CustomConfig)
	if err != nil {
		customConfigBytes = []byte("{}")
	}

	// 创建数据库模型
	dbConfig := &RateLimitConfig{
		TenantId:            tenantId,
		RateLimitConfigId:   limiterConfig.ID,
		LimitName:           limiterConfig.Name,
		Algorithm:           string(limiterConfig.Algorithm),
		KeyStrategy:         limiterConfig.KeyStrategy,
		LimitRate:           limiterConfig.Rate,
		BurstCapacity:       limiterConfig.Burst,
		TimeWindowSeconds:   limiterConfig.WindowSize,
		RejectionStatusCode: limiterConfig.ErrorStatusCode,
		RejectionMessage:    limiterConfig.ErrorMessage,
		CustomConfig:        string(customConfigBytes),
		ActiveFlag:          "Y",
	}

	if limiterConfig.Enabled {
		dbConfig.ActiveFlag = "Y"
	} else {
		dbConfig.ActiveFlag = "N"
	}

	return dbConfig, nil
}

// normalizeAlgorithm 标准化算法格式
func (c *RateLimitConfigConverter) normalizeAlgorithm(algorithm string) string {
	// 将旧格式转换为新格式
	switch algorithm {
	case "TOKEN_BUCKET":
		return "token-bucket"
	case "LEAKY_BUCKET":
		return "leaky-bucket"
	case "SLIDING_WINDOW":
		return "sliding-window"
	case "FIXED_WINDOW":
		return "fixed-window"
	case "NONE":
		return "none"
	default:
		// 如果已经是新格式，直接返回
		return algorithm
	}
}

// ValidateKeyStrategy 验证键策略
func (c *RateLimitConfigConverter) ValidateKeyStrategy(keyStrategy string) bool {
	validStrategies := []string{"ip", "user", "path", "service", "route"}
	for _, valid := range validStrategies {
		if keyStrategy == valid {
			return true
		}
	}
	return false
}

// ValidateAlgorithm 验证算法类型
func (c *RateLimitConfigConverter) ValidateAlgorithm(algorithm string) bool {
	validAlgorithms := []string{"token-bucket", "leaky-bucket", "sliding-window", "fixed-window", "none"}
	for _, valid := range validAlgorithms {
		if algorithm == valid {
			return true
		}
	}
	return false
}

// GetDefaultCustomConfig 获取默认自定义配置
func (c *RateLimitConfigConverter) GetDefaultCustomConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":     true,
		"description": "Default rate limit configuration",
		"metadata": map[string]interface{}{
			"created_by": "system",
			"version":    "1.0",
		},
	}
}

// MergeCustomConfig 合并自定义配置
func (c *RateLimitConfigConverter) MergeCustomConfig(base, override map[string]interface{}) map[string]interface{} {
	if base == nil {
		base = make(map[string]interface{})
	}
	if override == nil {
		return base
	}

	result := make(map[string]interface{})
	// 复制基础配置
	for k, v := range base {
		result[k] = v
	}
	// 覆盖配置
	for k, v := range override {
		result[k] = v
	}

	return result
}
