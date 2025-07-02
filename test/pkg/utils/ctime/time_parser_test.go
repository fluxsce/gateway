package ctime

import (
	"fmt"
	"gohub/pkg/utils/ctime"
	"testing"
	"time"
)

func TestTimeParser(t *testing.T) {
	fmt.Println("=== 时区解析测试（默认本地时区）===")
	
	// 测试1: 原始解析（现在使用本地时区）
	timeStr := "2025-06-30 15:04:05"
	t1, err := ctime.ParseTimeString(timeStr)
	if err != nil {
		fmt.Printf("原始解析失败: %v\n", err)
	} else {
		fmt.Printf("原始解析结果（本地时区）: %s (时区: %s)\n", t1.Format("2006-01-02 15:04:05 MST"), t1.Location())
	}
	
	// 测试2: 显式本地时区解析
	t2, err := ctime.ParseTimeStringInLocal(timeStr)
	if err != nil {
		fmt.Printf("本地时区解析失败: %v\n", err)
	} else {
		fmt.Printf("本地时区解析结果: %s (时区: %s)\n", t2.Format("2006-01-02 15:04:05 MST"), t2.Location())
	}
	
	// 测试3: 使用默认时区常量
	t3, err := ctime.ParseTimeStringInTimezone(timeStr, ctime.DefaultTimezone)
	if err != nil {
		fmt.Printf("默认时区解析失败: %v\n", err)
	} else {
		fmt.Printf("默认时区解析结果: %s (时区: %s)\n", t3.Format("2006-01-02 15:04:05 MST"), t3.Location())
	}
	
	// 测试4: 在上海时区解析
	t4, err := ctime.ParseTimeStringInTimezone(timeStr, ctime.ShanghaiTimezone)
	if err != nil {
		fmt.Printf("上海时区解析失败: %v\n", err)
	} else {
		fmt.Printf("上海时区解析结果: %s (时区: %s)\n", t4.Format("2006-01-02 15:04:05 MST"), t4.Location())
	}
	
	// 测试5: UTC时区解析
	t5, err := ctime.ParseTimeStringInTimezone(timeStr, ctime.UTCTimezone)
	if err != nil {
		fmt.Printf("UTC时区解析失败: %v\n", err)
	} else {
		fmt.Printf("UTC时区解析结果: %s (时区: %s)\n", t5.Format("2006-01-02 15:04:05 MST"), t5.Location())
	}
	
	fmt.Println("\n=== 时间转换测试 ===")
	
	// 测试6: UTC时间转换到上海时区
	utcTime := time.Now().UTC()
	shanghaiTime, err := ctime.ConvertTimeToTimezone(utcTime, ctime.ShanghaiTimezone)
	if err != nil {
		fmt.Printf("时区转换失败: %v\n", err)
	} else {
		fmt.Printf("UTC时间: %s\n", utcTime.Format("2006-01-02 15:04:05 MST"))
		fmt.Printf("上海时间: %s\n", shanghaiTime.Format("2006-01-02 15:04:05 MST"))
	}
	
	// 测试7: 获取时区偏移量
	offset, err := ctime.GetTimezoneOffset(ctime.ShanghaiTimezone, time.Now())
	if err != nil {
		fmt.Printf("获取时区偏移失败: %v\n", err)
	} else {
		fmt.Printf("上海时区偏移: %d秒 (%.1f小时)\n", offset, float64(offset)/3600)
	}
	
	fmt.Println("\n=== 当前时间获取测试 ===")
	
	// 测试8: 获取当前时间在不同时区（使用默认参数）
	localNow, err := ctime.GetCurrentTimeStringInTimezone(ctime.FormatDateTime, "")
	if err != nil {
		fmt.Printf("获取本地当前时间失败: %v\n", err)
	} else {
		fmt.Printf("本地当前时间（默认）: %s\n", localNow)
	}
	
	// 测试9: 获取上海当前时间
	shanghaiNow, err := ctime.GetCurrentTimeStringInTimezone(ctime.FormatDateTime, ctime.ShanghaiTimezone)
	if err != nil {
		fmt.Printf("获取上海当前时间失败: %v\n", err)
	} else {
		fmt.Printf("上海当前时间: %s\n", shanghaiNow)
	}
	
	// 测试10: 获取UTC当前时间
	utcNow, err := ctime.GetCurrentTimeStringInTimezone(ctime.FormatDateTime, ctime.UTCTimezone)
	if err != nil {
		fmt.Printf("获取UTC当前时间失败: %v\n", err)
	} else {
		fmt.Printf("UTC当前时间: %s\n", utcNow)
	}
	
	fmt.Println("\n=== UTC时间字符串解析测试 ===")
	
	// 测试11: 解析UTC时间字符串
	utcTimeStr := "2025-06-30T15:04:05Z"
	t11, err := ctime.ParseTimeStringInTimezone(utcTimeStr, ctime.ShanghaiTimezone)
	if err != nil {
		fmt.Printf("UTC时间字符串解析失败: %v\n", err)
	} else {
		fmt.Printf("UTC时间字符串解析结果: %s (时区: %s)\n", t11.Format("2006-01-02 15:04:05 MST"), t11.Location())
	}
	
	// 测试12: 比较同一时间在不同时区的解析结果
	fmt.Println("\n=== 同一时间不同时区解析比较 ===")
	localTime, _ := ctime.ParseTimeStringInTimezone("2025-06-30 15:04:05", ctime.DefaultTimezone)
	shanghaiTimeFromLocal, _ := ctime.ParseTimeStringInTimezone("2025-06-30 15:04:05", ctime.ShanghaiTimezone)
	utcTimeFromLocal, _ := ctime.ParseTimeStringInTimezone("2025-06-30 15:04:05", ctime.UTCTimezone)
	
	fmt.Printf("本地时区解析: %s (Unix: %d)\n", localTime.Format("2006-01-02 15:04:05 MST"), localTime.Unix())
	fmt.Printf("上海时区解析: %s (Unix: %d)\n", shanghaiTimeFromLocal.Format("2006-01-02 15:04:05 MST"), shanghaiTimeFromLocal.Unix())
	fmt.Printf("UTC时区解析: %s (Unix: %d)\n", utcTimeFromLocal.Format("2006-01-02 15:04:05 MST"), utcTimeFromLocal.Unix())
	
	fmt.Println("\n=== 总结 ===")
	fmt.Printf("时区解析功能已更新，默认使用本地时区: %s\n", ctime.DefaultTimezone)
	fmt.Printf("支持的时区常量: DefaultTimezone=%s, UTCTimezone=%s, ShanghaiTimezone=%s\n", 
		ctime.DefaultTimezone, ctime.UTCTimezone, ctime.ShanghaiTimezone)
} 