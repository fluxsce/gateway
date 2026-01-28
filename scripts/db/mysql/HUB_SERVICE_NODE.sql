-- 服务节点表 - 存储服务节点的详细信息，包括网络地址、健康状态等
CREATE TABLE `HUB_SERVICE_NODE` (
  -- 主键和租户信息
  `nodeId` VARCHAR(32) NOT NULL COMMENT '节点ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联服务（通过联合主键关联HUB_SERVICE表）
  `namespaceId` VARCHAR(32) NOT NULL COMMENT '命名空间ID，关联HUB_SERVICE表',
  `groupName` VARCHAR(64) NOT NULL COMMENT '分组名称，关联HUB_SERVICE表',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称，关联HUB_SERVICE表',
  
  -- 网络连接信息
  `ipAddress` VARCHAR(50) NOT NULL COMMENT 'IP地址',
  `portNumber` INT NOT NULL COMMENT '端口号',
  
  -- 节点状态信息
  `instanceStatus` VARCHAR(20) NOT NULL DEFAULT 'UP' COMMENT '节点状态(UP:运行中,DOWN:已停止,STARTING:启动中,OUT_OF_SERVICE:暂停服务)',
  `healthyStatus` VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN' COMMENT '健康状态(HEALTHY:健康,UNHEALTHY:不健康,UNKNOWN:未知)',
  `ephemeral` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否临时节点(Y:临时节点,N:持久节点)',
  
  -- 负载均衡配置
  `weight` DECIMAL(6,2) NOT NULL DEFAULT 1.00 COMMENT '权重值，范围0.01-10000.00，用于负载均衡',
  
  -- 节点元数据
  `metadataJson` TEXT DEFAULT NULL COMMENT '节点元数据，JSON格式，存储节点的扩展信息',
  
  -- 时间戳信息
  `registerTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  `lastBeatTime` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  `lastCheckTime` DATETIME DEFAULT NULL COMMENT '最后健康检查时间',
  
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
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `nodeId`),
  UNIQUE KEY `IDX_SVC_NODE_UNIQUE` (`tenantId`, `namespaceId`, `groupName`, `serviceName`, `ipAddress`, `portNumber`),
  KEY `IDX_SVC_NODE_SERVICE` (`tenantId`, `namespaceId`, `groupName`, `serviceName`),
  KEY `IDX_SVC_NODE_SERVICE_NAME` (`tenantId`, `serviceName`),
  KEY `IDX_SVC_NODE_NS_ID` (`tenantId`, `namespaceId`),
  KEY `IDX_SVC_NODE_GROUP_NAME` (`tenantId`, `namespaceId`, `groupName`),
  KEY `IDX_SVC_NODE_IP_PORT` (`ipAddress`, `portNumber`),
  KEY `IDX_SVC_NODE_STATUS` (`instanceStatus`),
  KEY `IDX_SVC_NODE_HEALTHY` (`healthyStatus`),
  KEY `IDX_SVC_NODE_EPHEMERAL` (`ephemeral`),
  KEY `IDX_SVC_NODE_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务节点表 - 存储服务节点的详细信息，包括网络地址、健康状态、权重等';


