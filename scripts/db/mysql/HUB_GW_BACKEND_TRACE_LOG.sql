CREATE TABLE `HUB_GW_BACKEND_TRACE_LOG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `traceId` VARCHAR(64) NOT NULL COMMENT '链路追踪ID，关联主表 HUB_GW_ACCESS_LOG.traceId',
  `backendTraceId` VARCHAR(64) NOT NULL COMMENT '后端服务追踪ID，用于区分同一请求的多个后端服务',

  -- 服务信息（单个后端服务一次转发一条记录）
  `serviceDefinitionId` VARCHAR(32) DEFAULT NULL COMMENT '服务定义ID',
  `serviceName` VARCHAR(300) DEFAULT NULL COMMENT '服务名称(冗余字段,便于查询)',

  -- 转发信息
  `forwardAddress` TEXT DEFAULT NULL COMMENT '实际转发目标地址(完整URL)',
  `forwardMethod` VARCHAR(10) DEFAULT NULL COMMENT '转发HTTP方法',
  `forwardPath` VARCHAR(1000) DEFAULT NULL COMMENT '转发路径',
  `forwardQuery` TEXT DEFAULT NULL COMMENT '转发查询参数',
  `forwardHeaders` LONGTEXT DEFAULT NULL COMMENT '转发请求头(JSON格式)',
  `forwardBody` LONGTEXT DEFAULT NULL COMMENT '转发请求体',
  `requestSize` INT DEFAULT 0 COMMENT '请求大小(字节，向后端发送的请求体大小)',

  -- 负载均衡信息
  `loadBalancerStrategy` VARCHAR(100) DEFAULT NULL COMMENT '负载均衡策略(round-robin, random, weighted等)',
  `loadBalancerDecision` VARCHAR(500) DEFAULT NULL COMMENT '负载均衡选择决策信息',

  -- 时间信息
  `requestStartTime` DATETIME(3) NOT NULL COMMENT '向后端发起请求的时间',
  `responseReceivedTime` DATETIME(3) DEFAULT NULL COMMENT '接收到后端响应的时间',
  `requestDurationMs` INT DEFAULT NULL COMMENT '请求耗时(毫秒,0表示未完成)',

  -- 响应信息
  `statusCode` INT DEFAULT NULL COMMENT '后端服务返回的HTTP状态码(0表示未收到响应)',
  `responseSize` INT DEFAULT 0 COMMENT '后端响应大小(字节)',
  `responseHeaders` LONGTEXT DEFAULT NULL COMMENT '后端响应头信息(JSON格式)',
  `responseBody` LONGTEXT DEFAULT NULL COMMENT '后端响应体内容',

  -- 错误信息
  `errorCode` VARCHAR(100) DEFAULT NULL COMMENT '错误代码',
  `errorMessage` LONGTEXT DEFAULT NULL COMMENT '详细错误信息',

  -- 状态信息
  `successFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否成功(Y成功,N失败)',
  `traceStatus` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '后端调用状态(pending,success,failed,timeout)',
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',

  -- 扩展信息
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性(JSON格式)',

  -- 标准数据库字段
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '记录创建者',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '记录修改者',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',

  PRIMARY KEY (`traceId`, `backendTraceId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关后端追踪日志表 - 记录每个后端服务的转发明细';

-- 索引设计：参考数据库规范，兼顾多租户与常用查询场景
CREATE INDEX `IDX_GW_BTRACE_TRACE` ON `HUB_GW_BACKEND_TRACE_LOG` (`tenantId`, `traceId`);
CREATE INDEX `IDX_GW_BTRACE_SERVICE` ON `HUB_GW_BACKEND_TRACE_LOG` (`tenantId`, `serviceDefinitionId`, `requestStartTime`);
CREATE INDEX `IDX_GW_BTRACE_TIME` ON `HUB_GW_BACKEND_TRACE_LOG` (`requestStartTime`);
CREATE INDEX `IDX_GW_BTRACE_TSTATUS` ON `HUB_GW_BACKEND_TRACE_LOG` (`tenantId`, `traceStatus`, `requestStartTime`);
CREATE INDEX `IDX_GW_BTRACE_ADDTIME` ON `HUB_GW_BACKEND_TRACE_LOG` (`tenantId`, `addTime`);
