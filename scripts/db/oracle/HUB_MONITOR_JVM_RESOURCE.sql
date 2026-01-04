CREATE TABLE HUB_MONITOR_JVM_RESOURCE (
    jvmResourceId VARCHAR2(100) NOT NULL, -- JVM资源记录ID（由应用端生成的唯一标识），主键
    tenantId VARCHAR2(32) NOT NULL, -- 租户ID
    serviceGroupId VARCHAR2(32) NOT NULL, -- 服务分组ID，主键
    
    -- 应用标识信息
    applicationName VARCHAR2(100) NOT NULL, -- 应用名称
    groupName VARCHAR2(100) NOT NULL, -- 分组名称
    hostName VARCHAR2(100) DEFAULT NULL, -- 主机名
    hostIpAddress VARCHAR2(50) DEFAULT NULL, -- 主机IP地址
    
    -- 时间相关字段
    collectionTime DATE NOT NULL, -- 数据采集时间
    jvmStartTime DATE NOT NULL, -- JVM启动时间
    jvmUptimeMs NUMBER(19,0) DEFAULT 0 NOT NULL, -- JVM运行时长（毫秒）
    
    -- 健康状态字段
    healthyFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- JVM整体健康标记(Y健康,N异常)
    healthGrade VARCHAR2(20) DEFAULT NULL, -- JVM健康等级(EXCELLENT/GOOD/FAIR/POOR)
    requiresAttentionFlag VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否需要立即关注(Y是,N否)
    summaryText VARCHAR2(500) DEFAULT NULL, -- 监控摘要信息
    
    -- 系统属性（JSON格式）
    systemPropertiesJson CLOB DEFAULT NULL, -- JVM系统属性，JSON格式（可能包含大量系统属性）
    
    -- 通用字段
    addTime DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
    addWho VARCHAR2(32) DEFAULT NULL, -- 创建人ID
    editTime DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
    editWho VARCHAR2(32) DEFAULT NULL, -- 最后修改人ID
    oprSeqFlag VARCHAR2(32) DEFAULT NULL, -- 操作序列标识
    currentVersion NUMBER(10,0) DEFAULT 1 NOT NULL, -- 当前版本号
    activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
    noteText VARCHAR2(500) DEFAULT NULL, -- 备注信息
    
    CONSTRAINT PK_MONITOR_JVM_RES PRIMARY KEY (tenantId, serviceGroupId, jvmResourceId)
);

CREATE INDEX IDX_MONITOR_JVM_APP ON HUB_MONITOR_JVM_RESOURCE(applicationName);
CREATE INDEX IDX_MONITOR_JVM_TIME ON HUB_MONITOR_JVM_RESOURCE(collectionTime);
CREATE INDEX IDX_MONITOR_JVM_HEALTH ON HUB_MONITOR_JVM_RESOURCE(healthyFlag, requiresAttentionFlag);
CREATE INDEX IDX_MONITOR_JVM_HOST ON HUB_MONITOR_JVM_RESOURCE(hostIpAddress);
CREATE INDEX IDX_MONITOR_JVM_GROUP ON HUB_MONITOR_JVM_RESOURCE(serviceGroupId, groupName);

COMMENT ON TABLE HUB_MONITOR_JVM_RESOURCE IS 'JVM资源监控主表';
