-- =====================================================
-- 用户角色关联表 - 存储用户与角色的关联关系
-- =====================================================
CREATE TABLE IF NOT EXISTS `HUB_AUTH_USER_ROLE` (
  -- 主键和租户信息
  `userRoleId` VARCHAR(100) NOT NULL COMMENT '用户角色关联ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联信息
  `userId` VARCHAR(32) NOT NULL COMMENT '用户ID',
  `roleId` VARCHAR(32) NOT NULL COMMENT '角色ID',
  
  -- 授权控制
  `grantedBy` VARCHAR(32) NOT NULL COMMENT '授权人ID',
  `grantedTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
  `expireTime` DATETIME DEFAULT NULL COMMENT '过期时间，NULL表示永不过期',
  `primaryRoleFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '主要角色标记(Y:主要角色,N:次要角色)',
  
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
  PRIMARY KEY (`tenantId`, `userRoleId`),
  UNIQUE KEY `IDX_AUTH_USER_ROLE_UNIQUE` (`tenantId`, `userId`, `roleId`),
  KEY `IDX_AUTH_USER_ROLE_USER` (`tenantId`, `userId`),
  KEY `IDX_AUTH_USER_ROLE_ROLE` (`tenantId`, `roleId`),
  KEY `IDX_AUTH_USER_ROLE_PRIMARY` (`primaryRoleFlag`),
  KEY `IDX_AUTH_USER_ROLE_EXPIRE` (`expireTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表 - 存储用户与角色的关联关系';

