package retention

import (
	"context"
	"strings"
	"testing"
	"time"

	"gateway/pkg/syssetting"
)

func TestFuncJobRun(t *testing.T) {
	called := false
	job := NewFuncJob("demo", func(context.Context) error {
		called = true
		return nil
	})
	if job.Name() != "demo" {
		t.Fatalf("name=%s", job.Name())
	}
	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("fn not called")
	}
}

func TestTableJobSkipNil(t *testing.T) {
	job := NewTableJob(nil, Dataset{Name: "audit", Table: "HUB_AUTH_AUDIT_LOG", TimeCol: "addTime"})
	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestDaysCutoff(t *testing.T) {
	before, ok := DaysCutoff(func(string) int { return 7 })("default")
	if !ok || before.IsZero() {
		t.Fatal("expected cutoff")
	}
	if _, ok := DaysCutoff(func(string) int { return 0 })("default"); ok {
		t.Fatal("expected skip")
	}
}

func TestLimitedDeleteUsesPublicBuilder(t *testing.T) {
	sql, args, err := limitedDelete("mysql", "HUB_ALERT_LOG", "alertTimestamp", "default", "ts")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "DELETE FROM HUB_ALERT_LOG") || !strings.Contains(sql, "LIMIT ?") {
		t.Fatalf("mysql sql=%q", sql)
	}
	if strings.Contains(sql, " IN (") {
		t.Fatalf("mysql should not use IN list, sql=%q", sql)
	}
	if len(args) != 3 || args[0] != "default" || args[2] != deleteBatchSize {
		t.Fatalf("args=%v", args)
	}
	oracle, _, err := limitedDelete("oracle", "HUB_ALERT_LOG", "alertTimestamp", "default", "ts")
	if err != nil || !strings.Contains(oracle, "ROWNUM") || strings.Contains(oracle, " IN (") {
		t.Fatalf("oracle sql=%q err=%v", oracle, err)
	}
	if _, _, err := limitedDelete("mysql", "HUB_ALERT_LOG;DROP", "alertTimestamp", "default", "ts"); err == nil {
		t.Fatal("expected invalid table")
	}
}

func TestClusterAckFollowsEventDays(t *testing.T) {
	syssetting.PutRetention("ack-test", syssetting.RetentionSettings{
		AuditLogDays: 180, TaskLogDays: 30, AlertLogDays: 7,
		ClusterEventDays: 3, MetricsDays: 30, GatewayLogDefaultDays: 30,
	})
	var event, ack Dataset
	for _, ds := range Datasets() {
		switch ds.Name {
		case "cluster-event":
			event = ds
		case "cluster-ack":
			ack = ds
		}
	}
	if event.Cutoff == nil || ack.Cutoff == nil {
		t.Skip("cluster cleanup disabled")
	}
	eBefore, eOK := event.Cutoff("ack-test")
	aBefore, aOK := ack.Cutoff("ack-test")
	if !eOK || !aOK {
		t.Fatal("expected both cutoffs")
	}
	if eBefore.Sub(aBefore).Abs() > time.Second {
		t.Fatalf("ack cutoff %v vs event %v", aBefore, eBefore)
	}
}

func TestJobDue(t *testing.T) {
	now := time.Date(2026, 8, 21, 10, 0, 0, 0, time.Local)
	if !jobDue(now, time.Time{}, time.Hour, "") {
		t.Fatal("first run without start time should be due")
	}
	if jobDue(now, now.Add(-30*time.Minute), time.Hour, "") {
		t.Fatal("interval not elapsed")
	}
	if !jobDue(now, now.Add(-2*time.Hour), time.Hour, "") {
		t.Fatal("interval elapsed should be due")
	}
	if !jobDue(now, now.Add(-2*time.Minute), 2*time.Minute, "") {
		t.Fatal("minute interval elapsed should be due")
	}
	if jobDue(now, now.Add(-time.Minute), 2*time.Minute, "") {
		t.Fatal("minute interval not elapsed")
	}
	start := time.Date(2026, 8, 21, 15, 0, 0, 0, time.Local)
	if jobDue(start.Add(-time.Hour), time.Time{}, time.Hour, "15:00") {
		t.Fatal("should wait until start time")
	}
	if !jobDue(start.Add(time.Minute), time.Time{}, time.Hour, "15:00") {
		t.Fatal("after start time first run should be due")
	}
	if jobDue(start.Add(time.Hour), start.Add(time.Minute), time.Hour, "15:00") {
		t.Fatal("same day after start should run only once")
	}
	nextDay := start.Add(25 * time.Hour)
	if !jobDue(nextDay, start.Add(time.Minute), time.Hour, "15:00") {
		t.Fatal("next day after start time should be due")
	}
}
