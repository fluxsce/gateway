CREATE TABLE HUB_USER (
    userId          VARCHAR(32)   NOT NULL COMMENT '用户ID，联合主键',
    tenantId        VARCHAR(32)   NOT NULL COMMENT '租户ID，联合主键',
    userName        VARCHAR(50)   NOT NULL COMMENT '用户名，登录账号',
    password        VARCHAR(128)  NOT NULL COMMENT '密码，加密存储',
    realName        VARCHAR(50)   NOT NULL COMMENT '真实姓名',
    deptId          VARCHAR(32)   NOT NULL COMMENT '所属部门ID',
    email           VARCHAR(255)  NULL     COMMENT '电子邮箱',
    mobile          VARCHAR(20)   NULL     COMMENT '手机号码',
    avatar          LONGTEXT      NULL     COMMENT '头像URL或Base64数据',
    gender          INT           NULL     DEFAULT 0 COMMENT '性别：1-男，2-女，0-未知',
    statusFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y' COMMENT '状态：Y-启用，N-禁用',
    deptAdminFlag   VARCHAR(1)    NOT NULL DEFAULT 'N' COMMENT '是否部门管理员：Y-是，N-否',
    tenantAdminFlag VARCHAR(1)    NOT NULL DEFAULT 'N' COMMENT '是否租户管理员：Y-是，N-否',
    userExpireDate  DATETIME      NOT NULL COMMENT '用户过期时间',
    lastLoginTime   DATETIME      NULL     COMMENT '最后登录时间',
    lastLoginIp     VARCHAR(128)  NULL     COMMENT '最后登录IP',
    pwdUpdateTime   DATETIME      NULL     COMMENT '密码最后更新时间',
    pwdErrorCount   INT           NOT NULL DEFAULT 0 COMMENT '密码错误次数',
    lockTime        DATETIME      NULL     COMMENT '账号锁定时间',
    addTime         DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho          VARCHAR(32)   NOT NULL COMMENT '创建人',
    editTime        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho         VARCHAR(32)   NOT NULL COMMENT '修改人',
    oprSeqFlag      VARCHAR(32)   NOT NULL COMMENT '操作序列标识',
    currentVersion  INT           NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    noteText        TEXT          NULL     COMMENT '备注信息',
    extProperty     TEXT          DEFAULT NULL COMMENT '扩展属性，JSON格式',
    reserved1       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段1',
    reserved2       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段2',
    reserved3       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段3',
    reserved4       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段4',
    reserved5       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段5',
    reserved6       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段6',
    reserved7       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段7',
    reserved8       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段8',
    reserved9       VARCHAR(500)  DEFAULT NULL COMMENT '预留字段9',
    reserved10      VARCHAR(500)  DEFAULT NULL COMMENT '预留字段10',
    PRIMARY KEY (userId, tenantId),
    INDEX UK_USER_NAME_TENANT (userName, tenantId), -- 普通索引代替 UNIQUE KEY
    INDEX IDX_USER_TENANT (tenantId),
    INDEX IDX_USER_DEPT (deptId),
    INDEX IDX_USER_STATUS (statusFlag),
    INDEX IDX_USER_EMAIL (email),
    INDEX IDX_USER_MOBILE (mobile)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';

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
    'admin',                            -- userId
    'default',                          -- tenantId
    'admin',                            -- userName
    '123456',                      -- password（使用 MySQL 内置 MD5 加密）
    '系统管理员',                         -- realName
    'D00000001',                        -- deptId
    'admin@example.com',                -- email
    '13800000000',                      -- mobile
    'https://example.com/avatar.png',   -- avatar
    1,                                  -- gender (1:男)
    'Y',                                -- statusFlag
    'N',                                -- deptAdminFlag
    'Y',                                -- tenantAdminFlag
    NOW() + INTERVAL 5 YEAR,            -- userExpireDate（5年后过期）
    'SEQFLAG_001',                      -- oprSeqFlag
    1,                                  -- currentVersion
    'Y',                                -- activeFlag
		'system',
		'system',
    '系统初始化管理员账号'              -- noteText
);

-- =====================================================
-- ALTER 变更语句：用户头像字段类型调整
-- 变更日期：2025-10-10
-- 变更原因：支持存储Base64编码的图片数据
-- 兼容性：向后兼容，现有URL数据不受影响
-- =====================================================
ALTER TABLE HUB_USER MODIFY COLUMN avatar LONGTEXT NULL COMMENT '头像URL或Base64数据';

