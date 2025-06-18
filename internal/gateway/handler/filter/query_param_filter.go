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
func (f *QueryParamFilter) ConfigureRename(paramName, targetParamName string) *QueryParamFilter {
	f.ModifierType = RenameQueryParam
	f.ParamName = paramName
	f.TargetParamName = targetParamName
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
		// 设置查询参数（替换已有值）
		query.Set(f.ParamName, f.ParamValue)
	case RemoveQueryParam:
		// 移除查询参数
		query.Del(f.ParamName)
	case RenameQueryParam:
		// 重命名查询参数
		values := query[f.ParamName]
		if len(values) > 0 {
			// 复制值到新的参数
			for _, value := range values {
				query.Add(f.TargetParamName, value)
			}
			// 删除旧的参数
			query.Del(f.ParamName)
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

	// 单个参数操作配置
	if paramName, ok := config["param_name"].(string); ok {
		if paramValue, ok := config["param_value"].(string); ok {
			operation := "add"
			if op, ok := config["operation"].(string); ok {
				operation = strings.ToLower(op)
			}

			switch operation {
			case "add":
				queryFilter.ConfigureAdd(paramName, paramValue)
			case "set":
				queryFilter.ConfigureSet(paramName, paramValue)
			case "remove":
				queryFilter.ConfigureRemove(paramName)
			case "rename":
				if newName, ok := config["new_name"].(string); ok {
					queryFilter.ConfigureRename(paramName, newName)
				}
			default:
				queryFilter.ConfigureAdd(paramName, paramValue)
			}
		}
	}

	// 批量参数操作配置
	if params, ok := config["params"].(map[string]interface{}); ok {
		for name, value := range params {
			if valueStr, ok := value.(string); ok {
				queryFilter.ConfigureAdd(name, valueStr)
			}
		}
	}

	return nil
}
