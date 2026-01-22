package service

import "time"

// parseDuration 解析时间字符串，解析失败返回默认值
func parseDuration(s string, defaultValue time.Duration) time.Duration {
	if s == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultValue
	}
	return d
}

// getStringValue 获取字符串指针的值
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// stringPtr 创建字符串指针
func stringPtr(s string) *string {
	return &s
}
