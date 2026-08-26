-- SQL Server 方言，由 scripts/db/mysql/HUB_AUTH_ROLE.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- =====================================================
-- 角色表 - 存储系统角色信息和数据权限范围
-- =====================================================
IF OBJECT_ID(N'dbo.HUB_AUTH_ROLE', N'U') IS NULL
CREATE TABLE HUB_AUTH_ROLE (
  -- 主键和租户信息
  roleId NVARCHAR(32) NOT NULL,
  tenantId NVARCHAR(32) NOT NULL,
  
  -- 角色基本信息
  roleName NVARCHAR(100) NOT NULL,
  roleDescription NVARCHAR(500) DEFAULT NULL,
  
  -- 角色状态
  roleStatus NVARCHAR(1) NOT NULL DEFAULT N'Y',
  builtInFlag NVARCHAR(1) NOT NULL DEFAULT N'N',
  
  -- 数据权限范围
  dataScope NVARCHAR(MAX) DEFAULT NULL,
  
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
  PRIMARY KEY (tenantId, roleId),
  INDEX IDX_AUTH_ROLE_NAME (tenantId, roleName),
  INDEX IDX_AUTH_ROLE_STATUS (roleStatus)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN roleName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN roleDescription NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN roleStatus NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN builtInFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN noteText NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved1 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved2 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved3 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved4 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved5 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved6 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved7 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved8 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved9 NVARCHAR(500) NULL;
ALTER TABLE HUB_AUTH_ROLE ALTER COLUMN reserved10 NVARCHAR(500) NULL;


-- =====================================================
-- 初始化角色数据
-- =====================================================

-- 超级管理员角色
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_ROLE WHERE tenantId = N'default' AND roleId = N'ROLE_SUPER_ADMIN')
INSERT INTO HUB_AUTH_ROLE (roleId, tenantId, roleName, roleDescription, roleStatus, builtInFlag, dataScope, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'ROLE_SUPER_ADMIN', N'default', N'超级管理员', N'拥有系统所有权限的超级管理员', N'Y', N'Y', N'{"type":"ALL"}', GETDATE(), N'system', GETDATE(), N'system', N'INIT_001', 1, N'Y')
ELSE
UPDATE HUB_AUTH_ROLE SET roleName = N'超级管理员', roleDescription = N'拥有系统所有权限的超级管理员' WHERE tenantId = N'default' AND roleId = N'ROLE_SUPER_ADMIN';


-- 只读角色（最小权限：可进入模块并查询/查看，不可写）
IF NOT EXISTS (SELECT 1 FROM HUB_AUTH_ROLE WHERE tenantId = N'default' AND roleId = N'ROLE_VIEWER')
INSERT INTO HUB_AUTH_ROLE (roleId, tenantId, roleName, roleDescription, roleStatus, builtInFlag, dataScope, addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag) VALUES (N'ROLE_VIEWER', N'default', N'只读用户', N'仅拥有查询、查看和重置筛选的只读权限', N'Y', N'Y', N'{"type":"ALL"}', GETDATE(), N'system', GETDATE(), N'system', N'INIT_002', 1, N'Y')
ELSE
UPDATE HUB_AUTH_ROLE SET roleName = N'只读用户', roleDescription = N'仅拥有查询、查看和重置筛选的只读权限' WHERE tenantId = N'default' AND roleId = N'ROLE_VIEWER';
