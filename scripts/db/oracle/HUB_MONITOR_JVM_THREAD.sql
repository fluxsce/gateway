CREATE TABLE HUB_MONITOR_JVM_THREAD (
    jvmThreadId VARCHAR2(32) NOT NULL, -- JVM线程记录ID，主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    jvmResourceId VARCHAR2(100) NOT NULL, -- 关联的JVM资源ID
    
    -- 基础线程统计
    currentThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- 当前线程数
    daemonThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- 守护线程数
    userThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- 用户线程数
    peakThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- 峰值线程数
    totalStartedThreadCount NUMBER(19,0) DEFAULT 0 NOT NULL, -- 总启动线程数
    
    -- 性能指标
    threadGrowthRatePercent NUMBER(5,2) DEFAULT 0.00, -- 线程增长率（百分比）
    daemonThreadRatioPercent NUMBER(5,2) DEFAULT 0.00, -- 守护线程比例（百分比）
    
    -- 监控功能支持状态
    cpuTimeSupported VARCHAR2(1) DEFAULT 'N' NOT NULL, -- CPU时间监控是否支持(Y是,N否)
    cpuTimeEnabled VARCHAR2(1) DEFAULT 'N' NOT NULL, -- CPU时间监控是否启用(Y是,N否)
    memoryAllocSupported VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 内存分配监控是否支持(Y是,N否)
    memoryAllocEnabled VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 内存分配监控是否启用(Y是,N否)
    contentionSupported VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 争用监控是否支持(Y是,N否)
    contentionEnabled VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 争用监控是否启用(Y是,N否)
    
    -- 健康状态
    healthyFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 线程健康标记(Y健康,N异常)
    healthGrade VARCHAR2(20) DEFAULT NULL, -- 线程健康等级(EXCELLENT/GOOD/FAIR/POOR)
    requiresAttentionFlag VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否需要立即关注(Y是,N否)
    potentialIssuesJson CLOB DEFAULT NULL, -- 潜在问题列表，JSON格式
    
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
    
    CONSTRAINT PK_MONITOR_JVM_THR PRIMARY KEY (tenantId, jvmThreadId)
);

CREATE INDEX IDX_MONITOR_THR_RES ON HUB_MONITOR_JVM_THREAD(jvmResourceId);
CREATE INDEX IDX_MONITOR_THR_TIME ON HUB_MONITOR_JVM_THREAD(collectionTime);
CREATE INDEX IDX_MONITOR_THR_HEALTH ON HUB_MONITOR_JVM_THREAD(healthyFlag, requiresAttentionFlag);
CREATE INDEX IDX_MONITOR_THR_COUNT ON HUB_MONITOR_JVM_THREAD(currentThreadCount);

COMMENT ON TABLE HUB_MONITOR_JVM_THREAD IS 'JVM线程监控表';
