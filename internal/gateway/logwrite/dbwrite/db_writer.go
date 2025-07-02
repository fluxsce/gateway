package dbwrite

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"gohub/internal/gateway/logwrite/types"
	"gohub/pkg/database"
)

// DBWriter 实现了 LogWriter 接口，用于将网关访问日志写入数据库
// DBWriter 支持单条日志写入和批量日志写入，并确保线程安全
//
// 主要特性:
//   - 支持单条和批量日志写入
//   - 使用事务确保批量写入的原子性
//   - 线程安全的并发写入
//   - 预编译SQL语句提高性能
//   - 可配置的批量写入大小
type DBWriter struct {
	// 日志配置，包含批量写入大小等配置项
	config *types.LogConfig
	
	// 数据库连接实例，用于执行SQL操作
	db database.Database
	
	// SQL预编译语句
	// insertStmt 用于单条日志插入的SQL语句
	insertStmt string
	// batchStmt 用于批量日志插入的SQL语句
	batchStmt  string
	
	// 互斥锁，用于保证并发写入安全
	// 使用读写锁以支持并发读取但互斥写入
	mutex sync.RWMutex
}

// NewDBWriter 创建一个新的数据库日志写入器
// 
// 创建过程:
//   1. 获取默认数据库连接
//   2. 构建单条和批量插入SQL语句
//   3. 初始化DBWriter实例
// 
// 参数:
//   - config: 日志配置，包含批量写入大小等参数
//
// 返回:
//   - *DBWriter: 数据库日志写入器实例
//   - error: 如果创建过程中发生错误则返回错误信息
func NewDBWriter(config *types.LogConfig) (*DBWriter, error) {
	// 获取默认数据库连接
	db := database.GetDefaultConnection()
	if db == nil {
		return nil, fmt.Errorf("failed to get default database connection")
	}

	// 准备SQL语句
	insertStmt := buildInsertSQL()
	batchStmt := buildBatchInsertSQL(config.BatchSize)

	return &DBWriter{
		config:     config,
		db:         db,
		insertStmt: insertStmt,
		batchStmt:  batchStmt,
	}, nil
}

// Write 写入单条访问日志到数据库
// 
// 写入过程:
//   1. 获取读锁确保线程安全
//   2. 将日志对象转换为SQL参数
//   3. 执行单条插入SQL语句
// 
// 参数:
//   - ctx: 上下文，用于控制执行超时和取消
//   - log: 要写入的访问日志
//
// 返回:
//   - error: 如果写入过程中发生错误则返回错误信息
func (w *DBWriter) Write(ctx context.Context, log *types.AccessLog) error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// 将日志转换为SQL参数
	params := w.logToParams(log)

	// 执行插入操作
	_, err := w.db.Exec(ctx, w.insertStmt, params, true)
	if err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

