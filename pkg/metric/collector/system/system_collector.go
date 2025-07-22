package system

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"gohub/pkg/metric/collector"
	"gohub/pkg/metric/types"

	"github.com/shirou/gopsutil/v4/host"
	gopsutilNet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// 默认采集超时时间
const DefaultCollectTimeout = 5 * time.Second

// SystemCollector 系统信息采集器
// 使用 gopsutil 实现跨平台系统信息采集，包括主机信息、运行时间、进程统计等
type SystemCollector struct {
	*collector.BaseCollector
	// 采集超时时间，防止采集操作阻塞过久
	timeout time.Duration
}

// NewSystemCollector 创建系统信息采集器
// 使用 gopsutil 实现跨平台系统信息采集
//
// 返回值:
//   - *SystemCollector: 新创建的系统采集器实例
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{
		BaseCollector: collector.NewBaseCollector(
			types.CollectorNameSystem,
			"基于gopsutil的系统信息采集器，提供跨平台系统信息采集",
		),
		timeout: DefaultCollectTimeout,
	}
}

// Collect 执行采集操作
// 统一的采集入口，使用gopsutil简化实现
//
// 返回值:
//   - interface{}: 采集到的系统指标数据
//   - error: 采集过程中的错误
func (c *SystemCollector) Collect() (interface{}, error) {
	// 检查采集器是否已启用
	if !c.IsEnabled() {
		return nil, types.ErrCollectorDisabled
	}

	// 执行系统信息采集
	metrics, err := c.GetSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", types.ErrCollectFailed, err)
	}

	// 记录最后一次采集时间
	c.SetLastCollectTime(time.Now())
	return metrics, nil
}

// GetSystemInfo 获取系统信息
// 使用 gopsutil 采集系统的各种信息，包括基本信息、运行时间、统计信息等
//
// 返回值:
//   - *types.SystemMetrics: 系统指标数据
//   - error: 获取过程中的错误
func (c *SystemCollector) GetSystemInfo() (*types.SystemMetrics, error) {
	// 创建带超时的上下文，防止采集操作阻塞
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// 初始化指标数据结构
	metrics := &types.SystemMetrics{
		CollectTime: time.Now(),
	}

	// 获取基本主机信息
	if err := c.getHostInfo(ctx, metrics); err != nil {
		return nil, fmt.Errorf("获取主机信息失败: %w", err)
	}

	// 获取系统运行时间和启动时间
	if err := c.getBootInfo(ctx, metrics); err != nil {
		// 启动时间获取失败不影响其他信息
		metrics.Uptime = 0
		metrics.BootTime = time.Time{}
	}

	// 获取系统统计信息
	if err := c.getSystemStats(ctx, metrics); err != nil {
		// 统计信息获取失败不影响其他信息
		metrics.ProcessCount = 0
		metrics.UserCount = 0
	}

	// 获取温度信息（如果支持）
	c.getTemperatureInfo(ctx, metrics)

	// 获取网络信息
	if err := c.getNetworkInfo(ctx, metrics); err != nil {
		// 网络信息获取失败不影响其他信息
		metrics.NetworkInfo = &types.SystemNetworkInfo{}
	}

	// 检测服务器类型
	metrics.ServerType = c.getServerType(ctx)

	return metrics, nil
}

// getHostInfo 获取主机基本信息
// 使用 gopsutil 获取主机名、操作系统、架构、内核版本等信息
//
// 参数:
//   - ctx: 上下文，用于控制操作超时
//   - metrics: 系统指标数据结构
//
// 返回值:
//   - error: 获取过程中的错误
func (c *SystemCollector) getHostInfo(ctx context.Context, metrics *types.SystemMetrics) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 使用 gopsutil 获取主机信息
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return fmt.Errorf("gopsutil获取主机信息失败: %w", err)
	}

	// 填充主机基本信息
	metrics.Hostname = hostInfo.Hostname
	metrics.OS = hostInfo.OS
	metrics.Architecture = hostInfo.KernelArch
	metrics.KernelVersion = hostInfo.KernelVersion

	// 构建操作系统版本字符串
	if hostInfo.Platform != "" {
		if hostInfo.PlatformVersion != "" {
			metrics.OSVersion = fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
		} else {
			metrics.OSVersion = hostInfo.Platform
		}
	} else {
		metrics.OSVersion = "unknown"
	}

	return nil
}

