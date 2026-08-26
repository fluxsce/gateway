-- SQL Server 方言，由 scripts/db/mysql/HUB_MONITOR_JVM_GC.sql 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。
IF OBJECT_ID(N'dbo.HUB_MONITOR_JVM_GC', N'U') IS NULL
CREATE TABLE HUB_MONITOR_JVM_GC (
                                    gcSnapshotId NVARCHAR(32) NOT NULL,
                                    tenantId NVARCHAR(32) NOT NULL,
                                    jvmResourceId NVARCHAR(100) NOT NULL,

    -- GC累积统计（从JVM启动到当前采集时刻）
                                    collectionCount BIGINT NOT NULL DEFAULT 0,
                                    collectionTimeMs BIGINT NOT NULL DEFAULT 0,

    -- ===== jstat -gc 风格的内存区域数据（单位：KB） =====

    -- Survivor区
                                    s0c BIGINT DEFAULT 0,
                                    s1c BIGINT DEFAULT 0,
                                    s0u BIGINT DEFAULT 0,
                                    s1u BIGINT DEFAULT 0,

    -- Eden区
                                    ec BIGINT DEFAULT 0,
                                    eu BIGINT DEFAULT 0,

    -- Old区
                                    oc BIGINT DEFAULT 0,
                                    ou BIGINT DEFAULT 0,

    -- Metaspace
                                    mc BIGINT DEFAULT 0,
                                    mu BIGINT DEFAULT 0,

    -- 压缩类空间
                                    ccsc BIGINT DEFAULT 0,
                                    ccsu BIGINT DEFAULT 0,

    -- GC统计（jstat -gc 格式）
                                    ygc BIGINT DEFAULT 0,
                                    ygct DECIMAL(10,3) DEFAULT 0.000,
                                    fgc BIGINT DEFAULT 0,
                                    fgct DECIMAL(10,3) DEFAULT 0.000,
                                    gct DECIMAL(10,3) DEFAULT 0.000,

    -- 时间戳信息
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

                                    PRIMARY KEY (tenantId, gcSnapshotId),
                                    INDEX IDX_MONITOR_GC_RES (jvmResourceId),
                                    INDEX IDX_MONITOR_GC_TIME (collectionTime),
                                    INDEX IDX_MONITOR_GC_RES_TIME (jvmResourceId, collectionTime)
);
-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。
ALTER TABLE HUB_MONITOR_JVM_GC ALTER COLUMN addWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_GC ALTER COLUMN editWho NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_GC ALTER COLUMN oprSeqFlag NVARCHAR(32) NULL;
ALTER TABLE HUB_MONITOR_JVM_GC ALTER COLUMN activeFlag NVARCHAR(1) NOT NULL;
ALTER TABLE HUB_MONITOR_JVM_GC ALTER COLUMN noteText NVARCHAR(500) NULL;
