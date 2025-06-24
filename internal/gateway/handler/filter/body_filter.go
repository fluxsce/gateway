package filter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gohub/internal/gateway/core"
)

// BodyModifierType 请求体修改类型
type BodyModifierType string

const (
	// TransformBody 转换请求体
	TransformBody BodyModifierType = "transform"

	// ValidateBody 验证请求体
	ValidateBody BodyModifierType = "validate"

	// ModifyBody 修改请求体
	ModifyBody BodyModifierType = "modify"

	// FilterBody 过滤请求体字段
	FilterBody BodyModifierType = "filter"
)

// BodyFilter 请求体过滤器
// 用于处理和修改请求体内容
type BodyFilter struct {
	BaseFilter

	// 修改类型
	ModifierType BodyModifierType

	// 过滤器操作配置
	Operation string

	// 配置参数
	FilterConfig map[string]interface{}

	// 允许的内容类型
	AllowedContentTypes []string

	// 最大请求体大小（字节）
	MaxBodySize int64
}

// BodyFilterFromConfig 从配置创建请求体过滤器
func BodyFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	bodyFilter := NewBodyFilter(config.Name, action, order)
	bodyFilter.originalConfig = config

	// 从配置中提取请求体操作参数
	if err := configureBodyFilter(bodyFilter, config.Config); err != nil {
		return nil, fmt.Errorf("配置请求体过滤器失败: %w", err)
	}

	return bodyFilter, nil
}

// NewBodyFilter 创建请求体过滤器
func NewBodyFilter(name string, action FilterAction, priority int) *BodyFilter {
	baseFilter := NewBaseFilter(BodyFilterType, action, priority, true, name)
	return &BodyFilter{
		BaseFilter:          *baseFilter,
		AllowedContentTypes: []string{"application/json", "application/xml", "text/plain"},
		MaxBodySize:         1024 * 1024, // 默认1MB
	}
}

// Apply 实现Filter接口
func (f *BodyFilter) Apply(ctx *core.Context) error {
	if ctx.Request == nil {
		return fmt.Errorf("request is nil")
	}

	// 检查内容类型
	contentType := ctx.Request.Header.Get("Content-Type")
	if !f.isAllowedContentType(contentType) {
		// 如果内容类型不被允许，跳过处理
		return nil
	}

	// 检查是否有请求体
	if ctx.Request.Body == nil {
		return nil
	}

	// 根据修改类型执行不同的操作
	switch f.ModifierType {
	case TransformBody:
		return f.transformBody(ctx)
	case ValidateBody:
		return f.validateBody(ctx)
	case ModifyBody:
		return f.modifyBody(ctx)
	case FilterBody:
		return f.filterBody(ctx)
	default:
		// 默认不执行任何操作
		return nil
	}
}

// transformBody 转换请求体
func (f *BodyFilter) transformBody(ctx *core.Context) error {
	body, err := f.readBody(ctx)
	if err != nil {
		return err
	}

	// 根据配置执行转换
	if transformer, ok := f.FilterConfig["transformer"].(string); ok {
		switch transformer {
		case "json_to_xml":
			return f.jsonToXML(ctx, body)
		case "xml_to_json":
			return f.xmlToJSON(ctx, body)
		case "uppercase":
			return f.toUpperCase(ctx, body)
		case "lowercase":
			return f.toLowerCase(ctx, body)
		default:
			// 未知转换器，保持原样
			return f.setBody(ctx, body)
		}
	}

	return nil
}

// validateBody 验证请求体
func (f *BodyFilter) validateBody(ctx *core.Context) error {
	body, err := f.readBody(ctx)
	if err != nil {
		return err
	}

	// 根据配置执行验证
	if validationType, ok := f.FilterConfig["validation_type"].(string); ok {
		switch validationType {
		case "json_schema":
			return f.validateJSONSchema(ctx, body)
		case "required_fields":
			return f.validateRequiredFields(ctx, body)
		case "max_size":
			return f.validateMaxSize(ctx, body)
		default:
			// 未知验证类型，跳过验证
			return f.setBody(ctx, body)
		}
	}

	return f.setBody(ctx, body)
}

