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
CREATE TABLE IF NOT EXISTS HUB_GW_USERAGENT_ACCESS_CONFIG (
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

CREATE INDEX IF NOT EXISTS idx_HUB_GW_USERAGENT_ACCESS_CONFIG_security ON HUB_GW_USERAGENT_ACCESS_CONFIG(securityConfigId);

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
CREATE TABLE IF NOT EXISTS HUB_METRIC_TEMPERATURE_LOG (
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

CREATE INDEX IDX_METRIC_TEMP_SERVER ON HUB_METRIC_TEMPERATURE_LOG(metricServerId);
CREATE INDEX IDX_METRIC_TEMP_TIME ON HUB_METRIC_TEMPERATURE_LOG(collectTime);
CREATE INDEX IDX_METRIC_TEMP_SENSOR ON HUB_METRIC_TEMPERATURE_LOG(sensorName);
CREATE INDEX IDX_METRIC_TEMP_ACTIVE ON HUB_METRIC_TEMPERATURE_LOG(activeFlag);
CREATE INDEX IDX_METRIC_TEMP_SRV_TIME ON HUB_METRIC_TEMPERATURE_LOG(metricServerId, collectTime);
CREATE INDEX IDX_METRIC_TEMP_SRV_SENSOR ON HUB_METRIC_TEMPERATURE_LOG(metricServerId, sensorName);
CREATE INDEX IDX_METRIC_TEMP_TNT_TIME ON HUB_METRIC_TEMPERATURE_LOG(tenantId, collectTime);

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

-- 外部注册中心配置表 - 存储外部注册中心连接配置
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_EXTERNAL_CONFIG (
  -- 主键和租户信息
  externalConfigId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 配置基本信息
  configName TEXT NOT NULL,
  configDescription TEXT,
  registryType TEXT NOT NULL,
  environmentName TEXT NOT NULL DEFAULT 'default',
  
  -- 连接配置
  serverAddress TEXT NOT NULL,
  serverPort INTEGER,
  serverPath TEXT,
  serverScheme TEXT DEFAULT 'http',
  
  -- 认证配置
  authEnabled TEXT DEFAULT 'N',
  username TEXT,
  password TEXT,
  accessToken TEXT,
  secretKey TEXT,
  
  -- 连接配置
  connectionTimeout INTEGER DEFAULT 5000,
  readTimeout INTEGER DEFAULT 10000,
  maxRetries INTEGER DEFAULT 3,
  retryInterval INTEGER DEFAULT 1000,
  
  -- 特定配置
  specificConfig TEXT,
  fieldMapping TEXT,
  
  -- 故障转移配置
  failoverEnabled TEXT DEFAULT 'N',
  failoverConfigId TEXT,
  failoverStrategy TEXT DEFAULT 'MANUAL',
  
  -- 数据同步配置
  syncEnabled TEXT DEFAULT 'N',
  syncInterval INTEGER DEFAULT 30,
  conflictResolution TEXT DEFAULT 'primary_wins',
  
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
  
  PRIMARY KEY (tenantId, externalConfigId)
);

-- 外部注册中心状态表 - 存储外部注册中心运行状态
CREATE TABLE IF NOT EXISTS HUB_REGISTRY_EXTERNAL_STATUS (
  -- 主键和租户信息
  externalStatusId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  externalConfigId TEXT NOT NULL,
  
  -- 连接状态
  connectionStatus TEXT NOT NULL DEFAULT 'DISCONNECTED',
  healthStatus TEXT NOT NULL DEFAULT 'UNKNOWN',
  lastConnectTime DATETIME,
  lastDisconnectTime DATETIME,
  lastHealthCheckTime DATETIME,
  
  -- 性能指标
  responseTime INTEGER DEFAULT 0,
  successCount INTEGER DEFAULT 0,
  errorCount INTEGER DEFAULT 0,
  timeoutCount INTEGER DEFAULT 0,
  
  -- 故障转移状态
  failoverStatus TEXT DEFAULT 'NORMAL',
  failoverTime DATETIME,
  failoverCount INTEGER DEFAULT 0,
  recoverTime DATETIME,
  
  -- 同步状态
  syncStatus TEXT DEFAULT 'IDLE',
  lastSyncTime DATETIME,
  syncSuccessCount INTEGER DEFAULT 0,
  syncErrorCount INTEGER DEFAULT 0,
  
  -- 错误信息
  lastErrorMessage TEXT,
  lastErrorTime DATETIME,
  errorDetails TEXT,
  
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
  
  PRIMARY KEY (tenantId, externalStatusId)
);

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

-- HUB_REGISTRY_EXTERNAL_CONFIG 索引
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_CONFIG_NAME ON HUB_REGISTRY_EXTERNAL_CONFIG(tenantId, configName, environmentName);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_CONFIG_TYPE ON HUB_REGISTRY_EXTERNAL_CONFIG(registryType);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_CONFIG_ENV ON HUB_REGISTRY_EXTERNAL_CONFIG(environmentName);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_CONFIG_ACTIVE ON HUB_REGISTRY_EXTERNAL_CONFIG(activeFlag);

-- HUB_REGISTRY_EXTERNAL_STATUS 索引
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_STATUS_CONFIG ON HUB_REGISTRY_EXTERNAL_STATUS(tenantId, externalConfigId);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_STATUS_CONN ON HUB_REGISTRY_EXTERNAL_STATUS(connectionStatus);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_STATUS_HEALTH ON HUB_REGISTRY_EXTERNAL_STATUS(healthStatus);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_STATUS_FAILOVER ON HUB_REGISTRY_EXTERNAL_STATUS(failoverStatus);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_STATUS_SYNC ON HUB_REGISTRY_EXTERNAL_STATUS(syncStatus);
CREATE INDEX IF NOT EXISTS IDX_REGISTRY_EXT_STATUS_ACTIVE ON HUB_REGISTRY_EXTERNAL_STATUS(activeFlag);

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