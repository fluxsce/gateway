CREATE TABLE HUB_METRIC_CPU_LOG (
                                    metricCpuLogId VARCHAR2(32) NOT NULL,
                                    tenantId VARCHAR2(32) NOT NULL,
                                    metricServerId VARCHAR2(32) NOT NULL,
                                    usagePercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                    userPercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                    systemPercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                    idlePercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                    ioWaitPercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                    irqPercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                    softIrqPercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                    coreCount NUMBER(10,0) DEFAULT 0 NOT NULL,
                                    logicalCount NUMBER(10,0) DEFAULT 0 NOT NULL,
                                    loadAvg1 NUMBER(8,2) DEFAULT 0.00 NOT NULL,
                                    loadAvg5 NUMBER(8,2) DEFAULT 0.00 NOT NULL,
                                    loadAvg15 NUMBER(8,2) DEFAULT 0.00 NOT NULL,
                                    collectTime DATE NOT NULL,
                                    addTime DATE DEFAULT SYSDATE NOT NULL,
                                    addWho VARCHAR2(32) NOT NULL,
                                    editTime DATE DEFAULT SYSDATE NOT NULL,
                                    editWho VARCHAR2(32) NOT NULL,
                                    oprSeqFlag VARCHAR2(32) NOT NULL,
                                    currentVersion NUMBER(10,0) DEFAULT 1 NOT NULL,
                                    activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL,
                                    noteText VARCHAR2(500),
                                    extProperty CLOB,
                                    reserved1 VARCHAR2(500),
                                    reserved2 VARCHAR2(500),
                                    reserved3 VARCHAR2(500),
                                    reserved4 VARCHAR2(500),
                                    reserved5 VARCHAR2(500),
                                    reserved6 VARCHAR2(500),
                                    reserved7 VARCHAR2(500),
                                    reserved8 VARCHAR2(500),
                                    reserved9 VARCHAR2(500),
                                    reserved10 VARCHAR2(500),
                                    CONSTRAINT PK_METRIC_CPU_LOG PRIMARY KEY (tenantId, metricCpuLogId)
);
CREATE INDEX IDX_CPULOG_SERVER ON HUB_METRIC_CPU_LOG(metricServerId);
CREATE INDEX IDX_CPULOG_TIME ON HUB_METRIC_CPU_LOG(collectTime);
CREATE INDEX IDX_CPULOG_USAGE ON HUB_METRIC_CPU_LOG(usagePercent);
CREATE INDEX IDX_CPULOG_ACTIVE ON HUB_METRIC_CPU_LOG(activeFlag);
CREATE INDEX IDX_CPULOG_SRV_TIME ON HUB_METRIC_CPU_LOG(metricServerId, collectTime);
CREATE INDEX IDX_CPULOG_TNT_TIME ON HUB_METRIC_CPU_LOG(tenantId, collectTime);
COMMENT ON TABLE HUB_METRIC_CPU_LOG IS 'CPU采集日志表';
