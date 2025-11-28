-- =====================================================
-- 权限系统数据库表结构设计
-- 遵循 docs/database/naming-convention.md 规范
-- 基于 web/actions/permission-design.md 设计文档
-- 创建时间: 2024-12-19
-- =====================================================

-- =====================================================
-- 角色表 - 存储系统角色信息和数据权限范围
-- =====================================================
CREATE TABLE `HUB_AUTH_ROLE` (
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
-- 权限资源表 - 存储系统所有权限资源信息
-- =====================================================
CREATE TABLE `HUB_AUTH_RESOURCE` (
  -- 主键和租户信息
  `resourceId` VARCHAR(32) NOT NULL COMMENT '资源ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 资源基本信息
  `resourceName` VARCHAR(100) NOT NULL COMMENT '资源名称',
  `resourceCode` VARCHAR(100) NOT NULL COMMENT '资源编码，用于程序判断',
  `resourceType` VARCHAR(20) NOT NULL COMMENT '资源类型(MODULE:模块,MENU:菜单,BUTTON:按钮,API:接口) - 四级层次结构',
  `resourcePath` VARCHAR(500) DEFAULT NULL COMMENT '资源路径(菜单路径或API路径)',
  `resourceMethod` VARCHAR(10) DEFAULT NULL COMMENT '请求方法(GET,POST,PUT,DELETE等)',
  
  -- 层级关系
  `parentResourceId` VARCHAR(32) DEFAULT NULL COMMENT '父资源ID',
  `resourceLevel` INT NOT NULL DEFAULT 1 COMMENT '资源层级',
  `sortOrder` INT NOT NULL DEFAULT 0 COMMENT '排序顺序',
  
  -- 显示信息
  `displayName` VARCHAR(100) DEFAULT NULL COMMENT '显示名称',
  `iconClass` VARCHAR(100) DEFAULT NULL COMMENT '图标样式类',
  `description` VARCHAR(500) DEFAULT NULL COMMENT '资源描述',
  
  -- 状态信息
  `resourceStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '资源状态(Y:启用,N:禁用)',
  `builtInFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '内置资源标记(Y:内置,N:自定义)',
  
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
  PRIMARY KEY (`tenantId`, `resourceId`),
  UNIQUE KEY `IDX_AUTH_RES_CODE` (`tenantId`, `resourceCode`),
  KEY `IDX_AUTH_RES_TYPE` (`resourceType`),
  KEY `IDX_AUTH_RES_PARENT` (`parentResourceId`),
  KEY `IDX_AUTH_RES_PATH` (`resourcePath`),
  KEY `IDX_AUTH_RES_STATUS` (`resourceStatus`),
  KEY `IDX_AUTH_RES_LEVEL` (`resourceLevel`),
  KEY `IDX_AUTH_RES_SORT` (`sortOrder`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限资源表 - 存储系统所有权限资源信息';

-- =====================================================
-- 角色权限关联表 - 存储角色与权限资源的关联关系
-- =====================================================
CREATE TABLE `HUB_AUTH_ROLE_RESOURCE` (
  -- 主键和租户信息
  `roleResourceId` VARCHAR(100) NOT NULL COMMENT '角色资源关联ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联信息
  `roleId` VARCHAR(32) NOT NULL COMMENT '角色ID',
  `resourceId` VARCHAR(32) NOT NULL COMMENT '资源ID',
  
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
-- 用户角色关联表 - 存储用户与角色的关联关系
-- =====================================================
CREATE TABLE `HUB_AUTH_USER_ROLE` (
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

-- =====================================================
-- 数据权限表 - 存储用户和角色的数据访问权限
-- =====================================================
CREATE TABLE `HUB_AUTH_DATA_PERMISSION` (
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

-- =====================================================
-- 权限操作日志表 - 记录所有权限相关的操作日志
-- =====================================================
CREATE TABLE `HUB_AUTH_OPERATION_LOG` (
  -- 主键和租户信息
  `operationLogId` VARCHAR(32) NOT NULL COMMENT '操作日志ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 操作基本信息
  `operationType` VARCHAR(50) NOT NULL COMMENT '操作类型(ROLE_CREATE,ROLE_UPDATE,ROLE_DELETE,PERMISSION_GRANT,PERMISSION_REVOKE,USER_ROLE_ASSIGN等)',
  `operationTarget` VARCHAR(50) NOT NULL COMMENT '操作目标(ROLE,RESOURCE,USER_ROLE,DATA_PERMISSION)',
  `targetId` VARCHAR(32) NOT NULL COMMENT '目标ID',
  `targetName` VARCHAR(200) DEFAULT NULL COMMENT '目标名称',
  
  -- 操作人信息
  `operatorUserId` VARCHAR(32) NOT NULL COMMENT '操作人用户ID',
  `operatorUserName` VARCHAR(100) DEFAULT NULL COMMENT '操作人用户名',
  `operatorRealName` VARCHAR(100) DEFAULT NULL COMMENT '操作人真实姓名',
  `operatorIpAddress` VARCHAR(50) DEFAULT NULL COMMENT '操作人IP地址',
  
  -- 操作详情
  `operationDescription` VARCHAR(1000) DEFAULT NULL COMMENT '操作描述',
  `beforeData` TEXT DEFAULT NULL COMMENT '操作前数据，JSON格式',
  `afterData` TEXT DEFAULT NULL COMMENT '操作后数据，JSON格式',
  `operationResult` VARCHAR(20) NOT NULL DEFAULT 'SUCCESS' COMMENT '操作结果(SUCCESS:成功,FAILED:失败)',
  `errorMessage` VARCHAR(1000) DEFAULT NULL COMMENT '错误信息',
  
  -- 时间信息
  `operationTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  
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
  PRIMARY KEY (`tenantId`, `operationLogId`),
  KEY `IDX_AUTH_OP_LOG_TYPE` (`operationType`),
  KEY `IDX_AUTH_OP_LOG_TARGET` (`operationTarget`, `targetId`),
  KEY `IDX_AUTH_OP_LOG_OPERATOR` (`operatorUserId`),
  KEY `IDX_AUTH_OP_LOG_TIME` (`operationTime`),
  KEY `IDX_AUTH_OP_LOG_RESULT` (`operationResult`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限操作日志表 - 记录所有权限相关的操作日志';

-- =====================================================
-- 初始数据插入 - 内置角色
-- =====================================================

-- 超级管理员角色
INSERT INTO HUB_AUTH_ROLE (roleId, tenantId, roleName, roleDescription, dataScope, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('ROLE_SUPER_ADMIN', 'default', '超级管理员', '拥有系统所有权限的超级管理员', '{"type":"ALL"}', 'Y', 'system', 'system', 'INIT_001');

-- 租户管理员角色
INSERT INTO HUB_AUTH_ROLE (roleId, tenantId, roleName, roleDescription, dataScope, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('ROLE_TENANT_ADMIN', 'default', '租户管理员', '租户内所有权限的管理员', '{"type":"TENANT"}', 'Y', 'system', 'system', 'INIT_002');

-- 部门管理员角色
INSERT INTO HUB_AUTH_ROLE (roleId, tenantId, roleName, roleDescription, dataScope, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('ROLE_DEPT_ADMIN', 'default', '部门管理员', '部门内权限的管理员', '{"type":"DEPT"}', 'Y', 'system', 'system', 'INIT_003');

-- 普通用户角色
INSERT INTO HUB_AUTH_ROLE (roleId, tenantId, roleName, roleDescription, dataScope, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('ROLE_USER', 'default', '普通用户', '普通用户权限', '{"type":"SELF"}', 'Y', 'system', 'system', 'INIT_004');

-- =====================================================
-- 初始数据插入 - 模块权限资源
-- 资源类型层级关系：MODULE(模块) -> MENU(菜单) -> BUTTON(按钮)/API(接口)
-- =====================================================

-- 认证模块 (hub0001)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_AUTH_MODULE', 'default', '认证模块', 'hub0001', 'MODULE', '认证管理', 'Y', 'system', 'system', 'INIT_001');

-- 用户管理模块 (hub0002)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_USER_MODULE', 'default', '用户管理模块', 'hub0002', 'MODULE', '用户管理', 'Y', 'system', 'system', 'INIT_002');

-- 用户管理菜单
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourcePath, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_USER_MENU', 'default', '用户管理菜单', 'hub0002:user:menu', 'MENU', 'RES_USER_MODULE', '/user-management', '用户管理', 'Y', 'system', 'system', 'INIT_003');

INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourcePath, resourceMethod, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_USER_LIST', 'default', '用户列表', 'hub0002:user:list', 'API', 'RES_USER_MENU', '/api/hub0002/users', 'GET', '查看用户', 'Y', 'system', 'system', 'INIT_004');

INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourcePath, resourceMethod, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_USER_CREATE', 'default', '创建用户', 'hub0002:user:create', 'BUTTON', 'RES_USER_MENU', '/api/hub0002/users', 'POST', '创建用户', 'Y', 'system', 'system', 'INIT_005');

-- 定时任务模块 (hub0003)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_TIMER_MODULE', 'default', '定时任务模块', 'hub0003', 'MODULE', '定时任务', 'Y', 'system', 'system', 'INIT_006');

-- 网关实例模块 (hub0020)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_GW_INSTANCE_MODULE', 'default', '网关实例模块', 'hub0020', 'MODULE', '网关实例', 'Y', 'system', 'system', 'INIT_007');

-- 网关配置模块 (hub0021)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_GW_CONFIG_MODULE', 'default', '网关配置模块', 'hub0021', 'MODULE', '网关配置', 'Y', 'system', 'system', 'INIT_008');

-- 代理配置模块 (hub0022)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_PROXY_CONFIG_MODULE', 'default', '代理配置模块', 'hub0022', 'MODULE', '代理配置', 'Y', 'system', 'system', 'INIT_009');

-- 网关日志模块 (hub0023)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_GW_LOG_MODULE', 'default', '网关日志模块', 'hub0023', 'MODULE', '网关日志', 'Y', 'system', 'system', 'INIT_010');

-- 服务分组模块 (hub0040)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_SERVICE_GROUP_MODULE', 'default', '服务分组模块', 'hub0040', 'MODULE', '服务分组', 'Y', 'system', 'system', 'INIT_011');

-- 服务注册模块 (hub0041)
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, displayName, builtInFlag, addWho, editWho, oprSeqFlag) 
VALUES ('RES_SERVICE_REGISTRY_MODULE', 'default', '服务注册模块', 'hub0041', 'MODULE', '服务注册', 'Y', 'system', 'system', 'INIT_012');
