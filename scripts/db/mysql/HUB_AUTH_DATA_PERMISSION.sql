-- =====================================================
-- 数据权限表 - 存储用户和角色的数据访问权限
-- =====================================================
CREATE TABLE IF NOT EXISTS `HUB_AUTH_DATA_PERMISSION` (
  -- 主键和租户信息
  `dataPermissionId` VARCHAR(32) NOT NULL COMMENT '数据权限ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联信息
  `userId` VARCHAR(32) DEFAULT NULL COMMENT '用户ID，为空表示角色级权限',
  `roleId` VARCHAR(32) DEFAULT NULL COMMENT '角色ID，为空表示用户级权限',
  
  -- 数据权限信息
  `resourceType` VARCHAR(50) NOT NULL COMMENT '资源类型(TABLE:数据表,API:接口,MODULE:模块)',
  `resourceCode` VARCHAR(100) NOT NULL COMMENT '资源编码',
  `scopeValue` TEXT DEFAULT NULL COMMENT '权限范围值，JSON格式',
  
  -- 权限条件
  `filterCondition` TEXT DEFAULT NULL COMMENT '过滤条件，SQL WHERE条件',
  `columnPermissions` TEXT DEFAULT NULL COMMENT '字段权限，JSON格式',
  `operationPermissions` VARCHAR(50) DEFAULT 'read' COMMENT '操作权限(read:只读,write:读写,delete:删除)',
  
  -- 生效时间
  `effectiveTime` DATETIME DEFAULT NULL COMMENT '生效时间',
  `expireTime` DATETIME DEFAULT NULL COMMENT '过期时间',
  
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
  PRIMARY KEY (`tenantId`, `dataPermissionId`),
  KEY `IDX_AUTH_DATA_PERM_USER` (`tenantId`, `userId`),
  KEY `IDX_AUTH_DATA_PERM_ROLE` (`tenantId`, `roleId`),
  KEY `IDX_AUTH_DATA_PERM_RESOURCE` (`resourceType`, `resourceCode`),
  KEY `IDX_AUTH_DATA_PERM_EXPIRE` (`expireTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据权限表 - 存储用户和角色的数据访问权限';

