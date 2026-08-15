package circuitbreaker

import "time"

// maxWindowBuckets 限制持久化分桶数量，窗口超过 60 秒时加长单桶。
const maxWindowBuckets = 60

// windowSeconds 返回滑动窗口长度，未配置时按 60 秒。
func (cb *circuitBreakerImpl) windowSeconds() int64 {
	if cb.config != nil && cb.config.WindowSizeSeconds > 0 {
		return cb.config.WindowSizeSeconds
	}
	return 60
}

// bucketDurationSeconds 计算单桶时长，保证窗口内桶数不超过 maxWindowBuckets。
func (cb *circuitBreakerImpl) bucketDurationSeconds() int64 {
	window := cb.windowSeconds()
	d := window / maxWindowBuckets
	if d < 1 {
		d = 1
	}
	return d
}

// currentUnix 返回状态机使用的当前时间，测试可替换 now。
func (cb *circuitBreakerImpl) currentUnix() int64 {
	if cb.now != nil {
		return cb.now().Unix()
	}
	return time.Now().Unix()
}

// recordSample 把一次尝试写入当前时间桶，并回写窗口汇总。
func (cb *circuitBreakerImpl) recordSample(circuit *CircuitBreakerInfo, now int64, failed, slow bool) {
	cb.rollWindow(circuit, now)
	bucket := cb.ensureBucket(circuit, now)
	bucket.Total++
	if failed {
		bucket.Failures++
	}
	if slow {
		bucket.Slow++
	}
	syncWindowTotals(circuit)
}

// rollWindow 丢掉已经滑出窗口的桶。
func (cb *circuitBreakerImpl) rollWindow(circuit *CircuitBreakerInfo, now int64) {
	cutoff := now - cb.windowSeconds()
	kept := circuit.WindowBuckets[:0]
	for _, bucket := range circuit.WindowBuckets {
		if bucket.Start >= cutoff {
			kept = append(kept, bucket)
		}
	}
	circuit.WindowBuckets = kept
}

// ensureBucket 返回 now 所在时间片；不存在则追加。
func (cb *circuitBreakerImpl) ensureBucket(circuit *CircuitBreakerInfo, now int64) *CircuitWindowBucket {
	dur := cb.bucketDurationSeconds()
	start := now - (now % dur)
	for i := range circuit.WindowBuckets {
		if circuit.WindowBuckets[i].Start == start {
			return &circuit.WindowBuckets[i]
		}
	}
	circuit.WindowBuckets = append(circuit.WindowBuckets, CircuitWindowBucket{Start: start})
	return &circuit.WindowBuckets[len(circuit.WindowBuckets)-1]
}

// resetWindow 清空滑动窗口，半开探测成功关闸时使用。
func resetWindow(circuit *CircuitBreakerInfo) {
	circuit.WindowBuckets = nil
	circuit.TotalRequests = 0
	circuit.SuccessRequests = 0
	circuit.FailureRequests = 0
	circuit.SlowRequests = 0
	circuit.FailureRate = 0
	circuit.SlowRate = 0
	circuit.WindowStart = 0
	circuit.WindowEnd = 0
	circuit.HalfOpenCount = 0
	circuit.HalfOpenSuccess = 0
}

// syncWindowTotals 用分桶汇总覆盖对外计数，供开闸判断和查询。
func syncWindowTotals(circuit *CircuitBreakerInfo) {
	var total, failures, slow int64
	var start, end int64
	for i, bucket := range circuit.WindowBuckets {
		total += bucket.Total
		failures += bucket.Failures
		slow += bucket.Slow
		if i == 0 || bucket.Start < start {
			start = bucket.Start
		}
		if bucket.Start > end {
			end = bucket.Start
		}
	}
	circuit.TotalRequests = total
	circuit.FailureRequests = failures
	circuit.SlowRequests = slow
	circuit.SuccessRequests = total - failures
	circuit.WindowStart = start
	circuit.WindowEnd = end
	if total > 0 {
		circuit.FailureRate = float64(failures) / float64(total) * 100
		circuit.SlowRate = float64(slow) / float64(total) * 100
		return
	}
	circuit.FailureRate = 0
	circuit.SlowRate = 0
}
