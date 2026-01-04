
-- 10. 熔断配置表
CREATE TABLE IF NOT EXISTS HUB_GW_CIRCUIT_BREAKER_CONFIG (
    tenantId TEXT NOT NULL,
    circuitBreakerConfigId TEXT NOT NULL,
    routeConfigId TEXT,
    targetServiceId TEXT,
    breakerName TEXT NOT NULL,
    keyStrategy TEXT NOT NULL DEFAULT 'api',
    errorRatePercent INTEGER NOT NULL DEFAULT 50,
    minimumRequests INTEGER NOT NULL DEFAULT 10,
    halfOpenMaxRequests INTEGER NOT NULL DEFAULT 3,
    slowCallThreshold INTEGER NOT NULL DEFAULT 1000,
    slowCallRatePercent INTEGER NOT NULL DEFAULT 50,
    openTimeoutSeconds INTEGER NOT NULL DEFAULT 60,
    windowSizeSeconds INTEGER NOT NULL DEFAULT 60,
    errorStatusCode INTEGER NOT NULL DEFAULT 503,
    errorMessage TEXT DEFAULT 'Service temporarily unavailable due to circuit breaker',
    storageType TEXT NOT NULL DEFAULT 'memory',
    storageConfig TEXT,
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
    PRIMARY KEY (tenantId, circuitBreakerConfigId)
);
CREATE INDEX IDX_GW_CB_ROUTE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_CB_SERVICE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(targetServiceId);
CREATE INDEX IDX_GW_CB_STRATEGY ON HUB_GW_CIRCUIT_BREAKER_CONFIG(keyStrategy);
CREATE INDEX IDX_GW_CB_STORAGE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(storageType);
CREATE INDEX IDX_GW_CB_PRIORITY ON HUB_GW_CIRCUIT_BREAKER_CONFIG(configPriority);