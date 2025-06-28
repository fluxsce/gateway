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