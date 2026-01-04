CREATE TABLE HUB_MONITOR_JVM_GC (
                                    gcSnapshotId VARCHAR(32) NOT NULL COMMENT 'GC快照记录ID，主键',
                                    tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                    jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- GC累积统计（从JVM启动到当前采集时刻）
                                    collectionCount BIGINT NOT NULL DEFAULT 0 COMMENT 'GC总次数（累积，所有GC收集器汇总）',
                                    collectionTimeMs BIGINT NOT NULL DEFAULT 0 COMMENT 'GC总耗时（毫秒，累积，所有GC收集器汇总）',

    -- ===== jstat -gc 风格的内存区域数据（单位：KB） =====

    -- Survivor区
                                    s0c BIGINT DEFAULT 0 COMMENT 'Survivor 0 区容量（KB）',
                                    s1c BIGINT DEFAULT 0 COMMENT 'Survivor 1 区容量（KB）',
                                    s0u BIGINT DEFAULT 0 COMMENT 'Survivor 0 区使用量（KB）',
                                    s1u BIGINT DEFAULT 0 COMMENT 'Survivor 1 区使用量（KB）',

    -- Eden区
                                    ec BIGINT DEFAULT 0 COMMENT 'Eden 区容量（KB）',
                                    eu BIGINT DEFAULT 0 COMMENT 'Eden 区使用量（KB）',

    -- Old区
                                    oc BIGINT DEFAULT 0 COMMENT 'Old 区容量（KB）',
                                    ou BIGINT DEFAULT 0 COMMENT 'Old 区使用量（KB）',

    -- Metaspace
                                    mc BIGINT DEFAULT 0 COMMENT 'Metaspace 容量（KB）',
                                    mu BIGINT DEFAULT 0 COMMENT 'Metaspace 使用量（KB）',

    -- 压缩类空间
                                    ccsc BIGINT DEFAULT 0 COMMENT '压缩类空间容量（KB）',
                                    ccsu BIGINT DEFAULT 0 COMMENT '压缩类空间使用量（KB）',

    -- GC统计（jstat -gc 格式）
                                    ygc BIGINT DEFAULT 0 COMMENT '年轻代GC次数',
                                    ygct DECIMAL(10,3) DEFAULT 0.000 COMMENT '年轻代GC总时间（秒）',
                                    fgc BIGINT DEFAULT 0 COMMENT 'Full GC次数',
                                    fgct DECIMAL(10,3) DEFAULT 0.000 COMMENT 'Full GC总时间（秒）',
                                    gct DECIMAL(10,3) DEFAULT 0.000 COMMENT '总GC时间（秒）',

    -- 时间戳信息
                                    collectionTime DATETIME NOT NULL COMMENT '数据采集时间戳',

    -- 通用字段
                                    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                    addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                    editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                    oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                    currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                    activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                    noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                    PRIMARY KEY (tenantId, gcSnapshotId),
                                    KEY IDX_MONITOR_GC_RES (jvmResourceId),
                                    KEY IDX_MONITOR_GC_TIME (collectionTime),
                                    KEY IDX_MONITOR_GC_RES_TIME (jvmResourceId, collectionTime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM GC快照表（jstat -gc风格，每次采集一条汇总记录）';
