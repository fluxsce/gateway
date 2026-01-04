
-- 8. 跨域配置表
CREATE TABLE IF NOT EXISTS HUB_GW_CORS_CONFIG (
    tenantId TEXT NOT NULL,
    corsConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    configName TEXT NOT NULL,
    allowOrigins TEXT NOT NULL,
    allowMethods TEXT NOT NULL DEFAULT 'GET,POST,PUT,DELETE,OPTIONS',
    allowHeaders TEXT,
    exposeHeaders TEXT,
    allowCredentials TEXT NOT NULL DEFAULT 'N',
    maxAgeSeconds INTEGER NOT NULL DEFAULT 86400,
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
    PRIMARY KEY (tenantId, corsConfigId)
);
CREATE INDEX IDX_GW_CORS_INST ON HUB_GW_CORS_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_CORS_ROUTE ON HUB_GW_CORS_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_CORS_PRIORITY ON HUB_GW_CORS_CONFIG(configPriority);