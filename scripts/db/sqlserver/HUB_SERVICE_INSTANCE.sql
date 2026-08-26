-- SQL Server 方言，由 scripts/db/mysql/HUB_SERVICE_INSTANCE.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
-- =============================================
-- 表名：HUB_SERVICE_INSTANCE
-- 说明：服务中心监听配置表（gRPC 服务器配置）
-- 作者：System
-- 创建时间：2026-01-23
-- =============================================

IF OBJECT_ID(N'dbo.HUB_SERVICE_INSTANCE', N'U') IS NULL
CREATE TABLE HUB_SERVICE_INSTANCE (
    -- 主键和租户信息
    tenantId                NVARCHAR(32)     NOT NULL,
    instanceName            NVARCHAR(100)    NOT NULL,
    environment             NVARCHAR(32)     NOT NULL,
    
    -- 服务器类型和监听配置
    serverType              NVARCHAR(32)     NOT NULL DEFAULT N'GRPC',
    listenAddress           NVARCHAR(128)    NOT NULL DEFAULT N'0.0.0.0',
    listenPort              INT             NOT NULL DEFAULT 12004,
    
    -- gRPC 消息大小配置
    maxRecvMsgSize          INT             NOT NULL DEFAULT 16777216,
    maxSendMsgSize          INT             NOT NULL DEFAULT 16777216,
    
    -- gRPC Keep-Alive 配置
    keepAliveTime           INT             NOT NULL DEFAULT 30,
    keepAliveTimeout        INT             NOT NULL DEFAULT 10,
    keepAliveMinTime        INT             NOT NULL DEFAULT 15,
    permitWithoutStream     NVARCHAR(1)      NOT NULL DEFAULT N'Y',
    
    -- gRPC 连接管理配置
    maxConnectionIdle       INT             NOT NULL DEFAULT 0,
    maxConnectionAge        INT             NOT NULL DEFAULT 0,
    maxConnectionAgeGrace   INT             NOT NULL DEFAULT 20,
    
    -- gRPC 功能开关
    enableReflection        NVARCHAR(1)      NOT NULL DEFAULT N'Y',
    enableTLS               NVARCHAR(1)      NOT NULL DEFAULT N'N',
    
    -- 证书配置 - 支持文件路径和数据库存储（参考 HUB_GW_INSTANCE）
    certStorageType         NVARCHAR(20)     NOT NULL DEFAULT N'FILE',
    certFilePath            NVARCHAR(255)    DEFAULT NULL,
    keyFilePath             NVARCHAR(255)    DEFAULT NULL,
    certContent             NVARCHAR(MAX)            DEFAULT NULL,
    keyContent              NVARCHAR(MAX)            DEFAULT NULL,
    certChainContent        NVARCHAR(MAX)            DEFAULT NULL,
    certPassword            NVARCHAR(255)    DEFAULT NULL,
    enableMTLS              NVARCHAR(1)      NOT NULL DEFAULT N'N',
    
    -- 性能调优配置
    maxConcurrentStreams    INT             NOT NULL DEFAULT 250,
    readBufferSize          INT             NOT NULL DEFAULT 32768,
    writeBufferSize         INT             NOT NULL DEFAULT 32768,
    
    -- 健康检查配置
    healthCheckInterval     INT             NOT NULL DEFAULT 30,
    healthCheckTimeout      INT             NOT NULL DEFAULT 5,
    
    -- 实例状态管理
    instanceStatus          NVARCHAR(20)     NOT NULL DEFAULT N'STOPPED',
    statusMessage           NVARCHAR(MAX)            DEFAULT NULL,
    lastStatusTime          DATETIME2        DEFAULT NULL,
    lastHealthCheckTime     DATETIME2        DEFAULT NULL,
    
    -- 通用字段
    addTime                 DATETIME2        NOT NULL DEFAULT GETDATE(),
    addWho                  NVARCHAR(32)     NOT NULL,
    editTime                DATETIME2        NOT NULL DEFAULT GETDATE(),
    editWho                 NVARCHAR(32)     NOT NULL,
    oprSeqFlag              NVARCHAR(32)     NOT NULL,
    currentVersion          INT             NOT NULL DEFAULT 1,
    activeFlag              NVARCHAR(1)      NOT NULL DEFAULT N'Y',
    noteText                NVARCHAR(500)    DEFAULT NULL,
    extProperty             NVARCHAR(MAX)            DEFAULT NULL,
    
    -- 访问控制配置
    enableAuth              NVARCHAR(1)      NOT NULL DEFAULT N'N',
    ipWhitelist             NVARCHAR(MAX)            DEFAULT NULL,
    ipBlacklist             NVARCHAR(MAX)            DEFAULT NULL,
    
    -- 主键和索引
    PRIMARY KEY (tenantId, instanceName, environment),
    INDEX IDX_SC_INST_TENANT (tenantId),
    INDEX IDX_SC_INST_INSTANCE (instanceName),
    INDEX IDX_SC_INST_ENV (environment),
    INDEX IDX_SC_INST_ACTIVE (activeFlag),
    INDEX IDX_SC_INST_TYPE (serverType),
    INDEX IDX_SC_INST_PORT (listenPort),
    INDEX IDX_SC_INST_STATUS (instanceStatus),
    INDEX IDX_SC_INST_HEALTH_CHECK (lastHealthCheckTime)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN instanceName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN environment NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN serverType NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN listenAddress NVARCHAR(128) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN permitWithoutStream NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN enableReflection NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN enableTLS NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN certStorageType NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN certFilePath NVARCHAR(255) NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN keyFilePath NVARCHAR(255) NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN certPassword NVARCHAR(255) NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN enableMTLS NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN instanceStatus NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN addWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN editWho NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN oprSeqFlag NVARCHAR(32) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN noteText NVARCHAR(500) NULL;
ALTER TABLE HUB_SERVICE_INSTANCE ALTER COLUMN enableAuth NVARCHAR(1) NOT NULL;


-- =============================================
-- 初始化默认配置（开发环境）
-- =============================================
INSERT INTO HUB_SERVICE_INSTANCE (
    tenantId,
    instanceName,
    environment,
    serverType,
    listenAddress,
    listenPort,
    maxRecvMsgSize,
    maxSendMsgSize,
    keepAliveTime,
    keepAliveTimeout,
    keepAliveMinTime,
    permitWithoutStream,
    maxConnectionIdle,
    maxConnectionAge,
    maxConnectionAgeGrace,
    enableReflection,
    enableTLS,
    certStorageType,
    enableMTLS,
    maxConcurrentStreams,
    readBufferSize,
    writeBufferSize,
    healthCheckInterval,
    healthCheckTimeout,
    instanceStatus,
    statusMessage,
    enableAuth,
    ipWhitelist,
    ipBlacklist,
    addWho,
    editWho,
    oprSeqFlag,
    currentVersion,
    activeFlag,
    noteText
) VALUES (
    N'default',
    N'service-center-grpc',
    N'DEVELOPMENT',
    N'GRPC',
    N'0.0.0.0',
    12004,
    16777216,           -- 16MB
    16777216,           -- 16MB
    30,                 -- 30秒
    10,                 -- 10秒
    15,                 -- 15秒
    N'Y',
    0,                  -- 无限制（服务中心需要长连接）
    0,                  -- 无限制（服务中心需要长连接）
    20,                 -- 20秒（优雅关闭）
    N'Y',                -- 启用反射（方便调试）
    N'N',                -- 开发环境不启用TLS
    N'FILE',             -- 文件存储证书
    N'N',                -- 单向认证
    250,                -- 250个并发流
    32768,              -- 32KB
    32768,              -- 32KB
    30,                 -- 30秒（健康检查间隔）
    5,                  -- 5秒（健康检查超时）
    N'STOPPED',          -- 初始状态为停止
    N'实例已创建，等待启动',  -- 初始状态消息
    N'N',                -- 开发环境不启用认证
    NULL,               -- 开发环境无 IP 白名单限制
    NULL,               -- 开发环境无 IP 黑名单
    N'system',
    N'system',
    N'INIT',
    1,
    N'Y',
    N'开发环境默认配置（明文传输，启用反射，无认证，开放访问用于本地调试）'
);
