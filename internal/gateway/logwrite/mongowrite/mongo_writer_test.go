package mongowrite

import (
	"context"
	"testing"

	"gateway/internal/gateway/logwrite/types"
)

func TestUpdateAccessLogRequiresIDs(t *testing.T) {
	w := &MongoWriter{config: &types.LogConfig{}}
	if _, err := w.UpdateAccessLog(context.Background(), &types.AccessLog{}); err == nil {
		t.Fatal("缺少 tenantId/traceId 应失败")
	}
}

func TestMongoWriterRejectsAfterClose(t *testing.T) {
	w := &MongoWriter{config: &types.LogConfig{}}
	w.closed.Store(true)
	if err := w.Write(context.Background(), &types.AccessLog{}); err == nil {
		t.Fatal("关闭后 Write 应失败")
	}
	if _, err := w.UpdateAccessLog(context.Background(), &types.AccessLog{TenantID: "t", TraceID: "x"}); err == nil {
		t.Fatal("关闭后 UpdateAccessLog 应失败")
	}
}

func TestBatchLimitAndQueueSize(t *testing.T) {
	if got := types.BatchLimit(nil); got != types.DefaultBatchSize {
		t.Fatalf("BatchLimit(nil) = %d, want %d", got, types.DefaultBatchSize)
	}
	if got := types.QueueSize(nil); got != types.DefaultAsyncQueueSize {
		t.Fatalf("QueueSize(nil) = %d, want %d", got, types.DefaultAsyncQueueSize)
	}
	cfg := &types.LogConfig{BatchSize: 50, AsyncQueueSize: 200}
	if got := types.BatchLimit(cfg); got != 50 {
		t.Fatalf("BatchLimit = %d, want 50", got)
	}
	if got := types.QueueSize(cfg); got != 200 {
		t.Fatalf("QueueSize = %d, want 200", got)
	}
}
