-- 系统指标采集表结构创建脚本
-- 基于 pkg/metric/types/metrics.go 中的数据结构设计
-- 遵循项目数据库设计规范

-- ===================================================================
-- 字段长度优化说明 (HUB_METRIC_SERVER_INFO)
-- ===================================================================
-- 
-- 1. 主键字段调整:
--    - metricServerId: 32 -> 64 字符，适应MD5哈希和UUID格式
--    - tenantId: 32 -> 64 字符，支持更复杂的租户标识
-- 
-- 2. 系统信息字段调整:
--    - hostname: 100 -> 255 字符，支持FQDN和长主机名
--    - osType: 50 -> 100 字符，支持详细的操作系统描述
--    - osVersion: 100 -> 255 字符，支持完整的版本信息
--    - kernelVersion: 100 -> 255 字符，支持详细的内核版本
--    - architecture: 50 -> 100 字符，支持复杂的架构描述
--    - serverLocation: 100 -> 255 字符，支持详细的位置描述
-- 
-- 3. 网络信息字段优化:
--    - ipAddress: 50 -> 45 字符，IPv6最大长度为39字符，预留6字符
--    - macAddress: 50 -> 17 字符，MAC地址标准格式为17字符
-- 
-- 4. 新增TEXT字段用于存储复杂数据:
--    - networkInfo: 存储完整的网络信息（所有IP、MAC、接口等）
--    - systemInfo: 存储系统扩展信息（温度、负载、进程统计等）
--    - hardwareInfo: 存储硬件详细信息（CPU详情、内存详情等）
--    - noteText: VARCHAR(500) -> TEXT，支持更长的备注信息
-- 
-- 5. 操作字段调整:
--    - addWho/editWho: 32 -> 64 字符，支持更长的用户标识
--    - oprSeqFlag: 32 -> 64 字符，支持更复杂的操作序列标识
-- 
-- 6. 新增索引:
--    - IDX_METRIC_SERVER_TYPE: 支持按服务器类型查询
-- 
-- ===================================================================
-- JSON字段存储格式示例
-- ===================================================================
-- 
-- networkInfo 字段存储格式:
-- {
--   "primaryIP": "192.168.1.100",
--   "primaryMAC": "00:11:22:33:44:55",
--   "primaryInterface": "eth0",
--   "allIPs": ["192.168.1.100", "10.0.0.1"],
--   "allMACs": ["00:11:22:33:44:55", "00:11:22:33:44:56"],
--   "activeInterfaces": ["eth0", "lo"]
-- }
-- 
-- systemInfo 字段存储格式:
-- {
--   "uptime": 86400,
--   "userCount": 5,
--   "processCount": 150,
--   "loadAvg": {"1min": 0.5, "5min": 0.3, "15min": 0.2},
--   "temperatures": [
--     {"sensor": "CPU", "value": 45.5, "high": 80.0, "critical": 90.0}
--   ]
-- }
-- 
-- hardwareInfo 字段存储格式:
-- {
--   "cpu": {
--     "coreCount": 8,
--     "logicalCount": 16,
--     "model": "Intel Core i7-9700K",
--     "frequency": "3.6GHz"
--   },
--   "memory": {
--     "total": 17179869184,
--     "type": "DDR4",
--     "speed": "3200MHz"
--   },
--   "storage": {
--     "totalDisks": 2,
--     "totalCapacity": 2000000000000
--   }
-- }
-- 
-- ===================================================================

