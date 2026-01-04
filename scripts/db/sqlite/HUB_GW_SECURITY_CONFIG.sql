
-- 22. 安全配置表
CREATE TABLE IF NOT EXISTS HUB_GW_SECURITY_CONFIG (
    tenantId TEXT NOT NULL,
    securityConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    configName TEXT NOT NULL,
    configDesc TEXT,
    configPriority INTEGER NOT NULL DEFAULT 0,
    customConfigJson TEXT,
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
    PRIMARY KEY (tenantId, securityConfigId)
);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_SECURITY_CONFIG_instance ON HUB_GW_SECURITY_CONFIG(gatewayInstanceId);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_SECURITY_CONFIG_route ON HUB_GW_SECURITY_CONFIG(routeConfigId);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_SECURITY_CONFIG_priority ON HUB_GW_SECURITY_CONFIG(configPriority);