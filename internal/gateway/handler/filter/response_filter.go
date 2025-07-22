package filter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gateway/internal/gateway/core"
)

// ResponseOperation 响应操作类型
type ResponseOperation string

const (
	// AddHeaders 添加响应头
	AddHeaders ResponseOperation = "add_headers"

	// ModifyResponseBody 修改响应体
	ModifyResponseBody ResponseOperation = "modify_body"

	// SetStatus 设置响应状态码
	SetStatus ResponseOperation = "set_status"

	// FilterHeaders 过滤响应头
	FilterHeaders ResponseOperation = "filter_headers"

	// TransformResponseBody 转换响应体
	TransformResponseBody ResponseOperation = "transform_body"

	// ValidateResponseData 验证响应
	ValidateResponseData ResponseOperation = "validate_response"
)

// ResponseFilter 响应过滤器
// 用于处理和修改HTTP响应
type ResponseFilter struct {
	BaseFilter

	// 响应操作类型
	Operation ResponseOperation

	// 配置参数
	FilterConfig map[string]interface{}

	// 响应头操作配置
	HeaderOperations map[string]interface{}

	// 响应体操作配置
	BodyOperations map[string]interface{}

	// 状态码操作配置
	StatusOperations map[string]interface{}

	// 是否在请求阶段设置（将在响应阶段应用）
	SetInRequestPhase bool

	// 条件配置
	Conditions map[string]interface{}
}

// ResponseModification 响应修改信息
type ResponseModification struct {
	// 要添加的响应头
	AddHeaders map[string]string

	// 要移除的响应头
	RemoveHeaders []string

	// 要修改的响应头
	ModifyHeaders map[string]string

	// 新的状态码
	StatusCode int

	// 新的响应体
	Body []byte

	// 响应体修改类型
	BodyModificationType string

	// 条件匹配结果
	ConditionMet bool
}

// ResponseFilterFromConfig 从配置创建响应过滤器
func ResponseFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	responseFilter := NewResponseFilter(config.Name, action, order)
	responseFilter.originalConfig = config

	// 从配置中提取响应操作参数
	if err := configureResponseFilter(responseFilter, config.Config); err != nil {
		return nil, fmt.Errorf("配置响应过滤器失败: %w", err)
	}

	return responseFilter, nil
}

// NewResponseFilter 创建响应过滤器
func NewResponseFilter(name string, action FilterAction, priority int) *ResponseFilter {
	baseFilter := NewBaseFilter(ResponseFilterType, action, priority, true, name)
	return &ResponseFilter{
		BaseFilter:        *baseFilter,
		Operation:         AddHeaders,
		FilterConfig:      make(map[string]interface{}),
		HeaderOperations:  make(map[string]interface{}),
		BodyOperations:    make(map[string]interface{}),
		StatusOperations:  make(map[string]interface{}),
		SetInRequestPhase: true, // 默认在请求阶段设置，响应阶段应用
		Conditions:        make(map[string]interface{}),
	}
}

// Apply 实现Filter接口
func (f *ResponseFilter) Apply(ctx *core.Context) error {
	if ctx.Request == nil {
		return fmt.Errorf("request is nil")
	}

	// 响应过滤器通常在请求阶段设置配置，在响应阶段应用
	if f.SetInRequestPhase {
		return f.setResponseModifications(ctx)
	}

	// 如果不是在请求阶段设置，则直接应用（这种情况较少见）
	return f.applyResponseModifications(ctx)
}

// setResponseModifications 在请求阶段设置响应修改配置
func (f *ResponseFilter) setResponseModifications(ctx *core.Context) error {
	// 检查条件是否满足
	if !f.checkConditions(ctx) {
		ctx.Set("response_filter_condition_not_met", true)
		return nil
	}

	modification := &ResponseModification{
		AddHeaders:    make(map[string]string),
		RemoveHeaders: make([]string, 0),
		ModifyHeaders: make(map[string]string),
		ConditionMet:  true,
	}

	// 根据操作类型设置相应的修改
	switch f.Operation {
	case AddHeaders:
		f.setHeaderAdditions(modification)
	case ModifyResponseBody:
		f.setBodyModifications(modification)
	case SetStatus:
		f.setStatusModifications(modification)
	case FilterHeaders:
		f.setHeaderFiltering(modification)
	case TransformResponseBody:
		f.setBodyTransformations(modification)
	case ValidateResponseData:
		f.setResponseValidation(modification)
	}

	// 将修改信息存储在上下文中
	f.storeModificationInContext(ctx, modification)

	// 记录操作
	ctx.Set("response_filter_applied", true)
	ctx.Set("response_filter_name", f.Name)
	ctx.Set("response_operation", string(f.Operation))

	return nil
}

