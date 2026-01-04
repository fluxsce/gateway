CREATE TABLE HUB_GW_CORS_CONFIG (
                                    tenantId VARCHAR2(32) NOT NULL,
                                    corsConfigId VARCHAR2(32) NOT NULL,
                                    gatewayInstanceId VARCHAR2(32) DEFAULT NULL,
                                    routeConfigId VARCHAR2(32) DEFAULT NULL,
                                    configName VARCHAR2(100) NOT NULL,
                                    allowOrigins CLOB NOT NULL,
                                    allowMethods VARCHAR2(200) DEFAULT 'GET,POST,PUT,DELETE,OPTIONS' NOT NULL,
                                    allowHeaders CLOB DEFAULT NULL,
                                    exposeHeaders CLOB DEFAULT NULL,
                                    allowCredentials VARCHAR2(1) DEFAULT 'N' NOT NULL,
                                    maxAgeSeconds NUMBER(10) DEFAULT 86400 NOT NULL,
                                    configPriority NUMBER(10) DEFAULT 0 NOT NULL,
                                    reserved1 VARCHAR2(100) DEFAULT NULL,
                                    reserved2 VARCHAR2(100) DEFAULT NULL,
                                    reserved3 NUMBER(10) DEFAULT NULL,
                                    reserved4 NUMBER(10) DEFAULT NULL,
                                    reserved5 TIMESTAMP DEFAULT NULL,
                                    extProperty CLOB DEFAULT NULL,
                                    addTime TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
                                    addWho VARCHAR2(32) NOT NULL,
                                    editTime TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
                                    editWho VARCHAR2(32) NOT NULL,
                                    oprSeqFlag VARCHAR2(32) NOT NULL,
                                    currentVersion NUMBER(10) DEFAULT 1 NOT NULL,
                                    activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL,
                                    noteText VARCHAR2(500) DEFAULT NULL,
                                    CONSTRAINT PK_HUB_GW_CORS_CONFIG PRIMARY KEY (tenantId, corsConfigId)
);

COMMENT ON TABLE HUB_GW_CORS_CONFIG IS '跨域配置表 - 存储CORS相关配置';

CREATE INDEX IDX_GW_CORS_INST ON HUB_GW_CORS_CONFIG (gatewayInstanceId);
CREATE INDEX IDX_GW_CORS_ROUTE ON HUB_GW_CORS_CONFIG (routeConfigId);
CREATE INDEX IDX_GW_CORS_PRIORITY ON HUB_GW_CORS_CONFIG (configPriority);

