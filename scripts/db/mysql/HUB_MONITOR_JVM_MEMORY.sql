CREATE TABLE HUB_MONITOR_JVM_MEMORY (
                                        jvmMemoryId VARCHAR(32) NOT NULL COMMENT 'JVM内存记录ID，主键',
                                        tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                        jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 内存类型
                                        memoryType VARCHAR(20) NOT NULL COMMENT '内存类型(HEAP/NON_HEAP)',

    -- 内存使用情况（字节）
                                        initMemoryBytes BIGINT NOT NULL DEFAULT 0 COMMENT '初始内存大小（字节）',
                                        usedMemoryBytes BIGINT NOT NULL DEFAULT 0 COMMENT '已使用内存大小（字节）',
                                        committedMemoryBytes BIGINT NOT NULL DEFAULT 0 COMMENT '已提交内存大小（字节）',
                                        maxMemoryBytes BIGINT NOT NULL DEFAULT -1 COMMENT '最大内存大小（字节），-1表示无限制',

    -- 计算指标
                                        usagePercent DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '内存使用率（百分比）',
                                        healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '内存健康标记(Y健康,N异常)',

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

                                        PRIMARY KEY (tenantId, jvmMemoryId),
                                        KEY IDX_MONITOR_MEM_RES (jvmResourceId),
                                        KEY IDX_MONITOR_MEM_TYPE (memoryType),
                                        KEY IDX_MONITOR_MEM_TIME (collectionTime),
                                        KEY IDX_MONITOR_MEM_USAGE (usagePercent)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM内存监控表';
