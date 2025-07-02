package cron

import (
	"testing"
	"time"
)

func TestCronWithQuestionMark(t *testing.T) {
	parser := NewStandardCronParser()
	
	// 测试用户提到的表达式："0 * * * * ?"
	schedule, err := parser.Parse("0 * * * * ?")
	if err != nil {
		t.Fatalf("Failed to parse cron expression with ?: %v", err)
	}
	
	// 测试计算下次执行时间
	now := time.Date(2024, 1, 1, 12, 30, 45, 0, time.UTC)
	next := schedule.Next(now)
	
	// 应该在下一分钟的0秒执行
	expected := time.Date(2024, 1, 1, 12, 31, 0, 0, time.UTC)
	if next != expected {
		t.Errorf("Expected next time %v, got %v", expected, next)
	}
	
	t.Logf("Cron expression '0 * * * * ?' parsed successfully")
	t.Logf("Current time: %v", now)
	t.Logf("Next execution time: %v", next)
}

func TestCronWithMultipleQuestionMarks(t *testing.T) {
	parser := NewStandardCronParser()
	
	// 测试多个?字符
	testCases := []string{
		"0 0 12 ? * ?",  // 每天中午12点
		"0 15 10 ? * 1", // 每周一上午10:15
		"0 0 0 1 ? *",   // 每月1号
	}
	
	for _, expr := range testCases {
		schedule, err := parser.Parse(expr)
		if err != nil {
			t.Errorf("Failed to parse expression '%s': %v", expr, err)
			continue
		}
		
		// 验证能计算下次执行时间
		now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		next := schedule.Next(now)
		
		if next.IsZero() {
			t.Errorf("Expression '%s' returned zero time", expr)
		} else {
			t.Logf("Expression '%s' next execution: %v", expr, next)
		}
	}
} 