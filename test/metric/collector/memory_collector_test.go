package collector_test

import (
	"encoding/json"
	"testing"
	"time"

	"gohub/pkg/metric/collector/memory"
	"gohub/pkg/metric/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryCollector_Basic 测试内存采集器基本功能
func TestMemoryCollector_Basic(t *testing.T) {
	collector := memory.NewMemoryCollector()
	assert.NotNil(t, collector)

	// 验证采集器基本属性
	assert.True(t, collector.IsEnabled())
	assert.Equal(t, types.CollectorNameMemory, collector.GetName())
	assert.Contains(t, collector.GetDescription(), "gopsutil")
}

// TestMemoryCollector_Collect 测试采集功能
func TestMemoryCollector_Collect(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	// 执行采集
	result, err := collector.Collect()
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证数据类型
	metrics, ok := result.(*types.MemoryMetrics)
	json, _ := json.Marshal(metrics)
	t.Logf("MemoryCollector_Collect result = %s", string(json))
	assert.True(t, ok, "返回数据应该是 MemoryMetrics 类型")
	assert.False(t, metrics.CollectTime.IsZero())
}

// TestMemoryCollector_GetMemoryUsage 测试内存使用率获取
func TestMemoryCollector_GetMemoryUsage(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	metrics, err := collector.GetMemoryUsage()
	require.NoError(t, err)
	require.NotNil(t, metrics)

	// 验证内存基本信息
	assert.Greater(t, metrics.Total, uint64(0), "总内存应该大于0")
	assert.GreaterOrEqual(t, metrics.Available, uint64(0), "可用内存应该大于等于0")
	assert.GreaterOrEqual(t, metrics.Used, uint64(0), "已使用内存应该大于等于0")
	assert.GreaterOrEqual(t, metrics.Free, uint64(0), "空闲内存应该大于等于0")
	
	// 验证内存使用率范围
	assert.GreaterOrEqual(t, metrics.UsagePercent, float64(0), "内存使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.UsagePercent, float64(100), "内存使用率应该小于等于100")
	
	// 验证逻辑关系
	assert.LessOrEqual(t, metrics.Used, metrics.Total, "已使用内存不应超过总内存")
	assert.LessOrEqual(t, metrics.Available, metrics.Total, "可用内存不应超过总内存")
	
	// 验证缓存和缓冲区信息（可能为0）
	assert.GreaterOrEqual(t, metrics.Cached, uint64(0), "缓存内存应该大于等于0")
	assert.GreaterOrEqual(t, metrics.Buffers, uint64(0), "缓冲区内存应该大于等于0")
	assert.GreaterOrEqual(t, metrics.Shared, uint64(0), "共享内存应该大于等于0")
	
	// 验证交换区信息
	assert.GreaterOrEqual(t, metrics.SwapTotal, uint64(0), "交换区总大小应该大于等于0")
	assert.GreaterOrEqual(t, metrics.SwapUsed, uint64(0), "交换区已使用应该大于等于0")
	assert.GreaterOrEqual(t, metrics.SwapFree, uint64(0), "交换区空闲应该大于等于0")
	assert.GreaterOrEqual(t, metrics.SwapUsagePercent, float64(0), "交换区使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.SwapUsagePercent, float64(100), "交换区使用率应该小于等于100")
	
	// 验证交换区逻辑关系
	if metrics.SwapTotal > 0 {
		assert.LessOrEqual(t, metrics.SwapUsed, metrics.SwapTotal, "交换区已使用不应超过总大小")
		assert.LessOrEqual(t, metrics.SwapFree, metrics.SwapTotal, "交换区空闲不应超过总大小")
		
		// 验证交换区使用率计算是否正确
		expectedUsagePercent := float64(metrics.SwapUsed) / float64(metrics.SwapTotal) * 100
		assert.InDelta(t, expectedUsagePercent, metrics.SwapUsagePercent, 1.0, "交换区使用率计算应该正确")
	}
}

// TestMemoryCollector_Timeout 测试超时设置
func TestMemoryCollector_Timeout(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	// 测试默认超时时间
	defaultTimeout := collector.GetTimeout()
	assert.Equal(t, 5*time.Second, defaultTimeout)
	
	// 测试设置新的超时时间
	newTimeout := 10 * time.Second
	collector.SetTimeout(newTimeout)
	assert.Equal(t, newTimeout, collector.GetTimeout())
}

// TestMemoryCollector_ConvenienceMethods 测试便捷方法
func TestMemoryCollector_ConvenienceMethods(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	// 测试获取内存使用率百分比
	usagePercent, err := collector.GetMemoryUsagePercent()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, usagePercent, float64(0), "内存使用率应该大于等于0")
	assert.LessOrEqual(t, usagePercent, float64(100), "内存使用率应该小于等于100")
	
	// 测试获取可用内存
	availableMemory, err := collector.GetAvailableMemory()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, availableMemory, uint64(0), "可用内存应该大于等于0")
	
	// 测试获取已使用内存
	usedMemory, err := collector.GetUsedMemory()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, usedMemory, uint64(0), "已使用内存应该大于等于0")
	
	// 测试获取交换区使用率
	swapUsagePercent, err := collector.GetSwapUsagePercent()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, swapUsagePercent, float64(0), "交换区使用率应该大于等于0")
	assert.LessOrEqual(t, swapUsagePercent, float64(100), "交换区使用率应该小于等于100")
}

