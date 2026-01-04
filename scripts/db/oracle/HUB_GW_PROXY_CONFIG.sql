CREATE TABLE HUB_GW_PROXY_CONFIG (
                                          tenantId          VARCHAR2(32) NOT NULL, -- 租户ID
                                          proxyConfigId     VARCHAR2(32) NOT NULL, -- 代理配置ID
                                          gatewayInstanceId VARCHAR2(32) NOT NULL, -- 网关实例ID(代理配置仅支持实例级)
                                          proxyName         VARCHAR2(100) NOT NULL, -- 代理名称

                                          proxyType         VARCHAR2(50) DEFAULT 'http' NOT NULL, -- 代理类型(http,websocket,tcp,udp)

                                          proxyId           VARCHAR2(100), -- 代理ID(来自ProxyConfig.ID)
                                          configPriority    NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高

                                          proxyConfig       CLOB NOT NULL, -- 代理具体配置,JSON格式,根据proxyType存储对应配置
                                          customConfig      CLOB, -- 自定义配置,JSON格式

                                          reserved1         VARCHAR2(100), -- 预留字段1
                                          reserved2         VARCHAR2(100), -- 预留字段2
                                          reserved3         NUMBER(10), -- 预留字段3
                                          reserved4         NUMBER(10), -- 预留字段4
                                          reserved5         DATE, -- 预留字段5
                                          extProperty       CLOB, -- 扩展属性,JSON格式

                                          addTime           DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                          addWho            VARCHAR2(32) NOT NULL, -- 创建人ID
                                          editTime          DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                          editWho           VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                          oprSeqFlag        VARCHAR2(32) NOT NULL, -- 操作序列标识
                                          currentVersion    NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                          activeFlag        VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动/禁用,Y活动/启用)
                                          noteText          VARCHAR2(500), -- 备注信息

                                          CONSTRAINT PK_GW_PROXY_CONFIG PRIMARY KEY (tenantId, proxyConfigId)
);
CREATE INDEX IDX_GW_PROXY_INST ON HUB_GW_PROXY_CONFIG(gatewayInstanceId);
CREATE INDEX IDX_GW_PROXY_TYPE ON HUB_GW_PROXY_CONFIG(proxyType);
CREATE INDEX IDX_GW_PROXY_PRIORITY ON HUB_GW_PROXY_CONFIG(configPriority);
CREATE INDEX IDX_GW_PROXY_ACTIVE ON HUB_GW_PROXY_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GW_PROXY_CONFIG IS '代理配置表 - 根据proxy.go逻辑设计,仅支持实例级代理配置';