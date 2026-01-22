CREATE TABLE `HUB_ALERT_CONFIG` (
  -- 主键和租户
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，主键',
  `channelName` VARCHAR(100) NOT NULL COMMENT '渠道名称，主键',
  
  -- 渠道基本信息
  `channelType` VARCHAR(50) NOT NULL COMMENT '渠道类型：email/qq/wechat_work/dingtalk/webhook/sms',
  `channelDesc` VARCHAR(500) DEFAULT NULL COMMENT '渠道描述',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '启用状态：Y-启用，N-禁用',
  `defaultFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否默认渠道：Y-是，N-否',
  `priorityLevel` INT NOT NULL DEFAULT 10 COMMENT '优先级：1-10，数字越小优先级越高',
  `defaultTemplateName` VARCHAR(100) DEFAULT NULL COMMENT '默认关联的模板名称',
  
  -- 服务器配置（JSON格式）
  `serverConfig` TEXT DEFAULT NULL COMMENT '服务器配置，JSON格式，如SMTP配置、Webhook URL等',
  `sendConfig` TEXT DEFAULT NULL COMMENT '发送配置，JSON格式，如默认收件人、超时设置等',
  
  -- 消息格式配置
  `messageContentFormat` VARCHAR(20) DEFAULT NULL COMMENT '消息内容格式：text/html/markdown',
  
  -- 重试和超时配置
  `timeoutSeconds` INT NOT NULL DEFAULT 30 COMMENT '超时时间（秒）',
  `retryCount` INT NOT NULL DEFAULT 3 COMMENT '重试次数',
  `retryIntervalSecs` INT NOT NULL DEFAULT 5 COMMENT '重试间隔（秒）',
  `asyncSendFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否异步发送：Y-是，N-否',
  
  -- 统计信息
  `totalSentCount` BIGINT NOT NULL DEFAULT 0 COMMENT '总发送次数',
  `successCount` BIGINT NOT NULL DEFAULT 0 COMMENT '成功次数',
  `failureCount` BIGINT NOT NULL DEFAULT 0 COMMENT '失败次数',
  `lastSendTime` DATETIME DEFAULT NULL COMMENT '最后发送时间',
  `lastSuccessTime` DATETIME DEFAULT NULL COMMENT '最后成功时间',
  `lastFailureTime` DATETIME DEFAULT NULL COMMENT '最后失败时间',
  `lastErrorMessage` VARCHAR(1000) DEFAULT NULL COMMENT '最后错误信息',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
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
  PRIMARY KEY (`tenantId`, `channelName`),
  INDEX `IDX_ALERT_CONFIG_TENANT` (`tenantId`),
  INDEX `IDX_ALERT_CONFIG_TYPE` (`channelType`),
  INDEX `IDX_ALERT_CONFIG_ACTIVE` (`activeFlag`),
  INDEX `IDX_ALERT_CONFIG_DEFAULT` (`defaultFlag`),
  INDEX `IDX_ALERT_CONFIG_PRIORITY` (`priorityLevel`),
  INDEX `IDX_ALERT_CONFIG_TEMPLATE` (`defaultTemplateName`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警渠道配置表 - 存储多渠道告警配置信息';
