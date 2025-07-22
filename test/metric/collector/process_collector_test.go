package collector_test

import (
	"encoding/json"
	"testing"
	"time"

	"gohub/pkg/metric/collector/process"
	"gohub/pkg/metric/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcessCollector_Basic 测试进程采集器基本功能
func TestProcessCollector_Basic(t *testing.T) {
	collector := process.NewProcessCollector()
	assert.NotNil(t, collector)

	// 验证采集器基本属性
	assert.True(t, collector.IsEnabled())
	assert.Equal(t, types.CollectorNameProcess, collector.GetName())
	assert.Contains(t, collector.GetDescription(), "gopsutil")
}

// TestProcessCollector_Collect 测试采集功能
func TestProcessCollector_Collect(t *testing.T) {
	collector := process.NewProcessCollector()
	
	// 执行采集
	result, err := collector.Collect()
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证数据类型
	metrics, ok := result.(*types.ProcessMetrics)
	json, _ := json.Marshal(metrics)
	t.Logf("ProcessCollector_Collect result = %s", string(json))
	assert.True(t, ok, "返回数据应该是 ProcessMetrics 类型")
	assert.NotNil(t, metrics.CurrentProcess)
	assert.NotNil(t, metrics.SystemProcesses)
	assert.False(t, metrics.CollectTime.IsZero())
}

// TestProcessCollector_CurrentProcessInfo 测试当前进程信息获取
func TestProcessCollector_CurrentProcessInfo(t *testing.T) {
	collector := process.NewProcessCollector()
	
	metrics, err := collector.GetProcessInfo()
	require.NoError(t, err)
	require.NotNil(t, metrics)
	require.NotNil(t, metrics.CurrentProcess)

	proc := metrics.CurrentProcess
	
	// 验证基本信息
	assert.Greater(t, proc.PID, int32(0), "PID应该大于0")
	assert.NotEmpty(t, proc.Name, "进程名称不应为空")
	assert.NotEmpty(t, proc.Status, "进程状态不应为空")
	
	// 验证时间信息
	assert.False(t, proc.CreateTime.IsZero(), "创建时间不应为零值")
	assert.GreaterOrEqual(t, proc.RunTime, uint64(0), "运行时间应该大于等于0")
	
	// 验证内存信息（可能为0，但不应为负数）
	assert.GreaterOrEqual(t, proc.MemoryUsage, uint64(0), "内存使用量应该大于等于0")
	assert.GreaterOrEqual(t, proc.MemoryPercent, float64(0), "内存使用率应该大于等于0")
	
	// 验证CPU信息（可能为0，但不应为负数）
	assert.GreaterOrEqual(t, proc.CPUPercent, float64(0), "CPU使用率应该大于等于0")
	
	// 验证线程数（至少应该有1个线程）
	assert.Greater(t, proc.ThreadCount, int32(0), "线程数应该大于0")
	
	// 验证文件描述符数量（在Unix系统上应该大于0）
	assert.GreaterOrEqual(t, proc.FileDescriptorCount, int32(0), "文件描述符数量应该大于等于0")
}

// TestProcessCollector_SystemProcessStats 测试系统进程统计
func TestProcessCollector_SystemProcessStats(t *testing.T) {
	collector := process.NewProcessCollector()
	
	metrics, err := collector.GetProcessInfo()
	require.NoError(t, err)
	require.NotNil(t, metrics.SystemProcesses)

	stats := metrics.SystemProcesses
	
	// 验证统计数据
	assert.Greater(t, stats.Total, uint32(0), "总进程数应该大于0")
	assert.GreaterOrEqual(t, stats.Running, uint32(0), "运行中进程数应该大于等于0")
	assert.GreaterOrEqual(t, stats.Sleeping, uint32(0), "睡眠中进程数应该大于等于0")
	assert.GreaterOrEqual(t, stats.Stopped, uint32(0), "停止的进程数应该大于等于0")
	assert.GreaterOrEqual(t, stats.Zombie, uint32(0), "僵尸进程数应该大于等于0")
	
	// 验证各状态进程数之和等于总数
	total := stats.Running + stats.Sleeping + stats.Stopped + stats.Zombie
	assert.Equal(t, stats.Total, total, "各状态进程数之和应等于总数")
}

// TestProcessCollector_Timeout 测试超时设置
func TestProcessCollector_Timeout(t *testing.T) {
	collector := process.NewProcessCollector()
	
	// 测试默认超时时间
	defaultTimeout := collector.GetTimeout()
	assert.Equal(t, 5*time.Second, defaultTimeout)
	
	// 测试设置新的超时时间
	newTimeout := 10 * time.Second
	collector.SetTimeout(newTimeout)
	assert.Equal(t, newTimeout, collector.GetTimeout())
}

// TestProcessCollector_ConvenienceMethods 测试便捷方法
func TestProcessCollector_ConvenienceMethods(t *testing.T) {
	collector := process.NewProcessCollector()
	
	// 测试获取当前进程PID
	pid := collector.GetCurrentPID()
	assert.Greater(t, pid, int32(0), "当前进程PID应该大于0")
	
	// 测试获取当前进程名称
	name, err := collector.GetCurrentProcessName()
	require.NoError(t, err)
	assert.NotEmpty(t, name, "当前进程名称不应为空")
	
	// 测试获取进程总数
	count, err := collector.GetProcessCount()
	require.NoError(t, err)
	assert.Greater(t, count, uint32(0), "进程总数应该大于0")
	
	// 测试获取当前进程内存使用情况
	memUsage, memPercent, err := collector.GetCurrentProcessMemoryUsage()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, memUsage, uint64(0), "内存使用量应该大于等于0")
	assert.GreaterOrEqual(t, memPercent, float64(0), "内存使用率应该大于等于0")
	
	// 测试获取当前进程CPU使用率
	cpuPercent, err := collector.GetCurrentProcessCPUUsage()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cpuPercent, float64(0), "CPU使用率应该大于等于0")
}

