-- 配置历史表 - 记录配置的变更历史，支持配置回滚和审计
CREATE TABLE `HUB_SERVICE_CONFIG_HISTORY` (
  -- 主键和租户信息
  `configHistoryId` VARCHAR(32) NOT NULL COMMENT '配置历史ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联配置数据
  `configDataId` VARCHAR(100) NOT NULL COMMENT '配置数据ID，关联HUB_SERVICE_CONFIG_DATA表',
  `namespaceId` VARCHAR(32) NOT NULL COMMENT '命名空间ID，冗余字段便于查询',
  `groupName` VARCHAR(64) NOT NULL COMMENT '分组名称，冗余字段便于查询',
  
  -- 变更信息
  `changeType` VARCHAR(20) NOT NULL COMMENT '变更类型(CREATE:创建,UPDATE:更新,DELETE:删除,ROLLBACK:回滚)',
  `oldContent` LONGTEXT DEFAULT NULL COMMENT '旧配置内容',
  `newContent` LONGTEXT NOT NULL COMMENT '新配置内容',
  `oldVersion` BIGINT DEFAULT NULL COMMENT '旧版本号',
  `newVersion` BIGINT NOT NULL COMMENT '新版本号',
  `oldMd5Value` VARCHAR(32) DEFAULT NULL COMMENT '旧配置MD5值',
  `newMd5Value` VARCHAR(32) NOT NULL COMMENT '新配置MD5值',
  
  -- 变更原因和操作人
  `changeReason` VARCHAR(500) DEFAULT NULL COMMENT '变更原因',
  `changedBy` VARCHAR(32) NOT NULL COMMENT '变更人ID',
  `changedAt` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '变更时间',
  
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
  PRIMARY KEY (`tenantId`, `configHistoryId`),
  KEY `IDX_SVC_CFG_HIST_DATA_ID` (`tenantId`, `configDataId`),
  KEY `IDX_SVC_CFG_HIST_NS_ID` (`tenantId`, `namespaceId`),
  KEY `IDX_SVC_CFG_HIST_GROUP_NAME` (`tenantId`, `namespaceId`, `groupName`),
  KEY `IDX_SVC_CFG_HIST_TYPE` (`changeType`),
  KEY `IDX_SVC_CFG_HIST_TIME` (`changedAt`),
  KEY `IDX_SVC_CFG_HIST_BY` (`changedBy`),
  KEY `IDX_SVC_CFG_HIST_VERSION` (`newVersion`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置历史表 - 记录配置的变更历史，支持配置回滚和审计';

