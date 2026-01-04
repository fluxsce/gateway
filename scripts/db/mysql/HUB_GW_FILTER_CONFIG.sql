CREATE TABLE `HUB_GW_FILTER_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `filterConfigId` VARCHAR(32) NOT NULL COMMENT '过滤器配置ID',
  `gatewayInstanceId` VARCHAR(32) DEFAULT NULL COMMENT '网关实例ID(实例级过滤器)',
  `routeConfigId` VARCHAR(32) DEFAULT NULL COMMENT '路由配置ID(路由级过滤器)',
  `filterName` VARCHAR(100) NOT NULL COMMENT '过滤器名称',
  
  -- 根据FilterType枚举值设计
  `filterType` VARCHAR(50) NOT NULL COMMENT '过滤器类型(header,query-param,body,url,method,cookie,response)',
  
  -- 根据FilterAction枚举值设计
  `filterAction` VARCHAR(50) NOT NULL COMMENT '过滤器执行时机(pre-routing,post-routing,pre-response)',
  
  `filterOrder` INT NOT NULL DEFAULT 0 COMMENT '过滤器执行顺序(Priority)',
  `filterConfig` TEXT NOT NULL COMMENT '过滤器具体配置,JSON格式',
  `filterDesc` VARCHAR(200) DEFAULT NULL COMMENT '过滤器描述',
  
  -- 根据FilterConfig结构设计的附属字段
  `configId` VARCHAR(100) DEFAULT NULL COMMENT '过滤器配置ID(来自FilterConfig.ID)',
  
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
  PRIMARY KEY (`tenantId`, `filterConfigId`),
  INDEX `IDX_GW_FILTER_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_FILTER_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_FILTER_TYPE` (`filterType`),
  INDEX `IDX_GW_FILTER_ACTION` (`filterAction`),
  INDEX `IDX_GW_FILTER_ORDER` (`filterOrder`),
  INDEX `IDX_GW_FILTER_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='过滤器配置表 - 根据filter.go逻辑设计,支持7种类型和3种执行时机';

