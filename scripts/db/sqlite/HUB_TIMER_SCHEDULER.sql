
-- 15. 定时任务调度器表
CREATE TABLE IF NOT EXISTS HUB_TIMER_SCHEDULER (
    schedulerId TEXT NOT NULL,
    tenantId TEXT NOT NULL,
    schedulerName TEXT NOT NULL,
    schedulerInstanceId TEXT,
    maxWorkers INTEGER NOT NULL DEFAULT 5,
    queueSize INTEGER NOT NULL DEFAULT 100,
    defaultTimeoutSeconds INTEGER NOT NULL DEFAULT 1800,
    defaultRetries INTEGER NOT NULL DEFAULT 3,
    schedulerStatus INTEGER NOT NULL DEFAULT 1,
    lastStartTime DATETIME,
    lastStopTime DATETIME,
    serverName TEXT,
    serverIp TEXT,
    serverPort INTEGER,
    totalTaskCount INTEGER NOT NULL DEFAULT 0,
    runningTaskCount INTEGER NOT NULL DEFAULT 0,
    lastHeartbeatTime DATETIME,
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
    PRIMARY KEY (tenantId, schedulerId)
);
CREATE INDEX IDX_TIMER_SCHED_NAME ON HUB_TIMER_SCHEDULER(schedulerName);
CREATE INDEX IDX_TIMER_SCHED_INST ON HUB_TIMER_SCHEDULER(schedulerInstanceId);
CREATE INDEX IDX_TIMER_SCHED_STATUS ON HUB_TIMER_SCHEDULER(schedulerStatus);
CREATE INDEX IDX_TIMER_SCHED_HEART ON HUB_TIMER_SCHEDULER(lastHeartbeatTime);