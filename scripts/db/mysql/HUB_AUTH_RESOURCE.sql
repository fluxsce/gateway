-- =====================================================
-- 权限资源表 - 存储系统所有权限资源信息
-- =====================================================
CREATE TABLE IF NOT EXISTS `HUB_AUTH_RESOURCE` (
  -- 主键和租户信息
  `resourceId` VARCHAR(100) NOT NULL COMMENT '资源ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 资源基本信息
  `resourceName` VARCHAR(100) NOT NULL COMMENT '资源名称',
  `resourceCode` VARCHAR(100) NOT NULL COMMENT '资源编码，用于程序判断',
  `resourceType` VARCHAR(20) NOT NULL COMMENT '资源类型(MODULE:模块,MENU:菜单,BUTTON:按钮,API:接口) - 四级层次结构',
  `resourcePath` VARCHAR(500) DEFAULT NULL COMMENT '资源路径(菜单路径或API路径)',
  `resourceMethod` VARCHAR(10) DEFAULT NULL COMMENT '请求方法(GET,POST,PUT,DELETE等)',
  
  -- 层级关系
  `parentResourceId` VARCHAR(100) DEFAULT NULL COMMENT '父资源ID',
  `resourceLevel` INT NOT NULL DEFAULT 1 COMMENT '资源层级',
  `sortOrder` INT NOT NULL DEFAULT 0 COMMENT '排序顺序',
  
  -- 显示信息
  `displayName` VARCHAR(100) DEFAULT NULL COMMENT '显示名称',
  `iconClass` VARCHAR(100) DEFAULT NULL COMMENT '图标样式类',
  `description` VARCHAR(500) DEFAULT NULL COMMENT '资源描述',
  `language` VARCHAR(10) DEFAULT 'zh-CN' COMMENT '语言标识（如：zh-CN, en-US），用于多语言支持，默认zh-CN',
  
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
-- 初始化权限资源数据
-- 基于 staticRoutes.ts 中的路由配置
-- 层级结构：GROUP（分组）-> MODULE（模块）-> BUTTON（按钮）
-- =====================================================

-- =====================================================
-- 第一层：分组（GROUP）
-- =====================================================

-- 系统监控分组 (group0000)
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'group0000', 'default', '系统监控', 'group0000', 'GROUP',
  1, 1, 'HomeOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_GROUP_001', 1, 'Y'
);

-- 系统设置分组 (group0001)
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'group0001', 'default', '系统设置', 'group0001', 'GROUP',
  1, 2, 'SettingsOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_GROUP_002', 1, 'Y'
);

-- 网关管理分组 (group0020)
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'group0020', 'default', '网关管理', 'group0020', 'GROUP',
  1, 3, 'CloudOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_GROUP_003', 1, 'Y'
);

-- 服务治理分组 (group0040)
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'group0040', 'default', '服务治理', 'group0040', 'GROUP',
  1, 4, 'GitNetworkOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_GROUP_004', 1, 'Y'
);

-- 隧道管理分组 (group0060)
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'group0060', 'default', '隧道管理', 'group0060', 'GROUP',
  1, 5, 'SwapHorizontalOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_GROUP_005', 1, 'Y'
);

-- =====================================================
-- 第二层：模块（MODULE）
-- =====================================================

-- 系统监控模块 (hub0000) - 属于 group0000
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0000', 'default', '系统监控', 'hub0000', 'MODULE',
  '/dashboard', 'group0000', 2, 1, 'HomeOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_001', 1, 'Y'
);

-- 用户登录模块 (hub0001) - 不需要权限验证，独立模块
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0001', 'default', '用户登录', 'hub0001', 'MODULE',
  '/login', 1, 0, 'LogInOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_002', 1, 'Y'
);

-- 用户管理模块 (hub0002) - 属于 group0001
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002', 'default', '用户管理', 'hub0002', 'MODULE',
  '/system/userManagement', 'group0001', 2, 1, 'PeopleOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003', 1, 'Y'
);

-- 角色管理模块 (hub0005) - 属于 group0001
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0005', 'default', '角色管理', 'hub0005', 'MODULE',
  '/system/roleManagement', 'group0001', 2, 2, 'PeopleCircleOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_004', 1, 'Y'
);

-- 权限资源管理模块 (hub0006) - 属于 group0001
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0006', 'default', '权限资源管理', 'hub0006', 'MODULE',
  '/system/resourceManagement', 'group0001', 2, 3, 'KeyOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_005', 1, 'Y'
);

