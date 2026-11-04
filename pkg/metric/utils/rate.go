package utils

// CounterRate 计算累计计数在一段时间上的每秒增量。
// 间隔无效，或当前值小于上次值（计数器回绕、设备重置）时返回 0。
// uint64 相减在回绕时会变成极大正数，不能靠结果是否为负来判断。
func CounterRate(current, previous uint64, seconds float64) float64 {
	if seconds <= 0 || current < previous {
		return 0
	}
	return float64(current-previous) / seconds
}
