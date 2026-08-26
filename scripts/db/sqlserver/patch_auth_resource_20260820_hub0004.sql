-- SQL Server 方言，由 scripts/db/mysql/patch_auth_resource_20260820_hub0004.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- 已有库补丁：审计日志查看模块 hub0004 及超级管理员/只读角色授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（IF NOT EXISTS）。

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004', N'default', N'审计日志', N'hub0004', N'MODULE', N'/system/auditLogManagement', N'group0001', 2, 7, N'ShieldCheckmarkOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'审计日志' WHERE tenantId = N'default' AND resourceId = N'hub0004';


IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004:view', N'default', N'查看详情', N'hub0004:view', N'BUTTON', N'hub0004', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0004:view';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004:search', N'default', N'查询', N'hub0004:search', N'BUTTON', N'hub0004', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0004:search';
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004:reset', N'default', N'重置', N'hub0004:reset', N'BUTTON', N'hub0004', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0004:reset';


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
  AND resourceId IN (N'hub0004', N'hub0004:view', N'hub0004:search', N'hub0004:reset')
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
  AND resourceId IN (N'hub0004', N'hub0004:view', N'hub0004:search', N'hub0004:reset')
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = HUB_AUTH_RESOURCE.tenantId
      AND rr.roleId = N'ROLE_VIEWER'
      AND rr.resourceId = HUB_AUTH_RESOURCE.resourceId
  );
