-- 已有库补丁：审计日志导出按钮 hub0004:export。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（WHERE NOT EXISTS）。

INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT 'hub0004:export', 'default', '导出', 'hub0004:export', 'BUTTON',
       'hub0004', 3, 4, 'zh-CN', 'Y', 'Y',
       SYSDATE, 'system', SYSDATE, 'system', 'INIT_009_004', 1, 'Y'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0004:export'
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
  AND r.resourceId = 'hub0004:export'
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = r.tenantId
      AND rr.roleId = 'ROLE_SUPER_ADMIN'
      AND rr.resourceId = r.resourceId
  );

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_VIEWER_' || REPLACE(resourceId, ':', '_'),
  tenantId,
  'ROLE_VIEWER',
  resourceId,
  'ALLOW',
  'system',
  SYSDATE,
  SYSDATE,
  'system',
  SYSDATE,
  'system',
  'INIT_VIEWER',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE r
WHERE r.tenantId = 'default'
  AND r.resourceId = 'hub0004:export'
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = r.tenantId
      AND rr.roleId = 'ROLE_VIEWER'
      AND rr.resourceId = r.resourceId
  );
