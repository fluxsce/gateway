package circuitbreaker

import (
	"testing"
	"time"
)

type fakeClock struct {
	t time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.t
}

func newTestBreaker(t *testing.T, window, minRequests, errorRate int) (*circuitBreakerImpl, *fakeClock) {
	t.Helper()
	cfg := DefaultCircuitBreakerConfig()
	cfg.WindowSizeSeconds = int64(window)
	cfg.MinimumRequests = minRequests
	cfg.ErrorRatePercent = errorRate
	cfg.OpenTimeoutSeconds = 5
	cfg.StorageType = "memory"
	handler, err := NewCircuitBreaker(cfg)
	if err != nil {
		t.Fatalf("创建熔断器失败: %v", err)
	}
	impl := handler.(*circuitBreakerImpl)
	clock := &fakeClock{t: time.Unix(1_700_000_000, 0)}
	impl.now = clock.Now
	return impl, clock
}

func TestSlidingWindowExpiresOldFailures(t *testing.T) {
	cb, clock := newTestBreaker(t, 2, 2, 50)
	key := NodeCircuitKey("svc-window", "n-expire")

	cb.RecordFailure(key, 10, errTestUpstream)
	clock.t = clock.t.Add(3 * time.Second)
	cb.RecordFailure(key, 10, errTestUpstream)
	if cb.GetState(key) != StateClosed {
		t.Fatal("窗口外的旧失败不应再参与开闸")
	}
	if !cb.Check(key) {
		t.Fatal("仅窗口内 1 次失败不应摘除")
	}

	cb.RecordFailure(key, 10, errTestUpstream)
	if cb.GetState(key) != StateOpen {
		t.Fatalf("窗口内连续失败应开闸，实际 %s", cb.GetState(key))
	}
}

func TestSlidingWindowSuccessDilutesFailureRate(t *testing.T) {
	cb, _ := newTestBreaker(t, 60, 4, 50)
	key := NodeCircuitKey("svc-window", "n-dilute")

	cb.RecordFailure(key, 10, errTestUpstream)
	cb.RecordSuccess(key, 8)
	cb.RecordSuccess(key, 8)
	cb.RecordSuccess(key, 8)
	if cb.GetState(key) != StateClosed {
		t.Fatal("窗口内失败率 25% 不应开闸")
	}
}

func TestHalfOpenSuccessClearsWindow(t *testing.T) {
	cb, clock := newTestBreaker(t, 60, 2, 50)
	cb.config.HalfOpenMaxRequests = 1
	cb.config.OpenTimeoutSeconds = 2
	key := NodeCircuitKey("svc-window", "n-reset")

	cb.RecordFailure(key, 10, errTestUpstream)
	cb.RecordFailure(key, 10, errTestUpstream)
	if cb.GetState(key) != StateOpen {
		t.Fatal("应已开闸")
	}

	clock.t = clock.t.Add(3 * time.Second)
	if !cb.Check(key) {
		t.Fatal("开闸超时后应半开放行")
	}
	cb.RecordSuccess(key, 8)
	if cb.GetState(key) != StateClosed {
		t.Fatalf("半开成功应关闸，实际 %s", cb.GetState(key))
	}
	info := cb.loadCircuitForTest(key)
	if info.TotalRequests != 0 || len(info.WindowBuckets) != 0 {
		t.Fatalf("关闸应清空窗口: %+v", info)
	}
}

func TestHalfOpenFailureReopensImmediately(t *testing.T) {
	cb, clock := newTestBreaker(t, 60, 2, 50)
	cb.config.HalfOpenMaxRequests = 3
	cb.config.OpenTimeoutSeconds = 2
	key := NodeCircuitKey("svc-window", "n-reopen")

	cb.RecordFailure(key, 10, errTestUpstream)
	cb.RecordFailure(key, 10, errTestUpstream)
	clock.t = clock.t.Add(3 * time.Second)
	if !cb.Check(key) {
		t.Fatal("应进入半开")
	}
	cb.RecordSuccess(key, 8)
	cb.RecordFailure(key, 10, errTestUpstream)
	if cb.GetState(key) != StateOpen {
		t.Fatalf("半开任一失败应立即开闸，实际 %s", cb.GetState(key))
	}
	if cb.Check(key) {
		t.Fatal("重新开闸后在超时前应拒绝")
	}
}

