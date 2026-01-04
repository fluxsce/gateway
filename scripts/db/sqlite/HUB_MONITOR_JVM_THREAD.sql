-- 存储JVM线程的详细监控数据
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_THREAD (
    jvmThreadId TEXT NOT NULL, -- JVM线程记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 基础线程统计
    currentThreadCount INTEGER DEFAULT 0 NOT NULL, -- 当前线程数
    daemonThreadCount INTEGER DEFAULT 0 NOT NULL, -- 守护线程数
    userThreadCount INTEGER DEFAULT 0 NOT NULL, -- 用户线程数
    peakThreadCount INTEGER DEFAULT 0 NOT NULL, -- 峰值线程数
    totalStartedThreadCount INTEGER DEFAULT 0 NOT NULL, -- 总启动线程数
    
    -- 性能指标
    threadGrowthRatePercent REAL DEFAULT 0.00, -- 线程增长率（百分比）
    daemonThreadRatioPercent REAL DEFAULT 0.00, -- 守护线程比例（百分比）
    
    -- 监控功能支持状态
    cpuTimeSupported TEXT DEFAULT 'N' NOT NULL CHECK(cpuTimeSupported IN ('Y','N')), -- CPU时间监控是否支持(Y是,N否)
    cpuTimeEnabled TEXT DEFAULT 'N' NOT NULL CHECK(cpuTimeEnabled IN ('Y','N')), -- CPU时间监控是否启用(Y是,N否)
    memoryAllocSupported TEXT DEFAULT 'N' NOT NULL CHECK(memoryAllocSupported IN ('Y','N')), -- 内存分配监控是否支持(Y是,N否)
    memoryAllocEnabled TEXT DEFAULT 'N' NOT NULL CHECK(memoryAllocEnabled IN ('Y','N')), -- 内存分配监控是否启用(Y是,N否)
    contentionSupported TEXT DEFAULT 'N' NOT NULL CHECK(contentionSupported IN ('Y','N')), -- 争用监控是否支持(Y是,N否)
    contentionEnabled TEXT DEFAULT 'N' NOT NULL CHECK(contentionEnabled IN ('Y','N')), -- 争用监控是否启用(Y是,N否)
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 线程健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- 线程健康等级(EXCELLENT/GOOD/FAIR/POOR)
    requiresAttentionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresAttentionFlag IN ('Y','N')), -- 是否需要立即关注(Y是,N否)
    potentialIssuesJson TEXT DEFAULT NULL, -- 潜在问题列表，JSON格式
    
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
    
    PRIMARY KEY (tenantId, jvmThreadId)
);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THR_RES ON HUB_MONITOR_JVM_THREAD(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THR_TIME ON HUB_MONITOR_JVM_THREAD(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THR_HEALTH ON HUB_MONITOR_JVM_THREAD(healthyFlag, requiresAttentionFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_THR_COUNT ON HUB_MONITOR_JVM_THREAD(currentThreadCount);