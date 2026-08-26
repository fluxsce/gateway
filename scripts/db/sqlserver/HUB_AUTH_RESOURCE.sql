-- SQL Server 方言，由 scripts/db/mysql/HUB_AUTH_RESOURCE.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- =====================================================
-- 权限资源表 - 存储系统所有权限资源信息
-- =====================================================
IF OBJECT_ID(N'dbo.HUB_AUTH_RESOURCE', N'U') IS NULL
CREATE TABLE HUB_AUTH_RESOURCE (
  -- 主键和租户信息
  resourceId NVARCHAR(100) NOT NULL,
  tenantId NVARCHAR(32) NOT NULL,
  
  -- 资源基本信息
  resourceName NVARCHAR(100) NOT NULL,
  resourceCode NVARCHAR(100) NOT NULL,
  resourceType NVARCHAR(20) NOT NULL,
  resourcePath NVARCHAR(500) DEFAULT NULL,
  resourceMethod NVARCHAR(10) DEFAULT NULL,
  
  -- 层级关系
  parentResourceId NVARCHAR(100) DEFAULT NULL,
  resourceLevel INT NOT NULL DEFAULT 1,
  sortOrder INT NOT NULL DEFAULT 0,
  
  -- 显示信息
  displayName NVARCHAR(100) DEFAULT NULL,
  iconClass NVARCHAR(100) DEFAULT NULL,
  description NVARCHAR(500) DEFAULT NULL,
  language NVARCHAR(10) DEFAULT N'zh-CN',
  
  -- 状态信息
  resourceStatus NVARCHAR(1) NOT NULL DEFAULT N'Y',
  builtInFlag NVARCHAR(1) NOT NULL DEFAULT N'N',
  
  -- 通用字段
  addTime DATETIME2 NOT NULL DEFAULT GETDATE(),
  addWho NVARCHAR(32) NOT NULL,
  editTime DATETIME2 NOT NULL DEFAULT GETDATE(),
  editWho NVARCHAR(32) NOT NULL,
  oprSeqFlag NVARCHAR(32) NOT NULL,
  currentVersion INT NOT NULL DEFAULT 1,
  activeFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
  noteText NVARCHAR(500) DEFAULT NULL,
  extProperty NVARCHAR(MAX) DEFAULT NULL,
  reserved1 NVARCHAR(500) DEFAULT NULL,
  reserved2 NVARCHAR(500) DEFAULT NULL,
  reserved3 NVARCHAR(500) DEFAULT NULL,
  reserved4 NVARCHAR(500) DEFAULT NULL,
  reserved5 NVARCHAR(500) DEFAULT NULL,
  reserved6 NVARCHAR(500) DEFAULT NULL,
  reserved7 NVARCHAR(500) DEFAULT NULL,
  reserved8 NVARCHAR(500) DEFAULT NULL,
  reserved9 NVARCHAR(500) DEFAULT NULL,
  reserved10 NVARCHAR(500) DEFAULT NULL,
  
  -- 主键和索引
  PRIMARY KEY (tenantId, resourceId),
  CONSTRAINT IDX_AUTH_RES_CODE UNIQUE (tenantId, resourceCode),
  INDEX IDX_AUTH_RES_TYPE (resourceType),
  INDEX IDX_AUTH_RES_PARENT (parentResourceId),
  INDEX IDX_AUTH_RES_PATH (resourcePath),
  INDEX IDX_AUTH_RES_STATUS (resourceStatus),
  INDEX IDX_AUTH_RES_LEVEL (resourceLevel),
  INDEX IDX_AUTH_RES_SORT (sortOrder)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN resourceName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN resourceCode NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN resourceType NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN resourcePath NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN resourceMethod NVARCHAR(10) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN displayName NVARCHAR(100) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN iconClass NVARCHAR(100) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN description NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN language NVARCHAR(10) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN resourceStatus NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN builtInFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN noteText NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved1 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved2 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved3 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved4 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved5 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved6 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved7 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved8 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved9 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_RESOURCE ALTER COLUMN reserved10 NVARCHAR(500) NULL;



-- =====================================================
-- 初始化权限资源数据
-- 基于 staticRoutes.ts 中的路由配置
-- 层级结构：GROUP（分组）-> MODULE（模块）-> BUTTON（按钮）
-- =====================================================

-- =====================================================
-- 第一层：分组（GROUP）
-- =====================================================

-- 系统监控分组 (group0000)
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'group0000')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'group0000', N'default', N'系统监控', N'group0000', N'GROUP', 1, 1, N'HomeOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_GROUP_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'系统监控' WHERE tenantId = N'default' AND resourceId = N'group0000';


-- 系统设置分组 (group0001)
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'group0001')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'group0001', N'default', N'系统设置', N'group0001', N'GROUP', 1, 2, N'SettingsOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_GROUP_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'系统设置' WHERE tenantId = N'default' AND resourceId = N'group0001';


-- 网关管理分组 (group0020)
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'group0020')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'group0020', N'default', N'网关管理', N'group0020', N'GROUP', 1, 3, N'CloudOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_GROUP_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'网关管理' WHERE tenantId = N'default' AND resourceId = N'group0020';


-- 服务治理分组 (group0040)
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'group0040')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'group0040', N'default', N'服务治理', N'group0040', N'GROUP', 1, 4, N'GitNetworkOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_GROUP_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'服务治理' WHERE tenantId = N'default' AND resourceId = N'group0040';


-- 隧道管理分组 (group0060)
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'group0060')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'group0060', N'default', N'隧道管理', N'group0060', N'GROUP', 1, 5, N'SwapHorizontalOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_GROUP_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'隧道管理' WHERE tenantId = N'default' AND resourceId = N'group0060';


-- =====================================================
-- 第二层：模块（MODULE）
-- =====================================================

-- 系统监控模块 (hub0000) - 属于 group0000
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0000')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0000', N'default', N'系统监控', N'hub0000', N'MODULE', N'/dashboard', N'group0000', 2, 1, N'HomeOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'系统监控' WHERE tenantId = N'default' AND resourceId = N'hub0000';


-- 用户登录模块 (hub0001) - 不需要权限验证，独立模块
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0001')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0001', N'default', N'用户登录', N'hub0001', N'MODULE', N'/login', 1, 0, N'LogInOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'用户登录' WHERE tenantId = N'default' AND resourceId = N'hub0001';


-- 用户管理模块 (hub0002) - 属于 group0001
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002', N'default', N'用户管理', N'hub0002', N'MODULE', N'/system/userManagement', N'group0001', 2, 1, N'PeopleOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'用户管理' WHERE tenantId = N'default' AND resourceId = N'hub0002';


-- 角色管理模块 (hub0005) - 属于 group0001
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0005')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0005', N'default', N'角色管理', N'hub0005', N'MODULE', N'/system/roleManagement', N'group0001', 2, 2, N'PeopleCircleOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'角色管理' WHERE tenantId = N'default' AND resourceId = N'hub0005';


-- 权限资源管理模块 (hub0006) - 属于 group0001
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0006')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0006', N'default', N'权限资源管理', N'hub0006', N'MODULE', N'/system/resourceManagement', N'group0001', 2, 3, N'KeyOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'权限资源管理' WHERE tenantId = N'default' AND resourceId = N'hub0006';


-- 系统节点监控模块 (hub0007) - 属于 group0001
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0007')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0007', N'default', N'系统节点监控', N'hub0007', N'MODULE', N'/system/serverNodeManagement', N'group0001', 2, 4, N'HardwareChipOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'系统节点监控' WHERE tenantId = N'default' AND resourceId = N'hub0007';


-- 集群节点事件模块 (hub0008) - 属于 group0001
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008', N'default', N'集群节点事件', N'hub0008', N'MODULE', N'/system/clusterEventManagement', N'group0001', 2, 5, N'RadioOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'集群节点事件' WHERE tenantId = N'default' AND resourceId = N'hub0008';


-- 定时任务模块 (hub0003) - 属于 group0001，后端写接口鉴权用，无独立前端页
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003', N'default', N'定时任务', N'hub0003', N'MODULE', N'/system/taskScheduler', N'group0001', 2, 6, N'TimerOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'定时任务' WHERE tenantId = N'default' AND resourceId = N'hub0003';


-- 审计日志模块 (hub0004) - 属于 group0001
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004', N'default', N'审计日志', N'hub0004', N'MODULE', N'/system/auditLogManagement', N'group0001', 2, 7, N'ShieldCheckmarkOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'审计日志' WHERE tenantId = N'default' AND resourceId = N'hub0004';


-- 环境设置模块 (hub0009) - 属于 group0001
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0009')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0009', N'default', N'环境设置', N'hub0009', N'MODULE', N'/system/environmentSettings', N'group0001', 2, 8, N'OptionsOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'环境设置' WHERE tenantId = N'default' AND resourceId = N'hub0009';


-- 网关实例管理模块 (hub0020) - 属于 group0020
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020', N'default', N'实例管理', N'hub0020', N'MODULE', N'/gateway/gatewayInstanceManager', N'group0020', 2, 1, N'ServerOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'实例管理' WHERE tenantId = N'default' AND resourceId = N'hub0020';


-- 路由管理模块 (hub0021) - 属于 group0020
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021', N'default', N'路由管理', N'hub0021', N'MODULE', N'/gateway/routeManagement', N'group0020', 2, 2, N'GitNetworkOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'路由管理' WHERE tenantId = N'default' AND resourceId = N'hub0021';


-- 代理管理模块 (hub0022) - 属于 group0020
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022', N'default', N'代理管理', N'hub0022', N'MODULE', N'/gateway/proxyManagement', N'group0020', 2, 3, N'FlashOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'代理管理' WHERE tenantId = N'default' AND resourceId = N'hub0022';


