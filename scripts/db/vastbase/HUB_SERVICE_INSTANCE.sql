-- Vastbase（PostgreSQL 兼容模式）。由 scripts/db/mysql 转换。
CREATE TABLE IF NOT EXISTS HUB_SERVICE_INSTANCE (
  tenantId                VARCHAR(32)     NOT NULL,
  instanceName            VARCHAR(100)    NOT NULL,
  environment             VARCHAR(32)     NOT NULL,
  serverType              VARCHAR(32)     NOT NULL DEFAULT 'GRPC',
  listenAddress           VARCHAR(128)    NOT NULL DEFAULT '0.0.0.0',
  listenPort              INT             NOT NULL DEFAULT 12004,
  maxRecvMsgSize          INT             NOT NULL DEFAULT 16777216,
  maxSendMsgSize          INT             NOT NULL DEFAULT 16777216,
  keepAliveTime           INT             NOT NULL DEFAULT 30,
  keepAliveTimeout        INT             NOT NULL DEFAULT 10,
  keepAliveMinTime        INT             NOT NULL DEFAULT 15,
  permitWithoutStream     VARCHAR(1)      NOT NULL DEFAULT 'Y',
  maxConnectionIdle       INT             NOT NULL DEFAULT 0,
  maxConnectionAge        INT             NOT NULL DEFAULT 0,
  maxConnectionAgeGrace   INT             NOT NULL DEFAULT 20,
  enableReflection        VARCHAR(1)      NOT NULL DEFAULT 'Y',
  enableTLS               VARCHAR(1)      NOT NULL DEFAULT 'N',
  certStorageType         VARCHAR(20)     NOT NULL DEFAULT 'FILE',
  certFilePath            VARCHAR(255)    DEFAULT NULL,
  keyFilePath             VARCHAR(255)    DEFAULT NULL,
  certContent             TEXT            DEFAULT NULL,
  keyContent              TEXT            DEFAULT NULL,
  certChainContent        TEXT            DEFAULT NULL,
  certPassword            VARCHAR(255)    DEFAULT NULL,
  enableMTLS              VARCHAR(1)      NOT NULL DEFAULT 'N',
  maxConcurrentStreams    INT             NOT NULL DEFAULT 250,
  readBufferSize          INT             NOT NULL DEFAULT 32768,
  writeBufferSize         INT             NOT NULL DEFAULT 32768,
  healthCheckInterval     INT             NOT NULL DEFAULT 30,
  healthCheckTimeout      INT             NOT NULL DEFAULT 5,
  instanceStatus          VARCHAR(20)     NOT NULL DEFAULT 'STOPPED',
  statusMessage           TEXT            DEFAULT NULL,
  lastStatusTime          TIMESTAMP        DEFAULT NULL,
  lastHealthCheckTime     TIMESTAMP        DEFAULT NULL,
  addTime                 TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho                  VARCHAR(32)     NOT NULL,
  editTime                TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho                 VARCHAR(32)     NOT NULL,
  oprSeqFlag              VARCHAR(32)     NOT NULL,
  currentVersion          INT             NOT NULL DEFAULT 1,
  activeFlag              VARCHAR(1)      NOT NULL DEFAULT 'Y',
  noteText                VARCHAR(500)    DEFAULT NULL,
  extProperty             TEXT            DEFAULT NULL,
  enableAuth              VARCHAR(1)      NOT NULL DEFAULT 'N',
  ipWhitelist             TEXT            DEFAULT NULL,
  ipBlacklist             TEXT            DEFAULT NULL,
  PRIMARY KEY (tenantId, instanceName, environment)
);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_TENANT ON HUB_SERVICE_INSTANCE (tenantId);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_INSTANCE ON HUB_SERVICE_INSTANCE (instanceName);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_ENV ON HUB_SERVICE_INSTANCE (environment);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_ACTIVE ON HUB_SERVICE_INSTANCE (activeFlag);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_TYPE ON HUB_SERVICE_INSTANCE (serverType);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_PORT ON HUB_SERVICE_INSTANCE (listenPort);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_STATUS ON HUB_SERVICE_INSTANCE (instanceStatus);
CREATE INDEX IF NOT EXISTS IDX_SC_INST_HEALTH_CHECK ON HUB_SERVICE_INSTANCE (lastHealthCheckTime);

INSERT INTO HUB_SERVICE_INSTANCE ( tenantId, instanceName, environment, serverType, listenAddress, listenPort, maxRecvMsgSize, maxSendMsgSize, keepAliveTime, keepAliveTimeout, keepAliveMinTime, permitWithoutStream, maxConnectionIdle, maxConnectionAge, maxConnectionAgeGrace, enableReflection, enableTLS, certStorageType, enableMTLS, maxConcurrentStreams, readBufferSize, writeBufferSize, healthCheckInterval, healthCheckTimeout, instanceStatus, statusMessage, enableAuth, ipWhitelist, ipBlacklist, addWho, editWho, oprSeqFlag, currentVersion, activeFlag, noteText ) VALUES ( 'default', 'service-center-grpc', 'DEVELOPMENT', 'GRPC', '0.0.0.0', 12004, 16777216, 16777216, 30, 10, 15, 'Y', 0, 0, 20, 'Y', 'N', 'FILE', 'N', 250, 32768, 32768, 30, 5, 'STOPPED', '实例已创建，等待启动', 'N', NULL, NULL, 'system', 'system', 'INIT', 1, 'Y', '开发环境默认配置（明文传输，启用反射，无认证，开放访问用于本地调试）' );
