
-- 12. 服务定义表
CREATE TABLE IF NOT EXISTS HUB_GW_SERVICE_DEFINITION (
    tenantId TEXT NOT NULL,
    serviceDefinitionId TEXT NOT NULL,
    serviceName TEXT NOT NULL,
    serviceDesc TEXT,
    serviceType INTEGER NOT NULL DEFAULT 0,
    proxyConfigId TEXT NOT NULL,
    loadBalanceStrategy TEXT NOT NULL DEFAULT 'round-robin',
    discoveryType TEXT,
    discoveryConfig TEXT,
    sessionAffinity TEXT NOT NULL DEFAULT 'N',
    stickySession TEXT NOT NULL DEFAULT 'N',
    maxRetries INTEGER NOT NULL DEFAULT 3,
    retryTimeoutMs INTEGER NOT NULL DEFAULT 5000,
    enableCircuitBreaker TEXT NOT NULL DEFAULT 'N',
    healthCheckEnabled TEXT NOT NULL DEFAULT 'Y',
    healthCheckPath TEXT DEFAULT '/health',
    healthCheckMethod TEXT DEFAULT 'GET',
    healthCheckIntervalSeconds INTEGER DEFAULT 30,
    healthCheckTimeoutMs INTEGER DEFAULT 5000,
    healthyThreshold INTEGER DEFAULT 2,
    unhealthyThreshold INTEGER DEFAULT 3,
    expectedStatusCodes TEXT DEFAULT '200',
    healthCheckHeaders TEXT,
    loadBalancerConfig TEXT,
    serviceMetadata TEXT,
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
    PRIMARY KEY (tenantId, serviceDefinitionId)
);
CREATE INDEX IDX_GW_SVC_NAME ON HUB_GW_SERVICE_DEFINITION(serviceName);
CREATE INDEX IDX_GW_SVC_TYPE ON HUB_GW_SERVICE_DEFINITION(serviceType);
CREATE INDEX IDX_GW_SVC_STRATEGY ON HUB_GW_SERVICE_DEFINITION(loadBalanceStrategy);
CREATE INDEX IDX_GW_SVC_HEALTH ON HUB_GW_SERVICE_DEFINITION(healthCheckEnabled);
CREATE INDEX IDX_GW_SVC_PROXY ON HUB_GW_SERVICE_DEFINITION(proxyConfigId);