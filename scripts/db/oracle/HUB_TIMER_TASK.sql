CREATE TABLE HUB_TIMER_TASK (
                                taskId                  VARCHAR2(32) NOT NULL, -- 任务ID，主键
                                tenantId                VARCHAR2(32) NOT NULL, -- 租户ID

                                taskName                VARCHAR2(200) NOT NULL, -- 任务名称
                                taskDescription         VARCHAR2(500), -- 任务描述
                                taskPriority            NUMBER(10) DEFAULT 1 NOT NULL, -- 任务优先级(1低,2普通,3高)
                                schedulerId             VARCHAR2(32), -- 关联的调度器ID
                                schedulerName           VARCHAR2(100), -- 调度器名称（冗余字段）

                                scheduleType            NUMBER(10) NOT NULL, -- 调度类型(1一次性,2固定间隔,3Cron,4延迟执行,5实时执行)
                                cronExpression          VARCHAR2(100), -- Cron表达式（scheduleType=3时必填）
                                intervalSeconds         NUMBER(20), -- 执行间隔秒数（scheduleType=2时必填）
                                delaySeconds            NUMBER(20), -- 延迟秒数（scheduleType=4时必填）
                                startTime               DATE, -- 任务开始时间
                                endTime                 DATE, -- 任务结束时间

                                maxRetries              NUMBER(10) DEFAULT 0 NOT NULL, -- 最大重试次数
                                retryIntervalSeconds    NUMBER(20) DEFAULT 60 NOT NULL, -- 重试间隔秒数
                                timeoutSeconds          NUMBER(20) DEFAULT 1800 NOT NULL, -- 执行超时时间秒数
                                taskParams              CLOB, -- 任务参数，JSON格式存储

    -- 新增字段：任务执行器配置
                                executorType            VARCHAR2(50), -- 执行器类型(BUILTIN内置,SFTP,SSH,DATABASE,HTTP等)
                                toolConfigId            VARCHAR2(32), -- 工具配置ID（如SFTP配置ID、数据库配置ID等）
                                toolConfigName          VARCHAR2(100), -- 工具配置名称（冗余字段）
                                operationType           VARCHAR2(100), -- 执行操作类型（如文件上传、下载、SQL执行、接口调用等）
                                operationConfig         CLOB, -- 操作参数配置，JSON格式存储具体操作的参数

                                taskStatus              NUMBER(10) DEFAULT 1 NOT NULL, -- 任务状态(1待执行,2运行中,3已完成,4失败,5取消)
                                nextRunTime             DATE, -- 下次执行时间
                                lastRunTime             DATE, -- 上次执行时间
                                runCount                NUMBER(20) DEFAULT 0 NOT NULL, -- 执行总次数
                                successCount            NUMBER(20) DEFAULT 0 NOT NULL, -- 成功次数
                                failureCount            NUMBER(20) DEFAULT 0 NOT NULL, -- 失败次数

                                lastExecutionId         VARCHAR2(32), -- 最后执行ID
                                lastExecutionStartTime  DATE, -- 最后执行开始时间
                                lastExecutionEndTime    DATE, -- 最后执行结束时间
                                lastExecutionDurationMs NUMBER(20), -- 最后执行耗时毫秒数
                                lastExecutionStatus     NUMBER(10), -- 最后执行状态
                                lastResultSuccess       VARCHAR2(1), -- 最后执行是否成功(N失败,Y成功)
                                lastErrorMessage        CLOB, -- 最后错误信息
                                lastRetryCount          NUMBER(10), -- 最后重试次数

                                addTime                 DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                addWho                  VARCHAR2(32) NOT NULL, -- 创建人ID
                                editTime                DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                editWho                 VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                oprSeqFlag              VARCHAR2(32) NOT NULL, -- 操作序列标识
                                currentVersion          NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                activeFlag              VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N/Y)
                                noteText                VARCHAR2(500), -- 备注信息
                                extProperty             CLOB, -- 扩展属性，JSON格式

                                reserved1               VARCHAR2(500), -- 预留字段1
                                reserved2               VARCHAR2(500), -- 预留字段2
                                reserved3               VARCHAR2(500), -- 预留字段3
                                reserved4               VARCHAR2(500), -- 预留字段4
                                reserved5               VARCHAR2(500), -- 预留字段5
                                reserved6               VARCHAR2(500), -- 预留字段6
                                reserved7               VARCHAR2(500), -- 预留字段7
                                reserved8               VARCHAR2(500), -- 预留字段8
                                reserved9               VARCHAR2(500), -- 预留字段9
                                reserved10              VARCHAR2(500), -- 预留字段10

                                CONSTRAINT PK_TIMER_TASK PRIMARY KEY (tenantId, taskId)
);
CREATE INDEX IDX_TIMER_TASK_NAME ON HUB_TIMER_TASK(taskName);
CREATE INDEX IDX_TIMER_TASK_SCHED ON HUB_TIMER_TASK(schedulerId);
CREATE INDEX IDX_TIMER_TASK_TYPE ON HUB_TIMER_TASK(scheduleType);
CREATE INDEX IDX_TIMER_TASK_STATUS ON HUB_TIMER_TASK(taskStatus);
CREATE INDEX IDX_TIMER_TASK_ACTIVE ON HUB_TIMER_TASK(activeFlag);

CREATE INDEX IDX_TIMER_TASK_EXEC ON HUB_TIMER_TASK(executorType);
CREATE INDEX IDX_TIMER_TASK_TOOL ON HUB_TIMER_TASK(toolConfigId);
CREATE INDEX IDX_TIMER_TASK_OP ON HUB_TIMER_TASK(operationType);
COMMENT ON TABLE HUB_TIMER_TASK IS '定时任务表 - 合并任务配置、运行时信息和最后执行结果';