-- 网关实例管理模块 (hub0020) - 属于 group0020
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020', 'default', '实例管理', 'hub0020', 'MODULE',
  '/gateway/gatewayInstanceManager', 'group0020', 2, 1, 'ServerOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010', 1, 'Y'
);

-- 路由管理模块 (hub0021) - 属于 group0020
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021', 'default', '路由管理', 'hub0021', 'MODULE',
  '/gateway/routeManagement', 'group0020', 2, 2, 'GitNetworkOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011', 1, 'Y'
);

-- 代理管理模块 (hub0022) - 属于 group0020
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0022', 'default', '代理管理', 'hub0022', 'MODULE',
  '/gateway/proxyManagement', 'group0020', 2, 3, 'FlashOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_012', 1, 'Y'
);

-- 网关日志管理模块 (hub0023) - 属于 group0020
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0023', 'default', '网关日志管理', 'hub0023', 'MODULE',
  '/gateway/gatewayLogManagement', 'group0020', 2, 4, 'DocumentTextOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_013', 1, 'Y'
);

-- 命名空间管理模块 (hub0040) - 属于 group0040
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0040', 'default', '命名空间管理', 'hub0040', 'MODULE',
  '/serviceGovernance/namespaceManagement', 'group0040', 2, 1, 'LayersOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_020', 1, 'Y'
);

-- 服务注册管理模块 (hub0041) - 属于 group0040
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0041', 'default', '服务注册管理', 'hub0041', 'MODULE',
  '/serviceGovernance/serviceRegistryManagement', 'group0040', 2, 2, 'ListOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_021', 1, 'Y'
);

-- 服务监控模块 (hub0042) - 属于 group0040
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0042', 'default', '服务监控', 'hub0042', 'MODULE',
  '/serviceGovernance/serviceMonitoring', 'group0040', 2, 3, 'BarChartOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_022', 1, 'Y'
);

-- 隧道服务器模块 (hub0060) - 属于 group0060
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060', 'default', '隧道服务器', 'hub0060', 'MODULE',
  '/tunnel/tunnelServerManagement', 'group0060', 2, 1, 'ServerOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030', 1, 'Y'
);

-- 静态映射模块 (hub0061) - 属于 group0060
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061', 'default', '静态映射', 'hub0061', 'MODULE',
  '/tunnel/staticMappingManagement', 'group0060', 2, 2, 'GitNetworkOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031', 1, 'Y'
);

-- 隧道客户端模块 (hub0062) - 属于 group0060
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `resourcePath`, `parentResourceId`, `resourceLevel`, `sortOrder`, `iconClass`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062', 'default', '隧道客户端', 'hub0062', 'MODULE',
  '/tunnel/tunnelClientManagement', 'group0060', 2, 3, 'DesktopOutline', 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032', 1, 'Y'
);

-- =====================================================
-- 第三层：按钮（BUTTON）
-- =====================================================

-- 用户管理模块 - 按钮资源 (hub0002)
-- 新增按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002:add', 'default', '新增', 'hub0002:add', 'BUTTON',
  'hub0002', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003_001', 1, 'Y'
);

-- 编辑按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002:edit', 'default', '编辑', 'hub0002:edit', 'BUTTON',
  'hub0002', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003_002', 1, 'Y'
);

-- 删除按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002:delete', 'default', '删除', 'hub0002:delete', 'BUTTON',
  'hub0002', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003_003', 1, 'Y'
);

-- 重置密码按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002:resetPassword', 'default', '重置密码', 'hub0002:resetPassword', 'BUTTON',
  'hub0002', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003_004', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002:view', 'default', '查看详情', 'hub0002:view', 'BUTTON',
  'hub0002', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003_005', 1, 'Y'
);

-- 用户授权按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002:roleAuth', 'default', '用户授权', 'hub0002:roleAuth', 'BUTTON',
  'hub0002', 3, 8, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003_008', 1, 'Y'
);

-- 查询按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002:search', 'default', '查询', 'hub0002:search', 'BUTTON',
  'hub0002', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003_006', 1, 'Y'
);

-- 重置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0002:reset', 'default', '重置', 'hub0002:reset', 'BUTTON',
  'hub0002', 3, 7, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_003_007', 1, 'Y'
);

