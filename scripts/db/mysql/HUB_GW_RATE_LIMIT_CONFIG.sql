CREATE TABLE `HUB_GW_RATE_LIMIT_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `rateLimitConfigId` VARCHAR(32) NOT NULL COMMENT '限流配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级限流)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级限流)',
  `limitName` VARCHAR(100) NOT NULL COMMENT '限流规则名称',
  
  -- 修改：统一算法标识格式
  `algorithm` VARCHAR(50) NOT NULL DEFAULT 'token-bucket' COMMENT '限流算法(token-bucket,leaky-bucket,sliding-window,fixed-window,none)',
  
  -- 修改：限流键策略（替代原limitType和keyExpression）
  `keyStrategy` VARCHAR(50) NOT NULL DEFAULT 'ip' COMMENT '限流键策略(ip,user,path,service,route)',
  
  -- 保持原有字段但调整默认值
  `limitRate` INT NOT NULL COMMENT '限流速率(次/秒)',
  `burstCapacity` INT NOT NULL DEFAULT 0 COMMENT '突发容量',
  `timeWindowSeconds` INT NOT NULL DEFAULT 1 COMMENT '时间窗口(秒)',
  `rejectionStatusCode` INT NOT NULL DEFAULT 429 COMMENT '拒绝时的HTTP状态码',
  `rejectionMessage` VARCHAR(200) DEFAULT '请求过于频繁，请稍后再试' COMMENT '拒绝时的提示消息',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `customConfig` TEXT DEFAULT NULL COMMENT '自定义配置,JSON格式',
  
  -- 保留现有的标准字段
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
  
  PRIMARY KEY (`tenantId`, `rateLimitConfigId`),
  INDEX `IDX_GW_RATE_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_RATE_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_RATE_STRATEGY` (`keyStrategy`),
  INDEX `IDX_GW_RATE_ALGORITHM` (`algorithm`),
  INDEX `IDX_GW_RATE_PRIORITY` (`configPriority`),
  INDEX `IDX_GW_RATE_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限流配置表 - 存储流量限制规则';

