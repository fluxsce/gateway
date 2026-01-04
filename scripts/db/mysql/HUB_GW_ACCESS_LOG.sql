CREATE TABLE `HUB_GW_ACCESS_LOG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `traceId` VARCHAR(64) NOT NULL COMMENT '链路追踪ID(作为主键)',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '网关实例ID',
  `gatewayInstanceName` VARCHAR(300) DEFAULT NULL COMMENT '网关实例名称(冗余字段,便于查询显示)',
  `gatewayNodeIp` VARCHAR(50) NOT NULL COMMENT '网关节点IP地址',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID',
  `routeName` VARCHAR(300) DEFAULT NULL COMMENT '路由名称(冗余字段,便于查询显示)',
  `serviceDefinitionId` VARCHAR(32) DEFAULT NULL COMMENT '服务定义ID',
  `serviceName` VARCHAR(300) DEFAULT NULL COMMENT '服务名称(冗余字段,便于查询显示)',
  `proxyType` VARCHAR(50) DEFAULT NULL COMMENT '代理类型(http,websocket,tcp,udp,可为空)',
  `logConfigId` VARCHAR(32) DEFAULT NULL COMMENT '日志配置ID',
  
  -- 请求基本信息
  `requestMethod` VARCHAR(10) NOT NULL COMMENT '请求方法(GET,POST,PUT等)',
  `requestPath` VARCHAR(1000) NOT NULL COMMENT '请求路径',
  `requestQuery` TEXT DEFAULT NULL COMMENT '请求查询参数',
  `requestSize` INT DEFAULT 0 COMMENT '请求大小(字节)',
  `requestHeaders` TEXT DEFAULT NULL COMMENT '请求头信息,JSON格式',
  `requestBody` TEXT DEFAULT NULL COMMENT '请求体(可选,根据配置决定是否记录)',
  
  -- 客户端信息
  `clientIpAddress` VARCHAR(50) NOT NULL COMMENT '客户端IP地址',
  `clientPort` INT DEFAULT NULL COMMENT '客户端端口',
  `userAgent` VARCHAR(1000) DEFAULT NULL COMMENT '用户代理信息',
  `referer` VARCHAR(1000) DEFAULT NULL COMMENT '来源页面',
  `userIdentifier` VARCHAR(100) DEFAULT NULL COMMENT '用户标识(如有)',
  
  -- 关键时间点 (所有时间字段均为DATETIME类型，精确到毫秒)
  `gatewayStartProcessingTime` DATETIME(3) NOT NULL COMMENT '网关开始处理时间(请求开始处理，必填)',
  `backendRequestStartTime` DATETIME(3) DEFAULT NULL COMMENT '后端服务请求开始时间(可选)',
  `backendResponseReceivedTime` DATETIME(3) DEFAULT NULL COMMENT '后端服务响应接收时间(可选)',
  `gatewayFinishedProcessingTime` DATETIME(3) DEFAULT NULL COMMENT '网关处理完成时间(可选，正在处理中或异常中断时为空)',
  
  -- 计算的时间指标 (所有时间指标均为毫秒)
  `totalProcessingTimeMs` INT DEFAULT NULL COMMENT '总处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)',
  `gatewayProcessingTimeMs` INT DEFAULT NULL COMMENT '网关处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)',
  `backendResponseTimeMs` INT DEFAULT NULL COMMENT '后端服务响应时间(毫秒，可选)',
  
  -- 响应信息
  `gatewayStatusCode` INT NOT NULL COMMENT '网关响应状态码',
  `backendStatusCode` INT DEFAULT NULL COMMENT '后端服务状态码',
  `responseSize` INT DEFAULT 0 COMMENT '响应大小(字节)',
  `responseHeaders` TEXT DEFAULT NULL COMMENT '响应头信息,JSON格式',
  `responseBody` TEXT DEFAULT NULL COMMENT '响应体(可选,根据配置决定是否记录)',
  
  -- 转发基本信息
  `matchedRoute` VARCHAR(500) DEFAULT NULL COMMENT '匹配的路由路径',
  `forwardAddress` TEXT DEFAULT NULL COMMENT '转发地址',
  `forwardMethod` VARCHAR(10) DEFAULT NULL COMMENT '转发方法',
  `forwardParams` TEXT DEFAULT NULL COMMENT '转发参数,JSON格式',
  `forwardHeaders` TEXT DEFAULT NULL COMMENT '转发头信息,JSON格式',
  `forwardBody` TEXT DEFAULT NULL COMMENT '转发报文内容',
  `loadBalancerDecision` VARCHAR(500) DEFAULT NULL COMMENT '负载均衡决策信息',
  
  -- 错误信息
  `errorMessage` TEXT DEFAULT NULL COMMENT '错误信息(如有)',
  `errorCode` VARCHAR(100) DEFAULT NULL COMMENT '错误代码(如有)',
  
  -- 追踪信息
  `parentTraceId` VARCHAR(100) DEFAULT NULL COMMENT '父链路追踪ID',
  
  -- 日志重置标记和次数
  `resetFlag` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '日志重置标记(N否,Y是)',
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `resetCount` INT NOT NULL DEFAULT 0 COMMENT '重置次数',
  
  -- 标准数据库字段
  `logLevel` VARCHAR(20) NOT NULL DEFAULT 'INFO' COMMENT '日志级别',
  `logType` VARCHAR(50) NOT NULL DEFAULT 'ACCESS' COMMENT '日志类型',
  `reserved1` VARCHAR(100) DEFAULT NULL COMMENT '预留字段1',
  `reserved2` VARCHAR(100) DEFAULT NULL COMMENT '预留字段2',
  `reserved3` INT DEFAULT NULL COMMENT '预留字段3',
  `reserved4` INT DEFAULT NULL COMMENT '预留字段4',
  `reserved5` DATETIME DEFAULT NULL COMMENT '预留字段5',
  `extProperty` TEXT DEFAULT NULL COMMENT '扩展属性,JSON格式',
  `addTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `addWho` VARCHAR(32) NOT NULL COMMENT '创建人ID',
  `editTime` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `editWho` VARCHAR(32) NOT NULL COMMENT '最后修改人ID',
  `oprSeqFlag` VARCHAR(32) NOT NULL COMMENT '操作序列标识',
  `currentVersion` INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  
  PRIMARY KEY (`tenantId`, `traceId`),
  -- 核心查询索引（高频查询字段）
  INDEX `idx_HUB_GW_ACCESS_LOG_time_instance` (`gatewayStartProcessingTime`, `gatewayInstanceId`),
  INDEX `idx_HUB_GW_ACCESS_LOG_time_route` (`gatewayStartProcessingTime`, `routeConfigId`),
  INDEX `idx_HUB_GW_ACCESS_LOG_time_service` (`gatewayStartProcessingTime`, `serviceDefinitionId`),
  
  -- 名称字段查询索引（利用冗余字段，避免JOIN）
  INDEX `idx_HUB_GW_ACCESS_LOG_instance_name` (`gatewayInstanceName`, `gatewayStartProcessingTime`),
  INDEX `idx_HUB_GW_ACCESS_LOG_route_name` (`routeName`, `gatewayStartProcessingTime`),
  INDEX `idx_HUB_GW_ACCESS_LOG_service_name` (`serviceName`, `gatewayStartProcessingTime`),
  
  -- 业务查询索引
  INDEX `idx_HUB_GW_ACCESS_LOG_client_ip` (`clientIpAddress`, `gatewayStartProcessingTime`),
  INDEX `idx_HUB_GW_ACCESS_LOG_status_time` (`gatewayStatusCode`, `gatewayStartProcessingTime`),
  INDEX `idx_HUB_GW_ACCESS_LOG_proxy_type` (`proxyType`, `gatewayStartProcessingTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网关访问日志表 - 记录API网关的请求和响应详细信息,开始时间必填,完成时间可选(支持处理中状态),含冗余字段优化查询性能';
