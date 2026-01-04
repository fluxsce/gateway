
-- 23. IP访问控制配置表
CREATE TABLE IF NOT EXISTS HUB_GW_IP_ACCESS_CONFIG (
    tenantId TEXT NOT NULL,
    ipAccessConfigId TEXT NOT NULL,
    securityConfigId TEXT NOT NULL,
    configName TEXT NOT NULL,
    defaultPolicy TEXT NOT NULL DEFAULT 'allow',
    whitelistIps TEXT,
    blacklistIps TEXT,
    whitelistCidrs TEXT,
    blacklistCidrs TEXT,
    trustXForwardedFor TEXT NOT NULL DEFAULT 'Y',
    trustXRealIp TEXT NOT NULL DEFAULT 'Y',
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
    PRIMARY KEY (tenantId, ipAccessConfigId)
);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_IP_ACCESS_CONFIG_security ON HUB_GW_IP_ACCESS_CONFIG(securityConfigId);