-- 角色管理模块 - 按钮资源 (hub0005)
-- 新增按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0005:add', 'default', '新增角色', 'hub0005:add', 'BUTTON',
  'hub0005', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_004_001', 1, 'Y'
);

-- 编辑按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0005:edit', 'default', '编辑角色', 'hub0005:edit', 'BUTTON',
  'hub0005', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_004_002', 1, 'Y'
);

-- 删除按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0005:delete', 'default', '删除角色', 'hub0005:delete', 'BUTTON',
  'hub0005', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_004_003', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0005:view', 'default', '查看详情', 'hub0005:view', 'BUTTON',
  'hub0005', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_004_004', 1, 'Y'
);

-- 角色授权按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0005:roleAuth', 'default', '角色授权', 'hub0005:roleAuth', 'BUTTON',
  'hub0005', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_004_005', 1, 'Y'
);

-- 查询按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0005:search', 'default', '查询', 'hub0005:search', 'BUTTON',
  'hub0005', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_004_006', 1, 'Y'
);

-- 重置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0005:reset', 'default', '重置', 'hub0005:reset', 'BUTTON',
  'hub0005', 3, 7, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_004_007', 1, 'Y'
);

-- 权限资源管理模块 - 按钮资源 (hub0006)
-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0006:view', 'default', '查看详情', 'hub0006:view', 'BUTTON',
  'hub0006', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_005_001', 1, 'Y'
);

-- 查询按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0006:search', 'default', '查询', 'hub0006:search', 'BUTTON',
  'hub0006', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_005_002', 1, 'Y'
);

-- 重置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0006:reset', 'default', '重置', 'hub0006:reset', 'BUTTON',
  'hub0006', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_005_003', 1, 'Y'
);

-- =====================================================
-- 网关实例管理模块 - 按钮资源 (hub0020)
-- =====================================================

-- 新增按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:add', 'default', '新建实例', 'hub0020:add', 'BUTTON',
  'hub0020', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_001', 1, 'Y'
);

-- 编辑按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:edit', 'default', '编辑', 'hub0020:edit', 'BUTTON',
  'hub0020', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_002', 1, 'Y'
);

-- 删除按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:delete', 'default', '删除', 'hub0020:delete', 'BUTTON',
  'hub0020', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_003', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:view', 'default', '查看详情', 'hub0020:view', 'BUTTON',
  'hub0020', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_004', 1, 'Y'
);

-- 启动按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:start', 'default', '启动', 'hub0020:start', 'BUTTON',
  'hub0020', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_005', 1, 'Y'
);

-- 停止按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:stop', 'default', '停止', 'hub0020:stop', 'BUTTON',
  'hub0020', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_006', 1, 'Y'
);

-- IP访问控制按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:ipAccessControl', 'default', 'IP访问控制', 'hub0020:ipAccessControl', 'BUTTON',
  'hub0020', 3, 7, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_007', 1, 'Y'
);

-- User-Agent访问控制按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:userAgentAccessControl', 'default', 'User-Agent访问控制', 'hub0020:userAgentAccessControl', 'BUTTON',
  'hub0020', 3, 8, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_008', 1, 'Y'
);

-- API访问控制按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:apiAccessControl', 'default', 'API访问控制', 'hub0020:apiAccessControl', 'BUTTON',
  'hub0020', 3, 9, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_009', 1, 'Y'
);

-- 域名访问控制按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:domainAccessControl', 'default', '域名访问控制', 'hub0020:domainAccessControl', 'BUTTON',
  'hub0020', 3, 10, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_010', 1, 'Y'
);

-- 跨域配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:corsConfig', 'default', '跨域配置', 'hub0020:corsConfig', 'BUTTON',
  'hub0020', 3, 11, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_011', 1, 'Y'
);

-- 认证配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:authConfig', 'default', '认证配置', 'hub0020:authConfig', 'BUTTON',
  'hub0020', 3, 12, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_012', 1, 'Y'
);

-- 限流配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:rateLimitConfig', 'default', '限流配置', 'hub0020:rateLimitConfig', 'BUTTON',
  'hub0020', 3, 13, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_013', 1, 'Y'
);

-- 日志配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:logConfig', 'default', '日志配置', 'hub0020:logConfig', 'BUTTON',
  'hub0020', 3, 14, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_014', 1, 'Y'
);

-- 网关重载按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:reload', 'default', '网关重载', 'hub0020:reload', 'BUTTON',
  'hub0020', 3, 15, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_015', 1, 'Y'
);

