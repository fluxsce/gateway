CREATE TABLE HUB_GW_UA_ACCESS_CONFIG (
                                               tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                               useragentAccessConfigId VARCHAR2(32) NOT NULL, -- User-Agent访问配置ID
                                               securityConfigId VARCHAR2(32) NOT NULL, -- 关联的安全配置ID
                                               configName VARCHAR2(100) NOT NULL, -- User-Agent访问配置名称
                                               defaultPolicy VARCHAR2(10) DEFAULT 'allow' NOT NULL, -- 默认策略(allow允许,deny拒绝)
                                               whitelistPatterns CLOB, -- User-Agent白名单模式,JSON数组格式,支持正则表达式
                                               blacklistPatterns CLOB, -- User-Agent黑名单模式,JSON数组格式,支持正则表达式
                                               blockEmptyUserAgent VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否阻止空User-Agent(N否,Y是)
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
                                               CONSTRAINT PK_GW_UA_ACCESS_CONFIG PRIMARY KEY (tenantId, useragentAccessConfigId)
);
CREATE INDEX IDX_GW_UA_SECURITY ON HUB_GW_UA_ACCESS_CONFIG(securityConfigId);
COMMENT ON TABLE HUB_GW_UA_ACCESS_CONFIG IS 'User-Agent访问控制配置表 - 存储User-Agent过滤规则';
