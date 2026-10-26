-- SQL Server 方言。已有库补丁：工具市场 hub0010 及超级管理员/只读角色授权。
-- 可重复执行（IF NOT EXISTS）。
-- 只读角色只有模块和查看；具体工具在角色管理里单独授予。

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010', N'default', N'工具市场', N'hub0010', N'MODULE', N'/system/toolMarketplace', N'group0001', 2, 9, N'AppsOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010', 1, N'Y');

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010:view', N'default', N'查看', N'hub0010:view', N'BUTTON', N'hub0010', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010_001', 1, N'Y');

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010:install')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010:install', N'default', N'安装', N'hub0010:install', N'BUTTON', N'hub0010', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010_002', 1, N'Y');

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010:uninstall')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010:uninstall', N'default', N'卸载', N'hub0010:uninstall', N'BUTTON', N'hub0010', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010_003', 1, N'Y');

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010:config')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010:config', N'default', N'配置', N'hub0010:config', N'BUTTON', N'hub0010', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010_004', 1, N'Y');

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010:tool:json-formatter')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010:tool:json-formatter', N'default', N'JSON格式化', N'hub0010:tool:json-formatter', N'BUTTON', N'hub0010', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010_005', 1, N'Y');

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010:tool:timestamp')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010:tool:timestamp', N'default', N'时间戳转换', N'hub0010:tool:timestamp', N'BUTTON', N'hub0010', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010_006', 1, N'Y');

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010:tool:base64')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010:tool:base64', N'default', N'Base64', N'hub0010:tool:base64', N'BUTTON', N'hub0010', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010_007', 1, N'Y');

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0010:tool:jwt-decoder')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0010:tool:jwt-decoder', N'default', N'JWT解码', N'hub0010:tool:jwt-decoder', N'BUTTON', N'hub0010', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0010_008', 1, N'Y');

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
    N'hub0010', N'hub0010:view', N'hub0010:install', N'hub0010:uninstall', N'hub0010:config',
    N'hub0010:tool:json-formatter', N'hub0010:tool:timestamp', N'hub0010:tool:base64', N'hub0010:tool:jwt-decoder'
  )
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
  AND resourceId IN (N'hub0010', N'hub0010:view')
  AND NOT EXISTS (
    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr
    WHERE rr.tenantId = HUB_AUTH_RESOURCE.tenantId
      AND rr.roleId = N'ROLE_VIEWER'
      AND rr.resourceId = HUB_AUTH_RESOURCE.resourceId
  );
