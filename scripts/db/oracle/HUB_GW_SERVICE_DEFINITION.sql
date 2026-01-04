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
