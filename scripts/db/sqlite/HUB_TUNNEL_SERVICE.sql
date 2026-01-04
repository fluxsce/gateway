-- =========================================

CREATE TABLE HUB_TUNNEL_SERVICE (
  tunnelServiceId TEXT NOT NULL PRIMARY KEY,
  tenantId TEXT NOT NULL,
  tunnelClientId TEXT NOT NULL,
  userId TEXT NOT NULL,
  serviceName TEXT NOT NULL UNIQUE,
  serviceDescription TEXT,
  serviceType TEXT NOT NULL,
  localAddress TEXT NOT NULL DEFAULT '127.0.0.1',
  localPort INTEGER NOT NULL,
  remotePort INTEGER,
  customDomains TEXT,
  subDomain TEXT,
  httpUser TEXT,
  httpPassword TEXT,
  hostHeaderRewrite TEXT,
  headers TEXT,
  locations TEXT,
  useEncryption TEXT NOT NULL DEFAULT 'N' CHECK(useEncryption IN ('Y', 'N')),
  useCompression TEXT NOT NULL DEFAULT 'Y' CHECK(useCompression IN ('Y', 'N')),
  secretKey TEXT,
  bandwidthLimit TEXT,
  maxConnections INTEGER DEFAULT 100,
  healthCheckType TEXT,
  healthCheckUrl TEXT,
  serviceStatus TEXT NOT NULL DEFAULT 'active',
  registeredTime TEXT NOT NULL DEFAULT (datetime('now')),
  lastActiveTime TEXT,
  connectionCount INTEGER DEFAULT 0,
  totalConnections INTEGER DEFAULT 0,
  totalTraffic INTEGER DEFAULT 0,
  serviceConfig TEXT,
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y' CHECK(activeFlag IN ('Y', 'N')),
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT
);

CREATE INDEX IDX_TUNNEL_SVC_TENANT ON HUB_TUNNEL_SERVICE(tenantId);
CREATE INDEX IDX_TUNNEL_SVC_CLIENT ON HUB_TUNNEL_SERVICE(tunnelClientId);
CREATE INDEX IDX_TUNNEL_SVC_USER ON HUB_TUNNEL_SERVICE(userId);
CREATE INDEX IDX_TUNNEL_SVC_TYPE ON HUB_TUNNEL_SERVICE(serviceType);
CREATE INDEX IDX_TUNNEL_SVC_STATUS ON HUB_TUNNEL_SERVICE(serviceStatus);
CREATE INDEX IDX_TUNNEL_SVC_PORT ON HUB_TUNNEL_SERVICE(remotePort);
CREATE INDEX IDX_TUNNEL_SVC_DOMAIN ON HUB_TUNNEL_SERVICE(subDomain);

-- =====================================================
-- 权限系统数据库表结构设计
-- 遵循 docs/database/naming-convention.md 规范
-- 基于 web/actions/permission-design.md 设计文档
-- 创建时间: 2024-12-19
-- =====================================================