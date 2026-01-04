CREATE TABLE HUB_MONITOR_JVM_CLASS (
                                       classLoadingId VARCHAR(32) NOT NULL COMMENT '类加载记录ID，主键',
                                       tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                       jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 类加载统计
                                       loadedClassCount INT NOT NULL DEFAULT 0 COMMENT '当前已加载类数量',
                                       totalLoadedClassCount BIGINT NOT NULL DEFAULT 0 COMMENT '总加载类数量',
                                       unloadedClassCount BIGINT NOT NULL DEFAULT 0 COMMENT '已卸载类数量',

    -- 比例指标
                                       classUnloadRatePercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '类卸载率（百分比）',
                                       classRetentionRatePercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '类保留率（百分比）',

    -- 配置状态
                                       verboseClassLoading VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用详细类加载输出(Y是,N否)',

    -- 性能指标
                                       loadingRatePerHour DECIMAL(10,2) DEFAULT 0.00 COMMENT '每小时平均类加载数量',
                                       loadingEfficiency DECIMAL(5,2) DEFAULT 0.00 COMMENT '类加载效率',
                                       memoryEfficiency VARCHAR(100) DEFAULT NULL COMMENT '内存使用效率评估',
                                       loaderHealth VARCHAR(50) DEFAULT NULL COMMENT '类加载器健康状况',

    -- 健康状态
                                       healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '类加载健康标记(Y健康,N异常)',
                                       healthGrade VARCHAR(20) DEFAULT NULL COMMENT '健康等级(EXCELLENT/GOOD/FAIR/POOR)',
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

                                       PRIMARY KEY (tenantId, classLoadingId),
                                       KEY IDX_MONITOR_CLS_RES (jvmResourceId),
                                       KEY IDX_MONITOR_CLS_TIME (collectionTime),
                                       KEY IDX_MONITOR_CLS_HEALTH (healthyFlag, requiresAttentionFlag),
                                       KEY IDX_MONITOR_CLS_COUNT (loadedClassCount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM类加载监控表';

