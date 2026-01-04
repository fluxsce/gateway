CREATE TABLE `HUB_GW_LOG_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `logConfigId` VARCHAR(32) NOT NULL COMMENT '日志配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称',
  `configDesc` VARCHAR(200) DEFAULT NULL COMMENT '配置描述',
  
  -- 日志内容控制
  `logFormat` VARCHAR(50) NOT NULL DEFAULT 'JSON' COMMENT '日志格式(JSON,TEXT,CSV)',
  `recordRequestBody` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否记录请求体(N否,Y是)',
  `recordResponseBody` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否记录响应体(N否,Y是)',
  `recordHeaders` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否记录请求/响应头(N否,Y是)',
  `maxBodySizeBytes` INT NOT NULL DEFAULT 4096 COMMENT '最大记录报文大小(字节)',
  
  -- 日志输出目标配置
  `outputTargets` VARCHAR(200) NOT NULL DEFAULT 'CONSOLE' COMMENT '输出目标,逗号分隔(CONSOLE,FILE,DATABASE,MONGODB,ELASTICSEARCH)',
  `fileConfig` TEXT DEFAULT NULL COMMENT '文件输出配置,JSON格式',
  `databaseConfig` TEXT DEFAULT NULL COMMENT '数据库输出配置,JSON格式',
  `mongoConfig` TEXT DEFAULT NULL COMMENT 'MongoDB输出配置,JSON格式',
  `elasticsearchConfig` TEXT DEFAULT NULL COMMENT 'Elasticsearch输出配置,JSON格式',
  `clickhouseConfig` TEXT DEFAULT NULL COMMENT 'clickhouseConfig输出配置,JSON格式',
  
  -- 异步和批量处理配置
  `enableAsyncLogging` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用异步日志(N否,Y是)',
  `asyncQueueSize` INT NOT NULL DEFAULT 10000 COMMENT '异步队列大小',
  `asyncFlushIntervalMs` INT NOT NULL DEFAULT 1000 COMMENT '异步刷新间隔(毫秒)',
  `enableBatchProcessing` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用批量处理(N否,Y是)',
  `batchSize` INT NOT NULL DEFAULT 100 COMMENT '批处理大小',
  `batchTimeoutMs` INT NOT NULL DEFAULT 5000 COMMENT '批处理超时时间(毫秒)',
  
  -- 日志保留和轮转配置
  `logRetentionDays` INT NOT NULL DEFAULT 30 COMMENT '日志保留天数',
  `enableFileRotation` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用文件轮转(N否,Y是)',
  `maxFileSizeMB` INT DEFAULT 100 COMMENT '最大文件大小(MB)',
  `maxFileCount` INT DEFAULT 10 COMMENT '最大文件数量',
  `rotationPattern` VARCHAR(100) DEFAULT 'DAILY' COMMENT '轮转模式(HOURLY,DAILY,WEEKLY,SIZE_BASED)',
  
  -- 敏感数据处理
  `enableSensitiveDataMasking` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用敏感数据脱敏(N否,Y是)',
  `sensitiveFields` TEXT DEFAULT NULL COMMENT '敏感字段列表,JSON数组格式',
  `maskingPattern` VARCHAR(100) DEFAULT '***' COMMENT '脱敏替换模式',
  
  -- 性能优化配置
  `bufferSize` INT NOT NULL DEFAULT 8192 COMMENT '缓冲区大小(字节)',
  `flushThreshold` INT NOT NULL DEFAULT 100 COMMENT '刷新阈值(条目数)',
  
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `logConfigId`),
  INDEX `idx_HUB_GW_LOG_CONFIG_name` (`configName`),
  INDEX `idx_HUB_GW_LOG_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志配置表 - 存储网关日志相关配置';
