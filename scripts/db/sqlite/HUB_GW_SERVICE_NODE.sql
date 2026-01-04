
-- 13. 服务节点表
CREATE TABLE IF NOT EXISTS HUB_GW_SERVICE_NODE (
    tenantId TEXT NOT NULL,
    serviceNodeId TEXT NOT NULL,
    serviceDefinitionId TEXT NOT NULL,
    nodeId TEXT NOT NULL,
    nodeUrl TEXT NOT NULL,
    nodeHost TEXT NOT NULL,
    nodePort INTEGER NOT NULL,
    nodeProtocol TEXT NOT NULL DEFAULT 'HTTP',
    nodeWeight INTEGER NOT NULL DEFAULT 100,
    healthStatus TEXT NOT NULL DEFAULT 'Y',
    nodeMetadata TEXT,
    nodeStatus INTEGER NOT NULL DEFAULT 1,
    lastHealthCheckTime DATETIME,
    healthCheckResult TEXT,
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
    PRIMARY KEY (tenantId, serviceNodeId)
);
CREATE INDEX IDX_GW_NODE_SERVICE ON HUB_GW_SERVICE_NODE(serviceDefinitionId);
CREATE INDEX IDX_GW_NODE_ENDPOINT ON HUB_GW_SERVICE_NODE(nodeHost, nodePort);
CREATE INDEX IDX_GW_NODE_HEALTH ON HUB_GW_SERVICE_NODE(healthStatus);
CREATE INDEX IDX_GW_NODE_STATUS ON HUB_GW_SERVICE_NODE(nodeStatus);