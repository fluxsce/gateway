CREATE TABLE HUB_USER (
                          userId          VARCHAR2(32)   NOT NULL, -- 用户ID，联合主键
                          tenantId        VARCHAR2(32)   NOT NULL,              -- 租户ID，联合主键
                          userName        VARCHAR2(50)   NOT NULL,              -- 用户名，登录账号
                          password        VARCHAR2(128)  NOT NULL,              -- 密码，加密存储
                          realName        VARCHAR2(50)   NOT NULL,              -- 真实姓名
                          deptId          VARCHAR2(32)   NOT NULL,              -- 所属部门ID
                          email           VARCHAR2(255),                         -- 电子邮箱
                          mobile          VARCHAR2(20),                          -- 手机号码
                          avatar          VARCHAR2(500),                         -- 头像URL
                          gender          NUMBER(10),                            -- 性别：1-男，2-女，0-未知
                          statusFlag      CHAR(1)        DEFAULT 'Y' NOT NULL,  -- 状态：Y-启用，N-禁用
                          deptAdminFlag   CHAR(1)        DEFAULT 'N' NOT NULL,  -- 是否部门管理员：Y-是，N-否
                          tenantAdminFlag CHAR(1)        DEFAULT 'N' NOT NULL,  -- 是否租户管理员：Y-是，N-否
                          userExpireDate  DATE           NOT NULL,              -- 用户过期时间
                          lastLoginTime   DATE,                                  -- 最后登录时间
                          lastLoginIp     VARCHAR2(128),                         -- 最后登录IP
                          pwdUpdateTime   DATE,                                  -- 密码最后更新时间
                          pwdErrorCount   NUMBER(10)     DEFAULT 0 NOT NULL,    -- 密码错误次数
                          lockTime        DATE,                                  -- 账号锁定时间
                          addTime         DATE           DEFAULT SYSDATE NOT NULL, -- 创建时间
                          addWho          VARCHAR2(32)   DEFAULT 'system' NOT NULL, -- 创建人
                          editTime        DATE           DEFAULT SYSDATE NOT NULL, -- 修改时间
                          editWho         VARCHAR2(32)   DEFAULT 'system' NOT NULL, -- 修改人
                          oprSeqFlag      VARCHAR2(32)   NOT NULL,              -- 操作序列标识
                          currentVersion  NUMBER(10)     DEFAULT 1 NOT NULL,    -- 当前版本号
                          activeFlag      CHAR(1)        DEFAULT 'Y' NOT NULL,  -- 活动状态标记：Y-活动，N-非活动
                          noteText        CLOB,                                  -- 备注信息
                          extProperty     CLOB,                                  -- 扩展属性，JSON格式
                          reserved1       VARCHAR2(500),                         -- 预留字段1
                          reserved2       VARCHAR2(500),                         -- 预留字段2
                          reserved3       VARCHAR2(500),                         -- 预留字段3
                          reserved4       VARCHAR2(500),                         -- 预留字段4
                          reserved5       VARCHAR2(500),                         -- 预留字段5
                          reserved6       VARCHAR2(500),                         -- 预留字段6
                          reserved7       VARCHAR2(500),                         -- 预留字段7
                          reserved8       VARCHAR2(500),                         -- 预留字段8
                          reserved9       VARCHAR2(500),                         -- 预留字段9
                          reserved10      VARCHAR2(500),                         -- 预留字段10
                          CONSTRAINT PK_USER PRIMARY KEY (tenantId,userId)
);

-- 创建索引
CREATE INDEX IDX_USER_TENANT ON HUB_USER(tenantId);
CREATE INDEX IDX_USER_DEPT ON HUB_USER(deptId);
CREATE INDEX IDX_USER_STATUS ON HUB_USER(statusFlag);
CREATE INDEX IDX_USER_EMAIL ON HUB_USER(email);
CREATE INDEX IDX_USER_MOBILE ON HUB_USER(mobile);
COMMENT ON TABLE HUB_USER IS '用户信息表';

INSERT INTO HUB_USER (userId,
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
                      noteText)
VALUES ('admin', -- userId
        'default', -- tenantId
        'admin', -- userName
        '123456', -- password（MD5("123456") 示例）
        '系统管理员', -- realName
        'D00000001', -- deptId
        'admin@example.com', -- email
        '13800000000', -- mobile
        'https://example.com/avatar.png', -- avatar
        1, -- gender (1:男)
        'Y', -- statusFlag
        'N', -- deptAdminFlag
        'Y', -- tenantAdminFlag
        SYSDATE + 365 * 5, -- userExpireDate（5年后过期）
        'SEQFLAG_001', -- oprSeqFlag
        1, -- currentVersion
        'Y', -- activeFlag
        '系统初始化管理员账号' -- noteText
       );


