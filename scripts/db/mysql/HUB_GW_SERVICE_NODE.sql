CREATE TABLE `HUB_GW_SERVICE_NODE` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `serviceNodeId` VARCHAR(32) NOT NULL COMMENT '服务节点ID',
  `serviceDefinitionId` VARCHAR(32) NOT NULL COMMENT '关联的服务定义ID',
  `nodeId` VARCHAR(100) NOT NULL COMMENT '节点标识ID',
  -- 根据NodeConfig.URL字段设计,分解为host+port+protocol便于查询和管理
  `nodeUrl` VARCHAR(500) NOT NULL COMMENT '节点完整URL(来自NodeConfig.URL)',
  `nodeHost` VARCHAR(100) NOT NULL COMMENT '节点主机地址(从URL解析)',
  `nodePort` INT NOT NULL COMMENT '节点端口(从URL解析)',
  `nodeProtocol` VARCHAR(10) NOT NULL DEFAULT 'HTTP' COMMENT '节点协议(HTTP,HTTPS,从URL解析)',
  
  -- 根据NodeConfig.Weight字段设计
  `nodeWeight` INT NOT NULL DEFAULT 100 COMMENT '节点权重(来自NodeConfig.Weight)',
  
  -- 根据NodeConfig.Health字段设计
  `healthStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '健康状态(N不健康,Y健康,来自NodeConfig.Health)',

  -- 根据NodeConfig.Metadata字段设计
  `nodeMetadata` TEXT DEFAULT NULL COMMENT '节点元数据,JSON格式(来自NodeConfig.Metadata)',
  
  -- 运行时状态字段(非NodeConfig结构,但运维需要)
  `nodeStatus` INT NOT NULL DEFAULT 1 COMMENT '节点运行状态(0下线,1在线,2维护)',
  `lastHealthCheckTime` DATETIME DEFAULT NULL COMMENT '最后健康检查时间',
  `healthCheckResult` TEXT DEFAULT NULL COMMENT '健康检查结果详情',
  
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
  PRIMARY KEY (`tenantId`, `serviceNodeId`),
  INDEX `IDX_GW_NODE_SERVICE` (`serviceDefinitionId`),
  INDEX `IDX_GW_NODE_ENDPOINT` (`nodeHost`, `nodePort`),
  INDEX `IDX_GW_NODE_HEALTH` (`healthStatus`),
  INDEX `IDX_GW_NODE_STATUS` (`nodeStatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务节点表 - 根据NodeConfig结构设计,存储服务节点实例信息';

