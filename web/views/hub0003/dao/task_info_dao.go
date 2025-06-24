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

// TaskInfoDAO 定时任务运行时信息数据访问对象
type TaskInfoDAO struct {
	db database.Database
}

// NewTaskInfoDAO 创建任务运行时信息DAO
func NewTaskInfoDAO(db database.Database) *TaskInfoDAO {
	return &TaskInfoDAO{
		db: db,
	}
}

// generateTaskInfoId 生成任务信息ID
func (dao *TaskInfoDAO) generateTaskInfoId() string {
	now := time.Now()
	timeStr := now.Format("20060102150405")
	randomStr := random.GenerateRandomString(4)
	return fmt.Sprintf("INFO%s%s", timeStr, randomStr)
}

// isTaskInfoIdExists 检查任务信息ID是否已存在
func (dao *TaskInfoDAO) isTaskInfoIdExists(ctx context.Context, taskInfoId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_TIMER_TASK_INFO WHERE taskInfoId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{taskInfoId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueTaskInfoId 生成唯一的任务信息ID
func (dao *TaskInfoDAO) generateUniqueTaskInfoId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		taskInfoId := dao.generateTaskInfoId()
		
		exists, err := dao.isTaskInfoIdExists(ctx, taskInfoId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查任务信息ID是否存在失败")
		}
		
		if !exists {
			return taskInfoId, nil
		}
		
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一任务信息ID失败，已达到最大尝试次数")
}

// AddTaskInfo 添加任务运行时信息
func (dao *TaskInfoDAO) AddTaskInfo(ctx context.Context, info *models.TaskInfo, operatorId string) error {
	if info.TenantId == "" || info.TaskConfigId == "" || info.TaskId == "" {
		return errors.New("tenantId、taskConfigId和taskId不能为空")
	}

	// 自动生成任务信息ID
	if info.TaskInfoId == "" {
		generatedId, err := dao.generateUniqueTaskInfoId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成任务信息ID失败")
		}
		info.TaskInfoId = generatedId
	}

	// 设置自动填充字段
	now := time.Now()
	info.AddTime = now
	info.AddWho = operatorId
	info.EditTime = now
	info.EditWho = operatorId
	info.OprSeqFlag = info.TaskInfoId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	info.CurrentVersion = 1
	info.ActiveFlag = "Y"

	// 设置默认值
	if info.TaskStatus == 0 {
		info.TaskStatus = 1 // 默认待执行
	}

	_, err := dao.db.Insert(ctx, "HUB_TIMER_TASK_INFO", info, true)
	if err != nil {
		return huberrors.WrapError(err, "添加任务运行时信息失败")
	}

	return nil
}