-- 查询按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:search', 'default', '查询', 'hub0020:search', 'BUTTON',
  'hub0020', 3, 16, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_016', 1, 'Y'
);

-- 重置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0020:reset', 'default', '重置', 'hub0020:reset', 'BUTTON',
  'hub0020', 3, 17, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_010_017', 1, 'Y'
);

-- =====================================================
-- 路由管理模块 - 按钮资源 (hub0021)
-- =====================================================

-- Router配置按钮（实例树右键菜单）
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:routerConfig', 'default', 'Router配置', 'hub0021:routerConfig', 'BUTTON',
  'hub0021', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_001', 1, 'Y'
);

-- 全局过滤器配置按钮（实例树右键菜单）
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:globalFilterConfig', 'default', '全局过滤器配置', 'hub0021:globalFilterConfig', 'BUTTON',
  'hub0021', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_002', 1, 'Y'
);

-- 新增路由按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:add', 'default', '新增路由', 'hub0021:add', 'BUTTON',
  'hub0021', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_003', 1, 'Y'
);

-- 删除路由按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:delete', 'default', '删除', 'hub0021:delete', 'BUTTON',
  'hub0021', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_004', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:view', 'default', '查看详情', 'hub0021:view', 'BUTTON',
  'hub0021', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_005', 1, 'Y'
);

-- 编辑按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:edit', 'default', '编辑', 'hub0021:edit', 'BUTTON',
  'hub0021', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_006', 1, 'Y'
);

-- 路由断言配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:assertConfig', 'default', '路由断言配置', 'hub0021:assertConfig', 'BUTTON',
  'hub0021', 3, 7, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_007', 1, 'Y'
);

-- 路由IP访问控制按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:ipAccessControl', 'default', 'IP访问控制', 'hub0021:ipAccessControl', 'BUTTON',
  'hub0021', 3, 8, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_008', 1, 'Y'
);

-- 路由User-Agent访问控制按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:userAgentAccessControl', 'default', 'User-Agent访问控制', 'hub0021:userAgentAccessControl', 'BUTTON',
  'hub0021', 3, 9, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_009', 1, 'Y'
);

-- 路由API访问控制按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:apiAccessControl', 'default', 'API访问控制', 'hub0021:apiAccessControl', 'BUTTON',
  'hub0021', 3, 10, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_010', 1, 'Y'
);

-- 路由域名访问控制按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:domainAccessControl', 'default', '域名访问控制', 'hub0021:domainAccessControl', 'BUTTON',
  'hub0021', 3, 11, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_011', 1, 'Y'
);

-- 路由跨域配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:corsConfig', 'default', '跨域配置', 'hub0021:corsConfig', 'BUTTON',
  'hub0021', 3, 12, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_012', 1, 'Y'
);

-- 路由认证配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:authConfig', 'default', '认证配置', 'hub0021:authConfig', 'BUTTON',
  'hub0021', 3, 13, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_013', 1, 'Y'
);

-- 路由限流配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:rateLimitConfig', 'default', '限流配置', 'hub0021:rateLimitConfig', 'BUTTON',
  'hub0021', 3, 14, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_014', 1, 'Y'
);

-- 路由过滤器按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:filters', 'default', '路由过滤器', 'hub0021:filters', 'BUTTON',
  'hub0021', 3, 15, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_015', 1, 'Y'
);

-- 查询按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:search', 'default', '查询', 'hub0021:search', 'BUTTON',
  'hub0021', 3, 16, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_016', 1, 'Y'
);

-- 重置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0021:reset', 'default', '重置', 'hub0021:reset', 'BUTTON',
  'hub0021', 3, 17, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_011_017', 1, 'Y'
);

-- =====================================================
-- 代理管理模块 - 按钮资源 (hub0022)
-- =====================================================

-- 代理配置按钮（实例树右键菜单）
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0022:addProxy', 'default', '代理配置', 'hub0022:addProxy', 'BUTTON',
  'hub0022', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_012_001', 1, 'Y'
);

-- 新增服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0022:add', 'default', '新增服务', 'hub0022:add', 'BUTTON',
  'hub0022', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_012_002', 1, 'Y'
);

-- 删除服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0022:delete', 'default', '删除', 'hub0022:delete', 'BUTTON',
  'hub0022', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_012_003', 1, 'Y'
);

-- 节点管理按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0022:manageNodes', 'default', '节点管理', 'hub0022:manageNodes', 'BUTTON',
  'hub0022', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_012_004', 1, 'Y'
);

