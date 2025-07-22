package collector_test

import (
	"encoding/json"
	"testing"
	"time"

	"gohub/pkg/metric/collector/system"
	"gohub/pkg/metric/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystemCollector_Basic 测试系统采集器基本功能
func TestSystemCollector_Basic(t *testing.T) {
	collector := system.NewSystemCollector()
	assert.NotNil(t, collector)

	// 验证采集器基本属性
	assert.True(t, collector.IsEnabled())
	assert.Equal(t, types.CollectorNameSystem, collector.GetName())
	assert.Contains(t, collector.GetDescription(), "gopsutil")
}

// TestSystemCollector_Collect 测试采集功能
func TestSystemCollector_Collect(t *testing.T) {
	collector := system.NewSystemCollector()
	
	// 执行采集
	result, err := collector.Collect()
	require.NoError(t, err)
	require.NotNil(t, result)
	json, _ := json.Marshal(result)
	t.Logf("SystemCollector_Collect result = %s", string(json))
	// 验证数据类型
	metrics, ok := result.(*types.SystemMetrics)
	assert.True(t, ok, "返回数据应该是 SystemMetrics 类型")
	assert.False(t, metrics.CollectTime.IsZero())
}

// TestSystemCollector_GetSystemInfo 测试系统信息获取
func TestSystemCollector_GetSystemInfo(t *testing.T) {
	collector := system.NewSystemCollector()
	
	metrics, err := collector.GetSystemInfo()
	require.NoError(t, err)
	require.NotNil(t, metrics)

	// 验证主机信息
	assert.NotEmpty(t, metrics.Hostname, "主机名不应为空")
	assert.NotEmpty(t, metrics.OS, "操作系统不应为空")
	assert.NotEmpty(t, metrics.Architecture, "架构信息不应为空")
	assert.NotEmpty(t, metrics.KernelVersion, "内核版本不应为空")
	assert.NotEmpty(t, metrics.OSVersion, "操作系统版本不应为空")
	
	// 验证运行时间信息
	assert.GreaterOrEqual(t, metrics.Uptime, uint64(0), "运行时间应该大于等于0")
	assert.False(t, metrics.BootTime.IsZero(), "启动时间不应为零值")
	
	// 验证统计信息
	assert.Greater(t, metrics.ProcessCount, uint32(0), "进程数应该大于0")
	assert.Greater(t, metrics.UserCount, uint32(0), "用户数应该大于0")
	
	// 验证温度信息（可能为空数组）
	assert.NotNil(t, metrics.Temperature, "温度信息数组不应为nil")
}

// TestSystemCollector_Timeout 测试超时设置
func TestSystemCollector_Timeout(t *testing.T) {
	collector := system.NewSystemCollector()
	
	// 测试默认超时时间
	defaultTimeout := collector.GetTimeout()
	assert.Equal(t, 5*time.Second, defaultTimeout)
	
	// 测试设置新的超时时间
	newTimeout := 10 * time.Second
	collector.SetTimeout(newTimeout)
	assert.Equal(t, newTimeout, collector.GetTimeout())
}

// TestSystemCollector_ConvenienceMethods 测试便捷方法
func TestSystemCollector_ConvenienceMethods(t *testing.T) {
	collector := system.NewSystemCollector()
	
	// 测试获取主机名
	hostname, err := collector.GetHostname()
	require.NoError(t, err)
	assert.NotEmpty(t, hostname, "主机名不应为空")
	
	// 测试获取操作系统信息
	osType, arch, kernel, err := collector.GetOSInfo()
	require.NoError(t, err)
	assert.NotEmpty(t, osType, "操作系统类型不应为空")
	assert.NotEmpty(t, arch, "架构信息不应为空")
	assert.NotEmpty(t, kernel, "内核版本不应为空")
	
	// 测试获取运行时间
	uptime, err := collector.GetUptime()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, uptime, uint64(0), "运行时间应该大于等于0")
	
	// 测试获取启动时间
	bootTime, err := collector.GetBootTime()
	require.NoError(t, err)
	assert.False(t, bootTime.IsZero(), "启动时间不应为零值")
	
	// 测试获取进程数量
	processCount, err := collector.GetProcessCount()
	require.NoError(t, err)
	assert.Greater(t, processCount, uint32(0), "进程数应该大于0")
	
	// 测试获取用户数量
	userCount, err := collector.GetUserCount()
	require.NoError(t, err)
	assert.Greater(t, userCount, uint32(0), "用户数应该大于0")
	
	// 测试获取温度信息
	temperatures, err := collector.GetTemperatures()
	require.NoError(t, err)
	assert.NotNil(t, temperatures, "温度信息不应为nil")
	// 温度信息可能为空数组，这是正常的
}

