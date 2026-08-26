-- SQL Server 方言，由 scripts/db/mysql/HUB_LOGIN_LOG.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_LOGIN_LOG', N'U') IS NULL
CREATE TABLE HUB_LOGIN_LOG (
    logId           NVARCHAR(32)   NOT NULL,
    userId          NVARCHAR(32)   NOT NULL,
    tenantId        NVARCHAR(32)   NOT NULL,
    userName        NVARCHAR(50)   NOT NULL,
    loginTime       DATETIME2      NOT NULL DEFAULT GETDATE(),
    loginIp         NVARCHAR(128)  NOT NULL,
    loginLocation   NVARCHAR(255)  NULL,
    loginType       INT           NOT NULL DEFAULT 1,
    deviceType      NVARCHAR(50)   NULL,
    deviceInfo      NVARCHAR(MAX)          NULL,
    browserInfo     NVARCHAR(MAX)          NULL,
    osInfo          NVARCHAR(255)  NULL,
    loginStatus     NVARCHAR(1)    NOT NULL DEFAULT N'N',
    logoutTime      DATETIME2      NULL,
    sessionDuration INT           NULL,
    failReason      NVARCHAR(MAX)          NULL,
    addTime         DATETIME2      NOT NULL DEFAULT GETDATE(),
    addWho          NVARCHAR(32)   NOT NULL,
    editTime        DATETIME2      NOT NULL DEFAULT GETDATE(),
    editWho         NVARCHAR(32)   NOT NULL,
    oprSeqFlag      NVARCHAR(32)   NOT NULL,
    currentVersion  INT           NOT NULL DEFAULT 1,
    activeFlag      NVARCHAR(1)    NOT NULL DEFAULT N'Y',
    PRIMARY KEY (logId),
    INDEX IDX_LOGIN_USER (userId),
    INDEX IDX_LOGIN_TIME (loginTime),
    INDEX IDX_LOGIN_TENANT (tenantId),
    INDEX IDX_LOGIN_STATUS (loginStatus),
    INDEX IDX_LOGIN_TYPE (loginType)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN userName NVARCHAR(50) NOT NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN loginIp NVARCHAR(128) NOT NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN loginLocation NVARCHAR(255) NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN deviceType NVARCHAR(50) NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN osInfo NVARCHAR(255) NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN loginStatus NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_LOGIN_LOG ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
