package dao

import "math"

// calculatePercentile 计算百分位数
// 使用线性插值方法计算精确的百分位数值
// values 必须是已排序的数组
func calculatePercentile(values []int, percentile float64) int {
	if len(values) == 0 {
		return 0
	}

	// 计算索引位置
	index := percentile * float64(len(values)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	// 如果索引相同，直接返回该值
	if lower == upper {
		return values[lower]
	}

	// 线性插值计算
	weight := index - float64(lower)
	return int(float64(values[lower])*(1-weight) + float64(values[upper])*weight)
}

// roundToTwoDecimalPlaces 将浮点数最多保留两位小数
// 自动去除不必要的尾随零
//
// 示例：
//   - 120.00 -> 120
//   - 120.50 -> 120.5
//   - 120.56 -> 120.56
//   - 120.567 -> 120.57 (四舍五入)
func roundToTwoDecimalPlaces(value float64) float64 {
	if value == 0 {
		return 0
	}

	// 先四舍五入到两位小数
	rounded := math.Round(value*100) / 100

	return rounded
}
