CREATE TABLE `HUB_GW_ROUTER_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `routerConfigId` VARCHAR(32) NOT NULL COMMENT 'Router配置ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '关联的网关实例ID',
  `routerName` VARCHAR(100) NOT NULL COMMENT 'Router名称',
  `routerDesc` VARCHAR(200) DEFAULT NULL COMMENT 'Router描述',
  
  -- Router基础配置
  `defaultPriority` INT NOT NULL DEFAULT 100 COMMENT '默认路由优先级',
  `enableRouteCache` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用路由缓存(N否,Y是)',
  `routeCacheTtlSeconds` INT NOT NULL DEFAULT 300 COMMENT '路由缓存TTL(秒)',
  `maxRoutes` INT DEFAULT 1000 COMMENT '最大路由数量限制',
  `routeMatchTimeout` INT DEFAULT 100 COMMENT '路由匹配超时时间(毫秒)',
  
  -- Router高级配置
  `enableStrictMode` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用严格模式(N否,Y是)',
  `enableMetrics` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用路由指标收集(N否,Y是)',
  `enableTracing` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用链路追踪(N否,Y是)',
  `caseSensitive` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '路径匹配是否区分大小写(N否,Y是)',
  `removeTrailingSlash` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否移除路径尾部斜杠(N否,Y是)',
  
  -- 路由处理配置
  `enableGlobalFilters` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用全局过滤器(N否,Y是)',
  `filterExecutionMode` VARCHAR(20) NOT NULL DEFAULT 'SEQUENTIAL' COMMENT '过滤器执行模式(SEQUENTIAL顺序,PARALLEL并行)',
  `maxFilterChainDepth` INT DEFAULT 50 COMMENT '最大过滤器链深度',
  
  -- 性能优化配置
  `enableRoutePooling` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用路由对象池(N否,Y是)',
  `routePoolSize` INT DEFAULT 100 COMMENT '路由对象池大小',
  `enableAsyncProcessing` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否启用异步处理(N否,Y是)',
  
  -- 错误处理配置
  `enableFallback` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否启用降级处理(N否,Y是)',
  `fallbackRoute` VARCHAR(200) DEFAULT NULL COMMENT '降级路由路径',
  `notFoundStatusCode` INT NOT NULL DEFAULT 404 COMMENT '路由未找到时的状态码',
  `notFoundMessage` VARCHAR(200) DEFAULT 'Route not found' COMMENT '路由未找到时的提示消息',
  
  -- 自定义配置
  `routerMetadata` TEXT DEFAULT NULL COMMENT 'Router元数据,JSON格式',
  `customConfig` TEXT DEFAULT NULL COMMENT '自定义配置,JSON格式',
  
  -- 标准数据库字段
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
  `activeFlag` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '活动状态标记(N非活动/禁用,Y活动/启用)',
  `noteText` VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
  
  PRIMARY KEY (`tenantId`, `routerConfigId`),
  INDEX `IDX_GW_ROUTER_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_ROUTER_NAME` (`routerName`),
  INDEX `IDX_GW_ROUTER_ACTIVE` (`activeFlag`),
  INDEX `IDX_GW_ROUTER_CACHE` (`enableRouteCache`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Router配置表 - 存储网关Router级别配置';

