CREATE TABLE HUB_METRIC_TEMP_LOG (
                                     metricTemperatureLogId VARCHAR2(32) NOT NULL,
                                     tenantId VARCHAR2(32) NOT NULL,
                                     metricServerId VARCHAR2(32) NOT NULL,
                                     sensorName VARCHAR2(100) NOT NULL,
                                     temperatureValue NUMBER(6,2) DEFAULT 0.00 NOT NULL,
                                     highThreshold NUMBER(6,2),
                                     criticalThreshold NUMBER(6,2),
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
                                     CONSTRAINT PK_TEMP_LOG PRIMARY KEY (tenantId, metricTemperatureLogId)
);
CREATE INDEX IDX_TEMPLOG_SERVER ON HUB_METRIC_TEMP_LOG(metricServerId);
CREATE INDEX IDX_TEMPLOG_TIME ON HUB_METRIC_TEMP_LOG(collectTime);
CREATE INDEX IDX_TEMPLOG_SENSOR ON HUB_METRIC_TEMP_LOG(sensorName);
CREATE INDEX IDX_TEMPLOG_ACTIVE ON HUB_METRIC_TEMP_LOG(activeFlag);
CREATE INDEX IDX_TEMPLOG_SRV_TIME ON HUB_METRIC_TEMP_LOG(metricServerId, collectTime);
CREATE INDEX IDX_TEMPLOG_SRV_SEN ON HUB_METRIC_TEMP_LOG(metricServerId, sensorName);
CREATE INDEX IDX_TEMPLOG_TNT_TIME ON HUB_METRIC_TEMP_LOG(tenantId, collectTime);
COMMENT ON TABLE HUB_METRIC_TEMP_LOG IS '温度信息采集日志表'; 