// BatchWrite 批量写入多条访问日志到数据库
// 
// 批量写入特性:
//   - 使用事务确保批量写入的原子性
//   - 在事务失败时自动回滚
//   - 支持可配置的批量大小
//   - 优化的SQL语句减少网络往返
// 
// 参数:
//   - ctx: 上下文，用于控制执行超时和取消
//   - logs: 要写入的访问日志数组
//
// 返回:
//   - error: 如果写入过程中发生错误则返回错误信息
func (w *DBWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// 开启事务，确保批量写入的原子性
	if err := w.db.BeginTx(ctx, &database.TxOptions{
		Isolation: database.IsolationDefault,
		ReadOnly:  false,
	}); err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer w.db.Rollback()

	// 准备批量参数
	// 预分配足够容量的切片以避免动态扩容
	batchParams := make([]interface{}, 0, len(logs)*57) // 每条日志57个字段
	for _, log := range logs {
		params := w.logToParams(log)
		batchParams = append(batchParams, params...)
	}

	// 执行批量插入
	_, err := w.db.Exec(ctx, w.batchStmt, batchParams, false)
	if err != nil {
		return fmt.Errorf("failed to write log batch: %w", err)
	}

	// 提交事务
	if err := w.db.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Flush 实现 LogWriter 接口的 Flush 方法
// 由于数据库写入器直接写入数据库而不使用额外缓冲，此方法为空实现
// 
// 参数:
//   - ctx: 上下文，用于控制执行超时和取消
// 返回:
//   - error: 总是返回nil，因为无需刷新操作
func (w *DBWriter) Flush(ctx context.Context) error {
	return nil
}

// Close 实现 LogWriter 接口的 Close 方法
// 由于数据库连接由连接池管理，此方法为空实现
// 
// 返回:
//   - error: 总是返回nil，因为数据库连接由连接池管理
func (w *DBWriter) Close() error {
	return nil
}

// logToParams 将访问日志转换为SQL参数数组
// 
// 转换规则:
//   - 按照数据库表字段顺序排列
//   - 保持与SQL语句占位符顺序一致
//   - 处理可空字段的指针类型
// 
// 参数:
//   - log: 要转换的访问日志
//
// 返回:
//   - []interface{}: SQL参数数组，顺序与表字段顺序一致
func (w *DBWriter) logToParams(log *types.AccessLog) []interface{} {
	return []interface{}{
		log.TenantID,                       // 租户ID
		log.TraceID,                        // 追踪ID
		log.GatewayInstanceID,              // 网关实例ID
		log.GatewayNodeIP,                  // 网关节点IP
		log.RouteConfigID,                  // 路由配置ID
		log.ServiceDefinitionID,            // 服务定义ID
		log.LogConfigID,                    // 日志配置ID
		log.RequestMethod,                  // 请求方法
		log.RequestPath,                    // 请求路径
		log.RequestQuery,                   // 请求查询参数
		log.RequestSize,                    // 请求大小
		log.RequestHeaders,                 // 请求头
		log.RequestBody,                    // 请求体
		log.ClientIPAddress,                // 客户端IP地址
		log.ClientPort,                     // 客户端端口
		log.UserAgent,                      // 用户代理
		log.Referer,                        // 来源页面
		log.UserIdentifier,                 // 用户标识
		log.GatewayReceivedTime,            // 网关接收时间
		log.GatewayStartProcessingTime,     // 网关开始处理时间
		log.BackendRequestStartTime,        // 后端请求开始时间
		log.BackendResponseReceivedTime,    // 后端响应接收时间
		log.GatewayFinishedProcessingTime,  // 网关完成处理时间
		log.GatewayResponseSentTime,        // 网关响应发送时间
		log.TotalProcessingTimeMs,          // 总处理时间(毫秒)
		log.GatewayProcessingTimeMs,        // 网关处理时间(毫秒)
		log.BackendResponseTimeMs,          // 后端响应时间(毫秒)
		log.NetworkLatencyMs,               // 网络延迟(毫秒)
		log.GatewayStatusCode,              // 网关状态码
		log.BackendStatusCode,              // 后端状态码
		log.ResponseSize,                   // 响应大小
		log.ResponseHeaders,                // 响应头
		log.ResponseBody,                   // 响应体
		log.MatchedRoute,                   // 匹配的路由
		log.ForwardAddress,                 // 转发地址
		log.ForwardMethod,                  // 转发方法
		log.ForwardParams,                  // 转发参数
		log.ForwardHeaders,                 // 转发头
		log.ForwardBody,                    // 转发体
		log.LoadBalancerDecision,           // 负载均衡决策
		log.ErrorMessage,                   // 错误信息
		log.ErrorCode,                      // 错误码
		log.ParentTraceID,                  // 父追踪ID
		log.ResetFlag,                      // 重置标记
		log.RetryCount,                     // 重试次数
		log.ResetCount,                     // 重置次数
		log.LogLevel,                       // 日志级别
		log.LogType,                        // 日志类型
		log.ExtProperty,                    // 扩展属性
		log.AddTime,                        // 添加时间
		log.AddWho,                         // 添加人
		log.EditTime,                       // 修改时间
		log.EditWho,                        // 修改人
		log.OprSeqFlag,                     // 操作序列标记
		log.CurrentVersion,                 // 当前版本
		log.ActiveFlag,                     // 活动标记
		log.NoteText,                       // 备注文本
	}
}

// buildInsertSQL 构建单条插入SQL语句
// 
// SQL语句特点:
//   - 包含所有必需的字段
//   - 使用占位符参数避免SQL注入
//   - 字段顺序与logToParams方法保持一致
// 
// 返回:
//   - string: 完整的INSERT语句
func buildInsertSQL() string {
	return `
		INSERT INTO HUB_GW_ACCESS_LOG (
			TENANT_ID,                 -- 租户ID
			TRACE_ID,                  -- 追踪ID
			GATEWAY_INSTANCE_ID,       -- 网关实例ID
			GATEWAY_NODE_IP,           -- 网关节点IP
			ROUTE_CONFIG_ID,           -- 路由配置ID
			SERVICE_DEFINITION_ID,     -- 服务定义ID
			LOG_CONFIG_ID,             -- 日志配置ID
			REQUEST_METHOD,            -- 请求方法
			REQUEST_PATH,              -- 请求路径
			REQUEST_QUERY,             -- 请求查询参数
			REQUEST_SIZE,              -- 请求大小
			REQUEST_HEADERS,           -- 请求头
			REQUEST_BODY,              -- 请求体
			CLIENT_IP_ADDRESS,         -- 客户端IP地址
			CLIENT_PORT,               -- 客户端端口
			USER_AGENT,                -- 用户代理
			REFERER,                   -- 来源页面
			USER_IDENTIFIER,           -- 用户标识
			GATEWAY_RECEIVED_TIME,     -- 网关接收时间
			GATEWAY_START_PROCESSING_TIME, -- 网关开始处理时间
			BACKEND_REQUEST_START_TIME, -- 后端请求开始时间
			BACKEND_RESPONSE_RECEIVED_TIME, -- 后端响应接收时间
			GATEWAY_FINISHED_PROCESSING_TIME, -- 网关完成处理时间
			GATEWAY_RESPONSE_SENT_TIME, -- 网关响应发送时间
			TOTAL_PROCESSING_TIME_MS,  -- 总处理时间(毫秒)
			GATEWAY_PROCESSING_TIME_MS, -- 网关处理时间(毫秒)
			BACKEND_RESPONSE_TIME_MS,  -- 后端响应时间(毫秒)
			NETWORK_LATENCY_MS,        -- 网络延迟(毫秒)
			GATEWAY_STATUS_CODE,       -- 网关状态码
			BACKEND_STATUS_CODE,       -- 后端状态码
			RESPONSE_SIZE,             -- 响应大小
			RESPONSE_HEADERS,          -- 响应头
			RESPONSE_BODY,             -- 响应体
			MATCHED_ROUTE,             -- 匹配的路由
			FORWARD_ADDRESS,           -- 转发地址
			FORWARD_METHOD,            -- 转发方法
			FORWARD_PARAMS,            -- 转发参数
			FORWARD_HEADERS,           -- 转发头
			FORWARD_BODY,              -- 转发体
			LOAD_BALANCER_DECISION,    -- 负载均衡决策
			ERROR_MESSAGE,             -- 错误信息
			ERROR_CODE,                -- 错误码
			PARENT_TRACE_ID,           -- 父追踪ID
			RESET_FLAG,                -- 重置标记
			RETRY_COUNT,               -- 重试次数
			RESET_COUNT,               -- 重置次数
			LOG_LEVEL,                 -- 日志级别
			LOG_TYPE,                  -- 日志类型
			EXT_PROPERTY,              -- 扩展属性
			ADD_TIME,                  -- 添加时间
			ADD_WHO,                   -- 添加人
			EDIT_TIME,                 -- 修改时间
			EDIT_WHO,                  -- 修改人
			OPR_SEQ_FLAG,              -- 操作序列标记
			CURRENT_VERSION,           -- 当前版本
			ACTIVE_FLAG,               -- 活动标记
			NOTE_TEXT                  -- 备注文本
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?
		)
	`
}

// buildBatchInsertSQL 构建批量插入SQL语句
// 
// 批量插入优化:
//   - 单个SQL语句插入多行数据
//   - 减少网络往返次数
//   - 提高插入性能
//   - 支持可配置的批量大小
// 
// 参数:
//   - batchSize: 批量插入的记录数
//
// 返回:
//   - string: 完整的批量INSERT语句
func buildBatchInsertSQL(batchSize int) string {
	// 创建每行的占位符
	// 57个字段对应57个占位符
	valueStr := "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	values := make([]string, batchSize)
	for i := 0; i < batchSize; i++ {
		values[i] = valueStr
	}

	return fmt.Sprintf(`
		INSERT INTO HUB_GW_ACCESS_LOG (
			TENANT_ID,                 -- 租户ID
			TRACE_ID,                  -- 追踪ID
			GATEWAY_INSTANCE_ID,       -- 网关实例ID
			GATEWAY_NODE_IP,           -- 网关节点IP
			ROUTE_CONFIG_ID,           -- 路由配置ID
			SERVICE_DEFINITION_ID,     -- 服务定义ID
			LOG_CONFIG_ID,             -- 日志配置ID
			REQUEST_METHOD,            -- 请求方法
			REQUEST_PATH,              -- 请求路径
			REQUEST_QUERY,             -- 请求查询参数
			REQUEST_SIZE,              -- 请求大小
			REQUEST_HEADERS,           -- 请求头
			REQUEST_BODY,              -- 请求体
			CLIENT_IP_ADDRESS,         -- 客户端IP地址
			CLIENT_PORT,               -- 客户端端口
			USER_AGENT,                -- 用户代理
			REFERER,                   -- 来源页面
			USER_IDENTIFIER,           -- 用户标识
			GATEWAY_RECEIVED_TIME,     -- 网关接收时间
			GATEWAY_START_PROCESSING_TIME, -- 网关开始处理时间
			BACKEND_REQUEST_START_TIME, -- 后端请求开始时间
			BACKEND_RESPONSE_RECEIVED_TIME, -- 后端响应接收时间
			GATEWAY_FINISHED_PROCESSING_TIME, -- 网关完成处理时间
			GATEWAY_RESPONSE_SENT_TIME, -- 网关响应发送时间
			TOTAL_PROCESSING_TIME_MS,  -- 总处理时间(毫秒)
			GATEWAY_PROCESSING_TIME_MS, -- 网关处理时间(毫秒)
			BACKEND_RESPONSE_TIME_MS,  -- 后端响应时间(毫秒)
			NETWORK_LATENCY_MS,        -- 网络延迟(毫秒)
			GATEWAY_STATUS_CODE,       -- 网关状态码
			BACKEND_STATUS_CODE,       -- 后端状态码
			RESPONSE_SIZE,             -- 响应大小
			RESPONSE_HEADERS,          -- 响应头
			RESPONSE_BODY,             -- 响应体
			MATCHED_ROUTE,             -- 匹配的路由
			FORWARD_ADDRESS,           -- 转发地址
			FORWARD_METHOD,            -- 转发方法
			FORWARD_PARAMS,            -- 转发参数
			FORWARD_HEADERS,           -- 转发头
			FORWARD_BODY,              -- 转发体
			LOAD_BALANCER_DECISION,    -- 负载均衡决策
			ERROR_MESSAGE,             -- 错误信息
			ERROR_CODE,                -- 错误码
			PARENT_TRACE_ID,           -- 父追踪ID
			RESET_FLAG,                -- 重置标记
			RETRY_COUNT,               -- 重试次数
			RESET_COUNT,               -- 重置次数
			LOG_LEVEL,                 -- 日志级别
			LOG_TYPE,                  -- 日志类型
			EXT_PROPERTY,              -- 扩展属性
			ADD_TIME,                  -- 添加时间
			ADD_WHO,                   -- 添加人
			EDIT_TIME,                 -- 修改时间
			EDIT_WHO,                  -- 修改人
			OPR_SEQ_FLAG,              -- 操作序列标记
			CURRENT_VERSION,           -- 当前版本
			ACTIVE_FLAG,               -- 活动标记
			NOTE_TEXT                  -- 备注文本
		) VALUES %s
	`, strings.Join(values, ","))
} 