CREATE TABLE HUB_GW_BACKEND_TRACE_LOG (
                                   tenantId VARCHAR2(32) NOT NULL,              -- 租户ID
                                   traceId VARCHAR2(64) NOT NULL,               -- 链路追踪ID，关联主表 HUB_GW_ACCESS_LOG.traceId
                                   backendTraceId VARCHAR2(64) NOT NULL,        -- 后端服务追踪ID，同一traceId下唯一

                                   -- 服务信息（单个后端服务一次转发一条记录）
                                   serviceDefinitionId VARCHAR2(32),            -- 服务定义ID
                                   serviceName VARCHAR2(300),                   -- 服务名称（冗余字段，便于查询）

                                   -- 转发信息
                                   forwardAddress CLOB,                         -- 实际转发目标地址（完整URL）
                                   forwardMethod VARCHAR2(10),                  -- 转发HTTP方法
                                   forwardPath VARCHAR2(1000),                  -- 转发路径
                                   forwardQuery CLOB,                           -- 转发查询参数
                                   forwardHeaders CLOB,                         -- 转发请求头（JSON格式）
                                   forwardBody CLOB,                            -- 转发请求体
                                   requestSize NUMBER(10) DEFAULT 0,            -- 请求大小（字节）

                                   -- 负载均衡信息
                                   loadBalancerStrategy VARCHAR2(100),          -- 负载均衡策略
                                   loadBalancerDecision VARCHAR2(500),          -- 负载均衡决策信息

                                   -- 时间信息
                                   requestStartTime TIMESTAMP(3) NOT NULL,      -- 向后端发起请求时间
                                   responseReceivedTime TIMESTAMP(3),           -- 接收到后端响应时间
                                   requestDurationMs NUMBER(10),                -- 请求耗时（毫秒）

                                   -- 响应信息
                                   statusCode NUMBER(10),                       -- 后端HTTP状态码
                                   responseSize NUMBER(10) DEFAULT 0,           -- 响应大小（字节）
                                   responseHeaders CLOB,                        -- 响应头信息（JSON格式）
                                   responseBody CLOB,                           -- 响应体内容

                                   -- 错误信息
                                   errorCode VARCHAR2(100),                     -- 错误代码
                                   errorMessage CLOB,                           -- 详细错误信息

                                   -- 状态信息
                                   successFlag VARCHAR2(1) DEFAULT 'N' NOT NULL,-- 是否成功(Y成功,N失败)
                                   traceStatus VARCHAR2(20) DEFAULT 'pending' NOT NULL, -- 后端调用状态(pending,success,failed,timeout)
                                   retryCount NUMBER(10) DEFAULT 0 NOT NULL,    -- 重试次数

                                   -- 扩展信息
                                   extProperty CLOB,                            -- 扩展属性(JSON格式)

                                   -- 标准数据库字段
                                   addTime TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL, -- 记录创建时间
                                   addWho VARCHAR2(32) NOT NULL,                -- 记录创建者
                                   editTime TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL, -- 记录修改时间
                                   editWho VARCHAR2(32) NOT NULL,               -- 记录修改者
                                   oprSeqFlag VARCHAR2(32) NOT NULL,            -- 操作序列标识
                                   currentVersion NUMBER(10) DEFAULT 1 NOT NULL,-- 当前版本号
                                   activeFlag VARCHAR2(1) DEFAULT 'Y' NOT NULL, -- 活动状态标记
                                   noteText VARCHAR2(500),                      -- 备注信息

                                   CONSTRAINT pk_HUB_GW_BACKEND_TRACE_LOG PRIMARY KEY (traceId, backendTraceId)
);

COMMENT ON TABLE HUB_GW_BACKEND_TRACE_LOG IS '网关后端追踪日志表 - 记录每个后端服务的转发明细';

CREATE INDEX idx_gw_btrace_trace ON HUB_GW_BACKEND_TRACE_LOG (tenantId, traceId);
CREATE INDEX idx_gw_btrace_service ON HUB_GW_BACKEND_TRACE_LOG (tenantId, serviceDefinitionId, requestStartTime);
CREATE INDEX idx_gw_btrace_time ON HUB_GW_BACKEND_TRACE_LOG (requestStartTime);
CREATE INDEX idx_gw_btrace_tstatus ON HUB_GW_BACKEND_TRACE_LOG (tenantId, traceStatus, requestStartTime);
CREATE INDEX idx_gw_btrace_addtime ON HUB_GW_BACKEND_TRACE_LOG (tenantId, addTime);