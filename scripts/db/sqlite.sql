-- =====================================================
-- SQLite数据库初始化脚本
-- 基于MySQL脚本直接翻译，保持原有表结构
-- 创建时间: 2024-12-19
-- 说明: 
-- 1. 保持与MySQL相同的表结构和字段名
-- 2. 将MySQL数据类型映射为SQLite对应类型
-- 3. 不添加额外的CHECK约束
-- 4. 保持原有的索引和约束逻辑
-- =====================================================

-- 启用外键约束
PRAGMA foreign_keys = ON;

-- 启用WAL模式以支持并发
PRAGMA journal_mode = WAL;

-- 1. 用户信息表
CREATE TABLE IF NOT EXISTS HUB_USER (
    userId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    userName TEXT NOT NULL,
    password TEXT NOT NULL,
    realName TEXT NOT NULL,
    deptId TEXT NOT NULL,
    email TEXT,
    mobile TEXT,
    avatar TEXT,
    gender INTEGER DEFAULT 0,
    statusFlag TEXT NOT NULL DEFAULT 'Y',
    deptAdminFlag TEXT NOT NULL DEFAULT 'N',
    tenantAdminFlag TEXT NOT NULL DEFAULT 'N',
    userExpireDate DATETIME NOT NULL,
    lastLoginTime DATETIME,
    lastLoginIp TEXT,
    pwdUpdateTime DATETIME,
    pwdErrorCount INTEGER NOT NULL DEFAULT 0,
    lockTime DATETIME,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (userId, tenantId)
);

CREATE INDEX UK_USER_NAME_TENANT ON HUB_USER(userName, tenantId);
CREATE INDEX IDX_USER_TENANT ON HUB_USER(tenantId);
CREATE INDEX IDX_USER_DEPT ON HUB_USER(deptId);
CREATE INDEX IDX_USER_STATUS ON HUB_USER(statusFlag);
CREATE INDEX IDX_USER_EMAIL ON HUB_USER(email);
CREATE INDEX IDX_USER_MOBILE ON HUB_USER(mobile);

INSERT INTO HUB_USER (
    userId,
    tenantId,
    userName,
    password,
    realName,
    deptId,
    email,
    mobile,
    avatar,
    gender,
    statusFlag,
    deptAdminFlag,
    tenantAdminFlag,
    userExpireDate,
    oprSeqFlag,
    currentVersion,
    activeFlag,
    addWho,
    editWho,
    noteText
) VALUES (
    'admin',
    'default',
    'admin',
    '123456',
    '系统管理员',
    'D00000001',
    'admin@example.com',
    '13800000000',
    'https://example.com/avatar.png',
    1,
    'Y',
    'N',
    'Y',
    datetime('now', '+5 years'),
    'SEQFLAG_001',
    1,
    'Y',
    'system',
    'system',
    '系统初始化管理员账号'
);

-- 2. 用户登录日志表
CREATE TABLE IF NOT EXISTS HUB_LOGIN_LOG (
    logId TEXT NOT NULL,
    userId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    userName TEXT NOT NULL,
    loginTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    loginIp TEXT NOT NULL,
    loginLocation TEXT,
    loginType INTEGER NOT NULL DEFAULT 1,
    deviceType TEXT,
    deviceInfo TEXT,
    browserInfo TEXT,
    osInfo TEXT,
    loginStatus TEXT NOT NULL DEFAULT 'N',
    logoutTime DATETIME,
    sessionDuration INTEGER,
    failReason TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    PRIMARY KEY (logId)
);

CREATE INDEX IDX_LOGIN_USER ON HUB_LOGIN_LOG(userId);
CREATE INDEX IDX_LOGIN_TIME ON HUB_LOGIN_LOG(loginTime);
CREATE INDEX IDX_LOGIN_TENANT ON HUB_LOGIN_LOG(tenantId);
CREATE INDEX IDX_LOGIN_STATUS ON HUB_LOGIN_LOG(loginStatus);
CREATE INDEX IDX_LOGIN_TYPE ON HUB_LOGIN_LOG(loginType);

-- 3. 网关实例表
CREATE TABLE IF NOT EXISTS HUB_GW_INSTANCE (
    tenantId TEXT NOT NULL,
    gatewayInstanceId TEXT NOT NULL,
    instanceName TEXT NOT NULL,
    instanceDesc TEXT,
    bindAddress TEXT DEFAULT '0.0.0.0',
    httpPort INTEGER,
    httpsPort INTEGER,
    tlsEnabled TEXT NOT NULL DEFAULT 'N',
    certStorageType TEXT NOT NULL DEFAULT 'FILE',
    certFilePath TEXT,
    keyFilePath TEXT,
    certContent TEXT,
    keyContent TEXT,
    certChainContent TEXT,
    certPassword TEXT,
    maxConnections INTEGER NOT NULL DEFAULT 10000,
    readTimeoutMs INTEGER NOT NULL DEFAULT 30000,
    writeTimeoutMs INTEGER NOT NULL DEFAULT 30000,
    idleTimeoutMs INTEGER NOT NULL DEFAULT 60000,
    maxHeaderBytes INTEGER NOT NULL DEFAULT 1048576,
    maxWorkers INTEGER NOT NULL DEFAULT 1000,
    keepAliveEnabled TEXT NOT NULL DEFAULT 'Y',
    tcpKeepAliveEnabled TEXT NOT NULL DEFAULT 'Y',
    gracefulShutdownTimeoutMs INTEGER NOT NULL DEFAULT 30000,
    enableHttp2 TEXT NOT NULL DEFAULT 'Y',
    tlsVersion TEXT DEFAULT '1.2',
    tlsCipherSuites TEXT,
    disableGeneralOptionsHandler TEXT NOT NULL DEFAULT 'N',
    logConfigId TEXT,
    healthStatus TEXT NOT NULL DEFAULT 'Y',
    lastHeartbeatTime DATETIME,
    instanceMetadata TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, gatewayInstanceId)
);

CREATE INDEX IDX_GW_INST_BIND_HTTP ON HUB_GW_INSTANCE(bindAddress, httpPort);
CREATE INDEX IDX_GW_INST_BIND_HTTPS ON HUB_GW_INSTANCE(bindAddress, httpsPort);
CREATE INDEX IDX_GW_INST_LOG ON HUB_GW_INSTANCE(logConfigId);
CREATE INDEX IDX_GW_INST_HEALTH ON HUB_GW_INSTANCE(healthStatus);
CREATE INDEX IDX_GW_INST_TLS ON HUB_GW_INSTANCE(tlsEnabled);

-- 4. Router配置表
CREATE TABLE IF NOT EXISTS HUB_GW_ROUTER_CONFIG (
    tenantId TEXT NOT NULL,
    routerConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT NOT NULL,
    routerName TEXT NOT NULL,
    routerDesc TEXT,
    defaultPriority INTEGER NOT NULL DEFAULT 100,
    enableRouteCache TEXT NOT NULL DEFAULT 'Y',
    routeCacheTtlSeconds INTEGER NOT NULL DEFAULT 300,
    maxRoutes INTEGER DEFAULT 1000,
    routeMatchTimeout INTEGER DEFAULT 100,
    enableStrictMode TEXT NOT NULL DEFAULT 'N',
    enableMetrics TEXT NOT NULL DEFAULT 'Y',
    enableTracing TEXT NOT NULL DEFAULT 'N',
    caseSensitive TEXT NOT NULL DEFAULT 'Y',
    removeTrailingSlash TEXT NOT NULL DEFAULT 'Y',
    enableGlobalFilters TEXT NOT NULL DEFAULT 'Y',
    filterExecutionMode TEXT NOT NULL DEFAULT 'SEQUENTIAL',
    maxFilterChainDepth INTEGER DEFAULT 50,
    enableRoutePooling TEXT NOT NULL DEFAULT 'N',
    routePoolSize INTEGER DEFAULT 100,
    enableAsyncProcessing TEXT NOT NULL DEFAULT 'N',
    enableFallback TEXT NOT NULL DEFAULT 'Y',
    fallbackRoute TEXT,
    notFoundStatusCode INTEGER NOT NULL DEFAULT 404,
    notFoundMessage TEXT DEFAULT 'Route not found',
    routerMetadata TEXT,
    customConfig TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, routerConfigId)
);

CREATE INDEX IDX_GW_ROUTER_INST ON HUB_GW_ROUTER_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_ROUTER_NAME ON HUB_GW_ROUTER_CONFIG(routerName);
CREATE INDEX IDX_GW_ROUTER_ACTIVE ON HUB_GW_ROUTER_CONFIG(activeFlag);
CREATE INDEX IDX_GW_ROUTER_CACHE ON HUB_GW_ROUTER_CONFIG(enableRouteCache);

-- 5. 路由定义表
CREATE TABLE IF NOT EXISTS HUB_GW_ROUTE_CONFIG (
    tenantId TEXT NOT NULL,
    routeConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT NOT NULL,
    routeName TEXT NOT NULL,
    routePath TEXT NOT NULL,
    allowedMethods TEXT,
    allowedHosts TEXT,
    matchType INTEGER NOT NULL DEFAULT 1,
    routePriority INTEGER NOT NULL DEFAULT 100,
    stripPathPrefix TEXT NOT NULL DEFAULT 'N',
    rewritePath TEXT,
    enableWebsocket TEXT NOT NULL DEFAULT 'N',
    timeoutMs INTEGER NOT NULL DEFAULT 30000,
    retryCount INTEGER NOT NULL DEFAULT 0,
    retryIntervalMs INTEGER NOT NULL DEFAULT 1000,
    serviceDefinitionId TEXT,
    logConfigId TEXT,
    routeMetadata TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, routeConfigId)
);

CREATE INDEX IDX_GW_ROUTE_INST ON HUB_GW_ROUTE_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_ROUTE_SERVICE ON HUB_GW_ROUTE_CONFIG(serviceDefinitionId);
CREATE INDEX IDX_GW_ROUTE_LOG ON HUB_GW_ROUTE_CONFIG(logConfigId);
CREATE INDEX IDX_GW_ROUTE_PRIORITY ON HUB_GW_ROUTE_CONFIG(routePriority);
CREATE INDEX IDX_GW_ROUTE_PATH ON HUB_GW_ROUTE_CONFIG(routePath);
CREATE INDEX IDX_GW_ROUTE_ACTIVE ON HUB_GW_ROUTE_CONFIG(activeFlag);

-- 6. 路由断言表
CREATE TABLE IF NOT EXISTS HUB_GW_ROUTE_ASSERTION (
    tenantId TEXT NOT NULL,
    routeAssertionId TEXT NOT NULL,
    routeConfigId TEXT NOT NULL,
    assertionName TEXT NOT NULL,
    assertionType TEXT NOT NULL,
    assertionOperator TEXT NOT NULL DEFAULT 'EQUAL',
    fieldName TEXT,
    expectedValue TEXT,
    patternValue TEXT,
    caseSensitive TEXT NOT NULL DEFAULT 'Y',
    assertionOrder INTEGER NOT NULL DEFAULT 0,
    isRequired TEXT NOT NULL DEFAULT 'Y',
    assertionDesc TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, routeAssertionId)
);

CREATE INDEX IDX_GW_ASSERT_ROUTE ON HUB_GW_ROUTE_ASSERTION(routeConfigId);
CREATE INDEX IDX_GW_ASSERT_TYPE ON HUB_GW_ROUTE_ASSERTION(assertionType);
CREATE INDEX IDX_GW_ASSERT_ORDER ON HUB_GW_ROUTE_ASSERTION(assertionOrder);

-- 7. 过滤器配置表
CREATE TABLE IF NOT EXISTS HUB_GW_FILTER_CONFIG (
    tenantId TEXT NOT NULL,
    filterConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    filterName TEXT NOT NULL,
    filterType TEXT NOT NULL,
    filterAction TEXT NOT NULL,
    filterOrder INTEGER NOT NULL DEFAULT 0,
    filterConfig TEXT NOT NULL,
    filterDesc TEXT,
    configId TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, filterConfigId)
);

CREATE INDEX IDX_GW_FILTER_INST ON HUB_GW_FILTER_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_FILTER_ROUTE ON HUB_GW_FILTER_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_FILTER_TYPE ON HUB_GW_FILTER_CONFIG(filterType);
CREATE INDEX IDX_GW_FILTER_ACTION ON HUB_GW_FILTER_CONFIG(filterAction);
CREATE INDEX IDX_GW_FILTER_ORDER ON HUB_GW_FILTER_CONFIG(filterOrder);
CREATE INDEX IDX_GW_FILTER_ACTIVE ON HUB_GW_FILTER_CONFIG(activeFlag);

-- 8. 跨域配置表
CREATE TABLE IF NOT EXISTS HUB_GW_CORS_CONFIG (
    tenantId TEXT NOT NULL,
    corsConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    configName TEXT NOT NULL,
    allowOrigins TEXT NOT NULL,
    allowMethods TEXT NOT NULL DEFAULT 'GET,POST,PUT,DELETE,OPTIONS',
    allowHeaders TEXT,
    exposeHeaders TEXT,
    allowCredentials TEXT NOT NULL DEFAULT 'N',
    maxAgeSeconds INTEGER NOT NULL DEFAULT 86400,
    configPriority INTEGER NOT NULL DEFAULT 0,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, corsConfigId)
);

CREATE INDEX IDX_GW_CORS_INST ON HUB_GW_CORS_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_CORS_ROUTE ON HUB_GW_CORS_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_CORS_PRIORITY ON HUB_GW_CORS_CONFIG(configPriority);

-- 9. 限流配置表
CREATE TABLE IF NOT EXISTS HUB_GW_RATE_LIMIT_CONFIG (
    tenantId TEXT NOT NULL,
    rateLimitConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    limitName TEXT NOT NULL,
    algorithm TEXT NOT NULL DEFAULT 'token-bucket',
    keyStrategy TEXT NOT NULL DEFAULT 'ip',
    limitRate INTEGER NOT NULL,
    burstCapacity INTEGER NOT NULL DEFAULT 0,
    timeWindowSeconds INTEGER NOT NULL DEFAULT 1,
    rejectionStatusCode INTEGER NOT NULL DEFAULT 429,
    rejectionMessage TEXT DEFAULT '请求过于频繁，请稍后再试',
    configPriority INTEGER NOT NULL DEFAULT 0,
    customConfig TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, rateLimitConfigId)
);

CREATE INDEX IDX_GW_RATE_INST ON HUB_GW_RATE_LIMIT_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_RATE_ROUTE ON HUB_GW_RATE_LIMIT_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_RATE_STRATEGY ON HUB_GW_RATE_LIMIT_CONFIG(keyStrategy);
CREATE INDEX IDX_GW_RATE_ALGORITHM ON HUB_GW_RATE_LIMIT_CONFIG(algorithm);
CREATE INDEX IDX_GW_RATE_PRIORITY ON HUB_GW_RATE_LIMIT_CONFIG(configPriority);
CREATE INDEX IDX_GW_RATE_ACTIVE ON HUB_GW_RATE_LIMIT_CONFIG(activeFlag);

