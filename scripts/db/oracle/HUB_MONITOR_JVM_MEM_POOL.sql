CREATE TABLE HUB_MONITOR_JVM_MEM_POOL (
    memoryPoolId VARCHAR2(32) NOT NULL, -- 内存池记录ID，主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    jvmResourceId VARCHAR2(100) NOT NULL, -- 关联的JVM资源ID
    
    -- 内存池基本信息
    poolName VARCHAR2(100) NOT NULL, -- 内存池名称
    poolType VARCHAR2(20) NOT NULL, -- 内存池类型(HEAP/NON_HEAP)
    poolCategory VARCHAR2(50) DEFAULT NULL, -- 内存池分类（年轻代/老年代/元数据空间/代码缓存/其他）
    
    -- 当前使用情况
    currentInitBytes NUMBER(19,0) DEFAULT 0 NOT NULL, -- 当前初始内存（字节）
    currentUsedBytes NUMBER(19,0) DEFAULT 0 NOT NULL, -- 当前已使用内存（字节）
    currentCommittedBytes NUMBER(19,0) DEFAULT 0 NOT NULL, -- 当前已提交内存（字节）
    currentMaxBytes NUMBER(19,0) DEFAULT -1 NOT NULL, -- 当前最大内存（字节）
    currentUsagePercent NUMBER(5,2) DEFAULT 0.00 NOT NULL, -- 当前使用率（百分比）
    
    -- 峰值使用情况
    peakInitBytes NUMBER(19,0) DEFAULT 0, -- 峰值初始内存（字节）
    peakUsedBytes NUMBER(19,0) DEFAULT 0, -- 峰值已使用内存（字节）
    peakCommittedBytes NUMBER(19,0) DEFAULT 0, -- 峰值已提交内存（字节）
    peakMaxBytes NUMBER(19,0) DEFAULT -1, -- 峰值最大内存（字节）
    peakUsagePercent NUMBER(5,2) DEFAULT 0.00, -- 峰值使用率（百分比）
    
    -- 阈值监控
    usageThresholdSupported VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否支持使用阈值监控(Y是,N否)
    usageThresholdBytes NUMBER(19,0) DEFAULT 0, -- 使用阈值（字节）
    usageThresholdCount NUMBER(19,0) DEFAULT 0, -- 使用阈值超越次数
    collectionUsageSupported VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否支持收集使用量监控(Y是,N否)
    
    -- 健康状态
    healthyFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 内存池健康标记(Y健康,N异常)
    
    -- 时间字段
    collectionTime DATE NOT NULL, -- 数据采集时间
    
    -- 通用字段
    addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
    addWho VARCHAR2(32) DEFAULT NULL, -- 创建人ID
    editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
    editWho VARCHAR2(32) DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag VARCHAR2(32) DEFAULT NULL, -- 操作序列标识
    currentVersion NUMBER(10,0) DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
    noteText VARCHAR2(500) DEFAULT NULL, -- 备注信息
    
    CONSTRAINT PK_MONITOR_JVM_MEM_POOL PRIMARY KEY (tenantId, memoryPoolId)
);

CREATE INDEX IDX_MONITOR_POOL_RES ON HUB_MONITOR_JVM_MEM_POOL(jvmResourceId);
CREATE INDEX IDX_MONITOR_POOL_NAME ON HUB_MONITOR_JVM_MEM_POOL(poolName);
CREATE INDEX IDX_MONITOR_POOL_TYPE ON HUB_MONITOR_JVM_MEM_POOL(poolType);
CREATE INDEX IDX_MONITOR_POOL_CAT ON HUB_MONITOR_JVM_MEM_POOL(poolCategory);
CREATE INDEX IDX_MONITOR_POOL_TIME ON HUB_MONITOR_JVM_MEM_POOL(collectionTime);

COMMENT ON TABLE HUB_MONITOR_JVM_MEM_POOL IS 'JVM内存池监控表';
