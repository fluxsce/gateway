-- 告警模板表
CREATE TABLE IF NOT EXISTS HUB_ALERT_TEMPLATE (
  -- 主键和租户
  tenantId TEXT NOT NULL,
  templateName TEXT NOT NULL,
  
  -- 模板基本信息
  templateDesc TEXT,
  channelType TEXT,
  
  -- 模板内容
  titleTemplate TEXT,
  contentTemplate TEXT,
  displayFormat TEXT NOT NULL DEFAULT 'table',
  templateVariables TEXT,
  
  -- 附件配置
  attachmentConfig TEXT,
  
  -- 通用字段
  addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho TEXT NOT NULL,
  editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  
  -- 预留字段
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, templateName)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS IDX_ALERT_TEMPLATE_TENANT ON HUB_ALERT_TEMPLATE(tenantId);
CREATE INDEX IF NOT EXISTS IDX_ALERT_TEMPLATE_CHANNEL ON HUB_ALERT_TEMPLATE(channelType);
CREATE INDEX IF NOT EXISTS IDX_ALERT_TEMPLATE_ACTIVE ON HUB_ALERT_TEMPLATE(activeFlag);