-- 10. 熔断配置表
CREATE TABLE IF NOT EXISTS HUB_GW_CIRCUIT_BREAKER_CONFIG (
    tenantId TEXT NOT NULL,
    circuitBreakerConfigId TEXT NOT NULL,
    routeConfigId TEXT,
    targetServiceId TEXT,
    breakerName TEXT NOT NULL,
    keyStrategy TEXT NOT NULL DEFAULT 'api',
    errorRatePercent INTEGER NOT NULL DEFAULT 50,
    minimumRequests INTEGER NOT NULL DEFAULT 10,
    halfOpenMaxRequests INTEGER NOT NULL DEFAULT 3,
    slowCallThreshold INTEGER NOT NULL DEFAULT 1000,
    slowCallRatePercent INTEGER NOT NULL DEFAULT 50,
    openTimeoutSeconds INTEGER NOT NULL DEFAULT 60,
    windowSizeSeconds INTEGER NOT NULL DEFAULT 60,
    errorStatusCode INTEGER NOT NULL DEFAULT 503,
    errorMessage TEXT DEFAULT 'Service temporarily unavailable due to circuit breaker',
    storageType TEXT NOT NULL DEFAULT 'memory',
    storageConfig TEXT,
    configPriority INTEGER NOT NULL DEFAULT 0,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, circuitBreakerConfigId)
);

CREATE INDEX IDX_GW_CB_ROUTE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_CB_SERVICE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(targetServiceId);
CREATE INDEX IDX_GW_CB_STRATEGY ON HUB_GW_CIRCUIT_BREAKER_CONFIG(keyStrategy);
CREATE INDEX IDX_GW_CB_STORAGE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(storageType);
CREATE INDEX IDX_GW_CB_PRIORITY ON HUB_GW_CIRCUIT_BREAKER_CONFIG(configPriority);

-- 11. 认证配置表
CREATE TABLE IF NOT EXISTS HUB_GW_AUTH_CONFIG (
    tenantId TEXT NOT NULL,
    authConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    authName TEXT NOT NULL,
    authType TEXT NOT NULL,
    authStrategy TEXT DEFAULT 'REQUIRED',
    authConfig TEXT NOT NULL,
    exemptPaths TEXT,
    exemptHeaders TEXT,
    failureStatusCode INTEGER NOT NULL DEFAULT 401,
    failureMessage TEXT DEFAULT '认证失败',
    configPriority INTEGER NOT NULL DEFAULT 0,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, authConfigId)
);

CREATE INDEX IDX_GW_AUTH_INST ON HUB_GW_AUTH_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_AUTH_ROUTE ON HUB_GW_AUTH_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_AUTH_TYPE ON HUB_GW_AUTH_CONFIG(authType);
CREATE INDEX IDX_GW_AUTH_PRIORITY ON HUB_GW_AUTH_CONFIG(configPriority);

-- 12. 服务定义表
CREATE TABLE IF NOT EXISTS HUB_GW_SERVICE_DEFINITION (
    tenantId TEXT NOT NULL,
    serviceDefinitionId TEXT NOT NULL,
    serviceName TEXT NOT NULL,
    serviceDesc TEXT,
    serviceType INTEGER NOT NULL DEFAULT 0,
    proxyConfigId TEXT NOT NULL,
    loadBalanceStrategy TEXT NOT NULL DEFAULT 'round-robin',
    discoveryType TEXT,
    discoveryConfig TEXT,
    sessionAffinity TEXT NOT NULL DEFAULT 'N',
    stickySession TEXT NOT NULL DEFAULT 'N',
    maxRetries INTEGER NOT NULL DEFAULT 3,
    retryTimeoutMs INTEGER NOT NULL DEFAULT 5000,
    enableCircuitBreaker TEXT NOT NULL DEFAULT 'N',
    healthCheckEnabled TEXT NOT NULL DEFAULT 'Y',
    healthCheckPath TEXT DEFAULT '/health',
    healthCheckMethod TEXT DEFAULT 'GET',
    healthCheckIntervalSeconds INTEGER DEFAULT 30,
    healthCheckTimeoutMs INTEGER DEFAULT 5000,
    healthyThreshold INTEGER DEFAULT 2,
    unhealthyThreshold INTEGER DEFAULT 3,
    expectedStatusCodes TEXT DEFAULT '200',
    healthCheckHeaders TEXT,
    loadBalancerConfig TEXT,
    serviceMetadata TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, serviceDefinitionId)
);

CREATE INDEX IDX_GW_SVC_NAME ON HUB_GW_SERVICE_DEFINITION(serviceName);
CREATE INDEX IDX_GW_SVC_TYPE ON HUB_GW_SERVICE_DEFINITION(serviceType);
CREATE INDEX IDX_GW_SVC_STRATEGY ON HUB_GW_SERVICE_DEFINITION(loadBalanceStrategy);
CREATE INDEX IDX_GW_SVC_HEALTH ON HUB_GW_SERVICE_DEFINITION(healthCheckEnabled);
CREATE INDEX IDX_GW_SVC_PROXY ON HUB_GW_SERVICE_DEFINITION(proxyConfigId);

-- 13. 服务节点表
CREATE TABLE IF NOT EXISTS HUB_GW_SERVICE_NODE (
    tenantId TEXT NOT NULL,
    serviceNodeId TEXT NOT NULL,
    serviceDefinitionId TEXT NOT NULL,
    nodeId TEXT NOT NULL,
    nodeUrl TEXT NOT NULL,
    nodeHost TEXT NOT NULL,
    nodePort INTEGER NOT NULL,
    nodeProtocol TEXT NOT NULL DEFAULT 'HTTP',
    nodeWeight INTEGER NOT NULL DEFAULT 100,
    healthStatus TEXT NOT NULL DEFAULT 'Y',
    nodeMetadata TEXT,
    nodeStatus INTEGER NOT NULL DEFAULT 1,
    lastHealthCheckTime DATETIME,
    healthCheckResult TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, serviceNodeId)
);

CREATE INDEX IDX_GW_NODE_SERVICE ON HUB_GW_SERVICE_NODE(serviceDefinitionId);
CREATE INDEX IDX_GW_NODE_ENDPOINT ON HUB_GW_SERVICE_NODE(nodeHost, nodePort);
CREATE INDEX IDX_GW_NODE_HEALTH ON HUB_GW_SERVICE_NODE(healthStatus);
CREATE INDEX IDX_GW_NODE_STATUS ON HUB_GW_SERVICE_NODE(nodeStatus);

-- 14. 代理配置表
CREATE TABLE IF NOT EXISTS HUB_GW_PROXY_CONFIG (
    tenantId TEXT NOT NULL,
    proxyConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT NOT NULL,
    proxyName TEXT NOT NULL,
    proxyType TEXT NOT NULL DEFAULT 'http',
    proxyId TEXT,
    configPriority INTEGER NOT NULL DEFAULT 0,
    proxyConfig TEXT NOT NULL,
    customConfig TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, proxyConfigId)
);

CREATE INDEX IDX_GW_PROXY_INST ON HUB_GW_PROXY_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_PROXY_TYPE ON HUB_GW_PROXY_CONFIG(proxyType);
CREATE INDEX IDX_GW_PROXY_PRIORITY ON HUB_GW_PROXY_CONFIG(configPriority);
CREATE INDEX IDX_GW_PROXY_ACTIVE ON HUB_GW_PROXY_CONFIG(activeFlag);

-- 15. 定时任务调度器表
CREATE TABLE IF NOT EXISTS HUB_TIMER_SCHEDULER (
    schedulerId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    schedulerName TEXT NOT NULL,
    schedulerInstanceId TEXT,
    maxWorkers INTEGER NOT NULL DEFAULT 5,
    queueSize INTEGER NOT NULL DEFAULT 100,
    defaultTimeoutSeconds INTEGER NOT NULL DEFAULT 1800,
    defaultRetries INTEGER NOT NULL DEFAULT 3,
    schedulerStatus INTEGER NOT NULL DEFAULT 1,
    lastStartTime DATETIME,
    lastStopTime DATETIME,
    serverName TEXT,
    serverIp TEXT,
    serverPort INTEGER,
    totalTaskCount INTEGER NOT NULL DEFAULT 0,
    runningTaskCount INTEGER NOT NULL DEFAULT 0,
    lastHeartbeatTime DATETIME,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, schedulerId)
);

CREATE INDEX IDX_TIMER_SCHED_NAME ON HUB_TIMER_SCHEDULER(schedulerName);
CREATE INDEX IDX_TIMER_SCHED_INST ON HUB_TIMER_SCHEDULER(schedulerInstanceId);
CREATE INDEX IDX_TIMER_SCHED_STATUS ON HUB_TIMER_SCHEDULER(schedulerStatus);
CREATE INDEX IDX_TIMER_SCHED_HEART ON HUB_TIMER_SCHEDULER(lastHeartbeatTime);

-- 16. 定时任务表
CREATE TABLE IF NOT EXISTS HUB_TIMER_TASK (
    taskId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    taskName TEXT NOT NULL,
    taskDescription TEXT,
    taskPriority INTEGER NOT NULL DEFAULT 1,
    schedulerId TEXT,
    schedulerName TEXT,
    scheduleType INTEGER NOT NULL,
    cronExpression TEXT,
    intervalSeconds INTEGER,
    delaySeconds INTEGER,
    startTime DATETIME,
    endTime DATETIME,
    maxRetries INTEGER NOT NULL DEFAULT 0,
    retryIntervalSeconds INTEGER NOT NULL DEFAULT 60,
    timeoutSeconds INTEGER NOT NULL DEFAULT 1800,
    taskParams TEXT,
    executorType TEXT,
    toolConfigId TEXT,
    toolConfigName TEXT,
    operationType TEXT,
    operationConfig TEXT,
    taskStatus INTEGER NOT NULL DEFAULT 1,
    nextRunTime DATETIME,
    lastRunTime DATETIME,
    runCount INTEGER NOT NULL DEFAULT 0,
    successCount INTEGER NOT NULL DEFAULT 0,
    failureCount INTEGER NOT NULL DEFAULT 0,
    lastExecutionId TEXT,
    lastExecutionStartTime DATETIME,
    lastExecutionEndTime DATETIME,
    lastExecutionDurationMs INTEGER,
    lastExecutionStatus INTEGER,
    lastResultSuccess TEXT,
    lastErrorMessage TEXT,
    lastRetryCount INTEGER,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, taskId)
);

CREATE INDEX IDX_TIMER_TASK_NAME ON HUB_TIMER_TASK(taskName);
CREATE INDEX IDX_TIMER_TASK_SCHED ON HUB_TIMER_TASK(schedulerId);
CREATE INDEX IDX_TIMER_TASK_TYPE ON HUB_TIMER_TASK(scheduleType);
CREATE INDEX IDX_TIMER_TASK_STATUS ON HUB_TIMER_TASK(taskStatus);
CREATE INDEX IDX_TIMER_TASK_NEXT ON HUB_TIMER_TASK(nextRunTime);
CREATE INDEX IDX_TIMER_TASK_LAST ON HUB_TIMER_TASK(lastRunTime);
CREATE INDEX IDX_TIMER_TASK_ACTIVE ON HUB_TIMER_TASK(activeFlag);
CREATE INDEX IDX_TIMER_TASK_EXEC ON HUB_TIMER_TASK(executorType);
CREATE INDEX IDX_TIMER_TASK_TOOL ON HUB_TIMER_TASK(toolConfigId);
CREATE INDEX IDX_TIMER_TASK_OP ON HUB_TIMER_TASK(operationType);

-- 17. 任务执行日志表
CREATE TABLE IF NOT EXISTS HUB_TIMER_EXECUTION_LOG (
    executionId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    taskId TEXT NOT NULL,
    taskName TEXT,
    schedulerId TEXT,
    executionStartTime DATETIME NOT NULL,
    executionEndTime DATETIME,
    executionDurationMs INTEGER,
    executionStatus INTEGER NOT NULL,
    resultSuccess TEXT NOT NULL DEFAULT 'N',
    errorMessage TEXT,
    errorStackTrace TEXT,
    retryCount INTEGER NOT NULL DEFAULT 0,
    maxRetryCount INTEGER NOT NULL DEFAULT 0,
    executionParams TEXT,
    executionResult TEXT,
    executorServerName TEXT,
    executorServerIp TEXT,
    logLevel TEXT,
    logMessage TEXT,
    logTimestamp DATETIME,
    executionPhase TEXT,
    threadName TEXT,
    className TEXT,
    methodName TEXT,
    exceptionClass TEXT,
    exceptionMessage TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, executionId)
);

CREATE INDEX IDX_TIMER_LOG_TASK ON HUB_TIMER_EXECUTION_LOG(taskId);
CREATE INDEX IDX_TIMER_LOG_NAME ON HUB_TIMER_EXECUTION_LOG(taskName);
CREATE INDEX IDX_TIMER_LOG_SCHED ON HUB_TIMER_EXECUTION_LOG(schedulerId);
CREATE INDEX IDX_TIMER_LOG_START ON HUB_TIMER_EXECUTION_LOG(executionStartTime);
CREATE INDEX IDX_TIMER_LOG_STATUS ON HUB_TIMER_EXECUTION_LOG(executionStatus);
CREATE INDEX IDX_TIMER_LOG_SUCCESS ON HUB_TIMER_EXECUTION_LOG(resultSuccess);
CREATE INDEX IDX_TIMER_LOG_LEVEL ON HUB_TIMER_EXECUTION_LOG(logLevel);
CREATE INDEX IDX_TIMER_LOG_TIME ON HUB_TIMER_EXECUTION_LOG(logTimestamp);

-- 18. 工具配置主表
CREATE TABLE IF NOT EXISTS HUB_TOOL_CONFIG (
    toolConfigId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    toolName TEXT NOT NULL,
    toolType TEXT NOT NULL,
    toolVersion TEXT,
    configName TEXT NOT NULL,
    configDescription TEXT,
    configGroupId TEXT,
    configGroupName TEXT,
    hostAddress TEXT,
    portNumber INTEGER,
    protocolType TEXT,
    authType TEXT,
    userName TEXT,
    passwordEncrypted TEXT,
    keyFilePath TEXT,
    keyFileContent TEXT,
    configParameters TEXT,
    environmentVariables TEXT,
    customSettings TEXT,
    configStatus TEXT NOT NULL DEFAULT 'Y',
    defaultFlag TEXT NOT NULL DEFAULT 'N',
    priorityLevel INTEGER DEFAULT 100,
    encryptionType TEXT,
    encryptionKey TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, toolConfigId)
);

