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
-- 登录权限按角色资源表精确匹配，不会把 MODULE 自动展开为按钮。
-- 超级管理员：授予资源目录中全部有效资源（含 GROUP/MODULE/MENU/BUTTON）。
-- 只读用户：仅授予导航资源以及查看/查询/重置筛选/刷新/返回，排除写操作。
-- hub0023:reset 为日志重发，不属于只读范围。
-- =====================================================

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_SUPER_ADMIN_' || replace(resourceId, ':', '_'),
  tenantId,
  'ROLE_SUPER_ADMIN',
  resourceId,
  'ALLOW',
  'system',
  datetime('now'),
  datetime('now'),
  'system',
  datetime('now'),
  'system',
  'INIT_SA',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = 'default'
  AND activeFlag = 'Y'
  AND resourceStatus = 'Y';

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_VIEWER_' || replace(resourceId, ':', '_'),
  tenantId,
  'ROLE_VIEWER',
  resourceId,
  'ALLOW',
  'system',
  datetime('now'),
  datetime('now'),
  'system',
  datetime('now'),
  'system',
  'INIT_VIEWER',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = 'default'
  AND activeFlag = 'Y'
  AND resourceStatus = 'Y'
  AND (
    resourceType IN ('GROUP', 'MODULE', 'MENU')
    OR resourceCode LIKE '%:view'
    OR resourceCode LIKE '%:search'
    OR resourceCode LIKE '%:resetQuery'
    OR resourceCode LIKE '%:back'
    OR resourceCode LIKE '%:refresh'
    OR (
      resourceCode LIKE '%:reset'
      AND resourceCode <> 'hub0023:reset'
    )
  );

-- 审计日志导出对只读角色开放（仍记 EXPORT 审计）
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_VIEWER_' || replace(resourceId, ':', '_'),
  tenantId,
  'ROLE_VIEWER',
  resourceId,
  'ALLOW',
  'system',
  datetime('now'),
  datetime('now'),
  'system',
  datetime('now'),
  'system',
  'INIT_VIEWER',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = 'default'
  AND resourceId = 'hub0004:export';
