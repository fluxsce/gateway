-- =====================================================
-- 角色表 - 存储系统角色信息和数据权限范围
-- =====================================================
CREATE TABLE IF NOT EXISTS `HUB_AUTH_ROLE` (
  -- 主键和租户信息
  `roleId` VARCHAR(32) NOT NULL COMMENT '角色ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 角色基本信息
  `roleName` VARCHAR(100) NOT NULL COMMENT '角色名称',
  `roleDescription` VARCHAR(500) DEFAULT NULL COMMENT '角色描述',
  
  -- 角色状态
  `roleStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '角色状态(Y:启用,N:禁用)',
  `builtInFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '内置角色标记(Y:内置,N:自定义)',
  
  -- 数据权限范围
  `dataScope` TEXT DEFAULT NULL COMMENT '数据权限范围，TEXT类型，支持存储复杂的权限配置(JSON格式)',
  
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
  PRIMARY KEY (`tenantId`, `roleId`),
  KEY `IDX_AUTH_ROLE_NAME` (`tenantId`, `roleName`),
  KEY `IDX_AUTH_ROLE_STATUS` (`roleStatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表 - 存储系统角色信息和数据权限范围';

-- =====================================================
-- 初始化角色数据
-- =====================================================

-- 超级管理员角色
INSERT INTO `HUB_AUTH_ROLE` (
  `roleId`, `tenantId`, `roleName`, `roleDescription`, 
  `roleStatus`, `builtInFlag`, `dataScope`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'ROLE_SUPER_ADMIN', 'default', '超级管理员', '拥有系统所有权限的超级管理员',
  'Y', 'Y', '{"type":"ALL"}',
  NOW(), 'system', NOW(), 'system', 'INIT_001', 1, 'Y'
);

