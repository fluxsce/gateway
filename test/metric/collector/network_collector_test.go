package collector

import (
	"testing"
	"time"

	"gohub/pkg/metric/collector/network"
	"gohub/pkg/metric/types"

	"github.com/stretchr/testify/assert"
)

func TestNetworkCollector_Basic(t *testing.T) {
	collector := network.NewNetworkCollector()
	assert.NotNil(t, collector)
	assert.Equal(t, types.CollectorNameNetwork, collector.GetName())
	assert.Equal(t, network.DefaultCollectTimeout, collector.GetTimeout())
}

func TestNetworkCollector_Collect(t *testing.T) {
	collector := network.NewNetworkCollector()
	collector.SetEnabled(true)

	// 执行采集
	result, err := collector.Collect()
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// 验证返回类型
	metrics, ok := result.(*types.NetworkMetrics)
	assert.True(t, ok)
	assert.NotNil(t, metrics)

	// 打印网络接口信息
	t.Log("\n=== 网络接口信息 ===")
	for _, iface := range metrics.Interfaces {
		t.Logf("接口名称: %s", iface.Name)
		t.Logf("MAC地址: %s", iface.HardwareAddr)
		t.Logf("IP地址: %v", iface.IPAddresses)
		t.Logf("接口状态: %s", iface.Status)
		t.Logf("接口类型: %s", iface.Type)
		t.Logf("接收字节数: %.2f MB", float64(iface.BytesReceived)/(1024*1024))
		t.Logf("发送字节数: %.2f MB", float64(iface.BytesSent)/(1024*1024))
		t.Logf("接收包数: %d", iface.PacketsReceived)
		t.Logf("发送包数: %d", iface.PacketsSent)
		t.Logf("接收错误数: %d", iface.ErrorsReceived)
		t.Logf("发送错误数: %d", iface.ErrorsSent)
		t.Logf("接收丢包数: %d", iface.DroppedReceived)
		t.Logf("发送丢包数: %d", iface.DroppedSent)
		t.Logf("接收速率: %.2f KB/s", iface.ReceiveRate/1024)
		t.Logf("发送速率: %.2f KB/s", iface.SendRate/1024)
		t.Logf("---")
	}
}

func TestNetworkCollector_NetworkRate(t *testing.T) {
	collector := network.NewNetworkCollector()
	collector.SetEnabled(true)

	// 第一次采集
	result1, err := collector.GetNetworkStats()
	assert.NoError(t, err)
	
	// 等待一段时间
	t.Log("等待3秒钟采集网络流量变化...")
	time.Sleep(3 * time.Second)
	
	// 第二次采集
	result2, err := collector.GetNetworkStats()
	assert.NoError(t, err)

	// 计算这段时间内的网络速率
	t.Log("\n=== 网络速率(3秒内) ===")
	for _, iface1 := range result1.Interfaces {
		// 查找对应的接口
		var iface2 *types.NetworkInterface
		for j := range result2.Interfaces {
			if result2.Interfaces[j].Name == iface1.Name {
				iface2 = &result2.Interfaces[j]
				break
			}
		}
		
		if iface2 == nil {
			continue
		}
		
		duration := iface2.LastCollectTime.Sub(iface1.LastCollectTime).Seconds()
		if duration <= 0 {
			continue
		}
		
		receiveRate := float64(iface2.BytesReceived-iface1.BytesReceived) / duration / 1024 // KB/s
		sendRate := float64(iface2.BytesSent-iface1.BytesSent) / duration / 1024 // KB/s
		
		t.Logf("接口: %s", iface1.Name)
		t.Logf("接收速率: %.2f KB/s", receiveRate)
		t.Logf("发送速率: %.2f KB/s", sendRate)
		t.Logf("内部计算的接收速率: %.2f KB/s", iface2.ReceiveRate/1024)
		t.Logf("内部计算的发送速率: %.2f KB/s", iface2.SendRate/1024)
		t.Logf("---")
	}
}

func TestNetworkCollector_ActiveInterfaces(t *testing.T) {
	collector := network.NewNetworkCollector()
	collector.SetEnabled(true)

	// 获取活动接口
	interfaces, err := collector.GetActiveInterfaces()
	assert.NoError(t, err)
	
	t.Log("\n=== 活动网络接口 ===")
	for _, iface := range interfaces {
		t.Logf("接口名称: %s", iface.Name)
		t.Logf("接口类型: %s", iface.Type)
		t.Logf("IP地址: %v", iface.IPAddresses)
		t.Logf("---")
	}
}

func TestNetworkCollector_TotalTraffic(t *testing.T) {
	collector := network.NewNetworkCollector()
	collector.SetEnabled(true)

	// 获取总流量
	bytesReceived, bytesSent, err := collector.GetTotalNetworkTraffic()
	assert.NoError(t, err)
	
	t.Log("\n=== 总网络流量 ===")
	t.Logf("总接收流量: %.2f MB", float64(bytesReceived)/(1024*1024))
	t.Logf("总发送流量: %.2f MB", float64(bytesSent)/(1024*1024))
}

func TestNetworkCollector_InterfaceByName(t *testing.T) {
	collector := network.NewNetworkCollector()
	collector.SetEnabled(true)
	
	// 获取所有接口
	metrics, err := collector.GetNetworkStats()
	assert.NoError(t, err)
	
	if len(metrics.Interfaces) > 0 {
		// 选择第一个接口进行测试
		firstInterface := metrics.Interfaces[0]
		
		// 测试GetInterfaceByName方法
		iface, err := collector.GetInterfaceByName(firstInterface.Name)
		assert.NoError(t, err)
		assert.NotNil(t, iface)
		assert.Equal(t, firstInterface.Name, iface.Name)
		
		t.Logf("\n=== 接口详情: %s ===", iface.Name)
		t.Logf("MAC地址: %s", iface.HardwareAddr)
		t.Logf("状态: %s", iface.Status)
		t.Logf("类型: %s", iface.Type)
	} else {
		t.Skip("没有可用的网络接口，跳过测试")
	}
}
