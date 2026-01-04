CREATE TABLE `HUB_TIMER_TASK` (
  -- 主键信息
  `taskId` VARCHAR(32) NOT NULL COMMENT '任务ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  
  -- 任务配置信息
  `taskName` VARCHAR(200) NOT NULL COMMENT '任务名称',
  `taskDescription` VARCHAR(500) DEFAULT NULL COMMENT '任务描述',
  `taskPriority` INT NOT NULL DEFAULT 1 COMMENT '任务优先级(1低优先级,2普通优先级,3高优先级)',
  `schedulerId` VARCHAR(32) DEFAULT NULL COMMENT '关联的调度器ID',
  `schedulerName` VARCHAR(100) DEFAULT NULL COMMENT '调度器名称（冗余字段，便于查询显示）',
  
  -- 调度配置
  `scheduleType` INT NOT NULL COMMENT '调度类型(1一次性执行,2固定间隔,3Cron表达式,4延迟执行,5实时执行)',
  `cronExpression` VARCHAR(100) DEFAULT NULL COMMENT 'Cron表达式，scheduleType=3时必填',
  `intervalSeconds` BIGINT DEFAULT NULL COMMENT '执行间隔秒数，scheduleType=2时必填',
  `delaySeconds` BIGINT DEFAULT NULL COMMENT '延迟秒数，scheduleType=4时必填',
  `startTime` DATETIME DEFAULT NULL COMMENT '任务开始时间',
  `endTime` DATETIME DEFAULT NULL COMMENT '任务结束时间',
  
  -- 执行配置
  `maxRetries` INT NOT NULL DEFAULT 0 COMMENT '最大重试次数',
  `retryIntervalSeconds` BIGINT NOT NULL DEFAULT 60 COMMENT '重试间隔秒数',
  `timeoutSeconds` BIGINT NOT NULL DEFAULT 1800 COMMENT '执行超时时间秒数',
  `taskParams` TEXT DEFAULT NULL COMMENT '任务参数，JSON格式存储',
  
  -- 任务执行器配置 - 关联到具体工具配置
  `executorType` VARCHAR(50) DEFAULT NULL COMMENT '执行器类型(BUILTIN内置,SFTP文件传输,SSH远程执行,DATABASE数据库,HTTP接口调用等)',
  `toolConfigId` VARCHAR(32) DEFAULT NULL COMMENT '关联的工具配置ID（如SFTP配置ID、数据库配置ID等）',
  `toolConfigName` VARCHAR(100) DEFAULT NULL COMMENT '工具配置名称（冗余字段，便于显示）',
  `operationType` VARCHAR(100) DEFAULT NULL COMMENT '执行操作类型（如文件上传、下载、SQL执行、接口调用等）',
  `operationConfig` TEXT DEFAULT NULL COMMENT '操作参数配置，JSON格式存储具体操作的参数',
  
  -- 运行时状态
  `taskStatus` INT NOT NULL DEFAULT 1 COMMENT '任务状态(1待执行,2运行中,3已完成,4执行失败,5已取消)',
  `nextRunTime` DATETIME DEFAULT NULL COMMENT '下次执行时间',
  `lastRunTime` DATETIME DEFAULT NULL COMMENT '上次执行时间',
  `runCount` BIGINT NOT NULL DEFAULT 0 COMMENT '执行总次数',
  `successCount` BIGINT NOT NULL DEFAULT 0 COMMENT '成功次数',
  `failureCount` BIGINT NOT NULL DEFAULT 0 COMMENT '失败次数',
  
  -- 最后执行结果
  `lastExecutionId` VARCHAR(32) DEFAULT NULL COMMENT '最后执行ID',
  `lastExecutionStartTime` DATETIME DEFAULT NULL COMMENT '最后执行开始时间',
  `lastExecutionEndTime` DATETIME DEFAULT NULL COMMENT '最后执行结束时间',
  `lastExecutionDurationMs` BIGINT DEFAULT NULL COMMENT '最后执行耗时毫秒数',
  `lastExecutionStatus` INT DEFAULT NULL COMMENT '最后执行状态',
  `lastResultSuccess` VARCHAR(1) DEFAULT NULL COMMENT '最后执行是否成功(N失败,Y成功)',
  `lastErrorMessage` TEXT DEFAULT NULL COMMENT '最后错误信息',
  `lastRetryCount` INT DEFAULT NULL COMMENT '最后重试次数',
  
  -- 通用字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性，JSON格式',
  `reserved1` VARCHAR(500) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(500) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` VARCHAR(500) DEFAULT NULL COMMENT '预留字段3',
  `reserved4` VARCHAR(500) DEFAULT NULL COMMENT '预留字段4',
  `reserved5` VARCHAR(500) DEFAULT NULL COMMENT '预留字段5',
  `reserved6` VARCHAR(500) DEFAULT NULL COMMENT '预留字段6',
  `reserved7` VARCHAR(500) DEFAULT NULL COMMENT '预留字段7',
  `reserved8` VARCHAR(500) DEFAULT NULL COMMENT '预留字段8',
  `reserved9` VARCHAR(500) DEFAULT NULL COMMENT '预留字段9',
  `reserved10` VARCHAR(500) DEFAULT NULL COMMENT '预留字段10',
  
  PRIMARY KEY (`tenantId`, `taskId`),
  KEY `IDX_TIMER_TASK_NAME` (`taskName`),
  KEY `IDX_TIMER_TASK_SCHED` (`schedulerId`),
  KEY `IDX_TIMER_TASK_TYPE` (`scheduleType`),
  KEY `IDX_TIMER_TASK_STATUS` (`taskStatus`),
  KEY `IDX_TIMER_TASK_NEXT` (`nextRunTime`),
  KEY `IDX_TIMER_TASK_LAST` (`lastRunTime`),
  KEY `IDX_TIMER_TASK_ACTIVE` (`activeFlag`),
  KEY `IDX_TIMER_TASK_EXEC` (`executorType`),
  KEY `IDX_TIMER_TASK_TOOL` (`toolConfigId`),
  KEY `IDX_TIMER_TASK_OP` (`operationType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表 - 合并任务配置、运行时信息和最后执行结果';
