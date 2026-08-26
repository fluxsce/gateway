-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_JVM_CLASS.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_JVM_CLASS', N'U') IS NULL
CREATE TABLE HUB_MONITOR_JVM_CLASS (
                                       classLoadingId NVARCHAR(32) NOT NULL,
                                       tenantId NVARCHAR(32) NOT NULL,
                                       jvmResourceId NVARCHAR(100) NOT NULL,

    -- 类加载统计
                                       loadedClassCount INT NOT NULL DEFAULT 0,
                                       totalLoadedClassCount BIGINT NOT NULL DEFAULT 0,
                                       unloadedClassCount BIGINT NOT NULL DEFAULT 0,

    -- 比例指标
                                       classUnloadRatePercent DECIMAL(5,2) DEFAULT 0.00,
                                       classRetentionRatePercent DECIMAL(5,2) DEFAULT 0.00,

    -- 配置状态
                                       verboseClassLoading NVARCHAR(1) NOT NULL DEFAULT N'N',

    -- 性能指标
                                       loadingRatePerHour DECIMAL(10,2) DEFAULT 0.00,
                                       loadingEfficiency DECIMAL(5,2) DEFAULT 0.00,
                                       memoryEfficiency NVARCHAR(100) DEFAULT NULL,
                                       loaderHealth NVARCHAR(50) DEFAULT NULL,

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

                                       PRIMARY KEY (tenantId, classLoadingId),
                                       INDEX IDX_MONITOR_CLS_RES (jvmResourceId),
                                       INDEX IDX_MONITOR_CLS_TIME (collectionTime),
                                       INDEX IDX_MONITOR_CLS_HEALTH (healthyFlag, requiresAttentionFlag),
                                       INDEX IDX_MONITOR_CLS_COUNT (loadedClassCount)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN verboseClassLoading NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN memoryEfficiency NVARCHAR(100) NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN loaderHealth NVARCHAR(50) NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN healthyFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN healthGrade NVARCHAR(20) NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN requiresAttentionFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_CLASS ALTER COLUMN noteText NVARCHAR(500) NULL;
