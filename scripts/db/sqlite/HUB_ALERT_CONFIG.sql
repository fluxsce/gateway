-- 告警渠道配置表
CREATE TABLE IF NOT EXISTS HUB_ALERT_CONFIG (
  -- 主键和租户
  tenantId TEXT NOT NULL,
  channelName TEXT NOT NULL,
  
  -- 渠道基本信息
  channelType TEXT NOT NULL,
  channelDesc TEXT,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  defaultFlag TEXT NOT NULL DEFAULT 'N',
  priorityLevel INTEGER NOT NULL DEFAULT 10,
  defaultTemplateName TEXT,
  
  -- 服务器配置（JSON格式）
  serverConfig TEXT,
  sendConfig TEXT,
  
  -- 消息格式配置
  messageContentFormat TEXT,
  
  -- 重试和超时配置
  timeoutSeconds INTEGER NOT NULL DEFAULT 30,
  retryCount INTEGER NOT NULL DEFAULT 3,
  retryIntervalSecs INTEGER NOT NULL DEFAULT 5,
  asyncSendFlag TEXT NOT NULL DEFAULT 'N',
  
  -- 统计信息
  totalSentCount INTEGER NOT NULL DEFAULT 0,
  successCount INTEGER NOT NULL DEFAULT 0,
  failureCount INTEGER NOT NULL DEFAULT 0,
  lastSendTime DATETIME,
  lastSuccessTime DATETIME,
  lastFailureTime DATETIME,
  lastErrorMessage TEXT,
  
  -- 通用字段
  addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho TEXT NOT NULL,
  editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
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
  
  PRIMARY KEY (tenantId, channelName)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS IDX_ALERT_CONFIG_TENANT ON HUB_ALERT_CONFIG(tenantId);
CREATE INDEX IF NOT EXISTS IDX_ALERT_CONFIG_TYPE ON HUB_ALERT_CONFIG(channelType);
CREATE INDEX IF NOT EXISTS IDX_ALERT_CONFIG_ACTIVE ON HUB_ALERT_CONFIG(activeFlag);
CREATE INDEX IF NOT EXISTS IDX_ALERT_CONFIG_DEFAULT ON HUB_ALERT_CONFIG(defaultFlag);
CREATE INDEX IF NOT EXISTS IDX_ALERT_CONFIG_PRIORITY ON HUB_ALERT_CONFIG(priorityLevel);
CREATE INDEX IF NOT EXISTS IDX_ALERT_CONFIG_TEMPLATE ON HUB_ALERT_CONFIG(defaultTemplateName);
