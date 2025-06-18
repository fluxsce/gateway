package cors

import (
	"fmt"
	"strings"
)

// CORSHandlerFactory CORS处理器工厂
type CORSHandlerFactory struct{}

// NewCORSHandlerFactory 创建CORS处理器工厂
func NewCORSHandlerFactory() *CORSHandlerFactory {
	return &CORSHandlerFactory{}
}

// CreateCORSHandler 根据配置创建CORS处理器
func (f *CORSHandlerFactory) CreateCORSHandler(config CORSConfig) (CORSHandler, error) {
	if config.Strategy == "" {
		config.Strategy = StrategyDefault
	}

	// 根据CORS策略创建相应的处理器
	switch config.Strategy {
	case StrategyDefault:
		return f.createDefaultCORSHandler(config)
	case StrategyStrict:
		return f.createStrictCORSHandler(config)
	case StrategyPermissive:
		return f.createPermissiveCORSHandler(config)
	case StrategyCustom:
		return f.createCustomCORSHandler(config)
	default:
		return nil, fmt.Errorf("不支持的CORS策略: %s", config.Strategy)
	}
}

// CreateCompositeCORSHandler 创建复合CORS处理器
func (f *CORSHandlerFactory) CreateCompositeCORSHandler(config CORSConfig) (*CORS, error) {
	cors := NewCORS(&config)

	// 根据策略创建相应的处理器
	handler, err := f.createCORSHandlerFromConfig(&config)
	if err != nil {
		return nil, fmt.Errorf("创建CORS处理器失败: %w", err)
	}

	cors.SetHandler(handler)
	return cors, nil
}

// createDefaultCORSHandler 创建默认CORS处理器
func (f *CORSHandlerFactory) createDefaultCORSHandler(config CORSConfig) (CORSHandler, error) {
	// 应用默认配置
	if config.Strategy == "" {
		config.Strategy = StrategyDefault
	}

	handler := NewDefaultCORSHandler(&config)
	handler.Enabled = config.Enabled

	// 验证配置
	if err := handler.Validate(); err != nil {
		return nil, fmt.Errorf("CORS配置验证失败: %w", err)
	}

	return handler, nil
}

// createStrictCORSHandler 创建严格CORS处理器
func (f *CORSHandlerFactory) createStrictCORSHandler(config CORSConfig) (CORSHandler, error) {
	// 应用严格配置
	strictConfig := StrictCORSConfig
	// 合并配置，但保持严格性
	if len(config.AllowOrigins) > 0 {
		strictConfig.AllowOrigins = config.AllowOrigins
	}
	if len(config.AllowMethods) > 0 {
		// 限制允许的方法
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}
		strictConfig.AllowMethods = f.filterMethods(config.AllowMethods, allowedMethods)
	}
	if len(config.AllowHeaders) > 0 {
		strictConfig.AllowHeaders = config.AllowHeaders
	}
	if config.MaxAge > 0 && config.MaxAge <= 3600 {
		strictConfig.MaxAge = config.MaxAge
	}
	// 严格模式不允许携带凭证
	strictConfig.AllowCredentials = false
	strictConfig.Strategy = StrategyStrict
	strictConfig.Enabled = config.Enabled

	handler := NewDefaultCORSHandler(&strictConfig)
	handler.Strategy = StrategyStrict

	return handler, nil
}

// createPermissiveCORSHandler 创建宽松CORS处理器
func (f *CORSHandlerFactory) createPermissiveCORSHandler(config CORSConfig) (CORSHandler, error) {
	// 应用宽松配置
	permissiveConfig := PermissiveCORSConfig
	// 合并配置，保持宽松性
	if len(config.AllowOrigins) > 0 {
		permissiveConfig.AllowOrigins = config.AllowOrigins
	}
	if len(config.AllowMethods) > 0 {
		permissiveConfig.AllowMethods = config.AllowMethods
	}
	if len(config.AllowHeaders) > 0 {
		permissiveConfig.AllowHeaders = config.AllowHeaders
	}
	if len(config.ExposeHeaders) > 0 {
		permissiveConfig.ExposeHeaders = config.ExposeHeaders
	}
	if config.MaxAge > 0 {
		permissiveConfig.MaxAge = config.MaxAge
	}
	// 保持宽松的凭证策略
	permissiveConfig.AllowCredentials = true
	permissiveConfig.Strategy = StrategyPermissive
	permissiveConfig.Enabled = config.Enabled

	handler := NewDefaultCORSHandler(&permissiveConfig)
	handler.Strategy = StrategyPermissive

	return handler, nil
}

