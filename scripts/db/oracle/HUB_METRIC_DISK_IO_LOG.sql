CREATE TABLE HUB_METRIC_DISK_IO_LOG (
                                        metricDiskIoLogId VARCHAR2(32) NOT NULL,
                                        tenantId VARCHAR2(32) NOT NULL,
                                        metricServerId VARCHAR2(32) NOT NULL,
                                        deviceName VARCHAR2(100) NOT NULL,
                                        readCount NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        writeCount NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        readBytes NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        writeBytes NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        readTime NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        writeTime NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        ioInProgress NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        ioTime NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        readRate NUMBER(20,2) DEFAULT 0.00 NOT NULL,
                                        writeRate NUMBER(20,2) DEFAULT 0.00 NOT NULL,
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
                                        CONSTRAINT PK_DISK_IO_LOG PRIMARY KEY (tenantId, metricDiskIoLogId)
);
CREATE INDEX IDX_DSKIO_SERVER ON HUB_METRIC_DISK_IO_LOG(metricServerId);
CREATE INDEX IDX_DSKIO_TIME ON HUB_METRIC_DISK_IO_LOG(collectTime);
CREATE INDEX IDX_DSKIO_DEVICE ON HUB_METRIC_DISK_IO_LOG(deviceName);
CREATE INDEX IDX_DSKIO_ACTIVE ON HUB_METRIC_DISK_IO_LOG(activeFlag);
CREATE INDEX IDX_DSKIO_SRV_TIME ON HUB_METRIC_DISK_IO_LOG(metricServerId, collectTime);
CREATE INDEX IDX_DSKIO_SRV_DEV ON HUB_METRIC_DISK_IO_LOG(metricServerId, deviceName);
CREATE INDEX IDX_DSKIO_TNT_TIME ON HUB_METRIC_DISK_IO_LOG(tenantId, collectTime);
COMMENT ON TABLE HUB_METRIC_DISK_IO_LOG IS '磁盘IO采集日志表';
