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