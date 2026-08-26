-- SQL Server 方言，由 scripts/db/mysql/patch_auth_resource_20260820_hub0004_export.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- 已有库补丁：审计日志导出按钮 hub0004:export。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（IF NOT EXISTS）。

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004:export')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004:export', N'default', N'导出', N'hub0004:export', N'BUTTON', N'hub0004', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'导出' WHERE tenantId = N'default' AND resourceId = N'hub0004:export';


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
  AND resourceId = N'hub0004:export'
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = HUB_AUTH_RESOURCE.tenantId
      AND rr.roleId = N'ROLE_SUPER_ADMIN'
      AND rr.resourceId = HUB_AUTH_RESOURCE.resourceId
  );


INSERT INTO HUB_AUTH_ROLE_RESOURCE (roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag)
SELECT
  CONCAT(N'ROLE_RES_VIEWER_', REPLACE(resourceId, N':', N'_')),
  tenantId,
  N'ROLE_VIEWER',
  resourceId,
  N'ALLOW',
  N'system',
  GETDATE(),
  GETDATE(),
  N'system',
  GETDATE(),
  N'system',
  N'INIT_VIEWER',
  1,
  N'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = N'default'
  AND resourceId = N'hub0004:export'
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = HUB_AUTH_RESOURCE.tenantId
      AND rr.roleId = N'ROLE_VIEWER'
      AND rr.resourceId = HUB_AUTH_RESOURCE.resourceId
  );
