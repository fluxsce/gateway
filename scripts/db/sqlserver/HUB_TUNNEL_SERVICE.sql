-- SQL Server 方言，由 scripts/db/mysql/HUB_TUNNEL_SERVICE.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_TUNNEL_SERVICE', N'U') IS NULL
CREATE TABLE HUB_TUNNEL_SERVICE (
                                      tunnelServiceId NVARCHAR(32) NOT NULL,
                                      tenantId NVARCHAR(32) NOT NULL,
                                      tunnelClientId NVARCHAR(32) NOT NULL,
                                      userId NVARCHAR(32) NOT NULL,
                                      serviceName NVARCHAR(100) NOT NULL,
                                      serviceDescription NVARCHAR(500) DEFAULT NULL,
                                      serviceType NVARCHAR(20) NOT NULL,
                                      localAddress NVARCHAR(50) NOT NULL DEFAULT N'127.0.0.1',
                                      localPort INT NOT NULL,
                                      remotePort INT DEFAULT NULL,
                                      customDomains NVARCHAR(MAX) DEFAULT NULL,
                                      subDomain NVARCHAR(100) DEFAULT NULL,
                                      httpUser NVARCHAR(50) DEFAULT NULL,
                                      httpPassword NVARCHAR(100) DEFAULT NULL,
                                      hostHeaderRewrite NVARCHAR(255) DEFAULT NULL,
                                      headers NVARCHAR(MAX) DEFAULT NULL,
                                      locations NVARCHAR(MAX) DEFAULT NULL,
                                      useEncryption NVARCHAR(1) NOT NULL DEFAULT N'N',
                                      useCompression NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                      secretKey NVARCHAR(100) DEFAULT NULL,
                                      bandwidthLimit NVARCHAR(20) DEFAULT NULL,
                                      maxConnections INT DEFAULT 100,
                                      healthCheckType NVARCHAR(20) DEFAULT NULL,
                                      healthCheckUrl NVARCHAR(255) DEFAULT NULL,
                                      serviceStatus NVARCHAR(20) NOT NULL DEFAULT N'active',
                                      registeredTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                      lastActiveTime DATETIME2 DEFAULT NULL,
                                      connectionCount INT DEFAULT 0,
                                      totalConnections BIGINT DEFAULT 0,
                                      totalTraffic BIGINT DEFAULT 0,
                                      serviceConfig NVARCHAR(MAX) DEFAULT NULL,
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
                                      CONSTRAINT PK_TUNNEL_SERVICE PRIMARY KEY (tunnelServiceId),
                                      CONSTRAINT IDX_TUNNEL_SVC_NAME UNIQUE (serviceName),
                                      INDEX IDX_TUNNEL_SVC_TENANT (tenantId),
                                      INDEX IDX_TUNNEL_SVC_CLIENT (tunnelClientId),
                                      INDEX IDX_TUNNEL_SVC_USER (userId),
                                      INDEX IDX_TUNNEL_SVC_TYPE (serviceType),
                                      INDEX IDX_TUNNEL_SVC_STATUS (serviceStatus),
                                      INDEX IDX_TUNNEL_SVC_PORT (remotePort),
                                      INDEX IDX_TUNNEL_SVC_DOMAIN (subDomain)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN serviceName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN serviceDescription NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN serviceType NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN localAddress NVARCHAR(50) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN subDomain NVARCHAR(100) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN httpUser NVARCHAR(50) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN httpPassword NVARCHAR(100) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN hostHeaderRewrite NVARCHAR(255) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN useEncryption NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN useCompression NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN secretKey NVARCHAR(100) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN bandwidthLimit NVARCHAR(20) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN healthCheckType NVARCHAR(20) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN healthCheckUrl NVARCHAR(255) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN serviceStatus NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN noteText NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved1 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved2 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved3 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved4 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved5 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved6 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved7 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved8 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved9 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVICE ALTER COLUMN reserved10 NVARCHAR(500) NULL;
