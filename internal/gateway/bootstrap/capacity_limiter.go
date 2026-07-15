package bootstrap

import "sync/atomic"

// requestAdmissionLimiter 限制整个网关当前正在处理的HTTP请求数量。
// 上限可在热重载时动态调整；已进入处理链的请求不会因上限降低而被中断。
type requestAdmissionLimiter struct {
	limit    atomic.Int64
	active   atomic.Int64
	rejected atomic.Uint64
}

func (l *requestAdmissionLimiter) setLimit(limit int) {
	l.limit.Store(int64(limit))
}

func (l *requestAdmissionLimiter) tryAcquire() bool {
	for {
		current := l.active.Load()
		limit := l.limit.Load()
		if limit > 0 && current >= limit {
			l.rejected.Add(1)
			return false
		}
		if l.active.CompareAndSwap(current, current+1) {
			return true
		}
	}
}

func (l *requestAdmissionLimiter) release() {
	l.active.Add(-1)
}

func (l *requestAdmissionLimiter) activeCount() int64 {
	return l.active.Load()
}

func (l *requestAdmissionLimiter) rejectedCount() uint64 {
	return l.rejected.Load()
}
