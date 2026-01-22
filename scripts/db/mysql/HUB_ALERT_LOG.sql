CREATE TABLE `HUB_ALERT_LOG` (
  -- 主键和租户
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，主键',
  `alertLogId` VARCHAR(32) NOT NULL COMMENT '告警日志ID，主键',
  
  -- 告警基本信息
  `alertLevel` VARCHAR(20) NOT NULL DEFAULT 'INFO' COMMENT '告警级别：INFO/WARN/ERROR/CRITICAL',
  `alertType` VARCHAR(100) DEFAULT NULL COMMENT '告警类型，业务自定义类型标识',
  `alertTitle` VARCHAR(500) NOT NULL COMMENT '告警标题',
  `alertContent` TEXT DEFAULT NULL COMMENT '告警内容',
  `alertTimestamp` DATETIME NOT NULL COMMENT '告警时间戳',
  
  -- 关联信息
  `channelName` VARCHAR(100) DEFAULT NULL COMMENT '使用的渠道名称',
  
  -- 发送信息
  `sendStatus` VARCHAR(20) DEFAULT NULL COMMENT '发送状态：PENDING待发送/SENDING发送中/SUCCESS成功/FAILED失败',
  `sendTime` DATETIME DEFAULT NULL COMMENT '发送时间',
  `sendResult` TEXT DEFAULT NULL COMMENT '发送结果详情，JSON格式',
  `sendErrorMessage` TEXT DEFAULT NULL COMMENT '发送错误信息',
  
  -- 标签和扩展信息
  `alertTags` TEXT DEFAULT NULL COMMENT '告警标签，JSON格式',
  `alertExtra` TEXT DEFAULT NULL COMMENT '告警额外数据，JSON格式',
  `tableData` TEXT DEFAULT NULL COMMENT '表格数据，JSON格式',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：N非活动，Y活动',
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
  PRIMARY KEY (`tenantId`, `alertLogId`),
  INDEX `IDX_ALERT_LOG_TENANT` (`tenantId`),
  INDEX `IDX_ALERT_LOG_LEVEL` (`alertLevel`),
  INDEX `IDX_ALERT_LOG_TYPE` (`alertType`),
  INDEX `IDX_ALERT_LOG_TIMESTAMP` (`alertTimestamp`),
  INDEX `IDX_ALERT_LOG_CHANNEL` (`channelName`),
  INDEX `IDX_ALERT_LOG_SEND_STATUS` (`sendStatus`),
  INDEX `idx_ALERT_LOG_TIME_STATUS` (`alertTimestamp`, `sendStatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警日志表 - 记录业务写入的告警日志和发送状态';

