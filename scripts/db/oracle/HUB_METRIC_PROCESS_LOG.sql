CREATE TABLE HUB_METRIC_PROCESS_LOG (
                                        metricProcessLogId VARCHAR2(32) NOT NULL,
                                        tenantId VARCHAR2(32) NOT NULL,
                                        metricServerId VARCHAR2(32) NOT NULL,
                                        processId NUMBER(10,0) NOT NULL,
                                        parentProcessId NUMBER(10,0),
                                        processName VARCHAR2(200) NOT NULL,
                                        processStatus VARCHAR2(50) NOT NULL,
                                        createTime DATE NOT NULL,
                                        runTime NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        memoryUsage NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        memoryPercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                        cpuPercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                        threadCount NUMBER(10,0) DEFAULT 0 NOT NULL,
                                        fileDescriptorCount NUMBER(10,0) DEFAULT 0 NOT NULL,
                                        commandLine CLOB,
                                        executablePath VARCHAR2(500),
                                        workingDirectory VARCHAR2(500),
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
                                        CONSTRAINT PK_PROC_LOG PRIMARY KEY (tenantId, metricProcessLogId)
);
CREATE INDEX IDX_PROCLOG_SERVER ON HUB_METRIC_PROCESS_LOG(metricServerId);
CREATE INDEX IDX_PROCLOG_TIME ON HUB_METRIC_PROCESS_LOG(collectTime);
CREATE INDEX IDX_PROCLOG_PID ON HUB_METRIC_PROCESS_LOG(processId);
CREATE INDEX IDX_PROCLOG_NAME ON HUB_METRIC_PROCESS_LOG(processName);
CREATE INDEX IDX_PROCLOG_STATUS ON HUB_METRIC_PROCESS_LOG(processStatus);
CREATE INDEX IDX_PROCLOG_ACTIVE ON HUB_METRIC_PROCESS_LOG(activeFlag);
CREATE INDEX IDX_PROCLOG_SRV_TIME ON HUB_METRIC_PROCESS_LOG(metricServerId, collectTime);
CREATE INDEX IDX_PROCLOG_SRV_PID ON HUB_METRIC_PROCESS_LOG(metricServerId, processId);
CREATE INDEX IDX_PROCLOG_TNT_TIME ON HUB_METRIC_PROCESS_LOG(tenantId, collectTime);
COMMENT ON TABLE HUB_METRIC_PROCESS_LOG IS '进程信息采集日志表';
