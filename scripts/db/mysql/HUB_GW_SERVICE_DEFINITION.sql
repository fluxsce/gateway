CREATE TABLE `HUB_GW_SERVICE_DEFINITION` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `serviceDefinitionId` VARCHAR(32) NOT NULL COMMENT '服务定义ID',
  `serviceName` VARCHAR(100) NOT NULL COMMENT '服务名称',
  `serviceDesc` VARCHAR(200) DEFAULT NULL COMMENT '服务描述',
  `serviceType` INT NOT NULL DEFAULT 0 COMMENT '服务类型(0静态配置,1服务发现)',
  
  -- 代理配置关联字段
  `proxyConfigId` VARCHAR(32) NOT NULL COMMENT '关联的代理配置ID',
  
  -- 根据ServiceConfig.Strategy字段设计负载均衡策略
  `loadBalanceStrategy` VARCHAR(50) NOT NULL DEFAULT 'round-robin' COMMENT '负载均衡策略(round-robin,random,ip-hash,least-conn,weighted-round-robin,consistent-hash)',
  
  -- 服务发现配置
  `discoveryType` VARCHAR(50) DEFAULT NULL COMMENT '服务发现类型(CONSUL,EUREKA,NACOS等)',
  `discoveryConfig` TEXT DEFAULT NULL COMMENT '服务发现配置,JSON格式',
  
  -- 根据LoadBalancerConfig结构设计负载均衡配置
  `sessionAffinity` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用会话亲和性(N否,Y是)',
  `stickySession` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用粘性会话(N否,Y是)',
  `maxRetries` INT NOT NULL DEFAULT 3 COMMENT '最大重试次数',
  `retryTimeoutMs` INT NOT NULL DEFAULT 5000 COMMENT '重试超时时间(毫秒)',
  `enableCircuitBreaker` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用熔断器(N否,Y是)',
  
  -- 根据HealthConfig结构设计健康检查配置
  `healthCheckEnabled` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用健康检查(N否,Y是)',
  `healthCheckPath` VARCHAR(200) DEFAULT '/health' COMMENT '健康检查路径',
  `healthCheckMethod` VARCHAR(10) DEFAULT 'GET' COMMENT '健康检查方法',
  `healthCheckIntervalSeconds` INT DEFAULT 30 COMMENT '健康检查间隔(秒)',
  `healthCheckTimeoutMs` INT DEFAULT 5000 COMMENT '健康检查超时(毫秒)',
  `healthyThreshold` INT DEFAULT 2 COMMENT '健康阈值',
  `unhealthyThreshold` INT DEFAULT 3 COMMENT '不健康阈值',
  `expectedStatusCodes` VARCHAR(200) DEFAULT '200' COMMENT '期望的状态码,逗号分隔',
  `healthCheckHeaders` TEXT DEFAULT NULL COMMENT '健康检查请求头,JSON格式',
  
  -- 负载均衡器配置(JSON格式存储完整的LoadBalancerConfig)
  `loadBalancerConfig` TEXT DEFAULT NULL COMMENT '负载均衡器完整配置,JSON格式',
  `serviceMetadata` TEXT DEFAULT NULL COMMENT '服务元数据,JSON格式',
  
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
  PRIMARY KEY (`tenantId`, `serviceDefinitionId`),
  INDEX `IDX_GW_SVC_NAME` (`serviceName`),
  INDEX `IDX_GW_SVC_TYPE` (`serviceType`),
  INDEX `IDX_GW_SVC_STRATEGY` (`loadBalanceStrategy`),
  INDEX `IDX_GW_SVC_HEALTH` (`healthCheckEnabled`),
  INDEX `IDX_GW_SVC_PROXY` (`proxyConfigId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务定义表 - 根据ServiceConfig结构设计,存储完整的服务配置';

