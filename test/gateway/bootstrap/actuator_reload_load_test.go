package bootstrap

import (
	"context"
	"fmt"
	"io"
	"net"
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
	defaultActuatorDuration = 5 * time.Minute
	defaultActuatorTimeout  = 30 * time.Second
	defaultActuatorDialTO   = 15 * time.Second
	defaultShutdownGrace    = 60 * time.Second
	maxSampleNetworkErrors  = 8
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
	netErrs   []string
	mu        sync.Mutex
}

func (s *actuatorLoadStats) recordLatency(d time.Duration) {
	s.mu.Lock()
	s.latencies = append(s.latencies, d)
	s.mu.Unlock()
}

func (s *actuatorLoadStats) recordNetworkError(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	if len(s.netErrs) < maxSampleNetworkErrors {
		s.netErrs = append(s.netErrs, err.Error())
	}
	s.mu.Unlock()
}

func (s *actuatorLoadStats) sampleNetworkErrors() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.netErrs))
	copy(out, s.netErrs)
	return out
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

// TestActuatorHighConcurrencyDuringGatewayReload 用 30 个并发线程压测 actuator（GET），
// 仅输出统计汇总，不因失败率或网络错误 Fail/抛异常。
//
// 运行示例（-timeout 必须大于压测时长，否则进程被强杀会出现 IO wait 堆栈）：
//
//	set ACTUATOR_LOAD_TEST=1
//	go test ./test/gateway/bootstrap -run TestActuatorHighConcurrencyDuringGatewayReload -v -count=1 -timeout 15m
//
// 可选环境变量：ACTUATOR_URL、ACTUATOR_WORKERS、ACTUATOR_DURATION（如 10m）
func TestActuatorHighConcurrencyDuringGatewayReload(t *testing.T) {
	// if os.Getenv("ACTUATOR_LOAD_TEST") != "1" {
	// 	t.Skip("跳过外部环境压测；设置 ACTUATOR_LOAD_TEST=1 后运行")
	// }
	if testing.Short() {
		t.Skip("短测试模式下跳过长时间压测")
	}

	targetURL := envOrDefault("ACTUATOR_URL", defaultActuatorURL)
	workers := envIntOrDefault("ACTUATOR_WORKERS", defaultActuatorWorkers)
	duration := envDurationOrDefault("ACTUATOR_DURATION", defaultActuatorDuration)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultActuatorDialTO,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        workers * 2,
		MaxIdleConnsPerHost: workers,
		MaxConnsPerHost:     workers,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   false,
	}
	client := &http.Client{
		Timeout:   defaultActuatorTimeout,
		Transport: transport,
	}
	defer transport.CloseIdleConnections()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// 压测窗口结束立即关空闲连接，缩短 worker 退出时间，避免收尾堆栈刷屏。
	go func() {
		<-ctx.Done()
		transport.CloseIdleConnections()
	}()

	stats := &actuatorLoadStats{
		latencies: make([]time.Duration, 0, workers*128),
	}

	t.Logf("开始压测: method=GET url=%s workers=%d duration=%s requestTimeout=%s",
		targetURL, workers, duration, defaultActuatorTimeout)
	t.Logf("请使用: go test ... -timeout 15m （需明显大于 duration=%s）", duration)

	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				doActuatorRequest(ctx, client, targetURL, stats)
			}
		}()
	}

	waitWorkers(&wg, defaultShutdownGrace)
	transport.CloseIdleConnections()
	elapsed := time.Since(start)
	printActuatorSummary(t, stats, workers, elapsed)

	// 只汇总，不 Fail，避免把失败率/收尾等待当成测试异常。
}

// waitWorkers 等待 worker 退出；超时后继续出汇总，避免永久卡死。
func waitWorkers(wg *sync.WaitGroup, grace time.Duration) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(grace):
	}
}

func printActuatorSummary(t *testing.T, stats *actuatorLoadStats, workers int, elapsed time.Duration) {
	t.Helper()
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
	for i, msg := range stats.sampleNetworkErrors() {
		t.Logf("网络错误样例[%d]: %s", i+1, msg)
	}
	if total == 0 {
		t.Log("提示: 未发出任何请求，请检查目标地址与网络（测试仍记为通过）")
	}
	t.Logf("==========================")
}

// doActuatorRequest 发起单次 GET，并按状态码/网络错误计入统计。
func doActuatorRequest(ctx context.Context, client *http.Client, targetURL string, stats *actuatorLoadStats) {
	if ctx.Err() != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		stats.total.Add(1)
		stats.failed.Add(1)
		stats.network.Add(1)
		stats.recordNetworkError(err)
		return
	}

	begin := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(begin)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		stats.total.Add(1)
		stats.recordLatency(elapsed)
		stats.failed.Add(1)
		stats.network.Add(1)
		stats.recordNetworkError(err)
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	stats.total.Add(1)
	stats.recordLatency(elapsed)
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
