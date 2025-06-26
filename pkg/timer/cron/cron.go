// Package cron 提供Cron表达式解析功能
package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronParser Cron表达式解析器接口
type CronParser interface {
	Parse(expr string) (CronSchedule, error)
}

// CronSchedule Cron调度接口
type CronSchedule interface {
	Next(t time.Time) time.Time
}

// StandardCronParser 标准Cron解析器
// 支持6字段格式：秒 分钟 小时 日 月 周
// 也支持5字段格式：分钟 小时 日 月 周（为了向后兼容）
type StandardCronParser struct{}

// NewStandardCronParser 创建标准Cron解析器实例
// 支持标准6字段Cron表达式格式：秒 分钟 小时 日 月 周
// 也支持5字段格式：分钟 小时 日 月 周（为了向后兼容）
// 返回:
//   *StandardCronParser: 初始化的Cron解析器实例
func NewStandardCronParser() *StandardCronParser {
	return &StandardCronParser{}
}

// Parse 解析Cron表达式字符串
// 支持6字段Cron表达式（秒 分钟 小时 日 月 周）和5字段格式（分钟 小时 日 月 周）
// 支持通配符(*)、范围(1-5)、列表(1,3,5)、步长(*/2)等语法
// 参数:
//   expr: Cron表达式字符串，格式为"秒 分钟 小时 日 月 周"或"分钟 小时 日 月 周"
// 返回:
//   CronSchedule: 解析后的调度对象，用于计算下次执行时间
//   error: 解析失败时返回错误信息
func (p *StandardCronParser) Parse(expr string) (CronSchedule, error) {
	fields := strings.Fields(expr)
	
	var second, minute, hour, day, month, weekday []int
	var err error
	
	if len(fields) == 6 {
		// 6字段格式：秒 分钟 小时 日 月 周
		second, err = parseField(fields[0], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid second field: %v", err)
		}
		
		minute, err = parseField(fields[1], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid minute field: %v", err)
		}
		
		hour, err = parseField(fields[2], 0, 23)
		if err != nil {
			return nil, fmt.Errorf("invalid hour field: %v", err)
		}
		
		day, err = parseField(fields[3], 1, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid day field: %v", err)
		}
		
		month, err = parseField(fields[4], 1, 12)
		if err != nil {
			return nil, fmt.Errorf("invalid month field: %v", err)
		}
		
		weekday, err = parseField(fields[5], 0, 6)
		if err != nil {
			return nil, fmt.Errorf("invalid weekday field: %v", err)
		}
	} else if len(fields) == 5 {
		// 5字段格式：分钟 小时 日 月 周（为了向后兼容）
		second = []int{0} // 默认在0秒执行
		
		minute, err = parseField(fields[0], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid minute field: %v", err)
		}
		
		hour, err = parseField(fields[1], 0, 23)
		if err != nil {
			return nil, fmt.Errorf("invalid hour field: %v", err)
		}
		
		day, err = parseField(fields[2], 1, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid day field: %v", err)
		}
		
		month, err = parseField(fields[3], 1, 12)
		if err != nil {
			return nil, fmt.Errorf("invalid month field: %v", err)
		}
		
		weekday, err = parseField(fields[4], 0, 6)
		if err != nil {
			return nil, fmt.Errorf("invalid weekday field: %v", err)
		}
	} else {
		return nil, fmt.Errorf("invalid cron expression: expected 5 or 6 fields, got %d", len(fields))
	}
	
	return &StandardCronSchedule{
		second:  second,
		minute:  minute,
		hour:    hour,
		day:     day,
		month:   month,
		weekday: weekday,
	}, nil
}

// StandardCronSchedule 标准Cron调度实现
type StandardCronSchedule struct {
	second  []int
	minute  []int
	hour    []int
	day     []int
	month   []int
	weekday []int
}

// Next 计算下次执行时间
// 基于当前时间和Cron规则计算下一次任务应该执行的时间
// 参数:
//   t: 当前时间，作为计算的起点
// 返回:
//   time.Time: 下次执行时间，如果找不到匹配时间则返回零值
func (s *StandardCronSchedule) Next(t time.Time) time.Time {
	// 从下一秒开始计算
	next := t.Add(time.Second).Truncate(time.Second)
	
	// 最多向前搜索4年（以秒为单位）
	for i := 0; i < 4*365*24*60*60; i++ {
		if s.matches(next) {
			return next
		}
		next = next.Add(time.Second)
	}
	
	// 如果找不到匹配的时间，返回零值
	return time.Time{}
}

