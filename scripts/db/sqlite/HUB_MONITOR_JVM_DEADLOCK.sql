-- 存储JVM中检测到的死锁情况
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_DEADLOCK (
    deadlockId TEXT NOT NULL, -- 死锁记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmThreadId TEXT NOT NULL, -- 关联的JVM线程记录ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 死锁基本信息
    hasDeadlockFlag TEXT DEFAULT 'N' NOT NULL CHECK(hasDeadlockFlag IN ('Y','N')), -- 是否检测到死锁(Y是,N否)
    deadlockThreadCount INTEGER DEFAULT 0 NOT NULL, -- 死锁线程数量
    deadlockThreadIds TEXT DEFAULT NULL, -- 死锁线程ID列表，逗号分隔
    deadlockThreadNames TEXT DEFAULT NULL, -- 死锁线程名称列表，逗号分隔
    
    -- 死锁严重程度
    severityLevel TEXT DEFAULT NULL, -- 严重程度(LOW/MEDIUM/HIGH/CRITICAL)
    severityDescription TEXT DEFAULT NULL, -- 严重程度描述
    affectedThreadGroups INTEGER DEFAULT 0, -- 影响的线程组数量
    
    -- 时间信息
    detectionTime TEXT DEFAULT NULL, -- 死锁检测时间
    deadlockDurationMs INTEGER DEFAULT 0, -- 死锁持续时间（毫秒）
    collectionTime TEXT NOT NULL, -- 数据采集时间
    
    -- 诊断信息
    descriptionText TEXT DEFAULT NULL, -- 死锁描述信息
    recommendedAction TEXT DEFAULT NULL, -- 建议的解决方案
    alertLevel TEXT DEFAULT NULL, -- 告警级别(INFO/WARNING/ERROR/CRITICAL/EMERGENCY)
    requiresActionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresActionFlag IN ('Y','N')), -- 是否需要立即处理(Y是,N否)
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, deadlockId)
);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_THR ON HUB_MONITOR_JVM_DEADLOCK(jvmThreadId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_RES ON HUB_MONITOR_JVM_DEADLOCK(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_TIME ON HUB_MONITOR_JVM_DEADLOCK(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_FLAG ON HUB_MONITOR_JVM_DEADLOCK(hasDeadlockFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_SEV ON HUB_MONITOR_JVM_DEADLOCK(severityLevel);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_DL_ALERT ON HUB_MONITOR_JVM_DEADLOCK(alertLevel);