CREATE INDEX IDX_TOOL_CONFIG_NAME ON HUB_TOOL_CONFIG(toolName);
CREATE INDEX IDX_TOOL_CONFIG_TYPE ON HUB_TOOL_CONFIG(toolType);
CREATE INDEX IDX_TOOL_CONFIG_CFGNAME ON HUB_TOOL_CONFIG(configName);
CREATE INDEX IDX_TOOL_CONFIG_GROUP ON HUB_TOOL_CONFIG(configGroupId);
CREATE INDEX IDX_TOOL_CONFIG_STATUS ON HUB_TOOL_CONFIG(configStatus);
CREATE INDEX IDX_TOOL_CONFIG_DEFAULT ON HUB_TOOL_CONFIG(defaultFlag);
CREATE INDEX IDX_TOOL_CONFIG_ACTIVE ON HUB_TOOL_CONFIG(activeFlag);

-- 19. 工具配置分组表
CREATE TABLE IF NOT EXISTS HUB_TOOL_CONFIG_GROUP (
    configGroupId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    groupName TEXT NOT NULL,
    groupDescription TEXT,
    parentGroupId TEXT,
    groupLevel INTEGER DEFAULT 1,
    groupPath TEXT,
    groupType TEXT,
    sortOrder INTEGER DEFAULT 100,
    groupIcon TEXT,
    groupColor TEXT,
    accessLevel TEXT DEFAULT 'private',
    allowedUsers TEXT,
    allowedRoles TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, configGroupId)
);

CREATE INDEX IDX_TOOL_GROUP_NAME ON HUB_TOOL_CONFIG_GROUP(groupName);
CREATE INDEX IDX_TOOL_GROUP_PARENT ON HUB_TOOL_CONFIG_GROUP(parentGroupId);
CREATE INDEX IDX_TOOL_GROUP_TYPE ON HUB_TOOL_CONFIG_GROUP(groupType);
CREATE INDEX IDX_TOOL_GROUP_SORT ON HUB_TOOL_CONFIG_GROUP(sortOrder);
CREATE INDEX IDX_TOOL_GROUP_ACTIVE ON HUB_TOOL_CONFIG_GROUP(activeFlag);

-- 20. 日志配置表
CREATE TABLE IF NOT EXISTS HUB_GW_LOG_CONFIG (
    tenantId TEXT NOT NULL,
    logConfigId TEXT NOT NULL,
    configName TEXT NOT NULL,
    configDesc TEXT,
    logFormat TEXT NOT NULL DEFAULT 'JSON',
    recordRequestBody TEXT NOT NULL DEFAULT 'N',
    recordResponseBody TEXT NOT NULL DEFAULT 'N',
    recordHeaders TEXT NOT NULL DEFAULT 'Y',
    maxBodySizeBytes INTEGER NOT NULL DEFAULT 4096,
    outputTargets TEXT NOT NULL DEFAULT 'CONSOLE',
    fileConfig TEXT,
    databaseConfig TEXT,
    mongoConfig TEXT,
    elasticsearchConfig TEXT,
    clickhouseConfig TEXT,
    enableAsyncLogging TEXT NOT NULL DEFAULT 'Y',
    asyncQueueSize INTEGER NOT NULL DEFAULT 10000,
    asyncFlushIntervalMs INTEGER NOT NULL DEFAULT 1000,
    enableBatchProcessing TEXT NOT NULL DEFAULT 'Y',
    batchSize INTEGER NOT NULL DEFAULT 100,
    batchTimeoutMs INTEGER NOT NULL DEFAULT 5000,
    logRetentionDays INTEGER NOT NULL DEFAULT 30,
    enableFileRotation TEXT NOT NULL DEFAULT 'Y',
    maxFileSizeMB INTEGER DEFAULT 100,
    maxFileCount INTEGER DEFAULT 10,
    rotationPattern TEXT DEFAULT 'DAILY',
    enableSensitiveDataMasking TEXT NOT NULL DEFAULT 'Y',
    sensitiveFields TEXT,
    maskingPattern TEXT DEFAULT '***',
    bufferSize INTEGER NOT NULL DEFAULT 8192,
    flushThreshold INTEGER NOT NULL DEFAULT 100,
    configPriority INTEGER NOT NULL DEFAULT 0,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, logConfigId)
);

CREATE INDEX IF NOT EXISTS idx_HUB_GW_LOG_CONFIG_name ON HUB_GW_LOG_CONFIG(configName);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_LOG_CONFIG_priority ON HUB_GW_LOG_CONFIG(configPriority);

-- 21. 网关访问日志表
CREATE TABLE IF NOT EXISTS HUB_GW_ACCESS_LOG (
    tenantId TEXT NOT NULL,
    traceId TEXT NOT NULL,
    gatewayInstanceId TEXT NOT NULL,
    gatewayInstanceName TEXT,
    gatewayNodeIp TEXT NOT NULL,
    routeConfigId TEXT,
    routeName TEXT,
    serviceDefinitionId TEXT,
    serviceName TEXT,
    proxyType TEXT,
    logConfigId TEXT,
    requestMethod TEXT NOT NULL,
    requestPath TEXT NOT NULL,
    requestQuery TEXT,
    requestSize INTEGER DEFAULT 0,
    requestHeaders TEXT,
    requestBody TEXT,
    clientIpAddress TEXT NOT NULL,
    clientPort INTEGER,
    userAgent TEXT,
    referer TEXT,
    userIdentifier TEXT,
    gatewayStartProcessingTime DATETIME NOT NULL,
    backendRequestStartTime DATETIME,
    backendResponseReceivedTime DATETIME,
    gatewayFinishedProcessingTime DATETIME,
    totalProcessingTimeMs INTEGER,
    gatewayProcessingTimeMs INTEGER,
    backendResponseTimeMs INTEGER,
    gatewayStatusCode INTEGER NOT NULL,
    backendStatusCode INTEGER,
    responseSize INTEGER DEFAULT 0,
    responseHeaders TEXT,
    responseBody TEXT,
    matchedRoute TEXT,
    forwardAddress TEXT,
    forwardMethod TEXT,
    forwardParams TEXT,
    forwardHeaders TEXT,
    forwardBody TEXT,
    loadBalancerDecision TEXT,
    errorMessage TEXT,
    errorCode TEXT,
    parentTraceId TEXT,
    resetFlag TEXT NOT NULL DEFAULT 'N',
    retryCount INTEGER NOT NULL DEFAULT 0,
    resetCount INTEGER NOT NULL DEFAULT 0,
    logLevel TEXT NOT NULL DEFAULT 'INFO',
    logType TEXT NOT NULL DEFAULT 'ACCESS',
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, traceId)
);

CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_time_instance ON HUB_GW_ACCESS_LOG(gatewayStartProcessingTime, gatewayInstanceId);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_time_route ON HUB_GW_ACCESS_LOG(gatewayStartProcessingTime, routeConfigId);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_time_service ON HUB_GW_ACCESS_LOG(gatewayStartProcessingTime, serviceDefinitionId);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_instance_name ON HUB_GW_ACCESS_LOG(gatewayInstanceName, gatewayStartProcessingTime);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_route_name ON HUB_GW_ACCESS_LOG(routeName, gatewayStartProcessingTime);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_service_name ON HUB_GW_ACCESS_LOG(serviceName, gatewayStartProcessingTime);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_client_ip ON HUB_GW_ACCESS_LOG(clientIpAddress, gatewayStartProcessingTime);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_status_time ON HUB_GW_ACCESS_LOG(gatewayStatusCode, gatewayStartProcessingTime);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_ACCESS_LOG_proxy_type ON HUB_GW_ACCESS_LOG(proxyType, gatewayStartProcessingTime);

-- 21. 后端追踪日志表 - HUB_GW_BACKEND_TRACE_LOG
-- 对应结构：internal/gateway/logwrite/types/backend_trace_log.go (BackendTraceLog)
-- 说明：
--   1. 作为 HUB_GW_ACCESS_LOG 的从表，记录每个后端服务转发的详细信息
--   2. traceId + backendTraceId 作为联合主键
CREATE TABLE IF NOT EXISTS HUB_GW_BACKEND_TRACE_LOG (
    tenantId TEXT NOT NULL,                 -- 租户ID
    traceId TEXT NOT NULL,                  -- 链路追踪ID，关联主表 HUB_GW_ACCESS_LOG.traceId
    backendTraceId TEXT NOT NULL,           -- 后端服务追踪ID，同一traceId下唯一

    -- 服务信息（单个后端服务一次转发一条记录）
    serviceDefinitionId TEXT,               -- 服务定义ID
    serviceName TEXT,                       -- 服务名称（冗余字段）

    -- 转发信息
    forwardAddress TEXT,                    -- 实际转发目标地址(完整URL)
    forwardMethod TEXT,                     -- 转发HTTP方法
    forwardPath TEXT,                       -- 转发路径
    forwardQuery TEXT,                      -- 转发查询参数
    forwardHeaders TEXT,                    -- 转发请求头(JSON格式)
    forwardBody TEXT,                       -- 转发请求体
    requestSize INTEGER DEFAULT 0,          -- 请求大小(字节)

    -- 负载均衡信息
    loadBalancerStrategy TEXT,             -- 负载均衡策略
    loadBalancerDecision TEXT,             -- 负载均衡决策信息

    -- 时间信息
    requestStartTime DATETIME NOT NULL,    -- 向后端发起请求时间
    responseReceivedTime DATETIME,         -- 接收到后端响应时间
    requestDurationMs INTEGER,             -- 请求耗时(毫秒)

    -- 响应信息
    statusCode INTEGER,                    -- 后端HTTP状态码
    responseSize INTEGER DEFAULT 0,        -- 响应大小(字节)
    responseHeaders TEXT,                  -- 响应头信息(JSON格式)
    responseBody TEXT,                     -- 响应体内容

    -- 错误信息
    errorCode TEXT,                        -- 错误代码
    errorMessage TEXT,                     -- 详细错误信息

    -- 状态信息
    successFlag TEXT NOT NULL DEFAULT 'N', -- 是否成功(Y成功,N失败)
    traceStatus TEXT NOT NULL DEFAULT 'pending',-- 后端调用状态(pending,success,failed,timeout)
    retryCount INTEGER NOT NULL DEFAULT 0, -- 重试次数

    -- 扩展信息
    extProperty TEXT,                      -- 扩展属性(JSON格式)

    -- 标准数据库字段
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,

    PRIMARY KEY (traceId, backendTraceId)
);

-- 索引设计：参考数据库规范，兼顾多租户与常用查询场景
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_TRACE ON HUB_GW_BACKEND_TRACE_LOG(tenantId, traceId);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_SERVICE ON HUB_GW_BACKEND_TRACE_LOG(tenantId, serviceDefinitionId, requestStartTime);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_TIME ON HUB_GW_BACKEND_TRACE_LOG(requestStartTime);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_TSTATUS ON HUB_GW_BACKEND_TRACE_LOG(tenantId, traceStatus, requestStartTime);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_ADDTIME ON HUB_GW_BACKEND_TRACE_LOG(tenantId, addTime);

-- 22. 安全配置表
CREATE TABLE IF NOT EXISTS HUB_GW_SECURITY_CONFIG (
    tenantId TEXT NOT NULL,
    securityConfigId TEXT NOT NULL,
    gatewayInstanceId TEXT,
    routeConfigId TEXT,
    configName TEXT NOT NULL,
    configDesc TEXT,
    configPriority INTEGER NOT NULL DEFAULT 0,
    customConfigJson TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, securityConfigId)
);

CREATE INDEX IF NOT EXISTS idx_HUB_GW_SECURITY_CONFIG_instance ON HUB_GW_SECURITY_CONFIG(gatewayInstanceId);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_SECURITY_CONFIG_route ON HUB_GW_SECURITY_CONFIG(routeConfigId);
CREATE INDEX IF NOT EXISTS idx_HUB_GW_SECURITY_CONFIG_priority ON HUB_GW_SECURITY_CONFIG(configPriority);

-- 23. IP访问控制配置表
CREATE TABLE IF NOT EXISTS HUB_GW_IP_ACCESS_CONFIG (
    tenantId TEXT NOT NULL,
    ipAccessConfigId TEXT NOT NULL,
    securityConfigId TEXT NOT NULL,
    configName TEXT NOT NULL,
    defaultPolicy TEXT NOT NULL DEFAULT 'allow',
    whitelistIps TEXT,
    blacklistIps TEXT,
    whitelistCidrs TEXT,
    blacklistCidrs TEXT,
    trustXForwardedFor TEXT NOT NULL DEFAULT 'Y',
    trustXRealIp TEXT NOT NULL DEFAULT 'Y',
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, ipAccessConfigId)
);

CREATE INDEX IF NOT EXISTS idx_HUB_GW_IP_ACCESS_CONFIG_security ON HUB_GW_IP_ACCESS_CONFIG(securityConfigId);

-- 24. User-Agent访问控制配置表
CREATE TABLE IF NOT EXISTS HUB_GW_UA_ACCESS_CONFIG (
    tenantId TEXT NOT NULL,
    useragentAccessConfigId TEXT NOT NULL,
    securityConfigId TEXT NOT NULL,
    configName TEXT NOT NULL,
    defaultPolicy TEXT NOT NULL DEFAULT 'allow',
    whitelistPatterns TEXT,
    blacklistPatterns TEXT,
    blockEmptyUserAgent TEXT NOT NULL DEFAULT 'N',
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, useragentAccessConfigId)
);

CREATE INDEX IF NOT EXISTS idx_HUB_GW_UA_ACCESS_CONFIG_security ON HUB_GW_UA_ACCESS_CONFIG(securityConfigId);

-- 25. API访问控制配置表
CREATE TABLE IF NOT EXISTS HUB_GW_API_ACCESS_CONFIG (
    tenantId TEXT NOT NULL,
    apiAccessConfigId TEXT NOT NULL,
    securityConfigId TEXT NOT NULL,
    configName TEXT NOT NULL,
    defaultPolicy TEXT NOT NULL DEFAULT 'allow',
    whitelistPaths TEXT,
    blacklistPaths TEXT,
    allowedMethods TEXT DEFAULT 'GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS',
    blockedMethods TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, apiAccessConfigId)
);

CREATE INDEX IF NOT EXISTS idx_HUB_GW_API_ACCESS_CONFIG_security ON HUB_GW_API_ACCESS_CONFIG(securityConfigId);

-- 26. 域名访问控制配置表
CREATE TABLE IF NOT EXISTS HUB_GW_DOMAIN_ACCESS_CONFIG (
    tenantId TEXT NOT NULL,
    domainAccessConfigId TEXT NOT NULL,
    securityConfigId TEXT NOT NULL,
    configName TEXT NOT NULL,
    defaultPolicy TEXT NOT NULL DEFAULT 'allow',
    whitelistDomains TEXT,
    blacklistDomains TEXT,
    allowSubdomains TEXT NOT NULL DEFAULT 'Y',
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 INTEGER,
    reserved4 INTEGER,
    reserved5 DATETIME,
    extProperty TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    PRIMARY KEY (tenantId, domainAccessConfigId)
);

CREATE INDEX IF NOT EXISTS idx_HUB_GW_DOMAIN_ACCESS_CONFIG_security ON HUB_GW_DOMAIN_ACCESS_CONFIG(securityConfigId);

