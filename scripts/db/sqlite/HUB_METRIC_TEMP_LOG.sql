
-- 35. 温度信息日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_TEMP_LOG (
    metricTemperatureLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    sensorName TEXT NOT NULL,
    temperatureValue REAL NOT NULL DEFAULT 0.00,
    highThreshold REAL,
    criticalThreshold REAL,
    collectTime DATETIME NOT NULL,
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
    PRIMARY KEY (tenantId, metricTemperatureLogId)
);

CREATE INDEX IDX_METRIC_TEMP_SERVER ON HUB_METRIC_TEMP_LOG(metricServerId);
CREATE INDEX IDX_METRIC_TEMP_TIME ON HUB_METRIC_TEMP_LOG(collectTime);
CREATE INDEX IDX_METRIC_TEMP_SENSOR ON HUB_METRIC_TEMP_LOG(sensorName);
CREATE INDEX IDX_METRIC_TEMP_ACTIVE ON HUB_METRIC_TEMP_LOG(activeFlag);
CREATE INDEX IDX_METRIC_TEMP_SRV_TIME ON HUB_METRIC_TEMP_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_TEMP_SRV_SENSOR ON HUB_METRIC_TEMP_LOG(metricServerId, sensorName);
CREATE INDEX IDX_METRIC_TEMP_TNT_TIME ON HUB_METRIC_TEMP_LOG(tenantId, collectTime);

-- =====================================================
-- 服务注册中心相关表结构
-- 基于 service_registry.sql 转换为 SQLite 格式
-- =====================================================