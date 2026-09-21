package controllers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gateway/web/views/hub0009/models"

	"go.mongodb.org/mongo-driver/bson"
)

func TestOverallHealthStatus(t *testing.T) {
	t.Parallel()
	if got := overallHealthStatus(nil); got != healthStatusOK {
		t.Fatalf("empty items status = %s, want %s", got, healthStatusOK)
	}
	if got := overallHealthStatus([]models.HealthItem{
		{Status: healthStatusOK},
		{Status: healthStatusSkipped},
	}); got != healthStatusOK {
		t.Fatalf("skipped should not degrade, got %s", got)
	}
	if got := overallHealthStatus([]models.HealthItem{
		{Status: healthStatusOK},
		{Status: healthStatusError},
	}); got != healthStatusDegraded {
		t.Fatalf("error should degrade, got %s", got)
	}
}

func TestSanitizeHealthError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		err  error
		want string
	}{
		{context.DeadlineExceeded, "探测超时"},
		{errors.New("i/o timeout"), "探测超时"},
		{errors.New("Error 1045: Access denied for user 'root'@'10.0.0.8' (using password: YES)"), "认证失败"},
		{errors.New("NOAUTH Authentication required"), "认证失败"},
		{errors.New("dial tcp 10.1.2.3:27017: connect: connection refused"), "连接失败"},
		{errors.New("mongodb://admin:secret@10.1.2.3:27017"), "连接失败"},
	}
	for _, tc := range cases {
		if got := sanitizeHealthError(tc.err); got != tc.want {
			t.Fatalf("sanitizeHealthError(%v) = %q, want %q", tc.err, got, tc.want)
		}
		if strings.Contains(sanitizeHealthError(tc.err), "secret") ||
			strings.Contains(sanitizeHealthError(tc.err), "10.1.2.3") ||
			strings.Contains(sanitizeHealthError(tc.err), "root") {
			t.Fatalf("sanitized message leaked details: %q", sanitizeHealthError(tc.err))
		}
	}
}

func TestFormatIndexKeys(t *testing.T) {
	t.Parallel()
	got := formatIndexKeys(bson.D{{Key: "tenantId", Value: 1}, {Key: "traceId", Value: 1}})
	if got != "tenantId, traceId" {
		t.Fatalf("formatIndexKeys = %q", got)
	}
	got = formatIndexKeys(bson.D{{Key: "gatewayInstanceId", Value: 1}, {Key: "gatewayStartProcessingTime", Value: -1}})
	if got != "gatewayInstanceId, gatewayStartProcessingTime desc" {
		t.Fatalf("formatIndexKeys desc = %q", got)
	}
}

func TestPlannedMongoObjectsIncludeDefaultIndex(t *testing.T) {
	t.Parallel()
	objects := plannedMongoObjects()
	if len(objects) == 0 {
		t.Fatal("expected system collections")
	}
	for _, obj := range objects {
		if len(obj.Indexes) == 0 || obj.Indexes[0].Name != defaultMongoIndexName {
			t.Fatalf("%s missing leading %s", obj.Name, defaultMongoIndexName)
		}
	}
}

func TestStoreMetaSkipMessage(t *testing.T) {
	t.Parallel()
	if got := (storeMeta{}).skipMessage(); got != "配置未启用" {
		t.Fatalf("root disabled = %q", got)
	}
	if got := (storeMeta{rootEnabled: true}).skipMessage(); got != "连接未启用" {
		t.Fatalf("conn disabled = %q", got)
	}
	if got := (storeMeta{rootEnabled: true, connEnabled: true}).skipMessage(); got != "进程未加载连接" {
		t.Fatalf("loaded skip = %q", got)
	}
}

func TestSortHealthItems(t *testing.T) {
	t.Parallel()
	items := []models.HealthItem{
		{Kind: healthKindCache, Name: "redis_main"},
		{Kind: healthKindMongo, Name: "b"},
		{Kind: healthKindMongo, Name: "a"},
		{Kind: healthKindProcess, Name: "controlPlane"},
	}
	sortHealthItems(items)
	want := []string{"controlPlane", "a", "b", "redis_main"}
	for i, name := range want {
		if items[i].Name != name {
			t.Fatalf("items[%d].Name = %s, want %s", i, items[i].Name, name)
		}
	}
}
