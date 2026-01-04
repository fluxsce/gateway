CREATE TABLE HUB_REGISTRY_SERVICE_GROUP (
                                            serviceGroupId VARCHAR2(32) NOT NULL, -- 服务分组ID，主键
                                            tenantId VARCHAR2(32) NOT NULL, -- 租户ID，用于多租户数据隔离

    -- 分组基本信息
                                            groupName VARCHAR2(100) NOT NULL, -- 分组名称
                                            groupDescription VARCHAR2(500), -- 分组描述
                                            groupType VARCHAR2(50) DEFAULT 'BUSINESS' NOT NULL, -- 分组类型(BUSINESS,SYSTEM,TEST)

    -- 授权信息
                                            ownerUserId VARCHAR2(32) NOT NULL, -- 分组所有者用户ID
                                            adminUserIds CLOB, -- 管理员用户ID列表，JSON格式
                                            readUserIds CLOB, -- 只读用户ID列表，JSON格式
                                            accessControlEnabled VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用访问控制(N否,Y是)

    -- 配置信息
                                            defaultProtocolType VARCHAR2(20) DEFAULT 'HTTP' NOT NULL, -- 默认协议类型
                                            defaultLoadBalanceStrategy VARCHAR2(50) DEFAULT 'ROUND_ROBIN' NOT NULL, -- 默认负载均衡策略
                                            defaultHealthCheckUrl VARCHAR2(500) DEFAULT '/health' NOT NULL, -- 默认健康检查URL
                                            defaultHealthCheckIntervalSeconds NUMBER(10) DEFAULT 30 NOT NULL, -- 默认健康检查间隔(秒)

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

                                            CONSTRAINT PK_REGISTRY_SERVICE_GROUP PRIMARY KEY (tenantId, serviceGroupId)
);
CREATE INDEX IDX_REG_GROUP_NAME ON HUB_REGISTRY_SERVICE_GROUP(tenantId, groupName);
CREATE INDEX IDX_REG_GROUP_TYPE ON HUB_REGISTRY_SERVICE_GROUP(groupType);
CREATE INDEX IDX_REG_GROUP_OWNER ON HUB_REGISTRY_SERVICE_GROUP(ownerUserId);
CREATE INDEX IDX_REG_GROUP_ACTIVE ON HUB_REGISTRY_SERVICE_GROUP(activeFlag);
COMMENT ON TABLE HUB_REGISTRY_SERVICE_GROUP IS '服务分组表 - 存储服务分组和授权信息';
