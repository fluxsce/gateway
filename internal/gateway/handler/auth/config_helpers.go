package auth

// configString 从配置 map 中按候选键顺序读取非空字符串。
// 用于兼容前端界面（camelCase）与 YAML/历史配置（snake_case）两种字段命名。
func configString(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// configBool 从配置 map 中按候选键顺序读取布尔值。
func configBool(m map[string]interface{}, keys ...string) (bool, bool) {
	for _, key := range keys {
		if v, ok := m[key].(bool); ok {
			return v, true
		}
	}
	return false, false
}

// configInt 从配置 map 中按候选键顺序读取整数值。
func configInt(m map[string]interface{}, keys ...string) (int, bool) {
	for _, key := range keys {
		switch v := m[key].(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		}
	}
	return 0, false
}