func TestSlowCallRateTripsIndependently(t *testing.T) {
	cb, _ := newTestBreaker(t, 60, 2, 100)
	cb.config.SlowCallThreshold = 100
	cb.config.SlowCallRatePercent = 50
	key := NodeCircuitKey("svc-window", "n-slow")

	cb.RecordSuccess(key, 200)
	cb.RecordSuccess(key, 200)
	if cb.GetState(key) != StateOpen {
		t.Fatal("慢调用率达标即使没有传输失败也应开闸")
	}
}

func TestBelowMinimumRequestsNeverTrips(t *testing.T) {
	cb, _ := newTestBreaker(t, 60, 5, 1)
	key := NodeCircuitKey("svc-window", "n-cold")
	for i := 0; i < 4; i++ {
		cb.RecordFailure(key, 10, errTestUpstream)
	}
	if cb.GetState(key) != StateClosed {
		t.Fatal("未达最小请求数即使全失败也不开闸")
	}
}

func TestDisabledAndEmptyKeyAlwaysAllow(t *testing.T) {
	cb, _ := newTestBreaker(t, 60, 1, 1)
	cb.config.Enabled = false
	key := NodeCircuitKey("svc-window", "n-off")
	cb.RecordFailure(key, 10, errTestUpstream)
	if !cb.Check(key) {
		t.Fatal("关闭熔断后应放行")
	}
	if !cb.Check("") {
		t.Fatal("空键应放行")
	}
	RecordKey(cb, "", false, time.Millisecond, errTestUpstream)
}

func TestFiveXXCountsAsFailureAndTrips(t *testing.T) {
	cb, _ := newTestBreaker(t, 60, 2, 50)
	key := NodeCircuitKey("svc-window", "n-5xx")
	RecordKey(cb, key, !AttemptFailed(nil, 404), time.Millisecond, AttemptError(nil, 404))
	RecordKey(cb, key, !AttemptFailed(nil, 200), time.Millisecond, AttemptError(nil, 200))
	if cb.GetState(key) != StateClosed {
		t.Fatal("4xx 与 2xx 不应开闸")
	}
	RecordKey(cb, key, !AttemptFailed(nil, 500), time.Millisecond, AttemptError(nil, 500))
	RecordKey(cb, key, !AttemptFailed(nil, 502), time.Millisecond, AttemptError(nil, 502))
	if cb.GetState(key) != StateOpen {
		t.Fatal("两次 5xx 应开闸，4xx 不计入失败")
	}
}

func TestHalfOpenReservesProbeSlots(t *testing.T) {
	cb, clock := newTestBreaker(t, 60, 2, 50)
	cb.config.HalfOpenMaxRequests = 2
	cb.config.OpenTimeoutSeconds = 2
	key := NodeCircuitKey("svc-window", "n-reserve")

	cb.RecordFailure(key, 10, errTestUpstream)
	cb.RecordFailure(key, 10, errTestUpstream)
	if cb.GetState(key) != StateOpen {
		t.Fatal("应已开闸")
	}

	clock.t = clock.t.Add(3 * time.Second)
	if !cb.Check(key) {
		t.Fatal("第一个探测应占坑放行")
	}
	if !cb.Check(key) {
		t.Fatal("第二个探测应占坑放行")
	}
	if cb.Check(key) {
		t.Fatal("超过半开配额的探测应拒绝")
	}
}

func (cb *circuitBreakerImpl) loadCircuitForTest(key string) *CircuitBreakerInfo {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.loadCircuit(key)
}
