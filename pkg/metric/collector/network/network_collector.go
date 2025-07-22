package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"gateway/pkg/metric/collector"
	"gateway/pkg/metric/types"

	gopsutilNet "github.com/shirou/gopsutil/v4/net"
)

// 默认采集超时时间
const DefaultCollectTimeout = 5 * time.Second

// NetworkCollector 网络采集器
// 使用 gopsutil 实现跨平台网络信息采集
type NetworkCollector struct {
	*collector.BaseCollector
	// 采集超时时间
	timeout time.Duration
	// 上次网络接口统计数据
	lastInterfaceStats map[string]types.NetworkInterface
	// 互斥锁保护lastInterfaceStats
	mu sync.RWMutex
	// 是否是第一次采集
	isFirstCollect bool
}

// NewNetworkCollector 创建网络采集器
// 使用 gopsutil 实现跨平台网络信息采集
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{
		BaseCollector: collector.NewBaseCollector(
			types.CollectorNameNetwork,
			"基于gopsutil的网络资源采集器，提供跨平台网络信息采集",
		),
		timeout:            DefaultCollectTimeout,
		lastInterfaceStats: make(map[string]types.NetworkInterface),
		isFirstCollect:     true,
	}
}

// Collect 执行采集
// 统一的采集入口，使用gopsutil简化实现
func (c *NetworkCollector) Collect() (interface{}, error) {
	if !c.IsEnabled() {
		return nil, types.ErrCollectorDisabled
	}

	metrics, err := c.GetNetworkStats()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", types.ErrCollectFailed, err)
	}

	c.SetLastCollectTime(time.Now())
	return metrics, nil
}

// GetNetworkStats 获取网络统计信息
// 使用 gopsutil 采集网络接口信息、流量统计等
func (c *NetworkCollector) GetNetworkStats() (*types.NetworkMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	metrics := &types.NetworkMetrics{
		CollectTime: time.Now(),
	}

	// 获取网络接口信息
	interfaces, err := c.getNetworkInterfaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取网络接口信息失败: %w", err)
	}
	metrics.Interfaces = interfaces

	// 第一次采集后设置标志为false
	if c.isFirstCollect {
		c.isFirstCollect = false
	}

	return metrics, nil
}

// getNetworkInterfaces 获取网络接口信息
// 使用 gopsutil 简化跨平台网络接口信息获取
func (c *NetworkCollector) getNetworkInterfaces(ctx context.Context) ([]types.NetworkInterface, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var interfaces []types.NetworkInterface
	now := time.Now()

	// 使用 gopsutil 获取网络接口信息
	gopsutilInterfaces, err := gopsutilNet.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("gopsutil获取网络接口失败: %w", err)
	}

	// 获取IO统计信息
	ioCounters, err := gopsutilNet.IOCountersWithContext(ctx, true)
	if err != nil {
		// IO统计获取失败不影响基础信息
		ioCounters = []gopsutilNet.IOCountersStat{}
	}

	// 创建IO统计映射表
	ioStatsMap := make(map[string]gopsutilNet.IOCountersStat)
	for _, io := range ioCounters {
		ioStatsMap[io.Name] = io
	}

	// 获取读写锁
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, gopsutilInterface := range gopsutilInterfaces {
		// 获取接口地址信息
		addresses, err := c.getInterfaceAddresses(ctx, gopsutilInterface.Name)
		if err != nil {
			// 地址获取失败不影响其他信息
			addresses = []string{}
		}

		// 获取接口统计信息
		ioStats, exists := ioStatsMap[gopsutilInterface.Name]
		var interfaceStats types.NetworkInterface
		if exists {
			interfaceStats = types.NetworkInterface{
				Name:            gopsutilInterface.Name,
				HardwareAddr:    gopsutilInterface.HardwareAddr,
				IPAddresses:     addresses,
				Status:          c.getInterfaceStatus(gopsutilInterface.Flags),
				Type:            c.getInterfaceType(gopsutilInterface.Name),
				BytesReceived:   ioStats.BytesRecv,
				BytesSent:       ioStats.BytesSent,
				PacketsReceived: ioStats.PacketsRecv,
				PacketsSent:     ioStats.PacketsSent,
				ErrorsReceived:  ioStats.Errin,
				ErrorsSent:      ioStats.Errout,
				DroppedReceived: ioStats.Dropin,
				DroppedSent:     ioStats.Dropout,
				ReceiveRate:     0, // 默认值，将在下面计算
				SendRate:        0, // 默认值，将在下面计算
				LastCollectTime: now,
			}

			// 计算网络速率
			if !c.isFirstCollect {
				if last, exists := c.lastInterfaceStats[gopsutilInterface.Name]; exists {
					duration := now.Sub(last.LastCollectTime).Seconds()
					if duration > 0 {
						// 计算接收速率 (字节/秒)
						interfaceStats.ReceiveRate = float64(interfaceStats.BytesReceived-last.BytesReceived) / duration
						// 计算发送速率 (字节/秒)
						interfaceStats.SendRate = float64(interfaceStats.BytesSent-last.BytesSent) / duration

						// 确保速率不为负数（可能由于计数器重置导致）
						if interfaceStats.ReceiveRate < 0 {
							interfaceStats.ReceiveRate = 0
						}
						if interfaceStats.SendRate < 0 {
							interfaceStats.SendRate = 0
						}
					}
				}
			}

			// 保存当前统计数据用于下次计算
			c.lastInterfaceStats[gopsutilInterface.Name] = interfaceStats
		} else {
			interfaceStats = types.NetworkInterface{
				Name:            gopsutilInterface.Name,
				HardwareAddr:    gopsutilInterface.HardwareAddr,
				IPAddresses:     addresses,
				Status:          c.getInterfaceStatus(gopsutilInterface.Flags),
				Type:            c.getInterfaceType(gopsutilInterface.Name),
				LastCollectTime: now,
			}

			// 即使没有IO统计，也保存基本信息用于下次比较
			c.lastInterfaceStats[gopsutilInterface.Name] = interfaceStats
		}

		interfaces = append(interfaces, interfaceStats)
	}

	return interfaces, nil
}

