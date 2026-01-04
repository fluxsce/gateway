CREATE TABLE HUB_MONITOR_JVM_MEM_POOL (
                                          memoryPoolId VARCHAR(32) NOT NULL COMMENT '内存池记录ID，主键',
                                          tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                          jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 内存池基本信息
                                          poolName VARCHAR(100) NOT NULL COMMENT '内存池名称',
                                          poolType VARCHAR(20) NOT NULL COMMENT '内存池类型(HEAP/NON_HEAP)',
                                          poolCategory VARCHAR(50) DEFAULT NULL COMMENT '内存池分类（年轻代/老年代/元数据空间/代码缓存/其他）',

    -- 当前使用情况
                                          currentInitBytes BIGINT NOT NULL DEFAULT 0 COMMENT '当前初始内存（字节）',
                                          currentUsedBytes BIGINT NOT NULL DEFAULT 0 COMMENT '当前已使用内存（字节）',
                                          currentCommittedBytes BIGINT NOT NULL DEFAULT 0 COMMENT '当前已提交内存（字节）',
                                          currentMaxBytes BIGINT NOT NULL DEFAULT -1 COMMENT '当前最大内存（字节）',
                                          currentUsagePercent DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '当前使用率（百分比）',

    -- 峰值使用情况
                                          peakInitBytes BIGINT DEFAULT 0 COMMENT '峰值初始内存（字节）',
                                          peakUsedBytes BIGINT DEFAULT 0 COMMENT '峰值已使用内存（字节）',
                                          peakCommittedBytes BIGINT DEFAULT 0 COMMENT '峰值已提交内存（字节）',
                                          peakMaxBytes BIGINT DEFAULT -1 COMMENT '峰值最大内存（字节）',
                                          peakUsagePercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '峰值使用率（百分比）',

    -- 阈值监控
                                          usageThresholdSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否支持使用阈值监控(Y是,N否)',
                                          usageThresholdBytes BIGINT DEFAULT 0 COMMENT '使用阈值（字节）',
                                          usageThresholdCount BIGINT DEFAULT 0 COMMENT '使用阈值超越次数',
                                          collectionUsageSupported VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否支持收集使用量监控(Y是,N否)',

    -- 健康状态
                                          healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '内存池健康标记(Y健康,N异常)',

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

                                          PRIMARY KEY (tenantId, memoryPoolId),
                                          KEY IDX_MONITOR_POOL_RES (jvmResourceId),
                                          KEY IDX_MONITOR_POOL_NAME (poolName),
                                          KEY IDX_MONITOR_POOL_TYPE (poolType),
                                          KEY IDX_MONITOR_POOL_CAT (poolCategory),
                                          KEY IDX_MONITOR_POOL_TIME (collectionTime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM内存池监控表';
