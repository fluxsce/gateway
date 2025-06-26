CREATE TABLE HUB_GATEWAY_INSTANCE (
                                      tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                      gatewayInstanceId VARCHAR2(32) NOT NULL, -- 网关实例ID
                                      instanceName VARCHAR2(100) NOT NULL, -- 实例名称
                                      instanceDesc VARCHAR2(200), -- 实例描述
                                      bindAddress VARCHAR2(100) DEFAULT '0.0.0.0', -- 绑定地址

    -- HTTP/HTTPS 端口配置
                                      httpPort NUMBER(10), -- HTTP监听端口
                                      httpsPort NUMBER(10), -- HTTPS监听端口
                                      tlsEnabled VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用TLS(N否,Y是)

    -- 证书配置 - 支持文件路径和数据库存储
                                      certStorageType VARCHAR2(20) DEFAULT 'FILE' NOT NULL, -- 证书存储类型(FILE文件,DATABASE数据库)
                                      certFilePath VARCHAR2(255), -- 证书文件路径
                                      keyFilePath VARCHAR2(255), -- 私钥文件路径
                                      certContent CLOB, -- 证书内容(PEM格式)
                                      keyContent CLOB, -- 私钥内容(PEM格式)
                                      certChainContent CLOB, -- 证书链内容(PEM格式)
                                      certPassword VARCHAR2(255), -- 证书密码(加密存储)

    -- Go HTTP Server 核心配置
                                      maxConnections NUMBER(10) DEFAULT 10000 NOT NULL, -- 最大连接数
                                      readTimeoutMs NUMBER(10) DEFAULT 30000 NOT NULL, -- 读取超时时间(毫秒)
                                      writeTimeoutMs NUMBER(10) DEFAULT 30000 NOT NULL, -- 写入超时时间(毫秒)
                                      idleTimeoutMs NUMBER(10) DEFAULT 60000 NOT NULL, -- 空闲连接超时时间(毫秒)
                                      maxHeaderBytes NUMBER(10) DEFAULT 1048576 NOT NULL, -- 最大请求头字节数(默认1MB)

    -- 性能和并发配置
                                      maxWorkers NUMBER(10) DEFAULT 1000 NOT NULL, -- 最大工作协程数
                                      keepAliveEnabled VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用Keep-Alive(N否,Y是)
                                      tcpKeepAliveEnabled VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用TCP Keep-Alive(N否,Y是)
                                      gracefulShutdownTimeoutMs NUMBER(10) DEFAULT 30000 NOT NULL, -- 优雅关闭超时时间(毫秒)

    -- TLS安全配置
                                      enableHttp2 VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用HTTP/2(N否,Y是)
                                      tlsVersion VARCHAR2(10) DEFAULT '1.2', -- TLS协议版本(1.0,1.1,1.2,1.3)
                                      tlsCipherSuites VARCHAR2(1000), -- TLS密码套件列表,逗号分隔
                                      disableGeneralOptionsHandler VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否禁用默认OPTIONS处理器(N否,Y是)

    -- 日志配置关联字段
                                      logConfigId VARCHAR2(32), -- 关联的日志配置ID
                                      healthStatus VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 健康状态(N不健康,Y健康)
                                      lastHeartbeatTime DATE, -- 最后心跳时间
                                      instanceMetadata CLOB, -- 实例元数据,JSON格式
                                      reserved1 VARCHAR2(100), -- 预留字段1
                                      reserved2 VARCHAR2(100), -- 预留字段2
                                      reserved3 NUMBER(10), -- 预留字段3
                                      reserved4 NUMBER(10), -- 预留字段4
                                      reserved5 DATE, -- 预留字段5
                                      extProperty CLOB, -- 扩展属性,JSON格式
                                      addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                      addWho VARCHAR2(32) NOT NULL, -- 创建人ID
                                      editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                      editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                      oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
                                      currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                      activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                      noteText VARCHAR2(500), -- 备注信息

                                      CONSTRAINT PK_HUB_GATEWAY_INSTANCE PRIMARY KEY (tenantId, gatewayInstanceId)
);
-- 然后使用以下语句创建索引
CREATE INDEX idx_HUB_GATEWAY_INSTANCE_bind_http ON HUB_GATEWAY_INSTANCE(bindAddress, httpPort);
CREATE INDEX idx_HUB_GATEWAY_INSTANCE_bind_https ON HUB_GATEWAY_INSTANCE(bindAddress, httpsPort);
CREATE INDEX idx_HUB_GATEWAY_INSTANCE_log ON HUB_GATEWAY_INSTANCE(logConfigId);
CREATE INDEX idx_HUB_GATEWAY_INSTANCE_health ON HUB_GATEWAY_INSTANCE(healthStatus);
CREATE INDEX idx_HUB_GATEWAY_INSTANCE_tls ON HUB_GATEWAY_INSTANCE(tlsEnabled);
-- Oracle 不直接支持在DDL中指定表级注释，需要使用单独的COMMENT ON语句。
COMMENT ON TABLE HUB_GATEWAY_INSTANCE IS '网关实例表 - 记录网关实例基础配置，完整支持Go HTTP Server配置';

