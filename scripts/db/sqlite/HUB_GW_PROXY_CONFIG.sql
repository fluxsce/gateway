
-- 14. 代理配置表
CREATE TABLE IF NOT EXISTS HUB_GW_PROXY_CONFIG (
    tenantId TEXT NOT NULL,
    proxyConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT NOT NULL,
    proxyName TEXT NOT NULL,
    proxyType TEXT NOT NULL DEFAULT 'http',
    proxyId TEXT,
    configPriority INTEGER NOT NULL DEFAULT 0,
    proxyConfig TEXT NOT NULL,
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
    PRIMARY KEY (tenantId, proxyConfigId)
);
CREATE INDEX IDX_GW_PROXY_INST ON HUB_GW_PROXY_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_PROXY_TYPE ON HUB_GW_PROXY_CONFIG(proxyType);
CREATE INDEX IDX_GW_PROXY_PRIORITY ON HUB_GW_PROXY_CONFIG(configPriority);
CREATE INDEX IDX_GW_PROXY_ACTIVE ON HUB_GW_PROXY_CONFIG(activeFlag);