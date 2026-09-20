-- SQL Server 方言，由 scripts/db/mysql/patch_auth_resource_20260919_hub0082.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- 已有库补丁：预警日志批量删除、忽略操作及超级管理员授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（IF NOT EXISTS）。

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:batchDelete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:batchDelete', N'default', N'批量删除', N'hub0082:batchDelete', N'BUTTON', N'hub0082', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'批量删除' WHERE tenantId = N'default' AND resourceId = N'hub0082:batchDelete';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:ignoreSelected')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:ignoreSelected', N'default', N'忽略选中', N'hub0082:ignoreSelected', N'BUTTON', N'hub0082', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'忽略选中' WHERE tenantId = N'default' AND resourceId = N'hub0082:ignoreSelected';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:ignoreGroup')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:ignoreGroup', N'default', N'分组忽略', N'hub0082:ignoreGroup', N'BUTTON', N'hub0082', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'分组忽略' WHERE tenantId = N'default' AND resourceId = N'hub0082:ignoreGroup';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:ignoreAll')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:ignoreAll', N'default', N'全部忽略', N'hub0082:ignoreAll', N'BUTTON', N'hub0082', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'全部忽略' WHERE tenantId = N'default' AND resourceId = N'hub0082:ignoreAll';


INSERT INTO HUB_AUTH_ROLE_RESOURCE (roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag)
SELECT
  CONCAT(N'ROLE_RES_SUPER_ADMIN_', REPLACE(resourceId, N':', N'_')),
  tenantId,
  N'ROLE_SUPER_ADMIN',
  resourceId,
  N'ALLOW',
  N'system',
  GETDATE(),
  GETDATE(),
  N'system',
  GETDATE(),
  N'system',
  N'INIT_SA',
  1,
  N'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = N'default'
  AND resourceId IN (N'hub0082:batchDelete', N'hub0082:ignoreSelected', N'hub0082:ignoreGroup', N'hub0082:ignoreAll')
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = HUB_AUTH_RESOURCE.tenantId
      AND rr.roleId = N'ROLE_SUPER_ADMIN'
      AND rr.resourceId = HUB_AUTH_RESOURCE.resourceId
  );