-- 27. 服务器信息主表
CREATE TABLE IF NOT EXISTS HUB_METRIC_SERVER_INFO (
    metricServerId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    hostname TEXT NOT NULL,
    osType TEXT NOT NULL,
    osVersion TEXT NOT NULL,
    kernelVersion TEXT,
    architecture TEXT NOT NULL,
    bootTime DATETIME NOT NULL,
    ipAddress TEXT,
    macAddress TEXT,
    serverLocation TEXT,
    serverType TEXT,
    lastUpdateTime DATETIME NOT NULL,
    networkInfo TEXT,
    systemInfo TEXT,
    hardwareInfo TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricServerId)
);

CREATE INDEX IDX_METRIC_SERVER_HOST ON HUB_METRIC_SERVER_INFO(hostname);
CREATE INDEX IDX_METRIC_SERVER_OS ON HUB_METRIC_SERVER_INFO(osType);
CREATE INDEX IDX_METRIC_SERVER_IP ON HUB_METRIC_SERVER_INFO(ipAddress);
CREATE INDEX IDX_METRIC_SERVER_TYPE ON HUB_METRIC_SERVER_INFO(serverType);
CREATE INDEX IDX_METRIC_SERVER_ACTIVE ON HUB_METRIC_SERVER_INFO(activeFlag);
CREATE INDEX IDX_METRIC_SERVER_UPDATE ON HUB_METRIC_SERVER_INFO(lastUpdateTime);

-- 28. CPU采集日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_CPU_LOG (
    metricCpuLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    usagePercent REAL NOT NULL DEFAULT 0.00,
    userPercent REAL NOT NULL DEFAULT 0.00,
    systemPercent REAL NOT NULL DEFAULT 0.00,
    idlePercent REAL NOT NULL DEFAULT 0.00,
    ioWaitPercent REAL NOT NULL DEFAULT 0.00,
    irqPercent REAL NOT NULL DEFAULT 0.00,
    softIrqPercent REAL NOT NULL DEFAULT 0.00,
    coreCount INTEGER NOT NULL DEFAULT 0,
    logicalCount INTEGER NOT NULL DEFAULT 0,
    loadAvg1 REAL NOT NULL DEFAULT 0.00,
    loadAvg5 REAL NOT NULL DEFAULT 0.00,
    loadAvg15 REAL NOT NULL DEFAULT 0.00,
    collectTime DATETIME NOT NULL,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricCpuLogId)
);

CREATE INDEX IDX_METRIC_CPU_SERVER ON HUB_METRIC_CPU_LOG(metricServerId);
CREATE INDEX IDX_METRIC_CPU_TIME ON HUB_METRIC_CPU_LOG(collectTime);
CREATE INDEX IDX_METRIC_CPU_USAGE ON HUB_METRIC_CPU_LOG(usagePercent);
CREATE INDEX IDX_METRIC_CPU_ACTIVE ON HUB_METRIC_CPU_LOG(activeFlag);
CREATE INDEX IDX_METRIC_CPU_SRV_TIME ON HUB_METRIC_CPU_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_CPU_TNT_TIME ON HUB_METRIC_CPU_LOG(tenantId, collectTime);

-- 29. 内存采集日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_MEMORY_LOG (
    metricMemoryLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    totalMemory INTEGER NOT NULL DEFAULT 0,
    availableMemory INTEGER NOT NULL DEFAULT 0,
    usedMemory INTEGER NOT NULL DEFAULT 0,
    usagePercent REAL NOT NULL DEFAULT 0.00,
    freeMemory INTEGER NOT NULL DEFAULT 0,
    cachedMemory INTEGER NOT NULL DEFAULT 0,
    buffersMemory INTEGER NOT NULL DEFAULT 0,
    sharedMemory INTEGER NOT NULL DEFAULT 0,
    swapTotal INTEGER NOT NULL DEFAULT 0,
    swapUsed INTEGER NOT NULL DEFAULT 0,
    swapFree INTEGER NOT NULL DEFAULT 0,
    swapUsagePercent REAL NOT NULL DEFAULT 0.00,
    collectTime DATETIME NOT NULL,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricMemoryLogId)
);

CREATE INDEX IDX_METRIC_MEMORY_SERVER ON HUB_METRIC_MEMORY_LOG(metricServerId);
CREATE INDEX IDX_METRIC_MEMORY_TIME ON HUB_METRIC_MEMORY_LOG(collectTime);
CREATE INDEX IDX_METRIC_MEMORY_USAGE ON HUB_METRIC_MEMORY_LOG(usagePercent);
CREATE INDEX IDX_METRIC_MEMORY_ACTIVE ON HUB_METRIC_MEMORY_LOG(activeFlag);
CREATE INDEX IDX_METRIC_MEMORY_SRV_TIME ON HUB_METRIC_MEMORY_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_MEMORY_TNT_TIME ON HUB_METRIC_MEMORY_LOG(tenantId, collectTime);

-- 30. 磁盘分区日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_DISK_PART_LOG (
    metricDiskPartitionLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    deviceName TEXT NOT NULL,
    mountPoint TEXT NOT NULL,
    fileSystem TEXT NOT NULL,
    totalSpace INTEGER NOT NULL DEFAULT 0,
    usedSpace INTEGER NOT NULL DEFAULT 0,
    freeSpace INTEGER NOT NULL DEFAULT 0,
    usagePercent REAL NOT NULL DEFAULT 0.00,
    inodesTotal INTEGER NOT NULL DEFAULT 0,
    inodesUsed INTEGER NOT NULL DEFAULT 0,
    inodesFree INTEGER NOT NULL DEFAULT 0,
    inodesUsagePercent REAL NOT NULL DEFAULT 0.00,
    collectTime DATETIME NOT NULL,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricDiskPartitionLogId)
);

CREATE INDEX IDX_METRIC_DISK_PART_SERVER ON HUB_METRIC_DISK_PART_LOG(metricServerId);
CREATE INDEX IDX_METRIC_DISK_PART_TIME ON HUB_METRIC_DISK_PART_LOG(collectTime);
CREATE INDEX IDX_METRIC_DISK_PART_DEVICE ON HUB_METRIC_DISK_PART_LOG(deviceName);
CREATE INDEX IDX_METRIC_DISK_PART_USAGE ON HUB_METRIC_DISK_PART_LOG(usagePercent);
CREATE INDEX IDX_METRIC_DISK_PART_ACTIVE ON HUB_METRIC_DISK_PART_LOG(activeFlag);
CREATE INDEX IDX_METRIC_DISK_PART_SRV_TIME ON HUB_METRIC_DISK_PART_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_DISK_PART_SRV_DEV ON HUB_METRIC_DISK_PART_LOG(metricServerId, deviceName);
CREATE INDEX IDX_METRIC_DISK_PART_TNT_TIME ON HUB_METRIC_DISK_PART_LOG(tenantId, collectTime);

-- 31. 磁盘IO日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_DISK_IO_LOG (
    metricDiskIoLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    deviceName TEXT NOT NULL,
    readCount INTEGER NOT NULL DEFAULT 0,
    writeCount INTEGER NOT NULL DEFAULT 0,
    readBytes INTEGER NOT NULL DEFAULT 0,
    writeBytes INTEGER NOT NULL DEFAULT 0,
    readTime INTEGER NOT NULL DEFAULT 0,
    writeTime INTEGER NOT NULL DEFAULT 0,
    ioInProgress INTEGER NOT NULL DEFAULT 0,
    ioTime INTEGER NOT NULL DEFAULT 0,
    readRate REAL NOT NULL DEFAULT 0.00,
    writeRate REAL NOT NULL DEFAULT 0.00,
    collectTime DATETIME NOT NULL,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricDiskIoLogId)
);

CREATE INDEX IDX_METRIC_DISK_IO_SERVER ON HUB_METRIC_DISK_IO_LOG(metricServerId);
CREATE INDEX IDX_METRIC_DISK_IO_TIME ON HUB_METRIC_DISK_IO_LOG(collectTime);
CREATE INDEX IDX_METRIC_DISK_IO_DEVICE ON HUB_METRIC_DISK_IO_LOG(deviceName);
CREATE INDEX IDX_METRIC_DISK_IO_ACTIVE ON HUB_METRIC_DISK_IO_LOG(activeFlag);
CREATE INDEX IDX_METRIC_DISK_IO_SRV_TIME ON HUB_METRIC_DISK_IO_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_DISK_IO_SRV_DEV ON HUB_METRIC_DISK_IO_LOG(metricServerId, deviceName);
CREATE INDEX IDX_METRIC_DISK_IO_TNT_TIME ON HUB_METRIC_DISK_IO_LOG(tenantId, collectTime);

-- 32. 网络接口日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_NETWORK_LOG (
    metricNetworkLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    interfaceName TEXT NOT NULL,
    hardwareAddr TEXT,
    ipAddresses TEXT,
    interfaceStatus TEXT NOT NULL,
    interfaceType TEXT,
    bytesReceived INTEGER NOT NULL DEFAULT 0,
    bytesSent INTEGER NOT NULL DEFAULT 0,
    packetsReceived INTEGER NOT NULL DEFAULT 0,
    packetsSent INTEGER NOT NULL DEFAULT 0,
    errorsReceived INTEGER NOT NULL DEFAULT 0,
    errorsSent INTEGER NOT NULL DEFAULT 0,
    droppedReceived INTEGER NOT NULL DEFAULT 0,
    droppedSent INTEGER NOT NULL DEFAULT 0,
    receiveRate REAL DEFAULT 0,
    sendRate REAL DEFAULT 0,
    collectTime DATETIME NOT NULL,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricNetworkLogId)
);

CREATE INDEX IDX_METRIC_NETWORK_SERVER ON HUB_METRIC_NETWORK_LOG(metricServerId);
CREATE INDEX IDX_METRIC_NETWORK_TIME ON HUB_METRIC_NETWORK_LOG(collectTime);
CREATE INDEX IDX_METRIC_NETWORK_INTERFACE ON HUB_METRIC_NETWORK_LOG(interfaceName);
CREATE INDEX IDX_METRIC_NETWORK_STATUS ON HUB_METRIC_NETWORK_LOG(interfaceStatus);
CREATE INDEX IDX_METRIC_NETWORK_ACTIVE ON HUB_METRIC_NETWORK_LOG(activeFlag);
CREATE INDEX IDX_METRIC_NETWORK_SRV_TIME ON HUB_METRIC_NETWORK_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_NETWORK_SRV_INT ON HUB_METRIC_NETWORK_LOG(metricServerId, interfaceName);
CREATE INDEX IDX_METRIC_NETWORK_TNT_TIME ON HUB_METRIC_NETWORK_LOG(tenantId, collectTime);

-- 33. 进程信息日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_PROCESS_LOG (
    metricProcessLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    processId INTEGER NOT NULL,
    parentProcessId INTEGER,
    processName TEXT NOT NULL,
    processStatus TEXT NOT NULL,
    createTime DATETIME NOT NULL,
    runTime INTEGER NOT NULL DEFAULT 0,
    memoryUsage INTEGER NOT NULL DEFAULT 0,
    memoryPercent REAL NOT NULL DEFAULT 0.00,
    cpuPercent REAL NOT NULL DEFAULT 0.00,
    threadCount INTEGER NOT NULL DEFAULT 0,
    fileDescriptorCount INTEGER NOT NULL DEFAULT 0,
    commandLine TEXT,
    executablePath TEXT,
    workingDirectory TEXT,
    collectTime DATETIME NOT NULL,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricProcessLogId)
);

CREATE INDEX IDX_METRIC_PROCESS_SERVER ON HUB_METRIC_PROCESS_LOG(metricServerId);
CREATE INDEX IDX_METRIC_PROCESS_TIME ON HUB_METRIC_PROCESS_LOG(collectTime);
CREATE INDEX IDX_METRIC_PROCESS_PID ON HUB_METRIC_PROCESS_LOG(processId);
CREATE INDEX IDX_METRIC_PROCESS_NAME ON HUB_METRIC_PROCESS_LOG(processName);
CREATE INDEX IDX_METRIC_PROCESS_STATUS ON HUB_METRIC_PROCESS_LOG(processStatus);
CREATE INDEX IDX_METRIC_PROCESS_ACTIVE ON HUB_METRIC_PROCESS_LOG(activeFlag);
CREATE INDEX IDX_METRIC_PROCESS_SRV_TIME ON HUB_METRIC_PROCESS_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_PROCESS_SRV_PID ON HUB_METRIC_PROCESS_LOG(metricServerId, processId);
CREATE INDEX IDX_METRIC_PROCESS_TNT_TIME ON HUB_METRIC_PROCESS_LOG(tenantId, collectTime);

-- 34. 进程统计日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_PROCSTAT_LOG (
    metricProcessStatsLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    runningCount INTEGER NOT NULL DEFAULT 0,
    sleepingCount INTEGER NOT NULL DEFAULT 0,
    stoppedCount INTEGER NOT NULL DEFAULT 0,
    zombieCount INTEGER NOT NULL DEFAULT 0,
    totalCount INTEGER NOT NULL DEFAULT 0,
    collectTime DATETIME NOT NULL,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricProcessStatsLogId)
);

CREATE INDEX IDX_METRIC_PROC_STATS_SERVER ON HUB_METRIC_PROCSTAT_LOG(metricServerId);
CREATE INDEX IDX_METRIC_PROC_STATS_TIME ON HUB_METRIC_PROCSTAT_LOG(collectTime);
CREATE INDEX IDX_METRIC_PROC_STATS_ACTIVE ON HUB_METRIC_PROCSTAT_LOG(activeFlag);
CREATE INDEX IDX_METRIC_PROC_STATS_SRV_TIME ON HUB_METRIC_PROCSTAT_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_PROC_STATS_TNT_TIME ON HUB_METRIC_PROCSTAT_LOG(tenantId, collectTime);

-- 35. 温度信息日志表
CREATE TABLE IF NOT EXISTS HUB_METRIC_TEMP_LOG (
    metricTemperatureLogId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    metricServerId TEXT NOT NULL,
    sensorName TEXT NOT NULL,
    temperatureValue REAL NOT NULL DEFAULT 0.00,
    highThreshold REAL,
    criticalThreshold REAL,
    collectTime DATETIME NOT NULL,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, metricTemperatureLogId)
);

CREATE INDEX IDX_METRIC_TEMP_SERVER ON HUB_METRIC_TEMP_LOG(metricServerId);
CREATE INDEX IDX_METRIC_TEMP_TIME ON HUB_METRIC_TEMP_LOG(collectTime);
CREATE INDEX IDX_METRIC_TEMP_SENSOR ON HUB_METRIC_TEMP_LOG(sensorName);
CREATE INDEX IDX_METRIC_TEMP_ACTIVE ON HUB_METRIC_TEMP_LOG(activeFlag);
CREATE INDEX IDX_METRIC_TEMP_SRV_TIME ON HUB_METRIC_TEMP_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_TEMP_SRV_SENSOR ON HUB_METRIC_TEMP_LOG(metricServerId, sensorName);
CREATE INDEX IDX_METRIC_TEMP_TNT_TIME ON HUB_METRIC_TEMP_LOG(tenantId, collectTime);

