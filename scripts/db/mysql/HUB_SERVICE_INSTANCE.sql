-- =============================================
-- 表名：HUB_SERVICE_INSTANCE
-- 说明：服务中心监听配置表（gRPC 服务器配置）
-- 作者：System
-- 创建时间：2026-01-23
-- =============================================

CREATE TABLE `HUB_SERVICE_INSTANCE` (
    -- 主键和租户信息
    `tenantId`                VARCHAR(32)     NOT NULL COMMENT '租户ID，用于多租户数据隔离',
    `instanceName`            VARCHAR(100)    NOT NULL COMMENT '实例名称，如 service-center-grpc',
    `environment`             VARCHAR(32)     NOT NULL COMMENT '部署环境（DEVELOPMENT开发,STAGING预发布,PRODUCTION生产）',
    
    -- 服务器类型和监听配置
    `serverType`              VARCHAR(32)     NOT NULL DEFAULT 'GRPC' COMMENT '服务器类型（GRPC, HTTP）',
    `listenAddress`           VARCHAR(128)    NOT NULL DEFAULT '0.0.0.0' COMMENT '监听地址',
    `listenPort`              INT             NOT NULL DEFAULT 12004 COMMENT '监听端口',
    
    -- gRPC 消息大小配置
    `maxRecvMsgSize`          INT             NOT NULL DEFAULT 16777216 COMMENT '最大接收消息大小（字节，默认16MB）',
    `maxSendMsgSize`          INT             NOT NULL DEFAULT 16777216 COMMENT '最大发送消息大小（字节，默认16MB）',
    
    -- gRPC Keep-Alive 配置
    `keepAliveTime`           INT             NOT NULL DEFAULT 30 COMMENT 'Keep-alive 发送间隔（秒）',
    `keepAliveTimeout`        INT             NOT NULL DEFAULT 10 COMMENT 'Keep-alive 超时时间（秒）',
    `keepAliveMinTime`        INT             NOT NULL DEFAULT 15 COMMENT '客户端最小 Keep-alive 间隔（秒）',
    `permitWithoutStream`     VARCHAR(1)      NOT NULL DEFAULT 'Y' COMMENT '是否允许无活跃流时发送 Keep-alive（N否,Y是）',
    
    -- gRPC 连接管理配置
    `maxConnectionIdle`       INT             NOT NULL DEFAULT 0 COMMENT '最大连接空闲时间（秒，0表示无限制）',
    `maxConnectionAge`        INT             NOT NULL DEFAULT 0 COMMENT '最大连接存活时间（秒，0表示无限制）',
    `maxConnectionAgeGrace`   INT             NOT NULL DEFAULT 20 COMMENT '连接关闭宽限期（秒）',
    
    -- gRPC 功能开关
    `enableReflection`        VARCHAR(1)      NOT NULL DEFAULT 'Y' COMMENT '是否启用 gRPC 反射（N否,Y是，用于 grpcurl）',
    `enableTLS`               VARCHAR(1)      NOT NULL DEFAULT 'N' COMMENT '是否启用 TLS 加密（N否,Y是）',
    
    -- 证书配置 - 支持文件路径和数据库存储（参考 HUB_GW_INSTANCE）
    `certStorageType`         VARCHAR(20)     NOT NULL DEFAULT 'FILE' COMMENT '证书存储类型（FILE文件,DATABASE数据库）',
    `certFilePath`            VARCHAR(255)    DEFAULT NULL COMMENT 'TLS 证书文件路径',
    `keyFilePath`             VARCHAR(255)    DEFAULT NULL COMMENT 'TLS 私钥文件路径',
    `certContent`             TEXT            DEFAULT NULL COMMENT 'TLS 证书内容（PEM格式）',
    `keyContent`              TEXT            DEFAULT NULL COMMENT 'TLS 私钥内容（PEM格式）',
    `certChainContent`        TEXT            DEFAULT NULL COMMENT 'TLS 证书链内容（PEM格式）',
    `certPassword`            VARCHAR(255)    DEFAULT NULL COMMENT '证书密码（加密存储）',
    `enableMTLS`              VARCHAR(1)      NOT NULL DEFAULT 'N' COMMENT '是否启用双向 TLS 认证（N否,Y是）',
    
    -- 性能调优配置
    `maxConcurrentStreams`    INT             NOT NULL DEFAULT 250 COMMENT '最大并发流数量（0表示无限制）',
    `readBufferSize`          INT             NOT NULL DEFAULT 32768 COMMENT '读缓冲区大小（字节，默认32KB）',
    `writeBufferSize`         INT             NOT NULL DEFAULT 32768 COMMENT '写缓冲区大小（字节，默认32KB）',
    
    -- 健康检查配置
    `healthCheckInterval`     INT             NOT NULL DEFAULT 30 COMMENT '健康检查间隔（秒），0表示禁用健康检查',
    `healthCheckTimeout`      INT             NOT NULL DEFAULT 5 COMMENT '健康检查超时时间（秒）',
    
    -- 实例状态管理
    `instanceStatus`          VARCHAR(20)     NOT NULL DEFAULT 'STOPPED' COMMENT '实例状态（STOPPED停止,STARTING启动中,RUNNING运行中,STOPPING停止中,ERROR异常）',
    `statusMessage`           TEXT            DEFAULT NULL COMMENT '状态消息，记录启动、停止、异常等详细信息',
    `lastStatusTime`          DATETIME        DEFAULT NULL COMMENT '最后状态变更时间（启动/停止/异常）',
    `lastHealthCheckTime`     DATETIME        DEFAULT NULL COMMENT '最后健康检查时间',
    
    -- 通用字段
    `addTime`                 DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `addWho`                  VARCHAR(32)     NOT NULL COMMENT '创建人ID',
    `editTime`                DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
    `editWho`                 VARCHAR(32)     NOT NULL COMMENT '最后修改人ID',
    `oprSeqFlag`              VARCHAR(32)     NOT NULL COMMENT '操作序列标识',
    `currentVersion`          INT             NOT NULL DEFAULT 1 COMMENT '当前版本号',
    `activeFlag`              VARCHAR(1)      NOT NULL DEFAULT 'Y' COMMENT '活动状态标记（N非活动,Y活动）',
    `noteText`                VARCHAR(500)    DEFAULT NULL COMMENT '备注信息',
    `extProperty`             TEXT            DEFAULT NULL COMMENT '扩展属性，JSON格式',
    
    -- 访问控制配置
    `enableAuth`              VARCHAR(1)      NOT NULL DEFAULT 'N' COMMENT '是否启用认证（N否,Y是）',
    `ipWhitelist`             TEXT            DEFAULT NULL COMMENT 'IP 白名单（JSON 数组格式，如 ["192.168.1.0/24", "10.0.0.1"]）',
    `ipBlacklist`             TEXT            DEFAULT NULL COMMENT 'IP 黑名单（JSON 数组格式）',
    
    -- 主键和索引
    PRIMARY KEY (`tenantId`, `instanceName`, `environment`),
    KEY `IDX_SC_INST_TENANT` (`tenantId`),
    KEY `IDX_SC_INST_INSTANCE` (`instanceName`),
    KEY `IDX_SC_INST_ENV` (`environment`),
    KEY `IDX_SC_INST_ACTIVE` (`activeFlag`),
    KEY `IDX_SC_INST_TYPE` (`serverType`),
    KEY `IDX_SC_INST_PORT` (`listenPort`),
    KEY `IDX_SC_INST_STATUS` (`instanceStatus`),
    KEY `IDX_SC_INST_HEALTH_CHECK` (`lastHealthCheckTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务中心监听配置表（gRPC 服务器配置）';

-- =============================================
-- 初始化默认配置（开发环境）
-- =============================================
INSERT INTO `HUB_SERVICE_INSTANCE` (
    `tenantId`,
    `instanceName`,
    `environment`,
    `serverType`,
    `listenAddress`,
    `listenPort`,
    `maxRecvMsgSize`,
    `maxSendMsgSize`,
    `keepAliveTime`,
    `keepAliveTimeout`,
    `keepAliveMinTime`,
    `permitWithoutStream`,
    `maxConnectionIdle`,
    `maxConnectionAge`,
    `maxConnectionAgeGrace`,
    `enableReflection`,
    `enableTLS`,
    `certStorageType`,
    `enableMTLS`,
    `maxConcurrentStreams`,
    `readBufferSize`,
    `writeBufferSize`,
    `healthCheckInterval`,
    `healthCheckTimeout`,
    `instanceStatus`,
    `statusMessage`,
    `enableAuth`,
    `ipWhitelist`,
    `ipBlacklist`,
    `addWho`,
    `editWho`,
    `oprSeqFlag`,
    `currentVersion`,
    `activeFlag`,
    `noteText`
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

