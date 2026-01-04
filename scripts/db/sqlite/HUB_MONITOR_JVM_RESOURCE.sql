-- 存储JVM整体资源监控信息的快照数据
-- ==========================================
CREATE TABLE IF NOT EXISTS HUB_MONITOR_JVM_RESOURCE (
    jvmResourceId TEXT NOT NULL, -- JVM资源记录ID（由应用端生成的唯一标识），主键
    tenantId TEXT NOT NULL, -- 租户ID
    serviceGroupId TEXT NOT NULL, -- 服务分组ID，主键
    
    -- 应用标识信息
    applicationName TEXT NOT NULL, -- 应用名称
    groupName TEXT NOT NULL, -- 分组名称
    hostName TEXT DEFAULT NULL, -- 主机名
    hostIpAddress TEXT DEFAULT NULL, -- 主机IP地址
    
    -- 时间相关字段
    collectionTime TEXT NOT NULL, -- 数据采集时间
    jvmStartTime TEXT NOT NULL, -- JVM启动时间
    jvmUptimeMs INTEGER DEFAULT 0 NOT NULL, -- JVM运行时长（毫秒）
    
    -- 健康状态字段
    healthyFlag TEXT DEFAULT 'Y' NOT NULL CHECK(healthyFlag IN ('Y','N')), -- JVM整体健康标记(Y健康,N异常)
    healthGrade TEXT DEFAULT NULL, -- JVM健康等级(EXCELLENT/GOOD/FAIR/POOR)
    requiresAttentionFlag TEXT DEFAULT 'N' NOT NULL CHECK(requiresAttentionFlag IN ('Y','N')), -- 是否需要立即关注(Y是,N否)
    summaryText TEXT DEFAULT NULL, -- 监控摘要信息
    
    -- 系统属性（JSON格式）
    systemPropertiesJson TEXT DEFAULT NULL, -- JVM系统属性，JSON格式（可能包含大量系统属性）
    
    -- 通用字段
    addTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 创建时间
    addWho TEXT DEFAULT NULL, -- 创建人ID
    editTime TEXT DEFAULT (datetime('now','localtime')) NOT NULL, -- 最后修改时间
    editWho TEXT DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag TEXT DEFAULT NULL, -- 操作序列标识
    currentVersion INTEGER DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag TEXT DEFAULT 'Y' NOT NULL CHECK(activeFlag IN ('Y','N')), -- 活动状态标记(N非活动,Y活动)
    noteText TEXT DEFAULT NULL, -- 备注信息
    
    PRIMARY KEY (tenantId, serviceGroupId, jvmResourceId)
);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_APP ON HUB_MONITOR_JVM_RESOURCE(applicationName);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_TIME ON HUB_MONITOR_JVM_RESOURCE(collectionTime);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_HEALTH ON HUB_MONITOR_JVM_RESOURCE(healthyFlag, requiresAttentionFlag);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_HOST ON HUB_MONITOR_JVM_RESOURCE(hostIpAddress);
CREATE INDEX IF NOT EXISTS IDX_MONITOR_JVM_GROUP ON HUB_MONITOR_JVM_RESOURCE(serviceGroupId, groupName);