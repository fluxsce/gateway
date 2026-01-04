
-- 服务表 - 存储服务基本信息
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_SERVICE (
  -- 主键和租户信息
  tenantId TEXT NOT NULL,
  serviceName TEXT NOT NULL,
  
  -- 关联分组（主键关联）
  serviceGroupId TEXT NOT NULL,
  -- 冗余字段（便于查询和展示）
  groupName TEXT NOT NULL,
  
  -- 服务基本信息
  serviceDescription TEXT,
  
  -- 注册管理配置
  registryType TEXT NOT NULL DEFAULT 'INTERNAL',
  externalRegistryConfig TEXT,
  
  -- 服务配置
  protocolType TEXT DEFAULT 'HTTP',
  contextPath TEXT DEFAULT '',
  loadBalanceStrategy TEXT DEFAULT 'ROUND_ROBIN',
  
  -- 健康检查配置
  healthCheckUrl TEXT DEFAULT '/health',
  healthCheckIntervalSeconds INTEGER DEFAULT 30,
  healthCheckTimeoutSeconds INTEGER DEFAULT 5,
  healthCheckType TEXT DEFAULT 'HTTP',
  healthCheckMode TEXT DEFAULT 'ACTIVE',
  
  -- 元数据和标签
  metadataJson TEXT,
  tagsJson TEXT,
  
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
  
  PRIMARY KEY (tenantId, serviceName)
);

CREATE INDEX IF NOT EXISTS IDX_REGISTRY_SVC_GROUP_ID ON HUB_REGISTRY_SERVICE(tenantId, serviceGroupId);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_SVC_GROUP_NAME ON HUB_REGISTRY_SERVICE(groupName);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_SVC_REGISTRY_TYPE ON HUB_REGISTRY_SERVICE(registryType);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_SVC_ACTIVE ON HUB_REGISTRY_SERVICE(activeFlag);