CREATE TABLE HUB_LOGIN_LOG (
                               logId           VARCHAR2(32)   NOT NULL,
                               userId          VARCHAR2(32)   NOT NULL,
                               tenantId        VARCHAR2(32)   NOT NULL,
                               userName        VARCHAR2(50)   NOT NULL,
                               loginTime       DATE           DEFAULT SYSDATE NOT NULL,
                               loginIp         VARCHAR2(128)  DEFAULT '0.0.0.0' NOT NULL,
                               loginLocation   VARCHAR2(255),
                               loginType       NUMBER(10)     DEFAULT 1 NOT NULL,
                               deviceType      VARCHAR2(50),
                               deviceInfo      CLOB,
                               browserInfo     CLOB,
                               osInfo          VARCHAR2(255),
                               loginStatus     CHAR(1)        DEFAULT 'N' NOT NULL,
                               logoutTime      DATE,
                               sessionDuration NUMBER(10),
                               failReason      CLOB,
                               addTime         DATE           DEFAULT SYSDATE NOT NULL,
                               addWho          VARCHAR2(32)   DEFAULT 'system' NOT NULL,
                               editTime        DATE           DEFAULT SYSDATE NOT NULL,
                               editWho         VARCHAR2(32)   DEFAULT 'system' NOT NULL,
                               oprSeqFlag      VARCHAR2(32)   NOT NULL,
                               currentVersion  NUMBER(10)     DEFAULT 1 NOT NULL,
                               activeFlag      CHAR(1)        DEFAULT 'Y' NOT NULL,
                               CONSTRAINT PK_LOGIN_LOG PRIMARY KEY (logId)
);

-- 创建索引
CREATE INDEX IDX_LOGIN_USER     ON HUB_LOGIN_LOG(userId);
CREATE INDEX IDX_LOGIN_TIME     ON HUB_LOGIN_LOG(loginTime);
CREATE INDEX IDX_LOGIN_TENANT   ON HUB_LOGIN_LOG(tenantId);
CREATE INDEX IDX_LOGIN_STATUS   ON HUB_LOGIN_LOG(loginStatus);
CREATE INDEX IDX_LOGIN_TYPE     ON HUB_LOGIN_LOG(loginType);
COMMENT ON TABLE HUB_LOGIN_LOG IS '用户登录日志表';

CREATE TABLE HUB_GW_INSTANCE (
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

                                      CONSTRAINT PK_GW_INSTANCE PRIMARY KEY (tenantId, gatewayInstanceId)
);
-- 然后使用以下语句创建索引
CREATE INDEX IDX_GW_INST_BIND_HTTP ON HUB_GW_INSTANCE(bindAddress, httpPort);
CREATE INDEX IDX_GW_INST_BIND_HTTPS ON HUB_GW_INSTANCE(bindAddress, httpsPort);
CREATE INDEX IDX_GW_INST_LOG ON HUB_GW_INSTANCE(logConfigId);
CREATE INDEX IDX_GW_INST_HEALTH ON HUB_GW_INSTANCE(healthStatus);
CREATE INDEX IDX_GW_INST_TLS ON HUB_GW_INSTANCE(tlsEnabled);
-- Oracle 不直接支持在DDL中指定表级注释，需要使用单独的COMMENT ON语句。
COMMENT ON TABLE HUB_GW_INSTANCE IS '网关实例表 - 记录网关实例基础配置，完整支持Go HTTP Server配置';

CREATE TABLE HUB_GW_ROUTER_CONFIG (
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

                                           CONSTRAINT PK_GW_ROUTER_CONFIG PRIMARY KEY (tenantId, routerConfigId)
);
CREATE INDEX IDX_GW_ROUTER_INST ON HUB_GW_ROUTER_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_ROUTER_NAME ON HUB_GW_ROUTER_CONFIG(routerName);
CREATE INDEX IDX_GW_ROUTER_ACTIVE ON HUB_GW_ROUTER_CONFIG(activeFlag);
CREATE INDEX IDX_GW_ROUTER_CACHE ON HUB_GW_ROUTER_CONFIG(enableRouteCache);
COMMENT ON TABLE HUB_GW_ROUTER_CONFIG IS 'Router配置表 - 存储网关Router级别配置';

CREATE TABLE HUB_GW_ROUTE_CONFIG (
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

  CONSTRAINT PK_GW_ROUTE_CONFIG PRIMARY KEY (tenantId, routeConfigId)
);
CREATE INDEX IDX_GW_ROUTE_INST ON HUB_GW_ROUTE_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_ROUTE_SERVICE ON HUB_GW_ROUTE_CONFIG(serviceDefinitionId);
CREATE INDEX IDX_GW_ROUTE_LOG ON HUB_GW_ROUTE_CONFIG(logConfigId);
CREATE INDEX IDX_GW_ROUTE_PRIORITY ON HUB_GW_ROUTE_CONFIG(routePriority);
CREATE INDEX IDX_GW_ROUTE_PATH ON HUB_GW_ROUTE_CONFIG(routePath);
CREATE INDEX IDX_GW_ROUTE_ACTIVE ON HUB_GW_ROUTE_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GW_ROUTE_CONFIG IS '路由定义表 - 存储API路由配置,使用activeFlag统一管理启用状态';