-- 网关日志管理模块 (hub0023) - 属于 group0020
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0023')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0023', N'default', N'网关日志管理', N'hub0023', N'MODULE', N'/gateway/gatewayLogManagement', N'group0020', 2, 4, N'DocumentTextOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_013', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'网关日志管理' WHERE tenantId = N'default' AND resourceId = N'hub0023';


-- 服务中心实例管理模块 (hub0040) - 属于 group0040
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040', N'default', N'服务中心实例管理', N'hub0040', N'MODULE', N'/serviceGovernance/serviceCenterInstanceManager', N'group0040', 2, 1, N'ServerOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'服务中心实例管理' WHERE tenantId = N'default' AND resourceId = N'hub0040';


-- 命名空间管理模块 (hub0041) - 属于 group0040
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0041')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0041', N'default', N'命名空间管理', N'hub0041', N'MODULE', N'/serviceGovernance/namespaceManagement', N'group0040', 2, 2, N'FolderOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_021', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'命名空间管理' WHERE tenantId = N'default' AND resourceId = N'hub0041';


-- 服务列表模块 (hub0042) - 属于 group0040
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042', N'default', N'服务列表', N'hub0042', N'MODULE', N'/serviceGovernance/serviceList', N'group0040', 2, 3, N'BarChartOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'服务列表' WHERE tenantId = N'default' AND resourceId = N'hub0042';


-- 配置中心模块 (hub0043) - 属于 group0040
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043', N'default', N'配置中心', N'hub0043', N'MODULE', N'/serviceGovernance/configManagement', N'group0040', 2, 4, N'CodeOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'配置中心' WHERE tenantId = N'default' AND resourceId = N'hub0043';


-- 隧道服务器模块 (hub0060) - 属于 group0060
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060', N'default', N'隧道服务器', N'hub0060', N'MODULE', N'/tunnel/tunnelServerManagement', N'group0060', 2, 1, N'ServerOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'隧道服务器' WHERE tenantId = N'default' AND resourceId = N'hub0060';


-- 静态映射模块 (hub0061) - 属于 group0060
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061', N'default', N'静态映射', N'hub0061', N'MODULE', N'/tunnel/staticMappingManagement', N'group0060', 2, 2, N'GitNetworkOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'静态映射' WHERE tenantId = N'default' AND resourceId = N'hub0061';


-- 隧道客户端模块 (hub0062) - 属于 group0060
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062', N'default', N'隧道客户端', N'hub0062', N'MODULE', N'/tunnel/tunnelClientManagement', N'group0060', 2, 3, N'DesktopOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'隧道客户端' WHERE tenantId = N'default' AND resourceId = N'hub0062';


-- =====================================================
-- 第三层：按钮（BUTTON）
-- =====================================================

-- 用户管理模块 - 按钮资源 (hub0002)
-- 新增按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002:add', N'default', N'新增', N'hub0002:add', N'BUTTON', N'hub0002', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增' WHERE tenantId = N'default' AND resourceId = N'hub0002:add';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002:edit', N'default', N'编辑', N'hub0002:edit', N'BUTTON', N'hub0002', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0002:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002:delete', N'default', N'删除', N'hub0002:delete', N'BUTTON', N'hub0002', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0002:delete';


-- 重置密码按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002:resetPassword')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002:resetPassword', N'default', N'重置密码', N'hub0002:resetPassword', N'BUTTON', N'hub0002', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置密码' WHERE tenantId = N'default' AND resourceId = N'hub0002:resetPassword';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002:view', N'default', N'查看详情', N'hub0002:view', N'BUTTON', N'hub0002', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0002:view';


-- 用户授权按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002:roleAuth')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002:roleAuth', N'default', N'用户授权', N'hub0002:roleAuth', N'BUTTON', N'hub0002', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'用户授权' WHERE tenantId = N'default' AND resourceId = N'hub0002:roleAuth';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002:search', N'default', N'查询', N'hub0002:search', N'BUTTON', N'hub0002', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0002:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0002:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0002:reset', N'default', N'重置', N'hub0002:reset', N'BUTTON', N'hub0002', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_003_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0002:reset';


-- 角色管理模块 - 按钮资源 (hub0005)
-- 新增按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0005:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0005:add', N'default', N'新增角色', N'hub0005:add', N'BUTTON', N'hub0005', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_004_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增角色' WHERE tenantId = N'default' AND resourceId = N'hub0005:add';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0005:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0005:edit', N'default', N'编辑角色', N'hub0005:edit', N'BUTTON', N'hub0005', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_004_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑角色' WHERE tenantId = N'default' AND resourceId = N'hub0005:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0005:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0005:delete', N'default', N'删除角色', N'hub0005:delete', N'BUTTON', N'hub0005', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_004_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除角色' WHERE tenantId = N'default' AND resourceId = N'hub0005:delete';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0005:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0005:view', N'default', N'查看详情', N'hub0005:view', N'BUTTON', N'hub0005', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_004_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0005:view';


-- 角色授权按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0005:roleAuth')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0005:roleAuth', N'default', N'角色授权', N'hub0005:roleAuth', N'BUTTON', N'hub0005', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_004_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'角色授权' WHERE tenantId = N'default' AND resourceId = N'hub0005:roleAuth';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0005:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0005:search', N'default', N'查询', N'hub0005:search', N'BUTTON', N'hub0005', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_004_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0005:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0005:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0005:reset', N'default', N'重置', N'hub0005:reset', N'BUTTON', N'hub0005', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_004_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0005:reset';


-- 权限资源管理模块 - 按钮资源 (hub0006)
-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0006:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0006:view', N'default', N'查看详情', N'hub0006:view', N'BUTTON', N'hub0006', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_005_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0006:view';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0006:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0006:search', N'default', N'查询', N'hub0006:search', N'BUTTON', N'hub0006', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_005_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0006:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0006:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0006:reset', N'default', N'重置', N'hub0006:reset', N'BUTTON', N'hub0006', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_005_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0006:reset';


-- 新增按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0006:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0006:add', N'default', N'新增', N'hub0006:add', N'BUTTON', N'hub0006', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_005_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增' WHERE tenantId = N'default' AND resourceId = N'hub0006:add';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0006:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0006:edit', N'default', N'编辑', N'hub0006:edit', N'BUTTON', N'hub0006', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_005_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0006:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0006:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0006:delete', N'default', N'删除', N'hub0006:delete', N'BUTTON', N'hub0006', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_005_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0006:delete';


-- =====================================================
-- 系统节点监控模块 - 按钮资源 (hub0007)
-- =====================================================

-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0007:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0007:view', N'default', N'查看详情', N'hub0007:view', N'BUTTON', N'hub0007', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_006_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0007:view';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0007:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0007:search', N'default', N'查询', N'hub0007:search', N'BUTTON', N'hub0007', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_006_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0007:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0007:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0007:reset', N'default', N'重置', N'hub0007:reset', N'BUTTON', N'hub0007', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_006_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0007:reset';


-- =====================================================
-- 集群节点事件模块 - 按钮资源 (hub0008)
-- =====================================================

-- 事件列表分组
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-list', N'default', N'事件列表', N'hub0008:event-list', N'BUTTON', N'hub0008', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'事件列表' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list';


-- 事件列表 - 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-list:view', N'default', N'查看详情', N'hub0008:event-list:view', N'BUTTON', N'hub0008:event-list', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_001_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list:view';


-- 事件列表 - 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-list:search', N'default', N'查询', N'hub0008:event-list:search', N'BUTTON', N'hub0008:event-list', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_001_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list:search';


-- 事件列表 - 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-list:reset', N'default', N'重置', N'hub0008:event-list:reset', N'BUTTON', N'hub0008:event-list', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_001_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list:reset';


-- 事件列表 - 收起/展开处理列表按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list:toggleAckList')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-list:toggleAckList', N'default', N'收起/展开处理列表', N'hub0008:event-list:toggleAckList', N'BUTTON', N'hub0008:event-list', 4, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_001_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'收起/展开处理列表' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-list:toggleAckList';


-- ACK处理列表分组
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-ack')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-ack', N'default', N'ACK处理列表', N'hub0008:event-ack', N'BUTTON', N'hub0008', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'ACK处理列表' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-ack';


-- ACK处理列表 - 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-ack:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-ack:view', N'default', N'查看详情', N'hub0008:event-ack:view', N'BUTTON', N'hub0008:event-ack', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_002_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-ack:view';


-- ACK处理列表 - 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-ack:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-ack:search', N'default', N'查询', N'hub0008:event-ack:search', N'BUTTON', N'hub0008:event-ack', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_002_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-ack:search';


-- ACK处理列表 - 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0008:event-ack:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0008:event-ack:reset', N'default', N'重置', N'hub0008:event-ack:reset', N'BUTTON', N'hub0008:event-ack', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_007_002_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0008:event-ack:reset';


-- =====================================================
-- 定时任务模块 - 按钮资源 (hub0003)
-- =====================================================

-- 调度器管理
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:scheduler')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:scheduler', N'default', N'调度器管理', N'hub0003:scheduler', N'BUTTON', N'hub0003', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'调度器管理' WHERE tenantId = N'default' AND resourceId = N'hub0003:scheduler';


-- 新增调度器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:scheduler:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:scheduler:add', N'default', N'新增调度器', N'hub0003:scheduler:add', N'BUTTON', N'hub0003:scheduler', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_001_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增调度器' WHERE tenantId = N'default' AND resourceId = N'hub0003:scheduler:add';


