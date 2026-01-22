package channel

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"gateway/pkg/alert"
)

// TemplateReplacer 模板替换工具类
// 支持替换模板中的 {{}} 包裹的字段内容
type TemplateReplacer struct {
	// 预编译的正则表达式，用于匹配 {{field}} 格式
	placeholderRegex *regexp.Regexp
}

// NewTemplateReplacer 创建模板替换器
func NewTemplateReplacer() *TemplateReplacer {
	// 匹配 {{field}} 或 {{field.subfield}} 格式
	// 支持字段名和子字段，如 {{title}}, {{tag.severity}}
	regex := regexp.MustCompile(`\{\{([a-zA-Z0-9_.]+)\}\}`)
	return &TemplateReplacer{
		placeholderRegex: regex,
	}
}

// Replace 替换模板中的占位符
// 参数:
//   - template: 模板字符串，包含 {{field}} 格式的占位符
//   - message: 告警消息，提供替换数据
//   - customData: 自定义数据映射，可以覆盖或扩展默认字段
//
// 返回:
//   - string: 替换后的字符串
//
// 支持的占位符:
//   - {{title}}: 消息标题
//   - {{content}}: 消息内容
//   - {{timestamp}}: 时间戳（格式：2006-01-02 15:04:05）
//   - {{timestamp_iso}}: ISO格式时间戳（格式：2006-01-02T15:04:05Z07:00）
//   - {{timestamp_unix}}: Unix时间戳
//   - {{tags}}: 所有标签（格式：key1: value1 | key2: value2）
//   - {{tag.key}}: 特定标签的值，如 {{tag.severity}}
//   - {{extra.key}}: 额外数据中的值，如 {{extra.custom_field}}
//   - {{table.key}}: 表格数据中的特定字段，如 {{table.service_name}}
//   - {{table}}: 所有表格数据（格式：key1: value1 | key2: value2）
func (tr *TemplateReplacer) Replace(template string, message *alert.Message, customData map[string]interface{}) string {
	if template == "" {
		return ""
	}

	// 构建数据映射
	data := tr.buildDataMap(message, customData)

	// 替换所有占位符
	result := tr.placeholderRegex.ReplaceAllStringFunc(template, func(match string) string {
		// 提取字段名（去掉 {{ 和 }}）
		fieldName := match[2 : len(match)-2]
		fieldName = strings.TrimSpace(fieldName)

		// 查找对应的值
		if value, ok := data[fieldName]; ok {
			return fmt.Sprintf("%v", value)
		}

		// 如果找不到，返回原占位符（不替换）
		return match
	})

	return result
}

// buildDataMap 构建数据映射
func (tr *TemplateReplacer) buildDataMap(message *alert.Message, customData map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})

	// 基本字段
	data["title"] = message.Title
	data["content"] = message.Content

	// 时间戳相关
	if !message.Timestamp.IsZero() {
		data["timestamp"] = message.Timestamp.Format("2006-01-02 15:04:05")
		data["timestamp_iso"] = message.Timestamp.Format(time.RFC3339)
		data["timestamp_unix"] = message.Timestamp.Unix()
		data["timestamp_unix_ms"] = message.Timestamp.UnixMilli()
	} else {
		now := time.Now()
		data["timestamp"] = now.Format("2006-01-02 15:04:05")
		data["timestamp_iso"] = now.Format(time.RFC3339)
		data["timestamp_unix"] = now.Unix()
		data["timestamp_unix_ms"] = now.UnixMilli()
	}

	// 标签相关
	if len(message.Tags) > 0 {
		// 所有标签的格式化字符串
		tagParts := make([]string, 0, len(message.Tags))
		for k, v := range message.Tags {
			tagParts = append(tagParts, fmt.Sprintf("%s: %s", k, v))
		}
		data["tags"] = strings.Join(tagParts, " | ")

		// 每个标签作为独立字段（tag.key 格式）
		for k, v := range message.Tags {
			data["tag."+k] = v
		}
	} else {
		data["tags"] = ""
	}

	// 额外数据
	if message.Extra != nil {
		for k, v := range message.Extra {
			// 跳过 send_config，因为它是特殊用途
			if k == "send_config" {
				continue
			}
			data["extra."+k] = v
		}
	}

	// 表格数据
	if len(message.TableData) > 0 {
		// 所有表格数据的格式化字符串
		tableParts := make([]string, 0, len(message.TableData))
		for k, v := range message.TableData {
			tableParts = append(tableParts, fmt.Sprintf("%s: %v", k, v))
		}
		data["table"] = strings.Join(tableParts, " | ")

		// 每个表格字段作为独立字段（table.key 格式）
		for k, v := range message.TableData {
			data["table."+k] = v
		}
	} else {
		data["table"] = ""
	}

	// 自定义数据（可以覆盖默认字段）
	if customData != nil {
		for k, v := range customData {
			data[k] = v
		}
	}

	return data
}

// GetPlaceholders 获取模板中所有的占位符
// 参数:
//   - template: 模板字符串
//
// 返回:
//   - []string: 占位符列表（不包含 {{ 和 }}）
func (tr *TemplateReplacer) GetPlaceholders(template string) []string {
	if template == "" {
		return nil
	}

	matches := tr.placeholderRegex.FindAllStringSubmatch(template, -1)
	placeholders := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			placeholders = append(placeholders, strings.TrimSpace(match[1]))
		}
	}

	return placeholders
}

// ValidateTemplate 验证模板是否有效
// 检查模板中的占位符格式是否正确
// 参数:
//   - template: 模板字符串
//
// 返回:
//   - bool: 模板是否有效
//   - []string: 无效的占位符列表
func (tr *TemplateReplacer) ValidateTemplate(template string) (bool, []string) {
	if template == "" {
		return true, nil
	}

	invalid := make([]string, 0)

	// 检查是否有未闭合的占位符
	if strings.Count(template, "{{") != strings.Count(template, "}}") {
		invalid = append(invalid, "未闭合的占位符")
	}

	// 检查是否有嵌套的占位符（不支持）
	if strings.Contains(template, "{{{{") {
		invalid = append(invalid, "检测到嵌套占位符（不支持）")
	}

	return len(invalid) == 0, invalid
}