CREATE TABLE HUB_GW_ROUTE_ASSERTION (
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

                                             CONSTRAINT PK_GW_ROUTE_ASSERTION PRIMARY KEY (tenantId, routeAssertionId)
);
CREATE INDEX IDX_GW_ASSERT_ROUTE ON HUB_GW_ROUTE_ASSERTION(routeConfigId);
CREATE INDEX IDX_GW_ASSERT_TYPE ON HUB_GW_ROUTE_ASSERTION(assertionType);
CREATE INDEX IDX_GW_ASSERT_ORDER ON HUB_GW_ROUTE_ASSERTION(assertionOrder);
COMMENT ON TABLE HUB_GW_ROUTE_ASSERTION IS '路由断言表 - 存储路由匹配断言规则';

CREATE TABLE HUB_GW_FILTER_CONFIG (
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

                                           CONSTRAINT PK_GW_FILTER_CONFIG PRIMARY KEY (tenantId, filterConfigId)
);
CREATE INDEX IDX_GW_FILTER_INST ON HUB_GW_FILTER_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_FILTER_ROUTE ON HUB_GW_FILTER_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_FILTER_TYPE ON HUB_GW_FILTER_CONFIG(filterType);
CREATE INDEX IDX_GW_FILTER_ACTION ON HUB_GW_FILTER_CONFIG(filterAction);
CREATE INDEX IDX_GW_FILTER_ORDER ON HUB_GW_FILTER_CONFIG(filterOrder);
CREATE INDEX IDX_GW_FILTER_ACTIVE ON HUB_GW_FILTER_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GW_FILTER_CONFIG IS '过滤器配置表 - 根据filter.go逻辑设计,支持7种类型和3种执行时机';

CREATE TABLE HUB_GW_RATE_LIMIT_CONFIG (
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
                                               CONSTRAINT PK_GW_RATE_LIMIT_CONFIG PRIMARY KEY (tenantId, rateLimitConfigId)
);
CREATE INDEX IDX_GW_RATE_INST ON HUB_GW_RATE_LIMIT_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_RATE_ROUTE ON HUB_GW_RATE_LIMIT_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_RATE_STRATEGY ON HUB_GW_RATE_LIMIT_CONFIG(keyStrategy);
CREATE INDEX IDX_GW_RATE_ALGORITHM ON HUB_GW_RATE_LIMIT_CONFIG(algorithm);
CREATE INDEX IDX_GW_RATE_PRIORITY ON HUB_GW_RATE_LIMIT_CONFIG(configPriority);
CREATE INDEX IDX_GW_RATE_ACTIVE ON HUB_GW_RATE_LIMIT_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GW_RATE_LIMIT_CONFIG IS '限流配置表 - 存储流量限制规则';


CREATE TABLE HUB_GW_CIRCUIT_BREAKER_CONFIG (
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

                                                    CONSTRAINT PK_GW_CIRCUIT_BREAKER_CONFIG PRIMARY KEY (tenantId, circuitBreakerConfigId)
);
CREATE INDEX IDX_GW_CB_ROUTE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_CB_SERVICE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(targetServiceId);
CREATE INDEX IDX_GW_CB_STRATEGY ON HUB_GW_CIRCUIT_BREAKER_CONFIG(keyStrategy);
CREATE INDEX IDX_GW_CB_STORAGE ON HUB_GW_CIRCUIT_BREAKER_CONFIG(storageType);
CREATE INDEX IDX_GW_CB_PRIORITY ON HUB_GW_CIRCUIT_BREAKER_CONFIG(configPriority);
COMMENT ON TABLE HUB_GW_CIRCUIT_BREAKER_CONFIG IS '熔断配置表 - 根据CircuitBreakerConfig结构设计,支持完整的熔断策略配置';

CREATE TABLE HUB_GW_AUTH_CONFIG (
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

                                         CONSTRAINT PK_GW_AUTH_CONFIG PRIMARY KEY (tenantId, authConfigId)
);
CREATE INDEX IDX_GW_AUTH_INST ON HUB_GW_AUTH_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_AUTH_ROUTE ON HUB_GW_AUTH_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_AUTH_TYPE ON HUB_GW_AUTH_CONFIG(authType);
CREATE INDEX IDX_GW_AUTH_PRIORITY ON HUB_GW_AUTH_CONFIG(configPriority);
COMMENT ON TABLE HUB_GW_AUTH_CONFIG IS '认证配置表 - 存储认证相关规则';

CREATE TABLE HUB_GW_SERVICE_DEFINITION (
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

                                                CONSTRAINT PK_GW_SERVICE_DEFINITION PRIMARY KEY (tenantId, serviceDefinitionId)
);
CREATE INDEX IDX_GW_SVC_NAME ON HUB_GW_SERVICE_DEFINITION(serviceName);
CREATE INDEX IDX_GW_SVC_TYPE ON HUB_GW_SERVICE_DEFINITION(serviceType);
CREATE INDEX IDX_GW_SVC_STRATEGY ON HUB_GW_SERVICE_DEFINITION(loadBalanceStrategy);
CREATE INDEX IDX_GW_SVC_HEALTH ON HUB_GW_SERVICE_DEFINITION(healthCheckEnabled);
CREATE INDEX IDX_GW_SVC_PROXY ON HUB_GW_SERVICE_DEFINITION(proxyConfigId);
COMMENT ON TABLE HUB_GW_SERVICE_DEFINITION IS '服务定义表 - 根据ServiceConfig结构设计,存储完整的服务配置';

