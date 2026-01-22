CREATE TABLE `HUB_ALERT_TEMPLATE` (
  -- 主键和租户
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，主键',
  `templateName` VARCHAR(100) NOT NULL COMMENT '模板名称，主键',
  
  -- 模板基本信息
  `templateDesc` VARCHAR(500) DEFAULT NULL COMMENT '模板描述',
  `channelType` VARCHAR(50) DEFAULT NULL COMMENT '适用的渠道类型：email/qq/wechat_work/dingtalk/webhook/sms，为空表示通用模板',
  
  -- 模板内容
  `titleTemplate` VARCHAR(500) DEFAULT NULL COMMENT '标题模板，支持变量占位符如{{.Title}}',
  `contentTemplate` TEXT DEFAULT NULL COMMENT '内容模板，支持变量占位符',
  `displayFormat` VARCHAR(20) NOT NULL DEFAULT 'table' COMMENT '显示格式：table表格格式/text文本格式',
  `templateVariables` TEXT DEFAULT NULL COMMENT '模板变量定义，JSON格式，描述可用的变量和说明',
  
  -- 附件配置
  `attachmentConfig` TEXT DEFAULT NULL COMMENT '附件配置，JSON格式，用于邮件附件等',
  
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
  PRIMARY KEY (`tenantId`, `templateName`),
  INDEX `IDX_ALERT_TEMPLATE_TENANT` (`tenantId`),
  INDEX `IDX_ALERT_TEMPLATE_CHANNEL` (`channelType`),
  INDEX `IDX_ALERT_TEMPLATE_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警模板表 - 存储告警消息模板，支持变量占位符和多种格式';

