
-- 服务分组表 - 存储服务分组和授权信息
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_SERVICE_GROUP (
  -- 主键和租户信息
  serviceGroupId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 分组基本信息
  groupName TEXT NOT NULL,
  groupDescription TEXT,
  groupType TEXT DEFAULT 'BUSINESS',
  
  -- 授权信息
  ownerUserId TEXT NOT NULL,
  adminUserIds TEXT,
  readUserIds TEXT,
  accessControlEnabled TEXT DEFAULT 'N',
  
  -- 配置信息
  defaultProtocolType TEXT DEFAULT 'HTTP',
  defaultLoadBalanceStrategy TEXT DEFAULT 'ROUND_ROBIN',
  defaultHealthCheckUrl TEXT DEFAULT '/health',
  defaultHealthCheckIntervalSeconds INTEGER DEFAULT 30,
  
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
  
  PRIMARY KEY (tenantId, serviceGroupId)
);

CREATE INDEX IF NOT EXISTS IDX_REGISTRY_GROUP_NAME ON HUB_REGISTRY_SERVICE_GROUP(tenantId, groupName);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_GROUP_TYPE ON HUB_REGISTRY_SERVICE_GROUP(groupType);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_GROUP_OWNER ON HUB_REGISTRY_SERVICE_GROUP(ownerUserId);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_GROUP_ACTIVE ON HUB_REGISTRY_SERVICE_GROUP(activeFlag);