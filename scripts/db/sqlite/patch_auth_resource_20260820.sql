-- 已有库补丁：安全配置写入口、预警日志编辑、超级管理员补授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（INSERT OR IGNORE）。

INSERT OR IGNORE INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES
  ('hub0020:securityConfig', 'default', '安全配置', 'hub0020:securityConfig', 'BUTTON',
   'hub0020:globalConfig', 4, 8, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_010_011_SEC', 1, 'Y'),
  ('hub0021:securityConfig', 'default', '安全配置', 'hub0021:securityConfig', 'BUTTON',
   'hub0021:routeConfig', 4, 20, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_011_013_SEC', 1, 'Y'),
  ('hub0022:securityConfig', 'default', '安全配置', 'hub0022:securityConfig', 'BUTTON',
   'hub0022', 3, 10, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_012_010', 1, 'Y'),
  ('hub0082:edit', 'default', '更新日志', 'hub0082:edit', 'BUTTON',
   'hub0082', 3, 5, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_042_005', 1, 'Y');

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
  AND resourceId IN (
    'hub0020:securityConfig',
    'hub0021:securityConfig',
    'hub0022:securityConfig',
    'hub0082:edit'
  );