// getBootInfo 获取系统启动信息
// 使用 gopsutil 获取系统启动时间和运行时间
//
// 参数:
//   - ctx: 上下文，用于控制操作超时
//   - metrics: 系统指标数据结构
//
// 返回值:
//   - error: 获取过程中的错误
func (c *SystemCollector) getBootInfo(ctx context.Context, metrics *types.SystemMetrics) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 使用 gopsutil 获取系统启动时间
	bootTime, err := host.BootTimeWithContext(ctx)
	if err != nil {
		return fmt.Errorf("gopsutil获取启动时间失败: %w", err)
	}

	// 设置启动时间（Unix时间戳转换为时间对象）
	metrics.BootTime = time.Unix(int64(bootTime), 0)

	// 计算系统运行时间（秒）
	metrics.Uptime = uint64(time.Since(metrics.BootTime).Seconds())

	return nil
}

// getSystemStats 获取系统统计信息
// 使用 gopsutil 获取进程数量和用户数量等统计信息
//
// 参数:
//   - ctx: 上下文，用于控制操作超时
//   - metrics: 系统指标数据结构
//
// 返回值:
//   - error: 获取过程中的错误
func (c *SystemCollector) getSystemStats(ctx context.Context, metrics *types.SystemMetrics) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 获取进程数量
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return fmt.Errorf("gopsutil获取进程列表失败: %w", err)
	}
	metrics.ProcessCount = uint32(len(processes))

	// 获取用户数量
	users, err := host.UsersWithContext(ctx)
	if err != nil {
		// 用户信息获取失败时设置默认值
		metrics.UserCount = 1
	} else {
		// 统计唯一用户数量
		userSet := make(map[string]struct{})
		for _, user := range users {
			if user.User != "" {
				userSet[user.User] = struct{}{}
			}
		}
		metrics.UserCount = uint32(len(userSet))
		if metrics.UserCount == 0 {
			metrics.UserCount = 1 // 至少有一个用户
		}
	}

	return nil
}

// getTemperatureInfo 获取系统温度信息
// 注意: 当前版本暂不支持温度信息获取，设置为空数组
//
// 参数:
//   - ctx: 上下文，用于控制操作超时
//   - metrics: 系统指标数据结构
//
// 注意: 温度信息获取在某些系统上可能不支持，失败时不影响其他信息
func (c *SystemCollector) getTemperatureInfo(ctx context.Context, metrics *types.SystemMetrics) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		metrics.Temperature = []types.TemperatureInfo{}
		return
	default:
	}

	// 温度信息获取暂不支持，设置为空数组
	// 未来可以考虑使用其他方式实现温度监控
	metrics.Temperature = []types.TemperatureInfo{}
}

// SetTimeout 设置采集超时时间
// 用于控制系统信息采集操作的最大执行时间
//
// 参数:
//   - timeout: 超时时间
func (c *SystemCollector) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout 获取采集超时时间
// 返回当前设置的采集超时时间
//
// 返回值:
//   - time.Duration: 当前的超时时间设置
func (c *SystemCollector) GetTimeout() time.Duration {
	return c.timeout
}

// GetHostname 获取主机名
// 这是一个便捷方法，直接返回系统主机名
//
// 返回值:
//   - string: 主机名
//   - error: 获取过程中的错误
func (c *SystemCollector) GetHostname() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("gopsutil获取主机名失败: %w", err)
	}
	return hostInfo.Hostname, nil
}