// applyResponseModifications 应用响应修改（在响应阶段调用）
func (f *ResponseFilter) applyResponseModifications(ctx *core.Context) error {
	// 从上下文中获取修改信息
	modificationsVal, exists := ctx.Get("response_modifications")
	if !exists || modificationsVal == nil {
		return nil // 没有需要应用的修改
	}
	if modificationsVal == nil {
		return nil // 没有需要应用的修改
	}

	modifications, ok := modificationsVal.([]ResponseModification)
	if !ok {
		return nil // 类型不匹配
	}

	// 应用所有修改
	for _, modification := range modifications {
		if !modification.ConditionMet {
			continue
		}

		// 应用响应头修改
		f.applyHeaderModifications(ctx, &modification)

		// 应用状态码修改
		if modification.StatusCode > 0 {
			ctx.Set("response_status_code", modification.StatusCode)
		}

		// 应用响应体修改
		if len(modification.Body) > 0 {
			ctx.Set("response_body", modification.Body)
			ctx.Set("response_body_type", modification.BodyModificationType)
		}
	}

	return nil
}

// checkConditions 检查应用条件
func (f *ResponseFilter) checkConditions(ctx *core.Context) bool {
	if len(f.Conditions) == 0 {
		return true // 没有条件限制，总是应用
	}

	// 检查请求方法条件
	if methods, ok := f.Conditions["methods"].([]interface{}); ok {
		methodMatch := false
		for _, method := range methods {
			if methodStr, ok := method.(string); ok && strings.EqualFold(ctx.Request.Method, methodStr) {
				methodMatch = true
				break
			}
		}
		if !methodMatch {
			return false
		}
	}

	// 检查路径条件
	if paths, ok := f.Conditions["paths"].([]interface{}); ok {
		pathMatch := false
		for _, path := range paths {
			if pathStr, ok := path.(string); ok && strings.Contains(ctx.Request.URL.Path, pathStr) {
				pathMatch = true
				break
			}
		}
		if !pathMatch {
			return false
		}
	}

	// 检查请求头条件
	if headers, ok := f.Conditions["headers"].(map[string]interface{}); ok {
		for headerName, expectedValue := range headers {
			actualValue := ctx.Request.Header.Get(headerName)
			if expectedValueStr, ok := expectedValue.(string); ok {
				if actualValue != expectedValueStr {
					return false
				}
			}
		}
	}

	// 检查查询参数条件
	if params, ok := f.Conditions["query_params"].(map[string]interface{}); ok {
		for paramName, expectedValue := range params {
			actualValue := ctx.Request.URL.Query().Get(paramName)
			if expectedValueStr, ok := expectedValue.(string); ok {
				if actualValue != expectedValueStr {
					return false
				}
			}
		}
	}

	return true
}

// setHeaderAdditions 设置响应头添加
func (f *ResponseFilter) setHeaderAdditions(modification *ResponseModification) {
	if headers, ok := f.HeaderOperations["add_headers"].(map[string]interface{}); ok {
		for name, value := range headers {
			if valueStr, ok := value.(string); ok {
				modification.AddHeaders[name] = valueStr
			}
		}
	}

	if headers, ok := f.FilterConfig["add_headers"].(map[string]interface{}); ok {
		for name, value := range headers {
			if valueStr, ok := value.(string); ok {
				modification.AddHeaders[name] = valueStr
			}
		}
	}
}

