package init

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gohub/internal/metric_collect/dao"
	"gohub/internal/metric_collect/types"
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/pkg/metric"
	metricTypes "gohub/pkg/metric/types"
)

// MetricCollectorManager 指标采集管理器
// 负责管理系统指标的采集、存储和清理工作
// 支持CPU、内存、磁盘、网络、进程等多种指标类型的采集
//
// 主要功能：
// - 指标数据采集：定时采集系统各项指标数据
// - 数据缓存：批量缓存采集数据，提高数据库写入效率
// - 数据存储：将采集的指标数据持久化到数据库
// - 数据清理：定期清理过期的历史数据
// - 生命周期管理：支持启动、停止、状态查询等操作
type MetricCollectorManager struct {
	// 配置信息
	config *MetricConfig // 指标采集配置
	
	// 数据库连接
	db database.Database // 数据库连接实例
	
	// DAO实例 - 数据访问对象，负责各类指标数据的数据库操作
	serverInfoDAO         *dao.ServerInfoDAO         // 服务器信息DAO
	cpuLogDAO            *dao.CpuLogDAO             // CPU日志DAO
	memoryLogDAO         *dao.MemoryLogDAO          // 内存日志DAO
	diskPartitionLogDAO  *dao.DiskPartitionLogDAO   // 磁盘分区日志DAO
	diskIoLogDAO         *dao.DiskIoLogDAO          // 磁盘IO日志DAO
	networkLogDAO        *dao.NetworkLogDAO         // 网络日志DAO
	processLogDAO        *dao.ProcessLogDAO         // 进程日志DAO
	processStatsLogDAO   *dao.ProcessStatsLogDAO    // 进程统计日志DAO
	temperatureLogDAO    *dao.TemperatureLogDAO     // 温度日志DAO
	
	// 运行状态控制
	running     bool              // 是否正在运行
	stopChan    chan struct{}     // 停止信号通道
	mu          sync.RWMutex      // 读写锁，保护运行状态
	
	// 数据缓存
	dataBuffer  *DataBuffer       // 数据缓冲区
	
	// 服务器信息管理器
	serverInfoManager *ServerInfoManager // 服务器信息管理器
	
	// 配置验证器
	configValidator *ConfigValidator // 配置验证器实例
}

// DataBuffer 数据缓冲区
// 用于批量缓存采集到的指标数据，提高数据库写入效率
// 
// 设计原理：
// - 减少数据库连接开销：批量插入比单条插入效率更高
// - 提高系统吞吐量：异步写入，不阻塞数据采集
// - 内存管理：定期刷新缓冲区，避免内存占用过高
type DataBuffer struct {
	cpuLogs            []*types.CpuLog            // CPU指标日志缓冲
	memoryLogs         []*types.MemoryLog         // 内存指标日志缓冲
	diskPartitionLogs  []*types.DiskPartitionLog  // 磁盘分区日志缓冲
	diskIoLogs         []*types.DiskIoLog         // 磁盘IO日志缓冲
	networkLogs        []*types.NetworkLog        // 网络指标日志缓冲
	processLogs        []*types.ProcessLog        // 进程信息日志缓冲
	processStatsLogs   []*types.ProcessStatsLog   // 进程统计日志缓冲
	temperatureLogs    []*types.TemperatureLog    // 温度指标日志缓冲
	mu                 sync.Mutex                 // 互斥锁，保护缓冲区数据
}

// 全局管理器实例
var (
	globalManager *MetricCollectorManager // 全局指标采集管理器实例
	initOnce      sync.Once                // 确保只初始化一次
)

// InitMetricCollector 初始化指标采集器
// 这是包的入口函数，负责创建和初始化指标采集管理器
//
// 参数：
//   - db: 数据库连接实例
//
// 返回：
//   - error: 初始化过程中的错误
//
// 注意：
//   - 使用sync.Once确保只初始化一次
//   - 支持自动启动功能
func InitMetricCollector(db database.Database) error {
	var err error
	initOnce.Do(func() {
		globalManager, err = NewMetricCollectorManager(db)
		if err != nil {
			logger.Error("初始化指标采集管理器失败", "error", err)
			return
		}
		
		// 如果配置了自动启动，则启动采集器
		if globalManager.config.AutoStart {
			if startErr := globalManager.Start(); startErr != nil {
				logger.Error("自动启动指标采集失败", "error", startErr)
				err = startErr
			} else {
				logger.Info("指标采集器自动启动成功")
			}
		}
	})
	
	return err
}

