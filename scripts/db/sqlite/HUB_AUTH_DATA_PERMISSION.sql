-- 数据权限表 - 存储用户和角色的数据访问权限
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_DATA_PERMISSION (
  -- 主键和租户信息
  dataPermissionId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联信息
  userId TEXT,
  roleId TEXT,
  
  -- 数据权限信息
  resourceType TEXT NOT NULL,
  resourceCode TEXT NOT NULL,
  scopeValue TEXT,
  
  -- 权限条件
  filterCondition TEXT,
  columnPermissions TEXT,
  operationPermissions TEXT DEFAULT 'read',
  
  -- 生效时间
  effectiveTime TEXT,
  expireTime TEXT,
  
  -- 通用字段
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
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
  
  PRIMARY KEY (tenantId, dataPermissionId)
);

CREATE INDEX IF NOT EXISTS IDX_AUTH_DATA_PERM_USER ON HUB_AUTH_DATA_PERMISSION(tenantId, userId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_DATA_PERM_ROLE ON HUB_AUTH_DATA_PERMISSION(tenantId, roleId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_DATA_PERM_RESOURCE ON HUB_AUTH_DATA_PERMISSION(resourceType, resourceCode);
CREATE INDEX IF NOT EXISTS IDX_AUTH_DATA_PERM_EXPIRE ON HUB_AUTH_DATA_PERMISSION(expireTime);

