
-- 16. 定时任务表
CREATE TABLE IF NOT EXISTS HUB_TIMER_TASK (
    taskId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    taskName TEXT NOT NULL,
    taskDescription TEXT,
    taskPriority INTEGER NOT NULL DEFAULT 1,
    schedulerId TEXT,
    schedulerName TEXT,
    scheduleType INTEGER NOT NULL,
    cronExpression TEXT,
    intervalSeconds INTEGER,
    delaySeconds INTEGER,
    startTime DATETIME,
    endTime DATETIME,
    maxRetries INTEGER NOT NULL DEFAULT 0,
    retryIntervalSeconds INTEGER NOT NULL DEFAULT 60,
    timeoutSeconds INTEGER NOT NULL DEFAULT 1800,
    taskParams TEXT,
    executorType TEXT,
    toolConfigId TEXT,
    toolConfigName TEXT,
    operationType TEXT,
    operationConfig TEXT,
    taskStatus INTEGER NOT NULL DEFAULT 1,
    nextRunTime DATETIME,
    lastRunTime DATETIME,
    runCount INTEGER NOT NULL DEFAULT 0,
    successCount INTEGER NOT NULL DEFAULT 0,
    failureCount INTEGER NOT NULL DEFAULT 0,
    lastExecutionId TEXT,
    lastExecutionStartTime DATETIME,
    lastExecutionEndTime DATETIME,
    lastExecutionDurationMs INTEGER,
    lastExecutionStatus INTEGER,
    lastResultSuccess TEXT,
    lastErrorMessage TEXT,
    lastRetryCount INTEGER,
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
    PRIMARY KEY (tenantId, taskId)
);
CREATE INDEX IDX_TIMER_TASK_NAME ON HUB_TIMER_TASK(taskName);
CREATE INDEX IDX_TIMER_TASK_SCHED ON HUB_TIMER_TASK(schedulerId);
CREATE INDEX IDX_TIMER_TASK_TYPE ON HUB_TIMER_TASK(scheduleType);
CREATE INDEX IDX_TIMER_TASK_STATUS ON HUB_TIMER_TASK(taskStatus);
CREATE INDEX IDX_TIMER_TASK_NEXT ON HUB_TIMER_TASK(nextRunTime);
CREATE INDEX IDX_TIMER_TASK_LAST ON HUB_TIMER_TASK(lastRunTime);
CREATE INDEX IDX_TIMER_TASK_ACTIVE ON HUB_TIMER_TASK(activeFlag);
CREATE INDEX IDX_TIMER_TASK_EXEC ON HUB_TIMER_TASK(executorType);
CREATE INDEX IDX_TIMER_TASK_TOOL ON HUB_TIMER_TASK(toolConfigId);
CREATE INDEX IDX_TIMER_TASK_OP ON HUB_TIMER_TASK(operationType);