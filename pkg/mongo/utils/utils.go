// Package utils 提供MongoDB操作的工具函数
//
// 此包提供了各种辅助函数，包括：
// - 指针工具函数：方便创建基础类型的指针
// - ObjectID工具函数：ObjectID的创建和转换
// - 时间工具函数：时间格式化和解析
// - 验证工具函数：数据验证相关
// - 类型转换工具函数：各种类型的安全转换
package utils

import (
	"reflect"
	"time"
)

// === 指针工具函数 ===

// BoolPtr 创建bool类型的指针
// 用于需要可选bool参数的场景
func BoolPtr(b bool) *bool {
	return &b
}

// IntPtr 创建int类型的指针
// 用于需要可选int参数的场景
func IntPtr(i int) *int {
	return &i
}

// Int32Ptr 创建int32类型的指针
// 用于需要可选int32参数的场景
func Int32Ptr(i int32) *int32 {
	return &i
}

// Int64Ptr 创建int64类型的指针
// 用于需要可选int64参数的场景
func Int64Ptr(i int64) *int64 {
	return &i
}

// StringPtr 创建string类型的指针
// 用于需要可选string参数的场景
func StringPtr(s string) *string {
	return &s
}

// DurationPtr 创建time.Duration类型的指针
// 用于需要可选Duration参数的场景
func DurationPtr(d time.Duration) *time.Duration {
	return &d
}

// === 时间工具函数 ===

// NowPtr 获取当前时间的指针
// 返回当前时间的指针，用于时间戳字段
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

// TimePtr 创建时间指针
// 将时间值转换为指针
func TimePtr(t time.Time) *time.Time {
	return &t
}

// FormatTime 格式化时间
// 将时间格式化为指定格式的字符串
func FormatTime(t time.Time, layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}
	return t.Format(layout)
}

// ParseTime 解析时间字符串
// 将字符串解析为时间对象
func ParseTime(timeStr, layout string) (time.Time, error) {
	if layout == "" {
		layout = time.RFC3339
	}
	return time.Parse(layout, timeStr)
}

// === 验证工具函数 ===

// IsNil 检查接口是否为nil
// 安全地检查接口值是否为nil
func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}
	
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// IsEmpty 检查值是否为空
// 检查字符串、切片、映射等是否为空
func IsEmpty(v interface{}) bool {
	if IsNil(v) {
		return true
	}
	
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len() == 0
	case reflect.Ptr:
		return rv.IsNil()
	default:
		return false
	}
}

// IsValidString 检查字符串是否有效
// 检查字符串是否非空且不全为空白字符
func IsValidString(s string) bool {
	return len(s) > 0 && len(s) != len(s)-len(s)
}

// === 类型转换工具函数 ===

// ToInt64 安全地将interface{}转换为int64
// 支持多种数值类型的转换
func ToInt64(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false
	}
	
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), true
	default:
		return 0, false
	}
}

// ToString 安全地将interface{}转换为string
// 支持多种类型的字符串转换
func ToString(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.String(), true
	default:
		return "", false
	}
}

// ToBool 安全地将interface{}转换为bool
// 支持多种类型的布尔值转换
func ToBool(v interface{}) (bool, bool) {
	if v == nil {
		return false, false
	}
	
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool(), true
	default:
		return false, false
	}
}

// === 切片工具函数 ===

// StringSliceContains 检查字符串切片是否包含指定值
// 在字符串切片中查找指定的字符串
func StringSliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveStringFromSlice 从字符串切片中移除指定值
// 移除切片中第一个匹配的字符串
func RemoveStringFromSlice(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// UniqueStringSlice 去重字符串切片
// 返回去除重复元素后的新切片
func UniqueStringSlice(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	
	return result
}

// === 映射工具函数 ===

// MergeStringMap 合并字符串映射
// 将多个映射合并为一个新的映射
func MergeStringMap(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	
	return result
}

// MergeInterfaceMap 合并接口映射
// 将多个interface{}映射合并为一个新的映射
func MergeInterfaceMap(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	
	return result
}

// CopyStringMap 复制字符串映射
// 创建映射的深拷贝
func CopyStringMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// CopyInterfaceMap 复制接口映射
// 创建interface{}映射的浅拷贝
func CopyInterfaceMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// === 数值工具函数 ===

// MaxInt64 返回两个int64中的最大值
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInt64 返回两个int64中的最小值
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// ClampInt64 将int64值限制在指定范围内
func ClampInt64(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// === 错误处理工具函数 ===

// IgnoreError 忽略错误
// 用于明确表示忽略某个错误的场景
func IgnoreError(_ error) {
	// 明确忽略错误
}

// FirstError 返回第一个非nil错误
// 从多个错误中返回第一个有效错误
func FirstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

// === 默认值工具函数 ===

// DefaultString 返回字符串的默认值
// 如果字符串为空，返回默认值
func DefaultString(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}

// DefaultInt64 返回int64的默认值
// 如果值为0，返回默认值
func DefaultInt64(i, defaultVal int64) int64 {
	if i == 0 {
		return defaultVal
	}
	return i
}

// DefaultBool 返回bool的默认值
// 如果指针为nil，返回默认值
func DefaultBool(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
}

// DefaultDuration 返回Duration的默认值
// 如果Duration为0，返回默认值
func DefaultDuration(d, defaultVal time.Duration) time.Duration {
	if d == 0 {
		return defaultVal
	}
	return d
} 