-- 编辑调度器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:scheduler:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:scheduler:edit', N'default', N'编辑调度器', N'hub0003:scheduler:edit', N'BUTTON', N'hub0003:scheduler', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_001_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑调度器' WHERE tenantId = N'default' AND resourceId = N'hub0003:scheduler:edit';


-- 删除调度器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:scheduler:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:scheduler:delete', N'default', N'删除调度器', N'hub0003:scheduler:delete', N'BUTTON', N'hub0003:scheduler', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_001_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除调度器' WHERE tenantId = N'default' AND resourceId = N'hub0003:scheduler:delete';


-- 任务管理
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:task')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:task', N'default', N'任务管理', N'hub0003:task', N'BUTTON', N'hub0003', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'任务管理' WHERE tenantId = N'default' AND resourceId = N'hub0003:task';


-- 新增任务
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:task:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:task:add', N'default', N'新增任务', N'hub0003:task:add', N'BUTTON', N'hub0003:task', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_002_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增任务' WHERE tenantId = N'default' AND resourceId = N'hub0003:task:add';


-- 编辑任务
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:task:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:task:edit', N'default', N'编辑任务', N'hub0003:task:edit', N'BUTTON', N'hub0003:task', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_002_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑任务' WHERE tenantId = N'default' AND resourceId = N'hub0003:task:edit';


-- 删除任务
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:task:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:task:delete', N'default', N'删除任务', N'hub0003:task:delete', N'BUTTON', N'hub0003:task', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_002_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除任务' WHERE tenantId = N'default' AND resourceId = N'hub0003:task:delete';


-- 启动任务
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:task:start')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:task:start', N'default', N'启动任务', N'hub0003:task:start', N'BUTTON', N'hub0003:task', 4, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_002_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'启动任务' WHERE tenantId = N'default' AND resourceId = N'hub0003:task:start';


-- 停止任务
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:task:stop')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:task:stop', N'default', N'停止任务', N'hub0003:task:stop', N'BUTTON', N'hub0003:task', 4, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_002_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'停止任务' WHERE tenantId = N'default' AND resourceId = N'hub0003:task:stop';


-- 立即执行
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:task:trigger')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:task:trigger', N'default', N'立即执行', N'hub0003:task:trigger', N'BUTTON', N'hub0003:task', 4, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_002_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'立即执行' WHERE tenantId = N'default' AND resourceId = N'hub0003:task:trigger';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:view', N'default', N'查看详情', N'hub0003:view', N'BUTTON', N'hub0003', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0003:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:search', N'default', N'查询', N'hub0003:search', N'BUTTON', N'hub0003', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0003:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0003:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0003:reset', N'default', N'重置', N'hub0003:reset', N'BUTTON', N'hub0003', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_008_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0003:reset';


-- =====================================================
-- 审计日志模块 - 按钮资源 (hub0004)
-- =====================================================

-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004:view', N'default', N'查看详情', N'hub0004:view', N'BUTTON', N'hub0004', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0004:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004:search', N'default', N'查询', N'hub0004:search', N'BUTTON', N'hub0004', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0004:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004:reset', N'default', N'重置', N'hub0004:reset', N'BUTTON', N'hub0004', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0004:reset';


-- 导出
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0004:export')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0004:export', N'default', N'导出', N'hub0004:export', N'BUTTON', N'hub0004', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_009_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'导出' WHERE tenantId = N'default' AND resourceId = N'hub0004:export';


-- =====================================================
-- 环境设置模块 - 按钮资源 (hub0009)
-- =====================================================

IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0009:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0009:view', N'default', N'查看', N'hub0009:view', N'BUTTON', N'hub0009', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0009_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看' WHERE tenantId = N'default' AND resourceId = N'hub0009:view';


IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0009:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0009:edit', N'default', N'保存', N'hub0009:edit', N'BUTTON', N'hub0009', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_HUB0009_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'保存' WHERE tenantId = N'default' AND resourceId = N'hub0009:edit';


-- =====================================================
-- 网关实例管理模块 - 按钮资源 (hub0020)
-- =====================================================

-- 新增按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:add', N'default', N'新建实例', N'hub0020:add', N'BUTTON', N'hub0020', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建实例' WHERE tenantId = N'default' AND resourceId = N'hub0020:add';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:edit', N'default', N'编辑', N'hub0020:edit', N'BUTTON', N'hub0020', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0020:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:delete', N'default', N'删除', N'hub0020:delete', N'BUTTON', N'hub0020', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0020:delete';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:view', N'default', N'查看详情', N'hub0020:view', N'BUTTON', N'hub0020', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0020:view';


-- 启动按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:start')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:start', N'default', N'启动', N'hub0020:start', N'BUTTON', N'hub0020', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'启动' WHERE tenantId = N'default' AND resourceId = N'hub0020:start';


-- 停止按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:stop')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:stop', N'default', N'停止', N'hub0020:stop', N'BUTTON', N'hub0020', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'停止' WHERE tenantId = N'default' AND resourceId = N'hub0020:stop';


-- 全局配置分组（右键菜单中的分组项）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:globalConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:globalConfig', N'default', N'全局配置', N'hub0020:globalConfig', N'BUTTON', N'hub0020', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'全局配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:globalConfig';


-- IP访问控制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:ipAccessControl', N'default', N'IP访问控制', N'hub0020:ipAccessControl', N'BUTTON', N'hub0020:globalConfig', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'IP访问控制' WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl';


-- IP访问控制子权限（来源于 common002/ip-config 模块的操作）
-- 新建配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:ipAccessControl:add', N'default', N'新建配置', N'hub0020:ipAccessControl:add', N'BUTTON', N'hub0020:ipAccessControl', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_007_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:ipAccessControl:edit', N'default', N'编辑配置', N'hub0020:ipAccessControl:edit', N'BUTTON', N'hub0020:ipAccessControl', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_007_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:edit';


-- 删除配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:ipAccessControl:delete', N'default', N'删除配置', N'hub0020:ipAccessControl:delete', N'BUTTON', N'hub0020:ipAccessControl', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_007_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:ipAccessControl:view', N'default', N'查看详情', N'hub0020:ipAccessControl:view', N'BUTTON', N'hub0020:ipAccessControl', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_007_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:ipAccessControl:search', N'default', N'查询', N'hub0020:ipAccessControl:search', N'BUTTON', N'hub0020:ipAccessControl', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_007_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:ipAccessControl:reset', N'default', N'重置', N'hub0020:ipAccessControl:reset', N'BUTTON', N'hub0020:ipAccessControl', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_007_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0020:ipAccessControl:reset';


-- User-Agent访问控制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:userAgentAccessControl', N'default', N'User-Agent访问控制', N'hub0020:userAgentAccessControl', N'BUTTON', N'hub0020:globalConfig', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'User-Agent访问控制' WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl';


-- User-Agent访问控制子权限（来源于 common002/agent-config 模块的操作）
-- 新建配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:userAgentAccessControl:add', N'default', N'新建配置', N'hub0020:userAgentAccessControl:add', N'BUTTON', N'hub0020:userAgentAccessControl', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_008_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:userAgentAccessControl:edit', N'default', N'编辑配置', N'hub0020:userAgentAccessControl:edit', N'BUTTON', N'hub0020:userAgentAccessControl', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_008_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:edit';


-- 删除配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:userAgentAccessControl:delete', N'default', N'删除配置', N'hub0020:userAgentAccessControl:delete', N'BUTTON', N'hub0020:userAgentAccessControl', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_008_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:userAgentAccessControl:view', N'default', N'查看详情', N'hub0020:userAgentAccessControl:view', N'BUTTON', N'hub0020:userAgentAccessControl', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_008_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:userAgentAccessControl:search', N'default', N'查询', N'hub0020:userAgentAccessControl:search', N'BUTTON', N'hub0020:userAgentAccessControl', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_008_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:userAgentAccessControl:reset', N'default', N'重置', N'hub0020:userAgentAccessControl:reset', N'BUTTON', N'hub0020:userAgentAccessControl', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_008_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0020:userAgentAccessControl:reset';


-- API访问控制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:apiAccessControl', N'default', N'API访问控制', N'hub0020:apiAccessControl', N'BUTTON', N'hub0020:globalConfig', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'API访问控制' WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl';


-- API访问控制子权限（来源于 common002/api-config 模块的操作）
-- 新建配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:apiAccessControl:add', N'default', N'新建配置', N'hub0020:apiAccessControl:add', N'BUTTON', N'hub0020:apiAccessControl', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_009_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:apiAccessControl:edit', N'default', N'编辑配置', N'hub0020:apiAccessControl:edit', N'BUTTON', N'hub0020:apiAccessControl', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_009_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:edit';


-- 删除配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:apiAccessControl:delete', N'default', N'删除配置', N'hub0020:apiAccessControl:delete', N'BUTTON', N'hub0020:apiAccessControl', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_009_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:apiAccessControl:view', N'default', N'查看详情', N'hub0020:apiAccessControl:view', N'BUTTON', N'hub0020:apiAccessControl', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_009_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:apiAccessControl:search', N'default', N'查询', N'hub0020:apiAccessControl:search', N'BUTTON', N'hub0020:apiAccessControl', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_009_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:apiAccessControl:reset', N'default', N'重置', N'hub0020:apiAccessControl:reset', N'BUTTON', N'hub0020:apiAccessControl', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_009_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0020:apiAccessControl:reset';


-- 域名访问控制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:domainAccessControl', N'default', N'域名访问控制', N'hub0020:domainAccessControl', N'BUTTON', N'hub0020:globalConfig', 4, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'域名访问控制' WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl';