CREATE TABLE HUB_GW_SERVICE_NODE (
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

                                          CONSTRAINT PK_GW_SERVICE_NODE PRIMARY KEY (tenantId, serviceNodeId)
);
CREATE INDEX IDX_GW_NODE_SERVICE ON HUB_GW_SERVICE_NODE(serviceDefinitionId);
CREATE INDEX IDX_GW_NODE_ENDPOINT ON HUB_GW_SERVICE_NODE(nodeHost, nodePort);
CREATE INDEX IDX_GW_NODE_HEALTH ON HUB_GW_SERVICE_NODE(healthStatus);
CREATE INDEX IDX_GW_NODE_STATUS ON HUB_GW_SERVICE_NODE(nodeStatus);
COMMENT ON TABLE HUB_GW_SERVICE_NODE IS '服务节点表 - 根据NodeConfig结构设计,存储服务节点实例信息';

CREATE TABLE HUB_GW_PROXY_CONFIG (
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

                                          CONSTRAINT PK_GW_PROXY_CONFIG PRIMARY KEY (tenantId, proxyConfigId)
);
CREATE INDEX IDX_GW_PROXY_INST ON HUB_GW_PROXY_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_PROXY_TYPE ON HUB_GW_PROXY_CONFIG(proxyType);
CREATE INDEX IDX_GW_PROXY_PRIORITY ON HUB_GW_PROXY_CONFIG(configPriority);
CREATE INDEX IDX_GW_PROXY_ACTIVE ON HUB_GW_PROXY_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GW_PROXY_CONFIG IS '代理配置表 - 根据proxy.go逻辑设计,仅支持实例级代理配置';

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

                                     CONSTRAINT PK_TIMER_SCHEDULER PRIMARY KEY (tenantId, schedulerId)
);
CREATE INDEX IDX_TIMER_SCHED_NAME ON HUB_TIMER_SCHEDULER(schedulerName);
CREATE INDEX IDX_TIMER_SCHED_INST ON HUB_TIMER_SCHEDULER(schedulerInstanceId);
CREATE INDEX IDX_TIMER_SCHED_STATUS ON HUB_TIMER_SCHEDULER(schedulerStatus);
CREATE INDEX IDX_TIMER_SCHED_HEART ON HUB_TIMER_SCHEDULER(lastHeartbeatTime);
COMMENT ON TABLE HUB_TIMER_SCHEDULER IS '定时任务调度器表 - 存储调度器配置和状态信息';

