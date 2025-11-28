package init

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/metric_collect/dao"
	"gateway/internal/metric_collect/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/metric"
	metricTypes "gateway/pkg/metric/types"
)

// ServerInfoManager 服务器信息管理器
// 负责服务器信息的初始化、更新和管理
//
// 主要功能：
// - 服务器信息初始化：从系统采集服务器基本信息
// - 服务器信息更新：定期更新服务器状态信息
// - 数据库操作：处理服务器信息的数据库存储和更新
// - 网络信息管理：管理服务器的网络配置信息
// - 硬件信息管理：管理服务器的硬件配置信息
type ServerInfoManager struct {
	// 配置信息
	config *MetricConfig // 指标采集配置

	// 数据库连接
	db database.Database // 数据库连接实例

	// DAO实例
	serverInfoDAO *dao.ServerInfoDAO // 服务器信息DAO

	// 服务器信息缓存
	serverInfo *types.ServerInfo // 当前服务器信息
}

// NewServerInfoManager 创建服务器信息管理器
// 创建一个新的服务器信息管理器实例
//
// 参数：
//   - config: 指标采集配置
//   - db: 数据库连接实例
//
// 返回：
//   - *ServerInfoManager: 服务器信息管理器实例
func NewServerInfoManager(config *MetricConfig, db database.Database) *ServerInfoManager {
	return &ServerInfoManager{
		config:        config,
		db:            db,
		serverInfoDAO: dao.NewServerInfoDAO(db),
	}
}

// InitializeServerInfo 初始化服务器信息
// 从数据库获取或创建服务器信息记录
//
// 流程：
// 1. 采集系统信息获取完整的系统信息（包括网络信息和服务器类型）
// 2. 尝试从数据库按hostname、IP和MAC地址获取已存在的服务器信息（确保唯一性）
// 3. 如果不存在，创建新记录；如果存在，更新信息（使用已存在的ServerId）
// 4. 将服务器信息缓存到内存中
//
// 返回：
//   - error: 初始化过程中的错误
func (m *ServerInfoManager) InitializeServerInfo() error {
	ctx := context.Background()

	// 使用底层完善的系统采集器获取完整的系统信息
	systemMetrics, err := metric.CollectSystem()
	if err != nil {
		return fmt.Errorf("采集系统信息失败: %w", err)
	}

	// 采集其他指标用于硬件信息
	cpuMetrics, _ := metric.CollectCPU()
	memoryMetrics, _ := metric.CollectMemory()
	diskMetrics, _ := metric.CollectDisk()

	// 获取网络信息
	primaryIP := m.getPrimaryIP(systemMetrics)
	primaryMAC := m.getPrimaryMAC(systemMetrics)

	var existingInfo *types.ServerInfo

	// 首先根据唯一约束字段（tenantId, metricServerId）查询，确保不会违反唯一约束
	existingInfo, err = m.serverInfoDAO.GetServerInfoById(ctx, m.config.TenantId, m.config.ServerId)
	if err != nil {
		logger.Error("根据ServerId查询服务器信息失败", "error", err, "server_id", m.config.ServerId)
	}

	// 如果通过ServerId没找到，再尝试通过hostname/IP/MAC查询（可能ServerId变化了）
	if existingInfo == nil {
		existingInfo, err = m.serverInfoDAO.GetServerInfoByHostnameAndNetwork(
			ctx,
			m.config.TenantId,
			systemMetrics.Hostname,
			primaryIP,
			primaryMAC,
		)
		if err != nil {
			logger.Error("根据hostname/IP/MAC查询服务器信息失败", "error", err)
		}
	}

	now := time.Now()

	if existingInfo != nil {
		// 更新已存在的服务器信息
		m.serverInfo = existingInfo

		// 重要：使用数据库中已存在的 MetricServerId，避免ID不一致
		// 同时更新配置中的 ServerId 以保持一致性
		if existingInfo.MetricServerId != "" && existingInfo.MetricServerId != m.config.ServerId {
			logger.Warn("检测到ServerId不一致，使用数据库中的ServerId",
				"config_server_id", m.config.ServerId,
				"db_server_id", existingInfo.MetricServerId,
				"hostname", systemMetrics.Hostname)
			m.config.ServerId = existingInfo.MetricServerId
		}

		if err := m.updateExistingServerInfo(systemMetrics, cpuMetrics, memoryMetrics, diskMetrics, now); err != nil {
			return fmt.Errorf("更新服务器信息失败: %w", err)
		}

		logger.Info("更新已存在的服务器信息",
			"server_id", m.serverInfo.MetricServerId,
			"hostname", systemMetrics.Hostname,
			"ip", primaryIP,
			"mac", primaryMAC,
			"server_type", systemMetrics.ServerType)
	} else {
		// 创建新的服务器信息记录
		if err := m.createNewServerInfo(systemMetrics, cpuMetrics, memoryMetrics, diskMetrics, now); err != nil {
			return fmt.Errorf("创建服务器信息失败: %w", err)
		}

		logger.Info("服务器信息初始化成功",
			"server_id", m.config.ServerId,
			"hostname", systemMetrics.Hostname,
			"os", systemMetrics.OS,
			"ip", primaryIP,
			"mac", primaryMAC,
			"server_type", systemMetrics.ServerType)
	}

	return nil
}