-- 域名访问控制子权限（来源于 common002/domain-config 模块的操作）
-- 新建配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:domainAccessControl:add', N'default', N'新建配置', N'hub0020:domainAccessControl:add', N'BUTTON', N'hub0020:domainAccessControl', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_010_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:domainAccessControl:edit', N'default', N'编辑配置', N'hub0020:domainAccessControl:edit', N'BUTTON', N'hub0020:domainAccessControl', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_010_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:edit';


-- 删除配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:domainAccessControl:delete', N'default', N'删除配置', N'hub0020:domainAccessControl:delete', N'BUTTON', N'hub0020:domainAccessControl', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_010_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:domainAccessControl:view', N'default', N'查看详情', N'hub0020:domainAccessControl:view', N'BUTTON', N'hub0020:domainAccessControl', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_010_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:domainAccessControl:search', N'default', N'查询', N'hub0020:domainAccessControl:search', N'BUTTON', N'hub0020:domainAccessControl', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_010_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:domainAccessControl:reset', N'default', N'重置', N'hub0020:domainAccessControl:reset', N'BUTTON', N'hub0020:domainAccessControl', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_010_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0020:domainAccessControl:reset';


-- 跨域配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:corsConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:corsConfig', N'default', N'跨域配置', N'hub0020:corsConfig', N'BUTTON', N'hub0020:globalConfig', 4, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_011', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'跨域配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:corsConfig';


-- 安全配置总表写入口（hubcommon002 add/edit/deleteSecurityConfig）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:securityConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:securityConfig', N'default', N'安全配置', N'hub0020:securityConfig', N'BUTTON', N'hub0020:globalConfig', 4, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_011_SEC', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'安全配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:securityConfig';


-- 认证配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:authConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:authConfig', N'default', N'认证配置', N'hub0020:authConfig', N'BUTTON', N'hub0020:globalConfig', 4, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_012', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'认证配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:authConfig';


-- 限流配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:rateLimitConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:rateLimitConfig', N'default', N'限流配置', N'hub0020:rateLimitConfig', N'BUTTON', N'hub0020:globalConfig', 4, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_013', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'限流配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:rateLimitConfig';


-- 限流配置子权限（来源于 common002/limit-config 模块的操作）
-- 新增配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:rateLimitConfig:create')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:rateLimitConfig:create', N'default', N'新增配置', N'hub0020:rateLimitConfig:create', N'BUTTON', N'hub0020:rateLimitConfig', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_013_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:rateLimitConfig:create';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:rateLimitConfig:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:rateLimitConfig:edit', N'default', N'编辑配置', N'hub0020:rateLimitConfig:edit', N'BUTTON', N'hub0020:rateLimitConfig', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_013_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:rateLimitConfig:edit';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:rateLimitConfig:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:rateLimitConfig:view', N'default', N'查看详情', N'hub0020:rateLimitConfig:view', N'BUTTON', N'hub0020:rateLimitConfig', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_013_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0020:rateLimitConfig:view';


-- 日志配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:logConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:logConfig', N'default', N'日志配置', N'hub0020:logConfig', N'BUTTON', N'hub0020', 3, 14, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_014', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'日志配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:logConfig';


-- 网关重载按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:reload')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:reload', N'default', N'网关重载', N'hub0020:reload', N'BUTTON', N'hub0020', 3, 15, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_015', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'网关重载' WHERE tenantId = N'default' AND resourceId = N'hub0020:reload';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:search', N'default', N'查询', N'hub0020:search', N'BUTTON', N'hub0020', 3, 16, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_016', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0020:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:reset', N'default', N'重置', N'hub0020:reset', N'BUTTON', N'hub0020', 3, 17, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_017', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0020:reset';


-- 工具子菜单（实例树右键：导出/导入 Excel 配置；与前端 GExport/GImport 权限码 hub0020:export / hub0020:import 一致）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:tools')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:tools', N'default', N'工具', N'hub0020:tools', N'BUTTON', N'hub0020', 3, 18, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_018', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'工具' WHERE tenantId = N'default' AND resourceId = N'hub0020:tools';


IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:export')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:export', N'default', N'导出实例配置', N'hub0020:export', N'BUTTON', N'hub0020:tools', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_018_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'导出实例配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:export';


IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0020:import')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0020:import', N'default', N'导入实例配置', N'hub0020:import', N'BUTTON', N'hub0020:tools', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_010_018_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'导入实例配置' WHERE tenantId = N'default' AND resourceId = N'hub0020:import';


-- =====================================================
-- 路由管理模块 - 按钮资源 (hub0021)
-- =====================================================

-- Router配置按钮（实例树右键菜单）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:routerConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:routerConfig', N'default', N'Router配置', N'hub0021:routerConfig', N'BUTTON', N'hub0021', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'Router配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:routerConfig';


-- 全局过滤器配置按钮（实例树右键菜单）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:globalFilterConfig', N'default', N'全局过滤器配置', N'hub0021:globalFilterConfig', N'BUTTON', N'hub0021', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'全局过滤器配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig';


-- 全局过滤器配置子权限（来源于 hub0021/filter-config 模块的操作，filterScope = 'global'）
-- 新增过滤器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:globalFilterConfig:add', N'default', N'新增过滤器', N'hub0021:globalFilterConfig:add', N'BUTTON', N'hub0021:globalFilterConfig', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_002_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增过滤器' WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:add';


-- 编辑过滤器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:globalFilterConfig:edit', N'default', N'编辑过滤器', N'hub0021:globalFilterConfig:edit', N'BUTTON', N'hub0021:globalFilterConfig', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_002_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑过滤器' WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:edit';


-- 删除过滤器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:globalFilterConfig:delete', N'default', N'删除过滤器', N'hub0021:globalFilterConfig:delete', N'BUTTON', N'hub0021:globalFilterConfig', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_002_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除过滤器' WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:globalFilterConfig:view', N'default', N'查看详情', N'hub0021:globalFilterConfig:view', N'BUTTON', N'hub0021:globalFilterConfig', 4, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_002_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:globalFilterConfig:search', N'default', N'查询', N'hub0021:globalFilterConfig:search', N'BUTTON', N'hub0021:globalFilterConfig', 4, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_002_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:globalFilterConfig:reset', N'default', N'重置', N'hub0021:globalFilterConfig:reset', N'BUTTON', N'hub0021:globalFilterConfig', 4, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_002_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0021:globalFilterConfig:reset';


-- 新增路由按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:add', N'default', N'新增路由', N'hub0021:add', N'BUTTON', N'hub0021', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增路由' WHERE tenantId = N'default' AND resourceId = N'hub0021:add';


-- 删除路由按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:delete', N'default', N'删除', N'hub0021:delete', N'BUTTON', N'hub0021', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0021:delete';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:view', N'default', N'查看详情', N'hub0021:view', N'BUTTON', N'hub0021', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:view';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:edit', N'default', N'编辑', N'hub0021:edit', N'BUTTON', N'hub0021', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0021:edit';


-- 路由配置分组（右键菜单中的分组项）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:routeConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:routeConfig', N'default', N'路由配置', N'hub0021:routeConfig', N'BUTTON', N'hub0021', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'路由配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:routeConfig';


-- 路由断言配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:assertConfig', N'default', N'路由断言配置', N'hub0021:assertConfig', N'BUTTON', N'hub0021:routeConfig', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'路由断言配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig';


-- 路由断言配置子权限（来源于 hub0021/assert-config 模块的操作）
-- 新增断言
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:assertConfig:add', N'default', N'新增断言', N'hub0021:assertConfig:add', N'BUTTON', N'hub0021:assertConfig', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_008_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增断言' WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:add';


-- 编辑断言
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:assertConfig:edit', N'default', N'编辑断言', N'hub0021:assertConfig:edit', N'BUTTON', N'hub0021:assertConfig', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_008_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑断言' WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:edit';


-- 删除断言
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:assertConfig:delete', N'default', N'删除断言', N'hub0021:assertConfig:delete', N'BUTTON', N'hub0021:assertConfig', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_008_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除断言' WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:assertConfig:view', N'default', N'查看详情', N'hub0021:assertConfig:view', N'BUTTON', N'hub0021:assertConfig', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_008_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:assertConfig:search', N'default', N'查询', N'hub0021:assertConfig:search', N'BUTTON', N'hub0021:assertConfig', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_008_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:assertConfig:reset', N'default', N'重置', N'hub0021:assertConfig:reset', N'BUTTON', N'hub0021:assertConfig', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_008_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0021:assertConfig:reset';


-- 路由IP访问控制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:ipAccessControl', N'default', N'IP访问控制', N'hub0021:ipAccessControl', N'BUTTON', N'hub0021:routeConfig', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'IP访问控制' WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl';


-- 路由IP访问控制子权限（来源于 common002/ip-config 模块的操作）
-- 新建配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:ipAccessControl:add', N'default', N'新建配置', N'hub0021:ipAccessControl:add', N'BUTTON', N'hub0021:ipAccessControl', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_009_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:ipAccessControl:edit', N'default', N'编辑配置', N'hub0021:ipAccessControl:edit', N'BUTTON', N'hub0021:ipAccessControl', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_009_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:edit';


-- 删除配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:ipAccessControl:delete', N'default', N'删除配置', N'hub0021:ipAccessControl:delete', N'BUTTON', N'hub0021:ipAccessControl', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_009_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:ipAccessControl:view', N'default', N'查看详情', N'hub0021:ipAccessControl:view', N'BUTTON', N'hub0021:ipAccessControl', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_009_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:ipAccessControl:search', N'default', N'查询', N'hub0021:ipAccessControl:search', N'BUTTON', N'hub0021:ipAccessControl', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_009_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:ipAccessControl:reset', N'default', N'重置', N'hub0021:ipAccessControl:reset', N'BUTTON', N'hub0021:ipAccessControl', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_009_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0021:ipAccessControl:reset';