CREATE TABLE HUB_GATEWAY_ROUTER_CONFIG (
                                           tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                           routerConfigId VARCHAR2(32) NOT NULL, -- Router配置ID
                                           gatewayInstanceId VARCHAR2(32) NOT NULL, -- 关联的网关实例ID
                                           routerName VARCHAR2(100) NOT NULL, -- Router名称
                                           routerDesc VARCHAR2(200), -- Router描述

    -- Router基础配置
                                           defaultPriority NUMBER(10) DEFAULT 100 NOT NULL, -- 默认路由优先级
                                           enableRouteCache VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用路由缓存(N否,Y是)
                                           routeCacheTtlSeconds NUMBER(10) DEFAULT 300 NOT NULL, -- 路由缓存TTL(秒)
                                           maxRoutes NUMBER(10) DEFAULT 1000, -- 最大路由数量限制
                                           routeMatchTimeout NUMBER(10) DEFAULT 100, -- 路由匹配超时时间(毫秒)

    -- Router高级配置
                                           enableStrictMode VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用严格模式(N否,Y是)
                                           enableMetrics VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用路由指标收集(N否,Y是)
                                           enableTracing VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用链路追踪(N否,Y是)
                                           caseSensitive VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 路径匹配是否区分大小写(N否,Y是)
                                           removeTrailingSlash VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否移除路径尾部斜杠(N否,Y是)

    -- 路由处理配置
                                           enableGlobalFilters VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用全局过滤器(N否,Y是)
                                           filterExecutionMode VARCHAR2(20) DEFAULT 'SEQUENTIAL' NOT NULL, -- 过滤器执行模式(SEQUENTIAL顺序,PARALLEL并行)
                                           maxFilterChainDepth NUMBER(10) DEFAULT 50, -- 最大过滤器链深度

    -- 性能优化配置
                                           enableRoutePooling VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用路由对象池(N否,Y是)
                                           routePoolSize NUMBER(10) DEFAULT 100, -- 路由对象池大小
                                           enableAsyncProcessing VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用异步处理(N否,Y是)

    -- 错误处理配置
                                           enableFallback VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用降级处理(N否,Y是)
                                           fallbackRoute VARCHAR2(200), -- 降级路由路径
                                           notFoundStatusCode NUMBER(10) DEFAULT 404 NOT NULL, -- 路由未找到时的状态码
                                           notFoundMessage VARCHAR2(200) DEFAULT 'Route not found', -- 路由未找到时的提示消息

    -- 自定义配置
                                           routerMetadata CLOB, -- Router元数据,JSON格式
                                           customConfig CLOB, -- 自定义配置,JSON格式

    -- 标准数据库字段
                                           reserved1 VARCHAR2(100), -- 预留字段1
                                           reserved2 VARCHAR2(100), -- 预留字段2
                                           reserved3 NUMBER(10), -- 预留字段3
                                           reserved4 NUMBER(10), -- 预留字段4
                                           reserved5 DATE, -- 预留字段5
                                           extProperty CLOB, -- 扩展属性,JSON格式
                                           addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                           addWho VARCHAR2(32) NOT NULL, -- 创建人ID
                                           editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                           editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                           oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
                                           currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                           activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动/禁用,Y活动/启用)
                                           noteText VARCHAR2(500), -- 备注信息

                                           CONSTRAINT PK_HUB_GATEWAY_ROUTER_CONFIG PRIMARY KEY (tenantId, routerConfigId)
);
CREATE INDEX idx_HUB_GATEWAY_ROUTER_CONFIG_instance ON HUB_GATEWAY_ROUTER_CONFIG(gatewayInstanceId);
CREATE INDEX idx_HUB_GATEWAY_ROUTER_CONFIG_name ON HUB_GATEWAY_ROUTER_CONFIG(routerName);
CREATE INDEX idx_HUB_GATEWAY_ROUTER_CONFIG_active ON HUB_GATEWAY_ROUTER_CONFIG(activeFlag);
CREATE INDEX idx_HUB_GATEWAY_ROUTER_CONFIG_cache ON HUB_GATEWAY_ROUTER_CONFIG(enableRouteCache);
COMMENT ON TABLE HUB_GATEWAY_ROUTER_CONFIG IS 'Router配置表 - 存储网关Router级别配置';

CREATE TABLE HUB_GATEWAY_ROUTE_CONFIG (
  tenantId VARCHAR2(32) NOT NULL, -- 租户ID
  routeConfigId VARCHAR2(32) NOT NULL, -- 路由配置ID
  gatewayInstanceId VARCHAR2(32) NOT NULL, -- 关联的网关实例ID
  routeName VARCHAR2(100) NOT NULL, -- 路由名称
  routePath VARCHAR2(200) NOT NULL, -- 路由路径
  allowedMethods VARCHAR2(200), -- 允许的HTTP方法,JSON数组格式["GET","POST"]
  allowedHosts VARCHAR2(500), -- 允许的域名,逗号分隔
  matchType NUMBER(10) DEFAULT 1 NOT NULL, -- 匹配类型(0精确匹配,1前缀匹配,2正则匹配)
  routePriority NUMBER(10) DEFAULT 100 NOT NULL, -- 路由优先级,数值越小优先级越高
  stripPathPrefix VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否剥离路径前缀(N否,Y是)
  rewritePath VARCHAR2(200), -- 重写路径
  enableWebsocket VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否支持WebSocket(N否,Y是)
  timeoutMs NUMBER(10) DEFAULT 30000 NOT NULL, -- 超时时间(毫秒)
  retryCount NUMBER(10) DEFAULT 0 NOT NULL, -- 重试次数
  retryIntervalMs NUMBER(10) DEFAULT 1000 NOT NULL, -- 重试间隔(毫秒)

  -- 服务关联字段，直接关联服务定义表
  serviceDefinitionId VARCHAR2(32), -- 关联的服务定义ID

  -- 日志配置关联字段
  logConfigId VARCHAR2(32), -- 关联的日志配置ID(路由级日志配置)

  -- 路由元数据，用于存储额外配置信息
  routeMetadata CLOB, -- 路由元数据,JSON格式,存储Methods等配置

  -- 预留字段
  reserved1 VARCHAR2(100), -- 预留字段1
  reserved2 VARCHAR2(100), -- 预留字段2
  reserved3 NUMBER(10), -- 预留字段3
  reserved4 NUMBER(10), -- 预留字段4
  reserved5 DATE, -- 预留字段5
  extProperty CLOB, -- 扩展属性,JSON格式

  -- 标准字段
  addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
  addWho VARCHAR2(32) NOT NULL, -- 创建人ID
  editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
  editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
  oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
  currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
  activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动/禁用,Y活动/启用)
  noteText VARCHAR2(500), -- 备注信息

  CONSTRAINT PK_HUB_GATEWAY_ROUTE_CONFIG PRIMARY KEY (tenantId, routeConfigId)
);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_CONFIG_instance ON HUB_GATEWAY_ROUTE_CONFIG(gatewayInstanceId);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_CONFIG_service ON HUB_GATEWAY_ROUTE_CONFIG(serviceDefinitionId);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_CONFIG_log ON HUB_GATEWAY_ROUTE_CONFIG(logConfigId);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_CONFIG_priority ON HUB_GATEWAY_ROUTE_CONFIG(routePriority);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_CONFIG_path ON HUB_GATEWAY_ROUTE_CONFIG(routePath);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_CONFIG_active ON HUB_GATEWAY_ROUTE_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GATEWAY_ROUTE_CONFIG IS '路由定义表 - 存储API路由配置,使用activeFlag统一管理启用状态';

