--   1. 作为 HUB_GW_ACCESS_LOG 的从表，记录每个后端服务转发的详细信息
--   2. traceId + backendTraceId 作为联合主键
CREATE TABLE IF NOT EXISTS HUB_GW_BACKEND_TRACE_LOG (
    tenantId TEXT NOT NULL,                 -- 租户ID
    traceId TEXT NOT NULL,                  -- 链路追踪ID，关联主表 HUB_GW_ACCESS_LOG.traceId
    backendTraceId TEXT NOT NULL,           -- 后端服务追踪ID，同一traceId下唯一

    -- 服务信息（单个后端服务一次转发一条记录）
    serviceDefinitionId TEXT,               -- 服务定义ID
    serviceName TEXT,                       -- 服务名称（冗余字段）

    -- 转发信息
    forwardAddress TEXT,                    -- 实际转发目标地址(完整URL)
    forwardMethod TEXT,                     -- 转发HTTP方法
    forwardPath TEXT,                       -- 转发路径
    forwardQuery TEXT,                      -- 转发查询参数
    forwardHeaders TEXT,                    -- 转发请求头(JSON格式)
    forwardBody TEXT,                       -- 转发请求体
    requestSize INTEGER DEFAULT 0,          -- 请求大小(字节)

    -- 负载均衡信息
    loadBalancerStrategy TEXT,             -- 负载均衡策略
    loadBalancerDecision TEXT,             -- 负载均衡决策信息

    -- 时间信息
    requestStartTime DATETIME NOT NULL,    -- 向后端发起请求时间
    responseReceivedTime DATETIME,         -- 接收到后端响应时间
    requestDurationMs INTEGER,             -- 请求耗时(毫秒)

    -- 响应信息
    statusCode INTEGER,                    -- 后端HTTP状态码
    responseSize INTEGER DEFAULT 0,        -- 响应大小(字节)
    responseHeaders TEXT,                  -- 响应头信息(JSON格式)
    responseBody TEXT,                     -- 响应体内容

    -- 错误信息
    errorCode TEXT,                        -- 错误代码
    errorMessage TEXT,                     -- 详细错误信息

    -- 状态信息
    successFlag TEXT NOT NULL DEFAULT 'N', -- 是否成功(Y成功,N失败)
    traceStatus TEXT NOT NULL DEFAULT 'pending',-- 后端调用状态(pending,success,failed,timeout)
    retryCount INTEGER NOT NULL DEFAULT 0, -- 重试次数

    -- 扩展信息
    extProperty TEXT,                      -- 扩展属性(JSON格式)

    -- 标准数据库字段
    addTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    addWho TEXT NOT NULL,
    editTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    editWho TEXT NOT NULL,
    oprSeqFlag TEXT NOT NULL,
    currentVersion INTEGER NOT NULL DEFAULT 1,
    activeFlag TEXT NOT NULL DEFAULT 'Y',
    noteText TEXT,

    PRIMARY KEY (traceId, backendTraceId)
);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_TRACE ON HUB_GW_BACKEND_TRACE_LOG(tenantId, traceId);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_SERVICE ON HUB_GW_BACKEND_TRACE_LOG(tenantId, serviceDefinitionId, requestStartTime);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_TIME ON HUB_GW_BACKEND_TRACE_LOG(requestStartTime);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_TSTATUS ON HUB_GW_BACKEND_TRACE_LOG(tenantId, traceStatus, requestStartTime);
CREATE INDEX IF NOT EXISTS IDX_GW_BTRACE_ADDTIME ON HUB_GW_BACKEND_TRACE_LOG(tenantId, addTime);

-- 索引设计：参考数据库规范，兼顾多租户与常用查询场景