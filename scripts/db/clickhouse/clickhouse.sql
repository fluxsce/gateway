-- ClickHouse 网关访问日志表
-- 基于MySQL版本翻译，针对ClickHouse的列式存储特性进行优化
-- TTL设置为30天自动过期
-- 时间精度：DateTime64(3) 支持毫秒精度
-- 字符串长度：String类型理论上无长度限制，实际受内存和配置限制

CREATE TABLE HUB_GW_ACCESS_LOG
(
    -- 主键字段
    `tenantId` String COMMENT '租户ID',
    `traceId` String COMMENT '链路追踪ID(作为主键)',
    
    -- 网关实例相关信息
    `gatewayInstanceId` String COMMENT '网关实例ID',
    `gatewayInstanceName` String COMMENT '网关实例名称(冗余字段,便于查询显示)',
    `gatewayNodeIp` String COMMENT '网关节点IP地址',
    
    -- 路由和服务相关信息
    `routeConfigId` String COMMENT '路由配置ID',
    `routeName` String COMMENT '路由名称(冗余字段,便于查询显示)',
    `serviceDefinitionId` String COMMENT '服务定义ID',
    `serviceName` String COMMENT '服务名称(冗余字段,便于查询显示)',
    `proxyType` String COMMENT '代理类型(http,websocket,tcp,udp,可为空)',
    `logConfigId` String COMMENT '日志配置ID',
    
    -- 请求基本信息
    `requestMethod` String COMMENT '请求方法(GET,POST,PUT等)',
    `requestPath` String COMMENT '请求路径',
    `requestQuery` String COMMENT '请求查询参数',
    `requestSize` Int32 DEFAULT 0 COMMENT '请求大小(字节)',
    `requestHeaders` String COMMENT '请求头信息,JSON格式',
    `requestBody` String COMMENT '请求体(可选,根据配置决定是否记录)',
    
    -- 客户端信息
    `clientIpAddress` String COMMENT '客户端IP地址',
    `clientPort` Nullable(Int32) COMMENT '客户端端口',
    `userAgent` String COMMENT '用户代理信息',
    `referer` String COMMENT '来源页面',
    `userIdentifier` String COMMENT '用户标识(如有)',
    
    -- 关键时间点 (使用DateTime64(3)精确到毫秒)
    `gatewayStartProcessingTime` DateTime64(3) COMMENT '网关开始处理时间(请求开始处理，必填)',
    `backendRequestStartTime` Nullable(DateTime64(3)) COMMENT '后端服务请求开始时间(可选)',
    `backendResponseReceivedTime` Nullable(DateTime64(3)) COMMENT '后端服务响应接收时间(可选)',
    `gatewayFinishedProcessingTime` Nullable(DateTime64(3)) COMMENT '网关处理完成时间(可选，正在处理中或异常中断时为空)',
    
    -- 计算的时间指标 (所有时间指标均为毫秒)
    `totalProcessingTimeMs` Nullable(Int32) COMMENT '总处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)',
    `gatewayProcessingTimeMs` Nullable(Int32) COMMENT '网关处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL)',
    `backendResponseTimeMs` Nullable(Int32) COMMENT '后端服务响应时间(毫秒，可选)',
    
    -- 响应信息
    `gatewayStatusCode` Int32 COMMENT '网关响应状态码',
    `backendStatusCode` Nullable(Int32) COMMENT '后端服务状态码',
    `responseSize` Int32 DEFAULT 0 COMMENT '响应大小(字节)',
    `responseHeaders` String COMMENT '响应头信息,JSON格式',
    `responseBody` String COMMENT '响应体(可选,根据配置决定是否记录)',
    
    -- 转发基本信息
    `matchedRoute` String COMMENT '匹配的路由路径',
    `forwardAddress` String COMMENT '转发地址',
    `forwardMethod` String COMMENT '转发方法',
    `forwardParams` String COMMENT '转发参数,JSON格式',
    `forwardHeaders` String COMMENT '转发头信息,JSON格式',
    `forwardBody` String COMMENT '转发报文内容',
    `loadBalancerDecision` String COMMENT '负载均衡决策信息',
    
    -- 错误信息
    `errorMessage` String COMMENT '错误信息(如有)',
    `errorCode` String COMMENT '错误代码(如有)',
    
    -- 追踪信息
    `parentTraceId` String COMMENT '父链路追踪ID',
    
    -- 日志重置标记和次数
    `resetFlag` String DEFAULT 'N' COMMENT '日志重置标记(N否,Y是)',
    `retryCount` Int32 DEFAULT 0 COMMENT '重试次数',
    `resetCount` Int32 DEFAULT 0 COMMENT '重置次数',
    
    -- 标准数据库字段
    `logLevel` String DEFAULT 'INFO' COMMENT '日志级别',
    `logType` String DEFAULT 'ACCESS' COMMENT '日志类型',
    `reserved1` String COMMENT '预留字段1',
    `reserved2` String COMMENT '预留字段2',
    `reserved3` Nullable(Int32) COMMENT '预留字段3',
    `reserved4` Nullable(Int32) COMMENT '预留字段4',
    `reserved5` Nullable(DateTime) COMMENT '预留字段5',
    `extProperty` String COMMENT '扩展属性,JSON格式',
    `addTime` DateTime DEFAULT now() COMMENT '创建时间',
    `addWho` String COMMENT '创建人ID',
    `editTime` DateTime DEFAULT now() COMMENT '最后修改时间',
    `editWho` String COMMENT '最后修改人ID',
    `oprSeqFlag` String COMMENT '操作序列标识',
    `currentVersion` Int32 DEFAULT 1 COMMENT '当前版本号',
    `activeFlag` String DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
    `noteText` String COMMENT '备注信息'
)
ENGINE = MergeTree()
-- 按天分区，提高查询性能并便于数据管理
PARTITION BY toDate(gatewayStartProcessingTime)
-- 主要排序字段：租户ID + 链路追踪ID + 时间，优化常见查询
ORDER BY (gatewayStartProcessingTime, tenantId, traceId)
-- TTL设置：30天后自动删除数据
TTL gatewayStartProcessingTime + INTERVAL 30 DAY
-- 表级别设置
SETTINGS 
    -- 启用索引颗粒度自适应
    index_granularity_bytes = 10485760,
    -- 设置索引颗粒度
    index_granularity = 8192;

