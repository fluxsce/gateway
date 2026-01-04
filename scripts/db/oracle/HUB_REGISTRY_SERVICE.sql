CREATE TABLE HUB_REGISTRY_SERVICE (
                                      tenantId VARCHAR2(32) NOT NULL, -- 租户ID，用于多租户数据隔离
                                      serviceName VARCHAR2(100) NOT NULL, -- 服务名称，主键

    -- 关联分组（主键关联）
                                      serviceGroupId VARCHAR2(32) NOT NULL, -- 服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键
    -- 冗余字段（便于查询和展示）
                                      groupName VARCHAR2(100) NOT NULL, -- 分组名称，冗余字段便于查询

    -- 服务基本信息
                                      serviceDescription VARCHAR2(500), -- 服务描述

    -- 注册管理配置
                                      registryType VARCHAR2(20) DEFAULT 'INTERNAL' NOT NULL, -- 注册类型(INTERNAL:内部管理,NACOS:Nacos注册中心,CONSUL:Consul,EUREKA:Eureka,ETCD:ETCD,ZOOKEEPER:ZooKeeper)
                                      externalRegistryConfig CLOB, -- 外部注册中心配置，JSON格式，仅当registryType非INTERNAL时使用

    -- 服务配置
                                      protocolType VARCHAR2(20) DEFAULT 'HTTP' NOT NULL, -- 协议类型(HTTP,HTTPS,TCP,UDP,GRPC)
                                      contextPath VARCHAR2(200) DEFAULT '' NOT NULL, -- 上下文路径
                                      loadBalanceStrategy VARCHAR2(50) DEFAULT 'ROUND_ROBIN' NOT NULL, -- 负载均衡策略

    -- 健康检查配置
                                      healthCheckUrl VARCHAR2(500) DEFAULT '/health' NOT NULL, -- 健康检查URL
                                      healthCheckIntervalSeconds NUMBER(10) DEFAULT 30 NOT NULL, -- 健康检查间隔(秒)
                                      healthCheckTimeoutSeconds NUMBER(10) DEFAULT 5 NOT NULL, -- 健康检查超时(秒)
                                      healthCheckType VARCHAR2(20) DEFAULT 'HTTP' NOT NULL, -- 健康检查类型(HTTP,TCP)
                                      healthCheckMode VARCHAR2(20) DEFAULT 'ACTIVE' NOT NULL, -- 健康检查模式(ACTIVE:主动探测,PASSIVE:客户端上报)

    -- 元数据和标签
                                      metadataJson CLOB, -- 服务元数据，JSON格式
                                      tagsJson CLOB, -- 服务标签，JSON格式

    -- 通用字段
                                      addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                      addWho VARCHAR2(32) NOT NULL, -- 创建人ID
                                      editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                      editWho VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                      oprSeqFlag VARCHAR2(32) NOT NULL, -- 操作序列标识
                                      currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                      activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                      noteText VARCHAR2(500), -- 备注信息
                                      extProperty CLOB, -- 扩展属性，JSON格式
                                      reserved1 VARCHAR2(500), -- 预留字段1
                                      reserved2 VARCHAR2(500), -- 预留字段2
                                      reserved3 VARCHAR2(500), -- 预留字段3
                                      reserved4 VARCHAR2(500), -- 预留字段4
                                      reserved5 VARCHAR2(500), -- 预留字段5
                                      reserved6 VARCHAR2(500), -- 预留字段6
                                      reserved7 VARCHAR2(500), -- 预留字段7
                                      reserved8 VARCHAR2(500), -- 预留字段8
                                      reserved9 VARCHAR2(500), -- 预留字段9
                                      reserved10 VARCHAR2(500), -- 预留字段10

                                      CONSTRAINT PK_REGISTRY_SERVICE PRIMARY KEY (tenantId, serviceName)
);
CREATE INDEX IDX_REG_SVC_GROUP_ID ON HUB_REGISTRY_SERVICE(tenantId, serviceGroupId);
CREATE INDEX IDX_REG_SVC_GROUP_NAME ON HUB_REGISTRY_SERVICE(groupName);
CREATE INDEX IDX_REG_SVC_REGISTRY_TYPE ON HUB_REGISTRY_SERVICE(registryType);
CREATE INDEX IDX_REG_SVC_ACTIVE ON HUB_REGISTRY_SERVICE(activeFlag);
COMMENT ON TABLE HUB_REGISTRY_SERVICE IS '服务表 - 存储服务的基本信息和配置，支持内部管理和外部注册中心代理模式';
