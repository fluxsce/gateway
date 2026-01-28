-- 服务表 - 存储服务的基本信息和元数据
CREATE TABLE IF NOT EXISTS HUB_SERVICE (
  -- 主键和租户信息
  tenantId TEXT NOT NULL,
  namespaceId TEXT NOT NULL,
  groupName TEXT NOT NULL DEFAULT 'DEFAULT_GROUP',
  serviceName TEXT NOT NULL,
  
  -- 服务类型
  serviceType TEXT NOT NULL DEFAULT 'INTERNAL',
  
  -- 服务基本信息
  serviceVersion TEXT,
  serviceDescription TEXT,
  
  -- 外部服务配置（仅当serviceType非INTERNAL时使用）
  externalServiceConfig TEXT,
  
  -- 服务元数据
  metadataJson TEXT,
  tagsJson TEXT,
  
  -- 服务保护阈值（0-1之间的小数，表示健康实例比例低于该值时触发保护）
  protectThreshold REAL DEFAULT 0.00,
  
  -- 服务选择器（用于服务路由）
  selectorJson TEXT,
  
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
  
  PRIMARY KEY (tenantId, namespaceId, groupName, serviceName)
);

CREATE INDEX IDX_SVC_NS_ID ON HUB_SERVICE(tenantId, namespaceId);
CREATE INDEX IDX_SVC_NAME ON HUB_SERVICE(serviceName);
CREATE INDEX IDX_SVC_TYPE ON HUB_SERVICE(serviceType);
CREATE INDEX IDX_SVC_ACTIVE ON HUB_SERVICE(activeFlag);