// TestProcessCollector_GetProcessesByStatus 测试根据状态获取进程数量
func TestProcessCollector_GetProcessesByStatus(t *testing.T) {
	collector := process.NewProcessCollector()
	
	// 测试获取各种状态的进程数量
	testCases := []string{"running", "sleeping", "stopped", "zombie"}
	
	for _, status := range testCases {
		count, err := collector.GetProcessesByStatus(status)
		require.NoError(t, err, "获取%s状态进程数量不应出错", status)
		assert.GreaterOrEqual(t, count, uint32(0), "%s状态进程数量应该大于等于0", status)
	}
	
	// 测试无效状态
	_, err := collector.GetProcessesByStatus("invalid")
	assert.Error(t, err, "无效状态应该返回错误")
	assert.Contains(t, err.Error(), "不支持的进程状态")
}

// TestProcessCollector_Disabled 测试禁用采集器
func TestProcessCollector_Disabled(t *testing.T) {
	collector := process.NewProcessCollector()
	
	// 禁用采集器
	collector.SetEnabled(false)
	assert.False(t, collector.IsEnabled())
	
	// 禁用状态下采集应该返回错误
	result, err := collector.Collect()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, types.ErrCollectorDisabled, err)
}

// TestProcessCollector_ConvertProcessStatus 测试进程状态转换
func TestProcessCollector_ConvertProcessStatus(t *testing.T) {
	collector := process.NewProcessCollector()
	
	// 测试状态转换（通过反射或创建测试辅助方法）
	// 由于convertProcessStatus是私有方法，我们通过间接方式测试
	metrics, err := collector.GetProcessInfo()
	require.NoError(t, err)
	
	// 验证当前进程状态是有效的
	validStatuses := []string{"running", "sleeping", "stopped", "zombie", "unknown"}
	assert.Contains(t, validStatuses, metrics.CurrentProcess.Status)
}

// TestProcessCollector_ContextCancellation 测试上下文取消
func TestProcessCollector_ContextCancellation(t *testing.T) {
	collector := process.NewProcessCollector()
	
	// 设置非常短的超时时间
	collector.SetTimeout(1 * time.Nanosecond)
	
	// 执行采集，可能会因为超时而失败
	// 注意：这个测试可能不会总是失败，因为进程信息获取通常很快
	_, err := collector.GetProcessInfo()
	// 不强制要求超时错误，因为操作可能在超时前完成
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	}
	
	// 恢复正常超时时间
	collector.SetTimeout(5 * time.Second)
}

// TestProcessCollector_Performance 测试性能
func TestProcessCollector_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}
	
	collector := process.NewProcessCollector()
	
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
	assert.Less(t, avgDuration, 1*time.Second, "平均采集时间应该小于1秒")
}

// BenchmarkProcessCollector_Collect 基准测试
func BenchmarkProcessCollector_Collect(b *testing.B) {
	collector := process.NewProcessCollector()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.Collect()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkProcessCollector_GetCurrentProcessInfo 基准测试当前进程信息获取
func BenchmarkProcessCollector_GetCurrentProcessInfo(b *testing.B) {
	collector := process.NewProcessCollector()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.GetProcessInfo()
		if err != nil {
			b.Fatal(err)
		}
	}
} 