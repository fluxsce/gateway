package cron

import (
	"testing"
	"time"
)

func TestStandardCronParser_Parse6Fields(t *testing.T) {
	parser := NewStandardCronParser()
	
	tests := []struct {
		name     string
		expr     string
		wantErr  bool
	}{
		{
			name:    "每秒执行",
			expr:    "* * * * * *",
			wantErr: false,
		},
		{
			name:    "每分钟执行",
			expr:    "0 * * * * *",
			wantErr: false,
		},
		{
			name:    "每5秒执行",
			expr:    "*/5 * * * * *",
			wantErr: false,
		},
		{
			name:    "每小时执行",
			expr:    "0 0 * * * *",
			wantErr: false,
		},
		{
			name:    "工作日上午9点",
			expr:    "0 0 9 * * 1-5",
			wantErr: false,
		},
		{
			name:    "无效字段数量",
			expr:    "* * *",
			wantErr: true,
		},
		{
			name:    "无效秒值",
			expr:    "60 * * * * *",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := parser.Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && schedule == nil {
				t.Error("Parse() returned nil schedule without error")
			}
		})
	}
}

func TestStandardCronParser_Parse5Fields(t *testing.T) {
	parser := NewStandardCronParser()
	
	tests := []struct {
		name     string
		expr     string
		wantErr  bool
	}{
		{
			name:    "每分钟执行（5字段）",
			expr:    "* * * * *",
			wantErr: false,
		},
		{
			name:    "每小时执行（5字段）",
			expr:    "0 * * * *",
			wantErr: false,
		},
		{
			name:    "每天执行（5字段）",
			expr:    "0 0 * * *",
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := parser.Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && schedule == nil {
				t.Error("Parse() returned nil schedule without error")
			}
		})
	}
}

func TestStandardCronSchedule_Next(t *testing.T) {
	parser := NewStandardCronParser()
	
	// 测试每5秒执行的任务
	schedule, err := parser.Parse("*/5 * * * * *")
	if err != nil {
		t.Fatalf("Failed to parse cron expression: %v", err)
	}
	
	now := time.Now()
	next := schedule.Next(now)
	
	if next.IsZero() {
		t.Error("Next() returned zero time")
	}
	
	if next.Before(now) {
		t.Error("Next() returned time in the past")
	}
	
	// 检查下次执行时间的秒数是否是5的倍数
	if next.Second()%5 != 0 {
		t.Errorf("Next() returned time with second %d, expected multiple of 5", next.Second())
	}
}

func TestPredefinedExpressions(t *testing.T) {
	parser := NewStandardCronParser()
	
	expressions := map[string]string{
		"EverySecond": EverySecond,
		"EveryMinute": EveryMinute,
		"Hourly":      Hourly,
		"Daily":       Daily,
		"Weekly":      Weekly,
		"Monthly":     Monthly,
		"Yearly":      Yearly,
	}
	
	for name, expr := range expressions {
		t.Run(name, func(t *testing.T) {
			schedule, err := parser.Parse(expr)
			if err != nil {
				t.Errorf("Failed to parse %s expression '%s': %v", name, expr, err)
			}
			if schedule == nil {
				t.Errorf("Parse() returned nil schedule for %s", name)
			}
		})
	}
} 