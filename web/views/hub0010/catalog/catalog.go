// Package catalog 是控制台内置工具目录。
// 工具随版本发布，不入库；角色是否可用由资源码 hub0010:tool:{id} 决定。
package catalog

import "fmt"

// ModuleCode 工具市场模块编码。
const ModuleCode = "hub0010"

// CategoryUtility 实用工具分类。
const CategoryUtility = "utility"

// ConfigField 工具可保存的一项配置。
type ConfigField struct {
	Key     string         `json:"key"`
	Type    string         `json:"type"`
	Options []ConfigOption `json:"options,omitempty"`
}

// ConfigOption 下拉配置的可选值。
type ConfigOption struct {
	Value any `json:"value"`
}

// Tool 一条内置工具。
type Tool struct {
	ID            string
	Category      string
	Version       string
	Author        string
	Icon          string
	Tags          []string
	ConfigSchema  []ConfigField
	DefaultConfig map[string]any
}

// ResourceCode 返回该工具在权限目录中的按钮码。
func ResourceCode(toolID string) string {
	return ModuleCode + ":tool:" + toolID
}

// All 返回内置工具，顺序即市场展示顺序。
func All() []Tool {
	return []Tool{
		{
			ID:       "json-formatter",
			Category: CategoryUtility,
			Version:  "1.0.0",
			Author:   "Gateway",
			Icon:     "code-outline",
			Tags:     []string{"JSON"},
			ConfigSchema: []ConfigField{{
				Key:  "indent",
				Type: "select",
				Options: []ConfigOption{
					{Value: 2},
					{Value: 4},
				},
			}},
			DefaultConfig: map[string]any{"indent": 2},
		},
		{
			ID:            "timestamp",
			Category:      CategoryUtility,
			Version:       "1.0.0",
			Author:        "Gateway",
			Icon:          "time-outline",
			Tags:          []string{"time"},
			DefaultConfig: map[string]any{},
		},
		{
			ID:            "base64",
			Category:      CategoryUtility,
			Version:       "1.0.0",
			Author:        "Gateway",
			Icon:          "key-outline",
			Tags:          []string{"base64"},
			DefaultConfig: map[string]any{},
		},
		{
			ID:            "jwt-decoder",
			Category:      CategoryUtility,
			Version:       "1.0.0",
			Author:        "Gateway",
			Icon:          "shield-outline",
			Tags:          []string{"JWT"},
			DefaultConfig: map[string]any{},
		},
	}
}

// Find 按 ID 查找工具。
func Find(toolID string) (Tool, bool) {
	for _, tool := range All() {
		if tool.ID == toolID {
			return tool, true
		}
	}
	return Tool{}, false
}

// NormalizeConfig 按工具模式整理用户配置。未知键、非法值返回错误。
func NormalizeConfig(tool Tool, raw map[string]any) (map[string]any, error) {
	if raw == nil {
		raw = map[string]any{}
	}
	if len(tool.ConfigSchema) == 0 {
		if len(raw) > 0 {
			return nil, fmt.Errorf("该工具没有可保存的配置")
		}
		return map[string]any{}, nil
	}
	out := map[string]any{}
	for key := range raw {
		if !schemaHas(tool, key) {
			return nil, fmt.Errorf("不支持的配置项: %s", key)
		}
	}
	for _, field := range tool.ConfigSchema {
		value, ok := raw[field.Key]
		if !ok {
			if def, has := tool.DefaultConfig[field.Key]; has {
				out[field.Key] = def
			}
			continue
		}
		normalized, err := normalizeField(field, value)
		if err != nil {
			return nil, err
		}
		out[field.Key] = normalized
	}
	return out, nil
}

func schemaHas(tool Tool, key string) bool {
	for _, field := range tool.ConfigSchema {
		if field.Key == key {
			return true
		}
	}
	return false
}

func normalizeField(field ConfigField, value any) (any, error) {
	if field.Type != "select" {
		return nil, fmt.Errorf("不支持的配置类型: %s", field.Type)
	}
	number, ok := asInt(value)
	if !ok {
		return nil, fmt.Errorf("配置项 %s 必须是整数", field.Key)
	}
	for _, option := range field.Options {
		allowed, ok := asInt(option.Value)
		if ok && allowed == number {
			return number, nil
		}
	}
	return nil, fmt.Errorf("配置项 %s 的值不在允许范围内", field.Key)
}

func asInt(value any) (int, bool) {
	switch n := value.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		if n != float64(int(n)) {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}