// NewMetricCollectorManager 创建指标采集管理器
// 创建一个新的指标采集管理器实例，包括配置加载、DAO初始化等
//
// 参数：
//   - db: 数据库连接实例
//
// 返回：
//   - *MetricCollectorManager: 管理器实例
//   - error: 创建过程中的错误
func NewMetricCollectorManager(db database.Database) (*MetricCollectorManager, error) {
	// 创建配置验证器
	configValidator := NewConfigValidator()
	
	// 加载配置
	cfg, err := configValidator.LoadMetricConfig()
	if err != nil {
		return nil, fmt.Errorf("加载指标采集配置失败: %w", err)
	}
	
	// 检查是否启用
	if !cfg.Enabled {
		logger.Info("指标采集功能已禁用")
		return nil, nil
	}
	
	// 创建管理器实例
	manager := &MetricCollectorManager{
		config:          cfg,
		db:              db,
		stopChan:        make(chan struct{}),
		dataBuffer:      &DataBuffer{},
		configValidator: configValidator,
	}
	
	// 初始化DAO
	manager.initDAOs()
	
	// 创建服务器信息管理器
	manager.serverInfoManager = NewServerInfoManager(cfg, db)
	
	// 初始化服务器信息
	if err := manager.serverInfoManager.InitializeServerInfo(); err != nil {
		return nil, fmt.Errorf("初始化服务器信息失败: %w", err)
	}
	
	// 配置metric采集器
	manager.configureCollectors()
	
	logger.Info("指标采集管理器初始化成功", 
		"config", cfg.String())
	
	return manager, nil
}

// initDAOs 初始化DAO实例
// 创建所有需要的数据访问对象，用于数据库操作
func (m *MetricCollectorManager) initDAOs() {
	m.serverInfoDAO = dao.NewServerInfoDAO(m.db)
	m.cpuLogDAO = dao.NewCpuLogDAO(m.db)
	m.memoryLogDAO = dao.NewMemoryLogDAO(m.db)
	m.diskPartitionLogDAO = dao.NewDiskPartitionLogDAO(m.db)
	m.diskIoLogDAO = dao.NewDiskIoLogDAO(m.db)
	m.networkLogDAO = dao.NewNetworkLogDAO(m.db)
	m.processLogDAO = dao.NewProcessLogDAO(m.db)
	m.processStatsLogDAO = dao.NewProcessStatsLogDAO(m.db)
	m.temperatureLogDAO = dao.NewTemperatureLogDAO(m.db)
	
	logger.Debug("DAO实例初始化完成")
}



// configureCollectors 配置采集器
// 根据配置启用或禁用相应的指标采集器
//
// 流程：
// 1. 先禁用所有采集器
// 2. 根据配置启用指定的采集器
// 3. 记录配置结果
func (m *MetricCollectorManager) configureCollectors() {
	// 禁用所有采集器
	for _, name := range metric.GetCollectorNames() {
		metric.DisableCollector(name)
	}
	
	// 根据配置启用采集器
	if m.config.Collectors.CPU {
		metric.EnableCollector(metricTypes.CollectorNameCPU)
	}
	if m.config.Collectors.Memory {
		metric.EnableCollector(metricTypes.CollectorNameMemory)
	}
	if m.config.Collectors.Disk {
		metric.EnableCollector(metricTypes.CollectorNameDisk)
	}
	if m.config.Collectors.Network {
		metric.EnableCollector(metricTypes.CollectorNameNetwork)
	}
	if m.config.Collectors.Process {
		metric.EnableCollector(metricTypes.CollectorNameProcess)
	}
	if m.config.Collectors.System {
		metric.EnableCollector(metricTypes.CollectorNameSystem)
	}
	
	logger.Info("采集器配置完成", 
		"enabled_collectors", m.config.GetEnabledCollectorNames())
}

