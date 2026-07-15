package bootstrap

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	defaultActuatorURL      = "http://datahub.flux.com.cn:28180/datahub01webApp/actuator"
	defaultActuatorWorkers  = 30
	defaultActuatorDuration = 120 * time.Second
	defaultActuatorTimeout  = 10 * time.Second
)

// actuatorLoadStats 汇总高并发访问结果，用于评估网关重载期间的失败率与延迟。
type actuatorLoadStats struct {
	total     atomic.Int64
	success   atomic.Int64
	failed    atomic.Int64
	status5xx atomic.Int64
	status4xx atomic.Int64
	network   atomic.Int64
	latencies []time.Duration
	mu        sync.Mutex
}

func (s *actuatorLoadStats) recordLatency(d time.Duration) {
	s.mu.Lock()
	s.latencies = append(s.latencies, d)
	s.mu.Unlock()
}

func (s *actuatorLoadStats) percentile(p float64) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.latencies)
	if n == 0 {
		return 0
	}
	cp := append([]time.Duration(nil), s.latencies...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	idx := int(float64(n-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return cp[idx]
}

func (s *actuatorLoadStats) average() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.latencies) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range s.latencies {
		sum += d
	}
	return sum / time.Duration(len(s.latencies))
}

// TestActuatorHighConcurrencyDuringGatewayReload 用 30 个并发线程压测 actuator，
// 统计成功/失败率与延迟分位。目标环境存在网关热重载时，可用于观察重载对可用性的影响。
//
// 默认跳过；开启方式：
//
//	set ACTUATOR_LOAD_TEST=1
//	go test ./test/gateway/bootstrap -run TestActuatorHighConcurrencyDuringGatewayReload -v -count=1 -timeout 5m
//
// 可选环境变量：
//   - ACTUATOR_URL：目标地址，默认 datahub actuator
//   - ACTUATOR_WORKERS：并发数，默认 30
//   - ACTUATOR_DURATION：压测时长，默认 60s（如 90s）
func TestActuatorHighConcurrencyDuringGatewayReload(t *testing.T) {
	if os.Getenv("ACTUATOR_LOAD_TEST") != "1" {
		t.Skip("跳过外部环境压测；设置 ACTUATOR_LOAD_TEST=1 后运行")
	}
	if testing.Short() {
		t.Skip("短测试模式下跳过长时间压测")
	}

	targetURL := envOrDefault("ACTUATOR_URL", defaultActuatorURL)
	workers := envIntOrDefault("ACTUATOR_WORKERS", defaultActuatorWorkers)
	duration := envDurationOrDefault("ACTUATOR_DURATION", defaultActuatorDuration)

	client := &http.Client{
		Timeout: defaultActuatorTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        workers * 2,
			MaxIdleConnsPerHost: workers * 2,
			MaxConnsPerHost:     workers * 2,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	stats := &actuatorLoadStats{
		latencies: make([]time.Duration, 0, workers*128),
	}

	t.Logf("开始压测: url=%s workers=%d duration=%s", targetURL, workers, duration)
	t.Log("请在压测期间对网关执行重载，以便观察失败率变化")

	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				doActuatorRequest(ctx, client, targetURL, stats)
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)

	total := stats.total.Load()
	success := stats.success.Load()
	failed := stats.failed.Load()
	failureRate := 0.0
	if total > 0 {
		failureRate = float64(failed) * 100 / float64(total)
	}
	qps := 0.0
	if elapsed.Seconds() > 0 {
		qps = float64(total) / elapsed.Seconds()
	}

	t.Logf("======== 压测汇总 ========")
	t.Logf("耗时:           %s", elapsed.Round(time.Millisecond))
	t.Logf("并发线程:       %d", workers)
	t.Logf("总请求:         %d", total)
	t.Logf("成功:           %d", success)
	t.Logf("失败:           %d", failed)
	t.Logf("  网络错误:     %d", stats.network.Load())
	t.Logf("  HTTP 4xx:     %d", stats.status4xx.Load())
	t.Logf("  HTTP 5xx:     %d", stats.status5xx.Load())
	t.Logf("失败率:         %.4f%%", failureRate)
	t.Logf("QPS:            %.2f", qps)
	t.Logf("延迟 avg:       %s", stats.average().Round(time.Microsecond))
	t.Logf("延迟 p50:       %s", stats.percentile(0.50).Round(time.Microsecond))
	t.Logf("延迟 p95:       %s", stats.percentile(0.95).Round(time.Microsecond))
	t.Logf("延迟 p99:       %s", stats.percentile(0.99).Round(time.Microsecond))
	t.Logf("==========================")

	// 不因失败率自动 Fail：重载期间允许短时失败，由人工根据汇总判断效率。
	if total == 0 {
		t.Fatal("未发出任何请求，请检查目标地址与网络")
	}
}

// doActuatorRequest 发起单次 GET，并按状态码/网络错误计入统计。
func doActuatorRequest(ctx context.Context, client *http.Client, targetURL string, stats *actuatorLoadStats) {
	stats.total.Add(1)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		stats.failed.Add(1)
		stats.network.Add(1)
		return
	}

	begin := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(begin)
	stats.recordLatency(elapsed)
	if err != nil {
		// 压测结束取消 context 不算失败
		if ctx.Err() != nil {
			stats.total.Add(-1)
			return
		}
		stats.failed.Add(1)
		stats.network.Add(1)
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 400:
		stats.success.Add(1)
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		stats.failed.Add(1)
		stats.status4xx.Add(1)
	default:
		stats.failed.Add(1)
		stats.status5xx.Add(1)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n <= 0 {
		return fallback
	}
	return n
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}
