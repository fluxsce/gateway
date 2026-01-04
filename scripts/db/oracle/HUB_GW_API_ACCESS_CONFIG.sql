CREATE TABLE HUB_GW_API_ACCESS_CONFIG (
                                          tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                          apiAccessConfigId VARCHAR2(32) NOT NULL, -- API访问配置ID
                                          securityConfigId VARCHAR2(32) NOT NULL, -- 关联的安全配置ID
                                          configName VARCHAR2(100) NOT NULL, -- API访问配置名称
                                          defaultPolicy VARCHAR2(10) DEFAULT 'allow' NOT NULL, -- 默认策略(allow允许,deny拒绝)
                                          whitelistPaths CLOB, -- API路径白名单,JSON数组格式,支持通配符
                                          blacklistPaths CLOB, -- API路径黑名单,JSON数组格式,支持通配符
                                          allowedMethods VARCHAR2(200) DEFAULT 'GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS', -- 允许的HTTP方法,逗号分隔
                                          blockedMethods VARCHAR2(200), -- 禁止的HTTP方法,逗号分隔
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
                                          CONSTRAINT PK_GW_API_ACCESS_CONFIG PRIMARY KEY (tenantId, apiAccessConfigId)
);
CREATE INDEX IDX_GW_API_SECURITY ON HUB_GW_API_ACCESS_CONFIG(securityConfigId);
COMMENT ON TABLE HUB_GW_API_ACCESS_CONFIG IS 'API访问控制配置表 - 存储API路径和方法过滤规则';