-- 编辑服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0022:edit', 'default', '编辑', 'hub0022:edit', 'BUTTON',
  'hub0022', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_012_005', 1, 'Y'
);

-- 查询按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0022:search', 'default', '查询', 'hub0022:search', 'BUTTON',
  'hub0022', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_012_006', 1, 'Y'
);

-- 重置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0022:reset', 'default', '重置', 'hub0022:reset', 'BUTTON',
  'hub0022', 3, 7, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_012_007', 1, 'Y'
);

-- =====================================================
-- 网关日志管理模块 - 按钮资源 (hub0023)
-- =====================================================

-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0023:view', 'default', '查看详情', 'hub0023:view', 'BUTTON',
  'hub0023', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_013_001', 1, 'Y'
);

-- 批量重发按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0023:batchReset', 'default', '批量重发', 'hub0023:batchReset', 'BUTTON',
  'hub0023', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_013_002', 1, 'Y'
);

-- 导出日志按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0023:export', 'default', '导出日志', 'hub0023:export', 'BUTTON',
  'hub0023', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_013_003', 1, 'Y'
);

-- 重发按钮（右键菜单）
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0023:reset', 'default', '重发', 'hub0023:reset', 'BUTTON',
  'hub0023', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_013_004', 1, 'Y'
);

-- 查询按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0023:search', 'default', '查询', 'hub0023:search', 'BUTTON',
  'hub0023', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_013_005', 1, 'Y'
);

-- 重置按钮（搜索表单）
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0023:resetQuery', 'default', '重置', 'hub0023:resetQuery', 'BUTTON',
  'hub0023', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_013_006', 1, 'Y'
);

-- =====================================================
-- 隧道服务器管理模块 - 按钮资源 (hub0060)
-- =====================================================

-- 新增按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:add', 'default', '新增服务器', 'hub0060:add', 'BUTTON',
  'hub0060', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_001', 1, 'Y'
);

-- 编辑按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:edit', 'default', '编辑', 'hub0060:edit', 'BUTTON',
  'hub0060', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_002', 1, 'Y'
);

-- 删除按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:delete', 'default', '删除', 'hub0060:delete', 'BUTTON',
  'hub0060', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_003', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:view', 'default', '查看详情', 'hub0060:view', 'BUTTON',
  'hub0060', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_004', 1, 'Y'
);

-- 启动服务器按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:start', 'default', '启动服务器', 'hub0060:start', 'BUTTON',
  'hub0060', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_005', 1, 'Y'
);

-- 停止服务器按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:stop', 'default', '停止服务器', 'hub0060:stop', 'BUTTON',
  'hub0060', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_006', 1, 'Y'
);

-- 重启服务器按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:restart', 'default', '重启服务器', 'hub0060:restart', 'BUTTON',
  'hub0060', 3, 7, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_007', 1, 'Y'
);

-- 客户端注册列表刷新按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:regist-client-list:refresh', 'default', '客户端注册列表刷新', 'hub0060:regist-client-list:refresh', 'BUTTON',
  'hub0060', 3, 8, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_008', 1, 'Y'
);

-- 服务注册列表刷新按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0060:regist-service-list:refresh', 'default', '服务注册列表刷新', 'hub0060:regist-service-list:refresh', 'BUTTON',
  'hub0060', 3, 9, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_030_009', 1, 'Y'
);

-- =====================================================
-- 静态映射管理模块 - 按钮资源 (hub0061)
-- =====================================================

-- 新增服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:add', 'default', '新增服务', 'hub0061:add', 'BUTTON',
  'hub0061', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_001', 1, 'Y'
);

-- 编辑按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:edit', 'default', '编辑', 'hub0061:edit', 'BUTTON',
  'hub0061', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_002', 1, 'Y'
);

-- 删除按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:delete', 'default', '删除', 'hub0061:delete', 'BUTTON',
  'hub0061', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_003', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:view', 'default', '查看详情', 'hub0061:view', 'BUTTON',
  'hub0061', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_004', 1, 'Y'
);

-- 启动服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:start', 'default', '启动服务', 'hub0061:start', 'BUTTON',
  'hub0061', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_005', 1, 'Y'
);

-- 停止服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:stop', 'default', '停止服务', 'hub0061:stop', 'BUTTON',
  'hub0061', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_006', 1, 'Y'
);

