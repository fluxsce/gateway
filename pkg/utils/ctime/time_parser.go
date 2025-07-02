// Package time 提供时间处理相关的工具函数
//
// 本包主要提供以下功能：
// 1. 时间字符串解析 - 支持多种常见的时间格式
// 2. 时间格式化 - 统一的时间格式化输出
// 3. 时间计算 - 常用的时间计算操作
// 4. 时区处理 - 时区转换和处理
//
// 使用示例：
//
//	// 解析时间字符串
//	t, err := time.ParseTimeString("2006-01-02 15:04:05")
//
//	// 格式化时间
//	str := time.FormatTime(time.Now(), time.FormatDateTime)
package ctime

import (
	"fmt"
	"time"
)

// 常用时间格式常量
const (
	FormatDateTime     = "2006-01-02 15:04:05"     // YYYY-MM-DD HH:mm:ss
	FormatDate         = "2006-01-02"              // YYYY-MM-DD
	FormatTimeOnly     = "15:04:05"                // HH:mm:ss
	FormatDateTimeSlash = "2006/01/02 15:04:05"    // YYYY/MM/DD HH:mm:ss
	FormatDateSlash    = "2006/01/02"              // YYYY/MM/DD
	FormatISO8601      = "2006-01-02T15:04:05"     // YYYY-MM-DDTHH:mm:ss
	FormatISO8601Z     = "2006-01-02T15:04:05Z"    // YYYY-MM-DDTHH:mm:ssZ
	FormatISO8601Milli = "2006-01-02T15:04:05.000Z" // YYYY-MM-DDTHH:mm:ss.sssZ
)

// 默认时区常量
const (
	DefaultTimezone = "Local" // 默认使用本地时区
	UTCTimezone     = "UTC"   // UTC时区
	ShanghaiTimezone = "Asia/Shanghai" // 上海时区（中国标准时间）
)

// ParseTimeString 解析时间字符串为time.Time类型
// 支持多种常见的时间格式，自动识别格式进行解析
// 对于没有时区信息的时间字符串，使用本地时区进行解析
//
// 支持的格式：
// - YYYY-MM-DD HH:mm:ss
// - YYYY-MM-DD
// - YYYY/MM/DD HH:mm:ss
// - YYYY/MM/DD
// - YYYY-MM-DDTHH:mm:ss
// - YYYY-MM-DDTHH:mm:ssZ
// - YYYY-MM-DDTHH:mm:ss.sssZ
// - RFC3339格式
//
// 参数:
//   timeStr: 要解析的时间字符串
// 返回:
//   time.Time: 解析后的时间对象
//   error: 解析失败时返回错误信息
//
// 使用示例:
//   t, err := ParseTimeString("2006-01-02 15:04:05")
//   if err != nil {
//       log.Printf("时间解析失败: %v", err)
//   }
func ParseTimeString(timeStr string) (time.Time, error) {
	// 使用本地时区作为默认时区
	return ParseTimeStringInLocation(timeStr, time.Local)
}

// ParseTimeStringInLocation 在指定时区解析时间字符串
// 对于没有时区信息的时间字符串，将使用指定的时区进行解析
//
// 参数:
//   timeStr: 要解析的时间字符串
//   location: 时区信息，可以使用time.LoadLocation()获取，如果为nil则使用本地时区
// 返回:
//   time.Time: 解析后的时间对象
//   error: 解析失败时返回错误信息
//
// 使用示例:
//   loc, _ := time.LoadLocation("Asia/Shanghai")
//   t, err := ParseTimeStringInLocation("2006-01-02 15:04:05", loc)
func ParseTimeStringInLocation(timeStr string, location *time.Location) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("时间字符串不能为空")
	}
	
	if location == nil {
		location = time.Local // 默认使用本地时区而不是UTC
	}

	// 定义支持的时间格式，按常用程度排序
	timeFormats := []string{
		FormatDateTime,     // 2006-01-02 15:04:05
		FormatDate,         // 2006-01-02
		FormatISO8601,      // 2006-01-02T15:04:05
		FormatISO8601Z,     // 2006-01-02T15:04:05Z
		FormatISO8601Milli, // 2006-01-02T15:04:05.000Z
		FormatDateTimeSlash, // 2006/01/02 15:04:05
		FormatDateSlash,    // 2006/01/02
		time.RFC3339,       // 标准RFC3339格式
		time.RFC3339Nano,   // RFC3339纳秒格式
		time.RFC822,        // RFC822格式
		time.RFC1123,       // RFC1123格式
	}
	
	// 尝试每种格式进行解析
	for _, format := range timeFormats {
		// 对于包含时区信息的格式，使用time.Parse
		if format == FormatISO8601Z || format == FormatISO8601Milli || 
		   format == time.RFC3339 || format == time.RFC3339Nano || 
		   format == time.RFC822 || format == time.RFC1123 {
			if parsedTime, err := time.Parse(format, timeStr); err == nil {
				return parsedTime, nil
			}
		} else {
			// 对于不包含时区信息的格式，使用指定时区
			if parsedTime, err := time.ParseInLocation(format, timeStr, location); err == nil {
				return parsedTime, nil
			}
		}
	}
	
	// 如果所有格式都失败，返回详细错误信息
	return time.Time{}, fmt.Errorf("无法解析时间字符串 '%s'，支持的格式包括: %s, %s, %s 等", 
		timeStr, FormatDateTime, FormatDate, FormatISO8601)
}

