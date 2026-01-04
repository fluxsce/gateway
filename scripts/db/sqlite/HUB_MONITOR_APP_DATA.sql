-- 对应 ThirdPartyMonitorData 采集的所有监控数据
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_APP_DATA (
    appDataId TEXT NOT NULL, -- 应用监控数据ID，主键
    tenantId TEXT NOT NULL, -- 租户ID
    jvmResourceId TEXT NOT NULL, -- 关联的JVM资源ID
    
    -- 数据分类标识
    dataType TEXT NOT NULL, -- 数据类型(THREAD_POOL:线程池/CONNECTION_POOL:连接池/CUSTOM_METRIC:自定义指标/CACHE_POOL:缓存池/MESSAGE_QUEUE:消息队列)
    dataName TEXT NOT NULL, -- 数据名称（如：线程池名称、指标名称等）
    dataCategory TEXT DEFAULT NULL, -- 数据分类（如：业务线程池/IO线程池/业务指标/技术指标）
    
    -- 监控数据（JSON格式存储，支持不同类型的数据结构）
    dataJson TEXT NOT NULL, -- 监控数据，JSON格式，包含具体的监控指标和值
    
    -- 核心指标（从JSON中提取的关键指标，便于查询和索引）
    primaryValue REAL DEFAULT NULL, -- 主要指标值（如：使用率、数量等）
    secondaryValue REAL DEFAULT NULL, -- 次要指标值（如：最大值、平均值等）
    statusValue TEXT DEFAULT NULL, -- 状态值（如：健康状态、连接状态等）
    
    -- 健康状态
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- 健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- 健康等级(EXCELLENT/GOOD/FAIR/POOR/CRITICAL)
    requiresAttentionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresAttentionFlag IN ('Y','N')), -- 是否需要立即关注(Y是,N否)
    
    -- 标签和维度（便于分组查询）
    tagsJson TEXT DEFAULT NULL, -- 标签信息，JSON格式（如：{"poolType":"business","environment":"prod"}）
    
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
    
    PRIMARY KEY (tenantId, appDataId)
);

CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_RES ON HUB_MONITOR_APP_DATA(jvmResourceId);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_TYPE ON HUB_MONITOR_APP_DATA(dataType);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_NAME ON HUB_MONITOR_APP_DATA(dataName);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_TIME ON HUB_MONITOR_APP_DATA(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_HEALTH ON HUB_MONITOR_APP_DATA(healthyFlag, requiresAttentionFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_PRIMARY ON HUB_MONITOR_APP_DATA(primaryValue);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_STATUS ON HUB_MONITOR_APP_DATA(statusValue);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_APP_DATA_COMPOSITE ON HUB_MONITOR_APP_DATA(jvmResourceId, dataType, dataName, collectionTime);
-- =========================================
-- 基于FRP架构的隧道管理系统 - SQLite数据库表结构设计
-- 参考FRP（Fast Reverse Proxy）设计模式
-- 遵循naming-convention.md数据库规范
-- =========================================

-- =========================================
-- 1. 隧道服务器表（控制端口）
-- =========================================