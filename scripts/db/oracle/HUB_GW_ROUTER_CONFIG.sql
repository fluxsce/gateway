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
