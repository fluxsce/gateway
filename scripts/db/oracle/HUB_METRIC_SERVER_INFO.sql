CREATE TABLE HUB_METRIC_SERVER_INFO (
                                        metricServerId VARCHAR2(64) NOT NULL,
                                        tenantId VARCHAR2(64) NOT NULL,
                                        hostname VARCHAR2(255) NOT NULL,
                                        osType VARCHAR2(100) NOT NULL,
                                        osVersion VARCHAR2(255) NOT NULL,
                                        kernelVersion VARCHAR2(255),
                                        architecture VARCHAR2(100) NOT NULL,
                                        bootTime DATE NOT NULL,
                                        ipAddress VARCHAR2(45),
                                        macAddress VARCHAR2(17),
                                        serverLocation VARCHAR2(255),
                                        serverType VARCHAR2(50),
                                        lastUpdateTime DATE NOT NULL,
                                        networkInfo CLOB,
                                        systemInfo CLOB,
                                        hardwareInfo CLOB,
                                        addTime DATE DEFAULT SYSDATE NOT NULL,
                                        addWho VARCHAR2(64) NOT NULL,
                                        editTime DATE DEFAULT SYSDATE NOT NULL,
                                        editWho VARCHAR2(64) NOT NULL,
                                        oprSeqFlag VARCHAR2(64) NOT NULL,
                                        currentVersion NUMBER(10,0) DEFAULT 1 NOT NULL,
                                        activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL,
                                        noteText CLOB,
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
                                        CONSTRAINT PK_METRIC_SRVINFO PRIMARY KEY (tenantId, metricServerId)
);

CREATE UNIQUE INDEX IDX_SRVINFO_HOST ON HUB_METRIC_SERVER_INFO(hostname);
CREATE INDEX IDX_SRVINFO_OS ON HUB_METRIC_SERVER_INFO(osType);
CREATE INDEX IDX_SRVINFO_IP ON HUB_METRIC_SERVER_INFO(ipAddress);
CREATE INDEX IDX_SRVINFO_TYPE ON HUB_METRIC_SERVER_INFO(serverType);
CREATE INDEX IDX_SRVINFO_ACTIVE ON HUB_METRIC_SERVER_INFO(activeFlag);
CREATE INDEX IDX_SRVINFO_UPDATE ON HUB_METRIC_SERVER_INFO(lastUpdateTime);

COMMENT ON TABLE HUB_METRIC_SERVER_INFO IS '服务器信息主表';

-- 兼容历史项目：删除hostname唯一索引
DROP INDEX IDX_SRVINFO_HOST;