// setBodyModifications 设置响应体修改
func (f *ResponseFilter) setBodyModifications(modification *ResponseModification) {
	if bodyConfig, ok := f.BodyOperations["modify_body"].(map[string]interface{}); ok {
		if bodyType, ok := bodyConfig["type"].(string); ok {
			modification.BodyModificationType = bodyType
		}

		if bodyContent, ok := bodyConfig["content"].(string); ok {
			modification.Body = []byte(bodyContent)
		}
	}

	if bodyConfig, ok := f.FilterConfig["modify_body"].(map[string]interface{}); ok {
		if bodyType, ok := bodyConfig["type"].(string); ok {
			modification.BodyModificationType = bodyType
		}

		if bodyContent, ok := bodyConfig["content"].(string); ok {
			modification.Body = []byte(bodyContent)
		}
	}
}

// setStatusModifications 设置状态码修改
func (f *ResponseFilter) setStatusModifications(modification *ResponseModification) {
	if status, ok := f.StatusOperations["status_code"].(int); ok {
		modification.StatusCode = status
	} else if statusFloat, ok := f.StatusOperations["status_code"].(float64); ok {
		modification.StatusCode = int(statusFloat)
	}

	if status, ok := f.FilterConfig["status_code"].(int); ok {
		modification.StatusCode = status
	} else if statusFloat, ok := f.FilterConfig["status_code"].(float64); ok {
		modification.StatusCode = int(statusFloat)
	}
}

// setHeaderFiltering 设置响应头过滤
func (f *ResponseFilter) setHeaderFiltering(modification *ResponseModification) {
	if removeHeaders, ok := f.HeaderOperations["remove_headers"].([]interface{}); ok {
		for _, header := range removeHeaders {
			if headerStr, ok := header.(string); ok {
				modification.RemoveHeaders = append(modification.RemoveHeaders, headerStr)
			}
		}
	}

	if removeHeaders, ok := f.FilterConfig["remove_headers"].([]interface{}); ok {
		for _, header := range removeHeaders {
			if headerStr, ok := header.(string); ok {
				modification.RemoveHeaders = append(modification.RemoveHeaders, headerStr)
			}
		}
	}
}

// setBodyTransformations 设置响应体转换
func (f *ResponseFilter) setBodyTransformations(modification *ResponseModification) {
	if transformConfig, ok := f.BodyOperations["transform_body"].(map[string]interface{}); ok {
		modification.BodyModificationType = "transform"

		if transformType, ok := transformConfig["type"].(string); ok {
			// 这里可以根据转换类型设置不同的处理逻辑
			switch transformType {
			case "json_prettify":
				modification.BodyModificationType = "json_prettify"
			case "json_minify":
				modification.BodyModificationType = "json_minify"
			case "xml_format":
				modification.BodyModificationType = "xml_format"
			}
		}
	}
}

// setResponseValidation 设置响应验证
func (f *ResponseFilter) setResponseValidation(modification *ResponseModification) {
	// 响应验证通常不修改响应，而是在验证失败时设置错误状态
	if validationConfig, ok := f.FilterConfig["validation"].(map[string]interface{}); ok {
		if failureStatus, ok := validationConfig["failure_status"].(int); ok {
			modification.StatusCode = failureStatus
		}

		if failureBody, ok := validationConfig["failure_body"].(string); ok {
			modification.Body = []byte(failureBody)
			modification.BodyModificationType = "validation_error"
		}
	}
}

// storeModificationInContext 将修改信息存储在上下文中
func (f *ResponseFilter) storeModificationInContext(ctx *core.Context, modification *ResponseModification) {
	// 获取现有的修改列表
	var modifications []ResponseModification
	existingVal, exists := ctx.Get("response_modifications")
	if exists && existingVal != nil {
		if existing, ok := existingVal.([]ResponseModification); ok {
			modifications = existing
		}
	}

	// 添加新的修改
	modifications = append(modifications, *modification)
	ctx.Set("response_modifications", modifications)
}

// applyHeaderModifications 应用响应头修改
func (f *ResponseFilter) applyHeaderModifications(ctx *core.Context, modification *ResponseModification) {
	// 添加响应头
	for name, value := range modification.AddHeaders {
		ctx.Set("response_header_"+name, value)
	}

	// 移除响应头
	for _, name := range modification.RemoveHeaders {
		ctx.Set("response_header_remove_"+name, true)
	}

	// 修改响应头
	for name, value := range modification.ModifyHeaders {
		ctx.Set("response_header_"+name, value)
	}
}

