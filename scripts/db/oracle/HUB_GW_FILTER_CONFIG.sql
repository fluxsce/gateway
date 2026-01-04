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