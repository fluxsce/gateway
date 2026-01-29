-- 命名空间表 - 存储服务和配置的命名空间信息，用于多租户和多环境隔离
CREATE TABLE IF NOT EXISTS HUB_SERVICE_NAMESPACE (
  -- 主键和租户信息
  namespaceId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联服务中心实例
  instanceName TEXT NOT NULL,
  environment TEXT NOT NULL,
  
  -- 命名空间基本信息
  namespaceName TEXT NOT NULL,
  namespaceDescription TEXT,
  
  -- 命名空间配置
  serviceQuotaLimit INTEGER DEFAULT 200,
  configQuotaLimit INTEGER DEFAULT 200,
  
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
  
  PRIMARY KEY (tenantId, namespaceId)
);

CREATE INDEX IDX_SVC_NS_NAME ON HUB_SERVICE_NAMESPACE(tenantId, namespaceName);
CREATE INDEX IDX_SVC_NS_INSTANCE ON HUB_SERVICE_NAMESPACE(tenantId, instanceName, environment);
CREATE INDEX IDX_SVC_NS_ENV ON HUB_SERVICE_NAMESPACE(environment);
CREATE INDEX IDX_SVC_NS_ACTIVE ON HUB_SERVICE_NAMESPACE(activeFlag);

