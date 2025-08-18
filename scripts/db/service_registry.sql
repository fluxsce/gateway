-- =====================================================
-- 服务注册中心数据库表结构设计
-- 遵循 docs/database/naming-convention.md 规范
-- 参考Nacos设计：服务分组 -> 服务名称 -> 服务实例
-- =====================================================

-- =====================================================
-- 服务分组表 - 存储服务分组和授权信息
-- =====================================================
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
  UNIQUE KEY `UK_REGISTRY_GROUP_NAME` (`tenantId`, `groupName`),
  KEY `IDX_REGISTRY_GROUP_TYPE` (`groupType`),
  KEY `IDX_REGISTRY_GROUP_OWNER` (`ownerUserId`),
  KEY `IDX_REGISTRY_GROUP_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务分组表 - 存储服务分组和授权信息';

-- =====================================================
-- 服务表 - 存储服务基本信息
-- =====================================================
CREATE TABLE `HUB_REGISTRY_SERVICE` (
  -- 主键和租户信息
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称，主键',
  
  -- 关联分组
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称',
  
  -- 服务基本信息
  `serviceDescription` VARCHAR(500) DEFAULT NULL COMMENT '服务描述',
  
  -- 服务配置
  `protocolType` VARCHAR(20) DEFAULT 'HTTP' COMMENT '协议类型(HTTP,HTTPS,TCP,UDP,GRPC)',
  `contextPath` VARCHAR(200) DEFAULT '' COMMENT '上下文路径',
  `loadBalanceStrategy` VARCHAR(50) DEFAULT 'ROUND_ROBIN' COMMENT '负载均衡策略',
  
  -- 健康检查配置
  `healthCheckUrl` VARCHAR(500) DEFAULT '/health' COMMENT '健康检查URL',
  `healthCheckIntervalSeconds` INT DEFAULT 30 COMMENT '健康检查间隔(秒)',
  `healthCheckTimeoutSeconds` INT DEFAULT 5 COMMENT '健康检查超时(秒)',
  
  -- 元数据和标签
  `metadataJson` TEXT DEFAULT NULL COMMENT '服务元数据，JSON格式',
  `tagsJson` TEXT DEFAULT NULL COMMENT '服务标签，JSON格式',
  
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
  PRIMARY KEY (`tenantId`, `serviceName`),
  KEY `IDX_REGISTRY_SVC_GROUP` (`groupName`),
  KEY `IDX_REGISTRY_SVC_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务表 - 存储服务的基本信息和配置';

