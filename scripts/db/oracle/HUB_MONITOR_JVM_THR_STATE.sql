CREATE TABLE HUB_MONITOR_JVM_THR_STATE (
    threadStateId VARCHAR2(32) NOT NULL, -- 线程状态记录ID，主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    jvmThreadId VARCHAR2(32) NOT NULL, -- 关联的JVM线程记录ID
    jvmResourceId VARCHAR2(100) NOT NULL, -- 关联的JVM资源ID
    
    -- 线程状态分布
    newThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- NEW状态线程数
    runnableThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- RUNNABLE状态线程数
    blockedThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- BLOCKED状态线程数
    waitingThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- WAITING状态线程数
    timedWaitingThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- TIMED_WAITING状态线程数
    terminatedThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- TERMINATED状态线程数
    totalThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- 总线程数
    
    -- 比例指标
    activeThreadRatioPercent NUMBER(5,2) DEFAULT 0.00, -- 活跃线程比例（百分比）
    blockedThreadRatioPercent NUMBER(5,2) DEFAULT 0.00, -- 阻塞线程比例（百分比）
    waitingThreadRatioPercent NUMBER(5,2) DEFAULT 0.00, -- 等待状态线程比例（百分比）
    
    -- 健康状态
    healthyFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 线程状态健康标记(Y健康,N异常)
    healthGrade VARCHAR2(20) DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR)
    
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
    
    CONSTRAINT PK_MONITOR_JVM_THR_ST PRIMARY KEY (tenantId, threadStateId)
);

CREATE INDEX IDX_MONITOR_THRST_THR ON HUB_MONITOR_JVM_THR_STATE(jvmThreadId);
CREATE INDEX IDX_MONITOR_THRST_RES ON HUB_MONITOR_JVM_THR_STATE(jvmResourceId);
CREATE INDEX IDX_MONITOR_THRST_TIME ON HUB_MONITOR_JVM_THR_STATE(collectionTime);
CREATE INDEX IDX_MONITOR_THRST_BLOCK ON HUB_MONITOR_JVM_THR_STATE(blockedThreadCount);

COMMENT ON TABLE HUB_MONITOR_JVM_THR_STATE IS 'JVM线程状态统计表';
