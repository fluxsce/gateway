CREATE TABLE HUB_MONITOR_JVM_GC (
    gcSnapshotId VARCHAR2(32) NOT NULL, -- GC快照记录ID，主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    jvmResourceId VARCHAR2(100) NOT NULL, -- 关联的JVM资源ID
    
    -- GC累积统计（从JVM启动到当前采集时刻）
    collectionCount NUMBER(19,0) DEFAULT 0 NOT NULL, -- GC总次数（累积，所有GC收集器汇总）
    collectionTimeMs NUMBER(19,0) DEFAULT 0 NOT NULL, -- GC总耗时（毫秒，累积，所有GC收集器汇总）
    
    -- ===== jstat -gc 风格的内存区域数据（单位：KB） =====
    
    -- Survivor区
    s0c NUMBER(19,0) DEFAULT 0, -- Survivor 0 区容量（KB）
    s1c NUMBER(19,0) DEFAULT 0, -- Survivor 1 区容量（KB）
    s0u NUMBER(19,0) DEFAULT 0, -- Survivor 0 区使用量（KB）
    s1u NUMBER(19,0) DEFAULT 0, -- Survivor 1 区使用量（KB）
    
    -- Eden区
    ec NUMBER(19,0) DEFAULT 0, -- Eden 区容量（KB）
    eu NUMBER(19,0) DEFAULT 0, -- Eden 区使用量（KB）
    
    -- Old区
    oc NUMBER(19,0) DEFAULT 0, -- Old 区容量（KB）
    ou NUMBER(19,0) DEFAULT 0, -- Old 区使用量（KB）
    
    -- Metaspace
    mc NUMBER(19,0) DEFAULT 0, -- Metaspace 容量（KB）
    mu NUMBER(19,0) DEFAULT 0, -- Metaspace 使用量（KB）
    
    -- 压缩类空间
    ccsc NUMBER(19,0) DEFAULT 0, -- 压缩类空间容量（KB）
    ccsu NUMBER(19,0) DEFAULT 0, -- 压缩类空间使用量（KB）
    
    -- GC统计（jstat -gc 格式）
    ygc NUMBER(19,0) DEFAULT 0, -- 年轻代GC次数
    ygct NUMBER(10,3) DEFAULT 0.000, -- 年轻代GC总时间（秒）
    fgc NUMBER(19,0) DEFAULT 0, -- Full GC次数
    fgct NUMBER(10,3) DEFAULT 0.000, -- Full GC总时间（秒）
    gct NUMBER(10,3) DEFAULT 0.000, -- 总GC时间（秒）
    
    -- 时间戳信息
    collectionTime DATE NOT NULL, -- 数据采集时间戳
    
    -- 通用字段
    addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
    addWho VARCHAR2(32) DEFAULT NULL, -- 创建人ID
    editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
    editWho VARCHAR2(32) DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag VARCHAR2(32) DEFAULT NULL, -- 操作序列标识
    currentVersion NUMBER(10,0) DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
    noteText VARCHAR2(500) DEFAULT NULL, -- 备注信息
    
    CONSTRAINT PK_MONITOR_JVM_GC PRIMARY KEY (tenantId, gcSnapshotId)
);

CREATE INDEX IDX_MONITOR_GC_RES ON HUB_MONITOR_JVM_GC(jvmResourceId);
CREATE INDEX IDX_MONITOR_GC_TIME ON HUB_MONITOR_JVM_GC(collectionTime);
CREATE INDEX IDX_MONITOR_GC_RES_TIME ON HUB_MONITOR_JVM_GC(jvmResourceId, collectionTime);

COMMENT ON TABLE HUB_MONITOR_JVM_GC IS 'JVM GC快照表（jstat -gc风格，每次采集一条汇总记录）';