CREATE TABLE HUB_GATEWAY_ROUTE_ASSERTION (
                                             tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                             routeAssertionId VARCHAR2(32) NOT NULL, -- 路由断言ID
                                             routeConfigId VARCHAR2(32) NOT NULL, -- 关联的路由配置ID
                                             assertionName VARCHAR2(100) NOT NULL, -- 断言名称
                                             assertionType VARCHAR2(50) NOT NULL, -- 断言类型(PATH,HEADER,QUERY,COOKIE,IP)
                                             assertionOperator VARCHAR2(20) DEFAULT 'EQUAL' NOT NULL, -- 断言操作符(EQUAL,NOT_EQUAL,CONTAINS,MATCHES等)
                                             fieldName VARCHAR2(100), -- 字段名称(header/query名称)
                                             expectedValue VARCHAR2(500), -- 期望值
                                             patternValue VARCHAR2(500), -- 匹配模式(正则表达式等)
                                             caseSensitive VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否区分大小写(N否,Y是)
                                             assertionOrder NUMBER(10) DEFAULT 0 NOT NULL, -- 断言执行顺序
                                             isRequired VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否必须匹配(N否,Y是)
                                             assertionDesc VARCHAR2(200), -- 断言描述

                                             reserved1 VARCHAR2(100), -- 预留字段1
                                             reserved2 VARCHAR2(100), -- 预留字段2
                                             reserved3 NUMBER(10), -- 预留字段3
                                             reserved4 NUMBER(10), -- 预留字段4
                                             reserved5 DATE, -- 预留字段5
                                             extProperty CLOB, -- 扩展属性,JSON格式

                                             addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                             addWho VARCHAR2(32) NOT NULL, -- 创建人ID
                                             editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                             editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                             oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
                                             currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                             activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                             noteText VARCHAR2(500), -- 备注信息

                                             CONSTRAINT PK_HUB_GATEWAY_ROUTE_ASSERTION PRIMARY KEY (tenantId, routeAssertionId)
);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_ASSERTION_route ON HUB_GATEWAY_ROUTE_ASSERTION(routeConfigId);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_ASSERTION_type ON HUB_GATEWAY_ROUTE_ASSERTION(assertionType);
CREATE INDEX idx_HUB_GATEWAY_ROUTE_ASSERTION_order ON HUB_GATEWAY_ROUTE_ASSERTION(assertionOrder);
COMMENT ON TABLE HUB_GATEWAY_ROUTE_ASSERTION IS '路由断言表 - 存储路由匹配断言规则';

CREATE TABLE HUB_GATEWAY_FILTER_CONFIG (
                                           tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                           filterConfigId VARCHAR2(32) NOT NULL, -- 过滤器配置ID
                                           gatewayInstanceId VARCHAR2(32), -- 网关实例ID(实例级过滤器)
                                           routeConfigId VARCHAR2(32), -- 路由配置ID(路由级过滤器)
                                           filterName VARCHAR2(100) NOT NULL, -- 过滤器名称

    -- 根据FilterType枚举值设计
                                           filterType VARCHAR2(50) NOT NULL, -- 过滤器类型(header,query-param,body,url,method,cookie,response)

    -- 根据FilterAction枚举值设计
                                           filterAction VARCHAR2(50) NOT NULL, -- 过滤器执行时机(pre-routing,post-routing,pre-response)

                                           filterOrder NUMBER(10) DEFAULT 0 NOT NULL, -- 过滤器执行顺序(Priority)
                                           filterConfig CLOB NOT NULL, -- 过滤器具体配置,JSON格式
                                           filterDesc VARCHAR2(200), -- 过滤器描述

    -- 根据FilterConfig结构设计的附属字段
                                           configId VARCHAR2(100), -- 过滤器配置ID(来自FilterConfig.ID)

                                           reserved1 VARCHAR2(100), -- 预留字段1
                                           reserved2 VARCHAR2(100), -- 预留字段2
                                           reserved3 NUMBER(10), -- 预留字段3
                                           reserved4 NUMBER(10), -- 预留字段4
                                           reserved5 DATE, -- 预留字段5
                                           extProperty CLOB, -- 扩展属性,JSON格式

                                           addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                           addWho VARCHAR2(32) NOT NULL, -- 创建人ID
                                           editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                           editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                           oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
                                           currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                           activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动/禁用,Y活动/启用)
                                           noteText VARCHAR2(500), -- 备注信息

                                           CONSTRAINT PK_HUB_GATEWAY_FILTER_CONFIG PRIMARY KEY (tenantId, filterConfigId)
);
CREATE INDEX idx_HUB_GATEWAY_FILTER_CONFIG_instance ON HUB_GATEWAY_FILTER_CONFIG(gatewayInstanceId);
CREATE INDEX idx_HUB_GATEWAY_FILTER_CONFIG_route ON HUB_GATEWAY_FILTER_CONFIG(routeConfigId);
CREATE INDEX idx_HUB_GATEWAY_FILTER_CONFIG_type ON HUB_GATEWAY_FILTER_CONFIG(filterType);
CREATE INDEX idx_HUB_GATEWAY_FILTER_CONFIG_action ON HUB_GATEWAY_FILTER_CONFIG(filterAction);
CREATE INDEX idx_HUB_GATEWAY_FILTER_CONFIG_order ON HUB_GATEWAY_FILTER_CONFIG(filterOrder);
CREATE INDEX idx_HUB_GATEWAY_FILTER_CONFIG_active ON HUB_GATEWAY_FILTER_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GATEWAY_FILTER_CONFIG IS '过滤器配置表 - 根据filter.go逻辑设计,支持7种类型和3种执行时机';

