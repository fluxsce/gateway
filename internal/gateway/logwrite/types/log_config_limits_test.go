package types

import "testing"

func TestQueueSizeBatchLimitFlushInterval(t *testing.T) {
	if got := QueueSize(nil); got != DefaultAsyncQueueSize {
		t.Fatalf("QueueSize(nil) = %d, want %d", got, DefaultAsyncQueueSize)
	}
	if got := QueueSize(&LogConfig{AsyncQueueSize: 50}); got != DefaultAsyncQueueSize {
		t.Fatalf("QueueSize(<100) = %d, want default", got)
	}
	if got := QueueSize(&LogConfig{AsyncQueueSize: 200}); got != 200 {
		t.Fatalf("QueueSize(200) = %d", got)
	}

	if got := BatchLimit(nil); got != DefaultBatchSize {
		t.Fatalf("BatchLimit(nil) = %d, want %d", got, DefaultBatchSize)
	}
	if got := BatchLimit(&LogConfig{BatchSize: 50}); got != 50 {
		t.Fatalf("BatchLimit(50) = %d", got)
	}

	if got := FlushIntervalMs(nil); got != DefaultAsyncFlushIntervalMs {
		t.Fatalf("FlushIntervalMs(nil) = %d, want %d", got, DefaultAsyncFlushIntervalMs)
	}
	if got := FlushIntervalMs(&LogConfig{AsyncFlushIntervalMs: 0}); got != DefaultAsyncFlushIntervalMs {
		t.Fatalf("FlushIntervalMs(0) 应用默认")
	}
	if got := FlushIntervalMs(&LogConfig{AsyncFlushIntervalMs: 3000}); got != 3000 {
		t.Fatalf("FlushIntervalMs(3000) = %d", got)
	}
}
