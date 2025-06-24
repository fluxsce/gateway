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

// TaskLogDAO 定时任务日志数据访问对象
type TaskLogDAO struct {
	db database.Database
}

// NewTaskLogDAO 创建任务日志DAO
func NewTaskLogDAO(db database.Database) *TaskLogDAO {
	return &TaskLogDAO{
		db: db,
	}
}

// generateTaskLogId 生成任务日志ID
func (dao *TaskLogDAO) generateTaskLogId() string {
	now := time.Now()
	timeStr := now.Format("20060102150405")
	randomStr := random.GenerateRandomString(4)
	return fmt.Sprintf("LOG%s%s", timeStr, randomStr)
}

// isTaskLogIdExists 检查任务日志ID是否已存在
func (dao *TaskLogDAO) isTaskLogIdExists(ctx context.Context, taskLogId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_TIMER_TASK_LOG WHERE taskLogId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{taskLogId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueTaskLogId 生成唯一的任务日志ID
func (dao *TaskLogDAO) generateUniqueTaskLogId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		taskLogId := dao.generateTaskLogId()
		
		exists, err := dao.isTaskLogIdExists(ctx, taskLogId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查任务日志ID是否存在失败")
		}
		
		if !exists {
			return taskLogId, nil
		}
		
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一任务日志ID失败，已达到最大尝试次数")
}

// AddTaskLog 添加任务执行日志
func (dao *TaskLogDAO) AddTaskLog(ctx context.Context, log *models.TaskLog, operatorId string) error {
	if log.TenantId == "" || log.TaskId == "" {
		return errors.New("tenantId和taskId不能为空")
	}

	// 自动生成任务日志ID
	if log.TaskLogId == "" {
		generatedId, err := dao.generateUniqueTaskLogId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成任务日志ID失败")
		}
		log.TaskLogId = generatedId
	}

	// 设置自动填充字段
	now := time.Now()
	log.AddTime = now
	log.AddWho = operatorId
	log.EditTime = now
	log.EditWho = operatorId
	log.OprSeqFlag = log.TaskLogId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	log.CurrentVersion = 1
	log.ActiveFlag = "Y"

	// 设置默认值
	if log.LogTimestamp.IsZero() {
		log.LogTimestamp = now
	}

	_, err := dao.db.Insert(ctx, "HUB_TIMER_TASK_LOG", log, true)
	if err != nil {
		return huberrors.WrapError(err, "添加任务执行日志失败")
	}

	return nil
}

// GetTaskLog 根据租户ID和日志ID获取任务日志
func (dao *TaskLogDAO) GetTaskLog(tenantId, taskLogId string) (*models.TaskLog, error) {
	if taskLogId == "" || tenantId == "" {
		return nil, errors.New("taskLogId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_LOG 
		WHERE taskLogId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var log models.TaskLog
	err := dao.db.QueryOne(context.Background(), &log, query, []interface{}{taskLogId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询任务执行日志失败")
	}

	return &log, nil
}

// QueryTaskLogs 查询任务执行日志列表（分页）
func (dao *TaskLogDAO) QueryTaskLogs(ctx context.Context, tenantId, taskId, taskResultId, logLevel, startTime, endTime string, page, pageSize int) (int, []*models.TaskLog, error) {
	if tenantId == "" {
		return 0, nil, errors.New("tenantId不能为空")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 构建查询条件
	conditions := []string{"tenantId = ? AND activeFlag = 'Y'"}
	params := []interface{}{tenantId}

	if taskId != "" {
		conditions = append(conditions, "taskId = ?")
		params = append(params, taskId)
	}
	if taskResultId != "" {
		conditions = append(conditions, "taskResultId = ?")
		params = append(params, taskResultId)
	}
	if logLevel != "" {
		conditions = append(conditions, "logLevel = ?")
		params = append(params, logLevel)
	}
	if startTime != "" {
		conditions = append(conditions, "logTimestamp >= ?")
		params = append(params, startTime)
	}
	if endTime != "" {
		conditions = append(conditions, "logTimestamp <= ?")
		params = append(params, endTime)
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询总数
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) as total 
		FROM HUB_TIMER_TASK_LOG 
		WHERE %s
	`, whereClause)

	var totalCount struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalCount, countSQL, params, true)
	if err != nil {
		return 0, nil, huberrors.WrapError(err, "查询任务执行日志总数失败")
	}

	// 查询数据
	offset := (page - 1) * pageSize
	dataSQL := fmt.Sprintf(`
		SELECT * FROM HUB_TIMER_TASK_LOG 
		WHERE %s
		ORDER BY logTimestamp DESC, addTime DESC 
		LIMIT ? OFFSET ?
	`, whereClause)

	params = append(params, pageSize, offset)

	var logs []*models.TaskLog
	err = dao.db.Query(ctx, &logs, dataSQL, params, true)
	if err != nil {
		return 0, nil, huberrors.WrapError(err, "查询任务执行日志列表失败")
	}

	return totalCount.Total, logs, nil
}

// GetTaskResultLogs 获取任务执行结果相关的所有日志
func (dao *TaskLogDAO) GetTaskResultLogs(tenantId, taskResultId string) ([]*models.TaskLog, error) {
	if tenantId == "" || taskResultId == "" {
		return nil, errors.New("tenantId和taskResultId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_LOG 
		WHERE tenantId = ? AND taskResultId = ? AND activeFlag = 'Y'
		ORDER BY logTimestamp ASC, addTime ASC
	`

	var logs []*models.TaskLog
	err := dao.db.Query(context.Background(), &logs, query, []interface{}{tenantId, taskResultId}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询任务执行结果相关日志失败")
	}

	return logs, nil
} 