// GetTaskInfo 根据租户ID和任务信息ID获取任务信息
func (dao *TaskInfoDAO) GetTaskInfo(tenantId, taskInfoId string) (*models.TaskInfo, error) {
	if taskInfoId == "" || tenantId == "" {
		return nil, errors.New("taskInfoId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_INFO 
		WHERE taskInfoId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var info models.TaskInfo
	err := dao.db.QueryOne(context.Background(), &info, query, []interface{}{taskInfoId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询任务运行时信息失败")
	}

	return &info, nil
}

// GetTaskInfoByTaskId 根据任务ID获取任务运行时信息
func (dao *TaskInfoDAO) GetTaskInfoByTaskId(tenantId, taskId string) (*models.TaskInfo, error) {
	if taskId == "" || tenantId == "" {
		return nil, errors.New("taskId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_INFO 
		WHERE taskId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var info models.TaskInfo
	err := dao.db.QueryOne(context.Background(), &info, query, []interface{}{taskId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询任务运行时信息失败")
	}

	return &info, nil
}

// UpdateTaskInfo 更新任务运行时信息
func (dao *TaskInfoDAO) UpdateTaskInfo(ctx context.Context, info *models.TaskInfo, operatorId string) error {
	if info.TaskInfoId == "" || info.TenantId == "" {
		return errors.New("taskInfoId和tenantId不能为空")
	}

	// 首先获取当前信息
	currentInfo, err := dao.GetTaskInfo(info.TenantId, info.TaskInfoId)
	if err != nil {
		return err
	}
	if currentInfo == nil {
		return errors.New("任务运行时信息不存在")
	}

	// 更新修改信息
	info.EditTime = time.Now()
	info.EditWho = operatorId
	info.CurrentVersion = currentInfo.CurrentVersion + 1
	info.OprSeqFlag = info.TaskInfoId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	sql := `
		UPDATE HUB_TIMER_TASK_INFO SET
			taskConfigId = ?, taskId = ?, taskStatus = ?, nextRunTime = ?, lastRunTime = ?,
			runCount = ?, successCount = ?, failureCount = ?, lastResultId = ?,
			lastExecutionDurationMs = ?, lastErrorMessage = ?,
			reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			reserved6 = ?, reserved7 = ?, reserved8 = ?, reserved9 = ?, reserved10 = ?,
			noteText = ?, editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?
		WHERE taskInfoId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, sql, []interface{}{
		info.TaskConfigId, info.TaskId, info.TaskStatus, info.NextRunTime, info.LastRunTime,
		info.RunCount, info.SuccessCount, info.FailureCount, info.LastResultId,
		info.LastExecutionDurationMs, info.LastErrorMessage,
		info.Reserved1, info.Reserved2, info.Reserved3, info.Reserved4, info.Reserved5,
		info.Reserved6, info.Reserved7, info.Reserved8, info.Reserved9, info.Reserved10,
		info.NoteText, info.EditTime, info.EditWho, info.CurrentVersion, info.OprSeqFlag,
		info.TaskInfoId, info.TenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新任务运行时信息失败")
	}

	return nil
}

// UpdateTaskStatus 更新任务状态
func (dao *TaskInfoDAO) UpdateTaskStatus(ctx context.Context, tenantId, taskId string, status int, operatorId string) error {
	if tenantId == "" || taskId == "" {
		return errors.New("tenantId和taskId不能为空")
	}

	sql := `
		UPDATE HUB_TIMER_TASK_INFO SET
			taskStatus = ?, editTime = ?, editWho = ?, currentVersion = currentVersion + 1,
			oprSeqFlag = ?
		WHERE taskId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	now := time.Now()
	oprSeqFlag := taskId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		status, now, operatorId, oprSeqFlag, taskId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新任务状态失败")
	}

	return nil
}

// UpdateTaskExecutionStats 更新任务执行统计信息
func (dao *TaskInfoDAO) UpdateTaskExecutionStats(ctx context.Context, tenantId, taskId string, 
	success bool, executionDurationMs int64, errorMessage string, operatorId string) error {
	
	if tenantId == "" || taskId == "" {
		return errors.New("tenantId和taskId不能为空")
	}

	var sql string
	var args []interface{}
	now := time.Now()
	oprSeqFlag := taskId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]

	if success {
		sql = `
			UPDATE HUB_TIMER_TASK_INFO SET
				runCount = runCount + 1, successCount = successCount + 1,
				lastRunTime = ?, lastExecutionDurationMs = ?, lastErrorMessage = NULL,
				editTime = ?, editWho = ?, currentVersion = currentVersion + 1, oprSeqFlag = ?
			WHERE taskId = ? AND tenantId = ? AND activeFlag = 'Y'
		`
		args = []interface{}{now, executionDurationMs, now, operatorId, oprSeqFlag, taskId, tenantId}
	} else {
		sql = `
			UPDATE HUB_TIMER_TASK_INFO SET
				runCount = runCount + 1, failureCount = failureCount + 1,
				lastRunTime = ?, lastExecutionDurationMs = ?, lastErrorMessage = ?,
				editTime = ?, editWho = ?, currentVersion = currentVersion + 1, oprSeqFlag = ?
			WHERE taskId = ? AND tenantId = ? AND activeFlag = 'Y'
		`
		args = []interface{}{now, executionDurationMs, errorMessage, now, operatorId, oprSeqFlag, taskId, tenantId}
	}

	_, err := dao.db.Exec(ctx, sql, args, true)
	if err != nil {
		return huberrors.WrapError(err, "更新任务执行统计信息失败")
	}

	return nil
}

// ListTaskInfos 查询任务运行时信息列表（分页）
func (dao *TaskInfoDAO) ListTaskInfos(ctx context.Context, tenantId string, page, pageSize int) ([]*models.TaskInfo, int, error) {
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
		FROM HUB_TIMER_TASK_INFO 
		WHERE tenantId = ? AND activeFlag = 'Y'
	`

	var totalCount struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalCount, countSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询任务运行时信息总数失败")
	}

	// 查询数据
	offset := (page - 1) * pageSize
	dataSQL := `
		SELECT * FROM HUB_TIMER_TASK_INFO 
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY lastRunTime DESC, addTime DESC 
		LIMIT ? OFFSET ?
	`

	var infos []*models.TaskInfo
	err = dao.db.Query(ctx, &infos, dataSQL, []interface{}{tenantId, pageSize, offset}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询任务运行时信息列表失败")
	}

	return infos, totalCount.Total, nil
}

// GetTaskInfosByStatus 根据任务状态查询任务信息
func (dao *TaskInfoDAO) GetTaskInfosByStatus(ctx context.Context, tenantId string, status int) ([]*models.TaskInfo, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_INFO 
		WHERE tenantId = ? AND taskStatus = ? AND activeFlag = 'Y'
		ORDER BY nextRunTime ASC, addTime ASC
	`

	var infos []*models.TaskInfo
	err := dao.db.Query(ctx, &infos, query, []interface{}{tenantId, status}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "根据状态查询任务运行时信息失败")
	}

	return infos, nil
}

// DeleteTaskInfo 删除任务运行时信息（软删除）
func (dao *TaskInfoDAO) DeleteTaskInfo(ctx context.Context, tenantId, taskInfoId, operatorId string) error {
	if taskInfoId == "" || tenantId == "" {
		return errors.New("taskInfoId和tenantId不能为空")
	}

	now := time.Now()
	oprSeqFlag := taskInfoId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]

	sql := `
		UPDATE HUB_TIMER_TASK_INFO SET
			activeFlag = 'N', editTime = ?, editWho = ?, 
			currentVersion = currentVersion + 1, oprSeqFlag = ?
		WHERE taskInfoId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		now, operatorId, oprSeqFlag, taskInfoId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除任务运行时信息失败")
	}

	return nil
} 