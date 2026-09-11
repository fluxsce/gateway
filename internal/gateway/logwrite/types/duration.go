package types

import "time"

// DurationMillis 把时长换成日志用的毫秒。
// 大于 0 且不足 1ms 记 1，避免真实耗时被截成 0；满 1ms 后按四舍五入。
// d<=0 记 0，表示没有发生这段耗时。
func DurationMillis(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	ms := (d + time.Millisecond/2) / time.Millisecond
	if ms < 1 {
		return 1
	}
	return int(ms)
}

// ElapsedMillis 计算 end-start 的日志毫秒；任一端为零时间或顺序颠倒则返回 0。
func ElapsedMillis(start, end time.Time) int {
	if start.IsZero() || end.IsZero() {
		return 0
	}
	return DurationMillis(end.Sub(start))
}
