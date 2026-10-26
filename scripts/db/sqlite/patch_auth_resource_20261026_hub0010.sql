-- 已有库补丁：工具市场 hub0010 及超级管理员/只读角色授权。
-- 新库执行 HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（INSERT OR IGNORE）。
-- 只读角色只有模块和查看；具体工具在角色管理里单独授予。

INSERT OR IGNORE INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0010', 'default', '工具市场', 'hub0010', 'MODULE',
  '/system/toolMarketplace', 'group0001', 2, 9, 'AppsOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010', 1, 'Y'
);

INSERT OR IGNORE INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES
  ('hub0010:view', 'default', '查看', 'hub0010:view', 'BUTTON',
   'hub0010', 3, 1, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010_001', 1, 'Y'),
  ('hub0010:install', 'default', '安装', 'hub0010:install', 'BUTTON',
   'hub0010', 3, 2, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010_002', 1, 'Y'),
  ('hub0010:uninstall', 'default', '卸载', 'hub0010:uninstall', 'BUTTON',
   'hub0010', 3, 3, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010_003', 1, 'Y'),
  ('hub0010:config', 'default', '配置', 'hub0010:config', 'BUTTON',
   'hub0010', 3, 4, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010_004', 1, 'Y'),
  ('hub0010:tool:json-formatter', 'default', 'JSON格式化', 'hub0010:tool:json-formatter', 'BUTTON',
   'hub0010', 3, 5, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010_005', 1, 'Y'),
  ('hub0010:tool:timestamp', 'default', '时间戳转换', 'hub0010:tool:timestamp', 'BUTTON',
   'hub0010', 3, 6, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010_006', 1, 'Y'),
  ('hub0010:tool:base64', 'default', 'Base64', 'hub0010:tool:base64', 'BUTTON',
   'hub0010', 3, 7, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010_007', 1, 'Y'),
  ('hub0010:tool:jwt-decoder', 'default', 'JWT解码', 'hub0010:tool:jwt-decoder', 'BUTTON',
   'hub0010', 3, 8, 'zh-CN', 'Y', 'Y',
   datetime('now'), 'system', datetime('now'), 'system', 'INIT_HUB0010_008', 1, 'Y');

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
  AND resourceId IN (
    'hub0010', 'hub0010:view', 'hub0010:install', 'hub0010:uninstall', 'hub0010:config',
    'hub0010:tool:json-formatter', 'hub0010:tool:timestamp', 'hub0010:tool:base64', 'hub0010:tool:jwt-decoder'
  );

INSERT OR IGNORE INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  'ROLE_RES_VIEWER_' || replace(resourceId, ':', '_'),
  tenantId,
  'ROLE_VIEWER',
  resourceId,
  'ALLOW',
  'system',
  datetime('now'),
  datetime('now'),
  'system',
  datetime('now'),
  'system',
  'INIT_VIEWER',
  1,
  'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = 'default'
  AND resourceId IN ('hub0010', 'hub0010:view');
