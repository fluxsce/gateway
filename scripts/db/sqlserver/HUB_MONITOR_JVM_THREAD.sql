-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_JVM_THREAD.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_JVM_THREAD', N'U') IS NULL
CREATE TABLE HUB_MONITOR_JVM_THREAD (
                                        jvmThreadId NVARCHAR(32) NOT NULL,
                                        tenantId NVARCHAR(32) NOT NULL,
                                        jvmResourceId NVARCHAR(100) NOT NULL,

    -- 基础线程统计
                                        currentThreadCount INT NOT NULL DEFAULT 0,
                                        daemonThreadCount INT NOT NULL DEFAULT 0,
                                        userThreadCount INT NOT NULL DEFAULT 0,
                                        peakThreadCount INT NOT NULL DEFAULT 0,
                                        totalStartedThreadCount BIGINT NOT NULL DEFAULT 0,

    -- 性能指标
                                        threadGrowthRatePercent DECIMAL(5,2) DEFAULT 0.00,
                                        daemonThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00,

    -- 监控功能支持状态
                                        cpuTimeSupported NVARCHAR(1) NOT NULL DEFAULT N'N',
                                        cpuTimeEnabled NVARCHAR(1) NOT NULL DEFAULT N'N',
                                        memoryAllocSupported NVARCHAR(1) NOT NULL DEFAULT N'N',
                                        memoryAllocEnabled NVARCHAR(1) NOT NULL DEFAULT N'N',
                                        contentionSupported NVARCHAR(1) NOT NULL DEFAULT N'N',
                                        contentionEnabled NVARCHAR(1) NOT NULL DEFAULT N'N',

    -- 健康状态
                                        healthyFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                        healthGrade NVARCHAR(20) DEFAULT NULL,
                                        requiresAttentionFlag NVARCHAR(1) NOT NULL DEFAULT N'N',
                                        potentialIssuesJson NVARCHAR(MAX) DEFAULT NULL,

    -- 时间字段
                                        collectionTime DATETIME2 NOT NULL,

    -- 通用字段
                                        addTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                        addWho NVARCHAR(32) DEFAULT NULL,
                                        editTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                        editWho NVARCHAR(32) DEFAULT NULL,
                                        oprSeqFlag NVARCHAR(32) DEFAULT NULL,
                                        currentVersion INT NOT NULL DEFAULT 1,
                                        activeFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                        noteText NVARCHAR(500) DEFAULT NULL,

                                        PRIMARY KEY (tenantId, jvmThreadId),
                                        INDEX IDX_MONITOR_THR_RES (jvmResourceId),
                                        INDEX IDX_MONITOR_THR_TIME (collectionTime),
                                        INDEX IDX_MONITOR_THR_HEALTH (healthyFlag, requiresAttentionFlag),
                                        INDEX IDX_MONITOR_THR_COUNT (currentThreadCount)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN cpuTimeSupported NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN cpuTimeEnabled NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN memoryAllocSupported NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN memoryAllocEnabled NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN contentionSupported NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN contentionEnabled NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN healthyFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN healthGrade NVARCHAR(20) NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN requiresAttentionFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THREAD ALTER COLUMN noteText NVARCHAR(500) NULL;