// createCustomCORSHandler 创建自定义CORS处理器
func (f *CORSHandlerFactory) createCustomCORSHandler(config CORSConfig) (CORSHandler, error) {
	if len(config.AllowOrigins) == 0 {
		return nil, fmt.Errorf("自定义CORS策略需要完整的配置")
	}

	config.Strategy = StrategyCustom

	handler := NewDefaultCORSHandler(&config)
	handler.Strategy = StrategyCustom

	// 验证配置
	if err := handler.Validate(); err != nil {
		return nil, fmt.Errorf("自定义CORS配置验证失败: %w", err)
	}

	return handler, nil
}

// createCORSHandlerFromConfig 从CORSConfig创建处理器
func (f *CORSHandlerFactory) createCORSHandlerFromConfig(config *CORSConfig) (CORSHandler, error) {
	if config == nil {
		return nil, fmt.Errorf("CORS配置不能为空")
	}

	switch config.Strategy {
	case StrategyDefault, "":
		return NewDefaultCORSHandler(config), nil
	case StrategyStrict:
		strictConfig := StrictCORSConfig
		if len(config.AllowOrigins) > 0 {
			strictConfig.AllowOrigins = config.AllowOrigins
		}
		return NewDefaultCORSHandler(&strictConfig), nil
	case StrategyPermissive:
		permissiveConfig := PermissiveCORSConfig
		if len(config.AllowOrigins) > 0 {
			permissiveConfig.AllowOrigins = config.AllowOrigins
		}
		return NewDefaultCORSHandler(&permissiveConfig), nil
	case StrategyCustom:
		return NewDefaultCORSHandler(config), nil
	default:
		return nil, fmt.Errorf("不支持的CORS策略: %s", config.Strategy)
	}
}

// filterMethods 过滤允许的HTTP方法
func (f *CORSHandlerFactory) filterMethods(requested, allowed []string) []string {
	result := make([]string, 0)
	allowedMap := make(map[string]bool)

	for _, method := range allowed {
		allowedMap[strings.ToUpper(method)] = true
	}

	for _, method := range requested {
		if allowedMap[strings.ToUpper(method)] {
			result = append(result, strings.ToUpper(method))
		}
	}

	return result
}

// GetSupportedStrategies 获取支持的CORS策略列表
func GetSupportedStrategies() []CORSStrategy {
	return []CORSStrategy{
		StrategyDefault,
		StrategyStrict,
		StrategyPermissive,
		StrategyCustom,
	}
}

// GetStrategyDescription 获取策略描述
func GetStrategyDescription(strategy CORSStrategy) string {
	descriptions := map[CORSStrategy]string{
		StrategyDefault:    "默认CORS策略，平衡安全性和兼容性",
		StrategyStrict:     "严格CORS策略，高安全性，限制性强",
		StrategyPermissive: "宽松CORS策略，高兼容性，安全性较低",
		StrategyCustom:     "自定义CORS策略，完全由用户配置",
	}

	if desc, exists := descriptions[strategy]; exists {
		return desc
	}
	return "未知CORS策略"
}

// GetPresetConfig 获取预设配置
func GetPresetConfig(strategy CORSStrategy) CORSConfig {
	switch strategy {
	case StrategyDefault:
		return DefaultCORSConfig
	case StrategyStrict:
		return StrictCORSConfig
	case StrategyPermissive:
		return PermissiveCORSConfig
	default:
		return DefaultCORSConfig
	}
}
