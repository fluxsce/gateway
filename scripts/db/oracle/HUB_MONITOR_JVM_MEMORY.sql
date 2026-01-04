CREATE TABLE HUB_MONITOR_JVM_MEMORY (
    jvmMemoryId VARCHAR2(32) NOT NULL, -- JVM内存记录ID，主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    jvmResourceId VARCHAR2(100) NOT NULL, -- 关联的JVM资源ID
    
    -- 内存类型
    memoryType VARCHAR2(20) NOT NULL, -- 内存类型(HEAP/NON_HEAP)
    
    -- 内存使用情况（字节）
    initMemoryBytes NUMBER(19,0) DEFAULT 0 NOT NULL, -- 初始内存大小（字节）
    usedMemoryBytes NUMBER(19,0) DEFAULT 0 NOT NULL, -- 已使用内存大小（字节）
    committedMemoryBytes NUMBER(19,0) DEFAULT 0 NOT NULL, -- 已提交内存大小（字节）
    maxMemoryBytes NUMBER(19,0) DEFAULT -1 NOT NULL, -- 最大内存大小（字节），-1表示无限制
    
    -- 计算指标
    usagePercent NUMBER(5,2) DEFAULT 0.00 NOT NULL, -- 内存使用率（百分比）
    healthyFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 内存健康标记(Y健康,N异常)
    
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
    
    CONSTRAINT PK_MONITOR_JVM_MEM PRIMARY KEY (tenantId, jvmMemoryId)
);

CREATE INDEX IDX_MONITOR_MEM_RES ON HUB_MONITOR_JVM_MEMORY(jvmResourceId);
CREATE INDEX IDX_MONITOR_MEM_TYPE ON HUB_MONITOR_JVM_MEMORY(memoryType);
CREATE INDEX IDX_MONITOR_MEM_TIME ON HUB_MONITOR_JVM_MEMORY(collectionTime);
CREATE INDEX IDX_MONITOR_MEM_USAGE ON HUB_MONITOR_JVM_MEMORY(usagePercent);

COMMENT ON TABLE HUB_MONITOR_JVM_MEMORY IS 'JVM内存监控表';
