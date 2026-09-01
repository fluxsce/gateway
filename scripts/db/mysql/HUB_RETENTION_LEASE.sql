CREATE TABLE `HUB_RETENTION_LEASE` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，集群级租约固定 default',
  `leaseKey` VARCHAR(64) NOT NULL COMMENT '租约键，当前仅 cleanup',
  `owner` VARCHAR(128) NOT NULL COMMENT '持有节点ID',
  `expireTime` DATETIME NOT NULL COMMENT '租约过期时间',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(64) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(64) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(64) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` TEXT DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性(JSON格式)',
  PRIMARY KEY (`tenantId`, `leaseKey`)
) ENGINE=InnoDB COMMENT='生命周期清理租约，保证集群内同一时刻只有一个节点执行清理';