// TransformResponseBody 转换响应体（工具方法）
func (f *ResponseFilter) TransformResponseBody(body []byte, transformType string) ([]byte, error) {
	switch transformType {
	case "json_prettify":
		return f.prettifyJSON(body)
	case "json_minify":
		return f.minifyJSON(body)
	case "xml_format":
		return f.formatXML(body)
	case "uppercase":
		return bytes.ToUpper(body), nil
	case "lowercase":
		return bytes.ToLower(body), nil
	default:
		return body, nil
	}
}

// prettifyJSON 美化JSON
func (f *ResponseFilter) prettifyJSON(body []byte) ([]byte, error) {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body, err // 不是有效JSON，返回原内容
	}

	prettified, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return body, err
	}

	return prettified, nil
}

// minifyJSON 压缩JSON
func (f *ResponseFilter) minifyJSON(body []byte) ([]byte, error) {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body, err // 不是有效JSON，返回原内容
	}

	minified, err := json.Marshal(data)
	if err != nil {
		return body, err
	}

	return minified, nil
}

// formatXML 格式化XML（简化实现）
func (f *ResponseFilter) formatXML(body []byte) ([]byte, error) {
	// 这里应该使用XML库进行格式化
	// 简化实现，直接返回原内容
	return body, nil
}

// ValidateResponse 验证响应（工具方法）
func (f *ResponseFilter) ValidateResponse(ctx *core.Context, statusCode int, headers http.Header, body []byte) error {
	validationRules, ok := f.FilterConfig["validation_rules"].(map[string]interface{})
	if !ok {
		return nil // 没有验证规则
	}

	// 验证状态码
	if expectedStatus, ok := validationRules["expected_status"].(int); ok {
		if statusCode != expectedStatus {
			return fmt.Errorf("响应状态码不匹配，期望: %d, 实际: %d", expectedStatus, statusCode)
		}
	}

	// 验证响应头
	if requiredHeaders, ok := validationRules["required_headers"].([]interface{}); ok {
		for _, header := range requiredHeaders {
			if headerStr, ok := header.(string); ok {
				if headers.Get(headerStr) == "" {
					return fmt.Errorf("缺少必需的响应头: %s", headerStr)
				}
			}
		}
	}

	// 验证响应体大小
	if maxSize, ok := validationRules["max_body_size"].(int); ok {
		if len(body) > maxSize {
			return fmt.Errorf("响应体大小超过限制: %d bytes", len(body))
		}
	}

	// 验证JSON格式
	if validateJSON, ok := validationRules["validate_json"].(bool); ok && validateJSON {
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			return fmt.Errorf("响应体不是有效的JSON格式: %w", err)
		}
	}

	return nil
}

// SetOperation 设置操作类型
func (f *ResponseFilter) SetOperation(operation ResponseOperation) *ResponseFilter {
	f.Operation = operation
	return f
}

// AddHeader 添加响应头
func (f *ResponseFilter) AddHeader(name, value string) *ResponseFilter {
	if f.HeaderOperations["add_headers"] == nil {
		f.HeaderOperations["add_headers"] = make(map[string]interface{})
	}
	f.HeaderOperations["add_headers"].(map[string]interface{})[name] = value
	return f
}

// RemoveHeader 移除响应头
func (f *ResponseFilter) RemoveHeader(name string) *ResponseFilter {
	if f.HeaderOperations["remove_headers"] == nil {
		f.HeaderOperations["remove_headers"] = make([]interface{}, 0)
	}
	headers := f.HeaderOperations["remove_headers"].([]interface{})
	f.HeaderOperations["remove_headers"] = append(headers, name)
	return f
}

// SetStatusCode 设置状态码
func (f *ResponseFilter) SetStatusCode(code int) *ResponseFilter {
	f.StatusOperations["status_code"] = code
	return f
}