// GetOSInfo 获取操作系统基本信息
// 这是一个便捷方法，返回操作系统类型、架构和内核版本信息
//
// 返回值:
//   - string: 操作系统类型
//   - string: 系统架构
//   - string: 内核版本
//   - error: 获取过程中的错误
func (c *SystemCollector) GetOSInfo() (string, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("gopsutil获取操作系统信息失败: %w", err)
	}

	return hostInfo.OS, hostInfo.KernelArch, hostInfo.KernelVersion, nil
}

// GetUptime 获取系统运行时间
// 这是一个便捷方法，只返回运行时间（秒）
//
// 返回值:
//   - uint64: 系统运行时间（秒）
//   - error: 获取过程中的错误
func (c *SystemCollector) GetUptime() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	bootTime, err := host.BootTimeWithContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("gopsutil获取启动时间失败: %w", err)
	}

	// 计算运行时间
	uptime := uint64(time.Since(time.Unix(int64(bootTime), 0)).Seconds())
	return uptime, nil
}

// GetBootTime 获取系统启动时间
// 这是一个便捷方法，返回系统启动时间
//
// 返回值:
//   - time.Time: 系统启动时间
//   - error: 获取过程中的错误
func (c *SystemCollector) GetBootTime() (time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	bootTime, err := host.BootTimeWithContext(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("gopsutil获取启动时间失败: %w", err)
	}

	return time.Unix(int64(bootTime), 0), nil
}

// GetProcessCount 获取系统进程总数
// 这是一个便捷方法，返回当前系统中的进程总数
//
// 返回值:
//   - uint32: 进程总数
//   - error: 获取过程中的错误
func (c *SystemCollector) GetProcessCount() (uint32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return 0, fmt.Errorf("gopsutil获取进程列表失败: %w", err)
	}

	return uint32(len(processes)), nil
}

// GetUserCount 获取系统用户数
// 这是一个便捷方法，返回当前登录的用户数量
//
// 返回值:
//   - uint32: 用户数量
//   - error: 获取过程中的错误
func (c *SystemCollector) GetUserCount() (uint32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	users, err := host.UsersWithContext(ctx)
	if err != nil {
		return 1, nil // 默认至少有一个用户
	}

	// 统计唯一用户数量
	userSet := make(map[string]struct{})
	for _, user := range users {
		if user.User != "" {
			userSet[user.User] = struct{}{}
		}
	}

	userCount := uint32(len(userSet))
	if userCount == 0 {
		userCount = 1 // 至少有一个用户
	}

	return userCount, nil
}

// GetTemperatures 获取系统温度信息
// 这是一个便捷方法，返回系统温度传感器信息
// 注意: 当前版本暂不支持温度信息获取
//
// 返回值:
//   - []types.TemperatureInfo: 温度信息列表（当前总是空数组）
//   - error: 获取过程中的错误
func (c *SystemCollector) GetTemperatures() ([]types.TemperatureInfo, error) {
	// 温度信息获取暂不支持，返回空数组
	// 未来可以考虑使用其他方式实现温度监控
	return []types.TemperatureInfo{}, nil
}

