CREATE TABLE HUB_GW_AUTH_CONFIG (
                                         tenantId         VARCHAR2(32) NOT NULL, -- 租户ID
                                         authConfigId     VARCHAR2(32) NOT NULL, -- 认证配置ID
                                         gatewayInstanceId VARCHAR2(32), -- 网关实例ID(实例级认证)
                                         routeConfigId    VARCHAR2(32), -- 路由配置ID(路由级认证)
                                         authName         VARCHAR2(100) NOT NULL, -- 认证配置名称
                                         authType         VARCHAR2(50) NOT NULL, -- 认证类型(JWT,API_KEY,OAUTH2,BASIC)
                                         authStrategy     VARCHAR2(50) DEFAULT 'REQUIRED', -- 认证策略(REQUIRED,OPTIONAL,DISABLED)
                                         authConfig       CLOB NOT NULL, -- 认证参数配置,JSON格式
                                         exemptPaths      CLOB, -- 豁免路径列表,JSON数组格式
                                         exemptHeaders    CLOB, -- 豁免请求头列表,JSON数组格式
                                         failureStatusCode NUMBER(10) DEFAULT 401 NOT NULL, -- 认证失败状态码
                                         failureMessage   VARCHAR2(200) DEFAULT '认证失败', -- 认证失败提示消息
                                         configPriority   NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高

                                         reserved1        VARCHAR2(100), -- 预留字段1
                                         reserved2        VARCHAR2(100), -- 预留字段2
                                         reserved3        NUMBER(10), -- 预留字段3
                                         reserved4        NUMBER(10), -- 预留字段4
                                         reserved5        DATE, -- 预留字段5
                                         extProperty      CLOB, -- 扩展属性,JSON格式

                                         addTime          DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                         addWho           VARCHAR2(32) NOT NULL, -- 创建人ID
                                         editTime         DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                         editWho          VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                         oprSeqFlag       VARCHAR2(32) NOT NULL, -- 操作序列标识
                                         currentVersion   NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                         activeFlag       VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                         noteText         VARCHAR2(500), -- 备注信息

                                         CONSTRAINT PK_GW_AUTH_CONFIG PRIMARY KEY (tenantId, authConfigId)
);
CREATE INDEX IDX_GW_AUTH_INST ON HUB_GW_AUTH_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_AUTH_ROUTE ON HUB_GW_AUTH_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_AUTH_TYPE ON HUB_GW_AUTH_CONFIG(authType);
CREATE INDEX IDX_GW_AUTH_PRIORITY ON HUB_GW_AUTH_CONFIG(configPriority);
COMMENT ON TABLE HUB_GW_AUTH_CONFIG IS '认证配置表 - 存储认证相关规则';