
-- 27. 服务器信息主表
CREATE TABLE IF NOT EXISTS HUB_METRIC_SERVER_INFO (
    metricServerId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    hostname TEXT NOT NULL,
    osType TEXT NOT NULL,
    osVersion TEXT NOT NULL,
    kernelVersion TEXT,
    architecture TEXT NOT NULL,
    bootTime DATETIME NOT NULL,
    ipAddress TEXT,
    macAddress TEXT,
    serverLocation TEXT,
    serverType TEXT,
    lastUpdateTime DATETIME NOT NULL,
    networkInfo TEXT,
    systemInfo TEXT,
    hardwareInfo TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricServerId)
);
CREATE INDEX IDX_METRIC_SERVER_HOST ON HUB_METRIC_SERVER_INFO(hostname);
CREATE INDEX IDX_METRIC_SERVER_OS ON HUB_METRIC_SERVER_INFO(osType);
CREATE INDEX IDX_METRIC_SERVER_IP ON HUB_METRIC_SERVER_INFO(ipAddress);
CREATE INDEX IDX_METRIC_SERVER_TYPE ON HUB_METRIC_SERVER_INFO(serverType);
CREATE INDEX IDX_METRIC_SERVER_ACTIVE ON HUB_METRIC_SERVER_INFO(activeFlag);
CREATE INDEX IDX_METRIC_SERVER_UPDATE ON HUB_METRIC_SERVER_INFO(lastUpdateTime);