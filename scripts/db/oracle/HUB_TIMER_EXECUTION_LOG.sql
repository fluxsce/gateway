CREATE TABLE HUB_TIMER_EXECUTION_LOG (
                                         executionId             VARCHAR2(32) NOT NULL, -- 执行ID，主键
                                         tenantId                VARCHAR2(32) NOT NULL, -- 租户ID
                                         taskId                  VARCHAR2(32) NOT NULL, -- 关联任务ID

                                         taskName                VARCHAR2(200), -- 任务名称（冗余）
                                         schedulerId             VARCHAR2(32), -- 调度器ID（冗余）

                                         executionStartTime      DATE NOT NULL, -- 执行开始时间
                                         executionEndTime        DATE, -- 执行结束时间
                                         executionDurationMs     NUMBER(20), -- 执行耗时毫秒数
                                         executionStatus         NUMBER(10) NOT NULL, -- 执行状态(1待执行,2运行中,3已完成,4失败,5取消)
                                         resultSuccess           VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否成功(N失败,Y成功)

                                         errorMessage              CLOB, -- 错误信息
                                         errorStackTrace           CLOB, -- 错误堆栈信息

                                         retryCount              NUMBER(10) DEFAULT 0 NOT NULL, -- 重试次数
                                         maxRetryCount           NUMBER(10) DEFAULT 0 NOT NULL, -- 最大重试次数

                                         executionParams         CLOB, -- 执行参数，JSON格式
                                         executionResult         CLOB, -- 执行结果，JSON格式

                                         executorServerName      VARCHAR2(100), -- 执行服务器名称
                                         executorServerIp        VARCHAR2(50), -- 执行服务器IP地址

                                         logLevel                VARCHAR2(10), -- 日志级别(DEBUG,INFO,WARN,ERROR)
                                         logMessage              CLOB, -- 日志消息内容
                                         logTimestamp            DATE, -- 日志时间戳

                                         executionPhase          VARCHAR2(50), -- 执行阶段(BEFORE_EXECUTE,EXECUTING,AFTER_EXECUTE,RETRY)
                                         threadName              VARCHAR2(100), -- 执行线程名称
                                         className               VARCHAR2(200), -- 执行类名
                                         methodName              VARCHAR2(100), -- 执行方法名

                                         exceptionClass          VARCHAR2(200), -- 异常类名
                                         exceptionMessage        CLOB, -- 异常消息

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

                                         CONSTRAINT PK_TIMER_EXECUTION_LOG PRIMARY KEY (tenantId, executionId)
);
CREATE INDEX IDX_TIMER_LOG_TASK ON HUB_TIMER_EXECUTION_LOG(taskId);
CREATE INDEX IDX_TIMER_LOG_NAME ON HUB_TIMER_EXECUTION_LOG(taskName);
CREATE INDEX IDX_TIMER_LOG_SCHED ON HUB_TIMER_EXECUTION_LOG(schedulerId);
CREATE INDEX IDX_TIMER_LOG_START ON HUB_TIMER_EXECUTION_LOG(executionStartTime);
CREATE INDEX IDX_TIMER_LOG_STATUS ON HUB_TIMER_EXECUTION_LOG(executionStatus);
CREATE INDEX IDX_TIMER_LOG_SUCCESS ON HUB_TIMER_EXECUTION_LOG(resultSuccess);
CREATE INDEX IDX_TIMER_LOG_LEVEL ON HUB_TIMER_EXECUTION_LOG(logLevel);
CREATE INDEX IDX_TIMER_LOG_TIME ON HUB_TIMER_EXECUTION_LOG(logTimestamp);
COMMENT ON TABLE HUB_TIMER_EXECUTION_LOG IS '任务执行日志表 - 合并执行记录和日志信息';

