package circuitbreaker

import (
	"testing"
	"time"
)

// TestCircuitBreakerTripsAndRecoversViaCache 验证状态写入 cache 后能开闸、半开探测并恢复。
func TestCircuitBreakerTripsAndRecoversViaCache(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cfg.MinimumRequests = 4
	cfg.ErrorRatePercent = 50
	cfg.HalfOpenMaxRequests = 2
	cfg.OpenTimeoutSeconds = 1
	cfg.StorageType = "memory"

	cb, err := NewCircuitBreaker(cfg)
	if err != nil {
		t.Fatalf("创建熔断器失败: %v", err)
	}

	key := NodeCircuitKey("svc-1", "n1")
	for i := 0; i < 4; i++ {
		if !cb.Check(key) {
			t.Fatalf("第%d次请求在熔断前被拒绝", i+1)
		}
		cb.RecordFailure(key, 10, errTestUpstream)
	}

	if cb.Check(key) {
		t.Fatal("错误率达到阈值后应拒绝该节点")
	}
	if cb.GetState(key) != StateOpen {
		t.Fatalf("期望 open，实际 %s", cb.GetState(key))
	}

	time.Sleep(1100 * time.Millisecond)
	if !cb.Check(key) {
		t.Fatal("打开超时后应进入半开并放行探测")
	}
	cb.RecordSuccess(key, 8)
	cb.RecordSuccess(key, 8)
	if cb.GetState(key) != StateClosed {
		t.Fatalf("半开探测成功后应关闭，实际 %s", cb.GetState(key))
	}
}

// TestCacheStorageRoundTrip 验证 cache 存储的写入、读取与重置。
func TestCacheStorageRoundTrip(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	store, err := newCacheCircuitBreakerStorage(cfg)
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}
	info := newClosedCircuit()
	info.TotalRequests = 3
	info.FailureRequests = 1
	if err := store.SetInfo("k1", info); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	got, err := store.GetInfo("k1")
	if err != nil || got == nil {
		t.Fatalf("读取失败: %v", err)
	}
	if got.TotalRequests != 3 || got.FailureRequests != 1 {
		t.Fatalf("状态不一致: %+v", got)
	}
	if err := store.Reset("k1"); err != nil {
		t.Fatalf("重置失败: %v", err)
	}
	got, err = store.GetInfo("k1")
	if err != nil {
		t.Fatalf("重置后读取失败: %v", err)
	}
	if got != nil {
		t.Fatalf("重置后应为空: %+v", got)
	}
}

var errTestUpstream = errString("upstream failed")

type errString string

func (e errString) Error() string { return string(e) }