// ParseTimeStringInTimezone 在指定时区名称解析时间字符串
// 这是ParseTimeStringInLocation的便捷版本，直接使用时区名称
//
// 参数:
//   timeStr: 要解析的时间字符串
//   timezone: 时区名称，如"Asia/Shanghai", "UTC", "Local"等，如果为空则使用本地时区
// 返回:
//   time.Time: 解析后的时间对象
//   error: 解析失败时返回错误信息
//
// 使用示例:
//   t, err := ParseTimeStringInTimezone("2006-01-02 15:04:05", "Asia/Shanghai")
func ParseTimeStringInTimezone(timeStr string, timezone string) (time.Time, error) {
	if timezone == "" {
		timezone = DefaultTimezone // 默认使用本地时区
	}
	
	// 特殊处理本地时区
	if timezone == "Local" || timezone == DefaultTimezone {
		return ParseTimeStringInLocation(timeStr, time.Local)
	}
	
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的时区名称 '%s': %w", timezone, err)
	}
	
	return ParseTimeStringInLocation(timeStr, location)
}

// ParseTimeStringInLocal 在本地时区解析时间字符串
// 这是ParseTimeStringInLocation的便捷版本，专门用于本地时区解析
//
// 参数:
//   timeStr: 要解析的时间字符串
// 返回:
//   time.Time: 解析后的时间对象
//   error: 解析失败时返回错误信息
//
// 使用示例:
//   t, err := ParseTimeStringInLocal("2006-01-02 15:04:05")
func ParseTimeStringInLocal(timeStr string) (time.Time, error) {
	return ParseTimeStringInLocation(timeStr, time.Local)
}

// ConvertTimeToTimezone 将时间转换到指定时区
// 用于时区转换操作
//
// 参数:
//   t: 要转换的时间
//   timezone: 目标时区名称
// 返回:
//   time.Time: 转换后的时间对象
//   error: 转换失败时返回错误信息
//
// 使用示例:
//   utcTime := time.Now().UTC()
//   shanghaiTime, err := ConvertTimeToTimezone(utcTime, "Asia/Shanghai")
func ConvertTimeToTimezone(t time.Time, timezone string) (time.Time, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的时区名称 '%s': %w", timezone, err)
	}
	
	return t.In(location), nil
}

// GetTimezoneOffset 获取指定时区相对于UTC的偏移量（秒）
// 用于计算时区差异
//
// 参数:
//   timezone: 时区名称
//   t: 参考时间（用于处理夏令时等情况）
// 返回:
//   int: 偏移量（秒），东时区为正数，西时区为负数
//   error: 获取失败时返回错误信息
//
// 使用示例:
//   offset, err := GetTimezoneOffset("Asia/Shanghai", time.Now())
//   // 中国时区通常返回 28800 (8小时 * 3600秒)
func GetTimezoneOffset(timezone string, t time.Time) (int, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, fmt.Errorf("无效的时区名称 '%s': %w", timezone, err)
	}
	
	_, offset := t.In(location).Zone()
	return offset, nil
}