// Start 启动指标采集
// 启动指标采集服务，包括数据采集、数据刷新、数据清理等协程
//
// 返回：
//   - error: 启动过程中的错误
//
// 注意：
//   - 使用协程异步执行各种任务
//   - 支持优雅停止机制
func (m *MetricCollectorManager) Start() error {
	if m == nil || !m.config.Enabled {
		return fmt.Errorf("指标采集未启用")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.running {
		return fmt.Errorf("指标采集已在运行")
	}
	
	m.running = true
	m.stopChan = make(chan struct{})
	
	// 启动采集协程
	go m.collectLoop()
	
	// 启动数据刷新协程
	go m.flushLoop()
	
	// 启动数据清理协程（如果启用）
	if m.config.Storage.Retention.Enabled {
		go m.cleanupLoop()
	}
	
	logger.Info("指标采集启动成功", 
		"collect_interval", m.config.CollectInterval,
		"flush_interval", m.config.Storage.FlushInterval)
	
	return nil
}

// Stop 停止指标采集
// 优雅停止指标采集服务，确保数据完整性
//
// 返回：
//   - error: 停止过程中的错误
//
// 流程：
// 1. 发送停止信号给所有协程
// 2. 执行最后一次数据刷新
// 3. 更新运行状态
func (m *MetricCollectorManager) Stop() error {
	if m == nil {
		return nil
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.running {
		return fmt.Errorf("指标采集未在运行")
	}
	
	m.running = false
	close(m.stopChan)
	
	// 最后一次刷新数据
	m.flushData()
	
	logger.Info("指标采集停止成功")
	return nil
}

// IsRunning 检查是否正在运行
// 线程安全地检查指标采集器的运行状态
//
// 返回：
//   - bool: true表示正在运行，false表示已停止
func (m *MetricCollectorManager) IsRunning() bool {
	if m == nil {
		return false
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetGlobalManager 获取全局管理器实例
// 返回全局的指标采集管理器实例
//
// 返回：
//   - *MetricCollectorManager: 全局管理器实例，可能为nil
func GetGlobalManager() *MetricCollectorManager {
	return globalManager
}

// GetConfig 获取当前配置
// 返回当前使用的配置信息
//
// 返回：
//   - *MetricConfig: 配置信息，如果管理器未初始化则返回nil
func (m *MetricCollectorManager) GetConfig() *MetricConfig {
	if m == nil {
		return nil
	}
	return m.config
}

// GetServerInfo 获取服务器信息
// 返回当前服务器的基本信息
//
// 返回：
//   - *types.ServerInfo: 服务器信息，如果未初始化则返回nil
func (m *MetricCollectorManager) GetServerInfo() *types.ServerInfo {
	if m == nil || m.serverInfoManager == nil {
		return nil
	}
	return m.serverInfoManager.GetServerInfo()
}

// collectLoop 采集循环
// 定时执行指标采集任务的主循环
//
// 特性：
// - 使用定时器确保采集间隔的准确性
// - 支持优雅停止机制
// - 错误处理不中断循环
func (m *MetricCollectorManager) collectLoop() {
	ticker := time.NewTicker(m.config.CollectInterval)
	defer ticker.Stop()
	
	logger.Debug("采集循环启动", "interval", m.config.CollectInterval)
	
	for {
		select {
		case <-ticker.C:
			m.collectMetrics()
		case <-m.stopChan:
			logger.Debug("采集循环停止")
			return
		}
	}
}

// flushLoop 数据刷新循环
// 定时将缓冲区数据刷新到数据库的主循环
//
// 特性：
// - 批量写入提高性能
// - 定时刷新避免数据丢失
// - 支持优雅停止机制
func (m *MetricCollectorManager) flushLoop() {
	ticker := time.NewTicker(m.config.Storage.FlushInterval)
	defer ticker.Stop()
	
	logger.Debug("数据刷新循环启动", "interval", m.config.Storage.FlushInterval)
	
	for {
		select {
		case <-ticker.C:
			m.flushData()
		case <-m.stopChan:
			logger.Debug("数据刷新循环停止")
			return
		}
	}
}

// cleanupLoop 数据清理循环
// 定时清理过期历史数据的主循环
//
// 特性：
// - 自动清理过期数据，节省存储空间
// - 可配置清理策略
// - 支持优雅停止机制
func (m *MetricCollectorManager) cleanupLoop() {
	ticker := time.NewTicker(m.config.Storage.Retention.CleanupInterval)
	defer ticker.Stop()
	
	logger.Debug("数据清理循环启动", "interval", m.config.Storage.Retention.CleanupInterval)
	
	for {
		select {
		case <-ticker.C:
			m.cleanupOldData()
		case <-m.stopChan:
			logger.Debug("数据清理循环停止")
			return
		}
	}
}

// collectMetrics 采集指标
// 执行一次完整的指标采集任务
//
// 流程：
// 1. 调用metric包采集所有指标
// 2. 转换指标数据为数据库记录格式
// 3. 将数据加入缓冲区
//
// 错误处理：
// - 采集失败时记录错误日志但不中断服务
func (m *MetricCollectorManager) collectMetrics() {
	collectTime := time.Now()
	oprSeqFlag := fmt.Sprintf("COLLECT_%d", collectTime.Unix())
	
	logger.Debug("开始采集指标", "collect_time", collectTime)
	
	// 采集所有指标
	allMetrics, err := metric.CollectAll()
	if err != nil {
		logger.Error("采集指标失败", "error", err)
		return
	}
	
	// 转换并缓存数据
	m.convertAndBufferMetrics(allMetrics, collectTime, oprSeqFlag)
	
	logger.Debug("指标采集完成", "collect_time", collectTime)
}

// convertAndBufferMetrics 转换并缓存指标数据
// 将采集到的原始指标数据转换为数据库记录格式，并加入缓冲区
//
// 参数：
//   - allMetrics: 采集到的所有指标数据
//   - collectTime: 采集时间
//   - oprSeqFlag: 操作序列标识
//
// 特性：
// - 使用类型转换方法确保数据格式正确
// - 线程安全的缓冲区操作
// - 根据配置选择性处理不同类型的指标
func (m *MetricCollectorManager) convertAndBufferMetrics(allMetrics *metricTypes.AllMetrics, collectTime time.Time, oprSeqFlag string) {
	m.dataBuffer.mu.Lock()
	defer m.dataBuffer.mu.Unlock()
	
	// CPU指标
	if allMetrics.CPU != nil && m.config.Collectors.CPU {
		cpuLog := types.NewCpuLogFromMetrics(allMetrics.CPU, m.config.TenantId, m.config.ServerId, m.config.Operator, collectTime, oprSeqFlag)
		m.dataBuffer.cpuLogs = append(m.dataBuffer.cpuLogs, cpuLog)
	}
	
	// 内存指标
	if allMetrics.Memory != nil && m.config.Collectors.Memory {
		memoryLog := types.NewMemoryLogFromMetrics(allMetrics.Memory, m.config.TenantId, m.config.ServerId, m.config.Operator, collectTime, oprSeqFlag)
		m.dataBuffer.memoryLogs = append(m.dataBuffer.memoryLogs, memoryLog)
	}
	
	// 磁盘指标
	if allMetrics.Disk != nil && m.config.Collectors.Disk {
		// 磁盘分区日志
		partitionLogs := types.NewDiskPartitionLogsFromMetrics(allMetrics.Disk, m.config.TenantId, m.config.ServerId, m.config.Operator, collectTime, oprSeqFlag)
		m.dataBuffer.diskPartitionLogs = append(m.dataBuffer.diskPartitionLogs, partitionLogs...)
		
		// 磁盘IO日志
		ioLogs := types.NewDiskIoLogsFromMetrics(allMetrics.Disk, m.config.TenantId, m.config.ServerId, m.config.Operator, collectTime, oprSeqFlag)
		m.dataBuffer.diskIoLogs = append(m.dataBuffer.diskIoLogs, ioLogs...)
	}
	
	// 网络指标
	if allMetrics.Network != nil && m.config.Collectors.Network {
		networkLogs := types.NewNetworkLogsFromMetrics(allMetrics.Network, m.config.TenantId, m.config.ServerId, m.config.Operator, collectTime, oprSeqFlag)
		m.dataBuffer.networkLogs = append(m.dataBuffer.networkLogs, networkLogs...)
	}
	
	// 进程指标
	if allMetrics.Process != nil && m.config.Collectors.Process {
		// 进程统计日志
		if allMetrics.Process.SystemProcesses != nil {
			statsLog := types.NewProcessStatsLogFromMetrics(allMetrics.Process.SystemProcesses, m.config.TenantId, m.config.ServerId, m.config.Operator, collectTime, oprSeqFlag)
			m.dataBuffer.processStatsLogs = append(m.dataBuffer.processStatsLogs, statsLog)
		}
		
		// 当前进程日志
		if allMetrics.Process.CurrentProcess != nil {
			processLog := types.NewProcessLogFromMetrics(allMetrics.Process.CurrentProcess, m.config.TenantId, m.config.ServerId, m.config.Operator, collectTime, oprSeqFlag)
			m.dataBuffer.processLogs = append(m.dataBuffer.processLogs, processLog)
		}
	}
	
	logger.Debug("指标数据已加入缓冲区",
		"cpu_logs", len(m.dataBuffer.cpuLogs),
		"memory_logs", len(m.dataBuffer.memoryLogs),
		"disk_partition_logs", len(m.dataBuffer.diskPartitionLogs),
		"disk_io_logs", len(m.dataBuffer.diskIoLogs),
		"network_logs", len(m.dataBuffer.networkLogs),
		"process_logs", len(m.dataBuffer.processLogs),
		"process_stats_logs", len(m.dataBuffer.processStatsLogs))
}

// flushData 刷新数据到数据库
func (m *MetricCollectorManager) flushData() {
	ctx := context.Background()
	
	// 获取缓冲区数据
	m.dataBuffer.mu.Lock()
	cpuLogs := m.dataBuffer.cpuLogs
	memoryLogs := m.dataBuffer.memoryLogs
	diskPartitionLogs := m.dataBuffer.diskPartitionLogs
	diskIoLogs := m.dataBuffer.diskIoLogs
	networkLogs := m.dataBuffer.networkLogs
	processLogs := m.dataBuffer.processLogs
	processStatsLogs := m.dataBuffer.processStatsLogs
	temperatureLogs := m.dataBuffer.temperatureLogs
	
	// 清空缓冲区
	m.dataBuffer.cpuLogs = nil
	m.dataBuffer.memoryLogs = nil
	m.dataBuffer.diskPartitionLogs = nil
	m.dataBuffer.diskIoLogs = nil
	m.dataBuffer.networkLogs = nil
	m.dataBuffer.processLogs = nil
	m.dataBuffer.processStatsLogs = nil
	m.dataBuffer.temperatureLogs = nil
	m.dataBuffer.mu.Unlock()
	
	// 批量插入数据
	if len(cpuLogs) > 0 {
		if err := m.cpuLogDAO.BatchInsertCpuLog(ctx, cpuLogs); err != nil {
			logger.Error("批量插入CPU日志失败", "error", err, "count", len(cpuLogs))
		} else {
			logger.Debug("CPU日志插入成功", "count", len(cpuLogs))
		}
	}
	
	if len(memoryLogs) > 0 {
		if err := m.memoryLogDAO.BatchInsertMemoryLog(ctx, memoryLogs); err != nil {
			logger.Error("批量插入内存日志失败", "error", err, "count", len(memoryLogs))
		} else {
			logger.Debug("内存日志插入成功", "count", len(memoryLogs))
		}
	}
	
	if len(diskPartitionLogs) > 0 {
		if err := m.diskPartitionLogDAO.BatchInsertDiskPartitionLog(ctx, diskPartitionLogs); err != nil {
			logger.Error("批量插入磁盘分区日志失败", "error", err, "count", len(diskPartitionLogs))
		} else {
			logger.Debug("磁盘分区日志插入成功", "count", len(diskPartitionLogs))
		}
	}
	
	if len(diskIoLogs) > 0 {
		if err := m.diskIoLogDAO.BatchInsertDiskIoLog(ctx, diskIoLogs); err != nil {
			logger.Error("批量插入磁盘IO日志失败", "error", err, "count", len(diskIoLogs))
		} else {
			logger.Debug("磁盘IO日志插入成功", "count", len(diskIoLogs))
		}
	}
	
	if len(networkLogs) > 0 {
		if err := m.networkLogDAO.BatchInsertNetworkLog(ctx, networkLogs); err != nil {
			logger.Error("批量插入网络日志失败", "error", err, "count", len(networkLogs))
		} else {
			logger.Debug("网络日志插入成功", "count", len(networkLogs))
		}
	}
	
	if len(processLogs) > 0 {
		if err := m.processLogDAO.BatchInsertProcessLog(ctx, processLogs); err != nil {
			logger.Error("批量插入进程日志失败", "error", err, "count", len(processLogs))
		} else {
			logger.Debug("进程日志插入成功", "count", len(processLogs))
		}
	}
	
	if len(processStatsLogs) > 0 {
		if err := m.processStatsLogDAO.BatchInsertProcessStatsLog(ctx, processStatsLogs); err != nil {
			logger.Error("批量插入进程统计日志失败", "error", err, "count", len(processStatsLogs))
		} else {
			logger.Debug("进程统计日志插入成功", "count", len(processStatsLogs))
		}
	}
	
	if len(temperatureLogs) > 0 {
		if err := m.temperatureLogDAO.BatchInsertTemperatureLog(ctx, temperatureLogs); err != nil {
			logger.Error("批量插入温度日志失败", "error", err, "count", len(temperatureLogs))
		} else {
			logger.Debug("温度日志插入成功", "count", len(temperatureLogs))
		}
	}
}

// cleanupOldData 清理过期数据
func (m *MetricCollectorManager) cleanupOldData() {
	if !m.config.Storage.Retention.Enabled {
		return
	}
	
	ctx := context.Background()
	beforeTime := time.Now().AddDate(0, 0, -m.config.Storage.Retention.KeepDays)
	
	logger.Info("开始清理过期数据", "before_time", beforeTime, "keep_days", m.config.Storage.Retention.KeepDays)
	
	// 清理各类日志数据
	if err := m.cpuLogDAO.DeleteCpuLogByTime(ctx, m.config.TenantId, beforeTime); err != nil {
		logger.Error("清理CPU日志失败", "error", err)
	}
	
	if err := m.memoryLogDAO.DeleteMemoryLogByTime(ctx, m.config.TenantId, beforeTime); err != nil {
		logger.Error("清理内存日志失败", "error", err)
	}
	
	if err := m.diskPartitionLogDAO.DeleteDiskPartitionLogByTime(ctx, m.config.TenantId, beforeTime); err != nil {
		logger.Error("清理磁盘分区日志失败", "error", err)
	}
	
	if err := m.diskIoLogDAO.DeleteDiskIoLogByTime(ctx, m.config.TenantId, beforeTime); err != nil {
		logger.Error("清理磁盘IO日志失败", "error", err)
	}
	
	if err := m.networkLogDAO.DeleteNetworkLogByTime(ctx, m.config.TenantId, beforeTime); err != nil {
		logger.Error("清理网络日志失败", "error", err)
	}
	
	if err := m.processLogDAO.DeleteProcessLogByTime(ctx, m.config.TenantId, beforeTime); err != nil {
		logger.Error("清理进程日志失败", "error", err)
	}
	
	if err := m.processStatsLogDAO.DeleteProcessStatsLogByTime(ctx, m.config.TenantId, beforeTime); err != nil {
		logger.Error("清理进程统计日志失败", "error", err)
	}
	
	if err := m.temperatureLogDAO.DeleteTemperatureLogByTime(ctx, m.config.TenantId, beforeTime); err != nil {
		logger.Error("清理温度日志失败", "error", err)
	}
	
	logger.Info("数据清理完成")
} 