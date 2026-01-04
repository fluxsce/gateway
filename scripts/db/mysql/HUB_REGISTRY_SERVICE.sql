CREATE TABLE `HUB_REGISTRY_SERVICE` (
  -- 主键和租户信息
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称，主键',
  
  -- 关联分组（主键关联）
  `serviceGroupId` VARCHAR(32) NOT NULL COMMENT '服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键',
  -- 冗余字段（便于查询和展示）
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称，冗余字段便于查询',
  
  -- 服务基本信息
  `serviceDescription` VARCHAR(500) DEFAULT NULL COMMENT '服务描述',
  
  -- 注册管理配置
  `registryType` VARCHAR(20) NOT NULL DEFAULT 'INTERNAL' COMMENT '注册类型(INTERNAL:内部管理,NACOS:Nacos注册中心,CONSUL:Consul,EUREKA:Eureka,ETCD:ETCD,ZOOKEEPER:ZooKeeper)',
  `externalRegistryConfig` TEXT DEFAULT NULL COMMENT '外部注册中心配置，JSON格式，仅当registryType非INTERNAL时使用',
  
  -- 服务配置
  `protocolType` VARCHAR(20) DEFAULT 'HTTP' COMMENT '协议类型(HTTP,HTTPS,TCP,UDP,GRPC)',
  `contextPath` VARCHAR(200) DEFAULT '' COMMENT '上下文路径',
  `loadBalanceStrategy` VARCHAR(50) DEFAULT 'ROUND_ROBIN' COMMENT '负载均衡策略',
  
  -- 健康检查配置
  `healthCheckUrl` VARCHAR(500) DEFAULT '/health' COMMENT '健康检查URL',
  `healthCheckIntervalSeconds` INT DEFAULT 30 COMMENT '健康检查间隔(秒)',
  `healthCheckTimeoutSeconds` INT DEFAULT 5 COMMENT '健康检查超时(秒)',
  `healthCheckType` VARCHAR(20) DEFAULT 'HTTP' COMMENT '健康检查类型(HTTP,TCP)',
  `healthCheckMode` VARCHAR(20) DEFAULT 'ACTIVE' COMMENT '健康检查模式(ACTIVE:主动探测,PASSIVE:客户端上报)',
  
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
  -- 主键关联索引（用于外键关联查询）
  KEY `IDX_REGISTRY_SVC_GROUP_ID` (`tenantId`, `serviceGroupId`),
  -- 冗余字段索引（用于业务查询和展示）
  KEY `IDX_REGISTRY_SVC_GROUP_NAME` (`groupName`),
  KEY `IDX_REGISTRY_SVC_REGISTRY_TYPE` (`registryType`),
  KEY `IDX_REGISTRY_SVC_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务表 - 存储服务的基本信息和配置，支持内部管理和外部注册中心代理模式';
