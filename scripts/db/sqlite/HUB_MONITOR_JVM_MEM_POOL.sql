-- 存储具体内存池的详细使用情况（Eden、Survivor、Old Gen、Metaspace等）
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_MEM_POOL (
    memoryPoolId TEXT NOT NULL, -- 内存池记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 内存池基本信息
    poolName TEXT NOT NULL, -- 内存池名称
    poolType TEXT NOT NULL, -- 内存池类型(HEAP/NON_HEAP)
    poolCategory TEXT DEFAULT NULL, -- 内存池分类（年轻代/老年代/元数据空间/代码缓存/其他）
    
    -- 当前使用情况
    currentInitBytes INTEGER DEFAULT 0 NOT NULL, -- 当前初始内存（字节）
    currentUsedBytes INTEGER DEFAULT 0 NOT NULL, -- 当前已使用内存（字节）
    currentCommittedBytes INTEGER DEFAULT 0 NOT NULL, -- 当前已提交内存（字节）
    currentMaxBytes INTEGER DEFAULT -1 NOT NULL, -- 当前最大内存（字节）
    currentUsagePercent REAL DEFAULT 0.00 NOT NULL, -- 当前使用率（百分比）
    
    -- 峰值使用情况
    peakInitBytes INTEGER DEFAULT 0, -- 峰值初始内存（字节）
    peakUsedBytes INTEGER DEFAULT 0, -- 峰值已使用内存（字节）
    peakCommittedBytes INTEGER DEFAULT 0, -- 峰值已提交内存（字节）
    peakMaxBytes INTEGER DEFAULT -1, -- 峰值最大内存（字节）
    peakUsagePercent REAL DEFAULT 0.00, -- 峰值使用率（百分比）
    
    -- 阈值监控
    usageThresholdSupported TEXT DEFAULT 'N' NOT NULL CHECK(usageThresholdSupported IN ('Y','N')), -- 是否支持使用阈值监控(Y是,N否)
    usageThresholdBytes INTEGER DEFAULT 0, -- 使用阈值（字节）
    usageThresholdCount INTEGER DEFAULT 0, -- 使用阈值超越次数
    collectionUsageSupported TEXT DEFAULT 'N' NOT NULL CHECK(collectionUsageSupported IN ('Y','N')), -- 是否支持收集使用量监控(Y是,N否)
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 内存池健康标记(Y健康,N异常)
    
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
    
    PRIMARY KEY (tenantId, memoryPoolId)
);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_RES ON HUB_MONITOR_JVM_MEM_POOL(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_NAME ON HUB_MONITOR_JVM_MEM_POOL(poolName);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_TYPE ON HUB_MONITOR_JVM_MEM_POOL(poolType);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_CAT ON HUB_MONITOR_JVM_MEM_POOL(poolCategory);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_POOL_TIME ON HUB_MONITOR_JVM_MEM_POOL(collectionTime);