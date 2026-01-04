
-- 5. 路由定义表
CREATE TABLE IF NOT EXISTS HUB_GW_ROUTE_CONFIG (
    tenantId TEXT NOT NULL,
    routeConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT NOT NULL,
    routeName TEXT NOT NULL,
    routePath TEXT NOT NULL,
    allowedMethods TEXT,
    allowedHosts TEXT,
    matchType INTEGER NOT NULL DEFAULT 1,
    routePriority INTEGER NOT NULL DEFAULT 100,
    stripPathPrefix TEXT NOT NULL DEFAULT 'N',
    rewritePath TEXT,
    enableWebsocket TEXT NOT NULL DEFAULT 'N',
    timeoutMs INTEGER NOT NULL DEFAULT 30000,
    retryCount INTEGER NOT NULL DEFAULT 0,
    retryIntervalMs INTEGER NOT NULL DEFAULT 1000,
    serviceDefinitionId TEXT,
    logConfigId TEXT,
    routeMetadata TEXT,
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
    PRIMARY KEY (tenantId, routeConfigId)
);
CREATE INDEX IDX_GW_ROUTE_INST ON HUB_GW_ROUTE_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_ROUTE_SERVICE ON HUB_GW_ROUTE_CONFIG(serviceDefinitionId);
CREATE INDEX IDX_GW_ROUTE_LOG ON HUB_GW_ROUTE_CONFIG(logConfigId);
CREATE INDEX IDX_GW_ROUTE_PRIORITY ON HUB_GW_ROUTE_CONFIG(routePriority);
CREATE INDEX IDX_GW_ROUTE_PATH ON HUB_GW_ROUTE_CONFIG(routePath);
CREATE INDEX IDX_GW_ROUTE_ACTIVE ON HUB_GW_ROUTE_CONFIG(activeFlag);