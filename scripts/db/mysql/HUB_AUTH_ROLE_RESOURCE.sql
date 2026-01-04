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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表 - 存储角色与权限资源的关联关系';

-- =====================================================
-- 初始化角色权限关联数据
-- 为超级管理员角色分配所有模块权限
-- =====================================================

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0000', 'default', 'ROLE_SUPER_ADMIN', 'hub0000', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_001', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0001', 'default', 'ROLE_SUPER_ADMIN', 'hub0001', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_002', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002', 'default', 'ROLE_SUPER_ADMIN', 'hub0002', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_003', 1, 'Y'
);

-- 超级管理员 - 用户管理模块按钮权限
INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_ADD', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:add', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_003_001', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_EDIT', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:edit', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_003_002', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_DELETE', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:delete', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_003_003', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_RESET_PWD', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:resetPassword', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_003_004', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:view', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_003_005', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:search', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_003_006', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:reset', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_003_007', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005', 'default', 'ROLE_SUPER_ADMIN', 'hub0005', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_004', 1, 'Y'
);

-- 超级管理员 - 角色管理模块按钮权限
INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_ADD', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:add', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_004_001', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_EDIT', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:edit', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_004_002', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_DELETE', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:delete', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_004_003', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:view', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_004_004', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_ROLE_AUTH', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:roleAuth', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_004_005', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:search', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_004_006', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:reset', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_004_007', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006', 'default', 'ROLE_SUPER_ADMIN', 'hub0006', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_005', 1, 'Y'
);

-- 超级管理员 - 权限资源管理模块按钮权限
INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:view', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_005_001', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:search', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_005_002', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:reset', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_005_003', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0020', 'default', 'ROLE_SUPER_ADMIN', 'hub0020', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_010', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0021', 'default', 'ROLE_SUPER_ADMIN', 'hub0021', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_011', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0022', 'default', 'ROLE_SUPER_ADMIN', 'hub0022', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_012', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0023', 'default', 'ROLE_SUPER_ADMIN', 'hub0023', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_013', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0040', 'default', 'ROLE_SUPER_ADMIN', 'hub0040', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_020', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0041', 'default', 'ROLE_SUPER_ADMIN', 'hub0041', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_021', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0042', 'default', 'ROLE_SUPER_ADMIN', 'hub0042', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_022', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0060', 'default', 'ROLE_SUPER_ADMIN', 'hub0060', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_030', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0061', 'default', 'ROLE_SUPER_ADMIN', 'hub0061', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_031', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0062', 'default', 'ROLE_SUPER_ADMIN', 'hub0062', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_032', 1, 'Y'
);

INSERT INTO `HUB_AUTH_ROLE_RESOURCE` (
  `roleResourceId`, `tenantId`, `roleId`, `resourceId`, `permissionType`, `grantedBy`, `grantedTime`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0063', 'default', 'ROLE_SUPER_ADMIN', 'hub0063', 'ALLOW', 'system', NOW(),
  NOW(), 'system', NOW(), 'system', 'INIT_033', 1, 'Y'
);

