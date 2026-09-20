-- 已有库补丁：服务列表批量删除及超级管理员授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（INSERT OR IGNORE）。

INSERT OR IGNORE INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES
  ('hub0042:batchDelete', 'default', '批量删除', 'hub0042:batchDelete', 'BUTTON',
   'hub0042', 3, 11, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_022_011', 1, 'Y');

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
  AND resourceId = 'hub0042:batchDelete';
