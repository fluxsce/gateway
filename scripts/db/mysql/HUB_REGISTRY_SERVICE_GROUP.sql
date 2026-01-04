CREATE TABLE `HUB_REGISTRY_SERVICE_GROUP` (
  -- 主键和租户信息
  `serviceGroupId` VARCHAR(32) NOT NULL COMMENT '服务分组ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 分组基本信息
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称',
  `groupDescription` VARCHAR(500) DEFAULT NULL COMMENT '分组描述',
  `groupType` VARCHAR(50) DEFAULT 'BUSINESS' COMMENT '分组类型(BUSINESS,SYSTEM,TEST)',
  
  -- 授权信息
  `ownerUserId` VARCHAR(32) NOT NULL COMMENT '分组所有者用户ID',
  `adminUserIds` TEXT DEFAULT NULL COMMENT '管理员用户ID列表，JSON格式',
  `readUserIds` TEXT DEFAULT NULL COMMENT '只读用户ID列表，JSON格式',
  `accessControlEnabled` VARCHAR(1) DEFAULT 'N' COMMENT '是否启用访问控制(N否,Y是)',
  
  -- 配置信息
  `defaultProtocolType` VARCHAR(20) DEFAULT 'HTTP' COMMENT '默认协议类型',
  `defaultLoadBalanceStrategy` VARCHAR(50) DEFAULT 'ROUND_ROBIN' COMMENT '默认负载均衡策略',
  `defaultHealthCheckUrl` VARCHAR(500) DEFAULT '/health' COMMENT '默认健康检查URL',
  `defaultHealthCheckIntervalSeconds` INT DEFAULT 30 COMMENT '默认健康检查间隔(秒)',
  
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
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `serviceGroupId`),
  KEY `IDX_REGISTRY_GROUP_NAME` (`tenantId`, `groupName`),
  KEY `IDX_REGISTRY_GROUP_TYPE` (`groupType`),
  KEY `IDX_REGISTRY_GROUP_OWNER` (`ownerUserId`),
  KEY `IDX_REGISTRY_GROUP_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务分组表 - 存储服务分组和授权信息';
