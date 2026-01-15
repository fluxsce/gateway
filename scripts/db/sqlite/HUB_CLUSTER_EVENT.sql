
-- 集群事件表 - 存储集群中各节点发布的事件
CREATE TABLE IF NOT EXISTS HUB_CLUSTER_EVENT (
  -- 主键和租户信息
  eventId TEXT NOT NULL,
  tenantId TEXT NOT NULL,
  
  -- 事件来源(发布者)
  sourceNodeId TEXT NOT NULL,
  sourceNodeIp TEXT,
  
  -- 事件信息
  eventType TEXT NOT NULL,
  eventAction TEXT NOT NULL,
  eventPayload TEXT,
  
  -- 事件时间
  eventTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expireTime DATETIME,
  
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
  reserved4 TEXT,
  reserved5 TEXT,
  
  PRIMARY KEY (tenantId, eventId)
);

-- =====================================================
-- HUB_CLUSTER_EVENT 表说明
-- =====================================================
-- 
-- 集群事件表用于实现基于数据库的集群节点间通信
-- 每个节点都可以发布事件，其他节点通过轮询消费事件
--
-- 事件类型(eventType)说明：
-- 1. ROUTE_CONFIG: 路由配置变更
-- 2. SERVICE_CONFIG: 服务配置变更
-- 3. FILTER_CONFIG: 过滤器配置变更
-- 4. CACHE_REFRESH: 缓存刷新通知
-- 5. NODE_HEARTBEAT: 节点心跳(可选)
--
-- 事件动作(eventAction)说明：
-- 1. CREATE: 新增
-- 2. UPDATE: 更新
-- 3. DELETE: 删除
-- 4. REFRESH: 刷新
-- 5. INVALIDATE: 失效
--
-- =====================================================

-- =====================================================
-- HUB_CLUSTER_EVENT 索引（优化后）
-- =====================================================
-- 优化后的复合索引：支持高效的待处理事件查询
-- 查询模式: WHERE tenantId=? AND activeFlag='Y' AND eventTime>? AND sourceNodeId!=?
CREATE INDEX IF NOT EXISTS IDX_CLS_EVT_QUERY ON HUB_CLUSTER_EVENT(tenantId, activeFlag, eventTime, eventId);
-- 辅助索引：用于特定场景查询
CREATE INDEX IF NOT EXISTS IDX_CLS_EVT_SOURCE ON HUB_CLUSTER_EVENT(sourceNodeId);
CREATE INDEX IF NOT EXISTS IDX_CLS_EVT_TYPE ON HUB_CLUSTER_EVENT(eventType, eventTime);
CREATE INDEX IF NOT EXISTS IDX_CLS_EVT_EXPIRE ON HUB_CLUSTER_EVENT(expireTime);