-- 重载配置按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:reload', 'default', '重载配置', 'hub0061:reload', 'BUTTON',
  'hub0061', 3, 7, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_007', 1, 'Y'
);

-- 管理节点按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:nodes', 'default', '管理节点', 'hub0061:nodes', 'BUTTON',
  'hub0061', 3, 8, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_008', 1, 'Y'
);

-- =====================================================
-- 静态节点管理 - 按钮资源 (hub0061:static-nodes)
-- =====================================================

-- 新增节点按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:static-nodes:add', 'default', '新增节点', 'hub0061:static-nodes:add', 'BUTTON',
  'hub0061', 3, 9, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_009', 1, 'Y'
);

-- 编辑节点按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:static-nodes:edit', 'default', '编辑节点', 'hub0061:static-nodes:edit', 'BUTTON',
  'hub0061', 3, 10, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_010', 1, 'Y'
);

-- 删除节点按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:static-nodes:delete', 'default', '删除节点', 'hub0061:static-nodes:delete', 'BUTTON',
  'hub0061', 3, 11, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_011', 1, 'Y'
);

-- 查看节点详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0061:static-nodes:view', 'default', '查看节点详情', 'hub0061:static-nodes:view', 'BUTTON',
  'hub0061', 3, 12, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_031_012', 1, 'Y'
);

-- =====================================================
-- 隧道客户端管理模块 - 按钮资源 (hub0062:tunnel-client)
-- =====================================================

-- 新增客户端按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:tunnel-client:add', 'default', '新增客户端', 'hub0062:tunnel-client:add', 'BUTTON',
  'hub0062', 3, 1, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_001', 1, 'Y'
);

-- 编辑按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:tunnel-client:edit', 'default', '编辑', 'hub0062:tunnel-client:edit', 'BUTTON',
  'hub0062', 3, 2, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_002', 1, 'Y'
);

-- 删除按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:tunnel-client:delete', 'default', '删除', 'hub0062:tunnel-client:delete', 'BUTTON',
  'hub0062', 3, 3, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_003', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:tunnel-client:view', 'default', '查看详情', 'hub0062:tunnel-client:view', 'BUTTON',
  'hub0062', 3, 4, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_004', 1, 'Y'
);

-- 连接按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:tunnel-client:connect', 'default', '连接', 'hub0062:tunnel-client:connect', 'BUTTON',
  'hub0062', 3, 5, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_005', 1, 'Y'
);

-- 断开连接按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:tunnel-client:disconnect', 'default', '断开连接', 'hub0062:tunnel-client:disconnect', 'BUTTON',
  'hub0062', 3, 6, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_006', 1, 'Y'
);

-- =====================================================
-- 隧道服务管理 - 按钮资源 (hub0062:service)
-- =====================================================

-- 新增服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:service:create', 'default', '新增服务', 'hub0062:service:create', 'BUTTON',
  'hub0062', 3, 7, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_007', 1, 'Y'
);

-- 编辑服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:service:edit', 'default', '编辑服务', 'hub0062:service:edit', 'BUTTON',
  'hub0062', 3, 8, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_008', 1, 'Y'
);

-- 删除服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:service:delete', 'default', '删除服务', 'hub0062:service:delete', 'BUTTON',
  'hub0062', 3, 9, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_009', 1, 'Y'
);

-- 查看服务详情按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:service:view', 'default', '查看服务详情', 'hub0062:service:view', 'BUTTON',
  'hub0062', 3, 10, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_010', 1, 'Y'
);

-- 注册服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:service:register', 'default', '注册服务', 'hub0062:service:register', 'BUTTON',
  'hub0062', 3, 11, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_011', 1, 'Y'
);

-- 注销服务按钮
INSERT INTO `HUB_AUTH_RESOURCE` (
  `resourceId`, `tenantId`, `resourceName`, `resourceCode`, `resourceType`,
  `parentResourceId`, `resourceLevel`, `sortOrder`, `language`,
  `resourceStatus`, `builtInFlag`,
  `addTime`, `addWho`, `editTime`, `editWho`, `oprSeqFlag`, `currentVersion`, `activeFlag`
) VALUES (
  'hub0062:service:unregister', 'default', '注销服务', 'hub0062:service:unregister', 'BUTTON',
  'hub0062', 3, 12, 'zh-CN',
  'Y', 'Y',
  NOW(), 'system', NOW(), 'system', 'INIT_032_012', 1, 'Y'
);