// FormatTime 格式化时间为指定格式的字符串
// 提供统一的时间格式化接口
//
// 参数:
//   t: 要格式化的时间对象
//   format: 时间格式，可以使用本包定义的常量或自定义格式
// 返回:
//   string: 格式化后的时间字符串
//
// 使用示例:
//   str := FormatTime(time.Now(), FormatDateTime)
//   // 输出: "2006-01-02 15:04:05"
func FormatTime(t time.Time, format string) string {
	return t.Format(format)
}

// ParseTimeStringWithDefault 解析时间字符串，失败时返回默认值
// 这是ParseTimeString的安全版本，不会返回错误
//
// 参数:
//   timeStr: 要解析的时间字符串
//   defaultTime: 解析失败时返回的默认时间
// 返回:
//   time.Time: 解析成功返回解析结果，失败返回默认值
//
// 使用示例:
//   t := ParseTimeStringWithDefault("invalid", time.Now())
func ParseTimeStringWithDefault(timeStr string, defaultTime time.Time) time.Time {
	if parsedTime, err := ParseTimeString(timeStr); err == nil {
		return parsedTime
	}
	return defaultTime
}

// ParseTimeStringWithDefaultInTimezone 在指定时区解析时间字符串，失败时返回默认值
// 这是ParseTimeStringInTimezone的安全版本，不会返回错误
//
// 参数:
//   timeStr: 要解析的时间字符串
//   timezone: 时区名称
//   defaultTime: 解析失败时返回的默认时间
// 返回:
//   time.Time: 解析成功返回解析结果，失败返回默认值
//
// 使用示例:
//   t := ParseTimeStringWithDefaultInTimezone("2006-01-02 15:04:05", "Asia/Shanghai", time.Now())
func ParseTimeStringWithDefaultInTimezone(timeStr string, timezone string, defaultTime time.Time) time.Time {
	if parsedTime, err := ParseTimeStringInTimezone(timeStr, timezone); err == nil {
		return parsedTime
	}
	return defaultTime
}

// IsValidTimeString 检查时间字符串是否有效
// 快速验证时间字符串格式是否正确
//
// 参数:
//   timeStr: 要检查的时间字符串
// 返回:
//   bool: true表示格式有效，false表示格式无效
//
// 使用示例:
//   if IsValidTimeString("2006-01-02 15:04:05") {
//       // 时间格式有效
//   }
func IsValidTimeString(timeStr string) bool {
	_, err := ParseTimeString(timeStr)
	return err == nil
}

// GetCurrentTimeString 获取当前时间的字符串表示
// 使用指定格式返回当前时间
//
// 参数:
//   format: 时间格式，可以使用本包定义的常量
// 返回:
//   string: 当前时间的字符串表示
//
// 使用示例:
//   now := GetCurrentTimeString(FormatDateTime)
//   // 输出类似: "2023-06-30 14:30:25"
func GetCurrentTimeString(format string) string {
	return time.Now().Format(format)
}

// GetCurrentTimeStringInTimezone 获取指定时区当前时间的字符串表示
// 使用指定格式和时区返回当前时间
//
// 参数:
//   format: 时间格式，可以使用本包定义的常量
//   timezone: 时区名称，如"Asia/Shanghai", "UTC", "Local"等，如果为空则使用本地时区
// 返回:
//   string: 指定时区当前时间的字符串表示
//   error: 时区无效时返回错误
//
// 使用示例:
//   now, err := GetCurrentTimeStringInTimezone(FormatDateTime, "Asia/Shanghai")
//   // 输出类似: "2023-06-30 22:30:25"
func GetCurrentTimeStringInTimezone(format string, timezone string) (string, error) {
	if timezone == "" {
		timezone = DefaultTimezone // 默认使用本地时区
	}
	
	// 特殊处理本地时区
	if timezone == "Local" || timezone == DefaultTimezone {
		return time.Now().Format(format), nil
	}
	
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("无效的时区名称 '%s': %w", timezone, err)
	}
	
	return time.Now().In(location).Format(format), nil
}

// BeginOfDay 获取指定日期的开始时间（00:00:00）
// 将时间设置为当天的0点0分0秒
//
// 参数:
//   t: 指定的时间
// 返回:
//   time.Time: 当天的开始时间
func BeginOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取指定日期的结束时间（23:59:59）
// 将时间设置为当天的23点59分59秒
//
// 参数:
//   t: 指定的时间
// 返回:
//   time.Time: 当天的结束时间
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
} 