-- 路由User-Agent访问控制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:userAgentAccessControl', N'default', N'User-Agent访问控制', N'hub0021:userAgentAccessControl', N'BUTTON', N'hub0021:routeConfig', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'User-Agent访问控制' WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl';


-- 路由User-Agent访问控制子权限（来源于 common002/agent-config 模块的操作）
-- 新建配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:userAgentAccessControl:add', N'default', N'新建配置', N'hub0021:userAgentAccessControl:add', N'BUTTON', N'hub0021:userAgentAccessControl', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_010_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:userAgentAccessControl:edit', N'default', N'编辑配置', N'hub0021:userAgentAccessControl:edit', N'BUTTON', N'hub0021:userAgentAccessControl', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_010_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:edit';


-- 删除配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:userAgentAccessControl:delete', N'default', N'删除配置', N'hub0021:userAgentAccessControl:delete', N'BUTTON', N'hub0021:userAgentAccessControl', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_010_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:userAgentAccessControl:view', N'default', N'查看详情', N'hub0021:userAgentAccessControl:view', N'BUTTON', N'hub0021:userAgentAccessControl', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_010_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:userAgentAccessControl:search', N'default', N'查询', N'hub0021:userAgentAccessControl:search', N'BUTTON', N'hub0021:userAgentAccessControl', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_010_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:userAgentAccessControl:reset', N'default', N'重置', N'hub0021:userAgentAccessControl:reset', N'BUTTON', N'hub0021:userAgentAccessControl', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_010_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0021:userAgentAccessControl:reset';


-- 路由API访问控制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:apiAccessControl', N'default', N'API访问控制', N'hub0021:apiAccessControl', N'BUTTON', N'hub0021:routeConfig', 4, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_011', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'API访问控制' WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl';


-- 路由API访问控制子权限（来源于 common002/api-config 模块的操作）
-- 新建配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:apiAccessControl:add', N'default', N'新建配置', N'hub0021:apiAccessControl:add', N'BUTTON', N'hub0021:apiAccessControl', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_011_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:apiAccessControl:edit', N'default', N'编辑配置', N'hub0021:apiAccessControl:edit', N'BUTTON', N'hub0021:apiAccessControl', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_011_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:edit';


-- 删除配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:apiAccessControl:delete', N'default', N'删除配置', N'hub0021:apiAccessControl:delete', N'BUTTON', N'hub0021:apiAccessControl', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_011_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:apiAccessControl:view', N'default', N'查看详情', N'hub0021:apiAccessControl:view', N'BUTTON', N'hub0021:apiAccessControl', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_011_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:apiAccessControl:search', N'default', N'查询', N'hub0021:apiAccessControl:search', N'BUTTON', N'hub0021:apiAccessControl', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_011_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:apiAccessControl:reset', N'default', N'重置', N'hub0021:apiAccessControl:reset', N'BUTTON', N'hub0021:apiAccessControl', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_011_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0021:apiAccessControl:reset';


-- 路由域名访问控制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:domainAccessControl', N'default', N'域名访问控制', N'hub0021:domainAccessControl', N'BUTTON', N'hub0021:routeConfig', 4, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_012', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'域名访问控制' WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl';


-- 路由域名访问控制子权限（来源于 common002/domain-config 模块的操作）
-- 新建配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:domainAccessControl:add', N'default', N'新建配置', N'hub0021:domainAccessControl:add', N'BUTTON', N'hub0021:domainAccessControl', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_012_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:domainAccessControl:edit', N'default', N'编辑配置', N'hub0021:domainAccessControl:edit', N'BUTTON', N'hub0021:domainAccessControl', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_012_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:edit';


-- 删除配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:domainAccessControl:delete', N'default', N'删除配置', N'hub0021:domainAccessControl:delete', N'BUTTON', N'hub0021:domainAccessControl', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_012_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:domainAccessControl:view', N'default', N'查看详情', N'hub0021:domainAccessControl:view', N'BUTTON', N'hub0021:domainAccessControl', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_012_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:domainAccessControl:search', N'default', N'查询', N'hub0021:domainAccessControl:search', N'BUTTON', N'hub0021:domainAccessControl', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_012_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:domainAccessControl:reset', N'default', N'重置', N'hub0021:domainAccessControl:reset', N'BUTTON', N'hub0021:domainAccessControl', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_012_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0021:domainAccessControl:reset';


-- 路由跨域配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:corsConfig', N'default', N'跨域配置', N'hub0021:corsConfig', N'BUTTON', N'hub0021:routeConfig', 4, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_013', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'跨域配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig';


-- 安全配置总表写入口（路由侧 hubcommon002）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:securityConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:securityConfig', N'default', N'安全配置', N'hub0021:securityConfig', N'BUTTON', N'hub0021:routeConfig', 4, 20, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_013_SEC', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'安全配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:securityConfig';


-- 路由跨域配置子权限（来源于 common002/cors-config 模块的操作）
-- 新增配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:corsConfig:add', N'default', N'新增配置', N'hub0021:corsConfig:add', N'BUTTON', N'hub0021:corsConfig', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_013_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:corsConfig:edit', N'default', N'编辑配置', N'hub0021:corsConfig:edit', N'BUTTON', N'hub0021:corsConfig', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_013_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig:edit';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:corsConfig:view', N'default', N'查看详情', N'hub0021:corsConfig:view', N'BUTTON', N'hub0021:corsConfig', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_013_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:corsConfig:search', N'default', N'查询', N'hub0021:corsConfig:search', N'BUTTON', N'hub0021:corsConfig', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_013_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:corsConfig:search';


-- 路由认证配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:authConfig', N'default', N'认证配置', N'hub0021:authConfig', N'BUTTON', N'hub0021:routeConfig', 4, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_014', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'认证配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig';


-- 路由认证配置子权限（来源于 common002/auth-config 模块的操作）
-- 新增配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:authConfig:add', N'default', N'新增配置', N'hub0021:authConfig:add', N'BUTTON', N'hub0021:authConfig', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_014_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:authConfig:edit', N'default', N'编辑配置', N'hub0021:authConfig:edit', N'BUTTON', N'hub0021:authConfig', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_014_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig:edit';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:authConfig:view', N'default', N'查看详情', N'hub0021:authConfig:view', N'BUTTON', N'hub0021:authConfig', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_014_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:authConfig:search', N'default', N'查询', N'hub0021:authConfig:search', N'BUTTON', N'hub0021:authConfig', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_014_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:authConfig:search';


-- 路由限流配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:rateLimitConfig', N'default', N'限流配置', N'hub0021:rateLimitConfig', N'BUTTON', N'hub0021:routeConfig', 4, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_015', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'限流配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig';


-- 路由限流配置子权限（来源于 common002/limit-config 模块的操作）
-- 新增配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:rateLimitConfig:add', N'default', N'新增配置', N'hub0021:rateLimitConfig:add', N'BUTTON', N'hub0021:rateLimitConfig', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_015_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig:add';


-- 编辑配置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:rateLimitConfig:edit', N'default', N'编辑配置', N'hub0021:rateLimitConfig:edit', N'BUTTON', N'hub0021:rateLimitConfig', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_015_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑配置' WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig:edit';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:rateLimitConfig:view', N'default', N'查看详情', N'hub0021:rateLimitConfig:view', N'BUTTON', N'hub0021:rateLimitConfig', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_015_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:rateLimitConfig:search', N'default', N'查询', N'hub0021:rateLimitConfig:search', N'BUTTON', N'hub0021:rateLimitConfig', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_015_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:rateLimitConfig:search';


-- 路由过滤器按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:filters')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:filters', N'default', N'路由过滤器', N'hub0021:filters', N'BUTTON', N'hub0021:routeConfig', 4, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_016', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'路由过滤器' WHERE tenantId = N'default' AND resourceId = N'hub0021:filters';


-- 路由过滤器子权限（来源于 hub0021/filter-config 模块的操作）
-- 新增过滤器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:filters:add', N'default', N'新增过滤器', N'hub0021:filters:add', N'BUTTON', N'hub0021:filters', 5, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_016_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增过滤器' WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:add';


-- 编辑过滤器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:filters:edit', N'default', N'编辑过滤器', N'hub0021:filters:edit', N'BUTTON', N'hub0021:filters', 5, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_016_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑过滤器' WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:edit';


-- 删除过滤器
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:filters:delete', N'default', N'删除过滤器', N'hub0021:filters:delete', N'BUTTON', N'hub0021:filters', 5, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_016_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除过滤器' WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:delete';


-- 查看详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:filters:view', N'default', N'查看详情', N'hub0021:filters:view', N'BUTTON', N'hub0021:filters', 5, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_016_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:view';


-- 查询
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:filters:search', N'default', N'查询', N'hub0021:filters:search', N'BUTTON', N'hub0021:filters', 5, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_016_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:search';


-- 重置
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:filters:reset', N'default', N'重置', N'hub0021:filters:reset', N'BUTTON', N'hub0021:filters', 5, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_016_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0021:filters:reset';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:search', N'default', N'查询', N'hub0021:search', N'BUTTON', N'hub0021', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_017', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0021:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:reset', N'default', N'重置', N'hub0021:reset', N'BUTTON', N'hub0021', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_018', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0021:reset';


-- 静态资源配置按钮（路由列表右键菜单）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0021:staticHostConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0021:staticHostConfig', N'default', N'静态资源', N'hub0021:staticHostConfig', N'BUTTON', N'hub0021', 3, 10, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_011_019', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'静态资源' WHERE tenantId = N'default' AND resourceId = N'hub0021:staticHostConfig';