// updateExistingServerInfo 更新已存在的服务器信息
// 更新数据库中已存在的服务器信息记录
//
// 参数：
//   - systemMetrics: 系统指标信息
//   - cpuMetrics: CPU指标信息
//   - memoryMetrics: 内存指标信息
//   - diskMetrics: 磁盘指标信息
//   - now: 当前时间
//
// 返回：
//   - error: 更新过程中的错误
func (m *ServerInfoManager) updateExistingServerInfo(
	systemMetrics *metricTypes.SystemMetrics,
	cpuMetrics *metricTypes.CPUMetrics,
	memoryMetrics *metricTypes.MemoryMetrics,
	diskMetrics *metricTypes.DiskMetrics,
	now time.Time,
) error {
	ctx := context.Background()

	// 更新基本信息
	m.serverInfo.OsType = systemMetrics.OS
	m.serverInfo.OsVersion = systemMetrics.OSVersion
	m.serverInfo.KernelVersion = &systemMetrics.KernelVersion
	m.serverInfo.Architecture = systemMetrics.Architecture
	m.serverInfo.BootTime = systemMetrics.BootTime
	m.serverInfo.LastUpdateTime = now
	m.serverInfo.EditTime = now
	m.serverInfo.EditWho = m.config.Operator
	m.serverInfo.OprSeqFlag = fmt.Sprintf("UPDATE_%d", now.Unix())

	// 更新网络信息
	m.updateNetworkInfo(systemMetrics)

	// 更新服务器类型
	if systemMetrics.ServerType != "" {
		m.serverInfo.ServerType = &systemMetrics.ServerType
	}

	// 更新系统信息
	m.updateSystemInfo(systemMetrics)

	// 更新硬件信息
	m.updateHardwareInfo(cpuMetrics, memoryMetrics, diskMetrics)

	// 注意：不再更新 MetricServerId，保持使用数据库中已存在的ID
	// 这样可以确保服务器ID的一致性，避免同一台服务器有多个不同的ID

	// 更新数据库
	if err := m.serverInfoDAO.UpdateServerInfo(ctx, m.serverInfo); err != nil {
		return fmt.Errorf("更新服务器信息失败: %w", err)
	}

	return nil
}

