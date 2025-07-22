package filter

import (
	"fmt"
	"gohub/internal/gateway/core"
	"strings"
)

// QueryParamModifierType 查询参数修改类型
type QueryParamModifierType string

const (
	// AddQueryParam 添加查询参数
	AddQueryParam QueryParamModifierType = "add"

	// SetQueryParam 设置查询参数（替换现有值）
	SetQueryParam QueryParamModifierType = "set"

	// RemoveQueryParam 移除查询参数
	RemoveQueryParam QueryParamModifierType = "remove"

	// RenameQueryParam 重命名查询参数
	RenameQueryParam QueryParamModifierType = "rename"
)

// QueryParamFilter 查询参数过滤器
// 用于修改URL查询参数
type QueryParamFilter struct {
	BaseFilter

	// 修改类型
	ModifierType QueryParamModifierType

	// 参数名称
	ParamName string

	// 参数值
	ParamValue string

	// 目标参数名称（用于RenameQueryParam）
	TargetParamName string
}

// QueryParamFilterFromConfig 从配置创建查询参数过滤器
func QueryParamFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	queryFilter := NewQueryParamFilter(config.Name, action, order)

	// 存储原始配置
	queryFilter.originalConfig = config

	// 从配置中提取参数操作
	if err := configureQueryParamFilter(queryFilter, config.Config); err != nil {
		return nil, fmt.Errorf("配置查询参数过滤器失败: %w", err)
	}

	return queryFilter, nil
}

// NewQueryParamFilter 创建查询参数过滤器
func NewQueryParamFilter(name string, action FilterAction, priority int) *QueryParamFilter {
	baseFilter := NewBaseFilter(QueryParamFilterType, action, priority, true, name)
	return &QueryParamFilter{
		BaseFilter: *baseFilter,
	}
}

// ConfigureAdd 配置为添加查询参数
func (f *QueryParamFilter) ConfigureAdd(paramName, paramValue string) *QueryParamFilter {
	f.ModifierType = AddQueryParam
	f.ParamName = paramName
	f.ParamValue = paramValue
	return f
}

// ConfigureSet 配置为设置查询参数
func (f *QueryParamFilter) ConfigureSet(paramName, paramValue string) *QueryParamFilter {
	f.ModifierType = SetQueryParam
	f.ParamName = paramName
	f.ParamValue = paramValue
	return f
}

// ConfigureRemove 配置为移除查询参数
func (f *QueryParamFilter) ConfigureRemove(paramName string) *QueryParamFilter {
	f.ModifierType = RemoveQueryParam
	f.ParamName = paramName
	return f
}

// ConfigureRename 配置为重命名查询参数
func (f *QueryParamFilter) ConfigureRename(paramName, targetParamName string, paramValue ...string) *QueryParamFilter {
	f.ModifierType = RenameQueryParam
	f.ParamName = paramName
	f.TargetParamName = targetParamName
	// 如果提供了新值，使用新值
	if len(paramValue) > 0 {
		f.ParamValue = paramValue[0]
	}
	return f
}