-- =====================================================
-- 代理管理模块 - 按钮资源 (hub0022)
-- =====================================================

-- 代理配置按钮（实例树右键菜单）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:addProxy')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:addProxy', N'default', N'代理配置', N'hub0022:addProxy', N'BUTTON', N'hub0022', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'代理配置' WHERE tenantId = N'default' AND resourceId = N'hub0022:addProxy';


-- 新增服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:add', N'default', N'新增服务', N'hub0022:add', N'BUTTON', N'hub0022', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增服务' WHERE tenantId = N'default' AND resourceId = N'hub0022:add';


-- 删除服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:delete', N'default', N'删除', N'hub0022:delete', N'BUTTON', N'hub0022', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0022:delete';


-- 节点管理按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:manageNodes', N'default', N'节点管理', N'hub0022:manageNodes', N'BUTTON', N'hub0022', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'节点管理' WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes';


-- 节点管理子权限（来源于 hub0022 节点管理页面的操作）
-- 新增节点
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:manageNodes:add', N'default', N'新增节点', N'hub0022:manageNodes:add', N'BUTTON', N'hub0022:manageNodes', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_004_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增节点' WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:add';


-- 编辑节点
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:manageNodes:edit', N'default', N'编辑节点', N'hub0022:manageNodes:edit', N'BUTTON', N'hub0022:manageNodes', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_004_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑节点' WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:edit';


-- 删除节点
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:manageNodes:delete', N'default', N'删除节点', N'hub0022:manageNodes:delete', N'BUTTON', N'hub0022:manageNodes', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_004_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除节点' WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:delete';


-- 查看节点详情
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:manageNodes:view', N'default', N'查看节点详情', N'hub0022:manageNodes:view', N'BUTTON', N'hub0022:manageNodes', 4, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_004_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看节点详情' WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:view';


-- 查询节点
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:manageNodes:search', N'default', N'查询', N'hub0022:manageNodes:search', N'BUTTON', N'hub0022:manageNodes', 4, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_004_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:search';


-- 重置节点列表
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:manageNodes:reset', N'default', N'重置', N'hub0022:manageNodes:reset', N'BUTTON', N'hub0022:manageNodes', 4, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_004_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0022:manageNodes:reset';


-- 编辑服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:edit', N'default', N'编辑', N'hub0022:edit', N'BUTTON', N'hub0022', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0022:edit';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:search', N'default', N'查询', N'hub0022:search', N'BUTTON', N'hub0022', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0022:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:reset', N'default', N'重置', N'hub0022:reset', N'BUTTON', N'hub0022', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0022:reset';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:view', N'default', N'查看详情', N'hub0022:view', N'BUTTON', N'hub0022', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0022:view';


-- 熔断配置按钮（服务列表右键菜单）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:circuitBreaker')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:circuitBreaker', N'default', N'熔断配置', N'hub0022:circuitBreaker', N'BUTTON', N'hub0022', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'熔断配置' WHERE tenantId = N'default' AND resourceId = N'hub0022:circuitBreaker';


-- 安全配置总表写入口（代理侧 hubcommon002）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0022:securityConfig')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0022:securityConfig', N'default', N'安全配置', N'hub0022:securityConfig', N'BUTTON', N'hub0022', 3, 10, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_012_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'安全配置' WHERE tenantId = N'default' AND resourceId = N'hub0022:securityConfig';


-- =====================================================
-- 网关日志管理模块 - 按钮资源 (hub0023)
-- =====================================================

-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0023:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0023:view', N'default', N'查看详情', N'hub0023:view', N'BUTTON', N'hub0023', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_013_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0023:view';


-- 批量重发按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0023:batchReset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0023:batchReset', N'default', N'批量重发', N'hub0023:batchReset', N'BUTTON', N'hub0023', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_013_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'批量重发' WHERE tenantId = N'default' AND resourceId = N'hub0023:batchReset';


-- 导出日志按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0023:export')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0023:export', N'default', N'导出日志', N'hub0023:export', N'BUTTON', N'hub0023', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_013_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'导出日志' WHERE tenantId = N'default' AND resourceId = N'hub0023:export';


-- 重发按钮（右键菜单）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0023:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0023:reset', N'default', N'重发', N'hub0023:reset', N'BUTTON', N'hub0023', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_013_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重发' WHERE tenantId = N'default' AND resourceId = N'hub0023:reset';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0023:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0023:search', N'default', N'查询', N'hub0023:search', N'BUTTON', N'hub0023', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_013_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0023:search';


-- 重置按钮（搜索表单）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0023:resetQuery')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0023:resetQuery', N'default', N'重置', N'hub0023:resetQuery', N'BUTTON', N'hub0023', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_013_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0023:resetQuery';


-- =====================================================
-- 服务中心实例管理模块 - 按钮资源 (hub0040)
-- =====================================================

-- 新建实例按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:add', N'default', N'新建实例', N'hub0040:add', N'BUTTON', N'hub0040', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建实例' WHERE tenantId = N'default' AND resourceId = N'hub0040:add';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:view', N'default', N'查看详情', N'hub0040:view', N'BUTTON', N'hub0040', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0040:view';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:edit', N'default', N'编辑', N'hub0040:edit', N'BUTTON', N'hub0040', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0040:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:delete', N'default', N'删除', N'hub0040:delete', N'BUTTON', N'hub0040', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0040:delete';


-- 启动实例按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:start')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:start', N'default', N'启动实例', N'hub0040:start', N'BUTTON', N'hub0040', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'启动实例' WHERE tenantId = N'default' AND resourceId = N'hub0040:start';


-- 停止实例按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:stop')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:stop', N'default', N'停止实例', N'hub0040:stop', N'BUTTON', N'hub0040', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'停止实例' WHERE tenantId = N'default' AND resourceId = N'hub0040:stop';


-- 重载配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:reload')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:reload', N'default', N'重载配置', N'hub0040:reload', N'BUTTON', N'hub0040', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重载配置' WHERE tenantId = N'default' AND resourceId = N'hub0040:reload';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:search', N'default', N'查询', N'hub0040:search', N'BUTTON', N'hub0040', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0040:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0040:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0040:reset', N'default', N'重置', N'hub0040:reset', N'BUTTON', N'hub0040', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_020_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0040:reset';


-- =====================================================
-- 命名空间管理模块 - 按钮资源 (hub0041)
-- =====================================================

-- 新建命名空间按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0041:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0041:add', N'default', N'新建命名空间', N'hub0041:add', N'BUTTON', N'hub0041', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_021_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建命名空间' WHERE tenantId = N'default' AND resourceId = N'hub0041:add';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0041:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0041:view', N'default', N'查看详情', N'hub0041:view', N'BUTTON', N'hub0041', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_021_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0041:view';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0041:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0041:edit', N'default', N'编辑', N'hub0041:edit', N'BUTTON', N'hub0041', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_021_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0041:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0041:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0041:delete', N'default', N'删除', N'hub0041:delete', N'BUTTON', N'hub0041', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_021_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0041:delete';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0041:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0041:search', N'default', N'查询', N'hub0041:search', N'BUTTON', N'hub0041', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_021_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0041:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0041:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0041:reset', N'default', N'重置', N'hub0041:reset', N'BUTTON', N'hub0041', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_021_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0041:reset';


-- =====================================================
-- 服务列表模块 - 按钮资源 (hub0042)
-- =====================================================

-- 命名空间列表区域权限 (hub0042-namespace)
-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:namespace:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:namespace:search', N'default', N'命名空间查询', N'hub0042:namespace:search', N'BUTTON', N'hub0042', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'命名空间查询' WHERE tenantId = N'default' AND resourceId = N'hub0042:namespace:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:namespace:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:namespace:reset', N'default', N'命名空间重置', N'hub0042:namespace:reset', N'BUTTON', N'hub0042', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'命名空间重置' WHERE tenantId = N'default' AND resourceId = N'hub0042:namespace:reset';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:namespace:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:namespace:view', N'default', N'命名空间查看详情', N'hub0042:namespace:view', N'BUTTON', N'hub0042', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'命名空间查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0042:namespace:view';


-- 服务列表区域权限 (hub0042)
-- 新建服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:add', N'default', N'新建服务', N'hub0042:add', N'BUTTON', N'hub0042', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建服务' WHERE tenantId = N'default' AND resourceId = N'hub0042:add';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:view', N'default', N'查看详情', N'hub0042:view', N'BUTTON', N'hub0042', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0042:view';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:edit', N'default', N'编辑', N'hub0042:edit', N'BUTTON', N'hub0042', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0042:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:delete', N'default', N'删除', N'hub0042:delete', N'BUTTON', N'hub0042', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0042:delete';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:search', N'default', N'查询', N'hub0042:search', N'BUTTON', N'hub0042', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0042:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:reset', N'default', N'重置', N'hub0042:reset', N'BUTTON', N'hub0042', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0042:reset';


-- 服务节点操作分组 (hub0042:nodeManagement)
-- 节点操作功能组
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:nodeManagement')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:nodeManagement', N'default', N'节点操作', N'hub0042:nodeManagement', N'MENU', N'hub0042', 3, 10, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'节点操作' WHERE tenantId = N'default' AND resourceId = N'hub0042:nodeManagement';


-- 节点操作子权限
-- 编辑节点按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:node:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:node:edit', N'default', N'编辑节点', N'hub0042:node:edit', N'BUTTON', N'hub0042:nodeManagement', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_010_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑节点' WHERE tenantId = N'default' AND resourceId = N'hub0042:node:edit';