// modifyBody 修改请求体
func (f *BodyFilter) modifyBody(ctx *core.Context) error {
	body, err := f.readBody(ctx)
	if err != nil {
		return err
	}

	// 根据配置执行修改
	if modification, ok := f.FilterConfig["modification"].(string); ok {
		switch modification {
		case "add_fields":
			return f.addFields(ctx, body)
		case "remove_fields":
			return f.removeFields(ctx, body)
		case "replace_values":
			return f.replaceValues(ctx, body)
		default:
			// 未知修改类型，保持原样
			return f.setBody(ctx, body)
		}
	}

	return f.setBody(ctx, body)
}

// filterBody 过滤请求体字段
func (f *BodyFilter) filterBody(ctx *core.Context) error {
	body, err := f.readBody(ctx)
	if err != nil {
		return err
	}

	// 如果是JSON，可以过滤字段
	if strings.Contains(ctx.Request.Header.Get("Content-Type"), "application/json") {
		return f.filterJSONFields(ctx, body)
	}

	// 其他类型暂时保持原样
	return f.setBody(ctx, body)
}

// 辅助方法

// readBody 读取请求体
func (f *BodyFilter) readBody(ctx *core.Context) ([]byte, error) {
	if ctx.Request.Body == nil {
		return []byte{}, nil
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("读取请求体失败: %w", err)
	}

	// 检查大小限制
	if f.MaxBodySize > 0 && int64(len(body)) > f.MaxBodySize {
		return nil, fmt.Errorf("请求体大小超过限制: %d bytes", len(body))
	}

	return body, nil
}

// setBody 设置请求体
func (f *BodyFilter) setBody(ctx *core.Context, body []byte) error {
	ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	ctx.Request.ContentLength = int64(len(body))
	return nil
}

// isAllowedContentType 检查内容类型是否被允许
func (f *BodyFilter) isAllowedContentType(contentType string) bool {
	if len(f.AllowedContentTypes) == 0 {
		return true // 如果没有限制，允许所有类型
	}

	for _, allowed := range f.AllowedContentTypes {
		if strings.Contains(contentType, allowed) {
			return true
		}
	}
	return false
}

// 具体转换方法

// jsonToXML JSON转XML
func (f *BodyFilter) jsonToXML(ctx *core.Context, body []byte) error {
	// 这里可以实现JSON到XML的转换逻辑
	// 简化实现，实际应该使用专门的转换库
	ctx.Request.Header.Set("Content-Type", "application/xml")
	return f.setBody(ctx, body)
}

// xmlToJSON XML转JSON
func (f *BodyFilter) xmlToJSON(ctx *core.Context, body []byte) error {
	// 这里可以实现XML到JSON的转换逻辑
	ctx.Request.Header.Set("Content-Type", "application/json")
	return f.setBody(ctx, body)
}

// toUpperCase 转换为大写
func (f *BodyFilter) toUpperCase(ctx *core.Context, body []byte) error {
	upperBody := bytes.ToUpper(body)
	return f.setBody(ctx, upperBody)
}

// toLowerCase 转换为小写
func (f *BodyFilter) toLowerCase(ctx *core.Context, body []byte) error {
	lowerBody := bytes.ToLower(body)
	return f.setBody(ctx, lowerBody)
}

// 验证方法

// validateJSONSchema 验证JSON模式
func (f *BodyFilter) validateJSONSchema(ctx *core.Context, body []byte) error {
	// 这里可以实现JSON模式验证
	// 简化实现，检查是否为有效JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return fmt.Errorf("无效的JSON格式: %w", err)
	}
	return f.setBody(ctx, body)
}

// validateRequiredFields 验证必需字段
func (f *BodyFilter) validateRequiredFields(ctx *core.Context, body []byte) error {
	if requiredFields, ok := f.FilterConfig["required_fields"].([]interface{}); ok {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			for _, field := range requiredFields {
				if fieldName, ok := field.(string); ok {
					if _, exists := jsonData[fieldName]; !exists {
						return fmt.Errorf("缺少必需字段: %s", fieldName)
					}
				}
			}
		}
	}
	return f.setBody(ctx, body)
}