CREATE TABLE HUB_TIMER_TASK (
                                taskId                  VARCHAR2(32) NOT NULL, -- 任务ID，主键
                                tenantId                VARCHAR2(32) NOT NULL, -- 租户ID

                                taskName                VARCHAR2(200) NOT NULL, -- 任务名称
                                taskDescription         VARCHAR2(500), -- 任务描述
                                taskPriority            NUMBER(10) DEFAULT 1 NOT NULL, -- 任务优先级(1低,2普通,3高)
                                schedulerId             VARCHAR2(32), -- 关联的调度器ID
                                schedulerName           VARCHAR2(100), -- 调度器名称（冗余字段）

                                scheduleType            NUMBER(10) NOT NULL, -- 调度类型(1一次性,2固定间隔,3Cron,4延迟执行,5实时执行)
                                cronExpression          VARCHAR2(100), -- Cron表达式（scheduleType=3时必填）
                                intervalSeconds         NUMBER(20), -- 执行间隔秒数（scheduleType=2时必填）
                                delaySeconds            NUMBER(20), -- 延迟秒数（scheduleType=4时必填）
                                startTime               DATE, -- 任务开始时间
                                endTime                 DATE, -- 任务结束时间

                                maxRetries              NUMBER(10) DEFAULT 0 NOT NULL, -- 最大重试次数
                                retryIntervalSeconds    NUMBER(20) DEFAULT 60 NOT NULL, -- 重试间隔秒数
                                timeoutSeconds          NUMBER(20) DEFAULT 1800 NOT NULL, -- 执行超时时间秒数
                                taskParams              CLOB, -- 任务参数，JSON格式存储

    -- 新增字段：任务执行器配置
                                executorType            VARCHAR2(50), -- 执行器类型(BUILTIN内置,SFTP,SSH,DATABASE,HTTP等)
                                toolConfigId            VARCHAR2(32), -- 工具配置ID（如SFTP配置ID、数据库配置ID等）
                                toolConfigName          VARCHAR2(100), -- 工具配置名称（冗余字段）
                                operationType           VARCHAR2(100), -- 执行操作类型（如文件上传、下载、SQL执行、接口调用等）
                                operationConfig         CLOB, -- 操作参数配置，JSON格式存储具体操作的参数

                                taskStatus              NUMBER(10) DEFAULT 1 NOT NULL, -- 任务状态(1待执行,2运行中,3已完成,4失败,5取消)
                                nextRunTime             DATE, -- 下次执行时间
                                lastRunTime             DATE, -- 上次执行时间
                                runCount                NUMBER(20) DEFAULT 0 NOT NULL, -- 执行总次数
                                successCount            NUMBER(20) DEFAULT 0 NOT NULL, -- 成功次数
                                failureCount            NUMBER(20) DEFAULT 0 NOT NULL, -- 失败次数

                                lastExecutionId         VARCHAR2(32), -- 最后执行ID
                                lastExecutionStartTime  DATE, -- 最后执行开始时间
                                lastExecutionEndTime    DATE, -- 最后执行结束时间
                                lastExecutionDurationMs NUMBER(20), -- 最后执行耗时毫秒数
                                lastExecutionStatus     NUMBER(10), -- 最后执行状态
                                lastResultSuccess       VARCHAR2(1), -- 最后执行是否成功(N失败,Y成功)
                                lastErrorMessage        CLOB, -- 最后错误信息
                                lastRetryCount          NUMBER(10), -- 最后重试次数

                                addTime                 DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                addWho                  VARCHAR2(32) NOT NULL, -- 创建人ID
                                editTime                DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                editWho                 VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                oprSeqFlag              VARCHAR2(32) NOT NULL, -- 操作序列标识
                                currentVersion          NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                activeFlag              VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N/Y)
                                noteText                VARCHAR2(500), -- 备注信息
                                extProperty             CLOB, -- 扩展属性，JSON格式

                                reserved1               VARCHAR2(500), -- 预留字段1
                                reserved2               VARCHAR2(500), -- 预留字段2
                                reserved3               VARCHAR2(500), -- 预留字段3
                                reserved4               VARCHAR2(500), -- 预留字段4
                                reserved5               VARCHAR2(500), -- 预留字段5
                                reserved6               VARCHAR2(500), -- 预留字段6
                                reserved7               VARCHAR2(500), -- 预留字段7
                                reserved8               VARCHAR2(500), -- 预留字段8
                                reserved9               VARCHAR2(500), -- 预留字段9
                                reserved10              VARCHAR2(500), -- 预留字段10

                                CONSTRAINT PK_TIMER_TASK PRIMARY KEY (tenantId, taskId)
);
CREATE INDEX IDX_TIMER_TASK_NAME ON HUB_TIMER_TASK(taskName);
CREATE INDEX IDX_TIMER_TASK_SCHED ON HUB_TIMER_TASK(schedulerId);
CREATE INDEX IDX_TIMER_TASK_TYPE ON HUB_TIMER_TASK(scheduleType);
CREATE INDEX IDX_TIMER_TASK_STATUS ON HUB_TIMER_TASK(taskStatus);
CREATE INDEX IDX_TIMER_TASK_ACTIVE ON HUB_TIMER_TASK(activeFlag);

-- 新增索引
CREATE INDEX IDX_TIMER_TASK_EXEC ON HUB_TIMER_TASK(executorType);
CREATE INDEX IDX_TIMER_TASK_TOOL ON HUB_TIMER_TASK(toolConfigId);
CREATE INDEX IDX_TIMER_TASK_OP ON HUB_TIMER_TASK(operationType);
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

                                         CONSTRAINT PK_TIMER_EXECUTION_LOG PRIMARY KEY (tenantId, executionId)
);
CREATE INDEX IDX_TIMER_LOG_TASK ON HUB_TIMER_EXECUTION_LOG(taskId);
CREATE INDEX IDX_TIMER_LOG_NAME ON HUB_TIMER_EXECUTION_LOG(taskName);
CREATE INDEX IDX_TIMER_LOG_SCHED ON HUB_TIMER_EXECUTION_LOG(schedulerId);
CREATE INDEX IDX_TIMER_LOG_START ON HUB_TIMER_EXECUTION_LOG(executionStartTime);
CREATE INDEX IDX_TIMER_LOG_STATUS ON HUB_TIMER_EXECUTION_LOG(executionStatus);
CREATE INDEX IDX_TIMER_LOG_SUCCESS ON HUB_TIMER_EXECUTION_LOG(resultSuccess);
CREATE INDEX IDX_TIMER_LOG_LEVEL ON HUB_TIMER_EXECUTION_LOG(logLevel);
CREATE INDEX IDX_TIMER_LOG_TIME ON HUB_TIMER_EXECUTION_LOG(logTimestamp);
COMMENT ON TABLE HUB_TIMER_EXECUTION_LOG IS '任务执行日志表 - 合并执行记录和日志信息';


