-- =============================================
-- 表名：HUB_SERVICE_INSTANCE
-- 说明：服务中心监听配置表（gRPC 服务器配置）
-- 作者：System
-- 创建时间：2026-01-23
-- =============================================

CREATE TABLE HUB_SERVICE_INSTANCE (
    -- 主键和租户信息
    tenantId VARCHAR2(32) NOT NULL,
    instanceName VARCHAR2(100) NOT NULL,
    environment VARCHAR2(32) NOT NULL,
    
    -- 服务器类型和监听配置
    serverType VARCHAR2(32) DEFAULT 'GRPC' NOT NULL,
    listenAddress VARCHAR2(128) DEFAULT '0.0.0.0' NOT NULL,
    listenPort NUMBER(10) DEFAULT 12004 NOT NULL,
    
    -- gRPC 消息大小配置
    maxRecvMsgSize NUMBER(10) DEFAULT 16777216 NOT NULL,
    maxSendMsgSize NUMBER(10) DEFAULT 16777216 NOT NULL,
    
    -- gRPC Keep-Alive 配置
    keepAliveTime NUMBER(10) DEFAULT 30 NOT NULL,
    keepAliveTimeout NUMBER(10) DEFAULT 10 NOT NULL,
    keepAliveMinTime NUMBER(10) DEFAULT 15 NOT NULL,
    permitWithoutStream VARCHAR2(1) DEFAULT 'Y' NOT NULL,
    
    -- gRPC 连接管理配置
    maxConnectionIdle NUMBER(10) DEFAULT 0 NOT NULL,
    maxConnectionAge NUMBER(10) DEFAULT 0 NOT NULL,
    maxConnectionAgeGrace NUMBER(10) DEFAULT 20 NOT NULL,
    
    -- gRPC 功能开关
    enableReflection VARCHAR2(1) DEFAULT 'Y' NOT NULL,
    enableTLS VARCHAR2(1) DEFAULT 'N' NOT NULL,
    
    -- 证书配置 - 支持文件路径和数据库存储
    certStorageType VARCHAR2(20) DEFAULT 'FILE' NOT NULL,
    certFilePath VARCHAR2(255),
    keyFilePath VARCHAR2(255),
    certContent CLOB,
    keyContent CLOB,
    certChainContent CLOB,
    certPassword VARCHAR2(255),
    enableMTLS VARCHAR2(1) DEFAULT 'N' NOT NULL,
    
    -- 性能调优配置
    maxConcurrentStreams NUMBER(10) DEFAULT 250 NOT NULL,
    readBufferSize NUMBER(10) DEFAULT 32768 NOT NULL,
    writeBufferSize NUMBER(10) DEFAULT 32768 NOT NULL,
    
    -- 健康检查配置
    healthCheckInterval NUMBER(10) DEFAULT 30 NOT NULL,
    healthCheckTimeout NUMBER(10) DEFAULT 5 NOT NULL,
    
    -- 实例状态管理
    instanceStatus VARCHAR2(20) DEFAULT 'STOPPED' NOT NULL,
    statusMessage CLOB,
    lastStatusTime DATE,
    lastHealthCheckTime DATE,
    
    -- 通用字段
    addTime DATE DEFAULT SYSDATE NOT NULL,
    addWho VARCHAR2(32) NOT NULL,
    editTime DATE DEFAULT SYSDATE NOT NULL,
    editWho VARCHAR2(32) NOT NULL,
    oprSeqFlag VARCHAR2(32) NOT NULL,
    currentVersion NUMBER(10) DEFAULT 1 NOT NULL,
    activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL,
    noteText VARCHAR2(500),
    extProperty CLOB,
    
    -- 访问控制配置
    enableAuth VARCHAR2(1) DEFAULT 'N' NOT NULL,
    ipWhitelist CLOB,
    ipBlacklist CLOB,
    
    CONSTRAINT PK_SC_INST PRIMARY KEY (tenantId, instanceName, environment)
);

CREATE INDEX IDX_SC_INST_TENANT ON HUB_SERVICE_INSTANCE(tenantId);
CREATE INDEX IDX_SC_INST_INSTANCE ON HUB_SERVICE_INSTANCE(instanceName);
CREATE INDEX IDX_SC_INST_ENV ON HUB_SERVICE_INSTANCE(environment);
CREATE INDEX IDX_SC_INST_ACTIVE ON HUB_SERVICE_INSTANCE(activeFlag);
CREATE INDEX IDX_SC_INST_TYPE ON HUB_SERVICE_INSTANCE(serverType);
CREATE INDEX IDX_SC_INST_PORT ON HUB_SERVICE_INSTANCE(listenPort);
CREATE INDEX IDX_SC_INST_STATUS ON HUB_SERVICE_INSTANCE(instanceStatus);
CREATE INDEX IDX_SC_INST_HEALTH_CHECK ON HUB_SERVICE_INSTANCE(lastHealthCheckTime);

COMMENT ON TABLE HUB_SERVICE_INSTANCE IS '服务中心监听配置表（gRPC 服务器配置）';

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
    16777216,
    16777216,
    30,
    10,
    15,
    'Y',
    0,
    0,
    20,
    'Y',
    'N',
    'FILE',
    'N',
    250,
    32768,
    32768,
    30,
    5,
    'STOPPED',
    '实例已创建，等待启动',
    'N',
    NULL,
    NULL,
    'system',
    'system',
    'INIT',
    1,
    'Y',
    '开发环境默认配置（明文传输，启用反射，无认证，开放访问用于本地调试）'
);

