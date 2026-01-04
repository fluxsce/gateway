CREATE TABLE HUB_METRIC_PROCSTAT_LOG (
                                         metricProcessStatsLogId VARCHAR2(32) NOT NULL,
                                         tenantId VARCHAR2(32) NOT NULL,
                                         metricServerId VARCHAR2(32) NOT NULL,
                                         runningCount NUMBER(10,0) DEFAULT 0 NOT NULL,
                                         sleepingCount NUMBER(10,0) DEFAULT 0 NOT NULL,
                                         stoppedCount NUMBER(10,0) DEFAULT 0 NOT NULL,
                                         zombieCount NUMBER(10,0) DEFAULT 0 NOT NULL,
                                         totalCount NUMBER(10,0) DEFAULT 0 NOT NULL,
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
                                         CONSTRAINT PK_PROCSTAT_LOG PRIMARY KEY (tenantId, metricProcessStatsLogId)
);
CREATE INDEX IDX_PROCSTAT_SERVER ON HUB_METRIC_PROCSTAT_LOG(metricServerId);
CREATE INDEX IDX_PROCSTAT_TIME ON HUB_METRIC_PROCSTAT_LOG(collectTime);
CREATE INDEX IDX_PROCSTAT_ACTIVE ON HUB_METRIC_PROCSTAT_LOG(activeFlag);
CREATE INDEX IDX_PROCSTAT_SRV_TIME ON HUB_METRIC_PROCSTAT_LOG(metricServerId, collectTime);
CREATE INDEX IDX_PROCSTAT_TNT_TIME ON HUB_METRIC_PROCSTAT_LOG(tenantId, collectTime);
COMMENT ON TABLE HUB_METRIC_PROCSTAT_LOG IS '进程统计采集日志表';
