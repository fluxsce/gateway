-- 已有库补丁：审计日志查看模块 hub0004 及超级管理员/只读角色授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（INSERT OR IGNORE）。

INSERT OR IGNORE INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0004', 'default', '审计日志', 'hub0004', 'MODULE',
  '/system/auditLogManagement', 'group0001', 2, 7, 'ShieldCheckmarkOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_009', 1, 'Y'
);

INSERT OR IGNORE INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES
  ('hub0004:view', 'default', '查看详情', 'hub0004:view', 'BUTTON',
   'hub0004', 3, 1, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_009_001', 1, 'Y'),
  ('hub0004:search', 'default', '查询', 'hub0004:search', 'BUTTON',
   'hub0004', 3, 2, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_009_002', 1, 'Y'),
  ('hub0004:reset', 'default', '重置', 'hub0004:reset', 'BUTTON',
   'hub0004', 3, 3, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_009_003', 1, 'Y');

INSERT OR IGNORE INTO HUB_AUTH_ROLE_RESOURCE (
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
  AND resourceId IN ('hub0004', 'hub0004:view', 'hub0004:search', 'hub0004:reset');

INSERT OR IGNORE INTO HUB_AUTH_ROLE_RESOURCE (
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
  AND resourceId IN ('hub0004', 'hub0004:view', 'hub0004:search', 'hub0004:reset');
