-- SQL Server 方言，由 scripts/db/mysql/HUB_TUNNEL_SERVER_NODE.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_TUNNEL_SERVER_NODE', N'U') IS NULL
CREATE TABLE HUB_TUNNEL_SERVER_NODE (
                                          serverNodeId NVARCHAR(32) NOT NULL,
                                          tenantId NVARCHAR(32) NOT NULL,
                                          tunnelServerId NVARCHAR(32) NOT NULL,
                                          nodeName NVARCHAR(100) NOT NULL,
                                          nodeType NVARCHAR(20) NOT NULL DEFAULT N'static',
                                          proxyType NVARCHAR(20) NOT NULL,
                                          listenAddress NVARCHAR(50) NOT NULL DEFAULT N'0.0.0.0',
                                          listenPort INT NOT NULL,
                                          targetAddress NVARCHAR(50) NOT NULL,
                                          targetPort INT NOT NULL,
                                          customDomains NVARCHAR(MAX) DEFAULT NULL,
                                          subDomain NVARCHAR(100) DEFAULT NULL,
                                          httpUser NVARCHAR(50) DEFAULT NULL,
                                          httpPassword NVARCHAR(100) DEFAULT NULL,
                                          hostHeaderRewrite NVARCHAR(255) DEFAULT NULL,
                                          headers NVARCHAR(MAX) DEFAULT NULL,
                                          locations NVARCHAR(MAX) DEFAULT NULL,
                                          compression NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                          encryption NVARCHAR(1) NOT NULL DEFAULT N'N',
                                          secretKey NVARCHAR(100) DEFAULT NULL,
                                          healthCheckType NVARCHAR(20) DEFAULT N'tcp',
                                          healthCheckUrl NVARCHAR(255) DEFAULT NULL,
                                          healthCheckInterval INT DEFAULT 60,
                                          maxConnections INT DEFAULT 100,
                                          nodeStatus NVARCHAR(20) NOT NULL DEFAULT N'active',
                                          lastHealthCheck DATETIME2 DEFAULT NULL,
                                          connectionCount INT DEFAULT 0,
                                          totalConnections BIGINT DEFAULT 0,
                                          totalBytes BIGINT DEFAULT 0,
                                          createdTime DATETIME2 NOT NULL DEFAULT GETDATE(),
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
                                          CONSTRAINT PK_TUNNEL_SVR_NODE PRIMARY KEY (serverNodeId),
                                          CONSTRAINT IDX_TUNNEL_NODE_NAME UNIQUE (nodeName),
                                          CONSTRAINT IDX_TUNNEL_NODE_PORT UNIQUE (listenAddress, listenPort, proxyType),
                                          INDEX IDX_TUNNEL_NODE_TENANT (tenantId),
                                          INDEX IDX_TUNNEL_NODE_SERVER (tunnelServerId),
                                          INDEX IDX_TUNNEL_NODE_TYPE (nodeType, proxyType),
                                          INDEX IDX_TUNNEL_NODE_STATUS (nodeStatus),
                                          INDEX IDX_TUNNEL_NODE_HEALTH (lastHealthCheck)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN nodeName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN nodeType NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN proxyType NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN listenAddress NVARCHAR(50) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN targetAddress NVARCHAR(50) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN subDomain NVARCHAR(100) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN httpUser NVARCHAR(50) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN httpPassword NVARCHAR(100) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN hostHeaderRewrite NVARCHAR(255) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN compression NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN encryption NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN secretKey NVARCHAR(100) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN healthCheckType NVARCHAR(20) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN healthCheckUrl NVARCHAR(255) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN nodeStatus NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN noteText NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved1 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved2 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved3 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved4 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved5 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved6 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved7 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved8 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved9 NVARCHAR(500) NULL;
ALTER TABLE HUB_TUNNEL_SERVER_NODE ALTER COLUMN reserved10 NVARCHAR(500) NULL;