CREATE TABLE HUB_LOGIN_LOG (
    logId           VARCHAR(32)   NOT NULL COMMENT '日志ID，主键',
    userId          VARCHAR(32)   NOT NULL COMMENT '用户ID',
    tenantId        VARCHAR(32)   NOT NULL COMMENT '租户ID',
    userName        VARCHAR(50)   NOT NULL COMMENT '用户名',
    loginTime       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    loginIp         VARCHAR(128)  NOT NULL COMMENT '登录IP',
    loginLocation   VARCHAR(255)  NULL     COMMENT '登录地点',
    loginType       INT           NOT NULL DEFAULT 1 COMMENT '登录类型：1-用户名密码，2-验证码，3-第三方',
    deviceType      VARCHAR(50)   NULL     COMMENT '设备类型',
    deviceInfo      TEXT          NULL     COMMENT '设备信息',
    browserInfo     TEXT          NULL     COMMENT '浏览器信息',
    osInfo          VARCHAR(255)  NULL     COMMENT '操作系统信息',
    loginStatus     VARCHAR(1)    NOT NULL DEFAULT 'N' COMMENT '登录状态：Y-成功，N-失败',
    logoutTime      DATETIME      NULL     COMMENT '登出时间',
    sessionDuration INT           NULL     COMMENT '会话持续时长(秒)',
    failReason      TEXT          NULL     COMMENT '失败原因',
    addTime         DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    addWho          VARCHAR(32)   NOT NULL COMMENT '创建人',
    editTime        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    editWho         VARCHAR(32)   NOT NULL COMMENT '修改人',
    oprSeqFlag      VARCHAR(32)   NOT NULL COMMENT '操作序列标识',
    currentVersion  INT           NOT NULL DEFAULT 1 COMMENT '当前版本号',
    activeFlag      VARCHAR(1)    NOT NULL DEFAULT 'Y' COMMENT '活动状态标记：Y-活动，N-非活动',
    PRIMARY KEY (logId),
    INDEX IDX_LOGIN_USER (userId),
    INDEX IDX_LOGIN_TIME (loginTime),
    INDEX IDX_LOGIN_TENANT (tenantId),
    INDEX IDX_LOGIN_STATUS (loginStatus),
    INDEX IDX_LOGIN_TYPE (loginType)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户登录日志表';

CREATE TABLE `HUB_GW_INSTANCE` (
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
  INDEX `IDX_GW_INST_BIND_HTTP` (`bindAddress`, `httpPort`),
  INDEX `IDX_GW_INST_BIND_HTTPS` (`bindAddress`, `httpsPort`),
  INDEX `IDX_GW_INST_LOG` (`logConfigId`),
  INDEX `IDX_GW_INST_HEALTH` (`healthStatus`),
  INDEX `IDX_GW_INST_TLS` (`tlsEnabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关实例表 - 记录网关实例基础配置，完整支持Go HTTP Server配置';


CREATE TABLE `HUB_GW_ROUTER_CONFIG` (
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
  INDEX `IDX_GW_ROUTER_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_ROUTER_NAME` (`routerName`),
  INDEX `IDX_GW_ROUTER_ACTIVE` (`activeFlag`),
  INDEX `IDX_GW_ROUTER_CACHE` (`enableRouteCache`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Router配置表 - 存储网关Router级别配置';


CREATE TABLE `HUB_GW_ROUTE_CONFIG` (
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
  INDEX `IDX_GW_ROUTE_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_ROUTE_SERVICE` (`serviceDefinitionId`),
  INDEX `IDX_GW_ROUTE_LOG` (`logConfigId`),
  INDEX `IDX_GW_ROUTE_PRIORITY` (`routePriority`),
  INDEX `IDX_GW_ROUTE_PATH` (`routePath`),
  INDEX `IDX_GW_ROUTE_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路由定义表 - 存储API路由配置,使用activeFlag统一管理启用状态';


CREATE TABLE `HUB_GW_ROUTE_ASSERTION` (
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
  INDEX `IDX_GW_ASSERT_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_ASSERT_TYPE` (`assertionType`),
  INDEX `IDX_GW_ASSERT_ORDER` (`assertionOrder`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路由断言表 - 存储路由匹配断言规则';


CREATE TABLE `HUB_GW_FILTER_CONFIG` (
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
  INDEX `IDX_GW_FILTER_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_FILTER_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_FILTER_TYPE` (`filterType`),
  INDEX `IDX_GW_FILTER_ACTION` (`filterAction`),
  INDEX `IDX_GW_FILTER_ORDER` (`filterOrder`),
  INDEX `IDX_GW_FILTER_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='过滤器配置表 - 根据filter.go逻辑设计,支持7种类型和3种执行时机';


CREATE TABLE `HUB_GW_CORS_CONFIG` (
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
  INDEX `IDX_GW_CORS_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_CORS_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_CORS_PRIORITY` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='跨域配置表 - 存储CORS相关配置';


CREATE TABLE `HUB_GW_RATE_LIMIT_CONFIG` (
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
  `customConfig` TEXT DEFAULT NULL COMMENT '自定义配置,JSON格式',
  
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
  INDEX `IDX_GW_RATE_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_RATE_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_RATE_STRATEGY` (`keyStrategy`),
  INDEX `IDX_GW_RATE_ALGORITHM` (`algorithm`),
  INDEX `IDX_GW_RATE_PRIORITY` (`configPriority`),
  INDEX `IDX_GW_RATE_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限流配置表 - 存储流量限制规则';


CREATE TABLE `HUB_GW_CIRCUIT_BREAKER_CONFIG` (
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
  INDEX `IDX_GW_CB_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_CB_SERVICE` (`targetServiceId`),
  INDEX `IDX_GW_CB_STRATEGY` (`keyStrategy`),
  INDEX `IDX_GW_CB_STORAGE` (`storageType`),
  INDEX `IDX_GW_CB_PRIORITY` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='熔断配置表 - 根据CircuitBreakerConfig结构设计,支持完整的熔断策略配置';


CREATE TABLE `HUB_GW_AUTH_CONFIG` (
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
  INDEX `IDX_GW_AUTH_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_AUTH_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_AUTH_TYPE` (`authType`),
  INDEX `IDX_GW_AUTH_PRIORITY` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证配置表 - 存储认证相关规则';


CREATE TABLE `HUB_GW_SERVICE_DEFINITION` (
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
  INDEX `IDX_GW_SVC_NAME` (`serviceName`),
  INDEX `IDX_GW_SVC_TYPE` (`serviceType`),
  INDEX `IDX_GW_SVC_STRATEGY` (`loadBalanceStrategy`),
  INDEX `IDX_GW_SVC_HEALTH` (`healthCheckEnabled`),
  INDEX `IDX_GW_SVC_PROXY` (`proxyConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务定义表 - 根据ServiceConfig结构设计,存储完整的服务配置';


CREATE TABLE `HUB_GW_SERVICE_NODE` (
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
  INDEX `IDX_GW_NODE_SERVICE` (`serviceDefinitionId`),
  INDEX `IDX_GW_NODE_ENDPOINT` (`nodeHost`, `nodePort`),
  INDEX `IDX_GW_NODE_HEALTH` (`healthStatus`),
  INDEX `IDX_GW_NODE_STATUS` (`nodeStatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务节点表 - 根据NodeConfig结构设计,存储服务节点实例信息';


CREATE TABLE `HUB_GW_PROXY_CONFIG` (
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
  INDEX `IDX_GW_PROXY_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_PROXY_TYPE` (`proxyType`),
  INDEX `IDX_GW_PROXY_PRIORITY` (`configPriority`),
  INDEX `IDX_GW_PROXY_ACTIVE` (`activeFlag`)
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
  KEY `IDX_TIMER_SCHED_NAME` (`schedulerName`),
  KEY `IDX_TIMER_SCHED_INST` (`schedulerInstanceId`),
  KEY `IDX_TIMER_SCHED_STATUS` (`schedulerStatus`),
  KEY `IDX_TIMER_SCHED_HEART` (`lastHeartbeatTime`)
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
  KEY `IDX_TIMER_TASK_NAME` (`taskName`),
  KEY `IDX_TIMER_TASK_SCHED` (`schedulerId`),
  KEY `IDX_TIMER_TASK_TYPE` (`scheduleType`),
  KEY `IDX_TIMER_TASK_STATUS` (`taskStatus`),
  KEY `IDX_TIMER_TASK_NEXT` (`nextRunTime`),
  KEY `IDX_TIMER_TASK_LAST` (`lastRunTime`),
  KEY `IDX_TIMER_TASK_ACTIVE` (`activeFlag`),
  KEY `IDX_TIMER_TASK_EXEC` (`executorType`),
  KEY `IDX_TIMER_TASK_TOOL` (`toolConfigId`),
  KEY `IDX_TIMER_TASK_OP` (`operationType`)
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
  KEY `IDX_TIMER_LOG_TASK` (`taskId`),
  KEY `IDX_TIMER_LOG_NAME` (`taskName`),
  KEY `IDX_TIMER_LOG_SCHED` (`schedulerId`),
  KEY `IDX_TIMER_LOG_START` (`executionStartTime`),
  KEY `IDX_TIMER_LOG_STATUS` (`executionStatus`),
  KEY `IDX_TIMER_LOG_SUCCESS` (`resultSuccess`),
  KEY `IDX_TIMER_LOG_LEVEL` (`logLevel`),
  KEY `IDX_TIMER_LOG_TIME` (`logTimestamp`)
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
  KEY `IDX_TOOL_CONFIG_NAME` (`toolName`),
  KEY `IDX_TOOL_CONFIG_TYPE` (`toolType`),
  KEY `IDX_TOOL_CONFIG_CFGNAME` (`configName`),
  KEY `IDX_TOOL_CONFIG_GROUP` (`configGroupId`),
  KEY `IDX_TOOL_CONFIG_STATUS` (`configStatus`),
  KEY `IDX_TOOL_CONFIG_DEFAULT` (`defaultFlag`),
  KEY `IDX_TOOL_CONFIG_ACTIVE` (`activeFlag`)
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
  KEY `IDX_TOOL_GROUP_NAME` (`groupName`),
  KEY `IDX_TOOL_GROUP_PARENT` (`parentGroupId`),
  KEY `IDX_TOOL_GROUP_TYPE` (`groupType`),
  KEY `IDX_TOOL_GROUP_SORT` (`sortOrder`),
  KEY `IDX_TOOL_GROUP_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工具配置分组表 - 用于对工具配置进行分组管理';

CREATE TABLE `HUB_GW_LOG_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `logConfigId` VARCHAR(32) NOT NULL COMMENT '日志配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称',
  `configDesc` VARCHAR(200) DEFAULT NULL COMMENT '配置描述',
  
  -- 日志内容控制
  `logFormat` VARCHAR(50) NOT NULL DEFAULT 'JSON' COMMENT '日志格式(JSON,TEXT,CSV)',
  `recordRequestBody` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否记录请求体(N否,Y是)',
  `recordResponseBody` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否记录响应体(N否,Y是)',
  `recordHeaders` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否记录请求/响应头(N否,Y是)',
  `maxBodySizeBytes` INT NOT NULL DEFAULT 4096 COMMENT '最大记录报文大小(字节)',
  
  -- 日志输出目标配置
  `outputTargets` VARCHAR(200) NOT NULL DEFAULT 'CONSOLE' COMMENT '输出目标,逗号分隔(CONSOLE,FILE,DATABASE,MONGODB,ELASTICSEARCH)',
  `fileConfig` TEXT DEFAULT NULL COMMENT '文件输出配置,JSON格式',
  `databaseConfig` TEXT DEFAULT NULL COMMENT '数据库输出配置,JSON格式',
  `mongoConfig` TEXT DEFAULT NULL COMMENT 'MongoDB输出配置,JSON格式',
  `elasticsearchConfig` TEXT DEFAULT NULL COMMENT 'Elasticsearch输出配置,JSON格式',
  `clickhouseConfig` TEXT DEFAULT NULL COMMENT 'clickhouseConfig输出配置,JSON格式',
  
  -- 异步和批量处理配置
  `enableAsyncLogging` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用异步日志(N否,Y是)',
  `asyncQueueSize` INT NOT NULL DEFAULT 10000 COMMENT '异步队列大小',
  `asyncFlushIntervalMs` INT NOT NULL DEFAULT 1000 COMMENT '异步刷新间隔(毫秒)',
  `enableBatchProcessing` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用批量处理(N否,Y是)',
  `batchSize` INT NOT NULL DEFAULT 100 COMMENT '批处理大小',
  `batchTimeoutMs` INT NOT NULL DEFAULT 5000 COMMENT '批处理超时时间(毫秒)',
  
  -- 日志保留和轮转配置
  `logRetentionDays` INT NOT NULL DEFAULT 30 COMMENT '日志保留天数',
  `enableFileRotation` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用文件轮转(N否,Y是)',
  `maxFileSizeMB` INT DEFAULT 100 COMMENT '最大文件大小(MB)',
  `maxFileCount` INT DEFAULT 10 COMMENT '最大文件数量',
  `rotationPattern` VARCHAR(100) DEFAULT 'DAILY' COMMENT '轮转模式(HOURLY,DAILY,WEEKLY,SIZE_BASED)',
  
  -- 敏感数据处理
  `enableSensitiveDataMasking` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用敏感数据脱敏(N否,Y是)',
  `sensitiveFields` TEXT DEFAULT NULL COMMENT '敏感字段列表,JSON数组格式',
  `maskingPattern` VARCHAR(100) DEFAULT '***' COMMENT '脱敏替换模式',
  
  -- 性能优化配置
  `bufferSize` INT NOT NULL DEFAULT 8192 COMMENT '缓冲区大小(字节)',
  `flushThreshold` INT NOT NULL DEFAULT 100 COMMENT '刷新阈值(条目数)',
  
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
  PRIMARY KEY (`tenantId`, `logConfigId`),
  INDEX `idx_HUB_GW_LOG_CONFIG_name` (`configName`),
  INDEX `idx_HUB_GW_LOG_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志配置表 - 存储网关日志相关配置';

CREATE TABLE `HUB_GW_ACCESS_LOG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `traceId` VARCHAR(64) NOT NULL COMMENT '链路追踪ID(作为主键)',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '网关实例ID',
  `gatewayInstanceName` VARCHAR(300) DEFAULT NULL COMMENT '网关实例名称(冗余字段,便于查询显示)',
  `gatewayNodeIp` VARCHAR(50) NOT NULL COMMENT '网关节点IP地址',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID',
  `routeName` VARCHAR(300) DEFAULT NULL COMMENT '路由名称(冗余字段,便于查询显示)',
  `serviceDefinitionId` VARCHAR(32) DEFAULT NULL COMMENT '服务定义ID',
  `serviceName` VARCHAR(300) DEFAULT NULL COMMENT '服务名称(冗余字段,便于查询显示)',
  `proxyType` VARCHAR(50) DEFAULT NULL COMMENT '代理类型(http,websocket,tcp,udp,可为空)',
  `logConfigId` VARCHAR(32) DEFAULT NULL COMMENT '日志配置ID',
  
  -- 请求基本信息
  `requestMethod` VARCHAR(10) NOT NULL COMMENT '请求方法(GET,POST,PUT等)',
  `requestPath` VARCHAR(1000) NOT NULL COMMENT '请求路径',
  `requestQuery` TEXT DEFAULT NULL COMMENT '请求查询参数',
  `requestSize` INT DEFAULT 0 COMMENT '请求大小(字节)',
  `requestHeaders` TEXT DEFAULT NULL COMMENT '请求头信息,JSON格式',
  `requestBody` TEXT DEFAULT NULL COMMENT '请求体(可选,根据配置决定是否记录)',
  
  -- 客户端信息
  `clientIpAddress` VARCHAR(50) NOT NULL COMMENT '客户端IP地址',
  `clientPort` INT DEFAULT NULL COMMENT '客户端端口',
  `userAgent` VARCHAR(1000) DEFAULT NULL COMMENT '用户代理信息',
  `referer` VARCHAR(1000) DEFAULT NULL COMMENT '来源页面',
  `userIdentifier` VARCHAR(100) DEFAULT NULL COMMENT '用户标识(如有)',
  
  -- 关键时间点 (所有时间字段均为DATETIME类型，精确到毫秒)
  `gatewayStartProcessingTime` DATETIME(3) NOT NULL COMMENT '网关开始处理时间(请求开始处理，必填)',
  `backendRequestStartTime` DATETIME(3) DEFAULT NULL COMMENT '后端服务请求开始时间(可选)',
  `backendResponseReceivedTime` DATETIME(3) DEFAULT NULL COMMENT '后端服务响应接收时间(可选)',
  `gatewayFinishedProcessingTime` DATETIME(3) DEFAULT NULL COMMENT '网关处理完成时间(可选，正在处理中或异常中断时为空)',
  
  -- 计算的时间指标 (所有时间指标均为毫秒)
  `totalProcessingTimeMs` INT DEFAULT NULL COMMENT '总处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)',
  `gatewayProcessingTimeMs` INT DEFAULT NULL COMMENT '网关处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)',
  `backendResponseTimeMs` INT DEFAULT NULL COMMENT '后端服务响应时间(毫秒，可选)',
  
  -- 响应信息
  `gatewayStatusCode` INT NOT NULL COMMENT '网关响应状态码',
  `backendStatusCode` INT DEFAULT NULL COMMENT '后端服务状态码',
  `responseSize` INT DEFAULT 0 COMMENT '响应大小(字节)',
  `responseHeaders` TEXT DEFAULT NULL COMMENT '响应头信息,JSON格式',
  `responseBody` TEXT DEFAULT NULL COMMENT '响应体(可选,根据配置决定是否记录)',
  
  -- 转发基本信息
  `matchedRoute` VARCHAR(500) DEFAULT NULL COMMENT '匹配的路由路径',
  `forwardAddress` TEXT DEFAULT NULL COMMENT '转发地址',
  `forwardMethod` VARCHAR(10) DEFAULT NULL COMMENT '转发方法',
  `forwardParams` TEXT DEFAULT NULL COMMENT '转发参数,JSON格式',
  `forwardHeaders` TEXT DEFAULT NULL COMMENT '转发头信息,JSON格式',
  `forwardBody` TEXT DEFAULT NULL COMMENT '转发报文内容',
  `loadBalancerDecision` VARCHAR(500) DEFAULT NULL COMMENT '负载均衡决策信息',
  
  -- 错误信息
  `errorMessage` TEXT DEFAULT NULL COMMENT '错误信息(如有)',
  `errorCode` VARCHAR(100) DEFAULT NULL COMMENT '错误代码(如有)',
  
  -- 追踪信息
  `parentTraceId` VARCHAR(100) DEFAULT NULL COMMENT '父链路追踪ID',
  
  -- 日志重置标记和次数
  `resetFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '日志重置标记(N否,Y是)',
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `resetCount` INT NOT NULL DEFAULT 0 COMMENT '重置次数',
  
  -- 标准数据库字段
  `logLevel` VARCHAR(20) NOT NULL DEFAULT 'INFO' COMMENT '日志级别',
  `logType` VARCHAR(50) NOT NULL DEFAULT 'ACCESS' COMMENT '日志类型',
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
  
  PRIMARY KEY (`tenantId`, `traceId`),
  -- 核心查询索引（高频查询字段）
  INDEX `idx_HUB_GW_ACCESS_LOG_time_instance` (`gatewayStartProcessingTime`, `gatewayInstanceId`),
  INDEX `idx_HUB_GW_ACCESS_LOG_time_route` (`gatewayStartProcessingTime`, `routeConfigId`),
  INDEX `idx_HUB_GW_ACCESS_LOG_time_service` (`gatewayStartProcessingTime`, `serviceDefinitionId`),
  
  -- 名称字段查询索引（利用冗余字段，避免JOIN）
  INDEX `idx_HUB_GW_ACCESS_LOG_instance_name` (`gatewayInstanceName`, `gatewayStartProcessingTime`),
  INDEX `idx_HUB_GW_ACCESS_LOG_route_name` (`routeName`, `gatewayStartProcessingTime`),
  INDEX `idx_HUB_GW_ACCESS_LOG_service_name` (`serviceName`, `gatewayStartProcessingTime`),
  
  -- 业务查询索引
  INDEX `idx_HUB_GW_ACCESS_LOG_client_ip` (`clientIpAddress`, `gatewayStartProcessingTime`),
  INDEX `idx_HUB_GW_ACCESS_LOG_status_time` (`gatewayStatusCode`, `gatewayStartProcessingTime`),
  INDEX `idx_HUB_GW_ACCESS_LOG_proxy_type` (`proxyType`, `gatewayStartProcessingTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关访问日志表 - 记录API网关的请求和响应详细信息,开始时间必填,完成时间可选(支持处理中状态),含冗余字段优化查询性能';

-- =====================================================
-- 字段长度调整：支持多服务定义ID和服务名称（多服务场景）
-- 注意：此处使用独立ALTER语句，避免直接修改历史建表语句，保证向后兼容
-- 变更内容：
--  1) HUB_GW_ACCESS_LOG.serviceDefinitionId 扩展为 1000 字符
--  2) HUB_GW_ACCESS_LOG.serviceName 扩展为 1000 字符
--  3) HUB_GW_ROUTE_CONFIG.serviceDefinitionId 扩展为 1000 字符
-- =====================================================
ALTER TABLE `HUB_GW_ACCESS_LOG`
  MODIFY COLUMN `serviceDefinitionId` VARCHAR(1000) DEFAULT NULL COMMENT '服务定义ID（支持多服务，逗号分隔）';

ALTER TABLE `HUB_GW_ACCESS_LOG`
  MODIFY COLUMN `serviceName` VARCHAR(1000) DEFAULT NULL COMMENT '服务名称(冗余字段,便于查询显示,支持多服务)';

ALTER TABLE `HUB_GW_ROUTE_CONFIG`
  MODIFY COLUMN `serviceDefinitionId` VARCHAR(1000) DEFAULT NULL COMMENT '关联的服务定义ID（支持多服务，逗号分隔）';

-- =====================================================
-- 后端追踪日志表 - HUB_GW_BACKEND_TRACE_LOG
-- 对应结构：internal/gateway/logwrite/types/backend_trace_log.go (BackendTraceLog)
-- 说明：
--   1. 作为 HUB_GW_ACCESS_LOG 的从表，记录每个后端服务的转发明细
--   2. TraceId + BackendTraceId 作为联合主键
--   3. 包含租户ID和标准通用字段，支持多租户和审计
-- =====================================================
CREATE TABLE `HUB_GW_BACKEND_TRACE_LOG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `traceId` VARCHAR(64) NOT NULL COMMENT '链路追踪ID，关联主表 HUB_GW_ACCESS_LOG.traceId',
  `backendTraceId` VARCHAR(64) NOT NULL COMMENT '后端服务追踪ID，用于区分同一请求的多个后端服务',

  -- 服务信息（单个后端服务一次转发一条记录）
  `serviceDefinitionId` VARCHAR(32) DEFAULT NULL COMMENT '服务定义ID',
  `serviceName` VARCHAR(300) DEFAULT NULL COMMENT '服务名称(冗余字段,便于查询)',

  -- 转发信息
  `forwardAddress` TEXT DEFAULT NULL COMMENT '实际转发目标地址(完整URL)',
  `forwardMethod` VARCHAR(10) DEFAULT NULL COMMENT '转发HTTP方法',
  `forwardPath` VARCHAR(1000) DEFAULT NULL COMMENT '转发路径',
  `forwardQuery` TEXT DEFAULT NULL COMMENT '转发查询参数',
  `forwardHeaders` LONGTEXT DEFAULT NULL COMMENT '转发请求头(JSON格式)',
  `forwardBody` LONGTEXT DEFAULT NULL COMMENT '转发请求体',
  `requestSize` INT DEFAULT 0 COMMENT '请求大小(字节，向后端发送的请求体大小)',

  -- 负载均衡信息
  `loadBalancerStrategy` VARCHAR(100) DEFAULT NULL COMMENT '负载均衡策略(round-robin, random, weighted等)',
  `loadBalancerDecision` VARCHAR(500) DEFAULT NULL COMMENT '负载均衡选择决策信息',

  -- 时间信息
  `requestStartTime` DATETIME(3) NOT NULL COMMENT '向后端发起请求的时间',
  `responseReceivedTime` DATETIME(3) DEFAULT NULL COMMENT '接收到后端响应的时间',
  `requestDurationMs` INT DEFAULT NULL COMMENT '请求耗时(毫秒,0表示未完成)',

  -- 响应信息
  `statusCode` INT DEFAULT NULL COMMENT '后端服务返回的HTTP状态码(0表示未收到响应)',
  `responseSize` INT DEFAULT 0 COMMENT '后端响应大小(字节)',
  `responseHeaders` LONGTEXT DEFAULT NULL COMMENT '后端响应头信息(JSON格式)',
  `responseBody` LONGTEXT DEFAULT NULL COMMENT '后端响应体内容',

  -- 错误信息
  `errorCode` VARCHAR(100) DEFAULT NULL COMMENT '错误代码',
  `errorMessage` LONGTEXT DEFAULT NULL COMMENT '详细错误信息',

  -- 状态信息
  `successFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否成功(Y成功,N失败)',
  `traceStatus` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '后端调用状态(pending,success,failed,timeout)',
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',

  -- 扩展信息
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性(JSON格式)',

  -- 标准数据库字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '记录创建者',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '记录修改者',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

  PRIMARY KEY (`traceId`, `backendTraceId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关后端追踪日志表 - 记录每个后端服务的转发明细';

-- 索引设计：参考数据库规范，兼顾多租户与常用查询场景
CREATE INDEX `IDX_GW_BTRACE_TRACE` ON `HUB_GW_BACKEND_TRACE_LOG` (`tenantId`, `traceId`);
CREATE INDEX `IDX_GW_BTRACE_SERVICE` ON `HUB_GW_BACKEND_TRACE_LOG` (`tenantId`, `serviceDefinitionId`, `requestStartTime`);
CREATE INDEX `IDX_GW_BTRACE_TIME` ON `HUB_GW_BACKEND_TRACE_LOG` (`requestStartTime`);
CREATE INDEX `IDX_GW_BTRACE_TSTATUS` ON `HUB_GW_BACKEND_TRACE_LOG` (`tenantId`, `traceStatus`, `requestStartTime`);
CREATE INDEX `IDX_GW_BTRACE_ADDTIME` ON `HUB_GW_BACKEND_TRACE_LOG` (`tenantId`, `addTime`);

-- 安全配置表 - 存储网关安全策略配置
CREATE TABLE `HUB_GW_SECURITY_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '安全配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级安全配置)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级安全配置)',
  `configName` VARCHAR(100) NOT NULL COMMENT '安全配置名称',
  `configDesc` VARCHAR(200) DEFAULT NULL COMMENT '安全配置描述',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
  `customConfigJson` TEXT DEFAULT NULL COMMENT '自定义配置参数,JSON格式',
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
  PRIMARY KEY (`tenantId`, `securityConfigId`),
  INDEX `idx_HUB_GW_SECURITY_CONFIG_instance` (`gatewayInstanceId`),
  INDEX `idx_HUB_GW_SECURITY_CONFIG_route` (`routeConfigId`),
  INDEX `idx_HUB_GW_SECURITY_CONFIG_priority` (`configPriority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='安全配置表 - 存储网关安全策略配置';

-- IP访问控制配置表 - 存储IP白名单黑名单规则
CREATE TABLE `HUB_GW_IP_ACCESS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `ipAccessConfigId` VARCHAR(32) NOT NULL COMMENT 'IP访问配置ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '关联的安全配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT 'IP访问配置名称',
  `defaultPolicy` VARCHAR(10) NOT NULL DEFAULT 'allow' COMMENT '默认策略(allow允许,deny拒绝)',
  `whitelistIps` TEXT DEFAULT NULL COMMENT 'IP白名单,JSON数组格式',
  `blacklistIps` TEXT DEFAULT NULL COMMENT 'IP黑名单,JSON数组格式',
  `whitelistCidrs` TEXT DEFAULT NULL COMMENT 'CIDR白名单,JSON数组格式',
  `blacklistCidrs` TEXT DEFAULT NULL COMMENT 'CIDR黑名单,JSON数组格式',
  `trustXForwardedFor` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否信任X-Forwarded-For头(N否,Y是)',
  `trustXRealIp` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否信任X-Real-IP头(N否,Y是)',
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
  PRIMARY KEY (`tenantId`, `ipAccessConfigId`),
  INDEX `idx_HUB_GW_IP_ACCESS_CONFIG_security` (`securityConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IP访问控制配置表 - 存储IP白名单黑名单规则';

-- User-Agent访问控制配置表 - 存储User-Agent过滤规则
CREATE TABLE `HUB_GW_UA_ACCESS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `useragentAccessConfigId` VARCHAR(32) NOT NULL COMMENT 'User-Agent访问配置ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '关联的安全配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT 'User-Agent访问配置名称',
  `defaultPolicy` VARCHAR(10) NOT NULL DEFAULT 'allow' COMMENT '默认策略(allow允许,deny拒绝)',
  `whitelistPatterns` TEXT DEFAULT NULL COMMENT 'User-Agent白名单模式,JSON数组格式,支持正则表达式',
  `blacklistPatterns` TEXT DEFAULT NULL COMMENT 'User-Agent黑名单模式,JSON数组格式,支持正则表达式',
  `blockEmptyUserAgent` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否阻止空User-Agent(N否,Y是)',
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
  PRIMARY KEY (`tenantId`, `useragentAccessConfigId`),
  INDEX `idx_HUB_GW_UA_ACCESS_CONFIG_security` (`securityConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User-Agent访问控制配置表 - 存储User-Agent过滤规则';

-- API访问控制配置表 - 存储API路径和方法过滤规则
CREATE TABLE `HUB_GW_API_ACCESS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `apiAccessConfigId` VARCHAR(32) NOT NULL COMMENT 'API访问配置ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '关联的安全配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT 'API访问配置名称',
  `defaultPolicy` VARCHAR(10) NOT NULL DEFAULT 'allow' COMMENT '默认策略(allow允许,deny拒绝)',
  `whitelistPaths` TEXT DEFAULT NULL COMMENT 'API路径白名单,JSON数组格式,支持通配符',
  `blacklistPaths` TEXT DEFAULT NULL COMMENT 'API路径黑名单,JSON数组格式,支持通配符',
  `allowedMethods` VARCHAR(200) DEFAULT 'GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS' COMMENT '允许的HTTP方法,逗号分隔',
  `blockedMethods` VARCHAR(200) DEFAULT NULL COMMENT '禁止的HTTP方法,逗号分隔',
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
  PRIMARY KEY (`tenantId`, `apiAccessConfigId`),
  INDEX `idx_HUB_GW_API_ACCESS_CONFIG_security` (`securityConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API访问控制配置表 - 存储API路径和方法过滤规则';

-- 域名访问控制配置表 - 存储域名白名单黑名单规则
CREATE TABLE `HUB_GW_DOMAIN_ACCESS_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `domainAccessConfigId` VARCHAR(32) NOT NULL COMMENT '域名访问配置ID',
  `securityConfigId` VARCHAR(32) NOT NULL COMMENT '关联的安全配置ID',
  `configName` VARCHAR(100) NOT NULL COMMENT '域名访问配置名称',
  `defaultPolicy` VARCHAR(10) NOT NULL DEFAULT 'allow' COMMENT '默认策略(allow允许,deny拒绝)',
  `whitelistDomains` TEXT DEFAULT NULL COMMENT '域名白名单,JSON数组格式',
  `blacklistDomains` TEXT DEFAULT NULL COMMENT '域名黑名单,JSON数组格式',
  `allowSubdomains` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否允许子域名(N否,Y是)',
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
  PRIMARY KEY (`tenantId`, `domainAccessConfigId`),
  INDEX `idx_HUB_GW_DOMAIN_ACCESS_CONFIG_security` (`securityConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='域名访问控制配置表 - 存储域名白名单黑名单规则';


-- 系统指标采集表结构创建脚本
-- 基于 pkg/metric/types/metrics.go 中的数据结构设计
-- 遵循项目数据库设计规范

-- ===================================================================
-- 字段长度优化说明 (HUB_METRIC_SERVER_INFO)
-- ===================================================================
-- 
-- 1. 主键字段调整:
--    - metricServerId: 32 -> 64 字符，适应MD5哈希和UUID格式
--    - tenantId: 32 -> 64 字符，支持更复杂的租户标识
-- 
-- 2. 系统信息字段调整:
--    - hostname: 100 -> 255 字符，支持FQDN和长主机名
--    - osType: 50 -> 100 字符，支持详细的操作系统描述
--    - osVersion: 100 -> 255 字符，支持完整的版本信息
--    - kernelVersion: 100 -> 255 字符，支持详细的内核版本
--    - architecture: 50 -> 100 字符，支持复杂的架构描述
--    - serverLocation: 100 -> 255 字符，支持详细的位置描述
-- 
-- 3. 网络信息字段优化:
--    - ipAddress: 50 -> 45 字符，IPv6最大长度为39字符，预留6字符
--    - macAddress: 50 -> 17 字符，MAC地址标准格式为17字符
-- 
-- 4. 新增TEXT字段用于存储复杂数据:
--    - networkInfo: 存储完整的网络信息（所有IP、MAC、接口等）
--    - systemInfo: 存储系统扩展信息（温度、负载、进程统计等）
--    - hardwareInfo: 存储硬件详细信息（CPU详情、内存详情等）
--    - noteText: VARCHAR(500) -> TEXT，支持更长的备注信息
-- 
-- 5. 操作字段调整:
--    - addWho/editWho: 32 -> 64 字符，支持更长的用户标识
--    - oprSeqFlag: 32 -> 64 字符，支持更复杂的操作序列标识
-- 
-- 6. 新增索引:
--    - IDX_METRIC_SERVER_TYPE: 支持按服务器类型查询
-- 
-- ===================================================================
-- JSON字段存储格式示例
-- ===================================================================
-- 
-- networkInfo 字段存储格式:
-- {
--   "primaryIP": "192.168.1.100",
--   "primaryMAC": "00:11:22:33:44:55",
--   "primaryInterface": "eth0",
--   "allIPs": ["192.168.1.100", "10.0.0.1"],
--   "allMACs": ["00:11:22:33:44:55", "00:11:22:33:44:56"],
--   "activeInterfaces": ["eth0", "lo"]
-- }
-- 
-- systemInfo 字段存储格式:
-- {
--   "uptime": 86400,
--   "userCount": 5,
--   "processCount": 150,
--   "loadAvg": {"1min": 0.5, "5min": 0.3, "15min": 0.2},
--   "temperatures": [
--     {"sensor": "CPU", "value": 45.5, "high": 80.0, "critical": 90.0}
--   ]
-- }
-- 
-- hardwareInfo 字段存储格式:
-- {
--   "cpu": {
--     "coreCount": 8,
--     "logicalCount": 16,
--     "model": "Intel Core i7-9700K",
--     "frequency": "3.6GHz"
--   },
--   "memory": {
--     "total": 17179869184,
--     "type": "DDR4",
--     "speed": "3200MHz"
--   },
--   "storage": {
--     "totalDisks": 2,
--     "totalCapacity": 2000000000000
--   }
-- }
-- 
-- ===================================================================

-- 1. 服务器信息主表
CREATE TABLE `HUB_METRIC_SERVER_INFO` (
  `metricServerId` VARCHAR(64) NOT NULL COMMENT '服务器ID',
  `tenantId` VARCHAR(64) NOT NULL COMMENT '租户ID',
  `hostname` VARCHAR(255) NOT NULL COMMENT '主机名',
  `osType` VARCHAR(100) NOT NULL COMMENT '操作系统类型',
  `osVersion` VARCHAR(255) NOT NULL COMMENT '操作系统版本',
  `kernelVersion` VARCHAR(255) DEFAULT NULL COMMENT '内核版本',
  `architecture` VARCHAR(100) NOT NULL COMMENT '系统架构',
  `bootTime` DATETIME NOT NULL COMMENT '系统启动时间',
  `ipAddress` VARCHAR(45) DEFAULT NULL COMMENT '主IP地址',
  `macAddress` VARCHAR(50) DEFAULT NULL COMMENT '主MAC地址',
  `serverLocation` VARCHAR(255) DEFAULT NULL COMMENT '服务器位置',
  `serverType` VARCHAR(50) DEFAULT NULL COMMENT '服务器类型(physical/virtual/unknown)',
  `lastUpdateTime` DATETIME NOT NULL COMMENT '最后更新时间',
  -- 新增网络信息字段
  `networkInfo` TEXT DEFAULT NULL COMMENT '网络信息详情，JSON格式存储所有IP和MAC地址',
  `systemInfo` TEXT DEFAULT NULL COMMENT '系统详细信息，JSON格式存储温度、负载等扩展信息',
  `hardwareInfo` TEXT DEFAULT NULL COMMENT '硬件信息，JSON格式存储CPU、内存、磁盘等硬件详情',
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(64) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(64) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(64) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` TEXT DEFAULT NULL COMMENT '备注信息',
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
  PRIMARY KEY (`tenantId`, `metricServerId`),
  UNIQUE KEY `IDX_METRIC_SERVER_HOST` (`hostname`),
  KEY `IDX_METRIC_SERVER_OS` (`osType`),
  KEY `IDX_METRIC_SERVER_IP` (`ipAddress`),
  KEY `IDX_METRIC_SERVER_TYPE` (`serverType`),
  KEY `IDX_METRIC_SERVER_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_SERVER_UPDATE` (`lastUpdateTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器信息主表';

-- 2. CPU采集日志表
CREATE TABLE `HUB_METRIC_CPU_LOG` (
  `metricCpuLogId` VARCHAR(32) NOT NULL COMMENT 'CPU采集日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `usagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'CPU使用率(0-100)',
  `userPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '用户态CPU使用率',
  `systemPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '系统态CPU使用率',
  `idlePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '空闲CPU使用率',
  `ioWaitPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'I/O等待CPU使用率',
  `irqPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '中断处理CPU使用率',
  `softIrqPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '软中断处理CPU使用率',
  `coreCount` INT NOT NULL DEFAULT 0 COMMENT 'CPU核心数',
  `logicalCount` INT NOT NULL DEFAULT 0 COMMENT '逻辑CPU数',
  `loadAvg1` DECIMAL(8,2) NOT NULL DEFAULT 0.00 COMMENT '1分钟负载平均值',
  `loadAvg5` DECIMAL(8,2) NOT NULL DEFAULT 0.00 COMMENT '5分钟负载平均值',
  `loadAvg15` DECIMAL(8,2) NOT NULL DEFAULT 0.00 COMMENT '15分钟负载平均值',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
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
  PRIMARY KEY (`tenantId`, `metricCpuLogId`),
  KEY `IDX_METRIC_CPU_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_CPU_TIME` (`collectTime`),
  KEY `IDX_METRIC_CPU_USAGE` (`usagePercent`),
  KEY `IDX_METRIC_CPU_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_CPU_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_CPU_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CPU采集日志表';

-- 3. 内存采集日志表
CREATE TABLE `HUB_METRIC_MEMORY_LOG` (
  `metricMemoryLogId` VARCHAR(32) NOT NULL COMMENT '内存采集日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `totalMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '总内存(字节)',
  `availableMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '可用内存(字节)',
  `usedMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '已使用内存(字节)',
  `usagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '内存使用率(0-100)',
  `freeMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '空闲内存(字节)',
  `cachedMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '缓存内存(字节)',
  `buffersMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '缓冲区内存(字节)',
  `sharedMemory` BIGINT NOT NULL DEFAULT 0 COMMENT '共享内存(字节)',
  `swapTotal` BIGINT NOT NULL DEFAULT 0 COMMENT '交换区总大小(字节)',
  `swapUsed` BIGINT NOT NULL DEFAULT 0 COMMENT '交换区已使用(字节)',
  `swapFree` BIGINT NOT NULL DEFAULT 0 COMMENT '交换区空闲(字节)',
  `swapUsagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '交换区使用率(0-100)',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
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
  PRIMARY KEY (`tenantId`, `metricMemoryLogId`),
  KEY `IDX_METRIC_MEMORY_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_MEMORY_TIME` (`collectTime`),
  KEY `IDX_METRIC_MEMORY_USAGE` (`usagePercent`),
  KEY `IDX_METRIC_MEMORY_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_MEMORY_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_MEMORY_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='内存采集日志表';

-- 4. 磁盘分区日志表
CREATE TABLE `HUB_METRIC_DISK_PART_LOG` (
  `metricDiskPartitionLogId` VARCHAR(32) NOT NULL COMMENT '磁盘分区日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `deviceName` VARCHAR(100) NOT NULL COMMENT '设备名称',
  `mountPoint` VARCHAR(200) NOT NULL COMMENT '挂载点',
  `fileSystem` VARCHAR(50) NOT NULL COMMENT '文件系统类型',
  `totalSpace` BIGINT NOT NULL DEFAULT 0 COMMENT '总大小(字节)',
  `usedSpace` BIGINT NOT NULL DEFAULT 0 COMMENT '已使用(字节)',
  `freeSpace` BIGINT NOT NULL DEFAULT 0 COMMENT '可用(字节)',
  `usagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '使用率(0-100)',
  `inodesTotal` BIGINT NOT NULL DEFAULT 0 COMMENT 'inode总数',
  `inodesUsed` BIGINT NOT NULL DEFAULT 0 COMMENT 'inode已使用',
  `inodesFree` BIGINT NOT NULL DEFAULT 0 COMMENT 'inode空闲',
  `inodesUsagePercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'inode使用率(0-100)',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
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
  PRIMARY KEY (`tenantId`, `metricDiskPartitionLogId`),
  KEY `IDX_METRIC_DISK_PART_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_DISK_PART_TIME` (`collectTime`),
  KEY `IDX_METRIC_DISK_PART_DEVICE` (`deviceName`),
  KEY `IDX_METRIC_DISK_PART_USAGE` (`usagePercent`),
  KEY `IDX_METRIC_DISK_PART_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_DISK_PART_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_DISK_PART_SRV_DEV` (`metricServerId`, `deviceName`),
  KEY `IDX_METRIC_DISK_PART_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='磁盘分区采集日志表';

-- 5. 磁盘IO日志表
CREATE TABLE `HUB_METRIC_DISK_IO_LOG` (
  `metricDiskIoLogId` VARCHAR(32) NOT NULL COMMENT '磁盘IO日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `deviceName` VARCHAR(100) NOT NULL COMMENT '设备名称',
  `readCount` BIGINT NOT NULL DEFAULT 0 COMMENT '读取次数',
  `writeCount` BIGINT NOT NULL DEFAULT 0 COMMENT '写入次数',
  `readBytes` BIGINT NOT NULL DEFAULT 0 COMMENT '读取字节数',
  `writeBytes` BIGINT NOT NULL DEFAULT 0 COMMENT '写入字节数',
  `readTime` BIGINT NOT NULL DEFAULT 0 COMMENT '读取时间(毫秒)',
  `writeTime` BIGINT NOT NULL DEFAULT 0 COMMENT '写入时间(毫秒)',
  `ioInProgress` BIGINT NOT NULL DEFAULT 0 COMMENT 'IO进行中数量',
  `ioTime` BIGINT NOT NULL DEFAULT 0 COMMENT 'IO时间(毫秒)',
  `readRate` DECIMAL(20,2) NOT NULL DEFAULT 0.00 COMMENT '读取速率(字节/秒)',
  `writeRate` DECIMAL(20,2) NOT NULL DEFAULT 0.00 COMMENT '写入速率(字节/秒)',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
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
  PRIMARY KEY (`tenantId`, `metricDiskIoLogId`),
  KEY `IDX_METRIC_DISK_IO_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_DISK_IO_TIME` (`collectTime`),
  KEY `IDX_METRIC_DISK_IO_DEVICE` (`deviceName`),
  KEY `IDX_METRIC_DISK_IO_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_DISK_IO_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_DISK_IO_SRV_DEV` (`metricServerId`, `deviceName`),
  KEY `IDX_METRIC_DISK_IO_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='磁盘IO采集日志表';

-- 6. 网络接口日志表
CREATE TABLE `HUB_METRIC_NETWORK_LOG` (
  `metricNetworkLogId` VARCHAR(32) NOT NULL COMMENT '网络接口日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `interfaceName` VARCHAR(100) NOT NULL COMMENT '接口名称',
  `hardwareAddr` VARCHAR(50) DEFAULT NULL COMMENT 'MAC地址',
  `ipAddresses` TEXT DEFAULT NULL COMMENT 'IP地址列表，JSON格式',
  `interfaceStatus` VARCHAR(20) NOT NULL COMMENT '接口状态',
  `interfaceType` VARCHAR(50) DEFAULT NULL COMMENT '接口类型',
  `bytesReceived` BIGINT NOT NULL DEFAULT 0 COMMENT '接收字节数',
  `bytesSent` BIGINT NOT NULL DEFAULT 0 COMMENT '发送字节数',
  `packetsReceived` BIGINT NOT NULL DEFAULT 0 COMMENT '接收包数',
  `packetsSent` BIGINT NOT NULL DEFAULT 0 COMMENT '发送包数',
  `errorsReceived` BIGINT NOT NULL DEFAULT 0 COMMENT '接收错误数',
  `errorsSent` BIGINT NOT NULL DEFAULT 0 COMMENT '发送错误数',
  `droppedReceived` BIGINT NOT NULL DEFAULT 0 COMMENT '接收丢包数',
  `droppedSent` BIGINT NOT NULL DEFAULT 0 COMMENT '发送丢包数',
  `receiveRate` DECIMAL(20,2) DEFAULT 0 COMMENT '接收速率(字节/秒)',
  `sendRate` DECIMAL(20,2) DEFAULT 0 COMMENT '发送速率(字节/秒)',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
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
  PRIMARY KEY (`tenantId`, `metricNetworkLogId`),
  KEY `IDX_METRIC_NETWORK_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_NETWORK_TIME` (`collectTime`),
  KEY `IDX_METRIC_NETWORK_INTERFACE` (`interfaceName`),
  KEY `IDX_METRIC_NETWORK_STATUS` (`interfaceStatus`),
  KEY `IDX_METRIC_NETWORK_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_NETWORK_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_NETWORK_SRV_INT` (`metricServerId`, `interfaceName`),
  KEY `IDX_METRIC_NETWORK_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网络接口采集日志表';

-- 7. 进程信息日志表
CREATE TABLE `HUB_METRIC_PROCESS_LOG` (
  `metricProcessLogId` VARCHAR(32) NOT NULL COMMENT '进程信息日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `processId` INT NOT NULL COMMENT '进程ID',
  `parentProcessId` INT DEFAULT NULL COMMENT '父进程ID',
  `processName` VARCHAR(200) NOT NULL COMMENT '进程名称',
  `processStatus` VARCHAR(50) NOT NULL COMMENT '进程状态',
  `createTime` DATETIME NOT NULL COMMENT '进程启动时间',
  `runTime` BIGINT NOT NULL DEFAULT 0 COMMENT '进程运行时间(秒)',
  `memoryUsage` BIGINT NOT NULL DEFAULT 0 COMMENT '内存使用(字节)',
  `memoryPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '内存使用率(0-100)',
  `cpuPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'CPU使用率(0-100)',
  `threadCount` INT NOT NULL DEFAULT 0 COMMENT '线程数',
  `fileDescriptorCount` INT NOT NULL DEFAULT 0 COMMENT '文件句柄数',
  `commandLine` TEXT DEFAULT NULL COMMENT '命令行参数，JSON格式',
  `executablePath` VARCHAR(500) DEFAULT NULL COMMENT '执行路径',
  `workingDirectory` VARCHAR(500) DEFAULT NULL COMMENT '工作目录',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
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
  PRIMARY KEY (`tenantId`, `metricProcessLogId`),
  KEY `IDX_METRIC_PROCESS_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_PROCESS_TIME` (`collectTime`),
  KEY `IDX_METRIC_PROCESS_PID` (`processId`),
  KEY `IDX_METRIC_PROCESS_NAME` (`processName`),
  KEY `IDX_METRIC_PROCESS_STATUS` (`processStatus`),
  KEY `IDX_METRIC_PROCESS_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_PROCESS_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_PROCESS_SRV_PID` (`metricServerId`, `processId`),
  KEY `IDX_METRIC_PROCESS_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='进程信息采集日志表';

-- 8. 进程统计日志表
CREATE TABLE `HUB_METRIC_PROCSTAT_LOG` (
  `metricProcessStatsLogId` VARCHAR(32) NOT NULL COMMENT '进程统计日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `runningCount` INT NOT NULL DEFAULT 0 COMMENT '运行中进程数',
  `sleepingCount` INT NOT NULL DEFAULT 0 COMMENT '睡眠中进程数',
  `stoppedCount` INT NOT NULL DEFAULT 0 COMMENT '停止的进程数',
  `zombieCount` INT NOT NULL DEFAULT 0 COMMENT '僵尸进程数',
  `totalCount` INT NOT NULL DEFAULT 0 COMMENT '总进程数',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
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
  PRIMARY KEY (`tenantId`, `metricProcessStatsLogId`),
  KEY `IDX_METRIC_PROC_STATS_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_PROC_STATS_TIME` (`collectTime`),
  KEY `IDX_METRIC_PROC_STATS_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_PROC_STATS_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_PROC_STATS_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='进程统计采集日志表';

-- 9. 温度信息日志表
CREATE TABLE `HUB_METRIC_TEMP_LOG` (
  `metricTemperatureLogId` VARCHAR(32) NOT NULL COMMENT '温度信息日志ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `metricServerId` VARCHAR(32) NOT NULL COMMENT '关联服务器ID',
  `sensorName` VARCHAR(100) NOT NULL COMMENT '传感器名称',
  `temperatureValue` DECIMAL(6,2) NOT NULL DEFAULT 0.00 COMMENT '温度值(摄氏度)',
  `highThreshold` DECIMAL(6,2) DEFAULT NULL COMMENT '高温阈值',
  `criticalThreshold` DECIMAL(6,2) DEFAULT NULL COMMENT '严重高温阈值',
  `collectTime` DATETIME NOT NULL COMMENT '采集时间',
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
  PRIMARY KEY (`tenantId`, `metricTemperatureLogId`),
  KEY `IDX_METRIC_TEMP_SERVER` (`metricServerId`),
  KEY `IDX_METRIC_TEMP_TIME` (`collectTime`),
  KEY `IDX_METRIC_TEMP_SENSOR` (`sensorName`),
  KEY `IDX_METRIC_TEMP_ACTIVE` (`activeFlag`),
  KEY `IDX_METRIC_TEMP_SRV_TIME` (`metricServerId`, `collectTime`),
  KEY `IDX_METRIC_TEMP_SRV_SENSOR` (`metricServerId`, `sensorName`),
  KEY `IDX_METRIC_TEMP_TNT_TIME` (`tenantId`, `collectTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='温度信息采集日志表';

-- =====================================================
-- 服务注册中心数据库表结构设计 (MySQL版本)
-- =====================================================

-- 服务分组表 - 存储服务分组和授权信息
CREATE TABLE `HUB_REGISTRY_SERVICE_GROUP` (
  -- 主键和租户信息
  `serviceGroupId` VARCHAR(32) NOT NULL COMMENT '服务分组ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 分组基本信息
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称',
  `groupDescription` VARCHAR(500) DEFAULT NULL COMMENT '分组描述',
  `groupType` VARCHAR(50) DEFAULT 'BUSINESS' COMMENT '分组类型(BUSINESS,SYSTEM,TEST)',
  
  -- 授权信息
  `ownerUserId` VARCHAR(32) NOT NULL COMMENT '分组所有者用户ID',
  `adminUserIds` TEXT DEFAULT NULL COMMENT '管理员用户ID列表，JSON格式',
  `readUserIds` TEXT DEFAULT NULL COMMENT '只读用户ID列表，JSON格式',
  `accessControlEnabled` VARCHAR(1) DEFAULT 'N' COMMENT '是否启用访问控制(N否,Y是)',
  
  -- 配置信息
  `defaultProtocolType` VARCHAR(20) DEFAULT 'HTTP' COMMENT '默认协议类型',
  `defaultLoadBalanceStrategy` VARCHAR(50) DEFAULT 'ROUND_ROBIN' COMMENT '默认负载均衡策略',
  `defaultHealthCheckUrl` VARCHAR(500) DEFAULT '/health' COMMENT '默认健康检查URL',
  `defaultHealthCheckIntervalSeconds` INT DEFAULT 30 COMMENT '默认健康检查间隔(秒)',
  
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
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `serviceGroupId`),
  KEY `IDX_REGISTRY_GROUP_NAME` (`tenantId`, `groupName`),
  KEY `IDX_REGISTRY_GROUP_TYPE` (`groupType`),
  KEY `IDX_REGISTRY_GROUP_OWNER` (`ownerUserId`),
  KEY `IDX_REGISTRY_GROUP_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务分组表 - 存储服务分组和授权信息';

-- 服务表 - 存储服务基本信息
CREATE TABLE `HUB_REGISTRY_SERVICE` (
  -- 主键和租户信息
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称，主键',
  
  -- 关联分组（主键关联）
  `serviceGroupId` VARCHAR(32) NOT NULL COMMENT '服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键',
  -- 冗余字段（便于查询和展示）
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称，冗余字段便于查询',
  
  -- 服务基本信息
  `serviceDescription` VARCHAR(500) DEFAULT NULL COMMENT '服务描述',
  
  -- 注册管理配置
  `registryType` VARCHAR(20) NOT NULL DEFAULT 'INTERNAL' COMMENT '注册类型(INTERNAL:内部管理,NACOS:Nacos注册中心,CONSUL:Consul,EUREKA:Eureka,ETCD:ETCD,ZOOKEEPER:ZooKeeper)',
  `externalRegistryConfig` TEXT DEFAULT NULL COMMENT '外部注册中心配置，JSON格式，仅当registryType非INTERNAL时使用',
  
  -- 服务配置
  `protocolType` VARCHAR(20) DEFAULT 'HTTP' COMMENT '协议类型(HTTP,HTTPS,TCP,UDP,GRPC)',
  `contextPath` VARCHAR(200) DEFAULT '' COMMENT '上下文路径',
  `loadBalanceStrategy` VARCHAR(50) DEFAULT 'ROUND_ROBIN' COMMENT '负载均衡策略',
  
  -- 健康检查配置
  `healthCheckUrl` VARCHAR(500) DEFAULT '/health' COMMENT '健康检查URL',
  `healthCheckIntervalSeconds` INT DEFAULT 30 COMMENT '健康检查间隔(秒)',
  `healthCheckTimeoutSeconds` INT DEFAULT 5 COMMENT '健康检查超时(秒)',
  `healthCheckType` VARCHAR(20) DEFAULT 'HTTP' COMMENT '健康检查类型(HTTP,TCP)',
  `healthCheckMode` VARCHAR(20) DEFAULT 'ACTIVE' COMMENT '健康检查模式(ACTIVE:主动探测,PASSIVE:客户端上报)',
  
  -- 元数据和标签
  `metadataJson` TEXT DEFAULT NULL COMMENT '服务元数据，JSON格式',
  `tagsJson` TEXT DEFAULT NULL COMMENT '服务标签，JSON格式',
  
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
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `serviceName`),
  -- 主键关联索引（用于外键关联查询）
  KEY `IDX_REGISTRY_SVC_GROUP_ID` (`tenantId`, `serviceGroupId`),
  -- 冗余字段索引（用于业务查询和展示）
  KEY `IDX_REGISTRY_SVC_GROUP_NAME` (`groupName`),
  KEY `IDX_REGISTRY_SVC_REGISTRY_TYPE` (`registryType`),
  KEY `IDX_REGISTRY_SVC_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务表 - 存储服务的基本信息和配置，支持内部管理和外部注册中心代理模式';

-- 服务实例表 - 存储具体的服务实例
CREATE TABLE `HUB_REGISTRY_SERVICE_INSTANCE` (
  -- 主键和租户信息
  `serviceInstanceId` VARCHAR(100) NOT NULL COMMENT '服务实例ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联服务和分组（主键关联）
  `serviceGroupId` VARCHAR(32) NOT NULL COMMENT '服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键',
  -- 冗余字段（便于查询和展示）
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称，冗余字段便于查询',
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称，冗余字段便于查询',
  
  -- 网络连接信息
  `hostAddress` VARCHAR(100) NOT NULL COMMENT '主机地址',
  `portNumber` INT NOT NULL COMMENT '端口号',
  `contextPath` VARCHAR(200) DEFAULT '' COMMENT '上下文路径',
  
  -- 实例状态信息
  `instanceStatus` VARCHAR(20) NOT NULL DEFAULT 'UP' COMMENT '实例状态(UP,DOWN,STARTING,OUT_OF_SERVICE)',
  `healthStatus` VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN' COMMENT '健康状态(HEALTHY,UNHEALTHY,UNKNOWN)',
  
  -- 负载均衡配置
  `weightValue` INT NOT NULL DEFAULT 100 COMMENT '权重值',
  
  -- 客户端信息
  `clientId` VARCHAR(100) DEFAULT NULL COMMENT '客户端ID',
  `clientVersion` VARCHAR(50) DEFAULT NULL COMMENT '客户端版本',
  `clientType` VARCHAR(50) DEFAULT 'SERVICE' COMMENT '客户端类型(SERVICE,GATEWAY,ADMIN)',
  `tempInstanceFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '临时实例标记(Y是临时实例,N否)',
  
  -- 健康检查统计
  `heartbeatFailCount` INT NOT NULL DEFAULT 0 COMMENT '心跳检查失败次数，仅用于计数',
  
  -- 元数据和标签
  `metadataJson` TEXT DEFAULT NULL COMMENT '实例元数据，JSON格式',
  `tagsJson` TEXT DEFAULT NULL COMMENT '实例标签，JSON格式',
  
  -- 时间戳信息
  `registerTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  `lastHeartbeatTime` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
  `lastHealthCheckTime` DATETIME DEFAULT NULL COMMENT '最后健康检查时间',
  
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
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `serviceInstanceId`),
  KEY `IDX_REGISTRY_INSTANCE` (`tenantId`, `serviceGroupId`, `serviceName`, `hostAddress`, `portNumber`),
  -- 主键关联索引（用于外键关联查询）
  KEY `IDX_REGISTRY_INST_GROUP_ID` (`tenantId`, `serviceGroupId`),
  -- 冗余字段索引（用于业务查询和展示）
  KEY `IDX_REGISTRY_INST_SVC_NAME` (`serviceName`),
  KEY `IDX_REGISTRY_INST_GROUP_NAME` (`groupName`),
  -- 业务状态索引
  KEY `IDX_REGISTRY_INST_STATUS` (`instanceStatus`),
  KEY `IDX_REGISTRY_INST_HEALTH` (`healthStatus`),
  KEY `IDX_REGISTRY_INST_HEARTBEAT` (`lastHeartbeatTime`),
  KEY `IDX_REGISTRY_INST_HOST_PORT` (`hostAddress`, `portNumber`),
  KEY `IDX_REGISTRY_INST_CLIENT` (`clientId`),
  KEY `IDX_REGISTRY_INST_ACTIVE` (`activeFlag`),
  KEY `IDX_REGISTRY_INST_TEMP` (`tempInstanceFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务实例表 - 存储具体的服务实例信息';

-- 服务事件日志表 - 记录服务变更事件
CREATE TABLE `HUB_REGISTRY_SERVICE_EVENT` (
  -- 主键和租户信息
  `serviceEventId` VARCHAR(32) NOT NULL COMMENT '服务事件ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联主键字段（用于精确关联到对应表记录）
  `serviceGroupId` VARCHAR(32) DEFAULT NULL COMMENT '服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键',
  `serviceInstanceId` VARCHAR(100) DEFAULT NULL COMMENT '服务实例ID，关联HUB_REGISTRY_SERVICE_INSTANCE表主键',
  
  -- 事件基本信息（冗余字段，便于查询和展示）
  `groupName` VARCHAR(100) DEFAULT NULL COMMENT '分组名称，冗余字段便于查询',
  `serviceName` VARCHAR(100) DEFAULT NULL COMMENT '服务名称，冗余字段便于查询',
  `hostAddress` VARCHAR(100) DEFAULT NULL COMMENT '主机地址，冗余字段便于查询',
  `portNumber` INT DEFAULT NULL COMMENT '端口号，冗余字段便于查询',
  `nodeIpAddress` VARCHAR(100) DEFAULT NULL COMMENT '节点IP地址，记录程序运行的IP',
  `eventType` VARCHAR(50) NOT NULL COMMENT '事件类型(GROUP_CREATE,GROUP_UPDATE,GROUP_DELETE,SERVICE_CREATE,SERVICE_UPDATE,SERVICE_DELETE,INSTANCE_REGISTER,INSTANCE_DEREGISTER,INSTANCE_HEARTBEAT,INSTANCE_HEALTH_CHANGE,INSTANCE_STATUS_CHANGE)',
  `eventSource` VARCHAR(100) DEFAULT NULL COMMENT '事件来源',
  
  -- 事件数据
  `eventDataJson` TEXT DEFAULT NULL COMMENT '事件数据，JSON格式',
  `eventMessage` VARCHAR(1000) DEFAULT NULL COMMENT '事件消息描述',
  
  -- 时间信息
  `eventTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '事件发生时间',
  
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
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `serviceEventId`),
  -- 主键关联索引（用于精确关联查询）
  KEY `IDX_REGISTRY_EVENT_GROUP_ID` (`tenantId`, `serviceGroupId`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_INSTANCE_ID` (`tenantId`, `serviceInstanceId`, `eventTime`),
  -- 冗余字段索引（用于业务查询和展示）
  KEY `IDX_REGISTRY_EVENT_GROUP_NAME` (`tenantId`, `groupName`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_SVC_NAME` (`tenantId`, `serviceName`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_HOST` (`tenantId`, `hostAddress`, `portNumber`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_NODE_IP` (`tenantId`, `nodeIpAddress`, `eventTime`),
  -- 事件类型和时间索引
  KEY `IDX_REGISTRY_EVENT_TYPE` (`eventType`, `eventTime`),
  KEY `IDX_REGISTRY_EVENT_TIME` (`eventTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务事件日志表 - 记录服务注册发现相关的所有事件';

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
-- ==========================================
-- JVM监控系统数据库表结构设计
-- 基于 flux-datahub-registry-monitor JVM监控模型
-- 遵循数据库设计规范 (naming-convention.md)
-- ==========================================

-- ==========================================
-- 1. JVM资源信息主表
-- 存储JVM整体资源监控信息的快照数据
-- ==========================================
CREATE TABLE HUB_MONITOR_JVM_RESOURCE (
                                          jvmResourceId VARCHAR(100) NOT NULL COMMENT 'JVM资源记录ID（由应用端生成的唯一标识），主键',
                                          tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                          serviceGroupId VARCHAR(32) NOT NULL COMMENT '服务分组ID，主键',

    -- 应用标识信息
                                          applicationName VARCHAR(100) NOT NULL COMMENT '应用名称',
                                          groupName VARCHAR(100) NOT NULL COMMENT '分组名称',
                                          hostName VARCHAR(100) DEFAULT NULL COMMENT '主机名',
                                          hostIpAddress VARCHAR(50) DEFAULT NULL COMMENT '主机IP地址',

    -- 时间相关字段
                                          collectionTime DATETIME NOT NULL COMMENT '数据采集时间',
                                          jvmStartTime DATETIME NOT NULL COMMENT 'JVM启动时间',
                                          jvmUptimeMs BIGINT NOT NULL DEFAULT 0 COMMENT 'JVM运行时长（毫秒）',

    -- 健康状态字段
                                          healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT 'JVM整体健康标记(Y健康,N异常)',
                                          healthGrade VARCHAR(20) DEFAULT NULL COMMENT 'JVM健康等级(EXCELLENT/GOOD/FAIR/POOR)',
                                          requiresAttentionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即关注(Y是,N否)',
                                          summaryText VARCHAR(500) DEFAULT NULL COMMENT '监控摘要信息',

    -- 系统属性（JSON格式）
                                          systemPropertiesJson LONGTEXT DEFAULT NULL COMMENT 'JVM系统属性，JSON格式（可能包含大量系统属性）',

    -- 通用字段
                                          addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                          addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                          editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                          editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                          oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                          currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                          activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                          noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                          PRIMARY KEY (tenantId, serviceGroupId, jvmResourceId),
                                          KEY IDX_MONITOR_JVM_APP (applicationName),
                                          KEY IDX_MONITOR_JVM_TIME (collectionTime),
                                          KEY IDX_MONITOR_JVM_HEALTH (healthyFlag, requiresAttentionFlag),
                                          KEY IDX_MONITOR_JVM_HOST (hostIpAddress),
                                          KEY IDX_MONITOR_JVM_GROUP (serviceGroupId, groupName)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM资源监控主表';

-- ==========================================
-- 2. 内存信息表（堆内存和非堆内存）
-- 存储JVM堆内存和非堆内存的使用情况
-- ==========================================
CREATE TABLE HUB_MONITOR_JVM_MEMORY (
                                        jvmMemoryId VARCHAR(32) NOT NULL COMMENT 'JVM内存记录ID，主键',
                                        tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                        jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 内存类型
                                        memoryType VARCHAR(20) NOT NULL COMMENT '内存类型(HEAP/NON_HEAP)',

    -- 内存使用情况（字节）
                                        initMemoryBytes BIGINT NOT NULL DEFAULT 0 COMMENT '初始内存大小（字节）',
                                        usedMemoryBytes BIGINT NOT NULL DEFAULT 0 COMMENT '已使用内存大小（字节）',
                                        committedMemoryBytes BIGINT NOT NULL DEFAULT 0 COMMENT '已提交内存大小（字节）',
                                        maxMemoryBytes BIGINT NOT NULL DEFAULT -1 COMMENT '最大内存大小（字节），-1表示无限制',

    -- 计算指标
                                        usagePercent DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '内存使用率（百分比）',
                                        healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '内存健康标记(Y健康,N异常)',

    -- 时间字段
                                        collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 通用字段
                                        addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                        addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                        editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                        editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                        oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                        currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                        activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                        noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                        PRIMARY KEY (tenantId, jvmMemoryId),
                                        KEY IDX_MONITOR_MEM_RES (jvmResourceId),
                                        KEY IDX_MONITOR_MEM_TYPE (memoryType),
                                        KEY IDX_MONITOR_MEM_TIME (collectionTime),
                                        KEY IDX_MONITOR_MEM_USAGE (usagePercent)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM内存监控表';

-- ==========================================
-- 3. 内存池信息表
-- 存储具体内存池的详细使用情况（Eden、Survivor、Old Gen、Metaspace等）
-- ==========================================
CREATE TABLE HUB_MONITOR_JVM_MEM_POOL (
                                          memoryPoolId VARCHAR(32) NOT NULL COMMENT '内存池记录ID，主键',
                                          tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                          jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 内存池基本信息
                                          poolName VARCHAR(100) NOT NULL COMMENT '内存池名称',
                                          poolType VARCHAR(20) NOT NULL COMMENT '内存池类型(HEAP/NON_HEAP)',
                                          poolCategory VARCHAR(50) DEFAULT NULL COMMENT '内存池分类（年轻代/老年代/元数据空间/代码缓存/其他）',

    -- 当前使用情况
                                          currentInitBytes BIGINT NOT NULL DEFAULT 0 COMMENT '当前初始内存（字节）',
                                          currentUsedBytes BIGINT NOT NULL DEFAULT 0 COMMENT '当前已使用内存（字节）',
                                          currentCommittedBytes BIGINT NOT NULL DEFAULT 0 COMMENT '当前已提交内存（字节）',
                                          currentMaxBytes BIGINT NOT NULL DEFAULT -1 COMMENT '当前最大内存（字节）',
                                          currentUsagePercent DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '当前使用率（百分比）',

    -- 峰值使用情况
                                          peakInitBytes BIGINT DEFAULT 0 COMMENT '峰值初始内存（字节）',
                                          peakUsedBytes BIGINT DEFAULT 0 COMMENT '峰值已使用内存（字节）',
                                          peakCommittedBytes BIGINT DEFAULT 0 COMMENT '峰值已提交内存（字节）',
                                          peakMaxBytes BIGINT DEFAULT -1 COMMENT '峰值最大内存（字节）',
                                          peakUsagePercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '峰值使用率（百分比）',

    -- 阈值监控
                                          usageThresholdSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否支持使用阈值监控(Y是,N否)',
                                          usageThresholdBytes BIGINT DEFAULT 0 COMMENT '使用阈值（字节）',
                                          usageThresholdCount BIGINT DEFAULT 0 COMMENT '使用阈值超越次数',
                                          collectionUsageSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否支持收集使用量监控(Y是,N否)',

    -- 健康状态
                                          healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '内存池健康标记(Y健康,N异常)',

    -- 时间字段
                                          collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 通用字段
                                          addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                          addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                          editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                          editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                          oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                          currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                          activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                          noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                          PRIMARY KEY (tenantId, memoryPoolId),
                                          KEY IDX_MONITOR_POOL_RES (jvmResourceId),
                                          KEY IDX_MONITOR_POOL_NAME (poolName),
                                          KEY IDX_MONITOR_POOL_TYPE (poolType),
                                          KEY IDX_MONITOR_POOL_CAT (poolCategory),
                                          KEY IDX_MONITOR_POOL_TIME (collectionTime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM内存池监控表';

-- ==========================================
-- 4. GC快照表（jstat -gc 风格，每次采集一条汇总记录）
-- 存储每次采集时刻的GC状态快照，包含完整的内存区域数据
-- 每次采集插入一条记录，包含所有GC收集器的汇总数据
-- ==========================================
CREATE TABLE HUB_MONITOR_JVM_GC (
                                    gcSnapshotId VARCHAR(32) NOT NULL COMMENT 'GC快照记录ID，主键',
                                    tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                    jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- GC累积统计（从JVM启动到当前采集时刻）
                                    collectionCount BIGINT NOT NULL DEFAULT 0 COMMENT 'GC总次数（累积，所有GC收集器汇总）',
                                    collectionTimeMs BIGINT NOT NULL DEFAULT 0 COMMENT 'GC总耗时（毫秒，累积，所有GC收集器汇总）',

    -- ===== jstat -gc 风格的内存区域数据（单位：KB） =====

    -- Survivor区
                                    s0c BIGINT DEFAULT 0 COMMENT 'Survivor 0 区容量（KB）',
                                    s1c BIGINT DEFAULT 0 COMMENT 'Survivor 1 区容量（KB）',
                                    s0u BIGINT DEFAULT 0 COMMENT 'Survivor 0 区使用量（KB）',
                                    s1u BIGINT DEFAULT 0 COMMENT 'Survivor 1 区使用量（KB）',

    -- Eden区
                                    ec BIGINT DEFAULT 0 COMMENT 'Eden 区容量（KB）',
                                    eu BIGINT DEFAULT 0 COMMENT 'Eden 区使用量（KB）',

    -- Old区
                                    oc BIGINT DEFAULT 0 COMMENT 'Old 区容量（KB）',
                                    ou BIGINT DEFAULT 0 COMMENT 'Old 区使用量（KB）',

    -- Metaspace
                                    mc BIGINT DEFAULT 0 COMMENT 'Metaspace 容量（KB）',
                                    mu BIGINT DEFAULT 0 COMMENT 'Metaspace 使用量（KB）',

    -- 压缩类空间
                                    ccsc BIGINT DEFAULT 0 COMMENT '压缩类空间容量（KB）',
                                    ccsu BIGINT DEFAULT 0 COMMENT '压缩类空间使用量（KB）',

    -- GC统计（jstat -gc 格式）
                                    ygc BIGINT DEFAULT 0 COMMENT '年轻代GC次数',
                                    ygct DECIMAL(10,3) DEFAULT 0.000 COMMENT '年轻代GC总时间（秒）',
                                    fgc BIGINT DEFAULT 0 COMMENT 'Full GC次数',
                                    fgct DECIMAL(10,3) DEFAULT 0.000 COMMENT 'Full GC总时间（秒）',
                                    gct DECIMAL(10,3) DEFAULT 0.000 COMMENT '总GC时间（秒）',

    -- 时间戳信息
                                    collectionTime DATETIME NOT NULL COMMENT '数据采集时间戳',

    -- 通用字段
                                    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                    addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                    editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                    oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                    currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                    activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                    noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                    PRIMARY KEY (tenantId, gcSnapshotId),
                                    KEY IDX_MONITOR_GC_RES (jvmResourceId),
                                    KEY IDX_MONITOR_GC_TIME (collectionTime),
                                    KEY IDX_MONITOR_GC_RES_TIME (jvmResourceId, collectionTime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM GC快照表（jstat -gc风格，每次采集一条汇总记录）';

-- ==========================================
-- 5. 线程信息表
-- 存储JVM线程的详细监控数据
-- ==========================================
CREATE TABLE HUB_MONITOR_JVM_THREAD (
                                        jvmThreadId VARCHAR(32) NOT NULL COMMENT 'JVM线程记录ID，主键',
                                        tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                        jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 基础线程统计
                                        currentThreadCount INT NOT NULL DEFAULT 0 COMMENT '当前线程数',
                                        daemonThreadCount INT NOT NULL DEFAULT 0 COMMENT '守护线程数',
                                        userThreadCount INT NOT NULL DEFAULT 0 COMMENT '用户线程数',
                                        peakThreadCount INT NOT NULL DEFAULT 0 COMMENT '峰值线程数',
                                        totalStartedThreadCount BIGINT NOT NULL DEFAULT 0 COMMENT '总启动线程数',

    -- 性能指标
                                        threadGrowthRatePercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '线程增长率（百分比）',
                                        daemonThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '守护线程比例（百分比）',

    -- 监控功能支持状态
                                        cpuTimeSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT 'CPU时间监控是否支持(Y是,N否)',
                                        cpuTimeEnabled VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT 'CPU时间监控是否启用(Y是,N否)',
                                        memoryAllocSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '内存分配监控是否支持(Y是,N否)',
                                        memoryAllocEnabled VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '内存分配监控是否启用(Y是,N否)',
                                        contentionSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '争用监控是否支持(Y是,N否)',
                                        contentionEnabled VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '争用监控是否启用(Y是,N否)',

    -- 健康状态
                                        healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '线程健康标记(Y健康,N异常)',
                                        healthGrade VARCHAR(20) DEFAULT NULL COMMENT '线程健康等级(EXCELLENT/GOOD/FAIR/POOR)',
                                        requiresAttentionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即关注(Y是,N否)',
                                        potentialIssuesJson LONGTEXT DEFAULT NULL COMMENT '潜在问题列表，JSON格式',

    -- 时间字段
                                        collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 通用字段
                                        addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                        addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                        editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                        editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                        oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                        currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                        activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                        noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                        PRIMARY KEY (tenantId, jvmThreadId),
                                        KEY IDX_MONITOR_THR_RES (jvmResourceId),
                                        KEY IDX_MONITOR_THR_TIME (collectionTime),
                                        KEY IDX_MONITOR_THR_HEALTH (healthyFlag, requiresAttentionFlag),
                                        KEY IDX_MONITOR_THR_COUNT (currentThreadCount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM线程监控表';

-- ==========================================
-- 6. 线程状态统计表
-- 存储不同状态下的线程数量分布
-- ==========================================
CREATE TABLE HUB_MONITOR_JVM_THR_STATE (
                                           threadStateId VARCHAR(32) NOT NULL COMMENT '线程状态记录ID，主键',
                                           tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                           jvmThreadId VARCHAR(32) NOT NULL COMMENT '关联的JVM线程记录ID',
                                           jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 线程状态分布
                                           newThreadCount INT NOT NULL DEFAULT 0 COMMENT 'NEW状态线程数',
                                           runnableThreadCount INT NOT NULL DEFAULT 0 COMMENT 'RUNNABLE状态线程数',
                                           blockedThreadCount INT NOT NULL DEFAULT 0 COMMENT 'BLOCKED状态线程数',
                                           waitingThreadCount INT NOT NULL DEFAULT 0 COMMENT 'WAITING状态线程数',
                                           timedWaitingThreadCount INT NOT NULL DEFAULT 0 COMMENT 'TIMED_WAITING状态线程数',
                                           terminatedThreadCount INT NOT NULL DEFAULT 0 COMMENT 'TERMINATED状态线程数',
                                           totalThreadCount INT NOT NULL DEFAULT 0 COMMENT '总线程数',

    -- 比例指标
                                           activeThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '活跃线程比例（百分比）',
                                           blockedThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '阻塞线程比例（百分比）',
                                           waitingThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '等待状态线程比例（百分比）',

    -- 健康状态
                                           healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '线程状态健康标记(Y健康,N异常)',
                                           healthGrade VARCHAR(20) DEFAULT NULL COMMENT '健康等级(EXCELLENT/GOOD/FAIR/POOR)',

    -- 时间字段
                                           collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 通用字段
                                           addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                           addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                           editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                           editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                           oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                           currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                           activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                           noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                           PRIMARY KEY (tenantId, threadStateId),
                                           KEY IDX_MONITOR_THRST_THR (jvmThreadId),
                                           KEY IDX_MONITOR_THRST_RES (jvmResourceId),
                                           KEY IDX_MONITOR_THRST_TIME (collectionTime),
                                           KEY IDX_MONITOR_THRST_BLOCK (blockedThreadCount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM线程状态统计表';

-- ==========================================
-- 7. 死锁检测信息表
-- 存储JVM中检测到的死锁情况
-- ==========================================
CREATE TABLE HUB_MONITOR_JVM_DEADLOCK (
                                          deadlockId VARCHAR(32) NOT NULL COMMENT '死锁记录ID，主键',
                                          tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                          jvmThreadId VARCHAR(32) NOT NULL COMMENT '关联的JVM线程记录ID',
                                          jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 死锁基本信息
                                          hasDeadlockFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否检测到死锁(Y是,N否)',
                                          deadlockThreadCount INT NOT NULL DEFAULT 0 COMMENT '死锁线程数量',
                                          deadlockThreadIds TEXT DEFAULT NULL COMMENT '死锁线程ID列表，逗号分隔',
                                          deadlockThreadNames TEXT DEFAULT NULL COMMENT '死锁线程名称列表，逗号分隔',

    -- 死锁严重程度
                                          severityLevel VARCHAR(20) DEFAULT NULL COMMENT '严重程度(LOW/MEDIUM/HIGH/CRITICAL)',
                                          severityDescription VARCHAR(200) DEFAULT NULL COMMENT '严重程度描述',
                                          affectedThreadGroups INT DEFAULT 0 COMMENT '影响的线程组数量',

    -- 时间信息
                                          detectionTime DATETIME DEFAULT NULL COMMENT '死锁检测时间',
                                          deadlockDurationMs BIGINT DEFAULT 0 COMMENT '死锁持续时间（毫秒）',
                                          collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 诊断信息
                                          descriptionText VARCHAR(500) DEFAULT NULL COMMENT '死锁描述信息',
                                          recommendedAction VARCHAR(500) DEFAULT NULL COMMENT '建议的解决方案',
                                          alertLevel VARCHAR(20) DEFAULT NULL COMMENT '告警级别(INFO/WARNING/ERROR/CRITICAL/EMERGENCY)',
                                          requiresActionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即处理(Y是,N否)',

    -- 通用字段
                                          addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                          addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                          editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                          editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                          oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                          currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                          activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                          noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                          PRIMARY KEY (tenantId, deadlockId),
                                          KEY IDX_MONITOR_DL_THR (jvmThreadId),
                                          KEY IDX_MONITOR_DL_RES (jvmResourceId),
                                          KEY IDX_MONITOR_DL_TIME (collectionTime),
                                          KEY IDX_MONITOR_DL_FLAG (hasDeadlockFlag),
                                          KEY IDX_MONITOR_DL_SEV (severityLevel),
                                          KEY IDX_MONITOR_DL_ALERT (alertLevel)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM死锁检测信息表';

-- ==========================================
-- 8. 类加载信息表
-- 存储JVM类加载器的统计信息
-- ==========================================
CREATE TABLE HUB_MONITOR_JVM_CLASS (
                                       classLoadingId VARCHAR(32) NOT NULL COMMENT '类加载记录ID，主键',
                                       tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                       jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 类加载统计
                                       loadedClassCount INT NOT NULL DEFAULT 0 COMMENT '当前已加载类数量',
                                       totalLoadedClassCount BIGINT NOT NULL DEFAULT 0 COMMENT '总加载类数量',
                                       unloadedClassCount BIGINT NOT NULL DEFAULT 0 COMMENT '已卸载类数量',

    -- 比例指标
                                       classUnloadRatePercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '类卸载率（百分比）',
                                       classRetentionRatePercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '类保留率（百分比）',

    -- 配置状态
                                       verboseClassLoading VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用详细类加载输出(Y是,N否)',

    -- 性能指标
                                       loadingRatePerHour DECIMAL(10,2) DEFAULT 0.00 COMMENT '每小时平均类加载数量',
                                       loadingEfficiency DECIMAL(5,2) DEFAULT 0.00 COMMENT '类加载效率',
                                       memoryEfficiency VARCHAR(100) DEFAULT NULL COMMENT '内存使用效率评估',
                                       loaderHealth VARCHAR(50) DEFAULT NULL COMMENT '类加载器健康状况',

    -- 健康状态
                                       healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '类加载健康标记(Y健康,N异常)',
                                       healthGrade VARCHAR(20) DEFAULT NULL COMMENT '健康等级(EXCELLENT/GOOD/FAIR/POOR)',
                                       requiresAttentionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即关注(Y是,N否)',
                                       potentialIssuesJson LONGTEXT DEFAULT NULL COMMENT '潜在问题列表，JSON格式',

    -- 时间字段
                                       collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 通用字段
                                       addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                       addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                       editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                       editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                       oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                       currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                       activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                       noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                       PRIMARY KEY (tenantId, classLoadingId),
                                       KEY IDX_MONITOR_CLS_RES (jvmResourceId),
                                       KEY IDX_MONITOR_CLS_TIME (collectionTime),
                                       KEY IDX_MONITOR_CLS_HEALTH (healthyFlag, requiresAttentionFlag),
                                       KEY IDX_MONITOR_CLS_COUNT (loadedClassCount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM类加载监控表';

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
--   └── HUB_MONITOR_JVM_CLASS (1:1，一个JVM资源对应一个类加载信息记录)
-- ==========================================
-- 9. 应用监控数据表
-- 存储应用层面的各种监控数据（线程池、连接池、自定义指标等）
-- 对应 ThirdPartyMonitorData 采集的所有监控数据
-- ==========================================
CREATE TABLE HUB_MONITOR_APP_DATA (
                                      appDataId VARCHAR(32) NOT NULL COMMENT '应用监控数据ID，主键',
                                      tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                      jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 数据分类标识
                                      dataType VARCHAR(50) NOT NULL COMMENT '数据类型(THREAD_POOL:线程池/CONNECTION_POOL:连接池/CUSTOM_METRIC:自定义指标/CACHE_POOL:缓存池/MESSAGE_QUEUE:消息队列)',
                                      dataName VARCHAR(100) NOT NULL COMMENT '数据名称（如：线程池名称、指标名称等）',
                                      dataCategory VARCHAR(50) DEFAULT NULL COMMENT '数据分类（如：业务线程池/IO线程池/业务指标/技术指标）',

    -- 监控数据（JSON格式存储，支持不同类型的数据结构）
                                      dataJson LONGTEXT NOT NULL COMMENT '监控数据，JSON格式，包含具体的监控指标和值',

    -- 核心指标（从JSON中提取的关键指标，便于查询和索引）
                                      primaryValue DECIMAL(20,4) DEFAULT NULL COMMENT '主要指标值（如：使用率、数量等）',
                                      secondaryValue DECIMAL(20,4) DEFAULT NULL COMMENT '次要指标值（如：最大值、平均值等）',
                                      statusValue VARCHAR(50) DEFAULT NULL COMMENT '状态值（如：健康状态、连接状态等）',

    -- 健康状态
                                      healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '健康标记(Y健康,N异常)',
                                      healthGrade VARCHAR(20) DEFAULT NULL COMMENT '健康等级(EXCELLENT/GOOD/FAIR/POOR/CRITICAL)',
                                      requiresAttentionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即关注(Y是,N否)',

    -- 标签和维度（便于分组查询）
                                      tagsJson TEXT DEFAULT NULL COMMENT '标签信息，JSON格式（如：{"poolType":"business","environment":"prod"}）',

    -- 时间字段
                                      collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 通用字段
                                      addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                      addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                      editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                      editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                      oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                      currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                      activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                      noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                      PRIMARY KEY (tenantId, appDataId),
                                      KEY IDX_MONITOR_APP_DATA_RES (jvmResourceId),
                                      KEY IDX_MONITOR_APP_DATA_TYPE (dataType),
                                      KEY IDX_MONITOR_APP_DATA_NAME (dataName),
                                      KEY IDX_MONITOR_APP_DATA_TIME (collectionTime),
                                      KEY IDX_MONITOR_APP_DATA_HEALTH (healthyFlag, requiresAttentionFlag),
                                      KEY IDX_MONITOR_APP_DATA_PRIMARY (primaryValue),
                                      KEY IDX_MONITOR_APP_DATA_STATUS (statusValue),
                                      KEY IDX_MONITOR_APP_DATA_RES_TYPE_NAME (jvmResourceId, dataType, dataName, collectionTime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用监控数据表';
-- =========================================
-- 基于FRP架构的隧道管理系统 - 数据库表结构设计
-- 参考FRP（Fast Reverse Proxy）设计模式
-- 遵循naming-convention.md数据库规范
-- =========================================

-- =========================================
-- 1. 隧道服务器表（控制端口）
-- =========================================

-- 隧道服务器配置表 - 管理控制端口和核心配置
CREATE TABLE `HUB_TUNNEL_SERVER` (
                                     `tunnelServerId` VARCHAR(32) NOT NULL COMMENT '隧道服务器ID，主键',
                                     `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
                                     `serverName` VARCHAR(100) NOT NULL COMMENT '服务器名称',
                                     `serverDescription` VARCHAR(500) DEFAULT NULL COMMENT '服务器描述',
                                     `controlAddress` VARCHAR(50) NOT NULL DEFAULT '0.0.0.0' COMMENT '控制端口监听地址',
                                     `controlPort` INT NOT NULL DEFAULT 7000 COMMENT '控制端口（接受客户端连接）',
                                     `dashboardPort` INT DEFAULT 7500 COMMENT '管理面板端口',
                                     `vhostHttpPort` INT DEFAULT 80 COMMENT '虚拟主机HTTP端口',
                                     `vhostHttpsPort` INT DEFAULT 443 COMMENT '虚拟主机HTTPS端口',
                                     `maxClients` INT NOT NULL DEFAULT 1000 COMMENT '最大客户端连接数',
                                     `tokenAuth` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '启用Token认证(N禁用,Y启用)',
                                     `authToken` VARCHAR(100) DEFAULT NULL COMMENT '客户端认证Token',
                                     `tlsEnable` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT 'TLS启用状态(N禁用,Y启用)',
                                     `tlsCertFile` VARCHAR(255) DEFAULT NULL COMMENT 'TLS证书文件路径',
                                     `tlsKeyFile` VARCHAR(255) DEFAULT NULL COMMENT 'TLS私钥文件路径',
                                     `heartbeatInterval` INT NOT NULL DEFAULT 30 COMMENT '心跳间隔(秒)',
                                     `heartbeatTimeout` INT NOT NULL DEFAULT 90 COMMENT '心跳超时(秒)',
                                     `logLevel` VARCHAR(10) NOT NULL DEFAULT 'info' COMMENT '日志级别(debug,info,warn,error)',
                                     `maxPortsPerClient` INT DEFAULT 10 COMMENT '每个客户端最大端口数',
                                     `allowPorts` TEXT DEFAULT NULL COMMENT '允许的端口范围，JSON格式',
                                     `serverStatus` VARCHAR(20) NOT NULL DEFAULT 'stopped' COMMENT '服务器状态(running,stopped,error)',
                                     `startTime` DATETIME DEFAULT NULL COMMENT '服务启动时间',
                                     `configVersion` VARCHAR(32) DEFAULT NULL COMMENT '配置版本号',
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
                                     CONSTRAINT `PK_TUNNEL_SERVER` PRIMARY KEY (`tunnelServerId`),
                                     UNIQUE KEY `IDX_TUNNEL_SVR_NAME` (`serverName`),
                                     KEY `IDX_TUNNEL_SVR_TENANT` (`tenantId`),
                                     KEY `IDX_TUNNEL_SVR_CTRL` (`controlAddress`, `controlPort`),
                                     KEY `IDX_TUNNEL_SVR_STATUS` (`serverStatus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='隧道服务器配置表，管理控制端口和核心配置';

-- =========================================
-- 2. 服务器节点表（静态端口映射）
-- =========================================

-- 隧道服务器节点表 - 管理静态端口映射和数据端口
CREATE TABLE `HUB_TUNNEL_SERVER_NODE` (
                                          `serverNodeId` VARCHAR(32) NOT NULL COMMENT '服务器节点ID，主键',
                                          `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
                                          `tunnelServerId` VARCHAR(32) NOT NULL COMMENT '隧道服务器ID，关联HUB_TUNNEL_SERVER',
                                          `nodeName` VARCHAR(100) NOT NULL COMMENT '节点名称',
                                          `nodeType` VARCHAR(20) NOT NULL DEFAULT 'static' COMMENT '节点类型(static,dynamic)',
                                          `proxyType` VARCHAR(20) NOT NULL COMMENT '代理类型(tcp,udp,http,https,stcp,sudp)',
                                          `listenAddress` VARCHAR(50) NOT NULL DEFAULT '0.0.0.0' COMMENT '监听地址',
                                          `listenPort` INT NOT NULL COMMENT '监听端口（公网端口）',
                                          `targetAddress` VARCHAR(50) NOT NULL COMMENT '目标地址（内网地址）',
                                          `targetPort` INT NOT NULL COMMENT '目标端口（内网端口）',
                                          `customDomains` TEXT DEFAULT NULL COMMENT '自定义域名列表，JSON格式',
                                          `subDomain` VARCHAR(100) DEFAULT NULL COMMENT '子域名',
                                          `httpUser` VARCHAR(50) DEFAULT NULL COMMENT 'HTTP基础认证用户名',
                                          `httpPassword` VARCHAR(100) DEFAULT NULL COMMENT 'HTTP基础认证密码',
                                          `hostHeaderRewrite` VARCHAR(255) DEFAULT NULL COMMENT '重写Host头',
                                          `headers` TEXT DEFAULT NULL COMMENT '自定义HTTP头，JSON格式',
                                          `locations` TEXT DEFAULT NULL COMMENT 'HTTP路径配置，JSON格式',
                                          `compression` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '启用压缩(N禁用,Y启用)',
                                          `encryption` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '启用加密(N禁用,Y启用)',
                                          `secretKey` VARCHAR(100) DEFAULT NULL COMMENT '加密密钥',
                                          `healthCheckType` VARCHAR(20) DEFAULT 'tcp' COMMENT '健康检查类型(tcp,http)',
                                          `healthCheckUrl` VARCHAR(255) DEFAULT NULL COMMENT '健康检查URL',
                                          `healthCheckInterval` INT DEFAULT 60 COMMENT '健康检查间隔(秒)',
                                          `maxConnections` INT DEFAULT 100 COMMENT '最大连接数',
                                          `nodeStatus` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '节点状态(active,inactive,error)',
                                          `lastHealthCheck` DATETIME DEFAULT NULL COMMENT '最后健康检查时间',
                                          `connectionCount` INT DEFAULT 0 COMMENT '当前连接数',
                                          `totalConnections` BIGINT DEFAULT 0 COMMENT '总连接数',
                                          `totalBytes` BIGINT DEFAULT 0 COMMENT '总传输字节数',
                                          `createdTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '节点创建时间',
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
                                          CONSTRAINT `PK_TUNNEL_SVR_NODE` PRIMARY KEY (`serverNodeId`),
                                          UNIQUE KEY `IDX_TUNNEL_NODE_NAME` (`nodeName`),
                                          UNIQUE KEY `IDX_TUNNEL_NODE_PORT` (`listenAddress`, `listenPort`, `proxyType`),
                                          KEY `IDX_TUNNEL_NODE_TENANT` (`tenantId`),
                                          KEY `IDX_TUNNEL_NODE_SERVER` (`tunnelServerId`),
                                          KEY `IDX_TUNNEL_NODE_TYPE` (`nodeType`, `proxyType`),
                                          KEY `IDX_TUNNEL_NODE_STATUS` (`nodeStatus`),
                                          KEY `IDX_TUNNEL_NODE_HEALTH` (`lastHealthCheck`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='隧道服务器节点表，管理静态端口映射和数据端口转发';

-- =========================================
-- 3. 客户端注册表（动态连接）
-- =========================================

-- 隧道客户端表 - 管理客户端连接和认证
CREATE TABLE `HUB_TUNNEL_CLIENT` (
                                     `tunnelClientId` VARCHAR(32) NOT NULL COMMENT '隧道客户端ID，主键',
                                     `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
                                     `userId` VARCHAR(32) NOT NULL COMMENT '用户ID，关联外部用户系统',
                                     `clientName` VARCHAR(100) NOT NULL COMMENT '客户端名称',
                                     `clientDescription` VARCHAR(500) DEFAULT NULL COMMENT '客户端描述',
                                     `clientVersion` VARCHAR(20) DEFAULT NULL COMMENT '客户端版本',
                                     `operatingSystem` VARCHAR(50) DEFAULT NULL COMMENT '操作系统',
                                     `clientIpAddress` VARCHAR(50) DEFAULT NULL COMMENT '客户端IP地址',
                                     `clientMacAddress` VARCHAR(20) DEFAULT NULL COMMENT '客户端MAC地址',
                                     `serverAddress` VARCHAR(100) NOT NULL COMMENT '服务器地址',
                                     `serverPort` INT NOT NULL DEFAULT 7000 COMMENT '服务器控制端口',
                                     `authToken` VARCHAR(100) NOT NULL COMMENT '认证令牌',
                                     `tlsEnable` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '启用TLS(N禁用,Y启用)',
                                     `autoReconnect` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '自动重连(N禁用,Y启用)',
                                     `maxRetries` INT NOT NULL DEFAULT 5 COMMENT '最大重试次数',
                                     `retryInterval` INT NOT NULL DEFAULT 20 COMMENT '重试间隔(秒)',
                                     `heartbeatInterval` INT NOT NULL DEFAULT 30 COMMENT '心跳间隔(秒)',
                                     `heartbeatTimeout` INT NOT NULL DEFAULT 90 COMMENT '心跳超时(秒)',
                                     `connectionStatus` VARCHAR(20) NOT NULL DEFAULT 'disconnected' COMMENT '连接状态(connected,disconnected,connecting,error)',
                                     `lastConnectTime` DATETIME DEFAULT NULL COMMENT '最后连接时间',
                                     `lastDisconnectTime` DATETIME DEFAULT NULL COMMENT '最后断开时间',
                                     `totalConnectTime` BIGINT DEFAULT 0 COMMENT '总连接时长(秒)',
                                     `reconnectCount` INT DEFAULT 0 COMMENT '重连次数',
                                     `serviceCount` INT DEFAULT 0 COMMENT '注册的服务数量',
                                     `lastHeartbeat` DATETIME DEFAULT NULL COMMENT '最后心跳时间',
                                     `clientConfig` TEXT DEFAULT NULL COMMENT '客户端配置，JSON格式',
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
                                     CONSTRAINT `PK_TUNNEL_CLIENT` PRIMARY KEY (`tunnelClientId`),
                                     UNIQUE KEY `IDX_TUNNEL_CLIENT_NAME` (`clientName`),
                                     KEY `IDX_TUNNEL_CLIENT_TENANT` (`tenantId`),
                                     KEY `IDX_TUNNEL_CLIENT_USER` (`userId`),
                                     KEY `IDX_TUNNEL_CLIENT_STATUS` (`connectionStatus`),
                                     KEY `IDX_TUNNEL_CLIENT_IP` (`clientIpAddress`),
                                     KEY `IDX_TUNNEL_CLIENT_HB` (`lastHeartbeat`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='隧道客户端表，管理客户端连接认证和状态跟踪';

-- =========================================
-- 4. 服务配置表（动态注册的服务）
-- =========================================

-- 隧道服务配置表 - 管理客户端动态注册的服务
CREATE TABLE `HUB_TUNNEL_SERVICE` (
                                      `tunnelServiceId` VARCHAR(32) NOT NULL COMMENT '隧道服务ID，主键',
                                      `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
                                      `tunnelClientId` VARCHAR(32) NOT NULL COMMENT '隧道客户端ID，关联HUB_TUNNEL_CLIENT',
                                      `userId` VARCHAR(32) NOT NULL COMMENT '用户ID，关联外部用户系统',
                                      `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称',
                                      `serviceDescription` VARCHAR(500) DEFAULT NULL COMMENT '服务描述',
                                      `serviceType` VARCHAR(20) NOT NULL COMMENT '服务类型(tcp,udp,http,https,stcp,sudp,xtcp)',
                                      `localAddress` VARCHAR(50) NOT NULL DEFAULT '127.0.0.1' COMMENT '本地地址',
                                      `localPort` INT NOT NULL COMMENT '本地端口',
                                      `remotePort` INT DEFAULT NULL COMMENT '远程端口（服务端分配）',
                                      `customDomains` TEXT DEFAULT NULL COMMENT '自定义域名列表，JSON格式',
                                      `subDomain` VARCHAR(100) DEFAULT NULL COMMENT '子域名前缀',
                                      `httpUser` VARCHAR(50) DEFAULT NULL COMMENT 'HTTP基础认证用户名',
                                      `httpPassword` VARCHAR(100) DEFAULT NULL COMMENT 'HTTP基础认证密码',
                                      `hostHeaderRewrite` VARCHAR(255) DEFAULT NULL COMMENT '重写Host头',
                                      `headers` TEXT DEFAULT NULL COMMENT '自定义HTTP头，JSON格式',
                                      `locations` TEXT DEFAULT NULL COMMENT 'HTTP路径配置，JSON格式',
                                      `useEncryption` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '使用加密(N禁用,Y启用)',
                                      `useCompression` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '使用压缩(N禁用,Y启用)',
                                      `secretKey` VARCHAR(100) DEFAULT NULL COMMENT '加密密钥',
                                      `bandwidthLimit` VARCHAR(20) DEFAULT NULL COMMENT '带宽限制',
                                      `maxConnections` INT DEFAULT 100 COMMENT '最大连接数限制',
                                      `healthCheckType` VARCHAR(20) DEFAULT NULL COMMENT '健康检查类型(tcp,http)',
                                      `healthCheckUrl` VARCHAR(255) DEFAULT NULL COMMENT '健康检查URL',
                                      `serviceStatus` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '服务状态(active,inactive,error,offline)',
                                      `registeredTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '服务注册时间',
                                      `lastActiveTime` DATETIME DEFAULT NULL COMMENT '最后活跃时间',
                                      `connectionCount` INT DEFAULT 0 COMMENT '当前连接数',
                                      `totalConnections` BIGINT DEFAULT 0 COMMENT '总连接数',
                                      `totalTraffic` BIGINT DEFAULT 0 COMMENT '总流量(字节)',
                                      `serviceConfig` TEXT DEFAULT NULL COMMENT '服务配置，JSON格式',
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
                                      CONSTRAINT `PK_TUNNEL_SERVICE` PRIMARY KEY (`tunnelServiceId`),
                                      UNIQUE KEY `IDX_TUNNEL_SVC_NAME` (`serviceName`),
                                      KEY `IDX_TUNNEL_SVC_TENANT` (`tenantId`),
                                      KEY `IDX_TUNNEL_SVC_CLIENT` (`tunnelClientId`),
                                      KEY `IDX_TUNNEL_SVC_USER` (`userId`),
                                      KEY `IDX_TUNNEL_SVC_TYPE` (`serviceType`),
                                      KEY `IDX_TUNNEL_SVC_STATUS` (`serviceStatus`),
                                      KEY `IDX_TUNNEL_SVC_PORT` (`remotePort`),
                                      KEY `IDX_TUNNEL_SVC_DOMAIN` (`subDomain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='隧道服务配置表，管理客户端动态注册的服务配置';
