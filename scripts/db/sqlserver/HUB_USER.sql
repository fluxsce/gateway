-- SQL Server 方言，由 scripts/db/mysql/HUB_USER.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_USER', N'U') IS NULL
CREATE TABLE HUB_USER (
    userId          NVARCHAR(32)   NOT NULL,
    tenantId        NVARCHAR(32)   NOT NULL,
    userName        NVARCHAR(50)   NOT NULL,
    password        NVARCHAR(128)  NOT NULL,
    realName        NVARCHAR(50)   NOT NULL,
    deptId          NVARCHAR(32)   NOT NULL,
    email           NVARCHAR(255)  NULL,
    mobile          NVARCHAR(20)   NULL,
    avatar          NVARCHAR(MAX)      NULL,
    gender          INT           NULL     DEFAULT 0,
    statusFlag      NVARCHAR(1)    NOT NULL DEFAULT N'Y',
    deptAdminFlag   NVARCHAR(1)    NOT NULL DEFAULT N'N',
    tenantAdminFlag NVARCHAR(1)    NOT NULL DEFAULT N'N',
    userExpireDate  DATETIME2      NOT NULL,
    lastLoginTime   DATETIME2      NULL,
    lastLoginIp     NVARCHAR(128)  NULL,
    pwdUpdateTime   DATETIME2      NULL,
    pwdErrorCount   INT           NOT NULL DEFAULT 0,
    lockTime        DATETIME2      NULL,
    addTime         DATETIME2      NOT NULL DEFAULT GETDATE(),
    addWho          NVARCHAR(32)   NOT NULL,
    editTime        DATETIME2      NOT NULL DEFAULT GETDATE(),
    editWho         NVARCHAR(32)   NOT NULL,
    oprSeqFlag      NVARCHAR(32)   NOT NULL,
    currentVersion  INT           NOT NULL DEFAULT 1,
    activeFlag      NVARCHAR(1)    NOT NULL DEFAULT N'Y',
    noteText        NVARCHAR(MAX)          NULL,
    extProperty     NVARCHAR(MAX)          DEFAULT NULL,
    reserved1       NVARCHAR(500)  DEFAULT NULL,
    reserved2       NVARCHAR(500)  DEFAULT NULL,
    reserved3       NVARCHAR(500)  DEFAULT NULL,
    reserved4       NVARCHAR(500)  DEFAULT NULL,
    reserved5       NVARCHAR(500)  DEFAULT NULL,
    reserved6       NVARCHAR(500)  DEFAULT NULL,
    reserved7       NVARCHAR(500)  DEFAULT NULL,
    reserved8       NVARCHAR(500)  DEFAULT NULL,
    reserved9       NVARCHAR(500)  DEFAULT NULL,
    reserved10      NVARCHAR(500)  DEFAULT NULL,
    PRIMARY KEY (userId, tenantId),
    INDEX UK_USER_NAME_TENANT (userName, tenantId), -- 普通索引代替 UNIQUE KEY
    INDEX IDX_USER_TENANT (tenantId),
    INDEX IDX_USER_DEPT (deptId),
    INDEX IDX_USER_STATUS (statusFlag),
    INDEX IDX_USER_EMAIL (email),
    INDEX IDX_USER_MOBILE (mobile)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_USER ALTER COLUMN userName NVARCHAR(50) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN password NVARCHAR(128) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN realName NVARCHAR(50) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN email NVARCHAR(255) NULL;
ALTER TABLE HUB_USER ALTER COLUMN mobile NVARCHAR(20) NULL;
ALTER TABLE HUB_USER ALTER COLUMN statusFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN deptAdminFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN tenantAdminFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN lastLoginIp NVARCHAR(128) NULL;
ALTER TABLE HUB_USER ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved1 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved2 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved3 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved4 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved5 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved6 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved7 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved8 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved9 NVARCHAR(500) NULL;
ALTER TABLE HUB_USER ALTER COLUMN reserved10 NVARCHAR(500) NULL;


IF NOT EXISTS (SELECT 1 FROM HUB_USER WHERE userId = N'admin' AND tenantId = N'default')
INSERT INTO HUB_USER (userId, tenantId, userName, password, realName, deptId, email, mobile, avatar, gender, statusFlag, deptAdminFlag, tenantAdminFlag, userExpireDate, oprSeqFlag, currentVersion, activeFlag, addWho, editWho, noteText) VALUES (N'admin', N'default', N'admin', N'$2a$10$S9Yqyb9LI5PqAutYj.kR0OI/Zm7EcJSKbxKaLCThw8djqwqsPiDQi', N'系统管理员', N'D00000001', N'admin@example.com', N'13800000000', N'https://example.com/avatar.png', 1, N'Y', N'N', N'Y', DATEADD(YEAR, 5, GETDATE()), N'SEQFLAG_001', 1, N'Y', N'system', N'system', N'系统初始化管理员账号')
ELSE
UPDATE HUB_USER SET userName = N'admin', realName = N'系统管理员', noteText = N'系统初始化管理员账号' WHERE userId = N'admin' AND tenantId = N'default';


-- =====================================================
-- ALTER 变更语句：用户头像字段类型调整
-- 变更日期：2025-10-10
-- 变更原因：支持存储Base64编码的图片数据
-- 兼容性：向后兼容，现有URL数据不受影响
-- =====================================================
ALTER TABLE HUB_USER ALTER COLUMN avatar NVARCHAR(MAX) NULL;

-- 历史库升级：建表已执行过的不会再跑 CREATE，只补列。
IF COL_LENGTH(N'dbo.HUB_USER', N'mustChangePwd') IS NULL
ALTER TABLE HUB_USER ADD mustChangePwd NVARCHAR(1) NOT NULL DEFAULT N'N';

-- 种子管理员仍是默认口令时强制首次改密。已改密的现网账号口令哈希不同，不会被置 Y。
UPDATE HUB_USER
SET mustChangePwd = N'Y'
WHERE userId = N'admin'
  AND tenantId = N'default'
  AND password = N'$2a$10$S9Yqyb9LI5PqAutYj.kR0OI/Zm7EcJSKbxKaLCThw8djqwqsPiDQi';
