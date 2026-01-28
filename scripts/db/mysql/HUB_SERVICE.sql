-- 服务表 - 存储服务的基本信息和元数据
CREATE TABLE `HUB_SERVICE` (
  -- 主键和租户信息
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  `namespaceId` VARCHAR(32) NOT NULL COMMENT '命名空间ID，关联HUB_SERVICE_NAMESPACE表',
  `groupName` VARCHAR(64) NOT NULL DEFAULT 'DEFAULT_GROUP' COMMENT '分组名称，如DEFAULT_GROUP',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称，全局唯一标识',
  
  -- 服务类型
  `serviceType` VARCHAR(50) NOT NULL DEFAULT 'INTERNAL' COMMENT '服务类型(INTERNAL:内部服务,NACOS:Nacos注册中心,CONSUL:Consul,EUREKA:Eureka,ETCD:ETCD,ZOOKEEPER:ZooKeeper)',
  
  -- 服务基本信息
  `serviceVersion` VARCHAR(50) DEFAULT NULL COMMENT '服务版本号',
  `serviceDescription` VARCHAR(500) DEFAULT NULL COMMENT '服务描述',
  
  -- 外部服务配置（仅当serviceType非INTERNAL时使用）
  `externalServiceConfig` TEXT DEFAULT NULL COMMENT '外部服务配置，JSON格式，存储外部注册中心的连接配置等信息',
  
  -- 服务元数据
  `metadataJson` TEXT DEFAULT NULL COMMENT '服务元数据，JSON格式，存储服务的扩展信息',
  `tagsJson` TEXT DEFAULT NULL COMMENT '服务标签，JSON格式，用于服务分类和过滤',
  
  -- 服务保护阈值（0-1之间的小数，表示健康实例比例低于该值时触发保护）
  `protectThreshold` DECIMAL(3,2) DEFAULT 0.00 COMMENT '服务保护阈值，范围0.00-1.00',
  
  -- 服务选择器（用于服务路由）
  `selectorJson` TEXT DEFAULT NULL COMMENT '服务选择器，JSON格式，用于服务路由规则',
  
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
  
  -- 预留字段
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
  PRIMARY KEY (`tenantId`, `namespaceId`, `groupName`, `serviceName`),
  KEY `IDX_SVC_NS_ID` (`tenantId`, `namespaceId`),
  KEY `IDX_SVC_NAME` (`serviceName`),
  KEY `IDX_SVC_TYPE` (`serviceType`),
  KEY `IDX_SVC_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务表 - 存储服务的基本信息和元数据，支持多命名空间和多分组管理';

