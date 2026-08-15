package circuitbreaker

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrentHalfOpenCheckRespectsQuota(t *testing.T) {
	cb, clock := newTestBreaker(t, 60, 2, 50)
	cb.config.HalfOpenMaxRequests = 3
	cb.config.OpenTimeoutSeconds = 2
	key := NodeCircuitKey("svc-conc", "n-halfopen")

	cb.RecordFailure(key, 10, errTestUpstream)
	cb.RecordFailure(key, 10, errTestUpstream)
	if cb.GetState(key) != StateOpen {
		t.Fatal("应已开闸")
	}
	clock.t = clock.t.Add(3 * time.Second)

	const goroutines = 32
	var allowed int64
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if cb.Check(key) {
				atomic.AddInt64(&allowed, 1)
			}
		}()
	}
	wg.Wait()

	if allowed != 3 {
		t.Fatalf("半开并发探测应正好占 3 个名额，实际 %d", allowed)
	}
	if cb.GetState(key) != StateHalfOpen {
		t.Fatalf("名额用尽后应保持半开，实际 %s", cb.GetState(key))
	}
}

func TestConcurrentRecordFailureTripsOnce(t *testing.T) {
	cb, _ := newTestBreaker(t, 60, 10, 50)
	key := NodeCircuitKey("svc-conc", "n-trip")

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			cb.RecordFailure(key, 10, errTestUpstream)
		}()
	}
	wg.Wait()

	if cb.GetState(key) != StateOpen {
		t.Fatalf("并发失败应开闸，实际 %s", cb.GetState(key))
	}
	if cb.Check(key) {
		t.Fatal("开闸后并发 Check 应拒绝")
	}
	info := cb.loadCircuitForTest(key)
	if info.TotalRequests != n || info.FailureRequests != n {
		t.Fatalf("窗口计数应等于写入次数: %+v", info)
	}
}

func TestConcurrentMixedRecordsStayConsistent(t *testing.T) {
	cb, _ := newTestBreaker(t, 60, 100, 50)
	key := NodeCircuitKey("svc-conc", "n-mix")

	const perKind = 40
	var wg sync.WaitGroup
	wg.Add(perKind * 2)
	for i := 0; i < perKind; i++ {
		go func() {
			defer wg.Done()
			cb.RecordSuccess(key, 8)
		}()
		go func() {
			defer wg.Done()
			cb.RecordFailure(key, 10, errTestUpstream)
		}()
	}
	wg.Wait()

	info := cb.loadCircuitForTest(key)
	if info.TotalRequests != int64(perKind*2) {
		t.Fatalf("总请求应=%d，实际 %d", perKind*2, info.TotalRequests)
	}
	if info.FailureRequests+info.SuccessRequests != info.TotalRequests {
		t.Fatalf("成功+失败应等于总数: %+v", info)
	}
	if cb.GetState(key) != StateClosed {
		t.Fatal("未达最小请求数不应开闸")
	}
}

func TestIsolatedNodeKeysDoNotShareState(t *testing.T) {
	cb, _ := newTestBreaker(t, 60, 2, 50)
	openKey := NodeCircuitKey("svc-conc", "n-a")
	closedKey := NodeCircuitKey("svc-conc", "n-b")

	cb.RecordFailure(openKey, 10, errTestUpstream)
	cb.RecordFailure(openKey, 10, errTestUpstream)
	if cb.GetState(openKey) != StateOpen {
		t.Fatal("n-a 应开闸")
	}
	if cb.GetState(closedKey) != StateClosed || !cb.Check(closedKey) {
		t.Fatal("n-b 不应被 n-a 带开")
	}
}

func TestSharedCacheSurvivesNewHandler(t *testing.T) {
	cb1, _ := newTestBreaker(t, 60, 2, 50)
	key := NodeCircuitKey("svc-conc", "n-reload")
	cb1.RecordFailure(key, 10, errTestUpstream)
	cb1.RecordFailure(key, 10, errTestUpstream)

	cb2, err := NewCircuitBreaker(cb1.config)
	if err != nil {
		t.Fatalf("重建熔断器失败: %v", err)
	}
	cb2Impl := cb2.(*circuitBreakerImpl)
	cb2Impl.now = cb1.now
	if cb2.GetState(key) != StateOpen {
		t.Fatal("热重载后新 handler 应读到同一开闸状态")
	}
	if cb2.Check(key) {
		t.Fatal("共享 cache 上开闸节点仍应被跳过")
	}
}
