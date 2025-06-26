package hub0003dao

import (
	"context"
	"gohub/pkg/database"
	hub0003models "gohub/web/views/hub0003/models"
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
	query := "UPDATE " + (&hub0003models.TimerExecutionLog{}).TableName() + " SET activeFlag = 'N', editWho = ?, editTime = NOW() " +
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
	
	// 计算总记录数
	countQuery := "SELECT COUNT(*) FROM " + (&hub0003models.TimerExecutionLog{}).TableName() + " " + whereClause
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count
	
	// 查询数据
	query := "SELECT * FROM " + (&hub0003models.TimerExecutionLog{}).TableName() + " " + whereClause +
		"ORDER BY executionStartTime DESC LIMIT ?, ?"
	args = append(args, (page-1)*pageSize, pageSize)
	
	err = dao.db.Query(ctx, &logs, query, args, true)
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
	
	// 计算总记录数
	countQuery := "SELECT COUNT(*) FROM " + (&hub0003models.TimerExecutionLog{}).TableName() + " " + whereClause
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count
	
	// 查询数据
	query := "SELECT * FROM " + (&hub0003models.TimerExecutionLog{}).TableName() + " " + whereClause +
		"ORDER BY executionStartTime DESC LIMIT ?, ?"
	args = append(args, (page-1)*pageSize, pageSize)
	
	err = dao.db.Query(ctx, &logs, query, args, true)
	if err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
} 