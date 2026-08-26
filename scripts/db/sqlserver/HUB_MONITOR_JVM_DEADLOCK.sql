-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_JVM_DEADLOCK.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_JVM_DEADLOCK', N'U') IS NULL
CREATE TABLE HUB_MONITOR_JVM_DEADLOCK (
                                          deadlockId NVARCHAR(32) NOT NULL,
                                          tenantId NVARCHAR(32) NOT NULL,
                                          jvmThreadId NVARCHAR(32) NOT NULL,
                                          jvmResourceId NVARCHAR(100) NOT NULL,

    -- 死锁基本信息
                                          hasDeadlockFlag NVARCHAR(1) NOT NULL DEFAULT N'N',
                                          deadlockThreadCount INT NOT NULL DEFAULT 0,
                                          deadlockThreadIds NVARCHAR(MAX) DEFAULT NULL,
                                          deadlockThreadNames NVARCHAR(MAX) DEFAULT NULL,

    -- 死锁严重程度
                                          severityLevel NVARCHAR(20) DEFAULT NULL,
                                          severityDescription NVARCHAR(200) DEFAULT NULL,
                                          affectedThreadGroups INT DEFAULT 0,

    -- 时间信息
                                          detectionTime DATETIME2 DEFAULT NULL,
                                          deadlockDurationMs BIGINT DEFAULT 0,
                                          collectionTime DATETIME2 NOT NULL,

    -- 诊断信息
                                          descriptionText NVARCHAR(500) DEFAULT NULL,
                                          recommendedAction NVARCHAR(500) DEFAULT NULL,
                                          alertLevel NVARCHAR(20) DEFAULT NULL,
                                          requiresActionFlag NVARCHAR(1) NOT NULL DEFAULT N'N',

    -- 通用字段
                                          addTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                          addWho NVARCHAR(32) DEFAULT NULL,
                                          editTime DATETIME2 NOT NULL DEFAULT GETDATE(),
                                          editWho NVARCHAR(32) DEFAULT NULL,
                                          oprSeqFlag NVARCHAR(32) DEFAULT NULL,
                                          currentVersion INT NOT NULL DEFAULT 1,
                                          activeFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',
                                          noteText NVARCHAR(500) DEFAULT NULL,

                                          PRIMARY KEY (tenantId, deadlockId),
                                          INDEX IDX_MONITOR_DL_THR (jvmThreadId),
                                          INDEX IDX_MONITOR_DL_RES (jvmResourceId),
                                          INDEX IDX_MONITOR_DL_TIME (collectionTime),
                                          INDEX IDX_MONITOR_DL_FLAG (hasDeadlockFlag),
                                          INDEX IDX_MONITOR_DL_SEV (severityLevel),
                                          INDEX IDX_MONITOR_DL_ALERT (alertLevel)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN hasDeadlockFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN severityLevel NVARCHAR(20) NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN severityDescription NVARCHAR(200) NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN descriptionText NVARCHAR(500) NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN recommendedAction NVARCHAR(500) NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN alertLevel NVARCHAR(20) NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN requiresActionFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_DEADLOCK ALTER COLUMN noteText NVARCHAR(500) NULL;
