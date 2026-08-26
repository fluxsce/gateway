-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_JVM_MEM_POOL.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_JVM_MEM_POOL', N'U') IS NULL
CREATE TABLE HUB_MONITOR_JVM_MEM_POOL (
                                          memoryPoolId NVARCHAR(32) NOT NULL,
                                          tenantId NVARCHAR(32) NOT NULL,
                                          jvmResourceId NVARCHAR(100) NOT NULL,

    -- 内存池基本信息
                                          poolName NVARCHAR(100) NOT NULL,
                                          poolType NVARCHAR(20) NOT NULL,
                                          poolCategory NVARCHAR(50) DEFAULT NULL,

    -- 当前使用情况
                                          currentInitBytes BIGINT NOT NULL DEFAULT 0,
                                          currentUsedBytes BIGINT NOT NULL DEFAULT 0,
                                          currentCommittedBytes BIGINT NOT NULL DEFAULT 0,
                                          currentMaxBytes BIGINT NOT NULL DEFAULT -1,
                                          currentUsagePercent DECIMAL(5,2) NOT NULL DEFAULT 0.00,

    -- 峰值使用情况
                                          peakInitBytes BIGINT DEFAULT 0,
                                          peakUsedBytes BIGINT DEFAULT 0,
                                          peakCommittedBytes BIGINT DEFAULT 0,
                                          peakMaxBytes BIGINT DEFAULT -1,
                                          peakUsagePercent DECIMAL(5,2) DEFAULT 0.00,

    -- 阈值监控
                                          usageThresholdSupported NVARCHAR(1) NOT NULL DEFAULT N'N',
                                          usageThresholdBytes BIGINT DEFAULT 0,
                                          usageThresholdCount BIGINT DEFAULT 0,
                                          collectionUsageSupported NVARCHAR(1) NOT NULL DEFAULT N'N',

    -- 健康状态
                                          healthyFlag NVARCHAR(1) NOT NULL DEFAULT N'Y',

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

                                          PRIMARY KEY (tenantId, memoryPoolId),
                                          INDEX IDX_MONITOR_POOL_RES (jvmResourceId),
                                          INDEX IDX_MONITOR_POOL_NAME (poolName),
                                          INDEX IDX_MONITOR_POOL_TYPE (poolType),
                                          INDEX IDX_MONITOR_POOL_CAT (poolCategory),
                                          INDEX IDX_MONITOR_POOL_TIME (collectionTime)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN poolName NVARCHAR(100) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN poolType NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN poolCategory NVARCHAR(50) NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN usageThresholdSupported NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN collectionUsageSupported NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN healthyFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEM_POOL ALTER COLUMN noteText NVARCHAR(500) NULL;