-- =====================================================
-- 服务注册中心相关表结构
-- 基于 service_registry.sql 转换为 SQLite 格式
-- =====================================================

-- 服务分组表 - 存储服务分组和授权信息
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_SERVICE_GROUP (
  -- 主键和租户信息
  serviceGroupId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 分组基本信息
  groupName TEXT NOT NULL,
  groupDescription TEXT,
  groupType TEXT DEFAULT 'BUSINESS',
  
  -- 授权信息
  ownerUserId TEXT NOT NULL,
  adminUserIds TEXT,
  readUserIds TEXT,
  accessControlEnabled TEXT DEFAULT 'N',
  
  -- 配置信息
  defaultProtocolType TEXT DEFAULT 'HTTP',
  defaultLoadBalanceStrategy TEXT DEFAULT 'ROUND_ROBIN',
  defaultHealthCheckUrl TEXT DEFAULT '/health',
  defaultHealthCheckIntervalSeconds INTEGER DEFAULT 30,
  
  -- 通用字段
  addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho TEXT NOT NULL,
  editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, serviceGroupId)
);

-- 服务表 - 存储服务基本信息
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_SERVICE (
  -- 主键和租户信息
  tenantId TEXT NOT NULL,
  serviceName TEXT NOT NULL,
  
  -- 关联分组（主键关联）
  serviceGroupId TEXT NOT NULL,
  -- 冗余字段（便于查询和展示）
  groupName TEXT NOT NULL,
  
  -- 服务基本信息
  serviceDescription TEXT,
  
  -- 注册管理配置
  registryType TEXT NOT NULL DEFAULT 'INTERNAL',
  externalRegistryConfig TEXT,
  
  -- 服务配置
  protocolType TEXT DEFAULT 'HTTP',
  contextPath TEXT DEFAULT '',
  loadBalanceStrategy TEXT DEFAULT 'ROUND_ROBIN',
  
  -- 健康检查配置
  healthCheckUrl TEXT DEFAULT '/health',
  healthCheckIntervalSeconds INTEGER DEFAULT 30,
  healthCheckTimeoutSeconds INTEGER DEFAULT 5,
  healthCheckType TEXT DEFAULT 'HTTP',
  healthCheckMode TEXT DEFAULT 'ACTIVE',
  
  -- 元数据和标签
  metadataJson TEXT,
  tagsJson TEXT,
  
  -- 通用字段
  addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho TEXT NOT NULL,
  editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, serviceName)
);

-- 服务实例表 - 存储具体的服务实例
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_SERVICE_INSTANCE (
  -- 主键和租户信息
  serviceInstanceId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联服务和分组（主键关联）
  serviceGroupId TEXT NOT NULL,
  -- 冗余字段（便于查询和展示）
  serviceName TEXT NOT NULL,
  groupName TEXT NOT NULL,
  
  -- 网络连接信息
  hostAddress TEXT NOT NULL,
  portNumber INTEGER NOT NULL,
  contextPath TEXT DEFAULT '',
  
  -- 实例状态信息
  instanceStatus TEXT NOT NULL DEFAULT 'UP',
  healthStatus TEXT NOT NULL DEFAULT 'UNKNOWN',
  
  -- 负载均衡配置
  weightValue INTEGER NOT NULL DEFAULT 100,
  
  -- 客户端信息
  clientId TEXT,
  clientVersion TEXT,
  clientType TEXT DEFAULT 'SERVICE',
  tempInstanceFlag TEXT NOT NULL DEFAULT 'N',
  
  -- 健康检查统计
  heartbeatFailCount INTEGER NOT NULL DEFAULT 0,
  
  -- 元数据和标签
  metadataJson TEXT,
  tagsJson TEXT,
  
  -- 时间戳信息
  registerTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  lastHeartbeatTime DATETIME,
  lastHealthCheckTime DATETIME,
  
  -- 通用字段
  addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho TEXT NOT NULL,
  editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, serviceInstanceId)
);

-- 服务事件日志表 - 记录服务变更事件
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_SERVICE_EVENT (
  -- 主键和租户信息
  serviceEventId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联主键字段（用于精确关联到对应表记录）
  serviceGroupId TEXT,
  serviceInstanceId TEXT,
  
  -- 事件基本信息（冗余字段，便于查询和展示）
  groupName TEXT,
  serviceName TEXT,
  hostAddress TEXT,
  portNumber INTEGER,
  nodeIpAddress TEXT,
  eventType TEXT NOT NULL,
  eventSource TEXT,
  
  -- 事件数据
  eventDataJson TEXT,
  eventMessage TEXT,
  
  -- 时间信息
  eventTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  -- 通用字段
  addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho TEXT NOT NULL,
  editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, serviceEventId)
);

-- =====================================================
-- 数据库表结构设计说明
-- =====================================================
-- 
-- 注册类型说明：
-- 1. INTERNAL: 内部管理（默认）- 服务实例直接注册到本系统数据库
-- 2. NACOS: Nacos注册中心 - 服务实例注册到Nacos，本系统作为代理
-- 3. CONSUL: Consul注册中心 - 服务实例注册到Consul，本系统作为代理
-- 4. EUREKA: Eureka注册中心 - 服务实例注册到Eureka，本系统作为代理
-- 5. ETCD: ETCD注册中心 - 服务实例注册到ETCD，本系统作为代理
-- 6. ZOOKEEPER: ZooKeeper注册中心 - 服务实例注册到ZooKeeper，本系统作为代理
--
-- 外部注册中心配置格式（externalRegistryConfig字段JSON示例）：
-- {
--   "serverAddress": "192.168.0.120:8848",
--   "namespace": "ea63c755-3d65-4203-87d7-5ee6837f5bc9",
--   "groupName": "datahub-test-group",
--   "username": "nacos",
--   "password": "nacos",
--   "timeout": 10000,
--   "enableAuth": true,
--   "connectionPool": {
--     "maxConnections": 10,
--     "connectionTimeout": 5000
--   }
-- }
--
-- 使用场景：
-- - registryType = 'INTERNAL': 传统的服务注册，实例信息存储在本地数据库
-- - registryType = 'NACOS': 服务作为Nacos和第三方应用的代理，提供统一的服务发现接口
-- - 其他类型: 类似Nacos，作为对应注册中心的代理
-- =====================================================

-- =====================================================
-- 服务注册中心相关索引
-- =====================================================

-- HUB_REGISTRY_SERVICE_GROUP 索引
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_GROUP_NAME ON HUB_REGISTRY_SERVICE_GROUP(tenantId, groupName);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_GROUP_TYPE ON HUB_REGISTRY_SERVICE_GROUP(groupType);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_GROUP_OWNER ON HUB_REGISTRY_SERVICE_GROUP(ownerUserId);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_GROUP_ACTIVE ON HUB_REGISTRY_SERVICE_GROUP(activeFlag);

-- HUB_REGISTRY_SERVICE 索引
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_SVC_GROUP_ID ON HUB_REGISTRY_SERVICE(tenantId, serviceGroupId);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_SVC_GROUP_NAME ON HUB_REGISTRY_SERVICE(groupName);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_SVC_REGISTRY_TYPE ON HUB_REGISTRY_SERVICE(registryType);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_SVC_ACTIVE ON HUB_REGISTRY_SERVICE(activeFlag);

-- HUB_REGISTRY_SERVICE_INSTANCE 索引
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INSTANCE ON HUB_REGISTRY_SERVICE_INSTANCE(tenantId, serviceGroupId, serviceName, hostAddress, portNumber);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_GROUP_ID ON HUB_REGISTRY_SERVICE_INSTANCE(tenantId, serviceGroupId);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_SVC_NAME ON HUB_REGISTRY_SERVICE_INSTANCE(serviceName);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_GROUP_NAME ON HUB_REGISTRY_SERVICE_INSTANCE(groupName);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_STATUS ON HUB_REGISTRY_SERVICE_INSTANCE(instanceStatus);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_HEALTH ON HUB_REGISTRY_SERVICE_INSTANCE(healthStatus);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_HEARTBEAT ON HUB_REGISTRY_SERVICE_INSTANCE(lastHeartbeatTime);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_HOST_PORT ON HUB_REGISTRY_SERVICE_INSTANCE(hostAddress, portNumber);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_CLIENT ON HUB_REGISTRY_SERVICE_INSTANCE(clientId);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_ACTIVE ON HUB_REGISTRY_SERVICE_INSTANCE(activeFlag);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_INST_TEMP ON HUB_REGISTRY_SERVICE_INSTANCE(tempInstanceFlag);

-- HUB_REGISTRY_SERVICE_EVENT 索引
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EVENT_GROUP_ID ON HUB_REGISTRY_SERVICE_EVENT(tenantId, serviceGroupId, eventTime);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EVENT_INSTANCE_ID ON HUB_REGISTRY_SERVICE_EVENT(tenantId, serviceInstanceId, eventTime);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EVENT_GROUP_NAME ON HUB_REGISTRY_SERVICE_EVENT(tenantId, groupName, eventTime);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EVENT_SVC_NAME ON HUB_REGISTRY_SERVICE_EVENT(tenantId, serviceName, eventTime);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EVENT_HOST ON HUB_REGISTRY_SERVICE_EVENT(tenantId, hostAddress, portNumber, eventTime);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EVENT_NODE_IP ON HUB_REGISTRY_SERVICE_EVENT(tenantId, nodeIpAddress, eventTime);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EVENT_TYPE ON HUB_REGISTRY_SERVICE_EVENT(eventType, eventTime);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EVENT_TIME ON HUB_REGISTRY_SERVICE_EVENT(eventTime);