-- ============================================================================
-- 重要说明
-- ============================================================================
-- 1. 时间精度：DateTime64(3) 支持毫秒精度，存储格式为 YYYY-MM-DD HH:MM:SS.sss
-- 2. 字符串长度：String类型理论上无长度限制，实际受以下因素限制：
--    - 单行数据大小限制（默认max_row_size=1GB）
--    - 内存限制
--    - 网络传输限制
--    - 建议大字段控制在合理范围内（如requestBody、responseBody等）
-- 3. 性能优化：
--    - 使用countIf、avgIf、quantileIf等条件聚合函数避免子查询
--    - 合理设置时间范围，避免全表扫描
--    - 利用分区特性，查询条件包含时间范围
-- 4. 实时查询：所有统计都是实时计算，无需预聚合，ClickHouse性能足够支撑

-- ============================================================================
-- 后端追踪日志表 - HUB_GW_BACKEND_TRACE_LOG
-- 对应MySQL版本：scripts/db/mysql.sql 第1271-1335行
-- 说明：
--   1. 作为 HUB_GW_ACCESS_LOG 的从表，记录每个后端服务的转发明细
--   2. traceId + backendTraceId 作为联合排序键
--   3. 包含租户ID和标准通用字段，支持多租户和审计
-- ============================================================================
CREATE TABLE HUB_GW_BACKEND_TRACE_LOG
(
    -- 主键字段
    `tenantId` String COMMENT '租户ID',
    `traceId` String COMMENT '链路追踪ID，关联主表 HUB_GW_ACCESS_LOG.traceId',
    `backendTraceId` String COMMENT '后端服务追踪ID，用于区分同一请求的多个后端服务',

    -- 服务信息（单个后端服务一次转发一条记录）
    `serviceDefinitionId` String COMMENT '服务定义ID',
    `serviceName` String COMMENT '服务名称(冗余字段,便于查询)',

    -- 转发信息
    `forwardAddress` String COMMENT '实际转发目标地址(完整URL)',
    `forwardMethod` String COMMENT '转发HTTP方法',
    `forwardPath` String COMMENT '转发路径',
    `forwardQuery` String COMMENT '转发查询参数',
    `forwardHeaders` String COMMENT '转发请求头(JSON格式)',
    `forwardBody` String COMMENT '转发请求体',
    `requestSize` Int32 DEFAULT 0 COMMENT '请求大小(字节，向后端发送的请求体大小)',

    -- 负载均衡信息
    `loadBalancerStrategy` String COMMENT '负载均衡策略(round-robin, random, weighted等)',
    `loadBalancerDecision` String COMMENT '负载均衡选择决策信息',

    -- 时间信息
    `requestStartTime` DateTime64(3) COMMENT '向后端发起请求的时间',
    `responseReceivedTime` Nullable(DateTime64(3)) COMMENT '接收到后端响应的时间',
    `requestDurationMs` Nullable(Int32) COMMENT '请求耗时(毫秒,NULL表示未完成)',

    -- 响应信息
    `statusCode` Nullable(Int32) COMMENT '后端服务返回的HTTP状态码(NULL表示未收到响应)',
    `responseSize` Int32 DEFAULT 0 COMMENT '后端响应大小(字节)',
    `responseHeaders` String COMMENT '后端响应头信息(JSON格式)',
    `responseBody` String COMMENT '后端响应体内容',

    -- 错误信息
    `errorCode` String COMMENT '错误代码',
    `errorMessage` String COMMENT '详细错误信息',

    -- 状态信息
    `successFlag` String DEFAULT 'N' COMMENT '是否成功(Y成功,N失败)',
    `traceStatus` String DEFAULT 'pending' COMMENT '后端调用状态(pending,success,failed,timeout)',
    `retryCount` Int32 DEFAULT 0 COMMENT '重试次数',

    -- 扩展信息
    `extProperty` String COMMENT '扩展属性(JSON格式)',

    -- 标准数据库字段
    `addTime` DateTime DEFAULT now() COMMENT '记录创建时间',
    `addWho` String COMMENT '记录创建者',
    `editTime` DateTime DEFAULT now() COMMENT '记录修改时间',
    `editWho` String COMMENT '记录修改者',
    `oprSeqFlag` String COMMENT '操作序列标识',
    `currentVersion` Int32 DEFAULT 1 COMMENT '当前版本号',
    `activeFlag` String DEFAULT 'Y' COMMENT '活动状态标记(N非活动,Y活动)',
    `noteText` String COMMENT '备注信息'
)
ENGINE = MergeTree()
-- 按天分区，与主表保持一致
PARTITION BY toDate(requestStartTime)
-- 主要排序字段：租户ID + 链路追踪ID + 后端追踪ID + 时间
ORDER BY (requestStartTime, tenantId, traceId, backendTraceId)
-- TTL设置：30天后自动删除数据，与主表保持一致
TTL requestStartTime + INTERVAL 30 DAY
-- 表级别设置
SETTINGS 
    -- 启用索引颗粒度自适应
    index_granularity_bytes = 10485760,
    -- 设置索引颗粒度
    index_granularity = 8192;

-- ============================================================================
-- 后端追踪日志表说明
-- ============================================================================
-- 1. 与主表关系：通过 traceId 关联 HUB_GW_ACCESS_LOG
-- 2. 一对多关系：一个主日志可能对应多个后端追踪日志（多服务转发场景）
-- 3. 查询优化：
--    - 按 traceId 查询某个请求的所有后端转发明细
--    - 按 serviceDefinitionId 统计某个服务的调用情况
--    - 按 traceStatus 筛选失败的后端调用
--    - 按 requestStartTime 进行时间范围查询
-- 4. 索引说明：
--    - ORDER BY 已包含主要查询字段，ClickHouse会自动创建稀疏索引
--    - 无需额外创建二级索引，利用排序键即可实现高效查询
-- 5. 性能优化：
--    - 使用 requestStartTime 作为第一排序键，与分区键一致
--    - 查询时尽量包含时间范围条件，利用分区裁剪
--    - JOIN主表时使用 traceId，在排序键中靠前位置
-- ============================================================================