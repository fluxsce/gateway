package hub0003dao

import (
	"context"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/utils/ctime"
	hub0003models "gateway/web/views/hub0003/models"
)

// ExecutionLogDao 执行日志数据访问对象
type ExecutionLogDao struct {
	db database.Database
}

// NewExecutionLogDao 创建执行日志数据访问对象
func NewExecutionLogDao(db database.Database) *ExecutionLogDao {
	return &ExecutionLogDao{db: db}
}

// Add 添加执行日志
func (dao *ExecutionLogDao) Add(ctx context.Context, log *hub0003models.TimerExecutionLog) (int64, error) {
	return dao.db.Insert(ctx, log.TableName(), log, true)
}

// GetById 根据ID获取执行日志
func (dao *ExecutionLogDao) GetById(ctx context.Context, tenantId, executionId string) (*hub0003models.TimerExecutionLog, error) {
	log := &hub0003models.TimerExecutionLog{}
	query := "SELECT * FROM " + log.TableName() + " WHERE tenantId = ? AND executionId = ? AND activeFlag = 'Y'"
	err := dao.db.QueryOne(ctx, log, query, []interface{}{tenantId, executionId}, true)
	if err != nil {
		return nil, err
	}
	return log, nil
}

// Update 更新执行日志
func (dao *ExecutionLogDao) Update(ctx context.Context, log *hub0003models.TimerExecutionLog) (int64, error) {
	query := "UPDATE " + log.TableName() + " SET executionEndTime = ?, executionDurationMs = ?, " +
		"executionStatus = ?, resultSuccess = ?, errorMessage = ?, errorStackTrace = ?, " +
		"retryCount = ?, executionResult = ?, logLevel = ?, logMessage = ?, " +
		"logTimestamp = ?, executionPhase = ?, exceptionClass = ?, exceptionMessage = ?, " +
		"editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND executionId = ? AND activeFlag = 'Y'"

	args := []interface{}{
		log.ExecutionEndTime, log.ExecutionDurationMs,
		log.ExecutionStatus, log.ResultSuccess, log.ErrorMessage, log.ErrorStackTrace,
		log.RetryCount, log.ExecutionResult, log.LogLevel, log.LogMessage,
		log.LogTimestamp, log.ExecutionPhase, log.ExceptionClass, log.ExceptionMessage,
		log.EditTime, log.EditWho, log.OprSeqFlag,
		log.TenantId, log.ExecutionId,
	}

	return dao.db.Exec(ctx, query, args, true)
}

// Delete 删除执行日志（逻辑删除）
func (dao *ExecutionLogDao) Delete(ctx context.Context, tenantId, executionId, editWho string) (int64, error) {
	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 获取当前时间函数
	timeFunc, err := sqlutils.GetCurrentTimeFunction(dbType)
	if err != nil {
		return 0, err
	}

	query := "UPDATE " + (&hub0003models.TimerExecutionLog{}).TableName() +
		" SET activeFlag = 'N', editWho = ?, editTime = " + timeFunc + " " +
		"WHERE tenantId = ? AND executionId = ? AND activeFlag = 'Y'"
	return dao.db.Exec(ctx, query, []interface{}{editWho, tenantId, executionId}, true)
}

// Query 查询执行日志列表
func (dao *ExecutionLogDao) Query(ctx context.Context, params map[string]interface{}, page, pageSize int) ([]*hub0003models.TimerExecutionLog, int64, error) {
	var logs []*hub0003models.TimerExecutionLog

	// 构建查询条件
	whereClause := "WHERE activeFlag = 'Y' "
	args := []interface{}{}

	if tenantId, ok := params["tenantId"].(string); ok && tenantId != "" {
		whereClause += "AND tenantId = ? "
		args = append(args, tenantId)
	}

	if taskId, ok := params["taskId"].(string); ok && taskId != "" {
		whereClause += "AND taskId = ? "
		args = append(args, taskId)
	}

	if taskName, ok := params["taskName"].(string); ok && taskName != "" {
		whereClause += "AND taskName LIKE ? "
		args = append(args, "%"+taskName+"%")
	}

	if schedulerId, ok := params["schedulerId"].(string); ok && schedulerId != "" {
		whereClause += "AND schedulerId = ? "
		args = append(args, schedulerId)
	}

	if executionStatus, ok := params["executionStatus"].(int); ok && executionStatus > 0 {
		whereClause += "AND executionStatus = ? "
		args = append(args, executionStatus)
	}

	if resultSuccess, ok := params["resultSuccess"].(string); ok && resultSuccess != "" {
		whereClause += "AND resultSuccess = ? "
		args = append(args, resultSuccess)
	}

	if logLevel, ok := params["logLevel"].(string); ok && logLevel != "" {
		whereClause += "AND logLevel = ? "
		args = append(args, logLevel)
	}

	// 处理执行开始时间范围查询
	if startTime, ok := params["startTime"].(string); ok && startTime != "" {
		// 使用自定义时间包解析时间字符串为time.Time类型
		if parsedStartTime, err := ctime.ParseTimeString(startTime); err == nil {
			whereClause += "AND executionStartTime >= ? "
			args = append(args, parsedStartTime)
		}
	}

	if endTime, ok := params["endTime"].(string); ok && endTime != "" {
		// 使用自定义时间包解析时间字符串为time.Time类型
		if parsedEndTime, err := ctime.ParseTimeString(endTime); err == nil {
			whereClause += "AND executionStartTime <= ? "
			args = append(args, parsedEndTime)
		}
	}

	// 构建基础查询语句（用于COUNT查询）
	baseQuery := "SELECT * FROM " + (&hub0003models.TimerExecutionLog{}).TableName() + " " + whereClause

	// 使用公共方法构建COUNT查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询总记录数
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count

	// 构建数据查询的基础语句（包含ORDER BY）
	dataBaseQuery := "SELECT * FROM " + (&hub0003models.TimerExecutionLog{}).TableName() + " " + whereClause + " ORDER BY executionStartTime DESC"

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 使用公共方法构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, dataBaseQuery, pagination)
	if err != nil {
		return nil, 0, err
	}

	// 合并查询参数和分页参数
	finalArgs := append(args, paginationArgs...)

	// 执行分页查询
	err = dao.db.Query(ctx, &logs, paginatedQuery, finalArgs, true)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetTaskLogs 获取任务相关的执行日志
func (dao *ExecutionLogDao) GetTaskLogs(ctx context.Context, tenantId, taskId string, page, pageSize int) ([]*hub0003models.TimerExecutionLog, int64, error) {
	var logs []*hub0003models.TimerExecutionLog

	// 构建查询条件
	whereClause := "WHERE activeFlag = 'Y' AND tenantId = ? AND taskId = ? "
	args := []interface{}{tenantId, taskId}

	// 构建基础查询语句（用于COUNT查询）
	baseQuery := "SELECT * FROM " + (&hub0003models.TimerExecutionLog{}).TableName() + " " + whereClause

	// 使用公共方法构建COUNT查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询总记录数
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count

	// 构建数据查询的基础语句（包含ORDER BY）
	dataBaseQuery := "SELECT * FROM " + (&hub0003models.TimerExecutionLog{}).TableName() + " " + whereClause + " ORDER BY executionStartTime DESC"

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(page, pageSize)

	// 使用公共方法构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, dataBaseQuery, pagination)
	if err != nil {
		return nil, 0, err
	}

	// 合并查询参数和分页参数
	finalArgs := append(args, paginationArgs...)

	// 执行分页查询
	err = dao.db.Query(ctx, &logs, paginatedQuery, finalArgs, true)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