CREATE TABLE HUB_GATEWAY_RATE_LIMIT_CONFIG (
                                               tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                               rateLimitConfigId VARCHAR2(32) NOT NULL, -- 限流配置ID
                                               gatewayInstanceId VARCHAR2(32), -- 网关实例ID(实例级限流)
                                               routeConfigId VARCHAR2(32), -- 路由配置ID(路由级限流)
                                               limitName VARCHAR2(100) NOT NULL, -- 限流规则名称

    -- 限流算法标识（token-bucket,leaky-bucket,sliding-window,fixed-window,none）
                                               algorithm VARCHAR2(50) DEFAULT 'token-bucket' NOT NULL,

    -- 限流键策略（ip,user,path,service,route）
                                               keyStrategy VARCHAR2(50) DEFAULT 'ip' NOT NULL,

    -- 限流速率相关字段
                                               limitRate NUMBER(10) NOT NULL, -- 限流速率(次/秒)
                                               burstCapacity NUMBER(10) DEFAULT 0 NOT NULL, -- 突发容量
                                               timeWindowSeconds NUMBER(10) DEFAULT 1 NOT NULL, -- 时间窗口(秒)
                                               rejectionStatusCode NUMBER(10) DEFAULT 429 NOT NULL, -- 拒绝时的HTTP状态码
                                               rejectionMessage VARCHAR2(200), -- 拒绝时的提示消息
                                               configPriority NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高
                                               customConfig CLOB DEFAULT '{}' NOT NULL, -- 自定义配置,JSON格式

    -- 预留字段
                                               reserved1 VARCHAR2(100), -- 预留字段1
                                               reserved2 VARCHAR2(100), -- 预留字段2
                                               reserved3 NUMBER(10), -- 预留字段3
                                               reserved4 NUMBER(10), -- 预留字段4
                                               reserved5 DATE, -- 预留字段5
                                               extProperty CLOB, -- 扩展属性,JSON格式

    -- 标准字段
                                               addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                               addWho VARCHAR2(32) NOT NULL, -- 创建人ID
                                               editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                               editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                               oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
                                               currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                               activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                               noteText VARCHAR2(500), -- 备注信息

    -- 主键约束
                                               CONSTRAINT PK_HUB_GATEWAY_RATE_LIMIT_CONFIG PRIMARY KEY (tenantId, rateLimitConfigId)
);
CREATE INDEX idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_instance ON HUB_GATEWAY_RATE_LIMIT_CONFIG(gatewayInstanceId);
CREATE INDEX idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_route ON HUB_GATEWAY_RATE_LIMIT_CONFIG(routeConfigId);
CREATE INDEX idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_strategy ON HUB_GATEWAY_RATE_LIMIT_CONFIG(keyStrategy);
CREATE INDEX idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_algorithm ON HUB_GATEWAY_RATE_LIMIT_CONFIG(algorithm);
CREATE INDEX idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_priority ON HUB_GATEWAY_RATE_LIMIT_CONFIG(configPriority);
CREATE INDEX idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_active ON HUB_GATEWAY_RATE_LIMIT_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GATEWAY_RATE_LIMIT_CONFIG IS '限流配置表 - 存储流量限制规则';


CREATE TABLE HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG (
                                                    tenantId        VARCHAR2(32) NOT NULL, -- 租户ID
                                                    circuitBreakerConfigId VARCHAR2(32) NOT NULL, -- 熔断配置ID
                                                    routeConfigId   VARCHAR2(32), -- 路由配置ID(路由级熔断)
                                                    targetServiceId VARCHAR2(32), -- 目标服务ID(服务级熔断)
                                                    breakerName     VARCHAR2(100) NOT NULL, -- 熔断器名称

    -- 熔断Key策略（ip, service, api等）
                                                    keyStrategy     VARCHAR2(50) DEFAULT 'api' NOT NULL,

    -- 阈值配置
                                                    errorRatePercent      NUMBER(10) DEFAULT 50 NOT NULL, -- 错误率阈值(百分比)
                                                    minimumRequests       NUMBER(10) DEFAULT 10 NOT NULL, -- 最小请求数阈值
                                                    halfOpenMaxRequests   NUMBER(10) DEFAULT 3 NOT NULL, -- 半开状态最大请求数
                                                    slowCallThreshold     NUMBER(10) DEFAULT 1000 NOT NULL, -- 慢调用阈值(毫秒)
                                                    slowCallRatePercent   NUMBER(10) DEFAULT 50 NOT NULL, -- 慢调用率阈值(百分比)

    -- 时间配置
                                                    openTimeoutSeconds    NUMBER(10) DEFAULT 60 NOT NULL, -- 熔断器打开持续时间(秒)
                                                    windowSizeSeconds     NUMBER(10) DEFAULT 60 NOT NULL, -- 统计窗口大小(秒)

    -- 错误处理配置
                                                    errorStatusCode       NUMBER(10) DEFAULT 503 NOT NULL, -- 熔断时返回的HTTP状态码
                                                    errorMessage          VARCHAR2(500), -- 熔断时返回的错误信息

    -- 存储配置
                                                    storageType           VARCHAR2(50) DEFAULT 'memory' NOT NULL, -- 存储类型(memory, redis)
                                                    storageConfig         CLOB, -- 存储配置,JSON格式

    -- 优先级 & 预留字段
                                                    configPriority        NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高

                                                    reserved1             VARCHAR2(100), -- 预留字段1
                                                    reserved2             VARCHAR2(100), -- 预留字段2
                                                    reserved3             NUMBER(10), -- 预留字段3
                                                    reserved4             NUMBER(10), -- 预留字段4
                                                    reserved5             DATE, -- 预留字段5
                                                    extProperty           CLOB, -- 扩展属性,JSON格式

    -- 标准字段
                                                    addTime               DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                                    addWho                VARCHAR2(32) NOT NULL, -- 创建人ID
                                                    editTime              DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                                    editWho               VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                                    oprSeqFlag            VARCHAR2(32) NOT NULL, -- 操作序列标识
                                                    currentVersion        NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                                    activeFlag            VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                                    noteText              VARCHAR2(500), -- 备注信息

    -- 主键定义
                                                    CONSTRAINT PK_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG PRIMARY KEY (tenantId, circuitBreakerConfigId)
);
CREATE INDEX idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_route ON HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG(routeConfigId);
CREATE INDEX idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_service ON HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG(targetServiceId);
CREATE INDEX idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_strategy ON HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG(keyStrategy);
CREATE INDEX idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_storage ON HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG(storageType);
CREATE INDEX idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_priority ON HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG(configPriority);
COMMENT ON TABLE HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG IS '熔断配置表 - 根据CircuitBreakerConfig结构设计,支持完整的熔断策略配置';

