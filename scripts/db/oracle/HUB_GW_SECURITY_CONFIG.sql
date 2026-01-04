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
