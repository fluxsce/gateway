CREATE TABLE HUB_GW_IP_ACCESS_CONFIG (
                                         tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                         ipAccessConfigId VARCHAR2(32) NOT NULL, -- IP访问配置ID
                                         securityConfigId VARCHAR2(32) NOT NULL, -- 关联的安全配置ID
                                         configName VARCHAR2(100) NOT NULL, -- IP访问配置名称
                                         defaultPolicy VARCHAR2(10) DEFAULT 'allow' NOT NULL, -- 默认策略(allow允许,deny拒绝)
                                         whitelistIps CLOB, -- IP白名单,JSON数组格式
                                         blacklistIps CLOB, -- IP黑名单,JSON数组格式
                                         whitelistCidrs CLOB, -- CIDR白名单,JSON数组格式
                                         blacklistCidrs CLOB, -- CIDR黑名单,JSON数组格式
                                         trustXForwardedFor VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否信任X-Forwarded-For头(N否,Y是)
                                         trustXRealIp VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否信任X-Real-IP头(N否,Y是)
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
                                         CONSTRAINT PK_GW_IP_ACCESS_CONFIG PRIMARY KEY (tenantId, ipAccessConfigId)
);
CREATE INDEX IDX_GW_IP_SECURITY ON HUB_GW_IP_ACCESS_CONFIG(securityConfigId);
COMMENT ON TABLE HUB_GW_IP_ACCESS_CONFIG IS 'IP访问控制配置表 - 存储IP白名单黑名单规则';
