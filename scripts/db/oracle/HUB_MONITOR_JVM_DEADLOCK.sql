CREATE TABLE HUB_MONITOR_JVM_DEADLOCK (
    deadlockId VARCHAR2(32) NOT NULL, -- 死锁记录ID，主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    jvmThreadId VARCHAR2(32) NOT NULL, -- 关联的JVM线程记录ID
    jvmResourceId VARCHAR2(100) NOT NULL, -- 关联的JVM资源ID
    
    -- 死锁基本信息
    hasDeadlockFlag VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否检测到死锁(Y是,N否)
    deadlockThreadCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- 死锁线程数量
    deadlockThreadIds CLOB DEFAULT NULL, -- 死锁线程ID列表，逗号分隔
    deadlockThreadNames CLOB DEFAULT NULL, -- 死锁线程名称列表，逗号分隔
    
    -- 死锁严重程度
    severityLevel VARCHAR2(20) DEFAULT NULL, -- 严重程度(LOW/MEDIUM/HIGH/CRITICAL)
    severityDescription VARCHAR2(200) DEFAULT NULL, -- 严重程度描述
    affectedThreadGroups NUMBER(10,0) DEFAULT 0, -- 影响的线程组数量
    
    -- 时间信息
    detectionTime DATE DEFAULT NULL, -- 死锁检测时间
    deadlockDurationMs NUMBER(19,0) DEFAULT 0, -- 死锁持续时间（毫秒）
    collectionTime DATE NOT NULL, -- 数据采集时间
    
    -- 诊断信息
    descriptionText VARCHAR2(500) DEFAULT NULL, -- 死锁描述信息
    recommendedAction VARCHAR2(500) DEFAULT NULL, -- 建议的解决方案
    alertLevel VARCHAR2(20) DEFAULT NULL, -- 告警级别(INFO/WARNING/ERROR/CRITICAL/EMERGENCY)
    requiresActionFlag VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否需要立即处理(Y是,N否)
    
    -- 通用字段
    addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
    addWho VARCHAR2(32) DEFAULT NULL, -- 创建人ID
    editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
    editWho VARCHAR2(32) DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag VARCHAR2(32) DEFAULT NULL, -- 操作序列标识
    currentVersion NUMBER(10,0) DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
    noteText VARCHAR2(500) DEFAULT NULL, -- 备注信息
    
    CONSTRAINT PK_MONITOR_JVM_DL PRIMARY KEY (tenantId, deadlockId)
);

CREATE INDEX IDX_MONITOR_DL_THR ON HUB_MONITOR_JVM_DEADLOCK(jvmThreadId);
CREATE INDEX IDX_MONITOR_DL_RES ON HUB_MONITOR_JVM_DEADLOCK(jvmResourceId);
CREATE INDEX IDX_MONITOR_DL_TIME ON HUB_MONITOR_JVM_DEADLOCK(collectionTime);
CREATE INDEX IDX_MONITOR_DL_FLAG ON HUB_MONITOR_JVM_DEADLOCK(hasDeadlockFlag);
CREATE INDEX IDX_MONITOR_DL_SEV ON HUB_MONITOR_JVM_DEADLOCK(severityLevel);
CREATE INDEX IDX_MONITOR_DL_ALERT ON HUB_MONITOR_JVM_DEADLOCK(alertLevel);

COMMENT ON TABLE HUB_MONITOR_JVM_DEADLOCK IS 'JVM死锁检测信息表';
