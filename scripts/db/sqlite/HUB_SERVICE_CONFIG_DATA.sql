-- 配置数据表 - 存储具体的配置数据，支持多种配置格式
CREATE TABLE IF NOT EXISTS HUB_SERVICE_CONFIG_DATA (
  -- 主键和租户信息
  configDataId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联命名空间和分组
  namespaceId TEXT NOT NULL,
  groupName TEXT NOT NULL DEFAULT 'DEFAULT_GROUP',
  
  -- 配置基本信息
  configContent TEXT NOT NULL,
  contentType TEXT DEFAULT 'text',
  
  -- 配置描述和属性
  configDescription TEXT,
  encrypted TEXT DEFAULT 'N',
  
  -- 版本信息
  version INTEGER NOT NULL DEFAULT 1,
  
  -- MD5校验值（用于配置变更检测）
  md5Value TEXT,
  
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
  
  PRIMARY KEY (tenantId, namespaceId, groupName, configDataId)
);

CREATE INDEX IDX_SVC_CFG_DATA_NS_ID ON HUB_SERVICE_CONFIG_DATA(tenantId, namespaceId);
CREATE INDEX IDX_SVC_CFG_DATA_MD5 ON HUB_SERVICE_CONFIG_DATA(md5Value);
CREATE INDEX IDX_SVC_CFG_DATA_ACTIVE ON HUB_SERVICE_CONFIG_DATA(activeFlag);