// SetBody 设置响应体
func (f *ResponseFilter) SetBody(body string, bodyType string) *ResponseFilter {
	if f.BodyOperations["modify_body"] == nil {
		f.BodyOperations["modify_body"] = make(map[string]interface{})
	}
	bodyConfig := f.BodyOperations["modify_body"].(map[string]interface{})
	bodyConfig["content"] = body
	bodyConfig["type"] = bodyType
	return f
}

// AddCondition 添加应用条件
func (f *ResponseFilter) AddCondition(key string, value interface{}) *ResponseFilter {
	f.Conditions[key] = value
	return f
}

// configureResponseFilter 配置响应过滤器
func configureResponseFilter(responseFilter *ResponseFilter, config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	// 首先检查是否有嵌套的 responseConfig 配置
	var responseConfig map[string]interface{}
	if nestedConfig, ok := config["responseConfig"].(map[string]interface{}); ok {
		responseConfig = nestedConfig
	} else {
		// 如果没有嵌套配置，直接使用顶级配置
		responseConfig = config
	}

	// 优先支持前端驼峰命名格式
	if operation, ok := responseConfig["operation"].(string); ok {
		// 驼峰命名配置处理
		filterConfig, _ := responseConfig["filterConfig"].(map[string]interface{})
		headerOperations, _ := responseConfig["headerOperations"].(map[string]interface{})
		bodyOperations, _ := responseConfig["bodyOperations"].(map[string]interface{})
		statusOperations, _ := responseConfig["statusOperations"].(map[string]interface{})
		conditions, _ := responseConfig["conditions"].(map[string]interface{})

		// 参数验证
		if operation == "" {
			return fmt.Errorf("operation 不能为空")
		}

		// 设置操作类型
		switch strings.ToLower(operation) {
		case "add_headers", "modify_body", "set_status", "filter_headers", "transform_body", "validate_response":
			responseFilter.Operation = ResponseOperation(strings.ToLower(operation))
		default:
			return fmt.Errorf("无效的operation: %s，支持的类型: add_headers, modify_body, set_status, filter_headers, transform_body, validate_response", operation)
		}

		// 设置是否在请求阶段设置
		if setInRequestPhase, ok := responseConfig["setInRequestPhase"].(bool); ok {
			responseFilter.SetInRequestPhase = setInRequestPhase
		}

		// 设置过滤器配置
		if filterConfig != nil {
			responseFilter.FilterConfig = filterConfig
		} else {
			responseFilter.FilterConfig = make(map[string]interface{})
		}

		// 设置响应头操作
		if headerOperations != nil {
			responseFilter.HeaderOperations = headerOperations
		}

		// 设置响应体操作
		if bodyOperations != nil {
			responseFilter.BodyOperations = bodyOperations
		}

		// 设置状态码操作
		if statusOperations != nil {
			responseFilter.StatusOperations = statusOperations
		}

		// 设置条件
		if conditions != nil {
			responseFilter.Conditions = conditions
		}

		return nil
	}

	// 兼容旧的下划线命名格式
	// 设置操作类型
	if operation, ok := responseConfig["operation"].(string); ok {
		responseFilter.Operation = ResponseOperation(operation)
	}

	// 设置是否在请求阶段设置
	if setInRequest, ok := responseConfig["set_in_request_phase"].(bool); ok {
		responseFilter.SetInRequestPhase = setInRequest
	}

	// 设置响应头操作
	if headerOps, ok := responseConfig["header_operations"].(map[string]interface{}); ok {
		responseFilter.HeaderOperations = headerOps
	}

	// 设置响应体操作
	if bodyOps, ok := responseConfig["body_operations"].(map[string]interface{}); ok {
		responseFilter.BodyOperations = bodyOps
	}

	// 设置状态码操作
	if statusOps, ok := responseConfig["status_operations"].(map[string]interface{}); ok {
		responseFilter.StatusOperations = statusOps
	}

	// 设置条件
	if conditions, ok := responseConfig["conditions"].(map[string]interface{}); ok {
		responseFilter.Conditions = conditions
	}

	// 存储完整配置
	responseFilter.FilterConfig = responseConfig

	return nil
}
