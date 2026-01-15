
-- 集群事件确认表 - 跟踪各节点对事件的处理状态
CREATE TABLE IF NOT EXISTS HUB_CLUSTER_EVENT_ACK (
  -- 主键和租户信息
  ackId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  eventId TEXT NOT NULL,
  
  -- 处理节点
  nodeId TEXT NOT NULL,
  nodeIp TEXT,
  
  -- 处理状态
  ackStatus TEXT NOT NULL DEFAULT 'PENDING',
  processTime DATETIME,
  resultMessage TEXT,
  retryCount INTEGER NOT NULL DEFAULT 0,
  
  -- 通用字段
  addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  addWho TEXT NOT NULL,
  editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  editWho TEXT NOT NULL,
  oprSeqFlag TEXT NOT NULL,
  currentVersion INTEGER NOT NULL DEFAULT 1,
  activeFlag TEXT NOT NULL DEFAULT 'Y',
  noteText TEXT,
  extProperty TEXT,
  reserved1 TEXT,
  reserved2 TEXT,
  reserved3 TEXT,
  
  PRIMARY KEY (tenantId, ackId)
);

-- =====================================================
-- HUB_CLUSTER_EVENT_ACK 表说明
-- =====================================================
-- 
-- 集群事件确认表用于跟踪每个节点对事件的处理状态
-- 通过此表可以确保每个节点都能处理到所有事件
--
-- 确认状态(ackStatus)说明：
-- 1. PENDING: 待处理
-- 2. SUCCESS: 处理成功
-- 3. FAILED: 处理失败
-- 4. SKIPPED: 跳过(如事件已过期)
--
-- =====================================================

-- =====================================================
-- HUB_CLUSTER_EVENT_ACK 索引（优化后）
-- =====================================================
-- 优化后的复合索引：支持 NOT EXISTS 子查询高效执行
-- 查询模式: WHERE eventId=? AND nodeId=? AND ackStatus='SUCCESS'
CREATE INDEX IF NOT EXISTS IDX_CLS_ACK_EVT_NODE ON HUB_CLUSTER_EVENT_ACK(eventId, nodeId, ackStatus);
-- 辅助索引：用于按节点查询处理状态
CREATE INDEX IF NOT EXISTS IDX_CLS_ACK_NODE ON HUB_CLUSTER_EVENT_ACK(nodeId, ackStatus, processTime);
-- 辅助索引：用于清理任务
CREATE INDEX IF NOT EXISTS IDX_CLS_ACK_CLEANUP ON HUB_CLUSTER_EVENT_ACK(tenantId, addTime);