// createNewServerInfo 创建新的服务器信息
// 创建新的服务器信息记录并插入到数据库
//
// 参数：
//   - systemMetrics: 系统指标信息
//   - cpuMetrics: CPU指标信息
//   - memoryMetrics: 内存指标信息
//   - diskMetrics: 磁盘指标信息
//   - now: 当前时间
//
// 返回：
//   - error: 创建过程中的错误
func (m *ServerInfoManager) createNewServerInfo(
	systemMetrics *metricTypes.SystemMetrics,
	cpuMetrics *metricTypes.CPUMetrics,
	memoryMetrics *metricTypes.MemoryMetrics,
	diskMetrics *metricTypes.DiskMetrics,
	now time.Time,
) error {
	ctx := context.Background()

	// 创建新的服务器信息记录
	m.serverInfo = &types.ServerInfo{
		MetricServerId: m.config.ServerId,
		TenantId:       m.config.TenantId,
		Hostname:       systemMetrics.Hostname,
		OsType:         systemMetrics.OS,
		OsVersion:      systemMetrics.OSVersion,
		KernelVersion:  &systemMetrics.KernelVersion,
		Architecture:   systemMetrics.Architecture,
		BootTime:       systemMetrics.BootTime,
		LastUpdateTime: now,
		AddTime:        now,
		AddWho:         m.config.Operator,
		EditTime:       now,
		EditWho:        m.config.Operator,
		OprSeqFlag:     fmt.Sprintf("INIT_%d", now.Unix()),
		CurrentVersion: 1,
		ActiveFlag:     types.ActiveFlagYes,
	}

	// 设置网络信息
	m.updateNetworkInfo(systemMetrics)

	// 设置服务器类型
	if systemMetrics.ServerType != "" {
		m.serverInfo.ServerType = &systemMetrics.ServerType
	}

	// 设置系统信息
	m.updateSystemInfo(systemMetrics)

	// 设置硬件信息
	m.updateHardwareInfo(cpuMetrics, memoryMetrics, diskMetrics)

	// 插入服务器信息到数据库
	if err := m.serverInfoDAO.InsertServerInfo(ctx, m.serverInfo); err != nil {
		return fmt.Errorf("插入服务器信息失败: %w", err)
	}

	return nil
}

// updateNetworkInfo 更新网络信息
// 更新服务器的网络配置信息
//
// 参数：
//   - systemMetrics: 系统指标信息
func (m *ServerInfoManager) updateNetworkInfo(systemMetrics *metricTypes.SystemMetrics) {
	if systemMetrics.NetworkInfo == nil {
		return
	}

	// 更新基本网络信息
	if systemMetrics.NetworkInfo.PrimaryIP != "" {
		m.serverInfo.IpAddress = &systemMetrics.NetworkInfo.PrimaryIP
	}
	if systemMetrics.NetworkInfo.PrimaryMAC != "" {
		m.serverInfo.MacAddress = &systemMetrics.NetworkInfo.PrimaryMAC
	}

	// 设置完整网络信息
	networkData := &types.NetworkInfoData{
		PrimaryIP:        systemMetrics.NetworkInfo.PrimaryIP,
		PrimaryMAC:       systemMetrics.NetworkInfo.PrimaryMAC,
		PrimaryInterface: systemMetrics.NetworkInfo.PrimaryInterface,
		AllIPs:           systemMetrics.NetworkInfo.IPAddresses,
		AllMACs:          systemMetrics.NetworkInfo.MACAddresses,
		ActiveInterfaces: systemMetrics.NetworkInfo.ActiveInterfaces,
	}
	m.serverInfo.SetNetworkInfo(networkData)
}

// updateSystemInfo 更新系统信息
// 更新服务器的系统运行信息
//
// 参数：
//   - systemMetrics: 系统指标信息
func (m *ServerInfoManager) updateSystemInfo(systemMetrics *metricTypes.SystemMetrics) {
	systemData := &types.SystemInfoData{
		Uptime:       systemMetrics.Uptime,
		UserCount:    systemMetrics.UserCount,
		ProcessCount: systemMetrics.ProcessCount,
		LoadAvg:      make(map[string]float64),
		Temperatures: []types.TemperatureData{},
	}

	// 添加温度信息
	for _, temp := range systemMetrics.Temperature {
		systemData.Temperatures = append(systemData.Temperatures, types.TemperatureData{
			Sensor:   temp.SensorName,
			Value:    temp.Temperature,
			High:     temp.High,
			Critical: temp.Critical,
		})
	}

	m.serverInfo.SetSystemInfo(systemData)
}

// updateHardwareInfo 更新硬件信息
// 更新服务器的硬件配置信息
//
// 参数：
//   - cpuMetrics: CPU指标信息
//   - memoryMetrics: 内存指标信息
//   - diskMetrics: 磁盘指标信息
func (m *ServerInfoManager) updateHardwareInfo(
	cpuMetrics *metricTypes.CPUMetrics,
	memoryMetrics *metricTypes.MemoryMetrics,
	diskMetrics *metricTypes.DiskMetrics,
) {
	hardwareData := &types.HardwareInfoData{}

	// 设置CPU信息
	if cpuMetrics != nil {
		hardwareData.CPU = types.CPUHardwareInfo{
			CoreCount:    cpuMetrics.CoreCount,
			LogicalCount: cpuMetrics.LogicalCount,
			Model:        "unknown", // 需要从其他地方获取
			Frequency:    "unknown", // 需要从其他地方获取
		}
	}

	// 设置内存信息
	if memoryMetrics != nil {
		hardwareData.Memory = types.MemoryHardwareInfo{
			Total: memoryMetrics.Total,
			Type:  "unknown", // 需要从其他地方获取
			Speed: "unknown", // 需要从其他地方获取
		}
	}

	// 设置存储信息
	if diskMetrics != nil {
		var totalCapacity uint64
		for _, partition := range diskMetrics.Partitions {
			totalCapacity += partition.Total
		}
		hardwareData.Storage = types.StorageHardwareInfo{
			TotalDisks:    len(diskMetrics.Partitions),
			TotalCapacity: totalCapacity,
		}
	}

	m.serverInfo.SetHardwareInfo(hardwareData)
}

