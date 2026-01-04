CREATE TABLE `HUB_TIMER_SCHEDULER` (
  `schedulerId` VARCHAR(32) NOT NULL COMMENT '调度器ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `schedulerName` VARCHAR(100) NOT NULL COMMENT '调度器名称',
  `schedulerInstanceId` VARCHAR(100) DEFAULT NULL COMMENT '调度器实例ID，用于集群环境区分',
  
  -- 调度器配置
  `maxWorkers` INT NOT NULL DEFAULT 5 COMMENT '最大工作线程数',
  `queueSize` INT NOT NULL DEFAULT 100 COMMENT '任务队列大小',
  `defaultTimeoutSeconds` BIGINT NOT NULL DEFAULT 1800 COMMENT '默认超时时间秒数',
  `defaultRetries` INT NOT NULL DEFAULT 3 COMMENT '默认重试次数',
  
  -- 调度器状态
  `schedulerStatus` INT NOT NULL DEFAULT 1 COMMENT '调度器状态(1停止,2运行中,3暂停)',
  `lastStartTime` DATETIME DEFAULT NULL COMMENT '最后启动时间',
  `lastStopTime` DATETIME DEFAULT NULL COMMENT '最后停止时间',
  
  -- 服务器信息
  `serverName` VARCHAR(100) DEFAULT NULL COMMENT '服务器名称',
  `serverIp` VARCHAR(50) DEFAULT NULL COMMENT '服务器IP地址',
  `serverPort` INT DEFAULT NULL COMMENT '服务器端口',
  
  -- 监控信息
  `totalTaskCount` INT NOT NULL DEFAULT 0 COMMENT '总任务数',
  `runningTaskCount` INT NOT NULL DEFAULT 0 COMMENT '运行中任务数',
  `lastHeartbeatTime` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  
  -- 通用字段
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
  
  PRIMARY KEY (`tenantId`, `schedulerId`),
  KEY `IDX_TIMER_SCHED_NAME` (`schedulerName`),
  KEY `IDX_TIMER_SCHED_INST` (`schedulerInstanceId`),
  KEY `IDX_TIMER_SCHED_STATUS` (`schedulerStatus`),
  KEY `IDX_TIMER_SCHED_HEART` (`lastHeartbeatTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务调度器表 - 存储调度器配置和状态信息';
