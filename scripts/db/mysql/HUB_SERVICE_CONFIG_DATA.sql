-- 配置数据表 - 存储具体的配置数据，支持多种配置格式
CREATE TABLE `HUB_SERVICE_CONFIG_DATA` (
  -- 主键和租户信息
  `configDataId` VARCHAR(100) NOT NULL COMMENT '配置数据ID，主键',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID，用于多租户数据隔离',
  
  -- 关联命名空间和分组
  `namespaceId` VARCHAR(32) NOT NULL COMMENT '命名空间ID，关联HUB_SERVICE_NAMESPACE表',
  `groupName` VARCHAR(64) NOT NULL DEFAULT 'DEFAULT_GROUP' COMMENT '分组名称，如DEFAULT_GROUP',
  
  -- 配置基本信息
  `configContent` LONGTEXT NOT NULL COMMENT '配置内容，支持大文本',
  `contentType` VARCHAR(50) DEFAULT 'text' COMMENT '内容类型(text:文本,json:JSON,xml:XML,yaml:YAML,properties:Properties)',
  
  -- 配置描述和属性
  `configDescription` VARCHAR(500) DEFAULT NULL COMMENT '配置描述',
  `encrypted` VARCHAR(1) DEFAULT 'N' COMMENT '是否加密存储(N否,Y是)',
  
  -- 版本信息
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '配置版本号，每次修改递增',
  
  -- MD5校验值（用于配置变更检测）
  `md5Value` VARCHAR(32) DEFAULT NULL COMMENT '配置内容的MD5值，用于快速比较配置是否变更',
  
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
  
  -- 主键和索引
  PRIMARY KEY (`tenantId`, `namespaceId`, `groupName`, `configDataId`),
  KEY `IDX_SVC_CFG_DATA_NS_ID` (`tenantId`, `namespaceId`),
  KEY `IDX_SVC_CFG_DATA_MD5` (`md5Value`),
  KEY `IDX_SVC_CFG_DATA_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置数据表 - 存储具体的配置数据，支持多种配置格式和版本管理';

