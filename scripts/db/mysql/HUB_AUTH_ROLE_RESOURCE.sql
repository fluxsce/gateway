-- =====================================================
-- 角色权限关联表 - 存储角色与权限资源的关联关系
-- =====================================================
CREATE TABLE IF NOT EXISTS `HUB_AUTH_ROLE_RESOURCE` (
  -- 主键和租户信息
  `roleResourceId` VARCHAR(100) NOT NULL COMMENT '角色资源关联ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联信息
  `roleId` VARCHAR(32) NOT NULL COMMENT '角色ID',
  `resourceId` VARCHAR(100) NOT NULL COMMENT '资源ID',
  
  -- 权限控制
  `permissionType` VARCHAR(20) NOT NULL DEFAULT 'ALLOW' COMMENT '权限类型(ALLOW:允许,DENY:拒绝)',
  `grantedBy` VARCHAR(32) NOT NULL COMMENT '授权人ID',
  `grantedTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
  `expireTime` DATETIME DEFAULT NULL COMMENT '过期时间，NULL表示永不过期',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `roleResourceId`),
  UNIQUE KEY `IDX_AUTH_ROLE_RES_UNIQUE` (`tenantId`, `roleId`, `resourceId`),
  KEY `IDX_AUTH_ROLE_RES_ROLE` (`tenantId`, `roleId`),
  KEY `IDX_AUTH_ROLE_RES_RESOURCE` (`tenantId`, `resourceId`),
  KEY `IDX_AUTH_ROLE_RES_TYPE` (`permissionType`),
  KEY `IDX_AUTH_ROLE_RES_EXPIRE` (`expireTime`)
) ENGINE=InnoDB COMMENT='角色权限关联表 - 存储角色与权限资源的关联关系';

-- =====================================================
-- 初始化角色权限关联数据
-- 登录权限按角色资源表精确匹配，不会把 MODULE 自动展开为按钮。
-- 超级管理员：授予资源目录中全部有效资源（含 GROUP/MODULE/MENU/BUTTON）。
-- 只读用户：仅授予导航资源以及查看/查询/重置筛选/刷新/返回，排除写操作。
-- hub0023:reset 为日志重发，不属于只读范围。
-- =====================================================

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
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
  AND `activeFlag` = 'Y'
  AND `resourceStatus` = 'Y';

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
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
  AND `activeFlag` = 'Y'
  AND `resourceStatus` = 'Y'
  AND (
    `resourceType` IN ('GROUP', 'MODULE', 'MENU')
    OR `resourceCode` LIKE '%:view'
    OR `resourceCode` LIKE '%:search'
    OR `resourceCode` LIKE '%:resetQuery'
    OR `resourceCode` LIKE '%:back'
    OR `resourceCode` LIKE '%:refresh'
    OR (
      `resourceCode` LIKE '%:reset'
      AND `resourceCode` <> 'hub0023:reset'
    )
  );

-- 审计日志导出对只读角色开放（仍记 EXPORT 审计）
INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
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
