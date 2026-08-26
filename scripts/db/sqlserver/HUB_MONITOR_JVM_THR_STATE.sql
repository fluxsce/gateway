-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_JVM_THR_STATE.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_JVM_THR_STATE', N'U') IS NULL
CREATE TABLE HUB_MONITOR_JVM_THR_STATE (
                                           threadStateId NVARCHAR(32) NOT NULL,
                                           tenantId NVARCHAR(32) NOT NULL,
                                           jvmThreadId NVARCHAR(32) NOT NULL,
                                           jvmResourceId NVARCHAR(100) NOT NULL,

    -- 线程状态分布
                                           newThreadCount INT NOT NULL DEFAULT 0,
                                           runnableThreadCount INT NOT NULL DEFAULT 0,
                                           blockedThreadCount INT NOT NULL DEFAULT 0,
                                           waitingThreadCount INT NOT NULL DEFAULT 0,
                                           timedWaitingThreadCount INT NOT NULL DEFAULT 0,
                                           terminatedThreadCount INT NOT NULL DEFAULT 0,
                                           totalThreadCount INT NOT NULL DEFAULT 0,

    -- 比例指标
                                           activeThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00,
                                           blockedThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00,
                                           waitingThreadRatioPercent DECIMAL(5,2) DEFAULT 0.00,

    -- 健康状态
                                           healthyFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                           healthGrade NVARCHAR(20) DEFAULT NULL,

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

                                           PRIMARY KEY (tenantId, threadStateId),
                                           INDEX IDX_MONITOR_THRST_THR (jvmThreadId),
                                           INDEX IDX_MONITOR_THRST_RES (jvmResourceId),
                                           INDEX IDX_MONITOR_THRST_TIME (collectionTime),
                                           INDEX IDX_MONITOR_THRST_BLOCK (blockedThreadCount)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_JVM_THR_STATE ALTER COLUMN healthyFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THR_STATE ALTER COLUMN healthGrade NVARCHAR(20) NULL;
ALTER TABLE HUB_MONITOR_JVM_THR_STATE ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_THR_STATE ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_THR_STATE ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_THR_STATE ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_THR_STATE ALTER COLUMN noteText NVARCHAR(500) NULL;
