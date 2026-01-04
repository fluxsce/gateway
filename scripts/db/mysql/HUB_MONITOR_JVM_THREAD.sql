CREATE TABLE HUB_MONITOR_JVM_THREAD (
                                        jvmThreadId VARCHAR(32) NOT NULL COMMENT 'JVM线程记录ID，主键',
                                        tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                        jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 基础线程统计
                                        currentThreadCount INT NOT NULL DEFAULT 0 COMMENT '当前线程数',
                                        daemonThreadCount INT NOT NULL DEFAULT 0 COMMENT '守护线程数',
                                        userThreadCount INT NOT NULL DEFAULT 0 COMMENT '用户线程数',
                                        peakThreadCount INT NOT NULL DEFAULT 0 COMMENT '峰值线程数',
                                        totalStartedThreadCount BIGINT NOT NULL DEFAULT 0 COMMENT '总启动线程数',

    -- 性能指标
                                        threadGrowthRatePercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '线程增长率（百分比）',
                                        daemonThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '守护线程比例（百分比）',

    -- 监控功能支持状态
                                        cpuTimeSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT 'CPU时间监控是否支持(Y是,N否)',
                                        cpuTimeEnabled VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT 'CPU时间监控是否启用(Y是,N否)',
                                        memoryAllocSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '内存分配监控是否支持(Y是,N否)',
                                        memoryAllocEnabled VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '内存分配监控是否启用(Y是,N否)',
                                        contentionSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '争用监控是否支持(Y是,N否)',
                                        contentionEnabled VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '争用监控是否启用(Y是,N否)',

    -- 健康状态
                                        healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '线程健康标记(Y健康,N异常)',
                                        healthGrade VARCHAR(20) DEFAULT NULL COMMENT '线程健康等级(EXCELLENT/GOOD/FAIR/POOR)',
                                        requiresAttentionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即关注(Y是,N否)',
                                        potentialIssuesJson LONGTEXT DEFAULT NULL COMMENT '潜在问题列表，JSON格式',

    -- 时间字段
                                        collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 通用字段
                                        addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                        addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                        editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                        editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                        oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                        currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                        activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                        noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                        PRIMARY KEY (tenantId, jvmThreadId),
                                        KEY IDX_MONITOR_THR_RES (jvmResourceId),
                                        KEY IDX_MONITOR_THR_TIME (collectionTime),
                                        KEY IDX_MONITOR_THR_HEALTH (healthyFlag, requiresAttentionFlag),
                                        KEY IDX_MONITOR_THR_COUNT (currentThreadCount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM线程监控表';