// matches 检查时间是否匹配Cron表达式
func (s *StandardCronSchedule) matches(t time.Time) bool {
	return s.matchesField(s.second, t.Second()) &&
		s.matchesField(s.minute, t.Minute()) &&
		s.matchesField(s.hour, t.Hour()) &&
		s.matchesField(s.day, t.Day()) &&
		s.matchesField(s.month, int(t.Month())) &&
		s.matchesField(s.weekday, int(t.Weekday()))
}

// matchesField 检查字段是否匹配
func (s *StandardCronSchedule) matchesField(field []int, value int) bool {
	for _, v := range field {
		if v == value {
			return true
		}
	}
	return false
}

// parseField 解析Cron字段
func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		return makeRange(min, max), nil
	}
	
	var result []int
	
	// 处理逗号分隔的值
	parts := strings.Split(field, ",")
	for _, part := range parts {
		values, err := parseFieldPart(part, min, max)
		if err != nil {
			return nil, err
		}
		result = append(result, values...)
	}
	
	return result, nil
}

// parseFieldPart 解析字段的一部分
func parseFieldPart(part string, min, max int) ([]int, error) {
	// 处理步长 (例如: */2, 1-5/2)
	if strings.Contains(part, "/") {
		return parseStepValue(part, min, max)
	}
	
	// 处理范围 (例如: 1-5)
	if strings.Contains(part, "-") {
		return parseRange(part, min, max)
	}
	
	// 处理单个值
	value, err := strconv.Atoi(part)
	if err != nil {
		return nil, fmt.Errorf("invalid value: %s", part)
	}
	
	if value < min || value > max {
		return nil, fmt.Errorf("value %d out of range [%d, %d]", value, min, max)
	}
	
	return []int{value}, nil
}

// parseStepValue 解析步长值
func parseStepValue(part string, min, max int) ([]int, error) {
	parts := strings.Split(part, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid step value: %s", part)
	}
	
	step, err := strconv.Atoi(parts[1])
	if err != nil || step <= 0 {
		return nil, fmt.Errorf("invalid step: %s", parts[1])
	}
	
	var baseRange []int
	if parts[0] == "*" {
		baseRange = makeRange(min, max)
	} else {
		baseRange, err = parseFieldPart(parts[0], min, max)
		if err != nil {
			return nil, err
		}
	}
	
	var result []int
	for i, value := range baseRange {
		if i%step == 0 {
			result = append(result, value)
		}
	}
	
	return result, nil
}

// parseRange 解析范围值
func parseRange(part string, min, max int) ([]int, error) {
	parts := strings.Split(part, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range: %s", part)
	}
	
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid range start: %s", parts[0])
	}
	
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid range end: %s", parts[1])
	}
	
	if start < min || start > max || end < min || end > max {
		return nil, fmt.Errorf("range [%d, %d] out of bounds [%d, %d]", start, end, min, max)
	}
	
	if start > end {
		return nil, fmt.Errorf("invalid range: start %d > end %d", start, end)
	}
	
	return makeRange(start, end), nil
}

// makeRange 创建范围数组
func makeRange(start, end int) []int {
	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}

// 预定义的常用Cron表达式（6字段格式：秒 分钟 小时 日 月 周）
var (
	// EverySecond 每秒执行
	EverySecond = "* * * * * *"
	// EveryMinute 每分钟执行
	EveryMinute = "0 * * * * *"
	// Hourly 每小时执行
	Hourly = "0 0 * * * *"
	// Daily 每天执行
	Daily = "0 0 0 * * *"
	// Weekly 每周执行
	Weekly = "0 0 0 * * 0"
	// Monthly 每月执行
	Monthly = "0 0 0 1 * *"
	// Yearly 每年执行
	Yearly = "0 0 0 1 1 *"
)

// ParseCron 解析Cron表达式的便捷函数
func ParseCron(expr string) (CronSchedule, error) {
	parser := NewStandardCronParser()
	return parser.Parse(expr)
} 