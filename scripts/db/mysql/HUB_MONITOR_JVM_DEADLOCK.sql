CREATE TABLE HUB_MONITOR_JVM_DEADLOCK (
                                          deadlockId VARCHAR(32) NOT NULL COMMENT '死锁记录ID，主键',
                                          tenantId VARCHAR(32) NOT NULL COMMENT '租户ID',
                                          jvmThreadId VARCHAR(32) NOT NULL COMMENT '关联的JVM线程记录ID',
                                          jvmResourceId VARCHAR(100) NOT NULL COMMENT '关联的JVM资源ID',

    -- 死锁基本信息
                                          hasDeadlockFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否检测到死锁(Y是,N否)',
                                          deadlockThreadCount INT NOT NULL DEFAULT 0 COMMENT '死锁线程数量',
                                          deadlockThreadIds TEXT DEFAULT NULL COMMENT '死锁线程ID列表，逗号分隔',
                                          deadlockThreadNames TEXT DEFAULT NULL COMMENT '死锁线程名称列表，逗号分隔',

    -- 死锁严重程度
                                          severityLevel VARCHAR(20) DEFAULT NULL COMMENT '严重程度(LOW/MEDIUM/HIGH/CRITICAL)',
                                          severityDescription VARCHAR(200) DEFAULT NULL COMMENT '严重程度描述',
                                          affectedThreadGroups INT DEFAULT 0 COMMENT '影响的线程组数量',

    -- 时间信息
                                          detectionTime DATETIME DEFAULT NULL COMMENT '死锁检测时间',
                                          deadlockDurationMs BIGINT DEFAULT 0 COMMENT '死锁持续时间（毫秒）',
                                          collectionTime DATETIME NOT NULL COMMENT '数据采集时间',

    -- 诊断信息
                                          descriptionText VARCHAR(500) DEFAULT NULL COMMENT '死锁描述信息',
                                          recommendedAction VARCHAR(500) DEFAULT NULL COMMENT '建议的解决方案',
                                          alertLevel VARCHAR(20) DEFAULT NULL COMMENT '告警级别(INFO/WARNING/ERROR/CRITICAL/EMERGENCY)',
                                          requiresActionFlag VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否需要立即处理(Y是,N否)',

    -- 通用字段
                                          addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                          addWho VARCHAR(32) DEFAULT NULL COMMENT '创建人ID',
                                          editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
                                          editWho VARCHAR(32) DEFAULT NULL COMMENT '最后修改人ID',
                                          oprSeqFlag VARCHAR(32) DEFAULT NULL COMMENT '操作序列标识',
                                          currentVersion INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
                                          activeFlag VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
                                          noteText VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

                                          PRIMARY KEY (tenantId, deadlockId),
                                          KEY IDX_MONITOR_DL_THR (jvmThreadId),
                                          KEY IDX_MONITOR_DL_RES (jvmResourceId),
                                          KEY IDX_MONITOR_DL_TIME (collectionTime),
                                          KEY IDX_MONITOR_DL_FLAG (hasDeadlockFlag),
                                          KEY IDX_MONITOR_DL_SEV (severityLevel),
                                          KEY IDX_MONITOR_DL_ALERT (alertLevel)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='JVM死锁检测信息表';
