CREATE TABLE `HUB_GATEWAY_INSTANCE` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '网关实例ID',
    `instanceName` VARCHAR(100) NOT NULL COMMENT '实例名称',
  `instanceDesc` VARCHAR(200) DEFAULT NULL COMMENT '实例描述',
  `bindAddress` VARCHAR(100) DEFAULT '0.0.0.0' COMMENT '绑定地址',

  -- HTTP/HTTPS 端口配置
  `httpPort` INT DEFAULT NULL COMMENT 'HTTP监听端口',
  `httpsPort` INT DEFAULT NULL COMMENT 'HTTPS监听端口',
  `tlsEnabled` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用TLS(N否,Y是)',

  -- 证书配置 - 支持文件路径和数据库存储
  `certStorageType` VARCHAR(20) NOT NULL DEFAULT 'FILE' COMMENT '证书存储类型(FILE文件,DATABASE数据库)',
  `certFilePath` VARCHAR(255) DEFAULT NULL COMMENT '证书文件路径',
  `keyFilePath` VARCHAR(255) DEFAULT NULL COMMENT '私钥文件路径',
  `certContent` TEXT DEFAULT NULL COMMENT '证书内容(PEM格式)',
  `keyContent` TEXT DEFAULT NULL COMMENT '私钥内容(PEM格式)',
  `certChainContent` TEXT DEFAULT NULL COMMENT '证书链内容(PEM格式)',
  `certPassword` VARCHAR(255) DEFAULT NULL COMMENT '证书密码(加密存储)',

  -- Go HTTP Server 核心配置
  `maxConnections` INT NOT NULL DEFAULT 10000 COMMENT '最大连接数',
  `readTimeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '读取超时时间(毫秒)',
  `writeTimeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '写入超时时间(毫秒)',
  `idleTimeoutMs` INT NOT NULL DEFAULT 60000 COMMENT '空闲连接超时时间(毫秒)',
  `maxHeaderBytes` INT NOT NULL DEFAULT 1048576 COMMENT '最大请求头字节数(默认1MB)',

  -- 性能和并发配置
  `maxWorkers` INT NOT NULL DEFAULT 1000 COMMENT '最大工作协程数',
  `keepAliveEnabled` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用Keep-Alive(N否,Y是)',
  `tcpKeepAliveEnabled` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用TCP Keep-Alive(N否,Y是)',
  `gracefulShutdownTimeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '优雅关闭超时时间(毫秒)',

  -- TLS安全配置
  `enableHttp2` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用HTTP/2(N否,Y是)',
  `tlsVersion` VARCHAR(10) DEFAULT '1.2' COMMENT 'TLS协议版本(1.0,1.1,1.2,1.3)',
  `tlsCipherSuites` VARCHAR(1000) DEFAULT NULL COMMENT 'TLS密码套件列表,逗号分隔',
  `disableGeneralOptionsHandler` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否禁用默认OPTIONS处理器(N否,Y是)',
  -- 日志配置关联字段
  `logConfigId` VARCHAR(32) DEFAULT NULL COMMENT '关联的日志配置ID',
  `healthStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '健康状态(N不健康,Y健康)',
  `lastHeartbeatTime` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  `instanceMetadata` TEXT DEFAULT NULL COMMENT '实例元数据,JSON格式',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_bind_http` (`bindAddress`, `httpPort`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_bind_https` (`bindAddress`, `httpsPort`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_log` (`logConfigId`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_health` (`healthStatus`),
  INDEX `idx_HUB_GATEWAY_INSTANCE_tls` (`tlsEnabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关实例表 - 记录网关实例基础配置，完整支持Go HTTP Server配置';


CREATE TABLE `HUB_GATEWAY_ROUTER_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `routerConfigId` VARCHAR(32) NOT NULL COMMENT 'Router配置ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '关联的网关实例ID',
  `routerName` VARCHAR(100) NOT NULL COMMENT 'Router名称',
  `routerDesc` VARCHAR(200) DEFAULT NULL COMMENT 'Router描述',
  
  -- Router基础配置
  `defaultPriority` INT NOT NULL DEFAULT 100 COMMENT '默认路由优先级',
  `enableRouteCache` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用路由缓存(N否,Y是)',
  `routeCacheTtlSeconds` INT NOT NULL DEFAULT 300 COMMENT '路由缓存TTL(秒)',
  `maxRoutes` INT DEFAULT 1000 COMMENT '最大路由数量限制',
  `routeMatchTimeout` INT DEFAULT 100 COMMENT '路由匹配超时时间(毫秒)',
  
  -- Router高级配置
  `enableStrictMode` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用严格模式(N否,Y是)',
  `enableMetrics` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用路由指标收集(N否,Y是)',
  `enableTracing` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用链路追踪(N否,Y是)',
  `caseSensitive` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '路径匹配是否区分大小写(N否,Y是)',
  `removeTrailingSlash` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否移除路径尾部斜杠(N否,Y是)',
  
  -- 路由处理配置
  `enableGlobalFilters` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用全局过滤器(N否,Y是)',
  `filterExecutionMode` VARCHAR(20) NOT NULL DEFAULT 'SEQUENTIAL' COMMENT '过滤器执行模式(SEQUENTIAL顺序,PARALLEL并行)',
  `maxFilterChainDepth` INT DEFAULT 50 COMMENT '最大过滤器链深度',
  
  -- 性能优化配置
  `enableRoutePooling` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用路由对象池(N否,Y是)',
  `routePoolSize` INT DEFAULT 100 COMMENT '路由对象池大小',
  `enableAsyncProcessing` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用异步处理(N否,Y是)',
  
  -- 错误处理配置
  `enableFallback` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用降级处理(N否,Y是)',
  `fallbackRoute` VARCHAR(200) DEFAULT NULL COMMENT '降级路由路径',
  `notFoundStatusCode` INT NOT NULL DEFAULT 404 COMMENT '路由未找到时的状态码',
  `notFoundMessage` VARCHAR(200) DEFAULT 'Route not found' COMMENT '路由未找到时的提示消息',
  
  -- 自定义配置
  `routerMetadata` TEXT DEFAULT NULL COMMENT 'Router元数据,JSON格式',
  `customConfig` TEXT DEFAULT NULL COMMENT '自定义配置,JSON格式',
  
  -- 标准数据库字段
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动/禁用,Y活动/启用)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  
  PRIMARY KEY (`tenantId`, `routerConfigId`),
  INDEX `idx_HUB_GATEWAY_ROUTER_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_ROUTER_CONFIG_name` (`routerName`),
  INDEX `idx_HUB_GATEWAY_ROUTER_CONFIG_active` (`activeFlag`),
  INDEX `idx_HUB_GATEWAY_ROUTER_CONFIG_cache` (`enableRouteCache`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Router配置表 - 存储网关Router级别配置';


CREATE TABLE `HUB_GATEWAY_ROUTE_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `routeConfigId` VARCHAR(32) NOT NULL COMMENT '路由配置ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '关联的网关实例ID',
  `routeName` VARCHAR(100) NOT NULL COMMENT '路由名称',
  `routePath` VARCHAR(200) NOT NULL COMMENT '路由路径',
  `allowedMethods` VARCHAR(200) DEFAULT NULL COMMENT '允许的HTTP方法,JSON数组格式["GET","POST"]',
  `allowedHosts` VARCHAR(500) DEFAULT NULL COMMENT '允许的域名,逗号分隔',
  `matchType` INT NOT NULL DEFAULT 1 COMMENT '匹配类型(0精确匹配,1前缀匹配,2正则匹配)',
  `routePriority` INT NOT NULL DEFAULT 100 COMMENT '路由优先级,数值越小优先级越高',
  `stripPathPrefix` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否剥离路径前缀(N否,Y是)',
  `rewritePath` VARCHAR(200) DEFAULT NULL COMMENT '重写路径',
  `enableWebsocket` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否支持WebSocket(N否,Y是)',
  `timeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '超时时间(毫秒)',
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `retryIntervalMs` INT NOT NULL DEFAULT 1000 COMMENT '重试间隔(毫秒)',
  
  -- 服务关联字段，直接关联服务定义表
  `serviceDefinitionId` VARCHAR(32) DEFAULT NULL COMMENT '关联的服务定义ID',
  
  -- 日志配置关联字段
  `logConfigId` VARCHAR(32) DEFAULT NULL COMMENT '关联的日志配置ID(路由级日志配置)',
  
  -- 路由元数据，用于存储额外配置信息
  `routeMetadata` TEXT DEFAULT NULL COMMENT '路由元数据,JSON格式,存储Methods等配置',
  
  -- 注意：使用activeFlag代替enabled字段，保持数据库设计一致性
  -- activeFlag='Y'表示路由启用，activeFlag='N'表示路由禁用
  -- 在代码中将activeFlag映射为enabled字段
  
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动/禁用,Y活动/启用)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `routeConfigId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_service` (`serviceDefinitionId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_log` (`logConfigId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_priority` (`routePriority`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_path` (`routePath`),
  INDEX `idx_HUB_GATEWAY_ROUTE_CONFIG_active` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路由定义表 - 存储API路由配置,使用activeFlag统一管理启用状态';


CREATE TABLE `HUB_GATEWAY_ROUTE_ASSERTION` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `routeAssertionId` VARCHAR(32) NOT NULL COMMENT '路由断言ID',
  `routeConfigId` VARCHAR(32) NOT NULL COMMENT '关联的路由配置ID',
  `assertionName` VARCHAR(100) NOT NULL COMMENT '断言名称',
  `assertionType` VARCHAR(50) NOT NULL COMMENT '断言类型(PATH,HEADER,QUERY,COOKIE,IP)',
  `assertionOperator` VARCHAR(20) NOT NULL DEFAULT 'EQUAL' COMMENT '断言操作符(EQUAL,NOT_EQUAL,CONTAINS,MATCHES等)',
  `fieldName` VARCHAR(100) DEFAULT NULL COMMENT '字段名称(header/query名称)',
  `expectedValue` VARCHAR(500) DEFAULT NULL COMMENT '期望值',
  `patternValue` VARCHAR(500) DEFAULT NULL COMMENT '匹配模式(正则表达式等)',
  `caseSensitive` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否区分大小写(N否,Y是)',
  `assertionOrder` INT NOT NULL DEFAULT 0 COMMENT '断言执行顺序',
  `isRequired` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否必须匹配(N否,Y是)',
  `assertionDesc` VARCHAR(200) DEFAULT NULL COMMENT '断言描述',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `routeAssertionId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_ASSERTION_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_ROUTE_ASSERTION_type` (`assertionType`),
  INDEX `idx_HUB_GATEWAY_ROUTE_ASSERTION_order` (`assertionOrder`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路由断言表 - 存储路由匹配断言规则';


CREATE TABLE `HUB_GATEWAY_FILTER_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `filterConfigId` VARCHAR(32) NOT NULL COMMENT '过滤器配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级过滤器)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级过滤器)',
  `filterName` VARCHAR(100) NOT NULL COMMENT '过滤器名称',
  
  -- 根据FilterType枚举值设计
  `filterType` VARCHAR(50) NOT NULL COMMENT '过滤器类型(header,query-param,body,url,method,cookie,response)',
  
  -- 根据FilterAction枚举值设计
  `filterAction` VARCHAR(50) NOT NULL COMMENT '过滤器执行时机(pre-routing,post-routing,pre-response)',
  
  `filterOrder` INT NOT NULL DEFAULT 0 COMMENT '过滤器执行顺序(Priority)',
  `filterConfig` TEXT NOT NULL COMMENT '过滤器具体配置,JSON格式',
  `filterDesc` VARCHAR(200) DEFAULT NULL COMMENT '过滤器描述',
  
  -- 根据FilterConfig结构设计的附属字段
  `configId` VARCHAR(100) DEFAULT NULL COMMENT '过滤器配置ID(来自FilterConfig.ID)',
  
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动/禁用,Y活动/启用)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `filterConfigId`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_type` (`filterType`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_action` (`filterAction`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_order` (`filterOrder`),
  INDEX `idx_HUB_GATEWAY_FILTER_CONFIG_active` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='过滤器配置表 - 根据filter.go逻辑设计,支持7种类型和3种执行时机';


CREATE TABLE `HUB_GATEWAY_CORS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `corsConfigId` VARCHAR(32) NOT NULL COMMENT 'CORS配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级CORS)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级CORS)',
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称',
  `allowOrigins` TEXT NOT NULL COMMENT '允许的源,JSON数组格式',
  `allowMethods` VARCHAR(200) NOT NULL DEFAULT 'GET,POST,PUT,DELETE,OPTIONS' COMMENT '允许的HTTP方法',
  `allowHeaders` TEXT DEFAULT NULL COMMENT '允许的请求头,JSON数组格式',
  `exposeHeaders` TEXT DEFAULT NULL COMMENT '暴露的响应头,JSON数组格式',
  `allowCredentials` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否允许携带凭证(N否,Y是)',
  `maxAgeSeconds` INT NOT NULL DEFAULT 86400 COMMENT '预检请求缓存时间(秒)',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `corsConfigId`),
  INDEX `idx_HUB_GATEWAY_CORS_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_CORS_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_CORS_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='跨域配置表 - 存储CORS相关配置';


CREATE TABLE `HUB_GATEWAY_RATE_LIMIT_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `rateLimitConfigId` VARCHAR(32) NOT NULL COMMENT '限流配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级限流)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级限流)',
  `limitName` VARCHAR(100) NOT NULL COMMENT '限流规则名称',
  
  -- 修改：统一算法标识格式
  `algorithm` VARCHAR(50) NOT NULL DEFAULT 'token-bucket' COMMENT '限流算法(token-bucket,leaky-bucket,sliding-window,fixed-window,none)',
  
  -- 修改：限流键策略（替代原limitType和keyExpression）
  `keyStrategy` VARCHAR(50) NOT NULL DEFAULT 'ip' COMMENT '限流键策略(ip,user,path,service,route)',
  
  -- 保持原有字段但调整默认值
  `limitRate` INT NOT NULL COMMENT '限流速率(次/秒)',
  `burstCapacity` INT NOT NULL DEFAULT 0 COMMENT '突发容量',
  `timeWindowSeconds` INT NOT NULL DEFAULT 1 COMMENT '时间窗口(秒)',
  `rejectionStatusCode` INT NOT NULL DEFAULT 429 COMMENT '拒绝时的HTTP状态码',
  `rejectionMessage` VARCHAR(200) DEFAULT '请求过于频繁，请稍后再试' COMMENT '拒绝时的提示消息',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `customConfig` TEXT DEFAULT '{}' COMMENT '自定义配置,JSON格式',
  
  -- 保留现有的标准字段
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  
  PRIMARY KEY (`tenantId`, `rateLimitConfigId`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_strategy` (`keyStrategy`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_algorithm` (`algorithm`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_priority` (`configPriority`),
  INDEX `idx_HUB_GATEWAY_RATE_LIMIT_CONFIG_active` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限流配置表 - 存储流量限制规则';


CREATE TABLE `HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `circuitBreakerConfigId` VARCHAR(32) NOT NULL COMMENT '熔断配置ID',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级熔断)',
  `targetServiceId` VARCHAR(32) DEFAULT NULL COMMENT '目标服务ID(服务级熔断)',
  `breakerName` VARCHAR(100) NOT NULL COMMENT '熔断器名称',
  
  -- 根据CircuitBreakerConfig结构设计基础配置
  `keyStrategy` VARCHAR(50) NOT NULL DEFAULT 'api' COMMENT '熔断Key策略(ip,service,api等)',
  
  -- 阈值配置
  `errorRatePercent` INT NOT NULL DEFAULT 50 COMMENT '错误率阈值(百分比)',
  `minimumRequests` INT NOT NULL DEFAULT 10 COMMENT '最小请求数阈值',
  `halfOpenMaxRequests` INT NOT NULL DEFAULT 3 COMMENT '半开状态最大请求数',
  `slowCallThreshold` INT NOT NULL DEFAULT 1000 COMMENT '慢调用阈值(毫秒)',
  `slowCallRatePercent` INT NOT NULL DEFAULT 50 COMMENT '慢调用率阈值(百分比)',
  
  -- 时间配置
  `openTimeoutSeconds` INT NOT NULL DEFAULT 60 COMMENT '熔断器打开持续时间(秒)',
  `windowSizeSeconds` INT NOT NULL DEFAULT 60 COMMENT '统计窗口大小(秒)',
  
  -- 错误处理配置
  `errorStatusCode` INT NOT NULL DEFAULT 503 COMMENT '熔断时返回的HTTP状态码',
  `errorMessage` VARCHAR(500) DEFAULT 'Service temporarily unavailable due to circuit breaker' COMMENT '熔断时返回的错误信息',
  
  -- 存储配置
  `storageType` VARCHAR(50) NOT NULL DEFAULT 'memory' COMMENT '存储类型(memory,redis)',
  `storageConfig` TEXT DEFAULT NULL COMMENT '存储配置,JSON格式',
  
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `circuitBreakerConfigId`),
  INDEX `idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_service` (`targetServiceId`),
  INDEX `idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_strategy` (`keyStrategy`),
  INDEX `idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_storage` (`storageType`),
  INDEX `idx_HUB_GATEWAY_CIRCUIT_BREAKER_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='熔断配置表 - 根据CircuitBreakerConfig结构设计,支持完整的熔断策略配置';


CREATE TABLE `HUB_GATEWAY_AUTH_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `authConfigId` VARCHAR(32) NOT NULL COMMENT '认证配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级认证)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级认证)',
  `authName` VARCHAR(100) NOT NULL COMMENT '认证配置名称',
  `authType` VARCHAR(50) NOT NULL COMMENT '认证类型(JWT,API_KEY,OAUTH2,BASIC)',
  `authStrategy` VARCHAR(50) DEFAULT 'REQUIRED' COMMENT '认证策略(REQUIRED,OPTIONAL,DISABLED)',
  `authConfig` TEXT NOT NULL COMMENT '认证参数配置,JSON格式',
  `exemptPaths` TEXT DEFAULT NULL COMMENT '豁免路径列表,JSON数组格式',
  `exemptHeaders` TEXT DEFAULT NULL COMMENT '豁免请求头列表,JSON数组格式',
  `failureStatusCode` INT NOT NULL DEFAULT 401 COMMENT '认证失败状态码',
  `failureMessage` VARCHAR(200) DEFAULT '认证失败' COMMENT '认证失败提示消息',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `authConfigId`),
  INDEX `idx_HUB_GATEWAY_AUTH_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_AUTH_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GATEWAY_AUTH_CONFIG_type` (`authType`),
  INDEX `idx_HUB_GATEWAY_AUTH_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证配置表 - 存储认证相关规则';


CREATE TABLE `HUB_GATEWAY_SERVICE_DEFINITION` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `serviceDefinitionId` VARCHAR(32) NOT NULL COMMENT '服务定义ID',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称',
  `serviceDesc` VARCHAR(200) DEFAULT NULL COMMENT '服务描述',
  `serviceType` INT NOT NULL DEFAULT 0 COMMENT '服务类型(0静态配置,1服务发现)',
  
  -- 代理配置关联字段
  `proxyConfigId` VARCHAR(32) NOT NULL COMMENT '关联的代理配置ID',
  
  -- 根据ServiceConfig.Strategy字段设计负载均衡策略
  `loadBalanceStrategy` VARCHAR(50) NOT NULL DEFAULT 'round-robin' COMMENT '负载均衡策略(round-robin,random,ip-hash,least-conn,weighted-round-robin,consistent-hash)',
  
  -- 服务发现配置
  `discoveryType` VARCHAR(50) DEFAULT NULL COMMENT '服务发现类型(CONSUL,EUREKA,NACOS等)',
  `discoveryConfig` TEXT DEFAULT NULL COMMENT '服务发现配置,JSON格式',
  
  -- 根据LoadBalancerConfig结构设计负载均衡配置
  `sessionAffinity` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用会话亲和性(N否,Y是)',
  `stickySession` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用粘性会话(N否,Y是)',
  `maxRetries` INT NOT NULL DEFAULT 3 COMMENT '最大重试次数',
  `retryTimeoutMs` INT NOT NULL DEFAULT 5000 COMMENT '重试超时时间(毫秒)',
  `enableCircuitBreaker` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用熔断器(N否,Y是)',
  
  -- 根据HealthConfig结构设计健康检查配置
  `healthCheckEnabled` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用健康检查(N否,Y是)',
  `healthCheckPath` VARCHAR(200) DEFAULT '/health' COMMENT '健康检查路径',
  `healthCheckMethod` VARCHAR(10) DEFAULT 'GET' COMMENT '健康检查方法',
  `healthCheckIntervalSeconds` INT DEFAULT 30 COMMENT '健康检查间隔(秒)',
  `healthCheckTimeoutMs` INT DEFAULT 5000 COMMENT '健康检查超时(毫秒)',
  `healthyThreshold` INT DEFAULT 2 COMMENT '健康阈值',
  `unhealthyThreshold` INT DEFAULT 3 COMMENT '不健康阈值',
  `expectedStatusCodes` VARCHAR(200) DEFAULT '200' COMMENT '期望的状态码,逗号分隔',
  `healthCheckHeaders` TEXT DEFAULT NULL COMMENT '健康检查请求头,JSON格式',
  
  -- 负载均衡器配置(JSON格式存储完整的LoadBalancerConfig)
  `loadBalancerConfig` TEXT DEFAULT NULL COMMENT '负载均衡器完整配置,JSON格式',
  `serviceMetadata` TEXT DEFAULT NULL COMMENT '服务元数据,JSON格式',
  
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `serviceDefinitionId`),
  INDEX `idx_HUB_GATEWAY_SERVICE_DEFINITION_name` (`serviceName`),
  INDEX `idx_HUB_GATEWAY_SERVICE_DEFINITION_type` (`serviceType`),
  INDEX `idx_HUB_GATEWAY_SERVICE_DEFINITION_strategy` (`loadBalanceStrategy`),
  INDEX `idx_HUB_GATEWAY_SERVICE_DEFINITION_health` (`healthCheckEnabled`),
  INDEX `idx_HUB_GATEWAY_SERVICE_DEFINITION_proxy` (`proxyConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务定义表 - 根据ServiceConfig结构设计,存储完整的服务配置';


CREATE TABLE `HUB_GATEWAY_SERVICE_NODE` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `serviceNodeId` VARCHAR(32) NOT NULL COMMENT '服务节点ID',
  `serviceDefinitionId` VARCHAR(32) NOT NULL COMMENT '关联的服务定义ID',
  `nodeId` VARCHAR(100) NOT NULL COMMENT '节点标识ID',
  -- 根据NodeConfig.URL字段设计,分解为host+port+protocol便于查询和管理
  `nodeUrl` VARCHAR(500) NOT NULL COMMENT '节点完整URL(来自NodeConfig.URL)',
  `nodeHost` VARCHAR(100) NOT NULL COMMENT '节点主机地址(从URL解析)',
  `nodePort` INT NOT NULL COMMENT '节点端口(从URL解析)',
  `nodeProtocol` VARCHAR(10) NOT NULL DEFAULT 'HTTP' COMMENT '节点协议(HTTP,HTTPS,从URL解析)',
  
  -- 根据NodeConfig.Weight字段设计
  `nodeWeight` INT NOT NULL DEFAULT 100 COMMENT '节点权重(来自NodeConfig.Weight)',
  
  -- 根据NodeConfig.Health字段设计
  `healthStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '健康状态(N不健康,Y健康,来自NodeConfig.Health)',

  -- 根据NodeConfig.Metadata字段设计
  `nodeMetadata` TEXT DEFAULT NULL COMMENT '节点元数据,JSON格式(来自NodeConfig.Metadata)',
  
  -- 运行时状态字段(非NodeConfig结构,但运维需要)
  `nodeStatus` INT NOT NULL DEFAULT 1 COMMENT '节点运行状态(0下线,1在线,2维护)',
  `lastHealthCheckTime` DATETIME DEFAULT NULL COMMENT '最后健康检查时间',
  `healthCheckResult` TEXT DEFAULT NULL COMMENT '健康检查结果详情',
  
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `serviceNodeId`),
  INDEX `idx_HUB_GATEWAY_SERVICE_NODE_service` (`serviceDefinitionId`),
  INDEX `idx_HUB_GATEWAY_SERVICE_NODE_endpoint` (`nodeHost`, `nodePort`),
  INDEX `idx_HUB_GATEWAY_SERVICE_NODE_health` (`healthStatus`),
  INDEX `idx_HUB_GATEWAY_SERVICE_NODE_status` (`nodeStatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务节点表 - 根据NodeConfig结构设计,存储服务节点实例信息';


CREATE TABLE `HUB_GATEWAY_PROXY_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `proxyConfigId` VARCHAR(32) NOT NULL COMMENT '代理配置ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '网关实例ID(代理配置仅支持实例级)',
  `proxyName` VARCHAR(100) NOT NULL COMMENT '代理名称',
  
  -- 根据ProxyType枚举值设计
  `proxyType` VARCHAR(50) NOT NULL DEFAULT 'http' COMMENT '代理类型(http,websocket,tcp,udp)',
  
  -- 基础配置
  `proxyId` VARCHAR(100) DEFAULT NULL COMMENT '代理ID(来自ProxyConfig.ID)',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  
  -- 通用配置，JSON格式存储不同类型的具体配置
  `proxyConfig` TEXT NOT NULL COMMENT '代理具体配置,JSON格式,根据proxyType存储对应配置',
  `customConfig` TEXT DEFAULT NULL COMMENT '自定义配置,JSON格式',
  
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动/禁用,Y活动/启用)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  PRIMARY KEY (`tenantId`, `proxyConfigId`),
  INDEX `idx_HUB_GATEWAY_PROXY_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GATEWAY_PROXY_CONFIG_type` (`proxyType`),
  INDEX `idx_HUB_GATEWAY_PROXY_CONFIG_priority` (`configPriority`),
  INDEX `idx_HUB_GATEWAY_PROXY_CONFIG_active` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='代理配置表 - 根据proxy.go逻辑设计,仅支持实例级代理配置';

-- =====================================================
-- 定时任务模块新表结构设计
-- 模块前缀: HUB_TIMER
-- 设计说明：
-- 1. 合并任务配置、运行时信息和最后执行结果到一个表
-- 2. 历史执行记录单独存储
-- 3. 简化表结构，减少关联查询
-- =====================================================

-- 1. 调度器配置表 - 存储调度器实例的配置信息
CREATE TABLE `HUB_TIMER_SCHEDULER` (
  `schedulerId` VARCHAR(32) NOT NULL COMMENT '调度器ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `schedulerName` VARCHAR(100) NOT NULL COMMENT '调度器名称',
  `schedulerInstanceId` VARCHAR(100) DEFAULT NULL COMMENT '调度器实例ID，用于集群环境区分',
  
  -- 调度器配置
  `maxWorkers` INT NOT NULL DEFAULT 5 COMMENT '最大工作线程数',
  `queueSize` INT NOT NULL DEFAULT 100 COMMENT '任务队列大小',
  `defaultTimeoutSeconds` BIGINT NOT NULL DEFAULT 1800 COMMENT '默认超时时间秒数',
  `defaultRetries` INT NOT NULL DEFAULT 3 COMMENT '默认重试次数',
  
  -- 调度器状态
  `schedulerStatus` INT NOT NULL DEFAULT 1 COMMENT '调度器状态(1停止,2运行中,3暂停)',
  `lastStartTime` DATETIME DEFAULT NULL COMMENT '最后启动时间',
  `lastStopTime` DATETIME DEFAULT NULL COMMENT '最后停止时间',
  
  -- 服务器信息
  `serverName` VARCHAR(100) DEFAULT NULL COMMENT '服务器名称',
  `serverIp` VARCHAR(50) DEFAULT NULL COMMENT '服务器IP地址',
  `serverPort` INT DEFAULT NULL COMMENT '服务器端口',
  
  -- 监控信息
  `totalTaskCount` INT NOT NULL DEFAULT 0 COMMENT '总任务数',
  `runningTaskCount` INT NOT NULL DEFAULT 0 COMMENT '运行中任务数',
  `lastHeartbeatTime` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  
  PRIMARY KEY (`tenantId`, `schedulerId`),
  KEY `idx_HUB_TIMER_SCHEDULER_name` (`schedulerName`),
  KEY `idx_HUB_TIMER_SCHEDULER_instanceId` (`schedulerInstanceId`),
  KEY `idx_HUB_TIMER_SCHEDULER_status` (`schedulerStatus`),
  KEY `idx_HUB_TIMER_SCHEDULER_heartbeat` (`lastHeartbeatTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务调度器表 - 存储调度器配置和状态信息';

-- 2. 任务表 - 合并配置、运行时信息和最后执行结果
CREATE TABLE `HUB_TIMER_TASK` (
  -- 主键信息
  `taskId` VARCHAR(32) NOT NULL COMMENT '任务ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  
  -- 任务配置信息
  `taskName` VARCHAR(200) NOT NULL COMMENT '任务名称',
  `taskDescription` VARCHAR(500) DEFAULT NULL COMMENT '任务描述',
  `taskPriority` INT NOT NULL DEFAULT 1 COMMENT '任务优先级(1低优先级,2普通优先级,3高优先级)',
  `schedulerId` VARCHAR(32) DEFAULT NULL COMMENT '关联的调度器ID',
  `schedulerName` VARCHAR(100) DEFAULT NULL COMMENT '调度器名称（冗余字段，便于查询显示）',
  
  -- 调度配置
  `scheduleType` INT NOT NULL COMMENT '调度类型(1一次性执行,2固定间隔,3Cron表达式,4延迟执行,5实时执行)',
  `cronExpression` VARCHAR(100) DEFAULT NULL COMMENT 'Cron表达式，scheduleType=3时必填',
  `intervalSeconds` BIGINT DEFAULT NULL COMMENT '执行间隔秒数，scheduleType=2时必填',
  `delaySeconds` BIGINT DEFAULT NULL COMMENT '延迟秒数，scheduleType=4时必填',
  `startTime` DATETIME DEFAULT NULL COMMENT '任务开始时间',
  `endTime` DATETIME DEFAULT NULL COMMENT '任务结束时间',
  
  -- 执行配置
  `maxRetries` INT NOT NULL DEFAULT 0 COMMENT '最大重试次数',
  `retryIntervalSeconds` BIGINT NOT NULL DEFAULT 60 COMMENT '重试间隔秒数',
  `timeoutSeconds` BIGINT NOT NULL DEFAULT 1800 COMMENT '执行超时时间秒数',
  `taskParams` TEXT DEFAULT NULL COMMENT '任务参数，JSON格式存储',
  
  -- 任务执行器配置 - 关联到具体工具配置
  `executorType` VARCHAR(50) DEFAULT NULL COMMENT '执行器类型(BUILTIN内置,SFTP文件传输,SSH远程执行,DATABASE数据库,HTTP接口调用等)',
  `toolConfigId` VARCHAR(32) DEFAULT NULL COMMENT '关联的工具配置ID（如SFTP配置ID、数据库配置ID等）',
  `toolConfigName` VARCHAR(100) DEFAULT NULL COMMENT '工具配置名称（冗余字段，便于显示）',
  `operationType` VARCHAR(100) DEFAULT NULL COMMENT '执行操作类型（如文件上传、下载、SQL执行、接口调用等）',
  `operationConfig` TEXT DEFAULT NULL COMMENT '操作参数配置，JSON格式存储具体操作的参数',
  
  -- 运行时状态
  `taskStatus` INT NOT NULL DEFAULT 1 COMMENT '任务状态(1待执行,2运行中,3已完成,4执行失败,5已取消)',
  `nextRunTime` DATETIME DEFAULT NULL COMMENT '下次执行时间',
  `lastRunTime` DATETIME DEFAULT NULL COMMENT '上次执行时间',
  `runCount` BIGINT NOT NULL DEFAULT 0 COMMENT '执行总次数',
  `successCount` BIGINT NOT NULL DEFAULT 0 COMMENT '成功次数',
  `failureCount` BIGINT NOT NULL DEFAULT 0 COMMENT '失败次数',
  
  -- 最后执行结果
  `lastExecutionId` VARCHAR(32) DEFAULT NULL COMMENT '最后执行ID',
  `lastExecutionStartTime` DATETIME DEFAULT NULL COMMENT '最后执行开始时间',
  `lastExecutionEndTime` DATETIME DEFAULT NULL COMMENT '最后执行结束时间',
  `lastExecutionDurationMs` BIGINT DEFAULT NULL COMMENT '最后执行耗时毫秒数',
  `lastExecutionStatus` INT DEFAULT NULL COMMENT '最后执行状态',
  `lastResultSuccess` VARCHAR(1) DEFAULT NULL COMMENT '最后执行是否成功(N失败,Y成功)',
  `lastErrorMessage` TEXT DEFAULT NULL COMMENT '最后错误信息',
  `lastRetryCount` INT DEFAULT NULL COMMENT '最后重试次数',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  
  PRIMARY KEY (`tenantId`, `taskId`),
  KEY `idx_HUB_TIMER_TASK_name` (`taskName`),
  KEY `idx_HUB_TIMER_TASK_schedulerId` (`schedulerId`),
  KEY `idx_HUB_TIMER_TASK_scheduleType` (`scheduleType`),
  KEY `idx_HUB_TIMER_TASK_status` (`taskStatus`),
  KEY `idx_HUB_TIMER_TASK_nextRunTime` (`nextRunTime`),
  KEY `idx_HUB_TIMER_TASK_lastRunTime` (`lastRunTime`),
  KEY `idx_HUB_TIMER_TASK_activeFlag` (`activeFlag`),
  KEY `idx_HUB_TIMER_TASK_executorType` (`executorType`),
  KEY `idx_HUB_TIMER_TASK_toolConfigId` (`toolConfigId`),
  KEY `idx_HUB_TIMER_TASK_operationType` (`operationType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表 - 合并任务配置、运行时信息和最后执行结果';

-- 3. 任务执行历史表 - 存储所有执行记录
-- 创建新的合并后的执行日志表
CREATE TABLE `HUB_TIMER_EXECUTION_LOG` (
  -- 主键信息
  `executionId` VARCHAR(32) NOT NULL COMMENT '执行ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `taskId` VARCHAR(32) NOT NULL COMMENT '关联任务ID',
  
  -- 任务信息（冗余）
  `taskName` VARCHAR(200) DEFAULT NULL COMMENT '任务名称',
  `schedulerId` VARCHAR(32) DEFAULT NULL COMMENT '调度器ID',
  
  -- 执行信息
  `executionStartTime` DATETIME NOT NULL COMMENT '执行开始时间',
  `executionEndTime` DATETIME DEFAULT NULL COMMENT '执行结束时间',
  `executionDurationMs` BIGINT DEFAULT NULL COMMENT '执行耗时毫秒数',
  `executionStatus` INT NOT NULL COMMENT '执行状态(1待执行,2运行中,3已完成,4执行失败,5已取消)',
  `resultSuccess` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '执行是否成功(N失败,Y成功)',
  
  -- 错误信息
  `errorMessage` TEXT DEFAULT NULL COMMENT '错误信息',
  `errorStackTrace` TEXT DEFAULT NULL COMMENT '错误堆栈信息',
  
  -- 重试信息
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `maxRetryCount` INT NOT NULL DEFAULT 0 COMMENT '最大重试次数',
  
  -- 参数和结果
  `executionParams` TEXT DEFAULT NULL COMMENT '执行参数，JSON格式',
  `executionResult` TEXT DEFAULT NULL COMMENT '执行结果，JSON格式',
  
  -- 执行环境
  `executorServerName` VARCHAR(100) DEFAULT NULL COMMENT '执行服务器名称',
  `executorServerIp` VARCHAR(50) DEFAULT NULL COMMENT '执行服务器IP地址',
  
  -- 日志信息
  `logLevel` VARCHAR(10) DEFAULT NULL COMMENT '日志级别(DEBUG,INFO,WARN,ERROR)',
  `logMessage` TEXT DEFAULT NULL COMMENT '日志消息内容',
  `logTimestamp` DATETIME DEFAULT NULL COMMENT '日志时间戳',
  
  -- 执行上下文
  `executionPhase` VARCHAR(50) DEFAULT NULL COMMENT '执行阶段(BEFORE_EXECUTE,EXECUTING,AFTER_EXECUTE,RETRY)',
  `threadName` VARCHAR(100) DEFAULT NULL COMMENT '执行线程名称',
  `className` VARCHAR(200) DEFAULT NULL COMMENT '执行类名',
  `methodName` VARCHAR(100) DEFAULT NULL COMMENT '执行方法名',
  
  -- 异常信息
  `exceptionClass` VARCHAR(200) DEFAULT NULL COMMENT '异常类名',
  `exceptionMessage` TEXT DEFAULT NULL COMMENT '异常消息',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  
  PRIMARY KEY (`tenantId`, `executionId`),
  KEY `idx_HUB_TIMER_EXECUTION_LOG_taskId` (`taskId`),
  KEY `idx_HUB_TIMER_EXECUTION_LOG_taskName` (`taskName`),
  KEY `idx_HUB_TIMER_EXECUTION_LOG_schedulerId` (`schedulerId`),
  KEY `idx_HUB_TIMER_EXECUTION_LOG_startTime` (`executionStartTime`),
  KEY `idx_HUB_TIMER_EXECUTION_LOG_status` (`executionStatus`),
  KEY `idx_HUB_TIMER_EXECUTION_LOG_success` (`resultSuccess`),
  KEY `idx_HUB_TIMER_EXECUTION_LOG_logLevel` (`logLevel`),
  KEY `idx_HUB_TIMER_EXECUTION_LOG_logTimestamp` (`logTimestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行日志表 - 合并执行记录和日志信息';

-- ===================================================
-- 通用配置工具表设计
-- 说明: 用于管理系统中各种工具的配置信息
-- ===================================================

-- 1. 工具配置主表
CREATE TABLE `HUB_TOOL_CONFIG` (
  `toolConfigId` VARCHAR(32) NOT NULL COMMENT '工具配置ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  
  -- 工具基础信息
  `toolName` VARCHAR(100) NOT NULL COMMENT '工具名称，如SFTP、SSH、FTP等',
  `toolType` VARCHAR(50) NOT NULL COMMENT '工具类型，如transfer、database、monitor等',
  `toolVersion` VARCHAR(20) DEFAULT NULL COMMENT '工具版本号',
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称，用于区分同一工具的不同配置',
  `configDescription` VARCHAR(500) DEFAULT NULL COMMENT '配置描述信息',
  
  -- 分组信息
  `configGroupId` VARCHAR(32) DEFAULT NULL COMMENT '配置分组ID',
  `configGroupName` VARCHAR(100) DEFAULT NULL COMMENT '配置分组名称',
  
  -- 连接配置
  `hostAddress` VARCHAR(255) DEFAULT NULL COMMENT '主机地址或域名',
  `portNumber` INT DEFAULT NULL COMMENT '端口号',
  `protocolType` VARCHAR(20) DEFAULT NULL COMMENT '协议类型，如TCP、UDP、HTTP等',
  
  -- 认证配置
  `authType` VARCHAR(50) DEFAULT NULL COMMENT '认证类型，如password、publickey、oauth等',
  `userName` VARCHAR(100) DEFAULT NULL COMMENT '用户名',
  `passwordEncrypted` VARCHAR(500) DEFAULT NULL COMMENT '加密后的密码',
  `keyFilePath` VARCHAR(500) DEFAULT NULL COMMENT '密钥文件路径',
  `keyFileContent` TEXT DEFAULT NULL COMMENT '密钥文件内容，加密存储',
  
  -- 配置参数
  `configParameters` TEXT DEFAULT NULL COMMENT '配置参数，JSON格式存储',
  `environmentVariables` TEXT DEFAULT NULL COMMENT '环境变量配置，JSON格式存储',
  `customSettings` TEXT DEFAULT NULL COMMENT '自定义设置，JSON格式存储',
  
  -- 状态和控制
  `configStatus` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '配置状态(N禁用,Y启用)',
  `defaultFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否为默认配置(N否,Y是)',
  `priorityLevel` INT DEFAULT 100 COMMENT '优先级，数值越小优先级越高',
  
  -- 安全和加密
  `encryptionType` VARCHAR(50) DEFAULT NULL COMMENT '加密类型，如AES256、RSA等',
  `encryptionKey` VARCHAR(100) DEFAULT NULL COMMENT '加密密钥标识',
  
  -- 标准字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  
  PRIMARY KEY (`tenantId`, `toolConfigId`),
  KEY `idx_HUB_TOOL_CONFIG_toolName` (`toolName`),
  KEY `idx_HUB_TOOL_CONFIG_toolType` (`toolType`),
  KEY `idx_HUB_TOOL_CONFIG_configName` (`configName`),
  KEY `idx_HUB_TOOL_CONFIG_configGroupId` (`configGroupId`),
  KEY `idx_HUB_TOOL_CONFIG_configStatus` (`configStatus`),
  KEY `idx_HUB_TOOL_CONFIG_defaultFlag` (`defaultFlag`),
  KEY `idx_HUB_TOOL_CONFIG_activeFlag` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工具配置主表 - 存储各种工具的基础配置信息';

-- 2. 工具配置分组表
CREATE TABLE `HUB_TOOL_CONFIG_GROUP` (
  `configGroupId` VARCHAR(32) NOT NULL COMMENT '配置分组ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  
  -- 分组信息
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称',
  `groupDescription` VARCHAR(500) DEFAULT NULL COMMENT '分组描述',
  `parentGroupId` VARCHAR(32) DEFAULT NULL COMMENT '父分组ID，支持层级结构',
  `groupLevel` INT DEFAULT 1 COMMENT '分组层级，从1开始',
  `groupPath` VARCHAR(500) DEFAULT NULL COMMENT '分组路径，如/root/parent/child',
  
  -- 分组属性
  `groupType` VARCHAR(50) DEFAULT NULL COMMENT '分组类型，如environment、project、department',
  `sortOrder` INT DEFAULT 100 COMMENT '排序顺序，数值越小越靠前',
  `groupIcon` VARCHAR(100) DEFAULT NULL COMMENT '分组图标',
  `groupColor` VARCHAR(20) DEFAULT NULL COMMENT '分组颜色代码',
  
  -- 权限控制
  `accessLevel` VARCHAR(20) DEFAULT 'private' COMMENT '访问级别，如private、public、restricted',
  `allowedUsers` TEXT DEFAULT NULL COMMENT '允许访问的用户列表，JSON格式',
  `allowedRoles` TEXT DEFAULT NULL COMMENT '允许访问的角色列表，JSON格式',
  
  -- 标准字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  
  PRIMARY KEY (`tenantId`, `configGroupId`),
  KEY `idx_HUB_TOOL_CONFIG_GROUP_groupName` (`groupName`),
  KEY `idx_HUB_TOOL_CONFIG_GROUP_parentGroupId` (`parentGroupId`),
  KEY `idx_HUB_TOOL_CONFIG_GROUP_groupType` (`groupType`),
  KEY `idx_HUB_TOOL_CONFIG_GROUP_sortOrder` (`sortOrder`),
  KEY `idx_HUB_TOOL_CONFIG_GROUP_activeFlag` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工具配置分组表 - 用于对工具配置进行分组管理';