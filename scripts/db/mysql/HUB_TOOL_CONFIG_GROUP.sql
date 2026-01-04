CREATE TABLE `HUB_TOOL_CONFIG_GROUP` (
  `configGroupId` VARCHAR(32) NOT NULL COMMENT '配置分组ID',
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  
  -- 分组信息
  `groupName` VARCHAR(100) NOT NULL COMMENT '分组名称',
  `groupDescription` VARCHAR(500) DEFAULT NULL COMMENT '分组描述',
  `parentGroupId` VARCHAR(32) DEFAULT NULL COMMENT '父分组ID，支持层级结构',
  `groupLevel` INT DEFAULT 1 COMMENT '分组层级，从1开始',
  `groupPath` VARCHAR(500) DEFAULT NULL COMMENT '分组路径，如/root/parent/child',
  
  -- 分组属性
  `groupType` VARCHAR(50) DEFAULT NULL COMMENT '分组类型，如environment、project、department',
  `sortOrder` INT DEFAULT 100 COMMENT '排序顺序，数值越小越靠前',
  `groupIcon` VARCHAR(100) DEFAULT NULL COMMENT '分组图标',
  `groupColor` VARCHAR(20) DEFAULT NULL COMMENT '分组颜色代码',
  
  -- 权限控制
  `accessLevel` VARCHAR(20) DEFAULT 'private' COMMENT '访问级别，如private、public、restricted',
  `allowedUsers` TEXT DEFAULT NULL COMMENT '允许访问的用户列表，JSON格式',
  `allowedRoles` TEXT DEFAULT NULL COMMENT '允许访问的角色列表，JSON格式',
  
  -- 标准字段
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
  
  PRIMARY KEY (`tenantId`, `configGroupId`),
  KEY `IDX_TOOL_GROUP_NAME` (`groupName`),
  KEY `IDX_TOOL_GROUP_PARENT` (`parentGroupId`),
  KEY `IDX_TOOL_GROUP_TYPE` (`groupType`),
  KEY `IDX_TOOL_GROUP_SORT` (`sortOrder`),
  KEY `IDX_TOOL_GROUP_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工具配置分组表 - 用于对工具配置进行分组管理';
