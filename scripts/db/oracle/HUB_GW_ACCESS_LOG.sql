CREATE TABLE HUB_GW_ACCESS_LOG (
                                   tenantId VARCHAR2(32) NOT NULL,
                                   traceId VARCHAR2(64) NOT NULL,
                                   gatewayInstanceId VARCHAR2(32) NOT NULL,
                                   gatewayInstanceName VARCHAR2(300),
                                   gatewayNodeIp VARCHAR2(50) NOT NULL,
                                   routeConfigId VARCHAR2(32),
                                   routeName VARCHAR2(300),
                                   serviceDefinitionId VARCHAR2(32),
                                   serviceName VARCHAR2(300),
                                   proxyType VARCHAR2(50),
                                   logConfigId VARCHAR2(32),

    -- 请求基本信息
                                   requestMethod VARCHAR2(10) NOT NULL,
                                   requestPath VARCHAR2(1000) NOT NULL,
                                   requestQuery CLOB,
                                   requestSize NUMBER(10) DEFAULT 0,
                                   requestHeaders CLOB,
                                   requestBody CLOB,

    -- 客户端信息
                                   clientIpAddress VARCHAR2(50) NOT NULL,
                                   clientPort NUMBER(10),
                                   userAgent VARCHAR2(1000),
                                   referer VARCHAR2(1000),
                                   userIdentifier VARCHAR2(100),

    -- 关键时间点 (Oracle使用TIMESTAMP类型，精确到毫秒)
                                   gatewayStartProcessingTime TIMESTAMP(3) NOT NULL,
                                   backendRequestStartTime TIMESTAMP(3),
                                   backendResponseReceivedTime TIMESTAMP(3),
                                   gatewayFinishedProcessingTime TIMESTAMP(3),

    -- 计算的时间指标 (毫秒)
                                   totalProcessingTimeMs NUMBER(10),
                                   gatewayProcessingTimeMs NUMBER(10),
                                   backendResponseTimeMs NUMBER(10),

    -- 响应信息
                                   gatewayStatusCode NUMBER(10) NOT NULL,
                                   backendStatusCode NUMBER(10),
                                   responseSize NUMBER(10) DEFAULT 0,
                                   responseHeaders CLOB,
                                   responseBody CLOB,

    -- 转发基本信息
                                   matchedRoute VARCHAR2(500),
                                   forwardAddress CLOB,
                                   forwardMethod VARCHAR2(10),
                                   forwardParams CLOB,
                                   forwardHeaders CLOB,
                                   forwardBody CLOB,
                                   loadBalancerDecision VARCHAR2(1000),

    -- 错误信息
                                   errorMessage CLOB,
                                   errorCode VARCHAR2(100),

    -- 追踪信息
                                   parentTraceId VARCHAR2(100),

    -- 日志重置标记和次数
                                   resetFlag VARCHAR2(1) DEFAULT 'N' NOT NULL,
                                   retryCount NUMBER(10) DEFAULT 0 NOT NULL,
                                   resetCount NUMBER(10) DEFAULT 0 NOT NULL,

    -- 标准数据库字段
                                   logLevel VARCHAR2(20) DEFAULT 'INFO' NOT NULL,
                                   logType VARCHAR2(50) DEFAULT 'ACCESS' NOT NULL,
                                   reserved1 VARCHAR2(100),
                                   reserved2 VARCHAR2(100),
                                   reserved3 NUMBER(10),
                                   reserved4 NUMBER(10),
                                   reserved5 TIMESTAMP,
                                   extProperty CLOB,
                                   addTime TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
                                   addWho VARCHAR2(32) NOT NULL,
                                   editTime TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
                                   editWho VARCHAR2(32) NOT NULL,
                                   oprSeqFlag VARCHAR2(32) NOT NULL,
                                   currentVersion NUMBER(10) DEFAULT 1 NOT NULL,
                                   activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL,
                                   noteText VARCHAR2(500),

                                   CONSTRAINT pk_HUB_GW_ACCESS_LOG PRIMARY KEY (tenantId, traceId)
);

COMMENT ON TABLE HUB_GW_ACCESS_LOG IS '网关访问日志表 - 记录API网关的请求和响应详细信息,开始时间必填,完成时间可选(支持处理中状态),含冗余字段优化查询性能';
CREATE INDEX idx_gw_log_time_inst ON HUB_GW_ACCESS_LOG (gatewayStartProcessingTime, gatewayInstanceId);
CREATE INDEX idx_gw_log_time_route ON HUB_GW_ACCESS_LOG (gatewayStartProcessingTime, routeConfigId);
CREATE INDEX idx_gw_log_time_service ON HUB_GW_ACCESS_LOG (gatewayStartProcessingTime, serviceDefinitionId);
CREATE INDEX idx_gw_log_inst_name ON HUB_GW_ACCESS_LOG (gatewayInstanceName, gatewayStartProcessingTime);
CREATE INDEX idx_gw_log_route_name ON HUB_GW_ACCESS_LOG (routeName, gatewayStartProcessingTime);
CREATE INDEX idx_gw_log_service_name ON HUB_GW_ACCESS_LOG (serviceName, gatewayStartProcessingTime);
CREATE INDEX idx_gw_log_client_ip ON HUB_GW_ACCESS_LOG (clientIpAddress, gatewayStartProcessingTime);
CREATE INDEX idx_gw_log_status_time ON HUB_GW_ACCESS_LOG (gatewayStatusCode, gatewayStartProcessingTime);
CREATE INDEX idx_gw_log_proxy_type ON HUB_GW_ACCESS_LOG (proxyType, gatewayStartProcessingTime);

