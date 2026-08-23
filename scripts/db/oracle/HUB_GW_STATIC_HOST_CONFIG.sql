CREATE TABLE HUB_GW_STATIC_HOST_CONFIG (
                                                    tenantId             VARCHAR2(32) NOT NULL, -- 租户ID
                                                    staticHostConfigId   VARCHAR2(32) NOT NULL, -- 静态托管配置ID
                                                    routeConfigId        VARCHAR2(32) NOT NULL, -- 路由配置ID(路由级本机目录托管)
                                                    configName           VARCHAR2(100) NOT NULL, -- 配置名称
                                                    rootDirectory        VARCHAR2(500) NOT NULL, -- 本机根目录，绝对或相对路径
                                                    stripRoutePrefix     VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否剥离已匹配路由前缀(N否,Y是)
                                                    indexFiles           CLOB, -- 目录索引文件,JSON数组或逗号分隔
                                                    rewriteRules         CLOB, -- 路径重写规则,JSON数组[{mode,from,to}]
                                                    spaFallback          VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 无扩展名缺失路径是否回退索引(N否,Y是)
                                                    cacheControlMaxAge   NUMBER(10) DEFAULT 3600 NOT NULL, -- 普通文件Cache-Control max-age(秒)
                                                    allowedExtensions    CLOB, -- 允许的文件扩展名,JSON数组或逗号分隔
                                                    maxFileSizeBytes     NUMBER(19) DEFAULT 0 NOT NULL, -- 单文件大小上限(字节),0表示不限制
                                                    followSymlinks       VARCHAR2(1) DEFAULT 'N' NOT NULL, -- 是否跟随符号链接(N否,Y是)
                                                    enablePrecompress    VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 是否出预压缩文件.br/.gz
                                                    errorPage404         VARCHAR2(200), -- 根目录内自定义404页面路径
                                                    errorPage403         VARCHAR2(200), -- 根目录内自定义403页面路径
                                                    configPriority       NUMBER(10) DEFAULT 0 NOT NULL, -- 配置优先级,数值越小优先级越高
                                                    reserved1            VARCHAR2(100), -- 预留字段1
                                                    reserved2            VARCHAR2(100), -- 预留字段2
                                                    reserved3            NUMBER(10), -- 预留字段3
                                                    reserved4            NUMBER(10), -- 预留字段4
                                                    reserved5            DATE, -- 预留字段5
                                                    extProperty          CLOB, -- 扩展属性,JSON格式
                                                    addTime              DATE DEFAULT SYSDATE NOT NULL, -- 创建时间
                                                    addWho               VARCHAR2(32) NOT NULL, -- 创建人ID
                                                    editTime             DATE DEFAULT SYSDATE NOT NULL, -- 最后修改时间
                                                    editWho              VARCHAR2(32) NOT NULL, -- 最后修改人ID
                                                    oprSeqFlag           VARCHAR2(32) NOT NULL, -- 操作序列标识
                                                    currentVersion       NUMBER(10) DEFAULT 1 NOT NULL, -- 当前版本号
                                                    activeFlag           VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记(N非活动,Y活动)
                                                    noteText             VARCHAR2(500), -- 备注信息
                                                    CONSTRAINT PK_GW_STATIC_HOST_CONFIG PRIMARY KEY (tenantId, staticHostConfigId)
);
CREATE INDEX IDX_GW_STATIC_ROUTE ON HUB_GW_STATIC_HOST_CONFIG(routeConfigId);
CREATE INDEX IDX_GW_STATIC_PRIORITY ON HUB_GW_STATIC_HOST_CONFIG(configPriority);
CREATE INDEX IDX_GW_STATIC_ACTIVE ON HUB_GW_STATIC_HOST_CONFIG(activeFlag);
COMMENT ON TABLE HUB_GW_STATIC_HOST_CONFIG IS '静态托管配置表 - 路由级本机目录托管,命中后不再转发上游';

-- 历史库升级：建表已执行过的不会再跑 CREATE，只补列。
ALTER TABLE HUB_GW_STATIC_HOST_CONFIG ADD redirectDirectorySlash VARCHAR2(1) DEFAULT 'N';
ALTER TABLE HUB_GW_STATIC_HOST_CONFIG ADD rootTokenExact VARCHAR2(1) DEFAULT 'N';
UPDATE HUB_GW_STATIC_HOST_CONFIG SET redirectDirectorySlash = 'N' WHERE redirectDirectorySlash IS NULL;
UPDATE HUB_GW_STATIC_HOST_CONFIG SET rootTokenExact = 'N' WHERE rootTokenExact IS NULL;
ALTER TABLE HUB_GW_STATIC_HOST_CONFIG ADD fallbackRoots CLOB;
ALTER TABLE HUB_GW_STATIC_HOST_CONFIG ADD cacheControlByExt CLOB;
ALTER TABLE HUB_GW_STATIC_HOST_CONFIG ADD enableGzip VARCHAR2(1) DEFAULT 'N';
ALTER TABLE HUB_GW_STATIC_HOST_CONFIG ADD securityHeaders CLOB;
UPDATE HUB_GW_STATIC_HOST_CONFIG SET enableGzip = 'N' WHERE enableGzip IS NULL;
