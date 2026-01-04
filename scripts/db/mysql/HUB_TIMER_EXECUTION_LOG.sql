CREATE TABLE `HUB_TIMER_EXECUTION_LOG` (
  -- 主键信息
  `executionId` VARCHAR(32) NOT NULL COMMENT '执行ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `taskId` VARCHAR(32) NOT NULL COMMENT '关联任务ID',
  
  -- 任务信息（冗余）
  `taskName` VARCHAR(200) DEFAULT NULL COMMENT '任务名称',
  `schedulerId` VARCHAR(32) DEFAULT NULL COMMENT '调度器ID',
  
  -- 执行信息
  `executionStartTime` DATETIME NOT NULL COMMENT '执行开始时间',
  `executionEndTime` DATETIME DEFAULT NULL COMMENT '执行结束时间',
  `executionDurationMs` BIGINT DEFAULT NULL COMMENT '执行耗时毫秒数',
  `executionStatus` INT NOT NULL COMMENT '执行状态(1待执行,2运行中,3已完成,4执行失败,5已取消)',
  `resultSuccess` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '执行是否成功(N失败,Y成功)',
  
  -- 错误信息
  `errorMessage` TEXT DEFAULT NULL COMMENT '错误信息',
  `errorStackTrace` TEXT DEFAULT NULL COMMENT '错误堆栈信息',
  
  -- 重试信息
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `maxRetryCount` INT NOT NULL DEFAULT 0 COMMENT '最大重试次数',
  
  -- 参数和结果
  `executionParams` TEXT DEFAULT NULL COMMENT '执行参数，JSON格式',
  `executionResult` TEXT DEFAULT NULL COMMENT '执行结果，JSON格式',
  
  -- 执行环境
  `executorServerName` VARCHAR(100) DEFAULT NULL COMMENT '执行服务器名称',
  `executorServerIp` VARCHAR(50) DEFAULT NULL COMMENT '执行服务器IP地址',
  
  -- 日志信息
  `logLevel` VARCHAR(10) DEFAULT NULL COMMENT '日志级别(DEBUG,INFO,WARN,ERROR)',
  `logMessage` TEXT DEFAULT NULL COMMENT '日志消息内容',
  `logTimestamp` DATETIME DEFAULT NULL COMMENT '日志时间戳',
  
  -- 执行上下文
  `executionPhase` VARCHAR(50) DEFAULT NULL COMMENT '执行阶段(BEFORE_EXECUTE,EXECUTING,AFTER_EXECUTE,RETRY)',
  `threadName` VARCHAR(100) DEFAULT NULL COMMENT '执行线程名称',
  `className` VARCHAR(200) DEFAULT NULL COMMENT '执行类名',
  `methodName` VARCHAR(100) DEFAULT NULL COMMENT '执行方法名',
  
  -- 异常信息
  `exceptionClass` VARCHAR(200) DEFAULT NULL COMMENT '异常类名',
  `exceptionMessage` TEXT DEFAULT NULL COMMENT '异常消息',
  
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
  
  PRIMARY KEY (`tenantId`, `executionId`),
  KEY `IDX_TIMER_LOG_TASK` (`taskId`),
  KEY `IDX_TIMER_LOG_NAME` (`taskName`),
  KEY `IDX_TIMER_LOG_SCHED` (`schedulerId`),
  KEY `IDX_TIMER_LOG_START` (`executionStartTime`),
  KEY `IDX_TIMER_LOG_STATUS` (`executionStatus`),
  KEY `IDX_TIMER_LOG_SUCCESS` (`resultSuccess`),
  KEY `IDX_TIMER_LOG_LEVEL` (`logLevel`),
  KEY `IDX_TIMER_LOG_TIME` (`logTimestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行日志表 - 合并执行记录和日志信息';