// getNetworkInfo 获取网络信息
// 使用 gopsutil 获取网络接口信息，包括IP地址、MAC地址等
//
// 参数:
//   - ctx: 上下文，用于控制操作超时
//   - metrics: 系统指标数据结构
//
// 返回值:
//   - error: 获取过程中的错误
func (c *SystemCollector) getNetworkInfo(ctx context.Context, metrics *types.SystemMetrics) error {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 初始化网络信息结构
	networkInfo := &types.SystemNetworkInfo{
		IPAddresses:      []string{},
		MACAddresses:     []string{},
		ActiveInterfaces: []string{},
	}

	// 获取网络接口信息
	interfaces, err := gopsutilNet.InterfacesWithContext(ctx)
	if err != nil {
		return fmt.Errorf("gopsutil获取网络接口失败: %w", err)
	}

	var primaryIP, primaryMAC, primaryInterface string
	var hasEthernet bool
	var hasIPv4 bool // 新增：标记是否已找到IPv4地址

	for _, iface := range interfaces {
		// 跳过回环接口
		if strings.Contains(strings.ToLower(iface.Name), "lo") {
			continue
		}

		// 检查接口是否启用
		isUp := false
		for _, flag := range iface.Flags {
			if strings.ToUpper(flag) == "UP" {
				isUp = true
				break
			}
		}

		if !isUp {
			continue
		}

		// 获取接口地址
		addresses, err := c.getInterfaceAddresses(iface.Name)
		if err != nil {
			continue
		}

		// 添加到活动接口列表
		if len(addresses) > 0 {
			networkInfo.ActiveInterfaces = append(networkInfo.ActiveInterfaces, iface.Name)
		}

		// 处理IP地址
		for _, addr := range addresses {
			// 跳过IPv6链路本地地址和IPv4链路本地地址
			if strings.HasPrefix(addr, "fe80:") || strings.HasPrefix(addr, "169.254.") {
				continue
			}
			networkInfo.IPAddresses = append(networkInfo.IPAddresses, addr)

			// 判断是否为IPv4地址
			isIPv4 := c.isIPv4Address(addr)
			// 判断是否为常见的私网IPv4地址
			isCommonPrivateIPv4 := isIPv4 && c.isCommonPrivateIPv4Address(addr)
			
			// 选择主IP地址的优先级：
			// 1. 优先选择常见的私网IPv4地址（192.168.x.x）
			// 2. 优先选择其他IPv4地址
			// 3. 优先选择以太网接口
			// 4. 其次是第一个非回环地址
			shouldSelectAsPrimary := false
			
			// 获取当前主IP是否为常见私网IPv4地址
			currentIsCommonPrivateIPv4 := primaryIP != "" && hasIPv4 && c.isCommonPrivateIPv4Address(primaryIP)
			
			if primaryIP == "" {
				// 如果还没有主IP，直接选择
				shouldSelectAsPrimary = true
			} else if !currentIsCommonPrivateIPv4 && isCommonPrivateIPv4 {
				// 如果当前主IP不是常见私网IPv4，但这个是常见私网IPv4，则优先选择
				shouldSelectAsPrimary = true
			} else if currentIsCommonPrivateIPv4 == isCommonPrivateIPv4 {
				// 如果常见私网IPv4状态相同，则比较IPv4状态
				if !hasIPv4 && isIPv4 {
					// 如果当前主IP不是IPv4，但这个是IPv4，则优先选择IPv4
					shouldSelectAsPrimary = true
				} else if hasIPv4 == isIPv4 {
					// 如果IP版本相同，则比较接口类型
					if !hasEthernet && c.isEthernetInterface(iface.Name) {
						shouldSelectAsPrimary = true
					}
				}
			}
			
			if shouldSelectAsPrimary {
				primaryIP = addr
				primaryInterface = iface.Name
				hasIPv4 = isIPv4
				if c.isEthernetInterface(iface.Name) {
					hasEthernet = true
				}
			}
		}

		// 处理MAC地址
		if iface.HardwareAddr != "" {
			networkInfo.MACAddresses = append(networkInfo.MACAddresses, iface.HardwareAddr)
			
			// 选择主MAC地址：优先以太网接口的MAC
			if primaryMAC == "" || (!hasEthernet && c.isEthernetInterface(iface.Name)) {
				primaryMAC = iface.HardwareAddr
				if c.isEthernetInterface(iface.Name) {
					hasEthernet = true
				}
			}
		}
	}

	// 设置主要网络信息
	networkInfo.PrimaryIP = primaryIP
	networkInfo.PrimaryMAC = primaryMAC
	networkInfo.PrimaryInterface = primaryInterface

	metrics.NetworkInfo = networkInfo
	return nil
}

