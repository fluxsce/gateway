-- =============================================
-- 表名：HUB_SERVICE_INSTANCE
-- 说明：服务中心监听配置表（gRPC 服务器配置）
-- 作者：System
-- 创建时间：2026-01-23
-- =============================================

CREATE TABLE IF NOT EXISTS HUB_SERVICE_INSTANCE (
    -- 主键和租户信息
    tenantId TEXT NOT NULL,
    instanceName TEXT NOT NULL,
    environment TEXT NOT NULL,
    
    -- 服务器类型和监听配置
    serverType TEXT NOT NULL DEFAULT 'GRPC',
    listenAddress TEXT NOT NULL DEFAULT '0.0.0.0',
    listenPort INTEGER NOT NULL DEFAULT 12004,
    
    -- gRPC 消息大小配置
    maxRecvMsgSize INTEGER NOT NULL DEFAULT 16777216,
    maxSendMsgSize INTEGER NOT NULL DEFAULT 16777216,
    
    -- gRPC Keep-Alive 配置
    keepAliveTime INTEGER NOT NULL DEFAULT 30,
    keepAliveTimeout INTEGER NOT NULL DEFAULT 10,
    keepAliveMinTime INTEGER NOT NULL DEFAULT 15,
    permitWithoutStream TEXT NOT NULL DEFAULT 'Y',
    
    -- gRPC 连接管理配置
    maxConnectionIdle INTEGER NOT NULL DEFAULT 0,
    maxConnectionAge INTEGER NOT NULL DEFAULT 0,
    maxConnectionAgeGrace INTEGER NOT NULL DEFAULT 20,
    
    -- gRPC 功能开关
    enableReflection TEXT NOT NULL DEFAULT 'Y',
    enableTLS TEXT NOT NULL DEFAULT 'N',
    
    -- 证书配置 - 支持文件路径和数据库存储
    certStorageType TEXT NOT NULL DEFAULT 'FILE',
    certFilePath TEXT,
    keyFilePath TEXT,
    certContent TEXT,
    keyContent TEXT,
    certChainContent TEXT,
    certPassword TEXT,
    enableMTLS TEXT NOT NULL DEFAULT 'N',
    
    -- 性能调优配置
    maxConcurrentStreams INTEGER NOT NULL DEFAULT 250,
    readBufferSize INTEGER NOT NULL DEFAULT 32768,
    writeBufferSize INTEGER NOT NULL DEFAULT 32768,
    
    -- 健康检查配置
    healthCheckInterval INTEGER NOT NULL DEFAULT 30,
    healthCheckTimeout INTEGER NOT NULL DEFAULT 5,
    
    -- 实例状态管理
    instanceStatus TEXT NOT NULL DEFAULT 'STOPPED',
    statusMessage TEXT,
    lastStatusTime TEXT,
    lastHealthCheckTime TEXT,
    
    -- 通用字段
    addTime TEXT NOT NULL DEFAULT (datetime('now')),
    addWho TEXT NOT NULL,
    editTime TEXT NOT NULL DEFAULT (datetime('now')),
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    
    -- 访问控制配置
    enableAuth TEXT NOT NULL DEFAULT 'N',
    ipWhitelist TEXT,
    ipBlacklist TEXT,
    
    PRIMARY KEY (tenantId, instanceName, environment)
);

CREATE INDEX IF NOT EXISTS IDX_SC_INST_TENANT ON HUB_SERVICE_INSTANCE(tenantId);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_INSTANCE ON HUB_SERVICE_INSTANCE(instanceName);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_ENV ON HUB_SERVICE_INSTANCE(environment);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_ACTIVE ON HUB_SERVICE_INSTANCE(activeFlag);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_TYPE ON HUB_SERVICE_INSTANCE(serverType);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_PORT ON HUB_SERVICE_INSTANCE(listenPort);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_STATUS ON HUB_SERVICE_INSTANCE(instanceStatus);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_HEALTH_CHECK ON HUB_SERVICE_INSTANCE(lastHealthCheckTime);

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
    'default',
    'service-center-grpc',
    'DEVELOPMENT',
    'GRPC',
    '0.0.0.0',
    12004,
    16777216,           -- 16MB
    16777216,           -- 16MB
    30,                 -- 30秒
    10,                 -- 10秒
    15,                 -- 15秒
    'Y',
    0,                  -- 无限制（服务中心需要长连接）
    0,                  -- 无限制（服务中心需要长连接）
    20,                 -- 20秒（优雅关闭）
    'Y',                -- 启用反射（方便调试）
    'N',                -- 开发环境不启用TLS
    'FILE',             -- 文件存储证书
    'N',                -- 单向认证
    250,                -- 250个并发流
    32768,              -- 32KB
    32768,              -- 32KB
    30,                 -- 30秒（健康检查间隔）
    5,                  -- 5秒（健康检查超时）
    'STOPPED',          -- 初始状态为停止
    '实例已创建，等待启动',  -- 初始状态消息
    'N',                -- 开发环境不启用认证
    NULL,               -- 开发环境无 IP 白名单限制
    NULL,               -- 开发环境无 IP 黑名单
    'system',
    'system',
    'INIT',
    1,
    'Y',
    '开发环境默认配置（明文传输，启用反射，无认证，开放访问用于本地调试）'
);

