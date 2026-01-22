-- 告警日志表
CREATE TABLE IF NOT EXISTS HUB_ALERT_LOG (
  -- 主键和租户
  tenantId TEXT NOT NULL,
  alertLogId TEXT NOT NULL,
  
  -- 告警基本信息
  alertLevel TEXT NOT NULL DEFAULT 'INFO',
  alertType TEXT,
  alertTitle TEXT NOT NULL,
  alertContent TEXT,
  alertTimestamp DATETIME NOT NULL,
  
  -- 关联信息
  channelName TEXT,
  
  -- 发送信息
  sendStatus TEXT,
  sendTime DATETIME,
  sendResult TEXT,
  sendErrorMessage TEXT,
  
  -- 标签和扩展信息
  alertTags TEXT,
  alertExtra TEXT,
  tableData TEXT,
  
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
  
  PRIMARY KEY (tenantId, alertLogId)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS IDX_ALERT_LOG_TENANT ON HUB_ALERT_LOG(tenantId);
CREATE INDEX IF NOT EXISTS IDX_ALERT_LOG_LEVEL ON HUB_ALERT_LOG(alertLevel);
CREATE INDEX IF NOT EXISTS IDX_ALERT_LOG_TYPE ON HUB_ALERT_LOG(alertType);
CREATE INDEX IF NOT EXISTS IDX_ALERT_LOG_TIMESTAMP ON HUB_ALERT_LOG(alertTimestamp);
CREATE INDEX IF NOT EXISTS IDX_ALERT_LOG_CHANNEL ON HUB_ALERT_LOG(channelName);
CREATE INDEX IF NOT EXISTS IDX_ALERT_LOG_SEND_STATUS ON HUB_ALERT_LOG(sendStatus);
CREATE INDEX IF NOT EXISTS idx_ALERT_LOG_TIME_STATUS ON HUB_ALERT_LOG(alertTimestamp, sendStatus);

