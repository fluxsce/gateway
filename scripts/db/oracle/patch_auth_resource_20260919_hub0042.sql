-- 已有库补丁：服务列表批量删除及超级管理员授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（WHERE NOT EXISTS）。

INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT 'hub0042:batchDelete', 'default', '批量删除', 'hub0042:batchDelete', 'BUTTON',
       'hub0042', 3, 11, 'zh-CN', 'Y', 'Y',
       SYSDATE, 'system', SYSDATE, 'system', 'INIT_022_011', 1, 'Y'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0042:batchDelete'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_SUPER_ADMIN_' || REPLACE(resourceId, ':', '_'),
  tenantId,
  'ROLE_SUPER_ADMIN',
  resourceId,
  'ALLOW',
  'system',
  SYSDATE,
  SYSDATE,
  'system',
  SYSDATE,
  'system',
  'INIT_SA',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE r
WHERE r.tenantId = 'default'
  AND r.resourceId = 'hub0042:batchDelete'
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = r.tenantId
      AND rr.roleId = 'ROLE_SUPER_ADMIN'
      AND rr.resourceId = r.resourceId
  );

COMMIT;
