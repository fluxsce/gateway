-- 角色表 - 存储系统角色信息和数据权限范围
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_ROLE (
  -- 主键和租户信息
  roleId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 角色基本信息
  roleName TEXT NOT NULL,
  roleDescription TEXT,
  
  -- 角色状态
  roleStatus TEXT NOT NULL DEFAULT 'Y',
  builtInFlag TEXT NOT NULL DEFAULT 'N',
  
  -- 数据权限范围
  dataScope TEXT DEFAULT NULL,
  
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
  
  PRIMARY KEY (tenantId, roleId)
);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_NAME ON HUB_AUTH_ROLE(tenantId, roleName);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_STATUS ON HUB_AUTH_ROLE(roleStatus);

-- =====================================================
-- 初始化角色数据
-- =====================================================

-- 超级管理员角色
INSERT INTO HUB_AUTH_ROLE (
  roleId, tenantId, roleName, roleDescription, 
  roleStatus, builtInFlag, dataScope,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_SUPER_ADMIN', 'default', '超级管理员', '拥有系统所有权限的超级管理员',
  'Y', 'Y', '{"type":"ALL"}',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_001', 1, 'Y'
);