-- ==========================================
-- 1. JVM资源信息主表
-- 存储JVM整体资源监控信息的快照数据
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_RESOURCE (
    jvmResourceId TEXT NOT NULL, -- JVM资源记录ID（由应用端生成的唯一标识），主键
    tenantId TEXT NOT NULL, -- 租户ID
    serviceGroupId TEXT NOT NULL, -- 服务分组ID，主键
    
    -- 应用标识信息
    applicationName TEXT NOT NULL, -- 应用名称
    groupName TEXT NOT NULL, -- 分组名称
    hostName TEXT DEFAULT NULL, -- 主机名
    hostIpAddress TEXT DEFAULT NULL, -- 主机IP地址
    
    -- 时间相关字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    jvmStartTime TEXT NOT NULL, -- JVM启动时间
    jvmUptimeMs INTEGER DEFAULT 0 NOT NULL, -- JVM运行时长（毫秒）
    
    -- 健康状态字段
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- JVM整体健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- JVM健康等级(EXCELLENT/GOOD/FAIR/POOR)
    requiresAttentionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresAttentionFlag IN ('Y','N')), -- 是否需要立即关注(Y是,N否)
    summaryText TEXT DEFAULT NULL, -- 监控摘要信息
    
    -- 系统属性（JSON格式）
    systemPropertiesJson TEXT DEFAULT NULL, -- JVM系统属性，JSON格式（可能包含大量系统属性）
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, serviceGroupId, jvmResourceId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_APP ON HUB_MONITOR_JVM_RESOURCE(applicationName);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_TIME ON HUB_MONITOR_JVM_RESOURCE(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_HEALTH ON HUB_MONITOR_JVM_RESOURCE(healthyFlag, requiresAttentionFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_HOST ON HUB_MONITOR_JVM_RESOURCE(hostIpAddress);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_GROUP ON HUB_MONITOR_JVM_RESOURCE(serviceGroupId, groupName);

-- ==========================================
-- 2. 内存信息表（堆内存和非堆内存）
-- 存储JVM堆内存和非堆内存的使用情况
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_MEMORY (
    jvmMemoryId TEXT NOT NULL, -- JVM内存记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 内存类型
    memoryType TEXT NOT NULL, -- 内存类型(HEAP/NON_HEAP)
    
    -- 内存使用情况（字节）
    initMemoryBytes INTEGER DEFAULT 0 NOT NULL, -- 初始内存大小（字节）
    usedMemoryBytes INTEGER DEFAULT 0 NOT NULL, -- 已使用内存大小（字节）
    committedMemoryBytes INTEGER DEFAULT 0 NOT NULL, -- 已提交内存大小（字节）
    maxMemoryBytes INTEGER DEFAULT -1 NOT NULL, -- 最大内存大小（字节），-1表示无限制
    
    -- 计算指标
    usagePercent REAL DEFAULT 0.00 NOT NULL, -- 内存使用率（百分比）
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 内存健康标记(Y健康,N异常)
    
    -- 时间字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, jvmMemoryId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_MEM_RES ON HUB_MONITOR_JVM_MEMORY(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_MEM_TYPE ON HUB_MONITOR_JVM_MEMORY(memoryType);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_MEM_TIME ON HUB_MONITOR_JVM_MEMORY(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_MEM_USAGE ON HUB_MONITOR_JVM_MEMORY(usagePercent);

-- ==========================================
-- 3. 内存池信息表
-- 存储具体内存池的详细使用情况（Eden、Survivor、Old Gen、Metaspace等）
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_MEM_POOL (
    memoryPoolId TEXT NOT NULL, -- 内存池记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 内存池基本信息
    poolName TEXT NOT NULL, -- 内存池名称
    poolType TEXT NOT NULL, -- 内存池类型(HEAP/NON_HEAP)
    poolCategory TEXT DEFAULT NULL, -- 内存池分类（年轻代/老年代/元数据空间/代码缓存/其他）
    
    -- 当前使用情况
    currentInitBytes INTEGER DEFAULT 0 NOT NULL, -- 当前初始内存（字节）
    currentUsedBytes INTEGER DEFAULT 0 NOT NULL, -- 当前已使用内存（字节）
    currentCommittedBytes INTEGER DEFAULT 0 NOT NULL, -- 当前已提交内存（字节）
    currentMaxBytes INTEGER DEFAULT -1 NOT NULL, -- 当前最大内存（字节）
    currentUsagePercent REAL DEFAULT 0.00 NOT NULL, -- 当前使用率（百分比）
    
    -- 峰值使用情况
    peakInitBytes INTEGER DEFAULT 0, -- 峰值初始内存（字节）
    peakUsedBytes INTEGER DEFAULT 0, -- 峰值已使用内存（字节）
    peakCommittedBytes INTEGER DEFAULT 0, -- 峰值已提交内存（字节）
    peakMaxBytes INTEGER DEFAULT -1, -- 峰值最大内存（字节）
    peakUsagePercent REAL DEFAULT 0.00, -- 峰值使用率（百分比）
    
    -- 阈值监控
    usageThresholdSupported TEXT DEFAULT 'N' NOT NULL CHECK(usageThresholdSupported IN ('Y','N')), -- 是否支持使用阈值监控(Y是,N否)
    usageThresholdBytes INTEGER DEFAULT 0, -- 使用阈值（字节）
    usageThresholdCount INTEGER DEFAULT 0, -- 使用阈值超越次数
    collectionUsageSupported TEXT DEFAULT 'N' NOT NULL CHECK(collectionUsageSupported IN ('Y','N')), -- 是否支持收集使用量监控(Y是,N否)
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 内存池健康标记(Y健康,N异常)
    
    -- 时间字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, memoryPoolId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_RES ON HUB_MONITOR_JVM_MEM_POOL(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_NAME ON HUB_MONITOR_JVM_MEM_POOL(poolName);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_TYPE ON HUB_MONITOR_JVM_MEM_POOL(poolType);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_CAT ON HUB_MONITOR_JVM_MEM_POOL(poolCategory);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_TIME ON HUB_MONITOR_JVM_MEM_POOL(collectionTime);

-- ==========================================
-- 4. GC快照表（jstat -gc 风格，每次采集一条汇总记录）
-- 存储每次采集时刻的GC状态快照，包含完整的内存区域数据
-- 每次采集插入一条记录，包含所有GC收集器的汇总数据
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_GC (
    gcSnapshotId TEXT NOT NULL, -- GC快照记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- GC累积统计（从JVM启动到当前采集时刻）
    collectionCount INTEGER DEFAULT 0 NOT NULL, -- GC总次数（累积，所有GC收集器汇总）
    collectionTimeMs INTEGER DEFAULT 0 NOT NULL, -- GC总耗时（毫秒，累积，所有GC收集器汇总）
    
    -- ===== jstat -gc 风格的内存区域数据（单位：KB） =====
    
    -- Survivor区
    s0c INTEGER DEFAULT 0, -- Survivor 0 区容量（KB）
    s1c INTEGER DEFAULT 0, -- Survivor 1 区容量（KB）
    s0u INTEGER DEFAULT 0, -- Survivor 0 区使用量（KB）
    s1u INTEGER DEFAULT 0, -- Survivor 1 区使用量（KB）
    
    -- Eden区
    ec INTEGER DEFAULT 0, -- Eden 区容量（KB）
    eu INTEGER DEFAULT 0, -- Eden 区使用量（KB）
    
    -- Old区
    oc INTEGER DEFAULT 0, -- Old 区容量（KB）
    ou INTEGER DEFAULT 0, -- Old 区使用量（KB）
    
    -- Metaspace
    mc INTEGER DEFAULT 0, -- Metaspace 容量（KB）
    mu INTEGER DEFAULT 0, -- Metaspace 使用量（KB）
    
    -- 压缩类空间
    ccsc INTEGER DEFAULT 0, -- 压缩类空间容量（KB）
    ccsu INTEGER DEFAULT 0, -- 压缩类空间使用量（KB）
    
    -- GC统计（jstat -gc 格式）
    ygc INTEGER DEFAULT 0, -- 年轻代GC次数
    ygct REAL DEFAULT 0.000, -- 年轻代GC总时间（秒）
    fgc INTEGER DEFAULT 0, -- Full GC次数
    fgct REAL DEFAULT 0.000, -- Full GC总时间（秒）
    gct REAL DEFAULT 0.000, -- 总GC时间（秒）
    
    -- 时间戳信息
    collectionTime TEXT NOT NULL, -- 数据采集时间戳
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, gcSnapshotId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_GC_RES ON HUB_MONITOR_JVM_GC(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_GC_TIME ON HUB_MONITOR_JVM_GC(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_GC_RES_TIME ON HUB_MONITOR_JVM_GC(jvmResourceId, collectionTime);

-- ==========================================
-- 5. 线程信息表
-- 存储JVM线程的详细监控数据
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_THREAD (
    jvmThreadId TEXT NOT NULL, -- JVM线程记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 基础线程统计
    currentThreadCount INTEGER DEFAULT 0 NOT NULL, -- 当前线程数
    daemonThreadCount INTEGER DEFAULT 0 NOT NULL, -- 守护线程数
    userThreadCount INTEGER DEFAULT 0 NOT NULL, -- 用户线程数
    peakThreadCount INTEGER DEFAULT 0 NOT NULL, -- 峰值线程数
    totalStartedThreadCount INTEGER DEFAULT 0 NOT NULL, -- 总启动线程数
    
    -- 性能指标
    threadGrowthRatePercent REAL DEFAULT 0.00, -- 线程增长率（百分比）
    daemonThreadRatioPercent REAL DEFAULT 0.00, -- 守护线程比例（百分比）
    
    -- 监控功能支持状态
    cpuTimeSupported TEXT DEFAULT 'N' NOT NULL CHECK(cpuTimeSupported IN ('Y','N')), -- CPU时间监控是否支持(Y是,N否)
    cpuTimeEnabled TEXT DEFAULT 'N' NOT NULL CHECK(cpuTimeEnabled IN ('Y','N')), -- CPU时间监控是否启用(Y是,N否)
    memoryAllocSupported TEXT DEFAULT 'N' NOT NULL CHECK(memoryAllocSupported IN ('Y','N')), -- 内存分配监控是否支持(Y是,N否)
    memoryAllocEnabled TEXT DEFAULT 'N' NOT NULL CHECK(memoryAllocEnabled IN ('Y','N')), -- 内存分配监控是否启用(Y是,N否)
    contentionSupported TEXT DEFAULT 'N' NOT NULL CHECK(contentionSupported IN ('Y','N')), -- 争用监控是否支持(Y是,N否)
    contentionEnabled TEXT DEFAULT 'N' NOT NULL CHECK(contentionEnabled IN ('Y','N')), -- 争用监控是否启用(Y是,N否)
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 线程健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- 线程健康等级(EXCELLENT/GOOD/FAIR/POOR)
    requiresAttentionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresAttentionFlag IN ('Y','N')), -- 是否需要立即关注(Y是,N否)
    potentialIssuesJson TEXT DEFAULT NULL, -- 潜在问题列表，JSON格式
    
    -- 时间字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, jvmThreadId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_THR_RES ON HUB_MONITOR_JVM_THREAD(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THR_TIME ON HUB_MONITOR_JVM_THREAD(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THR_HEALTH ON HUB_MONITOR_JVM_THREAD(healthyFlag, requiresAttentionFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THR_COUNT ON HUB_MONITOR_JVM_THREAD(currentThreadCount);

-- ==========================================
-- 6. 线程状态统计表
-- 存储不同状态下的线程数量分布
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_THR_STATE (
    threadStateId TEXT NOT NULL, -- 线程状态记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmThreadId TEXT NOT NULL, -- 关联的JVM线程记录ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 线程状态分布
    newThreadCount INTEGER DEFAULT 0 NOT NULL, -- NEW状态线程数
    runnableThreadCount INTEGER DEFAULT 0 NOT NULL, -- RUNNABLE状态线程数
    blockedThreadCount INTEGER DEFAULT 0 NOT NULL, -- BLOCKED状态线程数
    waitingThreadCount INTEGER DEFAULT 0 NOT NULL, -- WAITING状态线程数
    timedWaitingThreadCount INTEGER DEFAULT 0 NOT NULL, -- TIMED_WAITING状态线程数
    terminatedThreadCount INTEGER DEFAULT 0 NOT NULL, -- TERMINATED状态线程数
    totalThreadCount INTEGER DEFAULT 0 NOT NULL, -- 总线程数
    
    -- 比例指标
    activeThreadRatioPercent REAL DEFAULT 0.00, -- 活跃线程比例（百分比）
    blockedThreadRatioPercent REAL DEFAULT 0.00, -- 阻塞线程比例（百分比）
    waitingThreadRatioPercent REAL DEFAULT 0.00, -- 等待状态线程比例（百分比）
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 线程状态健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR)
    
    -- 时间字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, threadStateId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_THRST_THR ON HUB_MONITOR_JVM_THR_STATE(jvmThreadId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THRST_RES ON HUB_MONITOR_JVM_THR_STATE(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THRST_TIME ON HUB_MONITOR_JVM_THR_STATE(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THRST_BLOCK ON HUB_MONITOR_JVM_THR_STATE(blockedThreadCount);

-- ==========================================
-- 7. 死锁检测信息表
-- 存储JVM中检测到的死锁情况
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_DEADLOCK (
    deadlockId TEXT NOT NULL, -- 死锁记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmThreadId TEXT NOT NULL, -- 关联的JVM线程记录ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 死锁基本信息
    hasDeadlockFlag TEXT DEFAULT 'N' NOT NULL CHECK(hasDeadlockFlag IN ('Y','N')), -- 是否检测到死锁(Y是,N否)
    deadlockThreadCount INTEGER DEFAULT 0 NOT NULL, -- 死锁线程数量
    deadlockThreadIds TEXT DEFAULT NULL, -- 死锁线程ID列表，逗号分隔
    deadlockThreadNames TEXT DEFAULT NULL, -- 死锁线程名称列表，逗号分隔
    
    -- 死锁严重程度
    severityLevel TEXT DEFAULT NULL, -- 严重程度(LOW/MEDIUM/HIGH/CRITICAL)
    severityDescription TEXT DEFAULT NULL, -- 严重程度描述
    affectedThreadGroups INTEGER DEFAULT 0, -- 影响的线程组数量
    
    -- 时间信息
    detectionTime TEXT DEFAULT NULL, -- 死锁检测时间
    deadlockDurationMs INTEGER DEFAULT 0, -- 死锁持续时间（毫秒）
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 诊断信息
    descriptionText TEXT DEFAULT NULL, -- 死锁描述信息
    recommendedAction TEXT DEFAULT NULL, -- 建议的解决方案
    alertLevel TEXT DEFAULT NULL, -- 告警级别(INFO/WARNING/ERROR/CRITICAL/EMERGENCY)
    requiresActionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresActionFlag IN ('Y','N')), -- 是否需要立即处理(Y是,N否)
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, deadlockId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_THR ON HUB_MONITOR_JVM_DEADLOCK(jvmThreadId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_RES ON HUB_MONITOR_JVM_DEADLOCK(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_TIME ON HUB_MONITOR_JVM_DEADLOCK(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_FLAG ON HUB_MONITOR_JVM_DEADLOCK(hasDeadlockFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_SEV ON HUB_MONITOR_JVM_DEADLOCK(severityLevel);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_ALERT ON HUB_MONITOR_JVM_DEADLOCK(alertLevel);

-- ==========================================
-- 8. 类加载信息表
-- 存储JVM类加载器的统计信息
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_CLASS (
    classLoadingId TEXT NOT NULL, -- 类加载记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 类加载统计
    loadedClassCount INTEGER DEFAULT 0 NOT NULL, -- 当前已加载类数量
    totalLoadedClassCount INTEGER DEFAULT 0 NOT NULL, -- 总加载类数量
    unloadedClassCount INTEGER DEFAULT 0 NOT NULL, -- 已卸载类数量
    
    -- 比例指标
    classUnloadRatePercent REAL DEFAULT 0.00, -- 类卸载率（百分比）
    classRetentionRatePercent REAL DEFAULT 0.00, -- 类保留率（百分比）
    
    -- 配置状态
    verboseClassLoading TEXT DEFAULT 'N' NOT NULL CHECK(verboseClassLoading IN ('Y','N')), -- 是否启用详细类加载输出(Y是,N否)
    
    -- 性能指标
    loadingRatePerHour REAL DEFAULT 0.00, -- 每小时平均类加载数量
    loadingEfficiency REAL DEFAULT 0.00, -- 类加载效率
    memoryEfficiency TEXT DEFAULT NULL, -- 内存使用效率评估
    loaderHealth TEXT DEFAULT NULL, -- 类加载器健康状况
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 类加载健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR)
    requiresAttentionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresAttentionFlag IN ('Y','N')), -- 是否需要立即关注(Y是,N否)
    potentialIssuesJson TEXT DEFAULT NULL, -- 潜在问题列表，JSON格式
    
    -- 时间字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, classLoadingId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_CLS_RES ON HUB_MONITOR_JVM_CLASS(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_CLS_TIME ON HUB_MONITOR_JVM_CLASS(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_CLS_HEALTH ON HUB_MONITOR_JVM_CLASS(healthyFlag, requiresAttentionFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_CLS_COUNT ON HUB_MONITOR_JVM_CLASS(loadedClassCount);

-- ==========================================
-- 9. 应用监控数据表
-- 存储应用层面的各种监控数据（线程池、连接池、自定义指标等）
-- 对应 ThirdPartyMonitorData 采集的所有监控数据
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_APP_DATA (
    appDataId TEXT NOT NULL, -- 应用监控数据ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 数据分类标识
    dataType TEXT NOT NULL, -- 数据类型(THREAD_POOL:线程池/CONNECTION_POOL:连接池/CUSTOM_METRIC:自定义指标/CACHE_POOL:缓存池/MESSAGE_QUEUE:消息队列)
    dataName TEXT NOT NULL, -- 数据名称（如：线程池名称、指标名称等）
    dataCategory TEXT DEFAULT NULL, -- 数据分类（如：业务线程池/IO线程池/业务指标/技术指标）
    
    -- 监控数据（JSON格式存储，支持不同类型的数据结构）
    dataJson TEXT NOT NULL, -- 监控数据，JSON格式，包含具体的监控指标和值
    
    -- 核心指标（从JSON中提取的关键指标，便于查询和索引）
    primaryValue REAL DEFAULT NULL, -- 主要指标值（如：使用率、数量等）
    secondaryValue REAL DEFAULT NULL, -- 次要指标值（如：最大值、平均值等）
    statusValue TEXT DEFAULT NULL, -- 状态值（如：健康状态、连接状态等）
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR/CRITICAL)
    requiresAttentionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresAttentionFlag IN ('Y','N')), -- 是否需要立即关注(Y是,N否)
    
    -- 标签和维度（便于分组查询）
    tagsJson TEXT DEFAULT NULL, -- 标签信息，JSON格式（如：{"poolType":"business","environment":"prod"}）
    
    -- 时间字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, appDataId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_RES ON HUB_MONITOR_APP_DATA(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_TYPE ON HUB_MONITOR_APP_DATA(dataType);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_NAME ON HUB_MONITOR_APP_DATA(dataName);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_TIME ON HUB_MONITOR_APP_DATA(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_HEALTH ON HUB_MONITOR_APP_DATA(healthyFlag, requiresAttentionFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_PRIMARY ON HUB_MONITOR_APP_DATA(primaryValue);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_STATUS ON HUB_MONITOR_APP_DATA(statusValue);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_COMPOSITE ON HUB_MONITOR_APP_DATA(jvmResourceId, dataType, dataName, collectionTime);
-- =========================================
-- 基于FRP架构的隧道管理系统 - SQLite数据库表结构设计
-- 参考FRP（Fast Reverse Proxy）设计模式
-- 遵循naming-convention.md数据库规范
-- =========================================

-- =========================================
-- 1. 隧道服务器表（控制端口）
-- =========================================

-- 隧道服务器配置表 - 管理控制端口和核心配置
CREATE TABLE HUB_TUNNEL_SERVER (
  tunnelServerId TEXT NOT NULL PRIMARY KEY,
  tenantId TEXT NOT NULL,
  serverName TEXT NOT NULL UNIQUE,
  serverDescription TEXT,
  controlAddress TEXT NOT NULL DEFAULT '0.0.0.0',
  controlPort INTEGER NOT NULL DEFAULT 7000,
  dashboardPort INTEGER DEFAULT 7500,
  vhostHttpPort INTEGER DEFAULT 80,
  vhostHttpsPort INTEGER DEFAULT 443,
  maxClients INTEGER NOT NULL DEFAULT 1000,
  tokenAuth TEXT NOT NULL DEFAULT 'Y' CHECK(tokenAuth IN ('Y', 'N')),
  authToken TEXT,
  tlsEnable TEXT NOT NULL DEFAULT 'N' CHECK(tlsEnable IN ('Y', 'N')),
  tlsCertFile TEXT,
  tlsKeyFile TEXT,
  heartbeatInterval INTEGER NOT NULL DEFAULT 30,
  heartbeatTimeout INTEGER NOT NULL DEFAULT 90,
  logLevel TEXT NOT NULL DEFAULT 'info',
  maxPortsPerClient INTEGER DEFAULT 10,
  allowPorts TEXT,
  serverStatus TEXT NOT NULL DEFAULT 'stopped',
  startTime TEXT,
  configVersion TEXT,
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y' CHECK(activeFlag IN ('Y', 'N')),
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT
);

CREATE INDEX IDX_TUNNEL_SVR_TENANT ON HUB_TUNNEL_SERVER(tenantId);
CREATE INDEX IDX_TUNNEL_SVR_CTRL ON HUB_TUNNEL_SERVER(controlAddress, controlPort);
CREATE INDEX IDX_TUNNEL_SVR_STATUS ON HUB_TUNNEL_SERVER(serverStatus);

-- =========================================
-- 2. 服务器节点表（静态端口映射）
-- =========================================

CREATE TABLE HUB_TUNNEL_SERVER_NODE (
  serverNodeId TEXT NOT NULL PRIMARY KEY,
  tenantId TEXT NOT NULL,
  tunnelServerId TEXT NOT NULL,
  nodeName TEXT NOT NULL UNIQUE,
  nodeType TEXT NOT NULL DEFAULT 'static',
  proxyType TEXT NOT NULL,
  listenAddress TEXT NOT NULL DEFAULT '0.0.0.0',
  listenPort INTEGER NOT NULL,
  targetAddress TEXT NOT NULL,
  targetPort INTEGER NOT NULL,
  customDomains TEXT,
  subDomain TEXT,
  httpUser TEXT,
  httpPassword TEXT,
  hostHeaderRewrite TEXT,
  headers TEXT,
  locations TEXT,
  compression TEXT NOT NULL DEFAULT 'Y' CHECK(compression IN ('Y', 'N')),
  encryption TEXT NOT NULL DEFAULT 'N' CHECK(encryption IN ('Y', 'N')),
  secretKey TEXT,
  healthCheckType TEXT DEFAULT 'tcp',
  healthCheckUrl TEXT,
  healthCheckInterval INTEGER DEFAULT 60,
  maxConnections INTEGER DEFAULT 100,
  nodeStatus TEXT NOT NULL DEFAULT 'active',
  lastHealthCheck TEXT,
  connectionCount INTEGER DEFAULT 0,
  totalConnections INTEGER DEFAULT 0,
  totalBytes INTEGER DEFAULT 0,
  createdTime TEXT NOT NULL DEFAULT (datetime('now')),
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y' CHECK(activeFlag IN ('Y', 'N')),
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  UNIQUE(listenAddress, listenPort, proxyType)
);

CREATE INDEX IDX_TUNNEL_NODE_TENANT ON HUB_TUNNEL_SERVER_NODE(tenantId);
CREATE INDEX IDX_TUNNEL_NODE_SERVER ON HUB_TUNNEL_SERVER_NODE(tunnelServerId);
CREATE INDEX IDX_TUNNEL_NODE_TYPE ON HUB_TUNNEL_SERVER_NODE(nodeType, proxyType);
CREATE INDEX IDX_TUNNEL_NODE_STATUS ON HUB_TUNNEL_SERVER_NODE(nodeStatus);
CREATE INDEX IDX_TUNNEL_NODE_HEALTH ON HUB_TUNNEL_SERVER_NODE(lastHealthCheck);

-- =========================================
-- 3. 客户端注册表（动态连接）
-- =========================================

CREATE TABLE HUB_TUNNEL_CLIENT (
  tunnelClientId TEXT NOT NULL PRIMARY KEY,
  tenantId TEXT NOT NULL,
  userId TEXT NOT NULL,
  clientName TEXT NOT NULL UNIQUE,
  clientDescription TEXT,
  clientVersion TEXT,
  operatingSystem TEXT,
  clientIpAddress TEXT,
  clientMacAddress TEXT,
  serverAddress TEXT NOT NULL,
  serverPort INTEGER NOT NULL DEFAULT 7000,
  authToken TEXT NOT NULL,
  tlsEnable TEXT NOT NULL DEFAULT 'N' CHECK(tlsEnable IN ('Y', 'N')),
  autoReconnect TEXT NOT NULL DEFAULT 'Y' CHECK(autoReconnect IN ('Y', 'N')),
  maxRetries INTEGER NOT NULL DEFAULT 5,
  retryInterval INTEGER NOT NULL DEFAULT 20,
  heartbeatInterval INTEGER NOT NULL DEFAULT 30,
  heartbeatTimeout INTEGER NOT NULL DEFAULT 90,
  connectionStatus TEXT NOT NULL DEFAULT 'disconnected',
  lastConnectTime TEXT,
  lastDisconnectTime TEXT,
  totalConnectTime INTEGER DEFAULT 0,
  reconnectCount INTEGER DEFAULT 0,
  serviceCount INTEGER DEFAULT 0,
  lastHeartbeat TEXT,
  clientConfig TEXT,
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y' CHECK(activeFlag IN ('Y', 'N')),
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT
);

CREATE INDEX IDX_TUNNEL_CLIENT_TENANT ON HUB_TUNNEL_CLIENT(tenantId);
CREATE INDEX IDX_TUNNEL_CLIENT_USER ON HUB_TUNNEL_CLIENT(userId);
CREATE INDEX IDX_TUNNEL_CLIENT_STATUS ON HUB_TUNNEL_CLIENT(connectionStatus);
CREATE INDEX IDX_TUNNEL_CLIENT_IP ON HUB_TUNNEL_CLIENT(clientIpAddress);
CREATE INDEX IDX_TUNNEL_CLIENT_HB ON HUB_TUNNEL_CLIENT(lastHeartbeat);

-- =========================================
-- 4. 服务配置表（动态注册的服务）
-- =========================================

CREATE TABLE HUB_TUNNEL_SERVICE (
  tunnelServiceId TEXT NOT NULL PRIMARY KEY,
  tenantId TEXT NOT NULL,
  tunnelClientId TEXT NOT NULL,
  userId TEXT NOT NULL,
  serviceName TEXT NOT NULL UNIQUE,
  serviceDescription TEXT,
  serviceType TEXT NOT NULL,
  localAddress TEXT NOT NULL DEFAULT '127.0.0.1',
  localPort INTEGER NOT NULL,
  remotePort INTEGER,
  customDomains TEXT,
  subDomain TEXT,
  httpUser TEXT,
  httpPassword TEXT,
  hostHeaderRewrite TEXT,
  headers TEXT,
  locations TEXT,
  useEncryption TEXT NOT NULL DEFAULT 'N' CHECK(useEncryption IN ('Y', 'N')),
  useCompression TEXT NOT NULL DEFAULT 'Y' CHECK(useCompression IN ('Y', 'N')),
  secretKey TEXT,
  bandwidthLimit TEXT,
  maxConnections INTEGER DEFAULT 100,
  healthCheckType TEXT,
  healthCheckUrl TEXT,
  serviceStatus TEXT NOT NULL DEFAULT 'active',
  registeredTime TEXT NOT NULL DEFAULT (datetime('now')),
  lastActiveTime TEXT,
  connectionCount INTEGER DEFAULT 0,
  totalConnections INTEGER DEFAULT 0,
  totalTraffic INTEGER DEFAULT 0,
  serviceConfig TEXT,
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y' CHECK(activeFlag IN ('Y', 'N')),
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT
);

CREATE INDEX IDX_TUNNEL_SVC_TENANT ON HUB_TUNNEL_SERVICE(tenantId);
CREATE INDEX IDX_TUNNEL_SVC_CLIENT ON HUB_TUNNEL_SERVICE(tunnelClientId);
CREATE INDEX IDX_TUNNEL_SVC_USER ON HUB_TUNNEL_SERVICE(userId);
CREATE INDEX IDX_TUNNEL_SVC_TYPE ON HUB_TUNNEL_SERVICE(serviceType);
CREATE INDEX IDX_TUNNEL_SVC_STATUS ON HUB_TUNNEL_SERVICE(serviceStatus);
CREATE INDEX IDX_TUNNEL_SVC_PORT ON HUB_TUNNEL_SERVICE(remotePort);
CREATE INDEX IDX_TUNNEL_SVC_DOMAIN ON HUB_TUNNEL_SERVICE(subDomain);

-- =====================================================
-- 权限系统数据库表结构设计
-- 遵循 docs/database/naming-convention.md 规范
-- 基于 web/actions/permission-design.md 设计文档
-- 创建时间: 2024-12-19
-- =====================================================

-- =====================================================
-- 角色表 - 存储系统角色信息和数据权限范围
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_ROLE (
  -- 主键和租户信息
  roleId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 角色基本信息
  roleName TEXT NOT NULL,
  roleDescription TEXT,
  
  -- 角色状态
  roleStatus TEXT NOT NULL DEFAULT 'Y',
  builtInFlag TEXT NOT NULL DEFAULT 'N',
  
  -- 数据权限范围
  dataScope TEXT DEFAULT NULL,
  
  -- 通用字段
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, roleId)
);

CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_NAME ON HUB_AUTH_ROLE(tenantId, roleName);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_STATUS ON HUB_AUTH_ROLE(roleStatus);

-- =====================================================
-- 权限资源表 - 存储系统所有权限资源信息
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_RESOURCE (
  -- 主键和租户信息
  resourceId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 资源基本信息
  resourceName TEXT NOT NULL,
  resourceCode TEXT NOT NULL,
  resourceType TEXT NOT NULL,
  resourcePath TEXT,
  resourceMethod TEXT,
  
  -- 层级关系
  parentResourceId TEXT,
  resourceLevel INTEGER NOT NULL DEFAULT 1,
  sortOrder INTEGER NOT NULL DEFAULT 0,
  
  -- 显示信息
  displayName TEXT,
  iconClass TEXT,
  description TEXT,
  language TEXT DEFAULT 'zh-CN', -- 语言标识（如：zh-CN, en-US），用于多语言支持，默认zh-CN
  
  -- 状态信息
  resourceStatus TEXT NOT NULL DEFAULT 'Y',
  builtInFlag TEXT NOT NULL DEFAULT 'N',
  
  -- 通用字段
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, resourceId)
);

CREATE UNIQUE INDEX IF NOT EXISTS IDX_AUTH_RES_CODE ON HUB_AUTH_RESOURCE(tenantId, resourceCode);
CREATE INDEX IF NOT EXISTS IDX_AUTH_RES_TYPE ON HUB_AUTH_RESOURCE(resourceType);
CREATE INDEX IF NOT EXISTS IDX_AUTH_RES_PARENT ON HUB_AUTH_RESOURCE(parentResourceId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_RES_PATH ON HUB_AUTH_RESOURCE(resourcePath);
CREATE INDEX IF NOT EXISTS IDX_AUTH_RES_STATUS ON HUB_AUTH_RESOURCE(resourceStatus);
CREATE INDEX IF NOT EXISTS IDX_AUTH_RES_LEVEL ON HUB_AUTH_RESOURCE(resourceLevel);
CREATE INDEX IF NOT EXISTS IDX_AUTH_RES_SORT ON HUB_AUTH_RESOURCE(sortOrder);

-- =====================================================
-- 角色权限关联表 - 存储角色与权限资源的关联关系
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_ROLE_RESOURCE (
  -- 主键和租户信息
  roleResourceId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联信息
  roleId TEXT NOT NULL,
  resourceId TEXT NOT NULL,
  
  -- 权限控制
  permissionType TEXT NOT NULL DEFAULT 'ALLOW',
  grantedBy TEXT NOT NULL,
  grantedTime TEXT NOT NULL DEFAULT (datetime('now')),
  expireTime TEXT,
  
  -- 通用字段
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, roleResourceId)
);

CREATE UNIQUE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_UNIQUE ON HUB_AUTH_ROLE_RESOURCE(tenantId, roleId, resourceId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_ROLE ON HUB_AUTH_ROLE_RESOURCE(tenantId, roleId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_RESOURCE ON HUB_AUTH_ROLE_RESOURCE(tenantId, resourceId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_TYPE ON HUB_AUTH_ROLE_RESOURCE(permissionType);
CREATE INDEX IF NOT EXISTS IDX_AUTH_ROLE_RES_EXPIRE ON HUB_AUTH_ROLE_RESOURCE(expireTime);

-- =====================================================
-- 用户角色关联表 - 存储用户与角色的关联关系
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_USER_ROLE (
  -- 主键和租户信息
  userRoleId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联信息
  userId TEXT NOT NULL,
  roleId TEXT NOT NULL,
  
  -- 授权控制
  grantedBy TEXT NOT NULL,
  grantedTime TEXT NOT NULL DEFAULT (datetime('now')),
  expireTime TEXT,
  primaryRoleFlag TEXT NOT NULL DEFAULT 'N',
  
  -- 通用字段
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, userRoleId)
);

CREATE UNIQUE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_UNIQUE ON HUB_AUTH_USER_ROLE(tenantId, userId, roleId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_USER ON HUB_AUTH_USER_ROLE(tenantId, userId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_ROLE ON HUB_AUTH_USER_ROLE(tenantId, roleId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_PRIMARY ON HUB_AUTH_USER_ROLE(primaryRoleFlag);
CREATE INDEX IF NOT EXISTS IDX_AUTH_USER_ROLE_EXPIRE ON HUB_AUTH_USER_ROLE(expireTime);

-- =====================================================
-- 数据权限表 - 存储用户和角色的数据访问权限
-- =====================================================
CREATE TABLE IF NOT EXISTS HUB_AUTH_DATA_PERMISSION (
  -- 主键和租户信息
  dataPermissionId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 关联信息
  userId TEXT,
  roleId TEXT,
  
  -- 数据权限信息
  resourceType TEXT NOT NULL,
  resourceCode TEXT NOT NULL,
  scopeValue TEXT,
  
  -- 权限条件
  filterCondition TEXT,
  columnPermissions TEXT,
  operationPermissions TEXT DEFAULT 'read',
  
  -- 生效时间
  effectiveTime TEXT,
  expireTime TEXT,
  
  -- 通用字段
  addTime TEXT NOT NULL DEFAULT (datetime('now')),
  addWho TEXT NOT NULL,
  editTime TEXT NOT NULL DEFAULT (datetime('now')),
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  reserved4 TEXT,
  reserved5 TEXT,
  reserved6 TEXT,
  reserved7 TEXT,
  reserved8 TEXT,
  reserved9 TEXT,
  reserved10 TEXT,
  
  PRIMARY KEY (tenantId, dataPermissionId)
);

CREATE INDEX IF NOT EXISTS IDX_AUTH_DATA_PERM_USER ON HUB_AUTH_DATA_PERMISSION(tenantId, userId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_DATA_PERM_ROLE ON HUB_AUTH_DATA_PERMISSION(tenantId, roleId);
CREATE INDEX IF NOT EXISTS IDX_AUTH_DATA_PERM_RESOURCE ON HUB_AUTH_DATA_PERMISSION(resourceType, resourceCode);
CREATE INDEX IF NOT EXISTS IDX_AUTH_DATA_PERM_EXPIRE ON HUB_AUTH_DATA_PERMISSION(expireTime);

-- =====================================================
-- 权限系统初始化数据
-- 基于 web/frontend/src/router/staticRoutes.ts 路由配置
-- =====================================================

-- =====================================================
-- 初始化角色数据
-- =====================================================

-- 超级管理员角色
INSERT INTO HUB_AUTH_ROLE (
  roleId, tenantId, roleName, roleDescription, 
  roleStatus, builtInFlag, dataScope,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_SUPER_ADMIN', 'default', '超级管理员', '拥有系统所有权限的超级管理员',
  'Y', 'Y', '{"type":"ALL"}',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_001', 1, 'Y'
);

-- =====================================================
-- 初始化模块资源数据
-- 基于 staticRoutes.ts 中的路由配置
-- =====================================================

-- 系统监控模块 (hub0000) - 路径: /dashboard
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0000', 'default', '系统监控模块', 'hub0000', 'MODULE',
  '/dashboard', 1, 1, '系统监控', 'HomeOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_001', 1, 'Y'
);

-- 用户登录模块 (hub0001) - 路径: /login - 不需要权限验证
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0001', 'default', '用户登录模块', 'hub0001', 'MODULE',
  '/login', 1, 2, '用户登录', 'LogInOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_002', 1, 'Y'
);

-- 用户管理模块 (hub0002) - 路径: /system/userManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0002', 'default', '用户管理模块', 'hub0002', 'MODULE',
  '/system/userManagement', 1, 3, '用户管理', 'PeopleOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003', 1, 'Y'
);

-- 用户管理模块 - 按钮资源
-- 新增按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0002:add', 'default', '新增用户', 'hub0002:add', 'BUTTON',
  'hub0002', 2, 1, '新增', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_001', 1, 'Y'
);

-- 编辑按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0002:edit', 'default', '编辑用户', 'hub0002:edit', 'BUTTON',
  'hub0002', 2, 2, '编辑', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_002', 1, 'Y'
);

-- 删除按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0002:delete', 'default', '删除用户', 'hub0002:delete', 'BUTTON',
  'hub0002', 2, 3, '删除', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_003', 1, 'Y'
);

-- 重置密码按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0002:resetPassword', 'default', '重置密码', 'hub0002:resetPassword', 'BUTTON',
  'hub0002', 2, 4, '重置密码', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_004', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0002:view', 'default', '查看详情', 'hub0002:view', 'BUTTON',
  'hub0002', 2, 5, '查看详情', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_005', 1, 'Y'
);

-- 查询按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0002:search', 'default', '查询用户', 'hub0002:search', 'BUTTON',
  'hub0002', 2, 6, '查询', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_006', 1, 'Y'
);

-- 重置按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0002:reset', 'default', '重置查询', 'hub0002:reset', 'BUTTON',
  'hub0002', 2, 7, '重置', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_007', 1, 'Y'
);

-- 角色管理模块 (hub0005) - 路径: /system/roleManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0005', 'default', '角色管理模块', 'hub0005', 'MODULE',
  '/system/roleManagement', 1, 4, '角色管理', 'PeopleCircleOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004', 1, 'Y'
);

-- 角色管理模块 - 按钮资源
-- 新增按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0005:add', 'default', '新增角色', 'hub0005:add', 'BUTTON',
  'hub0005', 2, 1, '新增', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_001', 1, 'Y'
);

-- 编辑按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0005:edit', 'default', '编辑角色', 'hub0005:edit', 'BUTTON',
  'hub0005', 2, 2, '编辑', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_002', 1, 'Y'
);

-- 删除按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0005:delete', 'default', '删除角色', 'hub0005:delete', 'BUTTON',
  'hub0005', 2, 3, '删除', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_003', 1, 'Y'
);

-- 查看详情按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0005:view', 'default', '查看详情', 'hub0005:view', 'BUTTON',
  'hub0005', 2, 4, '查看详情', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_004', 1, 'Y'
);

-- 角色授权按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0005:roleAuth', 'default', '角色授权', 'hub0005:roleAuth', 'BUTTON',
  'hub0005', 2, 5, '角色授权', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_005', 1, 'Y'
);

-- 查询按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0005:search', 'default', '查询角色', 'hub0005:search', 'BUTTON',
  'hub0005', 2, 6, '查询', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_006', 1, 'Y'
);

-- 重置按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0005:reset', 'default', '重置查询', 'hub0005:reset', 'BUTTON',
  'hub0005', 2, 7, '重置', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_007', 1, 'Y'
);

-- 权限资源管理模块 (hub0006) - 路径: /system/resourceManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0006', 'default', '权限资源管理模块', 'hub0006', 'MODULE',
  '/system/resourceManagement', 1, 5, '权限资源管理', 'KeyOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005', 1, 'Y'
);

-- 权限资源管理模块 - 按钮资源
-- 查看详情按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0006:view', 'default', '查看详情', 'hub0006:view', 'BUTTON',
  'hub0006', 2, 1, '查看详情', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_001', 1, 'Y'
);

-- 查询按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0006:search', 'default', '查询资源', 'hub0006:search', 'BUTTON',
  'hub0006', 2, 2, '查询', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_002', 1, 'Y'
);

-- 重置按钮
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  parentResourceId, resourceLevel, sortOrder, displayName, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0006:reset', 'default', '重置查询', 'hub0006:reset', 'BUTTON',
  'hub0006', 2, 3, '重置', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_003', 1, 'Y'
);

-- 网关实例管理模块 (hub0020) - 路径: /gateway/gatewayInstanceManager
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0020', 'default', '网关实例管理模块', 'hub0020', 'MODULE',
  '/gateway/gatewayInstanceManager', 1, 10, '实例管理', 'ServerOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_010', 1, 'Y'
);

-- 路由管理模块 (hub0021) - 路径: /gateway/routeManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0021', 'default', '路由管理模块', 'hub0021', 'MODULE',
  '/gateway/routeManagement', 1, 11, '路由管理', 'GitNetworkOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_011', 1, 'Y'
);

-- 代理管理模块 (hub0022) - 路径: /gateway/proxyManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0022', 'default', '代理管理模块', 'hub0022', 'MODULE',
  '/gateway/proxyManagement', 1, 12, '代理管理', 'FlashOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_012', 1, 'Y'
);

-- 网关日志管理模块 (hub0023) - 路径: /gateway/gatewayLogManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0023', 'default', '网关日志管理模块', 'hub0023', 'MODULE',
  '/gateway/gatewayLogManagement', 1, 13, '网关日志管理', 'DocumentTextOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_013', 1, 'Y'
);

-- 命名空间管理模块 (hub0040) - 路径: /serviceGovernance/namespaceManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0040', 'default', '命名空间管理模块', 'hub0040', 'MODULE',
  '/serviceGovernance/namespaceManagement', 1, 20, '命名空间管理', 'LayersOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_020', 1, 'Y'
);

-- 服务注册管理模块 (hub0041) - 路径: /serviceGovernance/serviceRegistryManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0041', 'default', '服务注册管理模块', 'hub0041', 'MODULE',
  '/serviceGovernance/serviceRegistryManagement', 1, 21, '服务注册管理', 'ListOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_021', 1, 'Y'
);

-- 服务监控模块 (hub0042) - 路径: /serviceGovernance/serviceMonitoring
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0042', 'default', '服务监控模块', 'hub0042', 'MODULE',
  '/serviceGovernance/serviceMonitoring', 1, 22, '服务监控', 'BarChartOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_022', 1, 'Y'
);

