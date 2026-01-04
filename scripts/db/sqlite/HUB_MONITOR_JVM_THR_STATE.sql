-- 存储不同状态下的线程数量分布
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_THR_STATE (
    threadStateId TEXT NOT NULL, -- 线程状态记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmThreadId TEXT NOT NULL, -- 关联的JVM线程记录ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 线程状态分布
    newThreadCount INTEGER DEFAULT 0 NOT NULL, -- NEW状态线程数
    runnableThreadCount INTEGER DEFAULT 0 NOT NULL, -- RUNNABLE状态线程数
    blockedThreadCount INTEGER DEFAULT 0 NOT NULL, -- BLOCKED状态线程数
    waitingThreadCount INTEGER DEFAULT 0 NOT NULL, -- WAITING状态线程数
    timedWaitingThreadCount INTEGER DEFAULT 0 NOT NULL, -- TIMED_WAITING状态线程数
    terminatedThreadCount INTEGER DEFAULT 0 NOT NULL, -- TERMINATED状态线程数
    totalThreadCount INTEGER DEFAULT 0 NOT NULL, -- 总线程数
    
    -- 比例指标
    activeThreadRatioPercent REAL DEFAULT 0.00, -- 活跃线程比例（百分比）
    blockedThreadRatioPercent REAL DEFAULT 0.00, -- 阻塞线程比例（百分比）
    waitingThreadRatioPercent REAL DEFAULT 0.00, -- 等待状态线程比例（百分比）
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 线程状态健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR)
    
    -- 时间字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, threadStateId)
);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THRST_THR ON HUB_MONITOR_JVM_THR_STATE(jvmThreadId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THRST_RES ON HUB_MONITOR_JVM_THR_STATE(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THRST_TIME ON HUB_MONITOR_JVM_THR_STATE(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THRST_BLOCK ON HUB_MONITOR_JVM_THR_STATE(blockedThreadCount);