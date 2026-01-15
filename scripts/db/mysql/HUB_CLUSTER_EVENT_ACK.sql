CREATE TABLE `HUB_CLUSTER_EVENT_ACK` (
  -- 主键和租户信息
  `ackId` VARCHAR(64) NOT NULL COMMENT '确认ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `eventId` VARCHAR(64) NOT NULL COMMENT '事件ID',
  
  -- 处理节点
  `nodeId` VARCHAR(100) NOT NULL COMMENT '处理节点ID(hostname:port)',
  `nodeIp` VARCHAR(100) DEFAULT NULL COMMENT '处理节点IP',
  
  -- 处理状态
  `ackStatus` VARCHAR(20) NOT NULL DEFAULT 'PENDING' COMMENT '确认状态(PENDING/SUCCESS/FAILED/SKIPPED)',
  `processTime` DATETIME DEFAULT NULL COMMENT '处理时间',
  `resultMessage` VARCHAR(2000) DEFAULT NULL COMMENT '结果信息或错误信息',
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  
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
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `ackId`),
  -- 优化后的复合索引：支持 NOT EXISTS 子查询高效执行
  -- 查询模式: WHERE eventId=? AND nodeId=? AND ackStatus='SUCCESS'
  KEY `IDX_CLS_ACK_EVT_NODE` (`eventId`, `nodeId`, `ackStatus`),
  -- 辅助索引：用于按节点查询处理状态
  KEY `IDX_CLS_ACK_NODE` (`nodeId`, `ackStatus`, `processTime`),
  -- 辅助索引：用于清理任务
  KEY `IDX_CLS_ACK_CLEANUP` (`tenantId`, `addTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群事件确认表 - 跟踪各节点对事件的处理状态';

