-- 已有库补丁：审计日志导出按钮 hub0004:export。
-- 新库执行 init.sql / HUB_AUTH_RESOURCE.sql 即可，不必再跑本文件。
-- 可重复执行（INSERT IGNORE）。

INSERT IGNORE INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0004:export', 'default', '导出', 'hub0004:export', 'BUTTON',
  'hub0004', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_009_004', 1, 'Y'
);

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
  AND `resourceId` = 'hub0004:export';

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
  AND `resourceId` = 'hub0004:export';