CREATE TABLE HUB_GATEWAY_AUTH_CONFIG (
                                         tenantId         VARCHAR2(32) NOT NULL, -- 租户ID
                                         authConfigId     VARCHAR2(32) NOT NULL, -- 认证配置ID
                                         gatewayInstanceId VARCHAR2(32), -- 网关实例ID(实例级认证)
                                         routeConfigId    VARCHAR2(32), -- 路由配置ID(路由级认证)
                                         authName         VARCHAR2(100) NOT NULL, -- 认证配置名称
                                         authType         VARCHAR2(50) NOT NULL, -- 认证类型(JWT,API_KEY,OAUTH2,BASIC)
                                         authStrategy     VARCHAR2(50) DEFAULT 'REQUIRED', -- 认证策略(REQUIRED,OPTIONAL,DISABLED)
                                         authConfig       CLOB NOT NULL, -- 认证参数配置,JSON格式
                                         exemptPaths      CLOB, -- 豁免路径列表,JSON数组格式
                                         exemptHeaders    CLOB, -- 豁免请求头列表,JSON数组格式
                                         failureStatusCode NUMBER(10) DEFAULT 401 NOT NULL, -- 认证失败状态码
                                         failureMessage   VARCHAR2(200) DEFAULT '认证失败', -- 认证失败提示消息
                                         configPriority   NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高

                                         reserved1        VARCHAR2(100), -- 预留字段1
                                         reserved2        VARCHAR2(100), -- 预留字段2
                                         reserved3        NUMBER(10), -- 预留字段3
                                         reserved4        NUMBER(10), -- 预留字段4
                                         reserved5        DATE, -- 预留字段5
                                         extProperty      CLOB, -- 扩展属性,JSON格式

                                         addTime          DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                         addWho           VARCHAR2(32) NOT NULL, -- 创建人ID
                                         editTime         DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                         editWho          VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                         oprSeqFlag       VARCHAR2(32) NOT NULL, -- 操作序列标识
                                         currentVersion   NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                         activeFlag       VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                         noteText         VARCHAR2(500), -- 备注信息

                                         CONSTRAINT PK_HUB_GATEWAY_AUTH_CONFIG PRIMARY KEY (tenantId, authConfigId)
);
CREATE INDEX idx_HUB_GATEWAY_AUTH_CONFIG_instance ON HUB_GATEWAY_AUTH_CONFIG(gatewayInstanceId);
CREATE INDEX idx_HUB_GATEWAY_AUTH_CONFIG_route ON HUB_GATEWAY_AUTH_CONFIG(routeConfigId);
CREATE INDEX idx_HUB_GATEWAY_AUTH_CONFIG_type ON HUB_GATEWAY_AUTH_CONFIG(authType);
CREATE INDEX idx_HUB_GATEWAY_AUTH_CONFIG_priority ON HUB_GATEWAY_AUTH_CONFIG(configPriority);
COMMENT ON TABLE HUB_GATEWAY_AUTH_CONFIG IS '认证配置表 - 存储认证相关规则';

CREATE TABLE HUB_GATEWAY_SERVICE_DEFINITION (
                                                tenantId              VARCHAR2(32) NOT NULL, -- 租户ID
                                                serviceDefinitionId   VARCHAR2(32) NOT NULL, -- 服务定义ID
                                                serviceName           VARCHAR2(100) NOT NULL, -- 服务名称
                                                serviceDesc           VARCHAR2(200), -- 服务描述
                                                serviceType           NUMBER(10) DEFAULT 0 NOT NULL, -- 服务类型(0静态配置,1服务发现)

                                                proxyConfigId         VARCHAR2(32) NOT NULL, -- 关联的代理配置ID
                                                loadBalanceStrategy   VARCHAR2(50) DEFAULT 'round-robin' NOT NULL, -- 负载均衡策略(round-robin,random,ip-hash,least-conn,weighted-round-robin,consistent-hash)

                                                discoveryType         VARCHAR2(50), -- 服务发现类型(CONSUL,EUREKA,NACOS等)
                                                discoveryConfig       CLOB, -- 服务发现配置,JSON格式

                                                sessionAffinity       VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用会话亲和性(N否,Y是)
                                                stickySession         VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用粘性会话(N否,Y是)
                                                maxRetries            NUMBER(10) DEFAULT 3 NOT NULL, -- 最大重试次数
                                                retryTimeoutMs        NUMBER(10) DEFAULT 5000 NOT NULL, -- 重试超时时间(毫秒)
                                                enableCircuitBreaker  VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用熔断器(N否,Y是)

                                                healthCheckEnabled    VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否启用健康检查(N否,Y是)
                                                healthCheckPath       VARCHAR2(200) DEFAULT '/health', -- 健康检查路径
                                                healthCheckMethod     VARCHAR2(10) DEFAULT 'GET', -- 健康检查方法
                                                healthCheckIntervalSeconds NUMBER(10) DEFAULT 30, -- 健康检查间隔(秒)
                                                healthCheckTimeoutMs  NUMBER(10) DEFAULT 5000, -- 健康检查超时(毫秒)
                                                healthyThreshold      NUMBER(10) DEFAULT 2, -- 健康阈值
                                                unhealthyThreshold    NUMBER(10) DEFAULT 3, -- 不健康阈值
                                                expectedStatusCodes   VARCHAR2(200) DEFAULT '200', -- 期望的状态码,逗号分隔
                                                healthCheckHeaders    CLOB, -- 健康检查请求头,JSON格式

                                                loadBalancerConfig    CLOB, -- 负载均衡器完整配置,JSON格式
                                                serviceMetadata       CLOB, -- 服务元数据,JSON格式

                                                reserved1             VARCHAR2(100), -- 预留字段1
                                                reserved2             VARCHAR2(100), -- 预留字段2
                                                reserved3             NUMBER(10), -- 预留字段3
                                                reserved4             NUMBER(10), -- 预留字段4
                                                reserved5             DATE, -- 预留字段5
                                                extProperty           CLOB, -- 扩展属性,JSON格式

                                                addTime               DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                                addWho                VARCHAR2(32) NOT NULL, -- 创建人ID
                                                editTime              DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                                editWho               VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                                oprSeqFlag            VARCHAR2(32) NOT NULL, -- 操作序列标识
                                                currentVersion        NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                                activeFlag            VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                                noteText              VARCHAR2(500), -- 备注信息

                                                CONSTRAINT PK_HUB_GATEWAY_SERVICE_DEFINITION PRIMARY KEY (tenantId, serviceDefinitionId)
);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_DEFINITION_name ON HUB_GATEWAY_SERVICE_DEFINITION(serviceName);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_DEFINITION_type ON HUB_GATEWAY_SERVICE_DEFINITION(serviceType);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_DEFINITION_strategy ON HUB_GATEWAY_SERVICE_DEFINITION(loadBalanceStrategy);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_DEFINITION_health ON HUB_GATEWAY_SERVICE_DEFINITION(healthCheckEnabled);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_DEFINITION_proxy ON HUB_GATEWAY_SERVICE_DEFINITION(proxyConfigId);
COMMENT ON TABLE HUB_GATEWAY_SERVICE_DEFINITION IS '服务定义表 - 根据ServiceConfig结构设计,存储完整的服务配置';

