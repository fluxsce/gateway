CREATE TABLE HUB_CLUSTER_EVENT (
                                            eventId VARCHAR2(64) NOT NULL, -- 事件ID
                                            tenantId VARCHAR2(32) NOT NULL, -- 租户ID

    -- 事件来源(发布者)
                                            sourceNodeId VARCHAR2(100) NOT NULL, -- 发布节点ID(hostname:port)
                                            sourceNodeIp VARCHAR2(100), -- 发布节点IP

    -- 事件信息
                                            eventType VARCHAR2(50) NOT NULL, -- 事件类型(ROUTE_CONFIG/SERVICE_CONFIG/FILTER_CONFIG/CACHE_REFRESH等)
                                            eventAction VARCHAR2(50) NOT NULL, -- 事件动作(CREATE/UPDATE/DELETE/REFRESH/INVALIDATE)
                                            eventPayload CLOB, -- 事件数据(JSON格式，包含所有业务信息)

    -- 事件时间
                                            eventTime DATE DEFAULT SYSDATE NOT NULL, -- 事件发生时间
                                            expireTime DATE, -- 事件过期时间

    -- 通用字段
                                            addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                            addWho VARCHAR2(64) NOT NULL, -- 创建人ID
                                            editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                            editWho VARCHAR2(64) NOT NULL, -- 最后修改人ID
                                            oprSeqFlag VARCHAR2(64) NOT NULL, -- 操作序列标识
                                            currentVersion NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                            activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                            noteText CLOB, -- 备注信息
                                            extProperty CLOB, -- 扩展属性(JSON格式)
                                            reserved1 VARCHAR2(500), -- 预留字段1
                                            reserved2 VARCHAR2(500), -- 预留字段2
                                            reserved3 VARCHAR2(500), -- 预留字段3
                                            reserved4 VARCHAR2(500), -- 预留字段4
                                            reserved5 VARCHAR2(500), -- 预留字段5

                                            CONSTRAINT PK_CLUSTER_EVENT PRIMARY KEY (tenantId, eventId)
);
-- 优化后的复合索引：支持高效的待处理事件查询
-- 查询模式: WHERE tenantId=? AND activeFlag='Y' AND eventTime>? AND sourceNodeId!=?
CREATE INDEX IDX_CLS_EVT_QUERY ON HUB_CLUSTER_EVENT(tenantId, activeFlag, eventTime, eventId);
-- 辅助索引：用于特定场景查询
CREATE INDEX IDX_CLS_EVT_SOURCE ON HUB_CLUSTER_EVENT(sourceNodeId);
CREATE INDEX IDX_CLS_EVT_TYPE ON HUB_CLUSTER_EVENT(eventType, eventTime);
CREATE INDEX IDX_CLS_EVT_EXPIRE ON HUB_CLUSTER_EVENT(expireTime);
COMMENT ON TABLE HUB_CLUSTER_EVENT IS '集群事件表 - 存储集群中各节点发布的事件';

