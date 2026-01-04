-- =========================================

CREATE TABLE HUB_TUNNEL_CLIENT (
  tunnelClientId TEXT NOT NULL PRIMARY KEY,
  tenantId TEXT NOT NULL,
  userId TEXT NOT NULL,
  clientName TEXT NOT NULL UNIQUE,
  clientDescription TEXT,
  clientVersion TEXT,
  operatingSystem TEXT,
  clientIpAddress TEXT,
  clientMacAddress TEXT,
  serverAddress TEXT NOT NULL,
  serverPort INTEGER NOT NULL DEFAULT 7000,
  authToken TEXT NOT NULL,
  tlsEnable TEXT NOT NULL DEFAULT 'N' CHECK(tlsEnable IN ('Y', 'N')),
  autoReconnect TEXT NOT NULL DEFAULT 'Y' CHECK(autoReconnect IN ('Y', 'N')),
  maxRetries INTEGER NOT NULL DEFAULT 5,
  retryInterval INTEGER NOT NULL DEFAULT 20,
  heartbeatInterval INTEGER NOT NULL DEFAULT 30,
  heartbeatTimeout INTEGER NOT NULL DEFAULT 90,
  connectionStatus TEXT NOT NULL DEFAULT 'disconnected',
  lastConnectTime TEXT,
  lastDisconnectTime TEXT,
  totalConnectTime INTEGER DEFAULT 0,
  reconnectCount INTEGER DEFAULT 0,
  serviceCount INTEGER DEFAULT 0,
  lastHeartbeat TEXT,
  clientConfig TEXT,
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
CREATE INDEX IDX_TUNNEL_CLIENT_TENANT ON HUB_TUNNEL_CLIENT(tenantId);
CREATE INDEX IDX_TUNNEL_CLIENT_USER ON HUB_TUNNEL_CLIENT(userId);
CREATE INDEX IDX_TUNNEL_CLIENT_STATUS ON HUB_TUNNEL_CLIENT(connectionStatus);
CREATE INDEX IDX_TUNNEL_CLIENT_IP ON HUB_TUNNEL_CLIENT(clientIpAddress);
CREATE INDEX IDX_TUNNEL_CLIENT_HB ON HUB_TUNNEL_CLIENT(lastHeartbeat);