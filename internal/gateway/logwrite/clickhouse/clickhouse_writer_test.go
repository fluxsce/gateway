package clickhouse

import (
	"testing"

	"gateway/internal/gateway/logwrite/types"
)

func TestEffectiveBatchSizeDoesNotMutateConfig(t *testing.T) {
	cfg := &types.LogConfig{BatchSize: 100}
	if got := effectiveBatchSize(cfg); got != minClickHouseBatchSize {
		t.Fatalf("effectiveBatchSize(100) = %d, want %d", got, minClickHouseBatchSize)
	}
	if cfg.BatchSize != 100 {
		t.Fatalf("不应改写共享配置, BatchSize=%d", cfg.BatchSize)
	}

	cfg.BatchSize = 8000
	if got := effectiveBatchSize(cfg); got != 8000 {
		t.Fatalf("effectiveBatchSize(8000) = %d, want 8000", got)
	}
}
