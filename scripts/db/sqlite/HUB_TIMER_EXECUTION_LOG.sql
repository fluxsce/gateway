
-- 17. 任务执行日志表
CREATE TABLE IF NOT EXISTS HUB_TIMER_EXECUTION_LOG (
    executionId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    taskId TEXT NOT NULL,
    taskName TEXT,
    schedulerId TEXT,
    executionStartTime DATETIME NOT NULL,
    executionEndTime DATETIME,
    executionDurationMs INTEGER,
    executionStatus INTEGER NOT NULL,
    resultSuccess TEXT NOT NULL DEFAULT 'N',
    errorMessage TEXT,
    errorStackTrace TEXT,
    retryCount INTEGER NOT NULL DEFAULT 0,
    maxRetryCount INTEGER NOT NULL DEFAULT 0,
    executionParams TEXT,
    executionResult TEXT,
    executorServerName TEXT,
    executorServerIp TEXT,
    logLevel TEXT,
    logMessage TEXT,
    logTimestamp DATETIME,
    executionPhase TEXT,
    threadName TEXT,
    className TEXT,
    methodName TEXT,
    exceptionClass TEXT,
    exceptionMessage TEXT,
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,
    extProperty TEXT,
    reserved1 TEXT,
    reserved2 TEXT,
    reserved3 TEXT,
    reserved4 TEXT,
    reserved5 TEXT,
    reserved6 TEXT,
    reserved7 TEXT,
    reserved8 TEXT,
    reserved9 TEXT,
    reserved10 TEXT,
    PRIMARY KEY (tenantId, executionId)
);
CREATE INDEX IDX_TIMER_LOG_TASK ON HUB_TIMER_EXECUTION_LOG(taskId);
CREATE INDEX IDX_TIMER_LOG_NAME ON HUB_TIMER_EXECUTION_LOG(taskName);
CREATE INDEX IDX_TIMER_LOG_SCHED ON HUB_TIMER_EXECUTION_LOG(schedulerId);
CREATE INDEX IDX_TIMER_LOG_START ON HUB_TIMER_EXECUTION_LOG(executionStartTime);
CREATE INDEX IDX_TIMER_LOG_STATUS ON HUB_TIMER_EXECUTION_LOG(executionStatus);
CREATE INDEX IDX_TIMER_LOG_SUCCESS ON HUB_TIMER_EXECUTION_LOG(resultSuccess);
CREATE INDEX IDX_TIMER_LOG_LEVEL ON HUB_TIMER_EXECUTION_LOG(logLevel);
CREATE INDEX IDX_TIMER_LOG_TIME ON HUB_TIMER_EXECUTION_LOG(logTimestamp);