CREATE TABLE HUB_CLUSTER_EVENT_ACK (
                                            ackId VARCHAR2(64) NOT NULL, -- 确认ID
                                            tenantId VARCHAR2(32) NOT NULL, -- 租户ID
                                            eventId VARCHAR2(64) NOT NULL, -- 事件ID

    -- 处理节点
                                            nodeId VARCHAR2(100) NOT NULL, -- 处理节点ID(hostname:port)
                                            nodeIp VARCHAR2(100), -- 处理节点IP

    -- 处理状态
                                            ackStatus VARCHAR2(20) DEFAULT 'PENDING' NOT NULL, -- 确认状态(PENDING/SUCCESS/FAILED/SKIPPED)
                                            processTime DATE, -- 处理时间
                                            resultMessage VARCHAR2(2000), -- 结果信息或错误信息
                                            retryCount NUMBER(10) DEFAULT 0 NOT NULL, -- 重试次数

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

                                            CONSTRAINT PK_CLUSTER_EVENT_ACK PRIMARY KEY (tenantId, ackId)
);
-- 优化后的复合索引：支持 NOT EXISTS 子查询高效执行
-- 查询模式: WHERE eventId=? AND nodeId=? AND ackStatus='SUCCESS'
CREATE INDEX IDX_CLS_ACK_EVT_NODE ON HUB_CLUSTER_EVENT_ACK(eventId, nodeId, ackStatus);
-- 辅助索引：用于按节点查询处理状态
CREATE INDEX IDX_CLS_ACK_NODE ON HUB_CLUSTER_EVENT_ACK(nodeId, ackStatus, processTime);
-- 辅助索引：用于清理任务
CREATE INDEX IDX_CLS_ACK_CLEANUP ON HUB_CLUSTER_EVENT_ACK(tenantId, addTime);
COMMENT ON TABLE HUB_CLUSTER_EVENT_ACK IS '集群事件确认表 - 跟踪各节点对事件的处理状态';

