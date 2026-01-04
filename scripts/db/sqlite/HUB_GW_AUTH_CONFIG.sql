
-- 11. 认证配置表
CREATE TABLE IF NOT EXISTS HUB_GW_AUTH_CONFIG (
    tenantId TEXT NOT NULL,
    authConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    authName TEXT NOT NULL,
    authType TEXT NOT NULL,
    authStrategy TEXT DEFAULT 'REQUIRED',
    authConfig TEXT NOT NULL,
    exemptPaths TEXT,
    exemptHeaders TEXT,
    failureStatusCode INTEGER NOT NULL DEFAULT 401,
    failureMessage TEXT DEFAULT '认证失败',
    configPriority INTEGER NOT NULL DEFAULT 0,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, authConfigId)
);
CREATE INDEX IDX_GW_AUTH_INST ON HUB_GW_AUTH_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_AUTH_ROUTE ON HUB_GW_AUTH_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_AUTH_TYPE ON HUB_GW_AUTH_CONFIG(authType);
CREATE INDEX IDX_GW_AUTH_PRIORITY ON HUB_GW_AUTH_CONFIG(configPriority);