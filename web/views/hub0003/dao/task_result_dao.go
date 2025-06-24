package dao

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/utils/huberrors"
	"gohub/pkg/utils/random"
	"gohub/web/views/hub0003/models"
	"strings"
	"time"
)

// TaskResultDAO 定时任务执行结果数据访问对象
type TaskResultDAO struct {
	db database.Database
}

// NewTaskResultDAO 创建任务执行结果DAO
func NewTaskResultDAO(db database.Database) *TaskResultDAO {
	return &TaskResultDAO{
		db: db,
	}
}

// generateTaskResultId 生成任务结果ID
func (dao *TaskResultDAO) generateTaskResultId() string {
	now := time.Now()
	timeStr := now.Format("20060102150405")
	randomStr := random.GenerateRandomString(4)
	return fmt.Sprintf("RSLT%s%s", timeStr, randomStr)
}

// isTaskResultIdExists 检查任务结果ID是否已存在
func (dao *TaskResultDAO) isTaskResultIdExists(ctx context.Context, taskResultId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_TIMER_TASK_RESULT WHERE taskResultId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{taskResultId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueTaskResultId 生成唯一的任务结果ID
func (dao *TaskResultDAO) generateUniqueTaskResultId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		taskResultId := dao.generateTaskResultId()
		
		exists, err := dao.isTaskResultIdExists(ctx, taskResultId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查任务结果ID是否存在失败")
		}
		
		if !exists {
			return taskResultId, nil
		}
		
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一任务结果ID失败，已达到最大尝试次数")
}

// AddTaskResult 添加任务执行结果
func (dao *TaskResultDAO) AddTaskResult(ctx context.Context, result *models.TaskResult, operatorId string) error {
	if result.TenantId == "" || result.TaskId == "" || result.TaskConfigId == "" {
		return errors.New("tenantId、taskId和taskConfigId不能为空")
	}

	// 自动生成任务结果ID
	if result.TaskResultId == "" {
		generatedId, err := dao.generateUniqueTaskResultId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成任务结果ID失败")
		}
		result.TaskResultId = generatedId
	}

	// 设置自动填充字段
	now := time.Now()
	result.AddTime = now
	result.AddWho = operatorId
	result.EditTime = now
	result.EditWho = operatorId
	result.OprSeqFlag = result.TaskResultId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	result.CurrentVersion = 1
	result.ActiveFlag = "Y"

	// 设置默认值
	if result.ExecutionStatus == 0 {
		result.ExecutionStatus = 1 // 默认待执行
	}
	if result.ResultSuccess == "" {
		result.ResultSuccess = "N" // 默认未成功
	}

	_, err := dao.db.Insert(ctx, "HUB_TIMER_TASK_RESULT", result, true)
	if err != nil {
		return huberrors.WrapError(err, "添加任务执行结果失败")
	}

	return nil
}

