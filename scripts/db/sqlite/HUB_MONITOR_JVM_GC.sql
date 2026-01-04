-- 每次采集插入一条记录，包含所有GC收集器的汇总数据
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_GC (
    gcSnapshotId TEXT NOT NULL, -- GC快照记录ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- GC累积统计（从JVM启动到当前采集时刻）
    collectionCount INTEGER DEFAULT 0 NOT NULL, -- GC总次数（累积，所有GC收集器汇总）
    collectionTimeMs INTEGER DEFAULT 0 NOT NULL, -- GC总耗时（毫秒，累积，所有GC收集器汇总）
    
    -- ===== jstat -gc 风格的内存区域数据（单位：KB） =====
    
    -- Survivor区
    s0c INTEGER DEFAULT 0, -- Survivor 0 区容量（KB）
    s1c INTEGER DEFAULT 0, -- Survivor 1 区容量（KB）
    s0u INTEGER DEFAULT 0, -- Survivor 0 区使用量（KB）
    s1u INTEGER DEFAULT 0, -- Survivor 1 区使用量（KB）
    
    -- Eden区
    ec INTEGER DEFAULT 0, -- Eden 区容量（KB）
    eu INTEGER DEFAULT 0, -- Eden 区使用量（KB）
    
    -- Old区
    oc INTEGER DEFAULT 0, -- Old 区容量（KB）
    ou INTEGER DEFAULT 0, -- Old 区使用量（KB）
    
    -- Metaspace
    mc INTEGER DEFAULT 0, -- Metaspace 容量（KB）
    mu INTEGER DEFAULT 0, -- Metaspace 使用量（KB）
    
    -- 压缩类空间
    ccsc INTEGER DEFAULT 0, -- 压缩类空间容量（KB）
    ccsu INTEGER DEFAULT 0, -- 压缩类空间使用量（KB）
    
    -- GC统计（jstat -gc 格式）
    ygc INTEGER DEFAULT 0, -- 年轻代GC次数
    ygct REAL DEFAULT 0.000, -- 年轻代GC总时间（秒）
    fgc INTEGER DEFAULT 0, -- Full GC次数
    fgct REAL DEFAULT 0.000, -- Full GC总时间（秒）
    gct REAL DEFAULT 0.000, -- 总GC时间（秒）
    
    -- 时间戳信息
    collectionTime TEXT NOT NULL, -- 数据采集时间戳
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, gcSnapshotId)
);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_GC_RES ON HUB_MONITOR_JVM_GC(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_GC_TIME ON HUB_MONITOR_JVM_GC(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_GC_RES_TIME ON HUB_MONITOR_JVM_GC(jvmResourceId, collectionTime);