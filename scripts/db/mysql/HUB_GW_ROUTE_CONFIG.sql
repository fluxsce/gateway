CREATE TABLE `HUB_GW_ROUTE_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `routeConfigId` VARCHAR(32) NOT NULL COMMENT '路由配置ID',
  `gatewayInstanceId` VARCHAR(32) NOT NULL COMMENT '关联的网关实例ID',
  `routeName` VARCHAR(100) NOT NULL COMMENT '路由名称',
  `routePath` VARCHAR(200) NOT NULL COMMENT '路由路径',
  `allowedMethods` VARCHAR(200) DEFAULT NULL COMMENT '允许的HTTP方法,JSON数组格式["GET","POST"]',
  `allowedHosts` VARCHAR(500) DEFAULT NULL COMMENT '允许的域名,逗号分隔',
  `matchType` INT NOT NULL DEFAULT 1 COMMENT '匹配类型(0精确匹配,1前缀匹配,2正则匹配)',
  `routePriority` INT NOT NULL DEFAULT 100 COMMENT '路由优先级,数值越小优先级越高',
  `stripPathPrefix` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否剥离路径前缀(N否,Y是)',
  `rewritePath` VARCHAR(200) DEFAULT NULL COMMENT '重写路径',
  `enableWebsocket` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否支持WebSocket(N否,Y是)',
  `timeoutMs` INT NOT NULL DEFAULT 30000 COMMENT '超时时间(毫秒)',
  `retryCount` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `retryIntervalMs` INT NOT NULL DEFAULT 1000 COMMENT '重试间隔(毫秒)',
  
  -- 服务关联字段，直接关联服务定义表
  `serviceDefinitionId` VARCHAR(32) DEFAULT NULL COMMENT '关联的服务定义ID',
  
  -- 日志配置关联字段
  `logConfigId` VARCHAR(32) DEFAULT NULL COMMENT '关联的日志配置ID(路由级日志配置)',
  
  -- 路由元数据，用于存储额外配置信息
  `routeMetadata` TEXT DEFAULT NULL COMMENT '路由元数据,JSON格式,存储Methods等配置',
  
  -- 注意：使用activeFlag代替enabled字段，保持数据库设计一致性
  -- activeFlag='Y'表示路由启用，activeFlag='N'表示路由禁用
  -- 在代码中将activeFlag映射为enabled字段
  
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
  PRIMARY KEY (`tenantId`, `routeConfigId`),
  INDEX `IDX_GW_ROUTE_INST` (`gatewayInstanceId`),
  INDEX `IDX_GW_ROUTE_SERVICE` (`serviceDefinitionId`),
  INDEX `IDX_GW_ROUTE_LOG` (`logConfigId`),
  INDEX `IDX_GW_ROUTE_PRIORITY` (`routePriority`),
  INDEX `IDX_GW_ROUTE_PATH` (`routePath`),
  INDEX `IDX_GW_ROUTE_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路由定义表 - 存储API路由配置,使用activeFlag统一管理启用状态';

-- 兼容已存在的数据库，修改字段长度
-- 注意：由于字段长度增加到 VARCHAR(1000)，需要先删除索引，修改字段后再重建前缀索引
ALTER TABLE `HUB_GW_ROUTE_CONFIG` DROP INDEX `IDX_GW_ROUTE_SERVICE`;
ALTER TABLE `HUB_GW_ROUTE_CONFIG` MODIFY COLUMN `serviceDefinitionId` VARCHAR(1000) DEFAULT NULL COMMENT '关联的服务定义ID';
-- 重建索引，使用前缀索引（前255个字符，255*4=1020字节，在3072字节限制内）
ALTER TABLE `HUB_GW_ROUTE_CONFIG` ADD INDEX `IDX_GW_ROUTE_SERVICE` (`serviceDefinitionId`(255));

