package cron

import (
	"testing"
	"time"

	"gohub/pkg/timer/cron"
)

func TestStandardCronParser_Parse(t *testing.T) {
	parser := cron.NewStandardCronParser()

	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"6字段有效表达式", "0 * * * * *", false},
		{"5字段有效表达式", "* * * * *", false},
		{"无效字段数", "* * *", true},
		{"无效秒值", "60 * * * * *", true},
		{"无效分钟值", "0 60 * * * *", true},
		{"无效小时值", "0 0 24 * * *", true},
		{"无效日期值", "0 0 0 32 * *", true},
		{"无效月份值", "0 0 0 1 13 *", true},
		{"无效星期值", "0 0 0 * * 7", true},
		{"范围表达式", "0-5 * * * * *", false},
		{"列表表达式", "1,2,3 * * * * *", false},
		{"步长表达式", "*/5 * * * * *", false},
		{"复杂表达式", "1-5/2 * * * * *", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStandardCronSchedule_Next(t *testing.T) {
	parser := cron.NewStandardCronParser()

	tests := []struct {
		name     string
		expr     string
		from     time.Time
		expected time.Time
	}{
		{
			name: "每分钟",
			expr: "0 * * * * *",
			from: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC),
		},
		{
			name: "每小时",
			expr: "0 0 * * * *",
			from: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			name: "每天午夜",
			expr: "0 0 0 * * *",
			from: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "每周日",
			expr: "0 0 0 * * 0",
			from: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "每月1号",
			expr: "0 0 0 1 * *",
			from: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := parser.Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			got := schedule.Next(tt.from)
			if !got.Equal(tt.expected) {
				t.Errorf("Next() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPredefinedExpressions(t *testing.T) {
	parser := cron.NewStandardCronParser()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		expr     string
		expected time.Time
	}{
		{
			name: "EverySecond",
			expr: cron.EverySecond,
			expected: now.Add(time.Second),
		},
		{
			name: "EveryMinute",
			expr: cron.EveryMinute,
			expected: time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC),
		},
		{
			name: "Hourly",
			expr: cron.Hourly,
			expected: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			name: "Daily",
			expr: cron.Daily,
			expected: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Monthly",
			expr: cron.Monthly,
			expected: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := parser.Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			got := schedule.Next(now)
			if !got.Equal(tt.expected) {
				t.Errorf("Next() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseFieldErrors(t *testing.T) {
	parser := cron.NewStandardCronParser()

	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"无效数字", "abc * * * * *", true},
		{"超出范围", "60 * * * * *", true},
		{"无效范围格式", "1- * * * * *", true},
		{"范围起点大于终点", "5-3 * * * * *", true},
		{"无效步长格式", "*/ * * * * *", true},
		{"步长为零", "*/0 * * * * *", true},
		{"步长为负", "*/-1 * * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkCronParsing(b *testing.B) {
	parser := cron.NewStandardCronParser()
	expr := "*/5 * * * * *"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(expr)
	}
}

func BenchmarkNextExecution(b *testing.B) {
	parser := cron.NewStandardCronParser()
	schedule, _ := parser.Parse("*/5 * * * * *")
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schedule.Next(now)
	}
}

// TestSpecificCronExpression 测试特定的cron表达式 "0 0 9 * * 1" (每周一上午9点)
func TestSpecificCronExpression(t *testing.T) {
	parser := cron.NewStandardCronParser()
	expr := "0 0 9 * * 1" // 每周一上午9点

	// 测试解析是否成功
	schedule, err := parser.Parse(expr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tests := []struct {
		name     string
		from     time.Time
		expected time.Time
	}{
		{
			name: "从周五到下周一",
			from: time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC), // 2024-01-05 是周五
			expected: time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC), // 2024-01-08 是周一
		},
		{
			name: "从周一9点前到周一9点",
			from: time.Date(2024, 1, 8, 8, 30, 0, 0, time.UTC), // 2024-01-08 周一 8:30
			expected: time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC), // 2024-01-08 周一 9:00
		},
		{
			name: "从周一9点后到下周一9点",
			from: time.Date(2024, 1, 8, 9, 30, 0, 0, time.UTC), // 2024-01-08 周一 9:30
			expected: time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC), // 2024-01-15 下周一 9:00
		},
		{
			name: "从周三到下周一",
			from: time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC), // 2024-01-10 周三
			expected: time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC), // 2024-01-15 周一
		},
		{
			name: "从周日到周一",
			from: time.Date(2024, 1, 7, 22, 0, 0, 0, time.UTC), // 2024-01-07 周日
			expected: time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC), // 2024-01-08 周一
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schedule.Next(tt.from)
			if !got.Equal(tt.expected) {
				t.Errorf("Next() = %v, want %v", got, tt.expected)
				t.Errorf("From: %v (%s)", tt.from, tt.from.Weekday())
				t.Errorf("Got:  %v (%s)", got, got.Weekday())
				t.Errorf("Want: %v (%s)", tt.expected, tt.expected.Weekday())
			}
		})
	}

	// 测试连续的几次执行时间
	t.Run("连续执行时间", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // 2024-01-01 周一
		var executions []time.Time
		
		current := start
		for i := 0; i < 5; i++ {
			next := schedule.Next(current)
			executions = append(executions, next)
			current = next.Add(time.Second) // 添加1秒避免重复
		}

		expected := []time.Time{
			time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),  // 2024-01-01 周一 9:00
			time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC),  // 2024-01-08 周一 9:00
			time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC), // 2024-01-15 周一 9:00
			time.Date(2024, 1, 22, 9, 0, 0, 0, time.UTC), // 2024-01-22 周一 9:00
			time.Date(2024, 1, 29, 9, 0, 0, 0, time.UTC), // 2024-01-29 周一 9:00
		}

		for i, exec := range executions {
			if !exec.Equal(expected[i]) {
				t.Errorf("执行时间[%d] = %v, want %v", i, exec, expected[i])
			}
		}

		// 打印执行时间以便查看
		t.Logf("Cron表达式 '%s' 的连续执行时间:", expr)
		for i, exec := range executions {
			t.Logf("  %d: %v (%s)", i+1, exec, exec.Weekday())
		}
	})
} 