-- =====================================================
-- 服务实例表 - 存储具体的服务实例
-- =====================================================
CREATE TABLE `HUB_REGISTRY_SERVICE_INSTANCE` (
  -- 主键和租户信息
  `serviceInstanceId` VARCHAR(32) NOT NULL COMMENT '服务实例ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联服务
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称',
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称',
  
  -- 网络连接信息
  `hostAddress` VARCHAR(100) NOT NULL COMMENT '主机地址',
  `portNumber` INT NOT NULL COMMENT '端口号',
  `contextPath` VARCHAR(200) DEFAULT '' COMMENT '上下文路径',
  
  -- 实例状态信息
  `instanceStatus` VARCHAR(20) NOT NULL DEFAULT 'UP' COMMENT '实例状态(UP,DOWN,STARTING,OUT_OF_SERVICE)',
  `healthStatus` VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN' COMMENT '健康状态(HEALTHY,UNHEALTHY,UNKNOWN)',
  
  -- 负载均衡配置
  `weightValue` INT NOT NULL DEFAULT 100 COMMENT '权重值',
  
  -- 客户端信息
  `clientId` VARCHAR(100) DEFAULT NULL COMMENT '客户端ID',
  `clientVersion` VARCHAR(50) DEFAULT NULL COMMENT '客户端版本',
  `clientType` VARCHAR(50) DEFAULT 'SERVICE' COMMENT '客户端类型(SERVICE,GATEWAY,ADMIN)',
  
  -- 元数据和标签
  `metadataJson` TEXT DEFAULT NULL COMMENT '实例元数据，JSON格式',
  `tagsJson` TEXT DEFAULT NULL COMMENT '实例标签，JSON格式',
  
  -- 时间戳信息
  `registerTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  `lastHeartbeatTime` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  `lastHealthCheckTime` DATETIME DEFAULT NULL COMMENT '最后健康检查时间',
  
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
  PRIMARY KEY (`tenantId`, `serviceInstanceId`),
  UNIQUE KEY `UK_REGISTRY_INSTANCE` (`tenantId`, `groupName`, `serviceName`, `hostAddress`, `portNumber`),
  KEY `IDX_REGISTRY_INST_SVC_NAME` (`serviceName`),
  KEY `IDX_REGISTRY_INST_GROUP` (`groupName`),
  KEY `IDX_REGISTRY_INST_STATUS` (`instanceStatus`),
  KEY `IDX_REGISTRY_INST_HEALTH` (`healthStatus`),
  KEY `IDX_REGISTRY_INST_HEARTBEAT` (`lastHeartbeatTime`),
  KEY `IDX_REGISTRY_INST_HOST_PORT` (`hostAddress`, `portNumber`),
  KEY `IDX_REGISTRY_INST_CLIENT` (`clientId`),
  KEY `IDX_REGISTRY_INST_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务实例表 - 存储具体的服务实例信息';

-- =====================================================
-- 服务事件日志表 - 记录服务变更事件
-- =====================================================
CREATE TABLE `HUB_REGISTRY_SERVICE_EVENT` (
  -- 主键和租户信息
  `serviceEventId` BIGINT AUTO_INCREMENT COMMENT '服务事件ID，自增主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 事件基本信息
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称',
  `hostAddress` VARCHAR(100) DEFAULT NULL COMMENT '主机地址',
  `portNumber` INT DEFAULT NULL COMMENT '端口号',
  `eventType` VARCHAR(50) NOT NULL COMMENT '事件类型(GROUP_CREATE,GROUP_UPDATE,GROUP_DELETE,SERVICE_CREATE,SERVICE_UPDATE,SERVICE_DELETE,INSTANCE_REGISTER,INSTANCE_DEREGISTER,INSTANCE_HEARTBEAT,INSTANCE_HEALTH_CHANGE,INSTANCE_STATUS_CHANGE)',
  `eventSource` VARCHAR(100) DEFAULT NULL COMMENT '事件来源',
  
  -- 事件数据
  `eventDataJson` TEXT DEFAULT NULL COMMENT '事件数据，JSON格式',
  `eventMessage` VARCHAR(1000) DEFAULT NULL COMMENT '事件消息描述',
  
  -- 时间信息
  `eventTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '事件发生时间',
  
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
  PRIMARY KEY (`serviceEventId`),
  KEY `IDX_REGISTRY_EVENT_GROUP` (`tenantId`, `groupName`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_SVC` (`tenantId`, `serviceName`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_HOST` (`tenantId`, `hostAddress`, `portNumber`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_TYPE` (`eventType`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_TIME` (`eventTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务事件日志表 - 记录服务注册发现相关的所有事件';

-- =====================================================
-- 外部注册中心配置表 - 存储外部注册中心连接配置
-- =====================================================
CREATE TABLE `HUB_REGISTRY_EXTERNAL_CONFIG` (
  -- 主键和租户信息
  `externalConfigId` VARCHAR(32) NOT NULL COMMENT '外部配置ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 配置基本信息
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称',
  `configDescription` VARCHAR(500) DEFAULT NULL COMMENT '配置描述',
  `registryType` VARCHAR(50) NOT NULL COMMENT '注册中心类型(CONSUL,NACOS,ETCD,EUREKA,ZOOKEEPER)',
  `environmentName` VARCHAR(50) NOT NULL DEFAULT 'default' COMMENT '环境名称(dev,test,prod,default)',
  
  -- 连接配置
  `serverAddress` VARCHAR(500) NOT NULL COMMENT '服务器地址，多个地址用逗号分隔',
  `serverPort` INT DEFAULT NULL COMMENT '服务器端口',
  `serverPath` VARCHAR(200) DEFAULT NULL COMMENT '服务器路径',
  `serverScheme` VARCHAR(10) DEFAULT 'http' COMMENT '连接协议(http,https)',
  
  -- 认证配置
  `authEnabled` VARCHAR(1) DEFAULT 'N' COMMENT '是否启用认证(N否,Y是)',
  `username` VARCHAR(100) DEFAULT NULL COMMENT '用户名',
  `password` VARCHAR(200) DEFAULT NULL COMMENT '密码',
  `accessToken` VARCHAR(500) DEFAULT NULL COMMENT '访问令牌',
  `secretKey` VARCHAR(200) DEFAULT NULL COMMENT '密钥',
  
  -- 连接配置
  `connectionTimeout` INT DEFAULT 5000 COMMENT '连接超时时间(毫秒)',
  `readTimeout` INT DEFAULT 10000 COMMENT '读取超时时间(毫秒)',
  `maxRetries` INT DEFAULT 3 COMMENT '最大重试次数',
  `retryInterval` INT DEFAULT 1000 COMMENT '重试间隔(毫秒)',
  
  -- 特定配置
  `specificConfig` TEXT DEFAULT NULL COMMENT '特定注册中心配置，JSON格式',
  `fieldMapping` TEXT DEFAULT NULL COMMENT '字段映射配置，JSON格式',
  
  -- 故障转移配置
  `failoverEnabled` VARCHAR(1) DEFAULT 'N' COMMENT '是否启用故障转移(N否,Y是)',
  `failoverConfigId` VARCHAR(32) DEFAULT NULL COMMENT '故障转移配置ID',
  `failoverStrategy` VARCHAR(50) DEFAULT 'MANUAL' COMMENT '故障转移策略(MANUAL,AUTO)',
  
  -- 数据同步配置
  `syncEnabled` VARCHAR(1) DEFAULT 'N' COMMENT '是否启用数据同步(N否,Y是)',
  `syncInterval` INT DEFAULT 30 COMMENT '同步间隔(秒)',
  `conflictResolution` VARCHAR(50) DEFAULT 'primary_wins' COMMENT '冲突解决策略(primary_wins,secondary_wins,merge)',
  
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
  PRIMARY KEY (`tenantId`, `externalConfigId`),
  UNIQUE KEY `UK_REGISTRY_EXT_CONFIG_NAME` (`tenantId`, `configName`, `environmentName`),
  KEY `IDX_REGISTRY_EXT_CONFIG_TYPE` (`registryType`),
  KEY `IDX_REGISTRY_EXT_CONFIG_ENV` (`environmentName`),
  KEY `IDX_REGISTRY_EXT_CONFIG_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部注册中心配置表 - 存储外部注册中心的连接和配置信息';

-- =====================================================
-- 外部注册中心状态表 - 存储外部注册中心运行状态
-- =====================================================
CREATE TABLE `HUB_REGISTRY_EXTERNAL_STATUS` (
  -- 主键和租户信息
  `externalStatusId` VARCHAR(32) NOT NULL COMMENT '外部状态ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  `externalConfigId` VARCHAR(32) NOT NULL COMMENT '外部配置ID',
  
  -- 连接状态
  `connectionStatus` VARCHAR(20) NOT NULL DEFAULT 'DISCONNECTED' COMMENT '连接状态(CONNECTED,DISCONNECTED,CONNECTING,ERROR)',
  `healthStatus` VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN' COMMENT '健康状态(HEALTHY,UNHEALTHY,UNKNOWN)',
  `lastConnectTime` DATETIME DEFAULT NULL COMMENT '最后连接时间',
  `lastDisconnectTime` DATETIME DEFAULT NULL COMMENT '最后断开时间',
  `lastHealthCheckTime` DATETIME DEFAULT NULL COMMENT '最后健康检查时间',
  
  -- 性能指标
  `responseTime` INT DEFAULT 0 COMMENT '响应时间(毫秒)',
  `successCount` BIGINT DEFAULT 0 COMMENT '成功次数',
  `errorCount` BIGINT DEFAULT 0 COMMENT '错误次数',
  `timeoutCount` BIGINT DEFAULT 0 COMMENT '超时次数',
  
  -- 故障转移状态
  `failoverStatus` VARCHAR(20) DEFAULT 'NORMAL' COMMENT '故障转移状态(NORMAL,FAILOVER,RECOVERING)',
  `failoverTime` DATETIME DEFAULT NULL COMMENT '故障转移时间',
  `failoverCount` INT DEFAULT 0 COMMENT '故障转移次数',
  `recoverTime` DATETIME DEFAULT NULL COMMENT '恢复时间',
  
  -- 同步状态
  `syncStatus` VARCHAR(20) DEFAULT 'IDLE' COMMENT '同步状态(IDLE,SYNCING,ERROR)',
  `lastSyncTime` DATETIME DEFAULT NULL COMMENT '最后同步时间',
  `syncSuccessCount` BIGINT DEFAULT 0 COMMENT '同步成功次数',
  `syncErrorCount` BIGINT DEFAULT 0 COMMENT '同步错误次数',
  
  -- 错误信息
  `lastErrorMessage` VARCHAR(1000) DEFAULT NULL COMMENT '最后错误消息',
  `lastErrorTime` DATETIME DEFAULT NULL COMMENT '最后错误时间',
  `errorDetails` TEXT DEFAULT NULL COMMENT '错误详情，JSON格式',
  
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
  PRIMARY KEY (`tenantId`, `externalStatusId`),
  UNIQUE KEY `UK_REGISTRY_EXT_STATUS_CONFIG` (`tenantId`, `externalConfigId`),
  KEY `IDX_REGISTRY_EXT_STATUS_CONN` (`connectionStatus`),
  KEY `IDX_REGISTRY_EXT_STATUS_HEALTH` (`healthStatus`),
  KEY `IDX_REGISTRY_EXT_STATUS_FAILOVER` (`failoverStatus`),
  KEY `IDX_REGISTRY_EXT_STATUS_SYNC` (`syncStatus`),
  KEY `IDX_REGISTRY_EXT_STATUS_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部注册中心状态表 - 存储外部注册中心的实时运行状态和性能指标';
