-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_JVM_MEMORY.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_JVM_MEMORY', N'U') IS NULL
CREATE TABLE HUB_MONITOR_JVM_MEMORY (
                                        jvmMemoryId NVARCHAR(32) NOT NULL,
                                        tenantId NVARCHAR(32) NOT NULL,
                                        jvmResourceId NVARCHAR(100) NOT NULL,

    -- 内存类型
                                        memoryType NVARCHAR(20) NOT NULL,

    -- 内存使用情况（字节）
                                        initMemoryBytes BIGINT NOT NULL DEFAULT 0,
                                        usedMemoryBytes BIGINT NOT NULL DEFAULT 0,
                                        committedMemoryBytes BIGINT NOT NULL DEFAULT 0,
                                        maxMemoryBytes BIGINT NOT NULL DEFAULT -1,

    -- 计算指标
                                        usagePercent DECIMAL(5,2) NOT NULL DEFAULT 0.00,
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

                                        PRIMARY KEY (tenantId, jvmMemoryId),
                                        INDEX IDX_MONITOR_MEM_RES (jvmResourceId),
                                        INDEX IDX_MONITOR_MEM_TYPE (memoryType),
                                        INDEX IDX_MONITOR_MEM_TIME (collectionTime),
                                        INDEX IDX_MONITOR_MEM_USAGE (usagePercent)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_JVM_MEMORY ALTER COLUMN memoryType NVARCHAR(20) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEMORY ALTER COLUMN healthyFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEMORY ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_MEMORY ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_MEMORY ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_MEMORY ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_MEMORY ALTER COLUMN noteText NVARCHAR(500) NULL;
