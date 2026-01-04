CREATE TABLE HUB_METRIC_DISK_PART_LOG (
                                          metricDiskPartitionLogId VARCHAR2(32) NOT NULL,
                                          tenantId VARCHAR2(32) NOT NULL,
                                          metricServerId VARCHAR2(32) NOT NULL,
                                          deviceName VARCHAR2(100) NOT NULL,
                                          mountPoint VARCHAR2(200) NOT NULL,
                                          fileSystem VARCHAR2(50) NOT NULL,
                                          totalSpace NUMBER(19,0) DEFAULT 0 NOT NULL,
                                          usedSpace NUMBER(19,0) DEFAULT 0 NOT NULL,
                                          freeSpace NUMBER(19,0) DEFAULT 0 NOT NULL,
                                          usagePercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
                                          inodesTotal NUMBER(19,0) DEFAULT 0 NOT NULL,
                                          inodesUsed NUMBER(19,0) DEFAULT 0 NOT NULL,
                                          inodesFree NUMBER(19,0) DEFAULT 0 NOT NULL,
                                          inodesUsagePercent NUMBER(5,2) DEFAULT 0.00 NOT NULL,
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
                                          CONSTRAINT PK_DISK_PART_LOG PRIMARY KEY (tenantId, metricDiskPartitionLogId)
);
CREATE INDEX IDX_DSKPART_SERVER ON HUB_METRIC_DISK_PART_LOG(metricServerId);
CREATE INDEX IDX_DSKPART_TIME ON HUB_METRIC_DISK_PART_LOG(collectTime);
CREATE INDEX IDX_DSKPART_DEVICE ON HUB_METRIC_DISK_PART_LOG(deviceName);
CREATE INDEX IDX_DSKPART_USAGE ON HUB_METRIC_DISK_PART_LOG(usagePercent);
CREATE INDEX IDX_DSKPART_ACTIVE ON HUB_METRIC_DISK_PART_LOG(activeFlag);
CREATE INDEX IDX_DSKPART_SRV_TIME ON HUB_METRIC_DISK_PART_LOG(metricServerId, collectTime);
CREATE INDEX IDX_DSKPART_SRV_DEV ON HUB_METRIC_DISK_PART_LOG(metricServerId, deviceName);
CREATE INDEX IDX_DSKPART_TNT_TIME ON HUB_METRIC_DISK_PART_LOG(tenantId, collectTime);
COMMENT ON TABLE HUB_METRIC_DISK_PART_LOG IS '磁盘分区采集日志表';
