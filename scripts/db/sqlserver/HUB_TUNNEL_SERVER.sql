-- SQL Server 方言，由 scripts/db/mysql/HUB_TUNNEL_SERVER.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_TUNNEL_SERVER', N'U') IS NULL
CREATE TABLE HUB_TUNNEL_SERVER (
                                     tunnelServerId NVARCHAR(32) NOT NULL,
                                     tenantId NVARCHAR(32) NOT NULL,
                                     serverName NVARCHAR(100) NOT NULL,
                                     serverDescription NVARCHAR(500) DEFAULT NULL,
                                     controlAddress NVARCHAR(50) NOT NULL DEFAULT N'0.0.0.0',
                                     controlPort INT NOT NULL DEFAULT 7000,
                                     dashboardPort INT DEFAULT 7500,
                                     vhostHttpPort INT DEFAULT 80,
                                     vhostHttpsPort INT DEFAULT 443,
                                     maxClients INT NOT NULL DEFAULT 1000,
                                     tokenAuth NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                     authToken NVARCHAR(100) DEFAULT NULL,
                                     tlsEnable NVARCHAR(1) NOT NULL DEFAULT N'N',
                                     tlsCertFile NVARCHAR(255) DEFAULT NULL,
                                     tlsKeyFile NVARCHAR(255) DEFAULT NULL,
                                     heartbeatInterval INT NOT NULL DEFAULT 30,
                                     heartbeatTimeout INT NOT NULL DEFAULT 90,
                                     logLevel NVARCHAR(10) NOT NULL DEFAULT N'info',
                                     maxPortsPerClient INT DEFAULT 10,
                                     allowPorts NVARCHAR(MAX) DEFAULT NULL,
                                     serverStatus NVARCHAR(20) NOT NULL DEFAULT N'stopped',
                                     startTime DATETIME2 DEFAULT NULL,
                                     configVersion NVARCHAR(32) DEFAULT NULL,
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
                                     CONSTRAINT PK_TUNNEL_SERVER PRIMARY KEY (tunnelServerId),
                                     CONSTRAINT IDX_TUNNEL_SVR_NAME UNIQUE (serverName),
                                     INDEX IDX_TUNNEL_SVR_TENANT (tenantId),
                                     INDEX IDX_TUNNEL_SVR_CTRL (controlAddress, controlPort),
                                     INDEX IDX_TUNNEL_SVR_STATUS (serverStatus)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN serverName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN serverDescription NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN controlAddress NVARCHAR(50) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN tokenAuth NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN authToken NVARCHAR(100) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN tlsEnable NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN tlsCertFile NVARCHAR(255) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN tlsKeyFile NVARCHAR(255) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN logLevel NVARCHAR(10) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN serverStatus NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN configVersion NVARCHAR(32) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN noteText NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved1 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved2 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved3 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved4 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved5 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved6 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved7 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved8 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved9 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER ALTER COLUMN reserved10 NVARCHAR(500) NULL;
