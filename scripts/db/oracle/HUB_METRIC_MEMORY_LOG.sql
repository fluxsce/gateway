CREATE TABLE HUB_METRIC_MEMORY_LOG (
                                       metricMemoryLogId VARCHAR2(32) NOT NULL,
                                       tenantId VARCHAR2(32) NOT NULL,
                                       metricServerId VARCHAR2(32) NOT NULL,
                                       totalMemory NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       availableMemory NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       usedMemory NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       usagePercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                       freeMemory NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       cachedMemory NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       buffersMemory NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       sharedMemory NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       swapTotal NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       swapUsed NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       swapFree NUMBER(19,0) DEFAULT 0 NOT NULL,
                                       swapUsagePercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
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
                                       CONSTRAINT PK_MEM_LOG PRIMARY KEY (tenantId, metricMemoryLogId)
);
CREATE INDEX IDX_MEMLOG_SERVER ON HUB_METRIC_MEMORY_LOG(metricServerId);
CREATE INDEX IDX_MEMLOG_TIME ON HUB_METRIC_MEMORY_LOG(collectTime);
CREATE INDEX IDX_MEMLOG_USAGE ON HUB_METRIC_MEMORY_LOG(usagePercent);
CREATE INDEX IDX_MEMLOG_ACTIVE ON HUB_METRIC_MEMORY_LOG(activeFlag);
CREATE INDEX IDX_MEMLOG_SRV_TIME ON HUB_METRIC_MEMORY_LOG(metricServerId, collectTime);
CREATE INDEX IDX_MEMLOG_TNT_TIME ON HUB_METRIC_MEMORY_LOG(tenantId, collectTime);
COMMENT ON TABLE HUB_METRIC_MEMORY_LOG IS '内存采集日志表';
