package hub0003dao

import (
	"context"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	hub0003models "gateway/web/views/hub0003/models"
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
	query := "SELECT * FROM " + task.TableName() + " WHERE tenantId = ? AND taskId = ?"
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
		"taskStatus = ?, activeFlag = ?, " +
		"editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = currentVersion + 1, noteText = ? " +
		"WHERE tenantId = ? AND taskId = ?"

	args := []interface{}{
		task.TaskName, task.TaskDescription, task.TaskPriority,
		task.SchedulerId, task.SchedulerName, task.ScheduleType, task.CronExpression,
		task.IntervalSeconds, task.DelaySeconds, task.StartTime, task.EndTime,
		task.MaxRetries, task.RetryIntervalSeconds, task.TimeoutSeconds, task.TaskParams,
		task.ExecutorType, task.ToolConfigId, task.ToolConfigName, task.OperationType, task.OperationConfig,
		task.TaskStatus, task.ActiveFlag,
		task.EditTime, task.EditWho, task.OprSeqFlag, task.NoteText,
		task.TenantId, task.TaskId,
	}

	return dao.db.Exec(ctx, query, args, true)
}

// Delete 删除任务配置（物理删除）
func (dao *TaskDao) Delete(ctx context.Context, tenantId, taskId, editWho string) (int64, error) {
	query := "DELETE FROM " + (&hub0003models.TimerTask{}).TableName() +
		" WHERE tenantId = ? AND taskId = ?"
	return dao.db.Exec(ctx, query, []interface{}{tenantId, taskId}, true)
}

// Query 查询任务配置列表
func (dao *TaskDao) Query(ctx context.Context, params map[string]interface{}, page, pageSize int) ([]*hub0003models.TimerTask, int64, error) {
	var tasks []*hub0003models.TimerTask

	// 构建查询条件
	whereClause := "WHERE 1=1 "
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

	// 支持活动状态过滤（可选）
	if activeFlag, ok := params["activeFlag"].(string); ok && activeFlag != "" {
		whereClause += "AND activeFlag = ? "
		args = append(args, activeFlag)
	}

	// 构建基础查询语句（用于COUNT查询）
	baseQuery := "SELECT * FROM " + (&hub0003models.TimerTask{}).TableName() + " " + whereClause

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
	dataBaseQuery := "SELECT * FROM " + (&hub0003models.TimerTask{}).TableName() + " " + whereClause + " ORDER BY editTime DESC"

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
	err = dao.db.Query(ctx, &tasks, paginatedQuery, finalArgs, true)
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// UpdateTaskStatus 更新任务状态
func (dao *TaskDao) UpdateTaskStatus(ctx context.Context, tenantId, taskId string, status int, editWho string) (int64, error) {
	query := "UPDATE " + (&hub0003models.TimerTask{}).TableName() +
		" SET taskStatus = ?, editWho = ?, editTime = NOW(), currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND taskId = ?"

	return dao.db.Exec(ctx, query, []interface{}{status, editWho, tenantId, taskId}, true)
}

// UpdateNextRunTime 更新任务下次执行时间
func (dao *TaskDao) UpdateNextRunTime(ctx context.Context, tenantId, taskId string, nextRunTime time.Time) (int64, error) {
	query := "UPDATE " + (&hub0003models.TimerTask{}).TableName() +
		" SET nextRunTime = ?, editTime = NOW(), currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND taskId = ?"

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
		"WHERE tenantId = ? AND taskId = ?"

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
		" WHERE tenantId = ? AND toolConfigId = ? ORDER BY editTime DESC"

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
		" WHERE tenantId = ? AND executorType = ? ORDER BY editTime DESC"

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
		"WHERE tenantId = ? AND taskId = ?"

	args := []interface{}{
		executorType, toolConfigId, toolConfigName, operationType, operationConfig,
		editWho, tenantId, taskId,
	}

	return dao.db.Exec(ctx, query, args, true)
}
