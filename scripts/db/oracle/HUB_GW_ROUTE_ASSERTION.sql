CREATE TABLE HUB_GW_ROUTE_ASSERTION (
                                             tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                             routeAssertionId VARCHAR2(32) NOT NULL, -- 路由断言ID
                                             routeConfigId VARCHAR2(32) NOT NULL, -- 关联的路由配置ID
                                             assertionName VARCHAR2(100) NOT NULL, -- 断言名称
                                             assertionType VARCHAR2(50) NOT NULL, -- 断言类型(PATH,HEADER,QUERY,COOKIE,IP)
                                             assertionOperator VARCHAR2(20) DEFAULT 'EQUAL' NOT NULL, -- 断言操作符(EQUAL,NOT_EQUAL,CONTAINS,MATCHES等)
                                             fieldName VARCHAR2(100), -- 字段名称(header/query名称)
                                             expectedValue VARCHAR2(500), -- 期望值
                                             patternValue VARCHAR2(500), -- 匹配模式(正则表达式等)
                                             caseSensitive VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否区分大小写(N否,Y是)
                                             assertionOrder NUMBER(10) DEFAULT 0 NOT NULL, -- 断言执行顺序
                                             isRequired VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否必须匹配(N否,Y是)
                                             assertionDesc VARCHAR2(200), -- 断言描述

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

                                             CONSTRAINT PK_GW_ROUTE_ASSERTION PRIMARY KEY (tenantId, routeAssertionId)
);
CREATE INDEX IDX_GW_ASSERT_ROUTE ON HUB_GW_ROUTE_ASSERTION(routeConfigId);
CREATE INDEX IDX_GW_ASSERT_TYPE ON HUB_GW_ROUTE_ASSERTION(assertionType);
CREATE INDEX IDX_GW_ASSERT_ORDER ON HUB_GW_ROUTE_ASSERTION(assertionOrder);
COMMENT ON TABLE HUB_GW_ROUTE_ASSERTION IS '路由断言表 - 存储路由匹配断言规则';