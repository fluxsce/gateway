CREATE TABLE HUB_MONITOR_APP_DATA (
    appDataId VARCHAR2(32) NOT NULL, -- 应用监控数据ID，主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    jvmResourceId VARCHAR2(100) NOT NULL, -- 关联的JVM资源ID
    
    -- 数据分类标识
    dataType VARCHAR2(50) NOT NULL, -- 数据类型(THREAD_POOL:线程池/CONNECTION_POOL:连接池/CUSTOM_METRIC:自定义指标/CACHE_POOL:缓存池/MESSAGE_QUEUE:消息队列)
    dataName VARCHAR2(100) NOT NULL, -- 数据名称（如：线程池名称、指标名称等）
    dataCategory VARCHAR2(50) DEFAULT NULL, -- 数据分类（如：业务线程池/IO线程池/业务指标/技术指标）
    
    -- 监控数据（JSON格式存储，支持不同类型的数据结构）
    dataJson CLOB NOT NULL, -- 监控数据，JSON格式，包含具体的监控指标和值
    
    -- 核心指标（从JSON中提取的关键指标，便于查询和索引）
    primaryValue NUMBER(20,4) DEFAULT NULL, -- 主要指标值（如：使用率、数量等）
    secondaryValue NUMBER(20,4) DEFAULT NULL, -- 次要指标值（如：最大值、平均值等）
    statusValue VARCHAR2(50) DEFAULT NULL, -- 状态值（如：健康状态、连接状态等）
    
    -- 健康状态
    healthyFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 健康标记(Y健康,N异常)
    healthGrade VARCHAR2(20) DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR/CRITICAL)
    requiresAttentionFlag VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否需要立即关注(Y是,N否)
    
    -- 标签和维度（便于分组查询）
    tagsJson CLOB DEFAULT NULL, -- 标签信息，JSON格式（如：{"poolType":"business","environment":"prod"}）
    
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
    
    CONSTRAINT PK_MONITOR_APP_DATA PRIMARY KEY (tenantId, appDataId)
);

CREATE INDEX IDX_MONITOR_APP_DATA_RES ON HUB_MONITOR_APP_DATA(jvmResourceId);
CREATE INDEX IDX_MONITOR_APP_DATA_TYPE ON HUB_MONITOR_APP_DATA(dataType);
CREATE INDEX IDX_MONITOR_APP_DATA_NAME ON HUB_MONITOR_APP_DATA(dataName);
CREATE INDEX IDX_MONITOR_APP_DATA_TIME ON HUB_MONITOR_APP_DATA(collectionTime);
CREATE INDEX IDX_MONITOR_APP_DATA_HEALTH ON HUB_MONITOR_APP_DATA(healthyFlag, requiresAttentionFlag);
CREATE INDEX IDX_MONITOR_APP_DATA_PRIMARY ON HUB_MONITOR_APP_DATA(primaryValue);
CREATE INDEX IDX_MONITOR_APP_DATA_STATUS ON HUB_MONITOR_APP_DATA(statusValue);
CREATE INDEX IDX_MONITOR_APP_DATA_COMPOSITE ON HUB_MONITOR_APP_DATA(jvmResourceId, dataType, dataName, collectionTime);

COMMENT ON TABLE HUB_MONITOR_APP_DATA IS '应用监控数据表';



