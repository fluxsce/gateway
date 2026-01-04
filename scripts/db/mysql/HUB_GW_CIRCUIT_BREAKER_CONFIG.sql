CREATE TABLE `HUB_GW_CIRCUIT_BREAKER_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `circuitBreakerConfigId` VARCHAR(32) NOT NULL COMMENT '熔断配置ID',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级熔断)',
  `targetServiceId` VARCHAR(32) DEFAULT NULL COMMENT '目标服务ID(服务级熔断)',
  `breakerName` VARCHAR(100) NOT NULL COMMENT '熔断器名称',
  
  -- 根据CircuitBreakerConfig结构设计基础配置
  `keyStrategy` VARCHAR(50) NOT NULL DEFAULT 'api' COMMENT '熔断Key策略(ip,service,api等)',
  
  -- 阈值配置
  `errorRatePercent` INT NOT NULL DEFAULT 50 COMMENT '错误率阈值(百分比)',
  `minimumRequests` INT NOT NULL DEFAULT 10 COMMENT '最小请求数阈值',
  `halfOpenMaxRequests` INT NOT NULL DEFAULT 3 COMMENT '半开状态最大请求数',
  `slowCallThreshold` INT NOT NULL DEFAULT 1000 COMMENT '慢调用阈值(毫秒)',
  `slowCallRatePercent` INT NOT NULL DEFAULT 50 COMMENT '慢调用率阈值(百分比)',
  
  -- 时间配置
  `openTimeoutSeconds` INT NOT NULL DEFAULT 60 COMMENT '熔断器打开持续时间(秒)',
  `windowSizeSeconds` INT NOT NULL DEFAULT 60 COMMENT '统计窗口大小(秒)',
  
  -- 错误处理配置
  `errorStatusCode` INT NOT NULL DEFAULT 503 COMMENT '熔断时返回的HTTP状态码',
  `errorMessage` VARCHAR(500) DEFAULT 'Service temporarily unavailable due to circuit breaker' COMMENT '熔断时返回的错误信息',
  
  -- 存储配置
  `storageType` VARCHAR(50) NOT NULL DEFAULT 'memory' COMMENT '存储类型(memory,redis)',
  `storageConfig` TEXT DEFAULT NULL COMMENT '存储配置,JSON格式',
  
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
  PRIMARY KEY (`tenantId`, `circuitBreakerConfigId`),
  INDEX `IDX_GW_CB_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_CB_SERVICE` (`targetServiceId`),
  INDEX `IDX_GW_CB_STRATEGY` (`keyStrategy`),
  INDEX `IDX_GW_CB_STORAGE` (`storageType`),
  INDEX `IDX_GW_CB_PRIORITY` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='熔断配置表 - 根据CircuitBreakerConfig结构设计,支持完整的熔断策略配置';