// TestSystemCollector_Disabled 测试禁用采集器
func TestSystemCollector_Disabled(t *testing.T) {
	collector := system.NewSystemCollector()
	
	// 禁用采集器
	collector.SetEnabled(false)
	assert.False(t, collector.IsEnabled())
	
	// 禁用状态下采集应该返回错误
	result, err := collector.Collect()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, types.ErrCollectorDisabled, err)
}

// TestSystemCollector_Performance 测试性能
func TestSystemCollector_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}
	
	collector := system.NewSystemCollector()
	
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
	assert.Less(t, avgDuration, 2*time.Second, "平均采集时间应该小于2秒")
}

// BenchmarkSystemCollector_Collect 基准测试
func BenchmarkSystemCollector_Collect(b *testing.B) {
	collector := system.NewSystemCollector()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.Collect()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSystemCollector_GetSystemInfo 基准测试系统信息获取
func BenchmarkSystemCollector_GetSystemInfo(b *testing.B) {
	collector := system.NewSystemCollector()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := collector.GetSystemInfo()
		if err != nil {
			b.Fatal(err)
		}
	}
} 

// func TestIsCommonPrivateIPv4Address(t *testing.T) {
// 	collector := system.NewSystemCollector()
	
// 	testCases := []struct {
// 		ip       string
// 		expected bool
// 		desc     string
// 	}{
// 		{"192.168.1.1", true, "192.168.x.x should be common private"},
// 		{"192.168.0.100", true, "192.168.x.x should be common private"},
// 		{"10.0.0.1", true, "10.x.x.x should be common private"},
// 		{"10.1.1.1", true, "10.x.x.x should be common private"},
// 		{"172.16.0.1", true, "172.16.x.x should be common private"},
// 		{"172.31.255.255", true, "172.31.x.x should be common private"},
// 		{"2.0.0.1", false, "2.0.0.1 should not be common private"},
// 		{"8.8.8.8", false, "8.8.8.8 should not be common private"},
// 		{"127.0.0.1", false, "127.0.0.1 should not be common private"},
// 		{"169.254.1.1", false, "169.254.x.x should not be common private"},
// 	}
	
// 	for _, tc := range testCases {
// 		t.Run(tc.desc, func(t *testing.T) {
// 			result := collector.IsCommonPrivateIPv4Address(tc.ip)
// 			if result != tc.expected {
// 				t.Errorf("isCommonPrivateIPv4Address(%s) = %v, expected %v", tc.ip, result, tc.expected)
// 			}
// 		})
// 	}
// }

// func TestIsIPv4Address(t *testing.T) {
// 	collector := system.NewSystemCollector()
	
// 	testCases := []struct {
// 		ip       string
// 		expected bool
// 		desc     string
// 	}{
// 		{"192.168.1.1", true, "IPv4 address"},
// 		{"10.0.0.1", true, "IPv4 address"},
// 		{"2.0.0.1", true, "IPv4 address"},
// 		{"::1", false, "IPv6 address"},
// 		{"fe80::1", false, "IPv6 address"},
// 		{"2001:db8::1", false, "IPv6 address"},
// 		{"invalid", false, "Invalid address"},
// 	}
	
// 	for _, tc := range testCases {
// 		t.Run(tc.desc, func(t *testing.T) {
// 			result := collector.isIPv4Address(tc.ip)
// 			if result != tc.expected {
// 				t.Errorf("isIPv4Address(%s) = %v, expected %v", tc.ip, result, tc.expected)
// 			}
// 		})
// 	}
// } 