-- 存储JVM类加载器的统计信息
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_CLASS (
    classLoadingId TEXT NOT NULL, -- 类加载记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 类加载统计
    loadedClassCount INTEGER DEFAULT 0 NOT NULL, -- 当前已加载类数量
    totalLoadedClassCount INTEGER DEFAULT 0 NOT NULL, -- 总加载类数量
    unloadedClassCount INTEGER DEFAULT 0 NOT NULL, -- 已卸载类数量
    
    -- 比例指标
    classUnloadRatePercent REAL DEFAULT 0.00, -- 类卸载率（百分比）
    classRetentionRatePercent REAL DEFAULT 0.00, -- 类保留率（百分比）
    
    -- 配置状态
    verboseClassLoading TEXT DEFAULT 'N' NOT NULL CHECK(verboseClassLoading IN ('Y','N')), -- 是否启用详细类加载输出(Y是,N否)
    
    -- 性能指标
    loadingRatePerHour REAL DEFAULT 0.00, -- 每小时平均类加载数量
    loadingEfficiency REAL DEFAULT 0.00, -- 类加载效率
    memoryEfficiency TEXT DEFAULT NULL, -- 内存使用效率评估
    loaderHealth TEXT DEFAULT NULL, -- 类加载器健康状况
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 类加载健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR)
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
    
    PRIMARY KEY (tenantId, classLoadingId)
);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_CLS_RES ON HUB_MONITOR_JVM_CLASS(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_CLS_TIME ON HUB_MONITOR_JVM_CLASS(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_CLS_HEALTH ON HUB_MONITOR_JVM_CLASS(healthyFlag, requiresAttentionFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_CLS_COUNT ON HUB_MONITOR_JVM_CLASS(loadedClassCount);