-- 上线节点按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:node:online')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:node:online', N'default', N'上线节点', N'hub0042:node:online', N'BUTTON', N'hub0042:nodeManagement', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_010_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'上线节点' WHERE tenantId = N'default' AND resourceId = N'hub0042:node:online';


-- 下线节点按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:node:offline')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:node:offline', N'default', N'下线节点', N'hub0042:node:offline', N'BUTTON', N'hub0042:nodeManagement', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_010_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'下线节点' WHERE tenantId = N'default' AND resourceId = N'hub0042:node:offline';


-- 刷新节点列表按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0042:node:refresh')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0042:node:refresh', N'default', N'刷新节点列表', N'hub0042:node:refresh', N'BUTTON', N'hub0042:nodeManagement', 4, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_022_010_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'刷新节点列表' WHERE tenantId = N'default' AND resourceId = N'hub0042:node:refresh';


-- =====================================================
-- 配置中心模块 - 按钮资源 (hub0043)
-- =====================================================

-- 配置管理区域权限
-- 新建配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:add', N'default', N'新建配置', N'hub0043:add', N'BUTTON', N'hub0043', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新建配置' WHERE tenantId = N'default' AND resourceId = N'hub0043:add';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:view', N'default', N'查看详情', N'hub0043:view', N'BUTTON', N'hub0043', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0043:view';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:edit', N'default', N'编辑', N'hub0043:edit', N'BUTTON', N'hub0043', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0043:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:delete', N'default', N'删除', N'hub0043:delete', N'BUTTON', N'hub0043', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0043:delete';


-- 历史版本按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:history')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:history', N'default', N'历史版本', N'hub0043:history', N'BUTTON', N'hub0043', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'历史版本' WHERE tenantId = N'default' AND resourceId = N'hub0043:history';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:search', N'default', N'查询', N'hub0043:search', N'BUTTON', N'hub0043', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0043:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:reset', N'default', N'重置', N'hub0043:reset', N'BUTTON', N'hub0043', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0043:reset';


-- 配置历史管理分组 (hub0043:historyManagement)
-- 历史管理功能组
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:historyManagement')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:historyManagement', N'default', N'历史管理', N'hub0043:historyManagement', N'MENU', N'hub0043', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'历史管理' WHERE tenantId = N'default' AND resourceId = N'hub0043:historyManagement';


-- 历史管理子权限
-- 查看历史详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:history:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:history:view', N'default', N'查看历史详情', N'hub0043:history:view', N'BUTTON', N'hub0043:historyManagement', 4, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_008_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看历史详情' WHERE tenantId = N'default' AND resourceId = N'hub0043:history:view';


-- 回滚按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:history:rollback')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:history:rollback', N'default', N'回滚', N'hub0043:history:rollback', N'BUTTON', N'hub0043:historyManagement', 4, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_008_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'回滚' WHERE tenantId = N'default' AND resourceId = N'hub0043:history:rollback';


-- 历史查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:history:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:history:search', N'default', N'历史查询', N'hub0043:history:search', N'BUTTON', N'hub0043:historyManagement', 4, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_008_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'历史查询' WHERE tenantId = N'default' AND resourceId = N'hub0043:history:search';


-- 历史重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:history:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:history:reset', N'default', N'历史重置', N'hub0043:history:reset', N'BUTTON', N'hub0043:historyManagement', 4, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_008_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'历史重置' WHERE tenantId = N'default' AND resourceId = N'hub0043:history:reset';


-- 返回配置列表按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0043:history:back')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0043:history:back', N'default', N'返回配置列表', N'hub0043:history:back', N'BUTTON', N'hub0043:historyManagement', 4, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_023_008_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'返回配置列表' WHERE tenantId = N'default' AND resourceId = N'hub0043:history:back';


-- =====================================================
-- 隧道服务器管理模块 - 按钮资源 (hub0060)
-- =====================================================

-- 新增按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:add', N'default', N'新增服务器', N'hub0060:add', N'BUTTON', N'hub0060', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增服务器' WHERE tenantId = N'default' AND resourceId = N'hub0060:add';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:edit', N'default', N'编辑', N'hub0060:edit', N'BUTTON', N'hub0060', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0060:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:delete', N'default', N'删除', N'hub0060:delete', N'BUTTON', N'hub0060', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0060:delete';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:view', N'default', N'查看详情', N'hub0060:view', N'BUTTON', N'hub0060', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0060:view';


-- 启动服务器按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:start')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:start', N'default', N'启动服务器', N'hub0060:start', N'BUTTON', N'hub0060', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'启动服务器' WHERE tenantId = N'default' AND resourceId = N'hub0060:start';


-- 停止服务器按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:stop')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:stop', N'default', N'停止服务器', N'hub0060:stop', N'BUTTON', N'hub0060', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'停止服务器' WHERE tenantId = N'default' AND resourceId = N'hub0060:stop';


-- 重启服务器按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:restart')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:restart', N'default', N'重启服务器', N'hub0060:restart', N'BUTTON', N'hub0060', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重启服务器' WHERE tenantId = N'default' AND resourceId = N'hub0060:restart';


-- 客户端注册列表刷新按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:regist-client-list:refresh')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:regist-client-list:refresh', N'default', N'客户端注册列表刷新', N'hub0060:regist-client-list:refresh', N'BUTTON', N'hub0060', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'客户端注册列表刷新' WHERE tenantId = N'default' AND resourceId = N'hub0060:regist-client-list:refresh';


-- 服务注册列表刷新按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:regist-service-list:refresh')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:regist-service-list:refresh', N'default', N'服务注册列表刷新', N'hub0060:regist-service-list:refresh', N'BUTTON', N'hub0060', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'服务注册列表刷新' WHERE tenantId = N'default' AND resourceId = N'hub0060:regist-service-list:refresh';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:search', N'default', N'查询', N'hub0060:search', N'BUTTON', N'hub0060', 3, 10, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0060:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0060:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0060:reset', N'default', N'重置', N'hub0060:reset', N'BUTTON', N'hub0060', 3, 11, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_030_011', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0060:reset';


-- =====================================================
-- 静态映射管理模块 - 按钮资源 (hub0061)
-- =====================================================

-- 新增服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:add', N'default', N'新增服务', N'hub0061:add', N'BUTTON', N'hub0061', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增服务' WHERE tenantId = N'default' AND resourceId = N'hub0061:add';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:edit', N'default', N'编辑', N'hub0061:edit', N'BUTTON', N'hub0061', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0061:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:delete', N'default', N'删除', N'hub0061:delete', N'BUTTON', N'hub0061', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0061:delete';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:view', N'default', N'查看详情', N'hub0061:view', N'BUTTON', N'hub0061', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0061:view';


-- 启动服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:start')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:start', N'default', N'启动服务', N'hub0061:start', N'BUTTON', N'hub0061', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'启动服务' WHERE tenantId = N'default' AND resourceId = N'hub0061:start';


-- 停止服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:stop')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:stop', N'default', N'停止服务', N'hub0061:stop', N'BUTTON', N'hub0061', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'停止服务' WHERE tenantId = N'default' AND resourceId = N'hub0061:stop';


-- 重载配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:reload')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:reload', N'default', N'重载配置', N'hub0061:reload', N'BUTTON', N'hub0061', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重载配置' WHERE tenantId = N'default' AND resourceId = N'hub0061:reload';


-- 管理节点按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:nodes')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:nodes', N'default', N'管理节点', N'hub0061:nodes', N'BUTTON', N'hub0061', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'管理节点' WHERE tenantId = N'default' AND resourceId = N'hub0061:nodes';


-- =====================================================
-- 静态节点管理 - 按钮资源 (hub0061:static-nodes)
-- =====================================================

-- 新增节点按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:static-nodes:add', N'default', N'新增节点', N'hub0061:static-nodes:add', N'BUTTON', N'hub0061', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增节点' WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:add';


-- 编辑节点按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:static-nodes:edit', N'default', N'编辑节点', N'hub0061:static-nodes:edit', N'BUTTON', N'hub0061', 3, 10, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑节点' WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:edit';


-- 删除节点按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:static-nodes:delete', N'default', N'删除节点', N'hub0061:static-nodes:delete', N'BUTTON', N'hub0061', 3, 11, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_011', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除节点' WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:delete';


-- 查看节点详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:static-nodes:view', N'default', N'查看节点详情', N'hub0061:static-nodes:view', N'BUTTON', N'hub0061', 3, 12, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_012', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看节点详情' WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:view';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:search', N'default', N'查询', N'hub0061:search', N'BUTTON', N'hub0061', 3, 13, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_013', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0061:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:reset', N'default', N'重置', N'hub0061:reset', N'BUTTON', N'hub0061', 3, 14, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_014', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0061:reset';


-- 静态节点查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:static-nodes:search', N'default', N'查询节点', N'hub0061:static-nodes:search', N'BUTTON', N'hub0061', 3, 15, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_015', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询节点' WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:search';


-- 静态节点重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0061:static-nodes:reset', N'default', N'重置节点列表', N'hub0061:static-nodes:reset', N'BUTTON', N'hub0061', 3, 16, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_031_016', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置节点列表' WHERE tenantId = N'default' AND resourceId = N'hub0061:static-nodes:reset';


-- =====================================================
-- 隧道客户端管理模块 - 按钮资源 (hub0062:tunnel-client)
-- =====================================================