// getInterfaceAddresses 获取指定接口的IP地址列表
// 使用标准库获取网络接口的地址信息
//
// 参数:
//   - interfaceName: 接口名称
//
// 返回值:
//   - []string: IP地址列表
//   - error: 获取过程中的错误
func (c *SystemCollector) getInterfaceAddresses(interfaceName string) ([]string, error) {
	var addresses []string

	// 使用标准库获取接口地址
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
			addresses = append(addresses, ipNet.IP.String())
		}
	}

	return addresses, nil
}

// isEthernetInterface 判断是否为以太网接口
// 根据接口名称判断接口类型
//
// 参数:
//   - interfaceName: 接口名称
//
// 返回值:
//   - bool: 是否为以太网接口
func (c *SystemCollector) isEthernetInterface(interfaceName string) bool {
	lowerName := strings.ToLower(interfaceName)
	return strings.HasPrefix(lowerName, "eth") || 
		   strings.HasPrefix(lowerName, "en") || 
		   strings.HasPrefix(lowerName, "em")
}

// isIPv4Address 判断是否为IPv4地址
// 根据地址格式判断IP版本
//
// 参数:
//   - addr: IP地址字符串
//
// 返回值:
//   - bool: 是否为IPv4地址
func (c *SystemCollector) isIPv4Address(addr string) bool {
	// 简单的IPv4地址判断：包含点号且不包含冒号
	return strings.Contains(addr, ".") && !strings.Contains(addr, ":")
}

// isCommonPrivateIPv4Address 判断是否为常见的私网IPv4地址
// 常见的私网IPv4地址包括：192.168.x.x, 10.x.x.x, 172.16.x.x, 172.17.x.x, 172.18.x.x, 172.19.x.x, 172.20.x.x, 172.21.x.x, 172.22.x.x, 172.23.x.x, 172.24.x.x, 172.25.x.x, 172.26.x.x, 172.27.x.x, 172.28.x.x, 172.29.x.x, 172.30.x.x, 172.31.x.x
// 注意：这里只检查前缀，不检查完整的IP地址
func (c *SystemCollector) isCommonPrivateIPv4Address(addr string) bool {
	lowerAddr := strings.ToLower(addr)

	// 192.168.x.x
	if strings.HasPrefix(lowerAddr, "192.168.") {
		return true
	}

	// 10.x.x.x
	if strings.HasPrefix(lowerAddr, "10.") {
		return true
	}

	// 172.16.x.x - 172.31.x.x
	if strings.HasPrefix(lowerAddr, "172.16.") || strings.HasPrefix(lowerAddr, "172.17.") ||
	   strings.HasPrefix(lowerAddr, "172.18.") || strings.HasPrefix(lowerAddr, "172.19.") ||
	   strings.HasPrefix(lowerAddr, "172.20.") || strings.HasPrefix(lowerAddr, "172.21.") ||
	   strings.HasPrefix(lowerAddr, "172.22.") || strings.HasPrefix(lowerAddr, "172.23.") ||
	   strings.HasPrefix(lowerAddr, "172.24.") || strings.HasPrefix(lowerAddr, "172.25.") ||
	   strings.HasPrefix(lowerAddr, "172.26.") || strings.HasPrefix(lowerAddr, "172.27.") ||
	   strings.HasPrefix(lowerAddr, "172.28.") || strings.HasPrefix(lowerAddr, "172.29.") ||
	   strings.HasPrefix(lowerAddr, "172.30.") || strings.HasPrefix(lowerAddr, "172.31.") {
		return true
	}

	return false
}

