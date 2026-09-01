CREATE TABLE `HUB_SYS_SETTING` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `groupCode` VARCHAR(64) NOT NULL COMMENT '设置分组编码(retention/webTimeout等)',
  `content` TEXT NOT NULL COMMENT '分组内容JSON',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(64) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(64) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(64) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` TEXT DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性(JSON格式)',
  PRIMARY KEY (`tenantId`, `groupCode`)
) ENGINE=InnoDB COMMENT='环境设置表 - 按租户与分组存储平台策略';