// Apply 实现Filter接口
func (f *QueryParamFilter) Apply(ctx *core.Context) error {
	req := ctx.Request

	// 解析查询参数
	query := req.URL.Query()

	// 修改查询参数
	switch f.ModifierType {
	case AddQueryParam:
		// 添加查询参数（不替换已有值）
		query.Add(f.ParamName, f.ParamValue)
	case SetQueryParam:
		// 设置查询参数（只对已存在的参数替换值）
		if _, exists := query[f.ParamName]; !exists {
			// 参数不存在，跳过设置操作
			return nil
		}
		query.Set(f.ParamName, f.ParamValue)
	case RemoveQueryParam:
		// 移除查询参数（只对已存在的参数执行删除）
		if _, exists := query[f.ParamName]; !exists {
			// 参数不存在，跳过删除操作
			return nil
		}
		query.Del(f.ParamName)
	case RenameQueryParam:
		// 重命名查询参数 - 只对已存在的参数执行
		values := query[f.ParamName]
		
		// 检查原参数是否存在
		if len(values) == 0 {
			// 原参数不存在，跳过重命名操作
			return nil
		}
		
		// 原参数存在，执行重命名
		// 删除旧的参数
		query.Del(f.ParamName)
		
		// 设置新参数
		if f.ParamValue != "" {
			// 如果提供了新值，使用新值
			query.Set(f.TargetParamName, f.ParamValue)
		} else {
			// 复制原有值到新的参数名
			for _, value := range values {
				query.Add(f.TargetParamName, value)
			}
		}
		
		// 支持同名参数修改（当目标参数名与源参数名相同时）
		if f.TargetParamName == f.ParamName && f.ParamValue != "" {
			query.Set(f.ParamName, f.ParamValue)
		}
	}

	// 更新URL查询字符串
	req.URL.RawQuery = query.Encode()

	return nil
}

// configureQueryParamFilter 配置查询参数过滤器
func configureQueryParamFilter(queryFilter *QueryParamFilter, config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	// 首先检查是否有嵌套的 queryParamConfig 配置
	var paramConfig map[string]interface{}
	if nestedConfig, ok := config["queryParamConfig"].(map[string]interface{}); ok {
		paramConfig = nestedConfig
	} else {
		// 如果没有嵌套配置，直接使用顶级配置
		paramConfig = config
	}

	// 优先支持前端驼峰命名格式
	if modifierType, ok := paramConfig["modifierType"].(string); ok {
		// 驼峰命名配置处理
		paramName, _ := paramConfig["paramName"].(string)
		paramValue, _ := paramConfig["paramValue"].(string)
		targetParamName, _ := paramConfig["targetParamName"].(string)
		
		// 参数验证
		if paramName == "" {
			return fmt.Errorf("paramName 不能为空")
		}
		
		// 根据操作类型配置
		switch strings.ToLower(modifierType) {
		case "add":
			if paramValue == "" {
				return fmt.Errorf("add操作需要paramValue参数")
			}
			queryFilter.ConfigureAdd(paramName, paramValue)
		case "set":
			if paramValue == "" {
				return fmt.Errorf("set操作需要paramValue参数")
			}
			queryFilter.ConfigureSet(paramName, paramValue)
		case "remove":
			queryFilter.ConfigureRemove(paramName)
		case "rename":
			if targetParamName == "" {
				return fmt.Errorf("rename操作需要targetParamName参数")
			}
			queryFilter.ConfigureRename(paramName, targetParamName, paramValue)
		default:
			return fmt.Errorf("无效的modifierType: %s", modifierType)
		}
		
		return nil
	}

	// 支持下划线命名格式 (modifier_type, param_name等)
	if modifierType, ok := paramConfig["modifier_type"].(string); ok {
		paramName, _ := paramConfig["param_name"].(string)
		paramValue, _ := paramConfig["param_value"].(string)
		targetParamName, _ := paramConfig["target_param_name"].(string)
		
		if paramName == "" {
			return fmt.Errorf("param_name 不能为空")
		}
		
		switch strings.ToLower(modifierType) {
		case "add":
			if paramValue == "" {
				return fmt.Errorf("add操作需要param_value参数")
			}
			queryFilter.ConfigureAdd(paramName, paramValue)
		case "set":
			if paramValue == "" {
				return fmt.Errorf("set操作需要param_value参数")
			}
			queryFilter.ConfigureSet(paramName, paramValue)
		case "remove":
			queryFilter.ConfigureRemove(paramName)
		case "rename":
			if targetParamName == "" {
				return fmt.Errorf("rename操作需要target_param_name参数")
			}
			queryFilter.ConfigureRename(paramName, targetParamName, paramValue)
		default:
			return fmt.Errorf("无效的modifier_type: %s", modifierType)
		}
		
		return nil
	}
	

	return nil
}
