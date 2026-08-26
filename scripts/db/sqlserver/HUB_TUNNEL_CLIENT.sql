-- SQL Server 方言，由 scripts/db/mysql/HUB_TUNNEL_CLIENT.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_TUNNEL_CLIENT', N'U') IS NULL
CREATE TABLE HUB_TUNNEL_CLIENT (
                                     tunnelClientId NVARCHAR(32) NOT NULL,
                                     tenantId NVARCHAR(32) NOT NULL,
                                     userId NVARCHAR(32) NOT NULL,
                                     clientName NVARCHAR(100) NOT NULL,
                                     clientDescription NVARCHAR(500) DEFAULT NULL,
                                     clientVersion NVARCHAR(20) DEFAULT NULL,
                                     operatingSystem NVARCHAR(50) DEFAULT NULL,
                                     clientIpAddress NVARCHAR(50) DEFAULT NULL,
                                     clientMacAddress NVARCHAR(20) DEFAULT NULL,
                                     serverAddress NVARCHAR(100) NOT NULL,
                                     serverPort INT NOT NULL DEFAULT 7000,
                                     authToken NVARCHAR(100) NOT NULL,
                                     tlsEnable NVARCHAR(1) NOT NULL DEFAULT N'N',
                                     autoReconnect NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                     maxRetries INT NOT NULL DEFAULT 5,
                                     retryInterval INT NOT NULL DEFAULT 20,
                                     heartbeatInterval INT NOT NULL DEFAULT 30,
                                     heartbeatTimeout INT NOT NULL DEFAULT 90,
                                     connectionStatus NVARCHAR(20) NOT NULL DEFAULT N'disconnected',
                                     lastConnectTime DATETIME2 DEFAULT NULL,
                                     lastDisconnectTime DATETIME2 DEFAULT NULL,
                                     totalConnectTime BIGINT DEFAULT 0,
                                     reconnectCount INT DEFAULT 0,
                                     serviceCount INT DEFAULT 0,
                                     lastHeartbeat DATETIME2 DEFAULT NULL,
                                     clientConfig NVARCHAR(MAX) DEFAULT NULL,
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
                                     CONSTRAINT PK_TUNNEL_CLIENT PRIMARY KEY (tunnelClientId),
                                     CONSTRAINT IDX_TUNNEL_CLIENT_NAME UNIQUE (clientName),
                                     INDEX IDX_TUNNEL_CLIENT_TENANT (tenantId),
                                     INDEX IDX_TUNNEL_CLIENT_USER (userId),
                                     INDEX IDX_TUNNEL_CLIENT_STATUS (connectionStatus),
                                     INDEX IDX_TUNNEL_CLIENT_IP (clientIpAddress),
                                     INDEX IDX_TUNNEL_CLIENT_HB (lastHeartbeat)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN clientName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN clientDescription NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN clientVersion NVARCHAR(20) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN operatingSystem NVARCHAR(50) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN clientIpAddress NVARCHAR(50) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN clientMacAddress NVARCHAR(20) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN serverAddress NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN authToken NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN tlsEnable NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN autoReconnect NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN connectionStatus NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN noteText NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved1 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved2 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved3 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved4 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved5 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved6 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved7 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved8 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved9 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_CLIENT ALTER COLUMN reserved10 NVARCHAR(500) NULL;
