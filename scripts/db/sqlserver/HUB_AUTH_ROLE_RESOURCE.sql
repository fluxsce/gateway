-- SQL Server 方言，由 scripts/db/mysql/HUB_AUTH_ROLE_RESOURCE.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- =====================================================
-- 角色权限关联表 - 存储角色与权限资源的关联关系
-- =====================================================
IF OBJECT_ID(N'dbo.HUB_AUTH_ROLE_RESOURCE', N'U') IS NULL
CREATE TABLE HUB_AUTH_ROLE_RESOURCE (
  -- 主键和租户信息
  roleResourceId NVARCHAR(100) NOT NULL,
  tenantId NVARCHAR(32) NOT NULL,
  
  -- 关联信息
  roleId NVARCHAR(32) NOT NULL,
  resourceId NVARCHAR(100) NOT NULL,
  
  -- 权限控制
  permissionType NVARCHAR(20) NOT NULL DEFAULT N'ALLOW',
  grantedBy NVARCHAR(32) NOT NULL,
  grantedTime DATETIME2 NOT NULL DEFAULT GETDATE(),
  expireTime DATETIME2 DEFAULT NULL,
  
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
  PRIMARY KEY (tenantId, roleResourceId),
  CONSTRAINT IDX_AUTH_ROLE_RES_UNIQUE UNIQUE (tenantId, roleId, resourceId),
  INDEX IDX_AUTH_ROLE_RES_ROLE (tenantId, roleId),
  INDEX IDX_AUTH_ROLE_RES_RESOURCE (tenantId, resourceId),
  INDEX IDX_AUTH_ROLE_RES_TYPE (permissionType),
  INDEX IDX_AUTH_ROLE_RES_EXPIRE (expireTime)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN permissionType NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN grantedBy NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN noteText NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved1 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved2 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved3 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved4 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved5 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved6 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved7 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved8 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved9 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE_RESOURCE ALTER COLUMN reserved10 NVARCHAR(500) NULL;


-- =====================================================
-- 初始化角色权限关联数据
-- 登录权限按角色资源表精确匹配，不会把 MODULE 自动展开为按钮。
-- 超级管理员：授予资源目录中全部有效资源（含 GROUP/MODULE/MENU/BUTTON）。
-- 只读用户：仅授予导航资源以及查看/查询/重置筛选/刷新/返回，排除写操作。
-- hub0023:reset 为日志重发，不属于只读范围。
-- =====================================================

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  CONCAT(N'ROLE_RES_SUPER_ADMIN_', REPLACE(resourceId, N':', N'_')),
  tenantId,
  N'ROLE_SUPER_ADMIN',
  resourceId,
  N'ALLOW',
  N'system',
  GETDATE(),
  GETDATE(),
  N'system',
  GETDATE(),
  N'system',
  N'INIT_SA',
  1,
  N'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = N'default'
  AND activeFlag = N'Y'
  AND resourceStatus = N'Y';

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  CONCAT(N'ROLE_RES_VIEWER_', REPLACE(resourceId, N':', N'_')),
  tenantId,
  N'ROLE_VIEWER',
  resourceId,
  N'ALLOW',
  N'system',
  GETDATE(),
  GETDATE(),
  N'system',
  GETDATE(),
  N'system',
  N'INIT_VIEWER',
  1,
  N'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = N'default'
  AND activeFlag = N'Y'
  AND resourceStatus = N'Y'
  AND (
    resourceType IN (N'GROUP', N'MODULE', N'MENU')
    OR resourceCode LIKE N'%:view'
    OR resourceCode LIKE N'%:search'
    OR resourceCode LIKE N'%:resetQuery'
    OR resourceCode LIKE N'%:back'
    OR resourceCode LIKE N'%:refresh'
    OR (
      resourceCode LIKE N'%:reset'
      AND resourceCode <> N'hub0023:reset'
    )
  );

-- 审计日志导出对只读角色开放（仍记 EXPORT 审计）
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
)
SELECT
  CONCAT(N'ROLE_RES_VIEWER_', REPLACE(resourceId, N':', N'_')),
  tenantId,
  N'ROLE_VIEWER',
  resourceId,
  N'ALLOW',
  N'system',
  GETDATE(),
  GETDATE(),
  N'system',
  GETDATE(),
  N'system',
  N'INIT_VIEWER',
  1,
  N'Y'
FROM HUB_AUTH_RESOURCE
WHERE tenantId = N'default'
  AND resourceId = N'hub0004:export';
