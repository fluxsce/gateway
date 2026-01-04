CREATE TABLE HUB_MONITOR_JVM_THR_STATE (
                                           threadStateId VARCHAR(32) NOT NULL COMMENT '线程状态记录ID，主键',
                                           tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                           jvmThreadId VARCHAR(32) NOT NULL COMMENT '关联的JVM线程记录ID',
                                           jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 线程状态分布
                                           newThreadCount INT NOT NULL DEFAULT 0 COMMENT 'NEW状态线程数',
                                           runnableThreadCount INT NOT NULL DEFAULT 0 COMMENT 'RUNNABLE状态线程数',
                                           blockedThreadCount INT NOT NULL DEFAULT 0 COMMENT 'BLOCKED状态线程数',
                                           waitingThreadCount INT NOT NULL DEFAULT 0 COMMENT 'WAITING状态线程数',
                                           timedWaitingThreadCount INT NOT NULL DEFAULT 0 COMMENT 'TIMED_WAITING状态线程数',
                                           terminatedThreadCount INT NOT NULL DEFAULT 0 COMMENT 'TERMINATED状态线程数',
                                           totalThreadCount INT NOT NULL DEFAULT 0 COMMENT '总线程数',

    -- 比例指标
                                           activeThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '活跃线程比例（百分比）',
                                           blockedThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '阻塞线程比例（百分比）',
                                           waitingThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00 COMMENT '等待状态线程比例（百分比）',

    -- 健康状态
                                           healthyFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '线程状态健康标记(Y健康,N异常)',
                                           healthGrade VARCHAR(20) DEFAULT NULL COMMENT '健康等级(EXCELLENT/GOOD/FAIR/POOR)',

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

                                           PRIMARY KEY (tenantId, threadStateId),
                                           KEY IDX_MONITOR_THRST_THR (jvmThreadId),
                                           KEY IDX_MONITOR_THRST_RES (jvmResourceId),
                                           KEY IDX_MONITOR_THRST_TIME (collectionTime),
                                           KEY IDX_MONITOR_THRST_BLOCK (blockedThreadCount)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM线程状态统计表';