-- 新增客户端按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:tunnel-client:add', N'default', N'新增客户端', N'hub0062:tunnel-client:add', N'BUTTON', N'hub0062', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增客户端' WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:add';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:tunnel-client:edit', N'default', N'编辑', N'hub0062:tunnel-client:edit', N'BUTTON', N'hub0062', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:tunnel-client:delete', N'default', N'删除', N'hub0062:tunnel-client:delete', N'BUTTON', N'hub0062', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:delete';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:tunnel-client:view', N'default', N'查看详情', N'hub0062:tunnel-client:view', N'BUTTON', N'hub0062', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:view';


-- 连接按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:connect')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:tunnel-client:connect', N'default', N'连接', N'hub0062:tunnel-client:connect', N'BUTTON', N'hub0062', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'连接' WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:connect';


-- 断开连接按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:disconnect')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:tunnel-client:disconnect', N'default', N'断开连接', N'hub0062:tunnel-client:disconnect', N'BUTTON', N'hub0062', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'断开连接' WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:disconnect';


-- =====================================================
-- 隧道服务管理 - 按钮资源 (hub0062:service)
-- =====================================================

-- 新增服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:service:create')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:service:create', N'default', N'新增服务', N'hub0062:service:create', N'BUTTON', N'hub0062', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增服务' WHERE tenantId = N'default' AND resourceId = N'hub0062:service:create';


-- 编辑服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:service:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:service:edit', N'default', N'编辑服务', N'hub0062:service:edit', N'BUTTON', N'hub0062', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑服务' WHERE tenantId = N'default' AND resourceId = N'hub0062:service:edit';


-- 删除服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:service:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:service:delete', N'default', N'删除服务', N'hub0062:service:delete', N'BUTTON', N'hub0062', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除服务' WHERE tenantId = N'default' AND resourceId = N'hub0062:service:delete';


-- 查看服务详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:service:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:service:view', N'default', N'查看服务详情', N'hub0062:service:view', N'BUTTON', N'hub0062', 3, 10, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看服务详情' WHERE tenantId = N'default' AND resourceId = N'hub0062:service:view';


-- 注册服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:service:register')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:service:register', N'default', N'注册服务', N'hub0062:service:register', N'BUTTON', N'hub0062', 3, 11, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_011', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'注册服务' WHERE tenantId = N'default' AND resourceId = N'hub0062:service:register';


-- 注销服务按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:service:unregister')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:service:unregister', N'default', N'注销服务', N'hub0062:service:unregister', N'BUTTON', N'hub0062', 3, 12, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_012', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'注销服务' WHERE tenantId = N'default' AND resourceId = N'hub0062:service:unregister';


-- 隧道客户端查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:tunnel-client:search', N'default', N'查询', N'hub0062:tunnel-client:search', N'BUTTON', N'hub0062', 3, 13, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_013', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:search';


-- 隧道客户端重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:tunnel-client:reset', N'default', N'重置', N'hub0062:tunnel-client:reset', N'BUTTON', N'hub0062', 3, 14, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_014', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0062:tunnel-client:reset';


-- 隧道服务查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:service:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:service:search', N'default', N'查询服务', N'hub0062:service:search', N'BUTTON', N'hub0062', 3, 15, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_015', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询服务' WHERE tenantId = N'default' AND resourceId = N'hub0062:service:search';


-- 隧道服务重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0062:service:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0062:service:reset', N'default', N'重置服务列表', N'hub0062:service:reset', N'BUTTON', N'hub0062', 3, 16, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_032_016', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置服务列表' WHERE tenantId = N'default' AND resourceId = N'hub0062:service:reset';


-- =====================================================
-- 预警管理分组 (group0080)
-- =====================================================

-- 预警管理分组 (group0080)
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'group0080')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'group0080', N'default', N'预警管理', N'group0080', N'GROUP', 1, 6, N'NotificationsOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_GROUP_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'预警管理' WHERE tenantId = N'default' AND resourceId = N'group0080';


-- =====================================================
-- 预警管理模块（第二层：MODULE）
-- =====================================================

-- 预警服务配置模块 (hub0080) - 属于 group0080
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080', N'default', N'预警服务配置', N'hub0080', N'MODULE', N'/alert/alertConfigManagement', N'group0080', 2, 1, N'MailOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'预警服务配置' WHERE tenantId = N'default' AND resourceId = N'hub0080';


-- 预警模板管理模块 (hub0081) - 属于 group0080
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0081')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0081', N'default', N'预警模板管理', N'hub0081', N'MODULE', N'/alert/alertTemplateManagement', N'group0080', 2, 2, N'JournalOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_041', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'预警模板管理' WHERE tenantId = N'default' AND resourceId = N'hub0081';


-- 预警日志管理模块 (hub0082) - 属于 group0080
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, resourcePath, parentResourceId, resourceLevel, sortOrder, iconClass, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082', N'default', N'预警日志管理', N'hub0082', N'MODULE', N'/alert/alertLogManagement', N'group0080', 2, 3, N'DocumentTextOutline', N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'预警日志管理' WHERE tenantId = N'default' AND resourceId = N'hub0082';


-- =====================================================
-- 预警服务配置模块 - 按钮资源 (hub0080)
-- =====================================================

-- 新增按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:add', N'default', N'新增渠道', N'hub0080:add', N'BUTTON', N'hub0080', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增渠道' WHERE tenantId = N'default' AND resourceId = N'hub0080:add';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:view', N'default', N'查看详情', N'hub0080:view', N'BUTTON', N'hub0080', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0080:view';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:edit', N'default', N'编辑', N'hub0080:edit', N'BUTTON', N'hub0080', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0080:edit';


-- 复制按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:copy')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:copy', N'default', N'复制', N'hub0080:copy', N'BUTTON', N'hub0080', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'复制' WHERE tenantId = N'default' AND resourceId = N'hub0080:copy';


-- 重载配置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:reload')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:reload', N'default', N'重载配置', N'hub0080:reload', N'BUTTON', N'hub0080', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重载配置' WHERE tenantId = N'default' AND resourceId = N'hub0080:reload';


-- 设为默认按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:setDefault')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:setDefault', N'default', N'设为默认', N'hub0080:setDefault', N'BUTTON', N'hub0080', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'设为默认' WHERE tenantId = N'default' AND resourceId = N'hub0080:setDefault';


-- 预警测试按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:test')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:test', N'default', N'预警测试', N'hub0080:test', N'BUTTON', N'hub0080', 3, 7, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_007', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'预警测试' WHERE tenantId = N'default' AND resourceId = N'hub0080:test';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:delete', N'default', N'删除', N'hub0080:delete', N'BUTTON', N'hub0080', 3, 8, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_008', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0080:delete';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:search', N'default', N'查询', N'hub0080:search', N'BUTTON', N'hub0080', 3, 9, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_009', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0080:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0080:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0080:reset', N'default', N'重置', N'hub0080:reset', N'BUTTON', N'hub0080', 3, 10, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_040_010', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0080:reset';


-- =====================================================
-- 预警模板管理模块 - 按钮资源 (hub0081)
-- =====================================================

-- 新增按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0081:add')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0081:add', N'default', N'新增模板', N'hub0081:add', N'BUTTON', N'hub0081', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_041_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'新增模板' WHERE tenantId = N'default' AND resourceId = N'hub0081:add';


-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0081:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0081:view', N'default', N'查看详情', N'hub0081:view', N'BUTTON', N'hub0081', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_041_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0081:view';


-- 编辑按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0081:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0081:edit', N'default', N'编辑', N'hub0081:edit', N'BUTTON', N'hub0081', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_041_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'编辑' WHERE tenantId = N'default' AND resourceId = N'hub0081:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0081:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0081:delete', N'default', N'删除', N'hub0081:delete', N'BUTTON', N'hub0081', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_041_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0081:delete';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0081:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0081:search', N'default', N'查询', N'hub0081:search', N'BUTTON', N'hub0081', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_041_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0081:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0081:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0081:reset', N'default', N'重置', N'hub0081:reset', N'BUTTON', N'hub0081', 3, 6, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_041_006', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0081:reset';


-- =====================================================
-- 预警日志管理模块 - 按钮资源 (hub0082)
-- =====================================================

-- 查看详情按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:view')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:view', N'default', N'查看详情', N'hub0082:view', N'BUTTON', N'hub0082', 3, 1, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查看详情' WHERE tenantId = N'default' AND resourceId = N'hub0082:view';


-- 编辑按钮（更新发送状态/结果）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:edit')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:edit', N'default', N'更新日志', N'hub0082:edit', N'BUTTON', N'hub0082', 3, 5, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_005', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'更新日志' WHERE tenantId = N'default' AND resourceId = N'hub0082:edit';


-- 删除按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:delete')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:delete', N'default', N'删除', N'hub0082:delete', N'BUTTON', N'hub0082', 3, 2, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'删除' WHERE tenantId = N'default' AND resourceId = N'hub0082:delete';


-- 查询按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:search')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:search', N'default', N'查询', N'hub0082:search', N'BUTTON', N'hub0082', 3, 3, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_003', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'查询' WHERE tenantId = N'default' AND resourceId = N'hub0082:search';


-- 重置按钮
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_RESOURCE WHERE tenantId = N'default' AND resourceId = N'hub0082:reset')
INSERT INTO HUB_AUTH_RESOURCE (resourceId, tenantId, resourceName, resourceCode, resourceType, parentResourceId, resourceLevel, sortOrder, language, resourceStatus, builtInFlag, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'hub0082:reset', N'default', N'重置', N'hub0082:reset', N'BUTTON', N'hub0082', 3, 4, N'zh-CN', N'Y', N'Y', GETDATE(), N'system', GETDATE(), N'system', N'INIT_042_004', 1, N'Y')
ELSE
UPDATE HUB_AUTH_RESOURCE SET resourceName = N'重置' WHERE tenantId = N'default' AND resourceId = N'hub0082:reset';