-- 隧道服务器模块 (hub0060) - 路径: /tunnel/tunnelServerManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0060', 'default', '隧道服务器模块', 'hub0060', 'MODULE',
  '/tunnel/tunnelServerManagement', 1, 30, '隧道服务器', 'ServerOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_030', 1, 'Y'
);

-- 静态映射模块 (hub0061) - 路径: /tunnel/staticMappingManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0061', 'default', '静态映射模块', 'hub0061', 'MODULE',
  '/tunnel/staticMappingManagement', 1, 31, '静态映射', 'GitNetworkOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_031', 1, 'Y'
);

-- 隧道客户端模块 (hub0062) - 路径: /tunnel/tunnelClientManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0062', 'default', '隧道客户端模块', 'hub0062', 'MODULE',
  '/tunnel/tunnelClientManagement', 1, 32, '隧道客户端', 'DesktopOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_032', 1, 'Y'
);

-- 动态服务管理模块 (hub0063) - 路径: /tunnel/tunnelServiceManagement
INSERT INTO HUB_AUTH_RESOURCE (
  resourceId, tenantId, resourceName, resourceCode, resourceType,
  resourcePath, resourceLevel, sortOrder, displayName, iconClass, language,
  resourceStatus, builtInFlag,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'hub0063', 'default', '动态服务管理模块', 'hub0063', 'MODULE',
  '/tunnel/tunnelServiceManagement', 1, 33, '动态服务管理', 'GridOutline', 'zh-CN',
  'Y', 'Y',
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_033', 1, 'Y'
);