// validateMaxSize 验证最大大小
func (f *BodyFilter) validateMaxSize(ctx *core.Context, body []byte) error {
	if maxSize, ok := f.FilterConfig["max_size"].(int64); ok {
		if int64(len(body)) > maxSize {
			return fmt.Errorf("请求体大小超过限制: %d bytes", len(body))
		}
	}
	return f.setBody(ctx, body)
}

// 修改方法

// addFields 添加字段
func (f *BodyFilter) addFields(ctx *core.Context, body []byte) error {
	if addFields, ok := f.FilterConfig["add_fields"].(map[string]interface{}); ok {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			// 添加字段
			for key, value := range addFields {
				jsonData[key] = value
			}
			
			// 重新序列化
			newBody, err := json.Marshal(jsonData)
			if err != nil {
				return fmt.Errorf("序列化JSON失败: %w", err)
			}
			return f.setBody(ctx, newBody)
		}
	}
	return f.setBody(ctx, body)
}

// removeFields 移除字段
func (f *BodyFilter) removeFields(ctx *core.Context, body []byte) error {
	if removeFields, ok := f.FilterConfig["remove_fields"].([]interface{}); ok {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			// 移除字段
			for _, field := range removeFields {
				if fieldName, ok := field.(string); ok {
					delete(jsonData, fieldName)
				}
			}
			
			// 重新序列化
			newBody, err := json.Marshal(jsonData)
			if err != nil {
				return fmt.Errorf("序列化JSON失败: %w", err)
			}
			return f.setBody(ctx, newBody)
		}
	}
	return f.setBody(ctx, body)
}

// replaceValues 替换值
func (f *BodyFilter) replaceValues(ctx *core.Context, body []byte) error {
	if replaceValues, ok := f.FilterConfig["replace_values"].(map[string]interface{}); ok {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			// 替换值
			for key, value := range replaceValues {
				if _, exists := jsonData[key]; exists {
					jsonData[key] = value
				}
			}
			
			// 重新序列化
			newBody, err := json.Marshal(jsonData)
			if err != nil {
				return fmt.Errorf("序列化JSON失败: %w", err)
			}
			return f.setBody(ctx, newBody)
		}
	}
	return f.setBody(ctx, body)
}

// filterJSONFields 过滤JSON字段
func (f *BodyFilter) filterJSONFields(ctx *core.Context, body []byte) error {
	if allowedFields, ok := f.FilterConfig["allowed_fields"].([]interface{}); ok {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			// 创建新的数据对象，只包含允许的字段
			filteredData := make(map[string]interface{})
			for _, field := range allowedFields {
				if fieldName, ok := field.(string); ok {
					if value, exists := jsonData[fieldName]; exists {
						filteredData[fieldName] = value
					}
				}
			}
			
			// 重新序列化
			newBody, err := json.Marshal(filteredData)
			if err != nil {
				return fmt.Errorf("序列化JSON失败: %w", err)
			}
			return f.setBody(ctx, newBody)
		}
	}
	return f.setBody(ctx, body)
}

// configureBodyFilter 配置请求体过滤器
func configureBodyFilter(bodyFilter *BodyFilter, config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	// 设置操作类型
	if operation, ok := config["operation"].(string); ok {
		bodyFilter.Operation = operation
		bodyFilter.ModifierType = BodyModifierType(operation)
	}

	// 设置允许的内容类型
	if contentTypes, ok := config["allowed_content_types"].([]interface{}); ok {
		bodyFilter.AllowedContentTypes = make([]string, len(contentTypes))
		for i, ct := range contentTypes {
			if ctStr, ok := ct.(string); ok {
				bodyFilter.AllowedContentTypes[i] = ctStr
			}
		}
	}

	// 设置最大请求体大小
	if maxSize, ok := config["max_body_size"].(int64); ok {
		bodyFilter.MaxBodySize = maxSize
	} else if maxSizeInt, ok := config["max_body_size"].(int); ok {
		bodyFilter.MaxBodySize = int64(maxSizeInt)
	}

	// 存储完整配置
	bodyFilter.FilterConfig = config

	return nil
} 