-- 创建表
CREATE TABLE HUB_TOOL_CONFIG (
                                 toolConfigId      VARCHAR2(32)   NOT NULL,
                                 tenantId          VARCHAR2(32)   NOT NULL,

    -- 工具基础信息
                                 toolName          VARCHAR2(100)  NOT NULL,
                                 toolType          VARCHAR2(50)   NOT NULL,
                                 toolVersion       VARCHAR2(20),
                                 configName        VARCHAR2(100)  NOT NULL,
                                 configDescription VARCHAR2(500),

    -- 分组信息
                                 configGroupId     VARCHAR2(32),
                                 configGroupName   VARCHAR2(100),

    -- 连接配置
                                 hostAddress       VARCHAR2(255),
                                 portNumber        NUMBER(10),
                                 protocolType      VARCHAR2(20),

    -- 认证配置
                                 authType          VARCHAR2(50),
                                 userName          VARCHAR2(100),
                                 passwordEncrypted VARCHAR2(500),
                                 keyFilePath       VARCHAR2(500),
                                 keyFileContent    CLOB,

    -- 配置参数
                                 configParameters  CLOB,
                                 environmentVariables CLOB,
                                 customSettings    CLOB,

    -- 状态和控制
                                 configStatus      CHAR(1)        DEFAULT 'Y' NOT NULL,
                                 defaultFlag       CHAR(1)        DEFAULT 'N' NOT NULL,
                                 priorityLevel     NUMBER(10)     DEFAULT 100,

    -- 安全和加密
                                 encryptionType    VARCHAR2(50),
                                 encryptionKey     VARCHAR2(100),

    -- 标准字段
                                 addTime           DATE           DEFAULT SYSDATE NOT NULL,
                                 addWho            VARCHAR2(32)   NOT NULL,
                                 editTime          DATE           DEFAULT SYSDATE NOT NULL,
                                 editWho           VARCHAR2(32)   NOT NULL,
                                 oprSeqFlag        VARCHAR2(32)   NOT NULL,
                                 currentVersion    NUMBER(10)     DEFAULT 1 NOT NULL,
                                 activeFlag        CHAR(1)        DEFAULT 'Y' NOT NULL,
                                 noteText          VARCHAR2(500),
                                 extProperty       CLOB,
                                 reserved1         VARCHAR2(500),
                                 reserved2         VARCHAR2(500),
                                 reserved3         VARCHAR2(500),
                                 reserved4         VARCHAR2(500),
                                 reserved5         VARCHAR2(500),
                                 reserved6         VARCHAR2(500),
                                 reserved7         VARCHAR2(500),
                                 reserved8         VARCHAR2(500),
                                 reserved9         VARCHAR2(500),
                                 reserved10        VARCHAR2(500),

    -- 主键定义
                                 CONSTRAINT PK_TOOL_CONFIG PRIMARY KEY (tenantId, toolConfigId)
);

-- 创建索引
CREATE INDEX IDX_TOOL_CONFIG_NAME      ON HUB_TOOL_CONFIG(toolName);
CREATE INDEX IDX_TOOL_CONFIG_TYPE      ON HUB_TOOL_CONFIG(toolType);
CREATE INDEX IDX_TOOL_CONFIG_CFGNAME   ON HUB_TOOL_CONFIG(configName);
CREATE INDEX IDX_TOOL_CONFIG_GROUP     ON HUB_TOOL_CONFIG(configGroupId);
CREATE INDEX IDX_TOOL_CONFIG_STATUS    ON HUB_TOOL_CONFIG(configStatus);
CREATE INDEX IDX_TOOL_CONFIG_DEFAULT   ON HUB_TOOL_CONFIG(defaultFlag);
CREATE INDEX IDX_TOOL_CONFIG_ACTIVE    ON HUB_TOOL_CONFIG(activeFlag);
-- 添加表注释
COMMENT ON TABLE HUB_TOOL_CONFIG IS '工具配置主表 - 存储各种工具的基础配置信息';

-- 创建表
CREATE TABLE HUB_TOOL_CONFIG_GROUP (
                                       configGroupId     VARCHAR2(32)   NOT NULL,
                                       tenantId          VARCHAR2(32)   NOT NULL,

    -- 分组信息
                                       groupName         VARCHAR2(100)  NOT NULL,
                                       groupDescription  VARCHAR2(500),
                                       parentGroupId     VARCHAR2(32),
                                       groupLevel        NUMBER(10)     DEFAULT 1,
                                       groupPath         VARCHAR2(500),

    -- 分组属性
                                       groupType         VARCHAR2(50),
                                       sortOrder         NUMBER(10)     DEFAULT 100,
                                       groupIcon         VARCHAR2(100),
                                       groupColor        VARCHAR2(20),

    -- 权限控制
                                       accessLevel       VARCHAR2(20)   DEFAULT 'private',
                                       allowedUsers      CLOB,
                                       allowedRoles      CLOB,

    -- 标准字段
                                       addTime           DATE           DEFAULT SYSDATE NOT NULL,
                                       addWho            VARCHAR2(32)   NOT NULL,
                                       editTime          DATE           DEFAULT SYSDATE NOT NULL,
                                       editWho           VARCHAR2(32)   NOT NULL,
                                       oprSeqFlag        VARCHAR2(32)   NOT NULL,
                                       currentVersion    NUMBER(10)     DEFAULT 1 NOT NULL,
                                       activeFlag        CHAR(1)        DEFAULT 'Y' NOT NULL,
                                       noteText          VARCHAR2(500),
                                       extProperty       CLOB,
                                       reserved1         VARCHAR2(500),
                                       reserved2         VARCHAR2(500),
                                       reserved3         VARCHAR2(500),
                                       reserved4         VARCHAR2(500),
                                       reserved5         VARCHAR2(500),
                                       reserved6         VARCHAR2(500),
                                       reserved7         VARCHAR2(500),
                                       reserved8         VARCHAR2(500),
                                       reserved9         VARCHAR2(500),
                                       reserved10        VARCHAR2(500),

    -- 主键定义
                                       CONSTRAINT PK_TOOL_CONFIG_GROUP PRIMARY KEY (tenantId, configGroupId)
);

