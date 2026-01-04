
-- 9. 限流配置表
CREATE TABLE IF NOT EXISTS HUB_GW_RATE_LIMIT_CONFIG (
    tenantId TEXT NOT NULL,
    rateLimitConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    limitName TEXT NOT NULL,
    algorithm TEXT NOT NULL DEFAULT 'token-bucket',
    keyStrategy TEXT NOT NULL DEFAULT 'ip',
    limitRate INTEGER NOT NULL,
    burstCapacity INTEGER NOT NULL DEFAULT 0,
    timeWindowSeconds INTEGER NOT NULL DEFAULT 1,
    rejectionStatusCode INTEGER NOT NULL DEFAULT 429,
    rejectionMessage TEXT DEFAULT '请求过于频繁，请稍后再试',
    configPriority INTEGER NOT NULL DEFAULT 0,
    customConfig TEXT,
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
    PRIMARY KEY (tenantId, rateLimitConfigId)
);
CREATE INDEX IDX_GW_RATE_INST ON HUB_GW_RATE_LIMIT_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_RATE_ROUTE ON HUB_GW_RATE_LIMIT_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_RATE_STRATEGY ON HUB_GW_RATE_LIMIT_CONFIG(keyStrategy);
CREATE INDEX IDX_GW_RATE_ALGORITHM ON HUB_GW_RATE_LIMIT_CONFIG(algorithm);
CREATE INDEX IDX_GW_RATE_PRIORITY ON HUB_GW_RATE_LIMIT_CONFIG(configPriority);
CREATE INDEX IDX_GW_RATE_ACTIVE ON HUB_GW_RATE_LIMIT_CONFIG(activeFlag);