CREATE TABLE HUB_GATEWAY_SERVICE_NODE (
                                          tenantId              VARCHAR2(32) NOT NULL, -- 租户ID
                                          serviceNodeId         VARCHAR2(32) NOT NULL, -- 服务节点ID
                                          serviceDefinitionId   VARCHAR2(32) NOT NULL, -- 关联的服务定义ID
                                          nodeId                VARCHAR2(100) NOT NULL, -- 节点标识ID

                                          nodeUrl               VARCHAR2(500) NOT NULL, -- 节点完整URL(来自NodeConfig.URL)
                                          nodeHost              VARCHAR2(100) NOT NULL, -- 节点主机地址(从URL解析)
                                          nodePort              NUMBER(10) NOT NULL, -- 节点端口(从URL解析)
                                          nodeProtocol          VARCHAR2(10) DEFAULT 'HTTP' NOT NULL, -- 节点协议(HTTP,HTTPS)

                                          nodeWeight            NUMBER(10) DEFAULT 100 NOT NULL, -- 节点权重(来自NodeConfig.Weight)
                                          healthStatus          VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 健康状态(N不健康,Y健康)

                                          nodeMetadata          CLOB, -- 节点元数据,JSON格式

                                          nodeStatus            NUMBER(10) DEFAULT 1 NOT NULL, -- 节点运行状态(0下线,1在线,2维护)
                                          lastHealthCheckTime   DATE, -- 最后健康检查时间
                                          healthCheckResult     CLOB, -- 健康检查结果详情

                                          reserved1             VARCHAR2(100), -- 预留字段1
                                          reserved2             VARCHAR2(100), -- 预留字段2
                                          reserved3             NUMBER(10), -- 预留字段3
                                          reserved4             NUMBER(10), -- 预留字段4
                                          reserved5             DATE, -- 预留字段5
                                          extProperty           CLOB, -- 扩展属性,JSON格式

                                          addTime               DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                          addWho                VARCHAR2(32) NOT NULL, -- 创建人ID
                                          editTime              DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                          editWho               VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                          oprSeqFlag            VARCHAR2(32) NOT NULL, -- 操作序列标识
                                          currentVersion        NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                          activeFlag            VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                          noteText              VARCHAR2(500), -- 备注信息

                                          CONSTRAINT PK_HUB_GATEWAY_SERVICE_NODE PRIMARY KEY (tenantId, serviceNodeId)
);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_NODE_service ON HUB_GATEWAY_SERVICE_NODE(serviceDefinitionId);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_NODE_endpoint ON HUB_GATEWAY_SERVICE_NODE(nodeHost, nodePort);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_NODE_health ON HUB_GATEWAY_SERVICE_NODE(healthStatus);
CREATE INDEX idx_HUB_GATEWAY_SERVICE_NODE_status ON HUB_GATEWAY_SERVICE_NODE(nodeStatus);
COMMENT ON TABLE HUB_GATEWAY_SERVICE_NODE IS '服务节点表 - 根据NodeConfig结构设计,存储服务节点实例信息';

CREATE TABLE HUB_GATEWAY_PROXY_CONFIG (
                                          tenantId          VARCHAR2(32) NOT NULL, -- 租户ID
                                          proxyConfigId     VARCHAR2(32) NOT NULL, -- 代理配置ID
                                          gatewayInstanceId VARCHAR2(32) NOT NULL, -- 网关实例ID(代理配置仅支持实例级)
                                          proxyName         VARCHAR2(100) NOT NULL, -- 代理名称

                                          proxyType         VARCHAR2(50) DEFAULT 'http' NOT NULL, -- 代理类型(http,websocket,tcp,udp)

                                          proxyId           VARCHAR2(100), -- 代理ID(来自ProxyConfig.ID)
                                          configPriority    NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高

                                          proxyConfig       CLOB NOT NULL, -- 代理具体配置,JSON格式,根据proxyType存储对应配置
                                          customConfig      CLOB, -- 自定义配置,JSON格式

                                          reserved1         VARCHAR2(100), -- 预留字段1
                                          reserved2         VARCHAR2(100), -- 预留字段2
                                          reserved3         NUMBER(10), -- 预留字段3
                                          reserved4         NUMBER(10), -- 预留字段4
                                          reserved5         DATE, -- 预留字段5
                                          extProperty       CLOB, -- 扩展属性,JSON格式

                                          addTime           DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                          addWho            VARCHAR2(32) NOT NULL, -- 创建人ID
                                          editTime          DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                          editWho           VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                          oprSeqFlag        VARCHAR2(32) NOT NULL, -- 操作序列标识
                                          currentVersion    NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                          activeFlag        VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动/禁用,Y活动/启用)
                                          noteText          VARCHAR2(500), -- 备注信息

                                          CONSTRAINT PK_HUB_GATEWAY_PROXY_CONFIG PRIMARY KEY (tenantId, proxyConfigId)
);
CREATE INDEX idx_HUB_GATEWAY_PROXY_CONFIG_instance ON HUB_GATEWAY_PROXY_CONFIG(gatewayInstanceId);
CREATE INDEX idx_HUB_GATEWAY_PROXY_CONFIG_type ON HUB_GATEWAY_PROXY_CONFIG(proxyType);
CREATE INDEX idx_HUB_GATEWAY_PROXY_CONFIG_priority ON HUB_GATEWAY_PROXY_CONFIG(configPriority);
CREATE INDEX idx_HUB_GATEWAY_PROXY_CONFIG_active ON HUB_GATEWAY_PROXY_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GATEWAY_PROXY_CONFIG IS '代理配置表 - 根据proxy.go逻辑设计,仅支持实例级代理配置';