// GetTaskResult 根据租户ID和任务结果ID获取任务结果
func (dao *TaskResultDAO) GetTaskResult(tenantId, taskResultId string) (*models.TaskResult, error) {
	if taskResultId == "" || tenantId == "" {
		return nil, errors.New("taskResultId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_RESULT 
		WHERE taskResultId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var result models.TaskResult
	err := dao.db.QueryOne(context.Background(), &result, query, []interface{}{taskResultId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询任务执行结果失败")
	}

	return &result, nil
}

// UpdateTaskResult 更新任务执行结果
func (dao *TaskResultDAO) UpdateTaskResult(ctx context.Context, result *models.TaskResult, operatorId string) error {
	if result.TaskResultId == "" || result.TenantId == "" {
		return errors.New("taskResultId和tenantId不能为空")
	}

	// 首先获取当前结果
	currentResult, err := dao.GetTaskResult(result.TenantId, result.TaskResultId)
	if err != nil {
		return err
	}
	if currentResult == nil {
		return errors.New("任务执行结果不存在")
	}

	// 更新修改信息
	result.EditTime = time.Now()
	result.EditWho = operatorId
	result.CurrentVersion = currentResult.CurrentVersion + 1
	result.OprSeqFlag = result.TaskResultId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	sql := `
		UPDATE HUB_TIMER_TASK_RESULT SET
			taskId = ?, taskConfigId = ?, executionStartTime = ?, executionEndTime = ?,
			executionDurationMs = ?, executionStatus = ?, resultSuccess = ?, errorMessage = ?,
			errorStackTrace = ?, retryCount = ?, maxRetryCount = ?, executionParams = ?,
			executionResult = ?, executorServerName = ?, executorServerIp = ?,
			reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			reserved6 = ?, reserved7 = ?, reserved8 = ?, reserved9 = ?, reserved10 = ?,
			noteText = ?, editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?
		WHERE taskResultId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, sql, []interface{}{
		result.TaskId, result.TaskConfigId, result.ExecutionStartTime, result.ExecutionEndTime,
		result.ExecutionDurationMs, result.ExecutionStatus, result.ResultSuccess, result.ErrorMessage,
		result.ErrorStackTrace, result.RetryCount, result.MaxRetryCount, result.ExecutionParams,
		result.ExecutionResult, result.ExecutorServerName, result.ExecutorServerIp,
		result.Reserved1, result.Reserved2, result.Reserved3, result.Reserved4, result.Reserved5,
		result.Reserved6, result.Reserved7, result.Reserved8, result.Reserved9, result.Reserved10,
		result.NoteText, result.EditTime, result.EditWho, result.CurrentVersion, result.OprSeqFlag,
		result.TaskResultId, result.TenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新任务执行结果失败")
	}

	return nil
}

// ListTaskResults 查询任务执行结果列表（分页）
func (dao *TaskResultDAO) ListTaskResults(ctx context.Context, tenantId string, page, pageSize int) ([]*models.TaskResult, int, error) {
	if tenantId == "" {
		return nil, 0, errors.New("tenantId不能为空")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询总数
	countSQL := `
		SELECT COUNT(*) as total 
		FROM HUB_TIMER_TASK_RESULT 
		WHERE tenantId = ? AND activeFlag = 'Y'
	`

	var totalCount struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalCount, countSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询任务执行结果总数失败")
	}

	// 查询数据
	offset := (page - 1) * pageSize
	dataSQL := `
		SELECT * FROM HUB_TIMER_TASK_RESULT 
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY executionStartTime DESC 
		LIMIT ? OFFSET ?
	`

	var results []*models.TaskResult
	err = dao.db.Query(ctx, &results, dataSQL, []interface{}{tenantId, pageSize, offset}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询任务执行结果列表失败")
	}

	return results, totalCount.Total, nil
}

// GetTaskResultsByTaskId 根据任务ID查询执行结果
func (dao *TaskResultDAO) GetTaskResultsByTaskId(ctx context.Context, tenantId, taskId string, page, pageSize int) ([]*models.TaskResult, int, error) {
	if tenantId == "" || taskId == "" {
		return nil, 0, errors.New("tenantId和taskId不能为空")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询总数
	countSQL := `
		SELECT COUNT(*) as total 
		FROM HUB_TIMER_TASK_RESULT 
		WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'
	`

	var totalCount struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalCount, countSQL, []interface{}{tenantId, taskId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询任务执行结果总数失败")
	}

	// 查询数据
	offset := (page - 1) * pageSize
	dataSQL := `
		SELECT * FROM HUB_TIMER_TASK_RESULT 
		WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'
		ORDER BY executionStartTime DESC 
		LIMIT ? OFFSET ?
	`

	var results []*models.TaskResult
	err = dao.db.Query(ctx, &results, dataSQL, []interface{}{tenantId, taskId, pageSize, offset}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询任务执行结果列表失败")
	}

	return results, totalCount.Total, nil
}

// GetTaskResultsByStatus 根据执行状态查询任务结果
func (dao *TaskResultDAO) GetTaskResultsByStatus(ctx context.Context, tenantId string, status int) ([]*models.TaskResult, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_RESULT 
		WHERE tenantId = ? AND executionStatus = ? AND activeFlag = 'Y'
		ORDER BY executionStartTime DESC
	`

	var results []*models.TaskResult
	err := dao.db.Query(ctx, &results, query, []interface{}{tenantId, status}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "根据状态查询任务执行结果失败")
	}

	return results, nil
}

// GetLatestTaskResultByTaskId 获取任务的最新执行结果
func (dao *TaskResultDAO) GetLatestTaskResultByTaskId(tenantId, taskId string) (*models.TaskResult, error) {
	if tenantId == "" || taskId == "" {
		return nil, errors.New("tenantId和taskId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_RESULT 
		WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'
		ORDER BY executionStartTime DESC 
		LIMIT 1
	`

	var result models.TaskResult
	err := dao.db.QueryOne(context.Background(), &result, query, []interface{}{tenantId, taskId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询任务最新执行结果失败")
	}

	return &result, nil
}

// QueryTaskResults 查询任务执行结果列表
func (dao *TaskResultDAO) QueryTaskResults(ctx context.Context, tenantId, taskId, status, startTime, endTime string, pageNum, pageSize int) (int64, []*models.TaskResult, error) {
	if tenantId == "" {
		return 0, nil, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	conditions := []string{"tenantId = ? AND activeFlag = 'Y'"}
	args := []interface{}{tenantId}

	if taskId != "" {
		conditions = append(conditions, "taskId = ?")
		args = append(args, taskId)
	}

	if status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}

	if startTime != "" {
		conditions = append(conditions, "startTime >= ?")
		args = append(args, startTime)
	}

	if endTime != "" {
		conditions = append(conditions, "endTime <= ?")
		args = append(args, endTime)
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as count FROM HUB_TIMER_TASK_RESULT WHERE %s", whereClause)
	var result struct {
		Count int64 `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return 0, nil, huberrors.WrapError(err, "查询任务执行结果总数失败")
	}

	// 如果没有数据，直接返回
	if result.Count == 0 {
		return 0, make([]*models.TaskResult, 0), nil
	}

	// 查询分页数据
	offset := (pageNum - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT * FROM HUB_TIMER_TASK_RESULT 
		WHERE %s 
		ORDER BY startTime DESC 
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pageSize, offset)
	var results []*models.TaskResult
	err = dao.db.Query(ctx, &results, query, args, true)
	if err != nil {
		return 0, nil, huberrors.WrapError(err, "查询任务执行结果列表失败")
	}

	return result.Count, results, nil
}

// GetLatestTaskResult 获取任务最新执行结果
func (dao *TaskResultDAO) GetLatestTaskResult(tenantId, taskId string) (*models.TaskResult, error) {
	if taskId == "" || tenantId == "" {
		return nil, errors.New("taskId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_RESULT 
		WHERE tenantId = ? AND taskId = ? AND activeFlag = 'Y'
		ORDER BY startTime DESC 
		LIMIT 1
	`

	var result models.TaskResult
	err := dao.db.QueryOne(context.Background(), &result, query, []interface{}{tenantId, taskId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "获取任务最新执行结果失败")
	}

	return &result, nil
}

// UpdateTaskResultStatus 更新任务执行结果状态
func (dao *TaskResultDAO) UpdateTaskResultStatus(ctx context.Context, tenantId, taskResultId, status, message, operatorId string) error {
	if taskResultId == "" || tenantId == "" || status == "" {
		return errors.New("taskResultId、tenantId和status不能为空")
	}

	now := time.Now()
	oprSeqFlag := taskResultId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]

	sql := `
		UPDATE HUB_TIMER_TASK_RESULT SET
			status = ?, message = ?, endTime = ?, editTime = ?, editWho = ?, 
			currentVersion = currentVersion + 1, oprSeqFlag = ?
		WHERE taskResultId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		status, message, now, now, operatorId, oprSeqFlag, taskResultId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新任务执行结果状态失败")
	}

	return nil
} 