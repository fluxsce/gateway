CREATE TABLE HUB_REGISTRY_SERVICE_INSTANCE (
                                               serviceInstanceId VARCHAR2(100) NOT NULL, -- 服务实例ID，主键
                                               tenantId VARCHAR2(32) NOT NULL, -- 租户ID，用于多租户数据隔离

    -- 关联服务和分组（主键关联）
                                               serviceGroupId VARCHAR2(32) NOT NULL, -- 服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键
    -- 冗余字段（便于查询和展示）
                                               serviceName VARCHAR2(100) NOT NULL, -- 服务名称，冗余字段便于查询
                                               groupName VARCHAR2(100) NOT NULL, -- 分组名称，冗余字段便于查询

    -- 网络连接信息
                                               hostAddress VARCHAR2(100) NOT NULL, -- 主机地址
                                               portNumber NUMBER(10) NOT NULL, -- 端口号
                                               contextPath VARCHAR2(200) DEFAULT '' NOT NULL, -- 上下文路径

    -- 实例状态信息
                                               instanceStatus VARCHAR2(20) DEFAULT 'UP' NOT NULL, -- 实例状态(UP,DOWN,STARTING,OUT_OF_SERVICE)
                                               healthStatus VARCHAR2(20) DEFAULT 'UNKNOWN' NOT NULL, -- 健康状态(HEALTHY,UNHEALTHY,UNKNOWN)

    -- 负载均衡配置
                                               weightValue NUMBER(10) DEFAULT 100 NOT NULL, -- 权重值

    -- 客户端信息
                                               clientId VARCHAR2(100), -- 客户端ID
                                               clientVersion VARCHAR2(50), -- 客户端版本
                                               clientType VARCHAR2(50) DEFAULT 'SERVICE' NOT NULL, -- 客户端类型(SERVICE,GATEWAY,ADMIN)
                                               tempInstanceFlag VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 临时实例标记(Y是临时实例,N否)

    -- 健康检查统计
                                               heartbeatFailCount NUMBER(10) DEFAULT 0 NOT NULL, -- 心跳检查失败次数，仅用于计数

    -- 元数据和标签
                                               metadataJson CLOB, -- 实例元数据，JSON格式
                                               tagsJson CLOB, -- 实例标签，JSON格式

    -- 时间戳信息
                                               registerTime DATE DEFAULT SYSDATE NOT NULL, -- 注册时间
                                               lastHeartbeatTime DATE, -- 最后心跳时间
                                               lastHealthCheckTime DATE, -- 最后健康检查时间

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

                                               CONSTRAINT PK_REGISTRY_SVC_INSTANCE PRIMARY KEY (tenantId, serviceInstanceId)
);
CREATE INDEX IDX_REG_INST_COMPOSITE ON HUB_REGISTRY_SERVICE_INSTANCE(tenantId, serviceGroupId, serviceName, hostAddress, portNumber);
CREATE INDEX IDX_REG_INST_GROUP_ID ON HUB_REGISTRY_SERVICE_INSTANCE(tenantId, serviceGroupId);
CREATE INDEX IDX_REG_INST_SVC_NAME ON HUB_REGISTRY_SERVICE_INSTANCE(serviceName);
CREATE INDEX IDX_REG_INST_GROUP_NAME ON HUB_REGISTRY_SERVICE_INSTANCE(groupName);
CREATE INDEX IDX_REG_INST_STATUS ON HUB_REGISTRY_SERVICE_INSTANCE(instanceStatus);
CREATE INDEX IDX_REG_INST_HEALTH ON HUB_REGISTRY_SERVICE_INSTANCE(healthStatus);
CREATE INDEX IDX_REG_INST_HEARTBEAT ON HUB_REGISTRY_SERVICE_INSTANCE(lastHeartbeatTime);
CREATE INDEX IDX_REG_INST_HOST_PORT ON HUB_REGISTRY_SERVICE_INSTANCE(hostAddress, portNumber);
CREATE INDEX IDX_REG_INST_CLIENT ON HUB_REGISTRY_SERVICE_INSTANCE(clientId);
CREATE INDEX IDX_REG_INST_ACTIVE ON HUB_REGISTRY_SERVICE_INSTANCE(activeFlag);
CREATE INDEX IDX_REG_INST_TEMP ON HUB_REGISTRY_SERVICE_INSTANCE(tempInstanceFlag);
COMMENT ON TABLE HUB_REGISTRY_SERVICE_INSTANCE IS '服务实例表 - 存储具体的服务实例信息';
