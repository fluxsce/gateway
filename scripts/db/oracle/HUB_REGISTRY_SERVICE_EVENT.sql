CREATE TABLE HUB_REGISTRY_SERVICE_EVENT (
                                            serviceEventId VARCHAR2(32) NOT NULL, -- 服务事件ID，主键
                                            tenantId VARCHAR2(32) NOT NULL, -- 租户ID，用于多租户数据隔离

    -- 关联主键字段（用于精确关联到对应表记录）
                                            serviceGroupId VARCHAR2(32), -- 服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键
                                            serviceInstanceId VARCHAR2(100), -- 服务实例ID，关联HUB_REGISTRY_SERVICE_INSTANCE表主键

    -- 事件基本信息（冗余字段，便于查询和展示）
                                            groupName VARCHAR2(100), -- 分组名称，冗余字段便于查询
                                            serviceName VARCHAR2(100), -- 服务名称，冗余字段便于查询
                                            hostAddress VARCHAR2(100), -- 主机地址，冗余字段便于查询
                                            portNumber NUMBER(10), -- 端口号，冗余字段便于查询
                                            nodeIpAddress VARCHAR2(100), -- 节点IP地址，记录程序运行的IP
                                            eventType VARCHAR2(50) NOT NULL, -- 事件类型(GROUP_CREATE,GROUP_UPDATE,GROUP_DELETE,SERVICE_CREATE,SERVICE_UPDATE,SERVICE_DELETE,INSTANCE_REGISTER,INSTANCE_DEREGISTER,INSTANCE_HEARTBEAT,INSTANCE_HEALTH_CHANGE,INSTANCE_STATUS_CHANGE)
                                            eventSource VARCHAR2(100), -- 事件来源

    -- 事件数据
                                            eventDataJson CLOB, -- 事件数据，JSON格式
                                            eventMessage VARCHAR2(1000), -- 事件消息描述

    -- 时间信息
                                            eventTime DATE DEFAULT SYSDATE NOT NULL, -- 事件发生时间

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

                                            CONSTRAINT PK_REGISTRY_SVC_EVENT PRIMARY KEY (tenantId, serviceEventId)
);
CREATE INDEX IDX_REG_EVENT_GROUP_ID ON HUB_REGISTRY_SERVICE_EVENT(tenantId, serviceGroupId, eventTime);
CREATE INDEX IDX_REG_EVENT_INST_ID ON HUB_REGISTRY_SERVICE_EVENT(tenantId, serviceInstanceId, eventTime);
CREATE INDEX IDX_REG_EVENT_GROUP_NAME ON HUB_REGISTRY_SERVICE_EVENT(tenantId, groupName, eventTime);
CREATE INDEX IDX_REG_EVENT_SVC_NAME ON HUB_REGISTRY_SERVICE_EVENT(tenantId, serviceName, eventTime);
CREATE INDEX IDX_REG_EVENT_HOST ON HUB_REGISTRY_SERVICE_EVENT(tenantId, hostAddress, portNumber, eventTime);
CREATE INDEX IDX_REG_EVENT_NODE_IP ON HUB_REGISTRY_SERVICE_EVENT(tenantId, nodeIpAddress, eventTime);
CREATE INDEX IDX_REG_EVENT_TYPE ON HUB_REGISTRY_SERVICE_EVENT(eventType, eventTime);
CREATE INDEX IDX_REG_EVENT_TIME ON HUB_REGISTRY_SERVICE_EVENT(eventTime);
COMMENT ON TABLE HUB_REGISTRY_SERVICE_EVENT IS '服务事件日志表 - 记录服务注册发现相关的所有事件';