CREATE TABLE HUB_TIMER_SCHEDULER (
                                     schedulerId           VARCHAR2(32) NOT NULL, -- 调度器ID，主键
                                     tenantId              VARCHAR2(32) NOT NULL, -- 租户ID
                                     schedulerName         VARCHAR2(100) NOT NULL, -- 调度器名称
                                     schedulerInstanceId   VARCHAR2(100) NOT NULL, -- 调度器实例ID，用于集群环境区分

                                     maxWorkers            NUMBER(10) DEFAULT 5 NOT NULL, -- 最大工作线程数
                                     queueSize             NUMBER(10) DEFAULT 100 NOT NULL, -- 任务队列大小
                                     defaultTimeoutSeconds NUMBER(20) DEFAULT 1800 NOT NULL, -- 默认超时时间秒数
                                     defaultRetries        NUMBER(10) DEFAULT 3 NOT NULL, -- 默认重试次数

                                     schedulerStatus       NUMBER(10) DEFAULT 1 NOT NULL, -- 调度器状态(1停止,2运行中,3暂停)
                                     lastStartTime         DATE, -- 最后启动时间
                                     lastStopTime          DATE, -- 最后停止时间

                                     serverName            VARCHAR2(100), -- 服务器名称
                                     serverIp              VARCHAR2(50), -- 服务器IP地址
                                     serverPort            NUMBER(10), -- 服务器端口

                                     totalTaskCount        NUMBER(10) DEFAULT 0 NOT NULL, -- 总任务数
                                     runningTaskCount      NUMBER(10) DEFAULT 0 NOT NULL, -- 运行中任务数
                                     lastHeartbeatTime     DATE, -- 最后心跳时间

                                     addTime               DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                     addWho                VARCHAR2(32) NOT NULL, -- 创建人ID
                                     editTime              DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                     editWho               VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                     oprSeqFlag            VARCHAR2(32) NOT NULL, -- 操作序列标识
                                     currentVersion        NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                     activeFlag            VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                     noteText              VARCHAR2(500), -- 备注信息

                                     reserved1             VARCHAR2(500), -- 预留字段1
                                     reserved2             VARCHAR2(500), -- 预留字段2
                                     reserved3             VARCHAR2(500), -- 预留字段3

                                     CONSTRAINT PK_HUB_TIMER_SCHEDULER PRIMARY KEY (tenantId, schedulerId)
);
CREATE INDEX idx_HUB_TIMER_SCHEDULER_name ON HUB_TIMER_SCHEDULER(schedulerName);
CREATE INDEX idx_HUB_TIMER_SCHEDULER_instanceId ON HUB_TIMER_SCHEDULER(schedulerInstanceId);
CREATE INDEX idx_HUB_TIMER_SCHEDULER_status ON HUB_TIMER_SCHEDULER(schedulerStatus);
CREATE INDEX idx_HUB_TIMER_SCHEDULER_heartbeat ON HUB_TIMER_SCHEDULER(lastHeartbeatTime);
COMMENT ON TABLE HUB_TIMER_SCHEDULER IS '定时任务调度器表 - 存储调度器配置和状态信息';

CREATE TABLE HUB_TIMER_TASK (
                                taskId                VARCHAR2(32) NOT NULL, -- 任务ID，主键
                                tenantId              VARCHAR2(32) NOT NULL, -- 租户ID

                                taskName              VARCHAR2(200) NOT NULL, -- 任务名称
                                taskDescription       VARCHAR2(500), -- 任务描述
                                taskPriority          NUMBER(10) DEFAULT 1 NOT NULL, -- 任务优先级(1低,2普通,3高)
                                schedulerId           VARCHAR2(32), -- 关联的调度器ID
                                schedulerName         VARCHAR2(100), -- 调度器名称（冗余字段）

                                scheduleType          NUMBER(10) NOT NULL, -- 调度类型(1一次性,2固定间隔,3Cron,4延迟执行,5实时执行)
                                cronExpression        VARCHAR2(100), -- Cron表达式（scheduleType=3时必填）
                                intervalSeconds       NUMBER(20), -- 执行间隔秒数（scheduleType=2时必填）
                                delaySeconds          NUMBER(20), -- 延迟秒数（scheduleType=4时必填）
                                startTime             DATE, -- 任务开始时间
                                endTime               DATE, -- 任务结束时间

                                maxRetries            NUMBER(10) DEFAULT 0 NOT NULL, -- 最大重试次数
                                retryIntervalSeconds  NUMBER(20) DEFAULT 60 NOT NULL, -- 重试间隔秒数
                                timeoutSeconds        NUMBER(20) DEFAULT 1800 NOT NULL, -- 执行超时时间秒数
                                taskParams            CLOB, -- 任务参数，JSON格式存储

                                taskStatus            NUMBER(10) DEFAULT 1 NOT NULL, -- 任务状态(1待执行,2运行中,3已完成,4失败,5取消)
                                nextRunTime           DATE, -- 下次执行时间
                                lastRunTime           DATE, -- 上次执行时间
                                runCount              NUMBER(20) DEFAULT 0 NOT NULL, -- 执行总次数
                                successCount          NUMBER(20) DEFAULT 0 NOT NULL, -- 成功次数
                                failureCount          NUMBER(20) DEFAULT 0 NOT NULL, -- 失败次数

                                lastExecutionId       VARCHAR2(32), -- 最后执行ID
                                lastExecutionStartTime DATE, -- 最后执行开始时间
                                lastExecutionEndTime   DATE, -- 最后执行结束时间
                                lastExecutionDurationMs NUMBER(20), -- 最后执行耗时毫秒数
                                lastExecutionStatus    NUMBER(10), -- 最后执行状态
                                lastResultSuccess      VARCHAR2(1), -- 最后执行是否成功(N失败,Y成功)
                                lastErrorMessage       CLOB, -- 最后错误信息
                                lastRetryCount         NUMBER(10), -- 最后重试次数

                                addTime               DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                addWho                VARCHAR2(32) NOT NULL, -- 创建人ID
                                editTime              DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                editWho               VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                oprSeqFlag            VARCHAR2(32) NOT NULL, -- 操作序列标识
                                currentVersion        NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                activeFlag            VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                noteText              VARCHAR2(500), -- 备注信息

                                reserved1             VARCHAR2(500), -- 预留字段1
                                reserved2             VARCHAR2(500), -- 预留字段2
                                reserved3             VARCHAR2(500), -- 预留字段3

                                CONSTRAINT PK_HUB_TIMER_TASK PRIMARY KEY (tenantId, taskId)
);

