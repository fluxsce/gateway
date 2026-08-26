-- SQL Server 方言，由 scripts/db/mysql/patch_auth_resource_20260820.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- 已有库补丁：安全配置写入口、预警日志编辑、超级管理员补授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（IF NOT EXISTS）。

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:securityConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:securityConfig', N'default', N'安全配置', N'hub0020:securityConfig', N'BUTTON', N'hub0020:globalConfig', 4, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_011_SEC', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'安全配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:securityConfig';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:securityConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:securityConfig', N'default', N'安全配置', N'hub0021:securityConfig', N'BUTTON', N'hub0021:routeConfig', 4, 20, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_013_SEC', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'安全配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:securityConfig';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:securityConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:securityConfig', N'default', N'安全配置', N'hub0022:securityConfig', N'BUTTON', N'hub0022', 3, 10, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'安全配置' WHERE tenantId = N'default' AND resourceId = N'hub0022:securityConfig';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:edit', N'default', N'更新日志', N'hub0082:edit', N'BUTTON', N'hub0082', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'更新日志' WHERE tenantId = N'default' AND resourceId = N'hub0082:edit';


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
  AND resourceId IN (
    N'hub0020:securityConfig',
    N'hub0021:securityConfig',
    N'hub0022:securityConfig',
    N'hub0082:edit'
  )
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = HUB_AUTH_RESOURCE.tenantId
      AND rr.roleId = N'ROLE_SUPER_ADMIN'
      AND rr.resourceId = HUB_AUTH_RESOURCE.resourceId
  );
