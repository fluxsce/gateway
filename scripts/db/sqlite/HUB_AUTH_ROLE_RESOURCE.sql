-- 角色权限关联表 - 存储角色与权限资源的关联关系
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_ROLE_RESOURCE (
  -- 主键和租户信息
  roleResourceId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联信息
  roleId TEXT NOT NULL,
  resourceId TEXT NOT NULL,
  
  -- 权限控制
  permissionType TEXT NOT NULL DEFAULT 'ALLOW',
  grantedBy TEXT NOT NULL,
  grantedTime TEXT NOT NULL DEFAULT (datetime('now')),
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
  
  PRIMARY KEY (tenantId, roleResourceId)
);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_ROLE ON HUB_AUTH_ROLE_RESOURCE(tenantId, roleId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_RESOURCE ON HUB_AUTH_ROLE_RESOURCE(tenantId, resourceId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_TYPE ON HUB_AUTH_ROLE_RESOURCE(permissionType);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_EXPIRE ON HUB_AUTH_ROLE_RESOURCE(expireTime);

CREATE UNIQUE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_UNIQUE ON HUB_AUTH_ROLE_RESOURCE(tenantId, roleId, resourceId);

-- =====================================================
-- 初始化角色权限关联数据
-- 为超级管理员角色分配所有模块权限
-- =====================================================

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0000', 'default', 'ROLE_SUPER_ADMIN', 'hub0000', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_001', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0001', 'default', 'ROLE_SUPER_ADMIN', 'hub0001', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_002', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002', 'default', 'ROLE_SUPER_ADMIN', 'hub0002', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003', 1, 'Y'
);

-- 超级管理员 - 用户管理模块按钮权限
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_ADD', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:add', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_001', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_EDIT', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:edit', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_002', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_DELETE', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:delete', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_003', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_RESET_PWD', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:resetPassword', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_004', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:view', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_005', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:search', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_006', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:reset', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_007', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005', 'default', 'ROLE_SUPER_ADMIN', 'hub0005', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004', 1, 'Y'
);

-- 超级管理员 - 角色管理模块按钮权限
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_ADD', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:add', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_001', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_EDIT', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:edit', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_002', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_DELETE', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:delete', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_003', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:view', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_004', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_ROLE_AUTH', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:roleAuth', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_005', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:search', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_006', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:reset', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_007', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006', 'default', 'ROLE_SUPER_ADMIN', 'hub0006', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005', 1, 'Y'
);

-- 超级管理员 - 权限资源管理模块按钮权限
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:view', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_001', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:search', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_002', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:reset', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_003', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0020', 'default', 'ROLE_SUPER_ADMIN', 'hub0020', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_010', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0021', 'default', 'ROLE_SUPER_ADMIN', 'hub0021', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_011', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0022', 'default', 'ROLE_SUPER_ADMIN', 'hub0022', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_012', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0023', 'default', 'ROLE_SUPER_ADMIN', 'hub0023', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_013', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0040', 'default', 'ROLE_SUPER_ADMIN', 'hub0040', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_020', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0041', 'default', 'ROLE_SUPER_ADMIN', 'hub0041', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_021', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0042', 'default', 'ROLE_SUPER_ADMIN', 'hub0042', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_022', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0060', 'default', 'ROLE_SUPER_ADMIN', 'hub0060', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_030', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0061', 'default', 'ROLE_SUPER_ADMIN', 'hub0061', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_031', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0062', 'default', 'ROLE_SUPER_ADMIN', 'hub0062', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_032', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0063', 'default', 'ROLE_SUPER_ADMIN', 'hub0063', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_033', 1, 'Y'
);