-- =====================================================
-- 初始化角色权限关联数据
-- 为超级管理员角色分配所有模块权限
-- =====================================================

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0000', 'default', 'ROLE_SUPER_ADMIN', 'hub0000', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_001', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0001', 'default', 'ROLE_SUPER_ADMIN', 'hub0001', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_002', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002', 'default', 'ROLE_SUPER_ADMIN', 'hub0002', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003', 1, 'Y'
);

-- 超级管理员 - 用户管理模块按钮权限
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_ADD', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:add', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_001', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_EDIT', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:edit', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_002', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_DELETE', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:delete', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_003', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_RESET_PWD', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:resetPassword', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_004', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:view', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_005', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:search', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_006', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0002_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0002:reset', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_003_007', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005', 'default', 'ROLE_SUPER_ADMIN', 'hub0005', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004', 1, 'Y'
);

-- 超级管理员 - 角色管理模块按钮权限
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_ADD', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:add', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_001', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_EDIT', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:edit', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_002', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_DELETE', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:delete', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_003', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:view', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_004', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_ROLE_AUTH', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:roleAuth', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_005', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:search', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_006', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0005_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0005:reset', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_004_007', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006', 'default', 'ROLE_SUPER_ADMIN', 'hub0006', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005', 1, 'Y'
);

-- 超级管理员 - 权限资源管理模块按钮权限
INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_VIEW', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:view', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_001', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_SEARCH', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:search', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_002', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0006_BTN_RESET', 'default', 'ROLE_SUPER_ADMIN', 'hub0006:reset', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_005_003', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0020', 'default', 'ROLE_SUPER_ADMIN', 'hub0020', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_010', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0021', 'default', 'ROLE_SUPER_ADMIN', 'hub0021', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_011', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0022', 'default', 'ROLE_SUPER_ADMIN', 'hub0022', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_012', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0023', 'default', 'ROLE_SUPER_ADMIN', 'hub0023', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_013', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0040', 'default', 'ROLE_SUPER_ADMIN', 'hub0040', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_020', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0041', 'default', 'ROLE_SUPER_ADMIN', 'hub0041', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_021', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0042', 'default', 'ROLE_SUPER_ADMIN', 'hub0042', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_022', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0060', 'default', 'ROLE_SUPER_ADMIN', 'hub0060', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_030', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0061', 'default', 'ROLE_SUPER_ADMIN', 'hub0061', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_031', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0062', 'default', 'ROLE_SUPER_ADMIN', 'hub0062', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_032', 1, 'Y'
);

INSERT INTO HUB_AUTH_ROLE_RESOURCE (
  roleResourceId, tenantId, roleId, resourceId, permissionType, grantedBy, grantedTime,
  addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag
) VALUES (
  'ROLE_RES_SUPER_ADMIN_HUB0063', 'default', 'ROLE_SUPER_ADMIN', 'hub0063', 'ALLOW', 'system', datetime('now'),
  datetime('now'), 'system', datetime('now'), 'system', 'INIT_033', 1, 'Y'
);

-- ==========================================
-- 索引说明
-- ==========================================
-- 1. 所有表都建立了tenantId相关的复合主键，支持多租户数据隔离
-- 2. 为关联字段（jvmResourceId等）创建了索引，提高关联查询性能
-- 3. 为时间字段（collectionTime）创建了索引，支持时间范围查询
-- 4. 为健康状态字段创建了索引，便于快速筛选异常数据
-- 5. 为常用查询条件字段创建了索引，提高查询效率

-- ==========================================
-- 表关系说明
-- ==========================================
-- HUB_MONITOR_JVM_RESOURCE (主表)
--   ├── HUB_MONITOR_JVM_MEMORY (1:N，一个JVM资源对应多个内存记录：堆内存+非堆内存)
--   ├── HUB_MONITOR_JVM_MEM_POOL (1:N，一个JVM资源对应多个内存池)
--   ├── HUB_MONITOR_JVM_GC (1:N，一个JVM资源对应多个GC收集器)
--   ├── HUB_MONITOR_JVM_THREAD (1:1，一个JVM资源对应一个线程信息记录)
--   │   ├── HUB_MONITOR_JVM_THR_STATE (1:1，一个线程信息对应一个线程状态统计)
--   │   └── HUB_MONITOR_JVM_DEADLOCK (1:1，一个线程信息对应一个死锁检测记录)
--   ├── HUB_MONITOR_JVM_CLASS (1:1，一个JVM资源对应一个类加载信息记录)
--   └── HUB_MONITOR_APP_DATA (1:N，一个JVM资源对应多个应用监控数据)
-- ==========================================

-- =====================================================
-- SQLite特殊配置和优化
-- =====================================================

-- 设置同步模式为NORMAL以平衡性能和安全性
PRAGMA synchronous = NORMAL;

-- 设置页面缓存大小（2MB）
PRAGMA cache_size = -2000;

-- 设置临时存储为内存模式
PRAGMA temp_store = MEMORY;

-- 设置锁定超时时间（30秒）
PRAGMA busy_timeout = 30000;

-- 分析数据库以优化查询计划
ANALYZE;

-- =====================================================
-- 脚本执行完成提示
-- =====================================================
SELECT 'SQLite数据库初始化完成！' as message,
       'Created ' || COUNT(*) || ' tables' as table_count
FROM sqlite_master 
WHERE type = 'table' AND name LIKE 'HUB_%';