// getServerType 检测服务器类型
// 通过多种方式检测服务器是物理机还是虚拟机
//
// 参数:
//   - ctx: 上下文，用于控制操作超时
//
// 返回值:
//   - string: 服务器类型（physical/virtual/unknown）
func (c *SystemCollector) getServerType(ctx context.Context) string {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return "unknown"
	default:
	}

	// 方法1: 检查虚拟化相关的系统信息
	if virtualization := c.detectVirtualization(ctx); virtualization != "" {
		return "virtual"
	}

	// 方法2: 检查DMI信息（仅限Linux）
	if runtime.GOOS == "linux" {
		if c.checkDMIInfo() {
			return "virtual"
		}
	}

	// 方法3: 检查CPU信息中的虚拟化标识
	if c.checkCPUVirtualization(ctx) {
		return "virtual"
	}

	// 默认假设为物理机
	return "physical"
}

// detectVirtualization 检测虚拟化环境
// 使用 gopsutil 检测虚拟化信息
//
// 参数:
//   - ctx: 上下文，用于控制操作超时
//
// 返回值:
//   - string: 虚拟化类型，空字符串表示未检测到
func (c *SystemCollector) detectVirtualization(ctx context.Context) string {
	// 使用 gopsutil 获取虚拟化信息
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return ""
	}

	// 检查虚拟化字段
	if hostInfo.VirtualizationSystem != "" {
		return hostInfo.VirtualizationSystem
	}

	// 检查主机ID是否表明虚拟化
	if hostInfo.HostID != "" {
		// 某些虚拟化平台会有特定的主机ID模式
		if c.isVirtualHostID(hostInfo.HostID) {
			return "virtual"
		}
	}

	return ""
}

// checkDMIInfo 检查DMI信息判断虚拟化
// 在Linux系统上检查DMI信息中的虚拟化标识
//
// 返回值:
//   - bool: 是否检测到虚拟化环境
func (c *SystemCollector) checkDMIInfo() bool {
	// 这里可以检查 /sys/class/dmi/id/ 下的文件
	// 或者使用其他方式检查DMI信息
	// 为了简化，这里返回false，实际实现时可以添加更详细的检查
	return false
}

// checkCPUVirtualization 检查CPU信息中的虚拟化标识
// 检查CPU信息中是否包含虚拟化相关的标识
//
// 参数:
//   - ctx: 上下文，用于控制操作超时
//
// 返回值:
//   - bool: 是否检测到虚拟化环境
func (c *SystemCollector) checkCPUVirtualization(ctx context.Context) bool {
	// 这里可以检查CPU信息中的虚拟化标识
	// 例如检查CPU型号名称中是否包含虚拟化相关的关键字
	// 为了简化，这里返回false，实际实现时可以添加更详细的检查
	return false
}

// isVirtualHostID 检查主机ID是否表明虚拟化环境
// 某些虚拟化平台会有特定的主机ID模式
//
// 参数:
//   - hostID: 主机ID
//
// 返回值:
//   - bool: 是否为虚拟化环境的主机ID
func (c *SystemCollector) isVirtualHostID(hostID string) bool {
	// 检查是否为VMware的主机ID模式
	if strings.HasPrefix(hostID, "VMware") {
		return true
	}
	
	// 检查是否为VirtualBox的主机ID模式
	if strings.Contains(hostID, "VirtualBox") {
		return true
	}
	
	// 可以添加更多虚拟化平台的检查
	return false
}

// GetNetworkInfo 获取网络信息
// 这是一个便捷方法，直接返回系统网络信息
//
// 返回值:
//   - *types.SystemNetworkInfo: 网络信息
//   - error: 获取过程中的错误
func (c *SystemCollector) GetNetworkInfo() (*types.SystemNetworkInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	metrics := &types.SystemMetrics{}
	err := c.getNetworkInfo(ctx, metrics)
	if err != nil {
		return nil, err
	}
	
	return metrics.NetworkInfo, nil
}

// GetServerType 获取服务器类型
// 这是一个便捷方法，直接返回服务器类型
//
// 返回值:
//   - string: 服务器类型（physical/virtual/unknown）
func (c *SystemCollector) GetServerType() string {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.getServerType(ctx)
} 