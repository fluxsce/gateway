-- 已有库补丁：环境设置模块 hub0009 及超级管理员/只读角色授权。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（INSERT IGNORE）。

INSERT IGNORE INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0009', 'default', '环境设置', 'hub0009', 'MODULE',
  '/system/environmentSettings', 'group0001', 2, 8, 'OptionsOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_HUB0009', 1, 'Y'
);

INSERT IGNORE INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES
  ('hub0009:view', 'default', '查看', 'hub0009:view', 'BUTTON',
   'hub0009', 3, 1, 'zh-CN', 'Y', 'Y',
   NOW(), 'system', NOW(), 'system', 'INIT_HUB0009_001', 1, 'Y'),
  ('hub0009:edit', 'default', '保存', 'hub0009:edit', 'BUTTON',
   'hub0009', 3, 2, 'zh-CN', 'Y', 'Y',
   NOW(), 'system', NOW(), 'system', 'INIT_HUB0009_002', 1, 'Y');

INSERT IGNORE INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
)
SELECT
  CONCAT('ROLE_RES_SUPER_ADMIN_', REPLACE(`resourceId`, ':', '_')),
  `tenantId`,
  'ROLE_SUPER_ADMIN',
  `resourceId`,
  'ALLOW',
  'system',
  NOW(),
  NOW(),
  'system',
  NOW(),
  'system',
  'INIT_SA',
  1,
  'Y'
FROM `HUB_AUTH_RESOURCE`
WHERE `tenantId` = 'default'
  AND `resourceId` IN ('hub0009', 'hub0009:view', 'hub0009:edit');

INSERT IGNORE INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
)
SELECT
  CONCAT('ROLE_RES_VIEWER_', REPLACE(`resourceId`, ':', '_')),
  `tenantId`,
  'ROLE_VIEWER',
  `resourceId`,
  'ALLOW',
  'system',
  NOW(),
  NOW(),
  'system',
  NOW(),
  'system',
  'INIT_VIEWER',
  1,
  'Y'
FROM `HUB_AUTH_RESOURCE`
WHERE `tenantId` = 'default'
  AND `resourceId` IN ('hub0009', 'hub0009:view');
