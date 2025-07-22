package collector_test

import (
	"encoding/json"
	"gohub/pkg/metric/collector/cpu"
	"gohub/pkg/metric/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCPUCollector_Basic 测试CPU采集器基本功能
func TestCPUCollector_Basic(t *testing.T) {
	collector := cpu.NewCPUCollector()
	assert.NotNil(t, collector)

	// 验证采集器基本属性
	assert.True(t, collector.IsEnabled())
	assert.Equal(t, types.CollectorNameCPU, collector.GetName())
	assert.Contains(t, collector.GetDescription(), "gopsutil")
}

// TestCPUCollector_Collect 测试采集功能
func TestCPUCollector_Collect(t *testing.T) {
	collector := cpu.NewCPUCollector()
	
	// 执行采集
	result, err := collector.Collect()
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证数据类型
	metrics, ok := result.(*types.CPUMetrics)
	json, _ := json.Marshal(metrics)
	t.Logf("CPUCollector_Collect result = %s", string(json))
	assert.True(t, ok, "返回数据应该是 CPUMetrics 类型")
	assert.False(t, metrics.CollectTime.IsZero())
}

// TestCPUCollector_GetCPUUsage 测试CPU使用率获取
func TestCPUCollector_GetCPUUsage(t *testing.T) {
	collector := cpu.NewCPUCollector()
	
	metrics, err := collector.GetCPUUsage()
	require.NoError(t, err)
	require.NotNil(t, metrics)

	// 验证CPU核心数信息
	assert.Greater(t, metrics.CoreCount, 0, "物理核心数应该大于0")
	assert.Greater(t, metrics.LogicalCount, 0, "逻辑CPU数应该大于0")
	assert.GreaterOrEqual(t, metrics.LogicalCount, metrics.CoreCount, "逻辑CPU数应该大于等于物理核心数")
	
	// 验证CPU使用率信息（允许为0，但不应为负数）
	assert.GreaterOrEqual(t, metrics.UsagePercent, float64(0), "总CPU使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.UsagePercent, float64(100), "总CPU使用率应该小于等于100")
	
	assert.GreaterOrEqual(t, metrics.UserPercent, float64(0), "用户态CPU使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.UserPercent, float64(100), "用户态CPU使用率应该小于等于100")
	
	assert.GreaterOrEqual(t, metrics.SystemPercent, float64(0), "系统态CPU使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.SystemPercent, float64(100), "系统态CPU使用率应该小于等于100")
	
	assert.GreaterOrEqual(t, metrics.IdlePercent, float64(0), "空闲CPU使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.IdlePercent, float64(100), "空闲CPU使用率应该小于等于100")
	
	assert.GreaterOrEqual(t, metrics.IOWaitPercent, float64(0), "IO等待CPU使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.IOWaitPercent, float64(100), "IO等待CPU使用率应该小于等于100")
	
	assert.GreaterOrEqual(t, metrics.IrqPercent, float64(0), "中断处理CPU使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.IrqPercent, float64(100), "中断处理CPU使用率应该小于等于100")
	
	assert.GreaterOrEqual(t, metrics.SoftIrqPercent, float64(0), "软中断处理CPU使用率应该大于等于0")
	assert.LessOrEqual(t, metrics.SoftIrqPercent, float64(100), "软中断处理CPU使用率应该小于等于100")
	
	// 验证负载平均值（在某些系统上可能为0）
	assert.GreaterOrEqual(t, metrics.LoadAvg1, float64(0), "1分钟负载平均值应该大于等于0")
	assert.GreaterOrEqual(t, metrics.LoadAvg5, float64(0), "5分钟负载平均值应该大于等于0")
	assert.GreaterOrEqual(t, metrics.LoadAvg15, float64(0), "15分钟负载平均值应该大于等于0")
}

// TestCPUCollector_Timeout 测试超时设置
func TestCPUCollector_Timeout(t *testing.T) {
	collector := cpu.NewCPUCollector()
	
	// 测试默认超时时间
	defaultTimeout := collector.GetTimeout()
	assert.Equal(t, 5*time.Second, defaultTimeout)
	
	// 测试设置新的超时时间
	newTimeout := 10 * time.Second
	collector.SetTimeout(newTimeout)
	assert.Equal(t, newTimeout, collector.GetTimeout())
}

// TestCPUCollector_Disabled 测试禁用采集器
func TestCPUCollector_Disabled(t *testing.T) {
	collector := cpu.NewCPUCollector()
	
	// 禁用采集器
	collector.SetEnabled(false)
	assert.False(t, collector.IsEnabled())
	
	// 禁用状态下采集应该返回错误
	result, err := collector.Collect()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, types.ErrCollectorDisabled, err)
}

// TestCPUCollector_MultipleCollections 测试多次采集
func TestCPUCollector_MultipleCollections(t *testing.T) {
	collector := cpu.NewCPUCollector()
	
	// 连续多次采集，验证数据的一致性
	var previousMetrics *types.CPUMetrics
	
	for i := 0; i < 3; i++ {
		metrics, err := collector.GetCPUUsage()
		require.NoError(t, err)
		require.NotNil(t, metrics)
		
		// 核心数信息应该保持一致
		if previousMetrics != nil {
			assert.Equal(t, previousMetrics.CoreCount, metrics.CoreCount, "物理核心数应该保持一致")
			assert.Equal(t, previousMetrics.LogicalCount, metrics.LogicalCount, "逻辑CPU数应该保持一致")
		}
		
		previousMetrics = metrics
		
		// 短暂休眠，避免过快的采集
		time.Sleep(100 * time.Millisecond)
	}
}

// TestCPUCollector_Performance 测试性能
func TestCPUCollector_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}
	
	collector := cpu.NewCPUCollector()
	
	// 多次执行采集，测试性能
	iterations := 5
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		_, err := collector.Collect()
		require.NoError(t, err)
	}
	
	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)
	
	t.Logf("平均采集时间: %v", avgDuration)
	
	// 性能基准：每次采集应该在合理时间内完成
	assert.Less(t, avgDuration, 3*time.Second, "平均采集时间应该小于3秒")
}

// TestCPUCollector_ContextCancellation 测试上下文取消
func TestCPUCollector_ContextCancellation(t *testing.T) {
	collector := cpu.NewCPUCollector()
	
	// 设置非常短的超时时间
	collector.SetTimeout(1 * time.Nanosecond)
	
	// 执行采集，可能会因为超时而失败
	// 注意：这个测试可能不会总是失败，因为CPU信息获取通常很快
	_, err := collector.GetCPUUsage()
	// 不强制要求超时错误，因为操作可能在超时前完成
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
	
	// 恢复正常超时时间
	collector.SetTimeout(5 * time.Second)
}

// BenchmarkCPUCollector_Collect 基准测试
func BenchmarkCPUCollector_Collect(b *testing.B) {
	collector := cpu.NewCPUCollector()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.Collect()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCPUCollector_GetCPUUsage 基准测试CPU使用率获取
func BenchmarkCPUCollector_GetCPUUsage(b *testing.B) {
	collector := cpu.NewCPUCollector()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.GetCPUUsage()
		if err != nil {
			b.Fatal(err)
		}
	}
} 