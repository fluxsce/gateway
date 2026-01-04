CREATE TABLE HUB_GW_SERVICE_NODE (
                                          tenantId              VARCHAR2(32) NOT NULL, -- 租户ID
                                          serviceNodeId         VARCHAR2(32) NOT NULL, -- 服务节点ID
                                          serviceDefinitionId   VARCHAR2(32) NOT NULL, -- 关联的服务定义ID
                                          nodeId                VARCHAR2(100) NOT NULL, -- 节点标识ID

                                          nodeUrl               VARCHAR2(500) NOT NULL, -- 节点完整URL(来自NodeConfig.URL)
                                          nodeHost              VARCHAR2(100) NOT NULL, -- 节点主机地址(从URL解析)
                                          nodePort              NUMBER(10) NOT NULL, -- 节点端口(从URL解析)
                                          nodeProtocol          VARCHAR2(10) DEFAULT 'HTTP' NOT NULL, -- 节点协议(HTTP,HTTPS)

                                          nodeWeight            NUMBER(10) DEFAULT 100 NOT NULL, -- 节点权重(来自NodeConfig.Weight)
                                          healthStatus          VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 健康状态(N不健康,Y健康)

                                          nodeMetadata          CLOB, -- 节点元数据,JSON格式

                                          nodeStatus            NUMBER(10) DEFAULT 1 NOT NULL, -- 节点运行状态(0下线,1在线,2维护)
                                          lastHealthCheckTime   DATE, -- 最后健康检查时间
                                          healthCheckResult     CLOB, -- 健康检查结果详情

                                          reserved1             VARCHAR2(100), -- 预留字段1
                                          reserved2             VARCHAR2(100), -- 预留字段2
                                          reserved3             NUMBER(10), -- 预留字段3
                                          reserved4             NUMBER(10), -- 预留字段4
                                          reserved5             DATE, -- 预留字段5
                                          extProperty           CLOB, -- 扩展属性,JSON格式

                                          addTime               DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                          addWho                VARCHAR2(32) NOT NULL, -- 创建人ID
                                          editTime              DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                          editWho               VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                          oprSeqFlag            VARCHAR2(32) NOT NULL, -- 操作序列标识
                                          currentVersion        NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                          activeFlag            VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                          noteText              VARCHAR2(500), -- 备注信息

                                          CONSTRAINT PK_GW_SERVICE_NODE PRIMARY KEY (tenantId, serviceNodeId)
);
CREATE INDEX IDX_GW_NODE_SERVICE ON HUB_GW_SERVICE_NODE(serviceDefinitionId);
CREATE INDEX IDX_GW_NODE_ENDPOINT ON HUB_GW_SERVICE_NODE(nodeHost, nodePort);
CREATE INDEX IDX_GW_NODE_HEALTH ON HUB_GW_SERVICE_NODE(healthStatus);
CREATE INDEX IDX_GW_NODE_STATUS ON HUB_GW_SERVICE_NODE(nodeStatus);
COMMENT ON TABLE HUB_GW_SERVICE_NODE IS '服务节点表 - 根据NodeConfig结构设计,存储服务节点实例信息';