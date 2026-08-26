-- SQL Server 方言，由 scripts/db/mysql/patch_auth_resource_20260821_hub0009.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- 已有库补丁：环境设置模块 hub0009 及超级管理员/只读角色授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（IF NOT EXISTS）。

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0009')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0009', N'default', N'环境设置', N'hub0009', N'MODULE', N'/system/environmentSettings', N'group0001', 2, 8, N'OptionsOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'环境设置' WHERE tenantId = N'default' AND resourceId = N'hub0009';


IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0009:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0009:view', N'default', N'查看', N'hub0009:view', N'BUTTON', N'hub0009', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0009_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看' WHERE tenantId = N'default' AND resourceId = N'hub0009:view';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0009:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0009:edit', N'default', N'保存', N'hub0009:edit', N'BUTTON', N'hub0009', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0009_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'保存' WHERE tenantId = N'default' AND resourceId = N'hub0009:edit';


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
  AND resourceId IN (N'hub0009', N'hub0009:view', N'hub0009:edit')
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
  AND resourceId IN (N'hub0009', N'hub0009:view')
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = HUB_AUTH_RESOURCE.tenantId
      AND rr.roleId = N'ROLE_VIEWER'
      AND rr.resourceId = HUB_AUTH_RESOURCE.resourceId
  );
