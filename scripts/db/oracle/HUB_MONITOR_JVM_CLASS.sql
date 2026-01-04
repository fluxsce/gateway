CREATE TABLE HUB_MONITOR_JVM_CLASS (
    classLoadingId VARCHAR2(32) NOT NULL, -- 类加载记录ID，主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    jvmResourceId VARCHAR2(100) NOT NULL, -- 关联的JVM资源ID
    
    -- 类加载统计
    loadedClassCount NUMBER(10,0) DEFAULT 0 NOT NULL, -- 当前已加载类数量
    totalLoadedClassCount NUMBER(19,0) DEFAULT 0 NOT NULL, -- 总加载类数量
    unloadedClassCount NUMBER(19,0) DEFAULT 0 NOT NULL, -- 已卸载类数量
    
    -- 比例指标
    classUnloadRatePercent NUMBER(5,2) DEFAULT 0.00, -- 类卸载率（百分比）
    classRetentionRatePercent NUMBER(5,2) DEFAULT 0.00, -- 类保留率（百分比）
    
    -- 配置状态
    verboseClassLoading VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否启用详细类加载输出(Y是,N否)
    
    -- 性能指标
    loadingRatePerHour NUMBER(10,2) DEFAULT 0.00, -- 每小时平均类加载数量
    loadingEfficiency NUMBER(5,2) DEFAULT 0.00, -- 类加载效率
    memoryEfficiency VARCHAR2(100) DEFAULT NULL, -- 内存使用效率评估
    loaderHealth VARCHAR2(50) DEFAULT NULL, -- 类加载器健康状况
    
    -- 健康状态
    healthyFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 类加载健康标记(Y健康,N异常)
    healthGrade VARCHAR2(20) DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR)
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
    
    CONSTRAINT PK_MONITOR_JVM_CLS PRIMARY KEY (tenantId, classLoadingId)
);

CREATE INDEX IDX_MONITOR_CLS_RES ON HUB_MONITOR_JVM_CLASS(jvmResourceId);
CREATE INDEX IDX_MONITOR_CLS_TIME ON HUB_MONITOR_JVM_CLASS(collectionTime);
CREATE INDEX IDX_MONITOR_CLS_HEALTH ON HUB_MONITOR_JVM_CLASS(healthyFlag, requiresAttentionFlag);
CREATE INDEX IDX_MONITOR_CLS_COUNT ON HUB_MONITOR_JVM_CLASS(loadedClassCount);

COMMENT ON TABLE HUB_MONITOR_JVM_CLASS IS 'JVM类加载监控表';
