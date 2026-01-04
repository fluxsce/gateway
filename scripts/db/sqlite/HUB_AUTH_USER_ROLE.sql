-- 用户角色关联表 - 存储用户与角色的关联关系
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_USER_ROLE (
  -- 主键和租户信息
  userRoleId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联信息
  userId TEXT NOT NULL,
  roleId TEXT NOT NULL,
  
  -- 授权控制
  grantedBy TEXT NOT NULL,
  grantedTime TEXT NOT NULL DEFAULT (datetime('now')),
  expireTime TEXT,
  primaryRoleFlag TEXT NOT NULL DEFAULT 'N',
  
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
  
  PRIMARY KEY (tenantId, userRoleId)
);
CREATE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_USER ON HUB_AUTH_USER_ROLE(tenantId, userId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_ROLE ON HUB_AUTH_USER_ROLE(tenantId, roleId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_PRIMARY ON HUB_AUTH_USER_ROLE(primaryRoleFlag);
CREATE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_EXPIRE ON HUB_AUTH_USER_ROLE(expireTime);

CREATE UNIQUE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_UNIQUE ON HUB_AUTH_USER_ROLE(tenantId, userId, roleId);