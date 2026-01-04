CREATE TABLE `HUB_REGISTRY_SERVICE_INSTANCE` (
  -- 主键和租户信息
  `serviceInstanceId` VARCHAR(100) NOT NULL COMMENT '服务实例ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联服务和分组（主键关联）
  `serviceGroupId` VARCHAR(32) NOT NULL COMMENT '服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键',
  -- 冗余字段（便于查询和展示）
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称，冗余字段便于查询',
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称，冗余字段便于查询',
  
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
  `tempInstanceFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '临时实例标记(Y是临时实例,N否)',
  
  -- 健康检查统计
  `heartbeatFailCount` INT NOT NULL DEFAULT 0 COMMENT '心跳检查失败次数，仅用于计数',
  
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
  KEY `IDX_REGISTRY_INSTANCE` (`tenantId`, `serviceGroupId`, `serviceName`, `hostAddress`, `portNumber`),
  -- 主键关联索引（用于外键关联查询）
  KEY `IDX_REGISTRY_INST_GROUP_ID` (`tenantId`, `serviceGroupId`),
  -- 冗余字段索引（用于业务查询和展示）
  KEY `IDX_REGISTRY_INST_SVC_NAME` (`serviceName`),
  KEY `IDX_REGISTRY_INST_GROUP_NAME` (`groupName`),
  -- 业务状态索引
  KEY `IDX_REGISTRY_INST_STATUS` (`instanceStatus`),
  KEY `IDX_REGISTRY_INST_HEALTH` (`healthStatus`),
  KEY `IDX_REGISTRY_INST_HEARTBEAT` (`lastHeartbeatTime`),
  KEY `IDX_REGISTRY_INST_HOST_PORT` (`hostAddress`, `portNumber`),
  KEY `IDX_REGISTRY_INST_CLIENT` (`clientId`),
  KEY `IDX_REGISTRY_INST_ACTIVE` (`activeFlag`),
  KEY `IDX_REGISTRY_INST_TEMP` (`tempInstanceFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务实例表 - 存储具体的服务实例信息';