// getInterfaceAddresses 获取接口地址信息
// 使用 gopsutil 获取指定接口的IP地址
func (c *NetworkCollector) getInterfaceAddresses(ctx context.Context, interfaceName string) ([]string, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var addresses []string

	// 使用标准库获取接口地址（gopsutil目前没有直接提供这个功能）
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return addresses, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return addresses, err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP != nil {
			ipStr := ipNet.IP.String()
			// 过滤掉IPv6链路本地地址和IPv4链路本地地址
			if strings.HasPrefix(ipStr, "fe80:") || strings.HasPrefix(ipStr, "169.254.") {
				continue
			}
			addresses = append(addresses, ipStr)
		}
	}

	return addresses, nil
}

// getInterfaceStatus 获取接口状态
// 根据接口标志位判断接口状态
func (c *NetworkCollector) getInterfaceStatus(flags []string) string {
	// 检查是否包含UP标志
	for _, flag := range flags {
		if strings.ToUpper(flag) == "UP" {
			return "up"
		}
	}
	return "down"
}

// getInterfaceType 获取接口类型
// 根据接口名称判断接口类型
func (c *NetworkCollector) getInterfaceType(interfaceName string) string {
	// 根据接口名称判断类型
	lowerName := strings.ToLower(interfaceName)

	if strings.HasPrefix(lowerName, "lo") {
		return "loopback"
	}
	if strings.HasPrefix(lowerName, "eth") || strings.HasPrefix(lowerName, "en") {
		return "ethernet"
	}
	if strings.HasPrefix(lowerName, "wlan") || strings.HasPrefix(lowerName, "wl") || strings.HasPrefix(lowerName, "wi-fi") {
		return "wifi"
	}
	if strings.HasPrefix(lowerName, "ppp") || strings.HasPrefix(lowerName, "tun") || strings.HasPrefix(lowerName, "tap") {
		return "vpn"
	}
	if strings.HasPrefix(lowerName, "docker") || strings.HasPrefix(lowerName, "br-") {
		return "bridge"
	}
	return "unknown"
}

// SetTimeout 设置采集超时时间
func (c *NetworkCollector) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout 获取采集超时时间
func (c *NetworkCollector) GetTimeout() time.Duration {
	return c.timeout
}

// GetInterfaceByName 根据名称获取指定网络接口信息
// 这是一个便捷方法，用于获取特定网络接口的信息
func (c *NetworkCollector) GetInterfaceByName(name string) (*types.NetworkInterface, error) {
	metrics, err := c.GetNetworkStats()
	if err != nil {
		return nil, err
	}

	for _, iface := range metrics.Interfaces {
		if iface.Name == name {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("未找到名称为 %s 的网络接口", name)
}

// GetActiveInterfaces 获取活动的网络接口
// 这是一个便捷方法，用于获取所有处于活动状态的网络接口
func (c *NetworkCollector) GetActiveInterfaces() ([]types.NetworkInterface, error) {
	metrics, err := c.GetNetworkStats()
	if err != nil {
		return nil, err
	}

	var activeInterfaces []types.NetworkInterface
	for _, iface := range metrics.Interfaces {
		if iface.Status == "up" {
			activeInterfaces = append(activeInterfaces, iface)
		}
	}

	return activeInterfaces, nil
}

// GetInterfaceIOStats 获取指定接口的IO统计信息
// 这是一个便捷方法，用于获取特定接口的流量统计
func (c *NetworkCollector) GetInterfaceIOStats(interfaceName string) (*types.NetworkInterface, error) {
	iface, err := c.GetInterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}

	return iface, nil
}

// GetTotalNetworkTraffic 获取总网络流量
// 这是一个便捷方法，用于获取所有接口的总流量
func (c *NetworkCollector) GetTotalNetworkTraffic() (bytesReceived, bytesSent uint64, err error) {
	metrics, err := c.GetNetworkStats()
	if err != nil {
		return 0, 0, err
	}

	for _, iface := range metrics.Interfaces {
		// 跳过回环接口
		if iface.Type == "loopback" {
			continue
		}
		bytesReceived += iface.BytesReceived
		bytesSent += iface.BytesSent
	}

	return bytesReceived, bytesSent, nil
}