-- 1. 服务器信息主表
CREATE TABLE `HUB_METRIC_SERVER_INFO` (
  `metricServerId` VARCHAR(64) NOT NULL COMMENT '服务器ID',
  `tenantId` VARCHAR(64) NOT NULL COMMENT '租户ID',
  `hostname` VARCHAR(255) NOT NULL COMMENT '主机名',
  `osType` VARCHAR(100) NOT NULL COMMENT '操作系统类型',
  `osVersion` VARCHAR(255) NOT NULL COMMENT '操作系统版本',
  `kernelVersion` VARCHAR(255) DEFAULT NULL COMMENT '内核版本',
  `architecture` VARCHAR(100) NOT NULL COMMENT '系统架构',
  `bootTime` DATETIME NOT NULL COMMENT '系统启动时间',
  `ipAddress` VARCHAR(45) DEFAULT NULL COMMENT '主IP地址',
  `macAddress` VARCHAR(50) DEFAULT NULL COMMENT '主MAC地址',
  `serverLocation` VARCHAR(255) DEFAULT NULL COMMENT '服务器位置',
  `serverType` VARCHAR(50) DEFAULT NULL COMMENT '服务器类型(physical/virtual/unknown)',
  `lastUpdateTime` DATETIME NOT NULL COMMENT '最后更新时间',
  -- 新增网络信息字段
  `networkInfo` TEXT DEFAULT NULL COMMENT '网络信息详情，JSON格式存储所有IP和MAC地址',
  `systemInfo` TEXT DEFAULT NULL COMMENT '系统详细信息，JSON格式存储温度、负载等扩展信息',
  `hardwareInfo` TEXT DEFAULT NULL COMMENT '硬件信息，JSON格式存储CPU、内存、磁盘等硬件详情',
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(64) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(64) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(64) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` TEXT DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricServerId`),
  UNIQUE KEY `IDX_METRIC_SERVER_HOST` (`hostname`),
  KEY `IDX_METRIC_SERVER_OS` (`osType`),
  KEY `IDX_METRIC_SERVER_IP` (`ipAddress`),
  KEY `IDX_METRIC_SERVER_TYPE` (`serverType`),
  KEY `IDX_METRIC_SERVER_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_SERVER_UPDATE` (`lastUpdateTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器信息主表';

-- 2. CPU采集日志表
CREATE TABLE `HUB_METRIC_CPU_LOG` (
  `metricCpuLogId` VARCHAR(32) NOT NULL COMMENT 'CPU采集日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `usagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'CPU使用率(0-100)',
  `userPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '用户态CPU使用率',
  `systemPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '系统态CPU使用率',
  `idlePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '空闲CPU使用率',
  `ioWaitPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'I/O等待CPU使用率',
  `irqPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '中断处理CPU使用率',
  `softIrqPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '软中断处理CPU使用率',
  `coreCount` INT NOT NULL DEFAULT 0 COMMENT 'CPU核心数',
  `logicalCount` INT NOT NULL DEFAULT 0 COMMENT '逻辑CPU数',
  `loadAvg1` DECIMAL(8,2) NOT NULL DEFAULT 0.00 COMMENT '1分钟负载平均值',
  `loadAvg5` DECIMAL(8,2) NOT NULL DEFAULT 0.00 COMMENT '5分钟负载平均值',
  `loadAvg15` DECIMAL(8,2) NOT NULL DEFAULT 0.00 COMMENT '15分钟负载平均值',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricCpuLogId`),
  KEY `IDX_METRIC_CPU_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_CPU_TIME` (`collectTime`),
  KEY `IDX_METRIC_CPU_USAGE` (`usagePercent`),
  KEY `IDX_METRIC_CPU_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_CPU_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_CPU_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CPU采集日志表';

-- 3. 内存采集日志表
CREATE TABLE `HUB_METRIC_MEMORY_LOG` (
  `metricMemoryLogId` VARCHAR(32) NOT NULL COMMENT '内存采集日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `totalMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '总内存(字节)',
  `availableMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '可用内存(字节)',
  `usedMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '已使用内存(字节)',
  `usagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '内存使用率(0-100)',
  `freeMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '空闲内存(字节)',
  `cachedMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '缓存内存(字节)',
  `buffersMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '缓冲区内存(字节)',
  `sharedMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '共享内存(字节)',
  `swapTotal` BIGINT NOT NULL DEFAULT 0 COMMENT '交换区总大小(字节)',
  `swapUsed` BIGINT NOT NULL DEFAULT 0 COMMENT '交换区已使用(字节)',
  `swapFree` BIGINT NOT NULL DEFAULT 0 COMMENT '交换区空闲(字节)',
  `swapUsagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '交换区使用率(0-100)',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricMemoryLogId`),
  KEY `IDX_METRIC_MEMORY_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_MEMORY_TIME` (`collectTime`),
  KEY `IDX_METRIC_MEMORY_USAGE` (`usagePercent`),
  KEY `IDX_METRIC_MEMORY_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_MEMORY_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_MEMORY_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='内存采集日志表';

-- 4. 磁盘分区日志表
CREATE TABLE `HUB_METRIC_DISK_PART_LOG` (
  `metricDiskPartitionLogId` VARCHAR(32) NOT NULL COMMENT '磁盘分区日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `deviceName` VARCHAR(100) NOT NULL COMMENT '设备名称',
  `mountPoint` VARCHAR(200) NOT NULL COMMENT '挂载点',
  `fileSystem` VARCHAR(50) NOT NULL COMMENT '文件系统类型',
  `totalSpace` BIGINT NOT NULL DEFAULT 0 COMMENT '总大小(字节)',
  `usedSpace` BIGINT NOT NULL DEFAULT 0 COMMENT '已使用(字节)',
  `freeSpace` BIGINT NOT NULL DEFAULT 0 COMMENT '可用(字节)',
  `usagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '使用率(0-100)',
  `inodesTotal` BIGINT NOT NULL DEFAULT 0 COMMENT 'inode总数',
  `inodesUsed` BIGINT NOT NULL DEFAULT 0 COMMENT 'inode已使用',
  `inodesFree` BIGINT NOT NULL DEFAULT 0 COMMENT 'inode空闲',
  `inodesUsagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'inode使用率(0-100)',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricDiskPartitionLogId`),
  KEY `IDX_METRIC_DISK_PART_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_DISK_PART_TIME` (`collectTime`),
  KEY `IDX_METRIC_DISK_PART_DEVICE` (`deviceName`),
  KEY `IDX_METRIC_DISK_PART_USAGE` (`usagePercent`),
  KEY `IDX_METRIC_DISK_PART_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_DISK_PART_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_DISK_PART_SRV_DEV` (`metricServerId`, `deviceName`),
  KEY `IDX_METRIC_DISK_PART_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='磁盘分区采集日志表';

-- 5. 磁盘IO日志表
CREATE TABLE `HUB_METRIC_DISK_IO_LOG` (
  `metricDiskIoLogId` VARCHAR(32) NOT NULL COMMENT '磁盘IO日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `deviceName` VARCHAR(100) NOT NULL COMMENT '设备名称',
  `readCount` BIGINT NOT NULL DEFAULT 0 COMMENT '读取次数',
  `writeCount` BIGINT NOT NULL DEFAULT 0 COMMENT '写入次数',
  `readBytes` BIGINT NOT NULL DEFAULT 0 COMMENT '读取字节数',
  `writeBytes` BIGINT NOT NULL DEFAULT 0 COMMENT '写入字节数',
  `readTime` BIGINT NOT NULL DEFAULT 0 COMMENT '读取时间(毫秒)',
  `writeTime` BIGINT NOT NULL DEFAULT 0 COMMENT '写入时间(毫秒)',
  `ioInProgress` BIGINT NOT NULL DEFAULT 0 COMMENT 'IO进行中数量',
  `ioTime` BIGINT NOT NULL DEFAULT 0 COMMENT 'IO时间(毫秒)',
  `readRate` DECIMAL(20,2) NOT NULL DEFAULT 0.00 COMMENT '读取速率(字节/秒)',
  `writeRate` DECIMAL(20,2) NOT NULL DEFAULT 0.00 COMMENT '写入速率(字节/秒)',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricDiskIoLogId`),
  KEY `IDX_METRIC_DISK_IO_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_DISK_IO_TIME` (`collectTime`),
  KEY `IDX_METRIC_DISK_IO_DEVICE` (`deviceName`),
  KEY `IDX_METRIC_DISK_IO_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_DISK_IO_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_DISK_IO_SRV_DEV` (`metricServerId`, `deviceName`),
  KEY `IDX_METRIC_DISK_IO_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='磁盘IO采集日志表';

-- 6. 网络接口日志表
CREATE TABLE `HUB_METRIC_NETWORK_LOG` (
  `metricNetworkLogId` VARCHAR(32) NOT NULL COMMENT '网络接口日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `interfaceName` VARCHAR(100) NOT NULL COMMENT '接口名称',
  `hardwareAddr` VARCHAR(50) DEFAULT NULL COMMENT 'MAC地址',
  `ipAddresses` TEXT DEFAULT NULL COMMENT 'IP地址列表，JSON格式',
  `interfaceStatus` VARCHAR(20) NOT NULL COMMENT '接口状态',
  `interfaceType` VARCHAR(50) DEFAULT NULL COMMENT '接口类型',
  `bytesReceived` BIGINT NOT NULL DEFAULT 0 COMMENT '接收字节数',
  `bytesSent` BIGINT NOT NULL DEFAULT 0 COMMENT '发送字节数',
  `packetsReceived` BIGINT NOT NULL DEFAULT 0 COMMENT '接收包数',
  `packetsSent` BIGINT NOT NULL DEFAULT 0 COMMENT '发送包数',
  `errorsReceived` BIGINT NOT NULL DEFAULT 0 COMMENT '接收错误数',
  `errorsSent` BIGINT NOT NULL DEFAULT 0 COMMENT '发送错误数',
  `droppedReceived` BIGINT NOT NULL DEFAULT 0 COMMENT '接收丢包数',
  `droppedSent` BIGINT NOT NULL DEFAULT 0 COMMENT '发送丢包数',
  `receiveRate` DECIMAL(20,2) DEFAULT 0 COMMENT '接收速率(字节/秒)',
  `sendRate` DECIMAL(20,2) DEFAULT 0 COMMENT '发送速率(字节/秒)',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricNetworkLogId`),
  KEY `IDX_METRIC_NETWORK_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_NETWORK_TIME` (`collectTime`),
  KEY `IDX_METRIC_NETWORK_INTERFACE` (`interfaceName`),
  KEY `IDX_METRIC_NETWORK_STATUS` (`interfaceStatus`),
  KEY `IDX_METRIC_NETWORK_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_NETWORK_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_NETWORK_SRV_INT` (`metricServerId`, `interfaceName`),
  KEY `IDX_METRIC_NETWORK_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网络接口采集日志表';

-- 7. 进程信息日志表
CREATE TABLE `HUB_METRIC_PROCESS_LOG` (
  `metricProcessLogId` VARCHAR(32) NOT NULL COMMENT '进程信息日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `processId` INT NOT NULL COMMENT '进程ID',
  `parentProcessId` INT DEFAULT NULL COMMENT '父进程ID',
  `processName` VARCHAR(200) NOT NULL COMMENT '进程名称',
  `processStatus` VARCHAR(50) NOT NULL COMMENT '进程状态',
  `createTime` DATETIME NOT NULL COMMENT '进程启动时间',
  `runTime` BIGINT NOT NULL DEFAULT 0 COMMENT '进程运行时间(秒)',
  `memoryUsage` BIGINT NOT NULL DEFAULT 0 COMMENT '内存使用(字节)',
  `memoryPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '内存使用率(0-100)',
  `cpuPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'CPU使用率(0-100)',
  `threadCount` INT NOT NULL DEFAULT 0 COMMENT '线程数',
  `fileDescriptorCount` INT NOT NULL DEFAULT 0 COMMENT '文件句柄数',
  `commandLine` TEXT DEFAULT NULL COMMENT '命令行参数，JSON格式',
  `executablePath` VARCHAR(500) DEFAULT NULL COMMENT '执行路径',
  `workingDirectory` VARCHAR(500) DEFAULT NULL COMMENT '工作目录',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricProcessLogId`),
  KEY `IDX_METRIC_PROCESS_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_PROCESS_TIME` (`collectTime`),
  KEY `IDX_METRIC_PROCESS_PID` (`processId`),
  KEY `IDX_METRIC_PROCESS_NAME` (`processName`),
  KEY `IDX_METRIC_PROCESS_STATUS` (`processStatus`),
  KEY `IDX_METRIC_PROCESS_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_PROCESS_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_PROCESS_SRV_PID` (`metricServerId`, `processId`),
  KEY `IDX_METRIC_PROCESS_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='进程信息采集日志表';

-- 8. 进程统计日志表
CREATE TABLE `HUB_METRIC_PROCSTAT_LOG` (
  `metricProcessStatsLogId` VARCHAR(32) NOT NULL COMMENT '进程统计日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `runningCount` INT NOT NULL DEFAULT 0 COMMENT '运行中进程数',
  `sleepingCount` INT NOT NULL DEFAULT 0 COMMENT '睡眠中进程数',
  `stoppedCount` INT NOT NULL DEFAULT 0 COMMENT '停止的进程数',
  `zombieCount` INT NOT NULL DEFAULT 0 COMMENT '僵尸进程数',
  `totalCount` INT NOT NULL DEFAULT 0 COMMENT '总进程数',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricProcessStatsLogId`),
  KEY `IDX_METRIC_PROC_STATS_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_PROC_STATS_TIME` (`collectTime`),
  KEY `IDX_METRIC_PROC_STATS_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_PROC_STATS_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_PROC_STATS_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='进程统计采集日志表';

-- 9. 温度信息日志表
CREATE TABLE `HUB_METRIC_TEMPERATURE_LOG` (
  `metricTemperatureLogId` VARCHAR(32) NOT NULL COMMENT '温度信息日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `sensorName` VARCHAR(100) NOT NULL COMMENT '传感器名称',
  `temperatureValue` DECIMAL(6,2) NOT NULL DEFAULT 0.00 COMMENT '温度值(摄氏度)',
  `highThreshold` DECIMAL(6,2) DEFAULT NULL COMMENT '高温阈值',
  `criticalThreshold` DECIMAL(6,2) DEFAULT NULL COMMENT '严重高温阈值',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  PRIMARY KEY (`tenantId`, `metricTemperatureLogId`),
  KEY `IDX_METRIC_TEMP_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_TEMP_TIME` (`collectTime`),
  KEY `IDX_METRIC_TEMP_SENSOR` (`sensorName`),
  KEY `IDX_METRIC_TEMP_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_TEMP_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_TEMP_SRV_SENSOR` (`metricServerId`, `sensorName`),
  KEY `IDX_METRIC_TEMP_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='温度信息采集日志表'; 