// GetServerInfo 获取服务器信息
// 返回当前缓存的服务器信息
//
// 返回：
//   - *types.ServerInfo: 服务器信息，如果未初始化则返回nil
func (m *ServerInfoManager) GetServerInfo() *types.ServerInfo {
	return m.serverInfo
}

// RefreshServerInfo 刷新服务器信息
// 重新采集系统信息并更新服务器信息
//
// 返回：
//   - error: 刷新过程中的错误
func (m *ServerInfoManager) RefreshServerInfo() error {
	if m.serverInfo == nil {
		return fmt.Errorf("服务器信息未初始化")
	}

	// 重新采集系统信息
	systemMetrics, err := metric.CollectSystem()
	if err != nil {
		return fmt.Errorf("采集系统信息失败: %w", err)
	}

	// 采集其他指标
	cpuMetrics, _ := metric.CollectCPU()
	memoryMetrics, _ := metric.CollectMemory()
	diskMetrics, _ := metric.CollectDisk()

	// 更新服务器信息
	now := time.Now()
	if err := m.updateExistingServerInfo(systemMetrics, cpuMetrics, memoryMetrics, diskMetrics, now); err != nil {
		return fmt.Errorf("刷新服务器信息失败: %w", err)
	}

	logger.Debug("服务器信息刷新成功",
		"server_id", m.config.ServerId,
		"hostname", systemMetrics.Hostname)

	return nil
}

// getPrimaryIP 获取主要IP地址
// 从系统指标中获取主要IP地址
//
// 参数：
//   - systemMetrics: 系统指标信息
//
// 返回：
//   - string: 主要IP地址
func (m *ServerInfoManager) getPrimaryIP(systemMetrics *metricTypes.SystemMetrics) string {
	if systemMetrics.NetworkInfo != nil {
		return systemMetrics.NetworkInfo.PrimaryIP
	}
	return ""
}

// getPrimaryMAC 获取主要MAC地址
// 从系统指标中获取主要MAC地址
//
// 参数：
//   - systemMetrics: 系统指标信息
//
// 返回：
//   - string: 主要MAC地址
func (m *ServerInfoManager) getPrimaryMAC(systemMetrics *metricTypes.SystemMetrics) string {
	if systemMetrics.NetworkInfo != nil {
		return systemMetrics.NetworkInfo.PrimaryMAC
	}
	return ""
}

// UpdateServerStatus 更新服务器状态
// 更新服务器的运行状态信息
//
// 参数：
//   - status: 服务器状态
//   - notes: 状态说明
//
// 返回：
//   - error: 更新过程中的错误
func (m *ServerInfoManager) UpdateServerStatus(status string, notes string) error {
	if m.serverInfo == nil {
		return fmt.Errorf("服务器信息未初始化")
	}

	ctx := context.Background()
	now := time.Now()

	// 更新状态信息
	m.serverInfo.LastUpdateTime = now
	m.serverInfo.EditTime = now
	m.serverInfo.EditWho = m.config.Operator
	m.serverInfo.OprSeqFlag = fmt.Sprintf("STATUS_%d", now.Unix())

	if notes != "" {
		m.serverInfo.NoteText = &notes
	}

	// 更新数据库
	if err := m.serverInfoDAO.UpdateServerInfo(ctx, m.serverInfo); err != nil {
		return fmt.Errorf("更新服务器状态失败: %w", err)
	}

	logger.Info("服务器状态更新成功",
		"server_id", m.config.ServerId,
		"status", status,
		"notes", notes)

	return nil
}
