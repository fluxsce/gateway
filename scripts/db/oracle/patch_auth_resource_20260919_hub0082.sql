-- 已有库补丁：预警日志批量删除、忽略操作及超级管理员授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（WHERE NOT EXISTS）。

INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT 'hub0082:batchDelete', 'default', '批量删除', 'hub0082:batchDelete', 'BUTTON',
       'hub0082', 3, 6, 'zh-CN', 'Y', 'Y',
       SYSDATE, 'system', SYSDATE, 'system', 'INIT_042_006', 1, 'Y'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0082:batchDelete'
);

INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT 'hub0082:ignoreSelected', 'default', '忽略选中', 'hub0082:ignoreSelected', 'BUTTON',
       'hub0082', 3, 7, 'zh-CN', 'Y', 'Y',
       SYSDATE, 'system', SYSDATE, 'system', 'INIT_042_007', 1, 'Y'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0082:ignoreSelected'
);

INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT 'hub0082:ignoreGroup', 'default', '分组忽略', 'hub0082:ignoreGroup', 'BUTTON',
       'hub0082', 3, 8, 'zh-CN', 'Y', 'Y',
       SYSDATE, 'system', SYSDATE, 'system', 'INIT_042_008', 1, 'Y'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0082:ignoreGroup'
);

INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT 'hub0082:ignoreAll', 'default', '全部忽略', 'hub0082:ignoreAll', 'BUTTON',
       'hub0082', 3, 9, 'zh-CN', 'Y', 'Y',
       SYSDATE, 'system', SYSDATE, 'system', 'INIT_042_009', 1, 'Y'
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = 'default' AND resourceId = 'hub0082:ignoreAll'
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
  AND r.resourceId IN ('hub0082:batchDelete', 'hub0082:ignoreSelected', 'hub0082:ignoreGroup', 'hub0082:ignoreAll')
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = r.tenantId
      AND rr.roleId = 'ROLE_SUPER_ADMIN'
      AND rr.resourceId = r.resourceId
  );

COMMIT;
