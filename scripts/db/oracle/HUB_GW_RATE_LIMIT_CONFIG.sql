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