-- 创建索引
CREATE INDEX IDX_TOOL_GROUP_NAME       ON HUB_TOOL_CONFIG_GROUP(groupName);
CREATE INDEX IDX_TOOL_GROUP_PARENT     ON HUB_TOOL_CONFIG_GROUP(parentGroupId);
CREATE INDEX IDX_TOOL_GROUP_TYPE       ON HUB_TOOL_CONFIG_GROUP(groupType);
CREATE INDEX IDX_TOOL_GROUP_SORT       ON HUB_TOOL_CONFIG_GROUP(sortOrder);
CREATE INDEX IDX_TOOL_GROUP_ACTIVE     ON HUB_TOOL_CONFIG_GROUP(activeFlag);
-- 添加表注释
COMMENT ON TABLE HUB_TOOL_CONFIG_GROUP IS '工具配置分组表 - 用于对工具配置进行分组管理';

CREATE TABLE HUB_GW_LOG_CONFIG (
                                        tenantId VARCHAR2(32) NOT NULL,
                                        logConfigId VARCHAR2(32) NOT NULL,
                                        configName VARCHAR2(100) NOT NULL,
                                        logLevel VARCHAR2(20) DEFAULT 'INFO' NOT NULL,
                                        logFormat VARCHAR2(50) DEFAULT 'JSON' NOT NULL,
                                        outputTargets VARCHAR2(200) DEFAULT 'CONSOLE' NOT NULL,
                                        fileConfig CLOB DEFAULT NULL,
                                        databaseConfig CLOB DEFAULT NULL,
                                        enableAccessLog VARCHAR2(1) DEFAULT 'Y' NOT NULL,
                                        enableErrorLog VARCHAR2(1) DEFAULT 'Y' NOT NULL,
                                        enableAuditLog VARCHAR2(1) DEFAULT 'N' NOT NULL,
                                        logRetentionDays NUMBER(10) DEFAULT 30 NOT NULL,
                                        sensitiveFields CLOB DEFAULT NULL,
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
                                        CONSTRAINT PK_GW_LOG_CONFIG PRIMARY KEY (tenantId, logConfigId)
);

-- 添加表注释
COMMENT ON TABLE HUB_GW_LOG_CONFIG IS '日志配置表 - 存储网关日志相关配置';
-- 创建索引（注意Oracle索引名最长30个字符）
CREATE INDEX IDX_GW_LOG_NAME ON HUB_GW_LOG_CONFIG (configName);
CREATE INDEX IDX_GW_LOG_PRIORITY ON HUB_GW_LOG_CONFIG (configPriority);

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

-- 添加表注释
COMMENT ON TABLE HUB_GW_CORS_CONFIG IS '跨域配置表 - 存储CORS相关配置';

-- 创建索引（注意Oracle索引名最长30个字符）
CREATE INDEX IDX_GW_CORS_INST ON HUB_GW_CORS_CONFIG (gatewayInstanceId);
CREATE INDEX IDX_GW_CORS_ROUTE ON HUB_GW_CORS_CONFIG (routeConfigId);
CREATE INDEX IDX_GW_CORS_PRIORITY ON HUB_GW_CORS_CONFIG (configPriority);


-- 安全配置表 - 存储网关安全策略配置
CREATE TABLE HUB_GW_SECURITY_CONFIG (
                                        tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                        securityConfigId VARCHAR2(32) NOT NULL, -- 安全配置ID
                                        gatewayInstanceId VARCHAR2(32), -- 网关实例ID(实例级安全配置)
                                        routeConfigId VARCHAR2(32), -- 路由配置ID(路由级安全配置)
                                        configName VARCHAR2(100) NOT NULL, -- 安全配置名称
                                        configDesc VARCHAR2(200), -- 安全配置描述
                                        configPriority NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高
                                        customConfigJson CLOB, -- 自定义配置参数,JSON格式
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
                                        CONSTRAINT PK_GW_SECURITY_CONFIG PRIMARY KEY (tenantId, securityConfigId)
);
CREATE INDEX IDX_GW_SEC_INST ON HUB_GW_SECURITY_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_SEC_ROUTE ON HUB_GW_SECURITY_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_SEC_PRIORITY ON HUB_GW_SECURITY_CONFIG(configPriority);
COMMENT ON TABLE HUB_GW_SECURITY_CONFIG IS '安全配置表 - 存储网关安全策略配置';

