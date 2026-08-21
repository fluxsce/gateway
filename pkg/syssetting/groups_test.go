package syssetting

import "testing"

func TestValidateRetention(t *testing.T) {
	v := DefaultRetention()
	if err := ValidateRetention(v); err != nil {
		t.Fatalf("default retention should be valid: %v", err)
	}
	v.AuditLogDays = 0
	if err := ValidateRetention(v); err == nil {
		t.Fatal("expected error for auditLogDays=0")
	}
	v = DefaultRetention()
	v.MetricsDays = 4000
	if err := ValidateRetention(v); err == nil {
		t.Fatal("expected error for metricsDays>3650")
	}
}

func TestValidateWebTimeout(t *testing.T) {
	v := WebTimeoutSettings{RequestTimeoutSeconds: 120, SessionExpireHours: 12}
	if err := ValidateWebTimeout(v); err != nil {
		t.Fatalf("valid timeout rejected: %v", err)
	}
	v.RequestTimeoutSeconds = 5
	if err := ValidateWebTimeout(v); err == nil {
		t.Fatal("expected error for request timeout < 10")
	}
}

func TestParseRetentionFillsDefaults(t *testing.T) {
	got := ParseRetention(`{"auditLogDays":90}`)
	if got.AuditLogDays != 90 {
		t.Fatalf("auditLogDays=%d", got.AuditLogDays)
	}
	if got.TaskLogDays != DefaultRetention().TaskLogDays {
		t.Fatalf("taskLogDays should fall back, got %d", got.TaskLogDays)
	}
}

func TestParseRetentionJob(t *testing.T) {
	got := ParseRetentionJob(`{"enabled":false,"intervalMinutes":15,"startTime":"02:00"}`)
	if got.Enabled || got.IntervalMinutes != 15 || got.StartTime != "02:00" {
		t.Fatalf("got %+v", got)
	}
	def := ParseRetentionJob(`{}`)
	if !def.Enabled || def.IntervalMinutes != defaultJobIntervalMin || def.StartTime != "" {
		t.Fatalf("empty json should keep defaults, got %+v", def)
	}
}

func TestValidateRetentionJob(t *testing.T) {
	v := DefaultRetentionJob()
	if err := ValidateRetentionJob(v); err != nil {
		t.Fatalf("default job should be valid: %v", err)
	}
	v.IntervalMinutes = 0
	if err := ValidateRetentionJob(v); err == nil {
		t.Fatal("expected error for intervalMinutes=0")
	}
	v = DefaultRetentionJob()
	v.StartTime = "25:00"
	if err := ValidateRetentionJob(v); err == nil {
		t.Fatal("expected error for invalid startTime")
	}
}

func TestStorePutAndGet(t *testing.T) {
	PutRetention("t1", RetentionSettings{AuditLogDays: 60, TaskLogDays: 10, AlertLogDays: 3, ClusterEventDays: 2, MetricsDays: 15, GatewayLogDefaultDays: 20})
	got := GetRetention("t1")
	if got.AuditLogDays != 60 {
		t.Fatalf("got %+v", got)
	}
	if GetRetention("missing").AuditLogDays != DefaultRetention().AuditLogDays {
		t.Fatal("missing tenant should return default")
	}
}
