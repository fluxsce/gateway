package hub0003dao

import (
	"context"
	"gohub/pkg/database"
	hub0003models "gohub/web/views/hub0003/models"
	"time"
)

// TaskDao 任务数据访问对象
type TaskDao struct {
	db database.Database
}

// NewTaskDao 创建任务数据访问对象
func NewTaskDao(db database.Database) *TaskDao {
	return &TaskDao{db: db}
}

// Add 添加任务配置
func (dao *TaskDao) Add(ctx context.Context, task *hub0003models.TimerTask) (int64, error) {
	return dao.db.Insert(ctx, task.TableName(), task, true)
}

// GetById 根据ID获取任务配置
func (dao *TaskDao) GetById(ctx context.Context, tenantId, taskId string) (*hub0003models.TimerTask, error) {
	task := &hub0003models.TimerTask{}
	query := "SELECT * FROM " + task.TableName() + " WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'"
	err := dao.db.QueryOne(ctx, task, query, []interface{}{tenantId, taskId}, true)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// Update 更新任务配置
func (dao *TaskDao) Update(ctx context.Context, task *hub0003models.TimerTask) (int64, error) {
	query := "UPDATE " + task.TableName() + " SET taskName = ?, taskDescription = ?, taskPriority = ?, " +
		"schedulerId = ?, schedulerName = ?, scheduleType = ?, cronExpression = ?, " +
		"intervalSeconds = ?, delaySeconds = ?, startTime = ?, endTime = ?, " +
		"maxRetries = ?, retryIntervalSeconds = ?, timeoutSeconds = ?, taskParams = ?, " +
		"executorType = ?, toolConfigId = ?, toolConfigName = ?, operationType = ?, operationConfig = ?, " +
		"editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = currentVersion + 1, noteText = ? " +
		"WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'"
	
	args := []interface{}{
		task.TaskName, task.TaskDescription, task.TaskPriority,
		task.SchedulerId, task.SchedulerName, task.ScheduleType, task.CronExpression,
		task.IntervalSeconds, task.DelaySeconds, task.StartTime, task.EndTime,
		task.MaxRetries, task.RetryIntervalSeconds, task.TimeoutSeconds, task.TaskParams,
		task.ExecutorType, task.ToolConfigId, task.ToolConfigName, task.OperationType, task.OperationConfig,
		task.EditTime, task.EditWho, task.OprSeqFlag, task.NoteText,
		task.TenantId, task.TaskId,
	}
	
	return dao.db.Exec(ctx, query, args, true)
}

// Delete 删除任务配置（逻辑删除）
func (dao *TaskDao) Delete(ctx context.Context, tenantId, taskId, editWho string) (int64, error) {
	query := "UPDATE " + (&hub0003models.TimerTask{}).TableName() + " SET activeFlag = 'N', editWho = ?, editTime = NOW() " +
		"WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'"
	return dao.db.Exec(ctx, query, []interface{}{editWho, tenantId, taskId}, true)
}

// Query 查询任务配置列表
func (dao *TaskDao) Query(ctx context.Context, params map[string]interface{}, page, pageSize int) ([]*hub0003models.TimerTask, int64, error) {
	var tasks []*hub0003models.TimerTask
	
	// 构建查询条件
	whereClause := "WHERE activeFlag = 'Y' "
	args := []interface{}{}
	
	if tenantId, ok := params["tenantId"].(string); ok && tenantId != "" {
		whereClause += "AND tenantId = ? "
		args = append(args, tenantId)
	}
	
	if taskName, ok := params["taskName"].(string); ok && taskName != "" {
		whereClause += "AND taskName LIKE ? "
		args = append(args, "%"+taskName+"%")
	}
	
	if schedulerId, ok := params["schedulerId"].(string); ok && schedulerId != "" {
		whereClause += "AND schedulerId = ? "
		args = append(args, schedulerId)
	}
	
	if taskStatus, ok := params["taskStatus"].(int); ok && taskStatus > 0 {
		whereClause += "AND taskStatus = ? "
		args = append(args, taskStatus)
	}
	
	if scheduleType, ok := params["scheduleType"].(int); ok && scheduleType > 0 {
		whereClause += "AND scheduleType = ? "
		args = append(args, scheduleType)
	}
	
	if executorType, ok := params["executorType"].(string); ok && executorType != "" {
		whereClause += "AND executorType = ? "
		args = append(args, executorType)
	}
	
	if toolConfigId, ok := params["toolConfigId"].(string); ok && toolConfigId != "" {
		whereClause += "AND toolConfigId = ? "
		args = append(args, toolConfigId)
	}
	
	if operationType, ok := params["operationType"].(string); ok && operationType != "" {
		whereClause += "AND operationType = ? "
		args = append(args, operationType)
	}
	
	// 计算总记录数
	countQuery := "SELECT COUNT(*) FROM " + (&hub0003models.TimerTask{}).TableName() + " " + whereClause
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count
	
	// 查询数据
	query := "SELECT * FROM " + (&hub0003models.TimerTask{}).TableName() + " " + whereClause +
		"ORDER BY editTime DESC LIMIT ?, ?"
	args = append(args, (page-1)*pageSize, pageSize)
	
	err = dao.db.Query(ctx, &tasks, query, args, true)
	if err != nil {
		return nil, 0, err
	}
	
	return tasks, total, nil
}

// UpdateTaskStatus 更新任务状态
func (dao *TaskDao) UpdateTaskStatus(ctx context.Context, tenantId, taskId string, status int, editWho string) (int64, error) {
	query := "UPDATE " + (&hub0003models.TimerTask{}).TableName() + 
		" SET taskStatus = ?, editWho = ?, editTime = NOW(), currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'"
	
	return dao.db.Exec(ctx, query, []interface{}{status, editWho, tenantId, taskId}, true)
}

// UpdateNextRunTime 更新任务下次执行时间
func (dao *TaskDao) UpdateNextRunTime(ctx context.Context, tenantId, taskId string, nextRunTime time.Time) (int64, error) {
	query := "UPDATE " + (&hub0003models.TimerTask{}).TableName() + 
		" SET nextRunTime = ?, editTime = NOW(), currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'"
	
	return dao.db.Exec(ctx, query, []interface{}{nextRunTime, tenantId, taskId}, true)
}

// UpdateExecutionResult 更新任务执行结果
func (dao *TaskDao) UpdateExecutionResult(ctx context.Context, tenantId, taskId string, executionId string, 
	status int, success bool, errorMsg *string, startTime, endTime time.Time, durationMs int64, retryCount int) (int64, error) {
	
	// 计算成功或失败次数增量
	successIncrement := "successCount"
	failureIncrement := "failureCount"
	if success {
		successIncrement = "successCount + 1"
	} else {
		failureIncrement = "failureCount + 1"
	}
	
	// 构建成功标志
	var resultSuccess string
	if success {
		resultSuccess = "Y"
	} else {
		resultSuccess = "N"
	}
	
	query := "UPDATE " + (&hub0003models.TimerTask{}).TableName() + 
		" SET lastExecutionId = ?, lastExecutionStartTime = ?, lastExecutionEndTime = ?, " +
		"lastExecutionDurationMs = ?, lastExecutionStatus = ?, lastResultSuccess = ?, " +
		"lastErrorMessage = ?, lastRetryCount = ?, lastRunTime = ?, " +
		"runCount = runCount + 1, successCount = " + successIncrement + ", failureCount = " + failureIncrement + ", " +
		"editTime = NOW(), currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'"
	
	args := []interface{}{
		executionId, startTime, endTime,
		durationMs, status, resultSuccess,
		errorMsg, retryCount, startTime,
		tenantId, taskId,
	}
	
	return dao.db.Exec(ctx, query, args, true)
}

// GetByToolConfigId 根据工具配置ID获取任务列表
func (dao *TaskDao) GetByToolConfigId(ctx context.Context, tenantId, toolConfigId string) ([]*hub0003models.TimerTask, error) {
	var tasks []*hub0003models.TimerTask
	query := "SELECT * FROM " + (&hub0003models.TimerTask{}).TableName() + 
		" WHERE tenantId = ? AND toolConfigId = ? AND activeFlag = 'Y' ORDER BY editTime DESC"
	
	err := dao.db.Query(ctx, &tasks, query, []interface{}{tenantId, toolConfigId}, true)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetByExecutorType 根据执行器类型获取任务列表
func (dao *TaskDao) GetByExecutorType(ctx context.Context, tenantId, executorType string) ([]*hub0003models.TimerTask, error) {
	var tasks []*hub0003models.TimerTask
	query := "SELECT * FROM " + (&hub0003models.TimerTask{}).TableName() + 
		" WHERE tenantId = ? AND executorType = ? AND activeFlag = 'Y' ORDER BY editTime DESC"
	
	err := dao.db.Query(ctx, &tasks, query, []interface{}{tenantId, executorType}, true)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateExecutorConfig 更新任务执行器配置
func (dao *TaskDao) UpdateExecutorConfig(ctx context.Context, tenantId, taskId string, 
	executorType, toolConfigId, toolConfigName, operationType *string, operationConfig *string, editWho string) (int64, error) {
	
	query := "UPDATE " + (&hub0003models.TimerTask{}).TableName() + 
		" SET executorType = ?, toolConfigId = ?, toolConfigName = ?, operationType = ?, operationConfig = ?, " +
		"editWho = ?, editTime = NOW(), currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'"
	
	args := []interface{}{
		executorType, toolConfigId, toolConfigName, operationType, operationConfig,
		editWho, tenantId, taskId,
	}
	
	return dao.db.Exec(ctx, query, args, true)
} 