-- IP访问控制配置表 - 存储IP白名单黑名单规则
CREATE TABLE HUB_GW_IP_ACCESS_CONFIG (
                                         tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                         ipAccessConfigId VARCHAR2(32) NOT NULL, -- IP访问配置ID
                                         securityConfigId VARCHAR2(32) NOT NULL, -- 关联的安全配置ID
                                         configName VARCHAR2(100) NOT NULL, -- IP访问配置名称
                                         defaultPolicy VARCHAR2(10) DEFAULT 'allow' NOT NULL, -- 默认策略(allow允许,deny拒绝)
                                         whitelistIps CLOB, -- IP白名单,JSON数组格式
                                         blacklistIps CLOB, -- IP黑名单,JSON数组格式
                                         whitelistCidrs CLOB, -- CIDR白名单,JSON数组格式
                                         blacklistCidrs CLOB, -- CIDR黑名单,JSON数组格式
                                         trustXForwardedFor VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否信任X-Forwarded-For头(N否,Y是)
                                         trustXRealIp VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否信任X-Real-IP头(N否,Y是)
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
                                         CONSTRAINT PK_GW_IP_ACCESS_CONFIG PRIMARY KEY (tenantId, ipAccessConfigId)
);
CREATE INDEX IDX_GW_IP_SECURITY ON HUB_GW_IP_ACCESS_CONFIG(securityConfigId);
COMMENT ON TABLE HUB_GW_IP_ACCESS_CONFIG IS 'IP访问控制配置表 - 存储IP白名单黑名单规则';

-- User-Agent访问控制配置表 - 存储User-Agent过滤规则
CREATE TABLE HUB_GW_UA_ACCESS_CONFIG (
                                               tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                               useragentAccessConfigId VARCHAR2(32) NOT NULL, -- User-Agent访问配置ID
                                               securityConfigId VARCHAR2(32) NOT NULL, -- 关联的安全配置ID
                                               configName VARCHAR2(100) NOT NULL, -- User-Agent访问配置名称
                                               defaultPolicy VARCHAR2(10) DEFAULT 'allow' NOT NULL, -- 默认策略(allow允许,deny拒绝)
                                               whitelistPatterns CLOB, -- User-Agent白名单模式,JSON数组格式,支持正则表达式
                                               blacklistPatterns CLOB, -- User-Agent黑名单模式,JSON数组格式,支持正则表达式
                                               blockEmptyUserAgent VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否阻止空User-Agent(N否,Y是)
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
                                               CONSTRAINT PK_GW_UA_ACCESS_CONFIG PRIMARY KEY (tenantId, useragentAccessConfigId)
);
CREATE INDEX IDX_GW_UA_SECURITY ON HUB_GW_UA_ACCESS_CONFIG(securityConfigId);
COMMENT ON TABLE HUB_GW_UA_ACCESS_CONFIG IS 'User-Agent访问控制配置表 - 存储User-Agent过滤规则';

-- API访问控制配置表 - 存储API路径和方法过滤规则
CREATE TABLE HUB_GW_API_ACCESS_CONFIG (
                                          tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                          apiAccessConfigId VARCHAR2(32) NOT NULL, -- API访问配置ID
                                          securityConfigId VARCHAR2(32) NOT NULL, -- 关联的安全配置ID
                                          configName VARCHAR2(100) NOT NULL, -- API访问配置名称
                                          defaultPolicy VARCHAR2(10) DEFAULT 'allow' NOT NULL, -- 默认策略(allow允许,deny拒绝)
                                          whitelistPaths CLOB, -- API路径白名单,JSON数组格式,支持通配符
                                          blacklistPaths CLOB, -- API路径黑名单,JSON数组格式,支持通配符
                                          allowedMethods VARCHAR2(200) DEFAULT 'GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS', -- 允许的HTTP方法,逗号分隔
                                          blockedMethods VARCHAR2(200), -- 禁止的HTTP方法,逗号分隔
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
                                          CONSTRAINT PK_GW_API_ACCESS_CONFIG PRIMARY KEY (tenantId, apiAccessConfigId)
);
CREATE INDEX IDX_GW_API_SECURITY ON HUB_GW_API_ACCESS_CONFIG(securityConfigId);
COMMENT ON TABLE HUB_GW_API_ACCESS_CONFIG IS 'API访问控制配置表 - 存储API路径和方法过滤规则';

-- 域名访问控制配置表 - 存储域名白名单黑名单规则
CREATE TABLE HUB_GW_DOMAIN_ACCESS_CONFIG (
                                             tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                             domainAccessConfigId VARCHAR2(32) NOT NULL, -- 域名访问配置ID
                                             securityConfigId VARCHAR2(32) NOT NULL, -- 关联的安全配置ID
                                             configName VARCHAR2(100) NOT NULL, -- 域名访问配置名称
                                             defaultPolicy VARCHAR2(10) DEFAULT 'allow' NOT NULL, -- 默认策略(allow允许,deny拒绝)
                                             whitelistDomains CLOB, -- 域名白名单,JSON数组格式
                                             blacklistDomains CLOB, -- 域名黑名单,JSON数组格式
                                             allowSubdomains VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否允许子域名(N否,Y是)
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
                                             CONSTRAINT PK_GW_DOMAIN_ACCESS_CONFIG PRIMARY KEY (tenantId, domainAccessConfigId)
);
CREATE INDEX IDX_GW_DOMAIN_SECURITY ON HUB_GW_DOMAIN_ACCESS_CONFIG(securityConfigId);
COMMENT ON TABLE HUB_GW_DOMAIN_ACCESS_CONFIG IS '域名访问控制配置表 - 存储域名白名单黑名单规则';
