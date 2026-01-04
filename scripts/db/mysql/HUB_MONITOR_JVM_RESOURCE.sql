CREATE TABLE HUB_MONITOR_JVM_RESOURCE (
                                          jvmResourceId VARCHAR(100) NOT NULL COMMENT 'JVM资源记录ID（由应用端生成的唯一标识），主键',
                                          tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                          serviceGroupId VARCHAR(32) NOT NULL COMMENT '服务分组ID，主键',

    -- 应用标识信息
                                          applicationName VARCHAR(100) NOT NULL COMMENT '应用名称',
                                          groupName VARCHAR(100) NOT NULL COMMENT '分组名称',
                                          hostName VARCHAR(100) DEFAULT NULL COMMENT '主机名',
                                          hostIpAddress VARCHAR(50) DEFAULT NULL COMMENT '主机IP地址',

    -- 时间相关字段
                                          collectionTime DATETIME NOT NULL COMMENT '数据采集时间',
                                          jvmStartTime DATETIME NOT NULL COMMENT 'JVM启动时间',
                                          jvmUptimeMs BIGINT NOT NULL DEFAULT 0 COMMENT 'JVM运行时长（毫秒）',

    -- 健康状态字段
                                          healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT 'JVM整体健康标记(Y健康,N异常)',
                                          healthGrade VARCHAR(20) DEFAULT NULL COMMENT 'JVM健康等级(EXCELLENT/GOOD/FAIR/POOR)',
                                          requiresAttentionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即关注(Y是,N否)',
                                          summaryText VARCHAR(500) DEFAULT NULL COMMENT '监控摘要信息',

    -- 系统属性（JSON格式）
                                          systemPropertiesJson LONGTEXT DEFAULT NULL COMMENT 'JVM系统属性，JSON格式（可能包含大量系统属性）',

    -- 通用字段
                                          addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                          addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                          editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                          editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                          oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                          currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                          activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                          noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                          PRIMARY KEY (tenantId, serviceGroupId, jvmResourceId),
                                          KEY IDX_MONITOR_JVM_APP (applicationName),
                                          KEY IDX_MONITOR_JVM_TIME (collectionTime),
                                          KEY IDX_MONITOR_JVM_HEALTH (healthyFlag, requiresAttentionFlag),
                                          KEY IDX_MONITOR_JVM_HOST (hostIpAddress),
                                          KEY IDX_MONITOR_JVM_GROUP (serviceGroupId, groupName)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM资源监控主表';
