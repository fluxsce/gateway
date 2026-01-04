CREATE TABLE HUB_METRIC_NETWORK_LOG (
                                        metricNetworkLogId VARCHAR2(32) NOT NULL,
                                        tenantId VARCHAR2(32) NOT NULL,
                                        metricServerId VARCHAR2(32) NOT NULL,
                                        interfaceName VARCHAR2(100) NOT NULL,
                                        hardwareAddr VARCHAR2(50),
                                        ipAddresses CLOB,
                                        interfaceStatus VARCHAR2(20) NOT NULL,
                                        interfaceType VARCHAR2(50),
                                        bytesReceived NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        bytesSent NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        packetsReceived NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        packetsSent NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        errorsReceived NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        errorsSent NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        droppedReceived NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        droppedSent NUMBER(19,0) DEFAULT 0 NOT NULL,
                                        receiveRate NUMBER(20,2) DEFAULT 0 NOT NULL,
                                        sendRate NUMBER(20,2) DEFAULT 0 NOT NULL,
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
                                        CONSTRAINT PK_NET_LOG PRIMARY KEY (tenantId, metricNetworkLogId)
);
CREATE INDEX IDX_NETLOG_SERVER ON HUB_METRIC_NETWORK_LOG(metricServerId);
CREATE INDEX IDX_NETLOG_TIME ON HUB_METRIC_NETWORK_LOG(collectTime);
CREATE INDEX IDX_NETLOG_IFACE ON HUB_METRIC_NETWORK_LOG(interfaceName);
CREATE INDEX IDX_NETLOG_STATUS ON HUB_METRIC_NETWORK_LOG(interfaceStatus);
CREATE INDEX IDX_NETLOG_ACTIVE ON HUB_METRIC_NETWORK_LOG(activeFlag);
CREATE INDEX IDX_NETLOG_SRV_TIME ON HUB_METRIC_NETWORK_LOG(metricServerId, collectTime);
CREATE INDEX IDX_NETLOG_SRV_IF ON HUB_METRIC_NETWORK_LOG(metricServerId, interfaceName);
CREATE INDEX IDX_NETLOG_TNT_TIME ON HUB_METRIC_NETWORK_LOG(tenantId, collectTime);
COMMENT ON TABLE HUB_METRIC_NETWORK_LOG IS '网络接口采集日志表';
