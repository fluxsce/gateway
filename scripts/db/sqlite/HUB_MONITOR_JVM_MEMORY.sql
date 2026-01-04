-- 存储JVM堆内存和非堆内存的使用情况
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_MEMORY (
    jvmMemoryId TEXT NOT NULL, -- JVM内存记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 内存类型
    memoryType TEXT NOT NULL, -- 内存类型(HEAP/NON_HEAP)
    
    -- 内存使用情况（字节）
    initMemoryBytes INTEGER DEFAULT 0 NOT NULL, -- 初始内存大小（字节）
    usedMemoryBytes INTEGER DEFAULT 0 NOT NULL, -- 已使用内存大小（字节）
    committedMemoryBytes INTEGER DEFAULT 0 NOT NULL, -- 已提交内存大小（字节）
    maxMemoryBytes INTEGER DEFAULT -1 NOT NULL, -- 最大内存大小（字节），-1表示无限制
    
    -- 计算指标
    usagePercent REAL DEFAULT 0.00 NOT NULL, -- 内存使用率（百分比）
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 内存健康标记(Y健康,N异常)
    
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
    
    PRIMARY KEY (tenantId, jvmMemoryId)
);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_MEM_RES ON HUB_MONITOR_JVM_MEMORY(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_MEM_TYPE ON HUB_MONITOR_JVM_MEMORY(memoryType);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_MEM_TIME ON HUB_MONITOR_JVM_MEMORY(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_MEM_USAGE ON HUB_MONITOR_JVM_MEMORY(usagePercent);