CREATE INDEX idx_HUB_TIMER_TASK_name ON HUB_TIMER_TASK(taskName);
CREATE INDEX idx_HUB_TIMER_TASK_schedulerId ON HUB_TIMER_TASK(schedulerId);
CREATE INDEX idx_HUB_TIMER_TASK_scheduleType ON HUB_TIMER_TASK(scheduleType);
CREATE INDEX idx_HUB_TIMER_TASK_status ON HUB_TIMER_TASK(taskStatus);
CREATE INDEX idx_HUB_TIMER_TASK_nextRunTime ON HUB_TIMER_TASK(nextRunTime);
CREATE INDEX idx_HUB_TIMER_TASK_lastRunTime ON HUB_TIMER_TASK(lastRunTime);
CREATE INDEX idx_HUB_TIMER_TASK_activeFlag ON HUB_TIMER_TASK(activeFlag);
COMMENT ON TABLE HUB_TIMER_TASK IS '定时任务表 - 合并任务配置、运行时信息和最后执行结果';
CREATE TABLE HUB_TIMER_EXECUTION_LOG (
                                         executionId             VARCHAR2(32) NOT NULL, -- 执行ID，主键
                                         tenantId                VARCHAR2(32) NOT NULL, -- 租户ID
                                         taskId                  VARCHAR2(32) NOT NULL, -- 关联任务ID

                                         taskName                VARCHAR2(200), -- 任务名称（冗余）
                                         schedulerId             VARCHAR2(32), -- 调度器ID（冗余）

                                         executionStartTime      DATE NOT NULL, -- 执行开始时间
                                         executionEndTime        DATE, -- 执行结束时间
                                         executionDurationMs     NUMBER(20), -- 执行耗时毫秒数
                                         executionStatus         NUMBER(10) NOT NULL, -- 执行状态(1待执行,2运行中,3已完成,4失败,5取消)
                                         resultSuccess           VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否成功(N失败,Y成功)

                                         errorMessage              CLOB, -- 错误信息
                                         errorStackTrace           CLOB, -- 错误堆栈信息

                                         retryCount              NUMBER(10) DEFAULT 0 NOT NULL, -- 重试次数
                                         maxRetryCount           NUMBER(10) DEFAULT 0 NOT NULL, -- 最大重试次数

                                         executionParams         CLOB, -- 执行参数，JSON格式
                                         executionResult         CLOB, -- 执行结果，JSON格式

                                         executorServerName      VARCHAR2(100), -- 执行服务器名称
                                         executorServerIp        VARCHAR2(50), -- 执行服务器IP地址

                                         logLevel                VARCHAR2(10), -- 日志级别(DEBUG,INFO,WARN,ERROR)
                                         logMessage              CLOB, -- 日志消息内容
                                         logTimestamp            DATE, -- 日志时间戳

                                         executionPhase          VARCHAR2(50), -- 执行阶段(BEFORE_EXECUTE,EXECUTING,AFTER_EXECUTE,RETRY)
                                         threadName              VARCHAR2(100), -- 执行线程名称
                                         className               VARCHAR2(200), -- 执行类名
                                         methodName              VARCHAR2(100), -- 执行方法名

                                         exceptionClass          VARCHAR2(200), -- 异常类名
                                         exceptionMessage        CLOB, -- 异常消息

                                         addTime                 DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                         addWho                  VARCHAR2(32) NOT NULL, -- 创建人ID
                                         editTime                DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                         editWho                 VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                         oprSeqFlag              VARCHAR2(32) NOT NULL, -- 操作序列标识
                                         currentVersion          NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                         activeFlag              VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N/Y)
                                         noteText                VARCHAR2(500), -- 备注信息

                                         reserved1               VARCHAR2(500), -- 预留字段1
                                         reserved2               VARCHAR2(500), -- 预留字段2
                                         reserved3               VARCHAR2(500), -- 预留字段3

                                         CONSTRAINT PK_HUB_TIMER_EXECUTION_LOG PRIMARY KEY (tenantId, executionId)
);
CREATE INDEX idx_HUB_TIMER_EXECUTION_LOG_taskId ON HUB_TIMER_EXECUTION_LOG(taskId);
CREATE INDEX idx_HUB_TIMER_EXECUTION_LOG_taskName ON HUB_TIMER_EXECUTION_LOG(taskName);
CREATE INDEX idx_HUB_TIMER_EXECUTION_LOG_schedulerId ON HUB_TIMER_EXECUTION_LOG(schedulerId);
CREATE INDEX idx_HUB_TIMER_EXECUTION_LOG_startTime ON HUB_TIMER_EXECUTION_LOG(executionStartTime);
CREATE INDEX idx_HUB_TIMER_EXECUTION_LOG_status ON HUB_TIMER_EXECUTION_LOG(executionStatus);
CREATE INDEX idx_HUB_TIMER_EXECUTION_LOG_success ON HUB_TIMER_EXECUTION_LOG(resultSuccess);
CREATE INDEX idx_HUB_TIMER_EXECUTION_LOG_logLevel ON HUB_TIMER_EXECUTION_LOG(logLevel);
CREATE INDEX idx_HUB_TIMER_EXECUTION_LOG_logTimestamp ON HUB_TIMER_EXECUTION_LOG(logTimestamp);
COMMENT ON TABLE HUB_TIMER_EXECUTION_LOG IS '任务执行日志表 - 合并执行记录和日志信息';