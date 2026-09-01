CREATE TABLE `HUB_GW_STATIC_HOST_CONFIG` (
  `tenantId` VARCHAR(32) NOT NULL COMMENT '租户ID',
  `staticHostConfigId` VARCHAR(32) NOT NULL COMMENT '静态托管配置ID',
  `routeConfigId` VARCHAR(32) NOT NULL COMMENT '路由配置ID(路由级本机目录托管)',
  `configName` VARCHAR(100) NOT NULL COMMENT '配置名称',
  `rootDirectory` VARCHAR(500) NOT NULL COMMENT '本机根目录，绝对或相对路径',
  `stripRoutePrefix` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否剥离已匹配路由前缀(N否,Y是)',
  `indexFiles` TEXT DEFAULT NULL COMMENT '目录索引文件,JSON数组或逗号分隔',
  `rewriteRules` TEXT DEFAULT NULL COMMENT '路径重写规则,JSON数组[{mode,from,to}]',
  `spaFallback` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '无扩展名缺失路径是否回退索引(N否,Y是)',
  `cacheControlMaxAge` INT NOT NULL DEFAULT 3600 COMMENT '普通文件Cache-Control max-age(秒)',
  `allowedExtensions` TEXT DEFAULT NULL COMMENT '允许的文件扩展名,JSON数组或逗号分隔,空表示不额外限制',
  `maxFileSizeBytes` BIGINT NOT NULL DEFAULT 0 COMMENT '单文件大小上限(字节),0表示不限制',
  `followSymlinks` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否跟随符号链接(N否,Y是),仍须落在根目录内',
  `enablePrecompress` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否出预压缩文件.br/.gz(N否,Y是)',
  `errorPage404` VARCHAR(200) DEFAULT NULL COMMENT '根目录内自定义404页面路径',
  `errorPage403` VARCHAR(200) DEFAULT NULL COMMENT '根目录内自定义403页面路径',
  `configPriority` INT NOT NULL DEFAULT 0 COMMENT '配置优先级,数值越小优先级越高',
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
  PRIMARY KEY (`tenantId`, `staticHostConfigId`),
  INDEX `IDX_GW_STATIC_ROUTE` (`routeConfigId`),
  INDEX `IDX_GW_STATIC_PRIORITY` (`configPriority`),
  INDEX `IDX_GW_STATIC_ACTIVE` (`activeFlag`)
) ENGINE=InnoDB COMMENT='静态托管配置表 - 路由级本机目录托管,命中后不再转发上游';

-- 已建表环境补列（新库 CREATE 已包含，重复执行会报列已存在，可忽略）
-- ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `rewriteRules` TEXT DEFAULT NULL COMMENT '路径重写规则,JSON数组[{mode,from,to}]';
-- ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `allowedExtensions` TEXT DEFAULT NULL COMMENT '允许的文件扩展名';
-- ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `maxFileSizeBytes` BIGINT NOT NULL DEFAULT 0 COMMENT '单文件大小上限(字节)';
-- ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `followSymlinks` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '是否跟随符号链接';
-- ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `enablePrecompress` VARCHAR(1) NOT NULL DEFAULT 'Y' COMMENT '是否出预压缩文件';
-- ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `errorPage404` VARCHAR(200) DEFAULT NULL COMMENT '自定义404页面';
-- ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `errorPage403` VARCHAR(200) DEFAULT NULL COMMENT '自定义403页面';

-- 历史库升级：建表已执行过的不会再跑 CREATE，只补列。
ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `redirectDirectorySlash` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '目录缺斜杠是否301到path/(N否,Y是)';
ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `rootTokenExact` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '占位符是否只精确匹配第一段(N否,Y是)';
ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `fallbackRoots` TEXT DEFAULT NULL COMMENT '备用查找目录,一行一个,最多3个';
ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `cacheControlByExt` TEXT DEFAULT NULL COMMENT '按扩展名覆盖缓存秒数,如.js=86400';
ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `enableGzip` VARCHAR(1) NOT NULL DEFAULT 'N' COMMENT '无预压缩时是否现场gzip(N否,Y是)';
ALTER TABLE `HUB_GW_STATIC_HOST_CONFIG` ADD COLUMN `securityHeaders` TEXT DEFAULT NULL COMMENT '页面安全头,一行一个 名: 值';