// TestMemoryCollector_Disabled 测试禁用采集器
func TestMemoryCollector_Disabled(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	// 禁用采集器
	collector.SetEnabled(false)
	assert.False(t, collector.IsEnabled())
	
	// 禁用状态下采集应该返回错误
	result, err := collector.Collect()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, types.ErrCollectorDisabled, err)
}

// TestMemoryCollector_MultipleCollections 测试多次采集
func TestMemoryCollector_MultipleCollections(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	// 连续多次采集，验证数据的稳定性
	var previousMetrics *types.MemoryMetrics
	
	for i := 0; i < 3; i++ {
		metrics, err := collector.GetMemoryUsage()
		require.NoError(t, err)
		require.NotNil(t, metrics)
		
		// 总内存应该保持一致
		if previousMetrics != nil {
			assert.Equal(t, previousMetrics.Total, metrics.Total, "总内存应该保持一致")
			
			// 交换区总大小应该保持一致
			if previousMetrics.SwapTotal > 0 {
				assert.Equal(t, previousMetrics.SwapTotal, metrics.SwapTotal, "交换区总大小应该保持一致")
			}
		}
		
		previousMetrics = metrics
		
		// 短暂休眠，避免过快的采集
		time.Sleep(100 * time.Millisecond)
	}
}

// TestMemoryCollector_MemoryCalculations 测试内存计算逻辑
func TestMemoryCollector_MemoryCalculations(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	metrics, err := collector.GetMemoryUsage()
	require.NoError(t, err)
	require.NotNil(t, metrics)
	
	// 验证使用率计算是否在合理范围内
	if metrics.Total > 0 {
		calculatedUsagePercent := float64(metrics.Used) / float64(metrics.Total) * 100
		
		// 允许一定的误差范围，因为不同的计算方式可能有细微差别
		assert.InDelta(t, calculatedUsagePercent, metrics.UsagePercent, 10.0, 
			"内存使用率计算应该在合理范围内")
	}
	
	// 验证可用内存是否合理
	if metrics.Available > 0 {
		assert.LessOrEqual(t, metrics.Available, metrics.Total, 
			"可用内存不应超过总内存")
	}
}

// TestMemoryCollector_Performance 测试性能
func TestMemoryCollector_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}
	
	collector := memory.NewMemoryCollector()
	
	// 多次执行采集，测试性能
	iterations := 10
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		_, err := collector.Collect()
		require.NoError(t, err)
	}
	
	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)
	
	t.Logf("平均采集时间: %v", avgDuration)
	
	// 性能基准：每次采集应该在合理时间内完成
	assert.Less(t, avgDuration, 500*time.Millisecond, "平均采集时间应该小于500毫秒")
}

// TestMemoryCollector_ContextCancellation 测试上下文取消
func TestMemoryCollector_ContextCancellation(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	// 设置非常短的超时时间
	collector.SetTimeout(1 * time.Nanosecond)
	
	// 执行采集，可能会因为超时而失败
	// 注意：这个测试可能不会总是失败，因为内存信息获取通常很快
	_, err := collector.GetMemoryUsage()
	// 不强制要求超时错误，因为操作可能在超时前完成
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
	
	// 恢复正常超时时间
	collector.SetTimeout(5 * time.Second)
}

// TestMemoryCollector_ConsistencyCheck 测试数据一致性
func TestMemoryCollector_ConsistencyCheck(t *testing.T) {
	collector := memory.NewMemoryCollector()
	
	// 执行多次采集，检查数据一致性
	const iterations = 5
	var totalMemoryValues []uint64
	
	for i := 0; i < iterations; i++ {
		metrics, err := collector.GetMemoryUsage()
		require.NoError(t, err)
		totalMemoryValues = append(totalMemoryValues, metrics.Total)
		
		time.Sleep(50 * time.Millisecond)
	}
	
	// 总内存在短时间内应该保持一致
	firstTotal := totalMemoryValues[0]
	for i, total := range totalMemoryValues {
		assert.Equal(t, firstTotal, total, "第%d次采集的总内存应该与第一次一致", i+1)
	}
}

// BenchmarkMemoryCollector_Collect 基准测试
func BenchmarkMemoryCollector_Collect(b *testing.B) {
	collector := memory.NewMemoryCollector()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.Collect()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryCollector_GetMemoryUsage 基准测试内存使用率获取
func BenchmarkMemoryCollector_GetMemoryUsage(b *testing.B) {
	collector := memory.NewMemoryCollector()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.GetMemoryUsage()
		if err != nil {
			b.Fatal(err)
		}
	}
} 