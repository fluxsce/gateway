CREATE TABLE `HUB_CLUSTER_EVENT` (
  -- 主键和租户信息
  `eventId` VARCHAR(64) NOT NULL COMMENT '事件ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  
  -- 事件来源(发布者)
  `sourceNodeId` VARCHAR(100) NOT NULL COMMENT '发布节点ID(hostname:port)',
  `sourceNodeIp` VARCHAR(100) DEFAULT NULL COMMENT '发布节点IP',
  
  -- 事件信息
  `eventType` VARCHAR(50) NOT NULL COMMENT '事件类型(ROUTE_CONFIG/SERVICE_CONFIG/FILTER_CONFIG/CACHE_REFRESH等)',
  `eventAction` VARCHAR(50) NOT NULL COMMENT '事件动作(CREATE/UPDATE/DELETE/REFRESH/INVALIDATE)',
  `eventPayload` TEXT DEFAULT NULL COMMENT '事件数据(JSON格式，包含所有业务信息)',
  
  -- 事件时间
  `eventTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '事件发生时间',
  `expireTime` DATETIME DEFAULT NULL COMMENT '事件过期时间',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(64) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(64) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(64) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` TEXT DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性(JSON格式)',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `eventId`),
  -- 优化后的复合索引：支持高效的待处理事件查询
  -- 查询模式: WHERE tenantId=? AND activeFlag='Y' AND eventTime>? AND sourceNodeId!=?
  KEY `IDX_CLS_EVT_QUERY` (`tenantId`, `activeFlag`, `eventTime`, `eventId`),
  -- 辅助索引：用于特定场景查询
  KEY `IDX_CLS_EVT_SOURCE` (`sourceNodeId`),
  KEY `IDX_CLS_EVT_TYPE` (`eventType`, `eventTime`),
  KEY `IDX_CLS_EVT_EXPIRE` (`expireTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群事件表 - 存储集群中各节点发布的事件';

