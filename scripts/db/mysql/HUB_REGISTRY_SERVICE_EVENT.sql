CREATE TABLE `HUB_REGISTRY_SERVICE_EVENT` (
  -- 主键和租户信息
  `serviceEventId` VARCHAR(32) NOT NULL COMMENT '服务事件ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联主键字段（用于精确关联到对应表记录）
  `serviceGroupId` VARCHAR(32) DEFAULT NULL COMMENT '服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键',
  `serviceInstanceId` VARCHAR(100) DEFAULT NULL COMMENT '服务实例ID，关联HUB_REGISTRY_SERVICE_INSTANCE表主键',
  
  -- 事件基本信息（冗余字段，便于查询和展示）
  `groupName` VARCHAR(100) DEFAULT NULL COMMENT '分组名称，冗余字段便于查询',
  `serviceName` VARCHAR(100) DEFAULT NULL COMMENT '服务名称，冗余字段便于查询',
  `hostAddress` VARCHAR(100) DEFAULT NULL COMMENT '主机地址，冗余字段便于查询',
  `portNumber` INT DEFAULT NULL COMMENT '端口号，冗余字段便于查询',
  `nodeIpAddress` VARCHAR(100) DEFAULT NULL COMMENT '节点IP地址，记录程序运行的IP',
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
  PRIMARY KEY (`tenantId`, `serviceEventId`),
  -- 主键关联索引（用于精确关联查询）
  KEY `IDX_REGISTRY_EVENT_GROUP_ID` (`tenantId`, `serviceGroupId`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_INSTANCE_ID` (`tenantId`, `serviceInstanceId`, `eventTime`),
  -- 冗余字段索引（用于业务查询和展示）
  KEY `IDX_REGISTRY_EVENT_GROUP_NAME` (`tenantId`, `groupName`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_SVC_NAME` (`tenantId`, `serviceName`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_HOST` (`tenantId`, `hostAddress`, `portNumber`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_NODE_IP` (`tenantId`, `nodeIpAddress`, `eventTime`),
  -- 事件类型和时间索引
  KEY `IDX_REGISTRY_EVENT_TYPE` (`eventType`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_TIME` (`eventTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务事件日志表 - 记录服务注册发现相关的所有事件';

