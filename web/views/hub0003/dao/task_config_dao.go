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

// TaskConfigDAO 定时任务配置数据访问对象
type TaskConfigDAO struct {
	db database.Database
}

// NewTaskConfigDAO 创建定时任务配置DAO
func NewTaskConfigDAO(db database.Database) *TaskConfigDAO {
	return &TaskConfigDAO{
		db: db,
	}
}

// generateTaskConfigId 生成任务配置ID
// 格式：TASK + YYYYMMDD + HHMMSS + 4位随机数
// 示例：TASK20240615143022A1B2
func (dao *TaskConfigDAO) generateTaskConfigId() string {
	now := time.Now()
	timeStr := now.Format("20060102150405")
	randomStr := random.GenerateRandomString(4)
	return fmt.Sprintf("TASK%s%s", timeStr, randomStr)
}

// isTaskConfigIdExists 检查任务配置ID是否已存在
func (dao *TaskConfigDAO) isTaskConfigIdExists(ctx context.Context, taskConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_TIMER_TASK_CONFIG WHERE taskConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{taskConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueTaskConfigId 生成唯一的任务配置ID
func (dao *TaskConfigDAO) generateUniqueTaskConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		taskConfigId := dao.generateTaskConfigId()
		
		exists, err := dao.isTaskConfigIdExists(ctx, taskConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查任务配置ID是否存在失败")
		}
		
		if !exists {
			return taskConfigId, nil
		}
		
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一任务配置ID失败，已达到最大尝试次数")
}

// AddTaskConfig 添加定时任务配置
func (dao *TaskConfigDAO) AddTaskConfig(ctx context.Context, config *models.TaskConfig, operatorId string) error {
	if config.TenantId == "" || config.TaskId == "" || config.TaskName == "" {
		return errors.New("tenantId、taskId和taskName不能为空")
	}

	// 自动生成任务配置ID
	if config.TaskConfigId == "" {
		generatedId, err := dao.generateUniqueTaskConfigId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成任务配置ID失败")
		}
		config.TaskConfigId = generatedId
	}

	// 验证调度类型相关参数
	if err := dao.validateScheduleConfig(config); err != nil {
		return err
	}

	// 设置自动填充字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.TaskConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.TaskPriority == 0 {
		config.TaskPriority = 1
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 0
	}
	if config.RetryIntervalSeconds == 0 {
		config.RetryIntervalSeconds = 60
	}
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 1800
	}

	_, err := dao.db.Insert(ctx, "HUB_TIMER_TASK_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加定时任务配置失败")
	}

	return nil
}

// validateScheduleConfig 验证调度配置参数
func (dao *TaskConfigDAO) validateScheduleConfig(config *models.TaskConfig) error {
	switch config.ScheduleType {
	case 2: // 固定间隔
		if config.IntervalSeconds == nil || *config.IntervalSeconds <= 0 {
			return errors.New("固定间隔调度类型需要设置intervalSeconds且必须大于0")
		}
	case 3: // Cron表达式
		if config.CronExpression == nil || *config.CronExpression == "" {
			return errors.New("Cron表达式调度类型需要设置cronExpression")
		}
		// TODO: 可以添加Cron表达式格式验证
	case 4: // 延迟执行
		if config.DelaySeconds == nil || *config.DelaySeconds <= 0 {
			return errors.New("延迟执行调度类型需要设置delaySeconds且必须大于0")
		}
	}
	return nil
}

// GetTaskConfig 根据租户ID和配置ID获取任务配置
func (dao *TaskConfigDAO) GetTaskConfig(tenantId, taskConfigId string) (*models.TaskConfig, error) {
	if taskConfigId == "" || tenantId == "" {
		return nil, errors.New("taskConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_CONFIG 
		WHERE taskConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var config models.TaskConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{taskConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询任务配置失败")
	}

	return &config, nil
}

// GetTaskConfigByTaskId 根据任务ID获取任务配置
func (dao *TaskConfigDAO) GetTaskConfigByTaskId(tenantId, taskId string) (*models.TaskConfig, error) {
	if taskId == "" || tenantId == "" {
		return nil, errors.New("taskId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_CONFIG 
		WHERE taskId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var config models.TaskConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{taskId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询任务配置失败")
	}

	return &config, nil
}

// UpdateTaskConfig 更新任务配置
func (dao *TaskConfigDAO) UpdateTaskConfig(ctx context.Context, config *models.TaskConfig, operatorId string) error {
	if config.TaskConfigId == "" || config.TenantId == "" {
		return errors.New("taskConfigId和tenantId不能为空")
	}

	// 首先获取当前配置
	currentConfig, err := dao.GetTaskConfig(config.TenantId, config.TaskConfigId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("任务配置不存在")
	}

	// 验证调度类型相关参数
	if err := dao.validateScheduleConfig(config); err != nil {
		return err
	}

	// 更新修改信息
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.OprSeqFlag = config.TaskConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	sql := `
		UPDATE HUB_TIMER_TASK_CONFIG SET
			taskId = ?, taskName = ?, taskDescription = ?, taskPriority = ?,
			scheduleType = ?, cronExpression = ?, intervalSeconds = ?, delaySeconds = ?,
			startTime = ?, endTime = ?, maxRetries = ?, retryIntervalSeconds = ?, timeoutSeconds = ?,
			taskParams = ?, reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			reserved6 = ?, reserved7 = ?, reserved8 = ?, reserved9 = ?, reserved10 = ?,
			noteText = ?, editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?
		WHERE taskConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, sql, []interface{}{
		config.TaskId, config.TaskName, config.TaskDescription, config.TaskPriority,
		config.ScheduleType, config.CronExpression, config.IntervalSeconds, config.DelaySeconds,
		config.StartTime, config.EndTime, config.MaxRetries, config.RetryIntervalSeconds, config.TimeoutSeconds,
		config.TaskParams, config.Reserved1, config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5,
		config.Reserved6, config.Reserved7, config.Reserved8, config.Reserved9, config.Reserved10,
		config.NoteText, config.EditTime, config.EditWho, config.CurrentVersion, config.OprSeqFlag,
		config.TaskConfigId, config.TenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新任务配置失败")
	}

	return nil
}

// DeleteTaskConfig 删除任务配置（软删除）
func (dao *TaskConfigDAO) DeleteTaskConfig(tenantId, taskConfigId, operatorId string) error {
	if taskConfigId == "" || tenantId == "" || operatorId == "" {
		return errors.New("taskConfigId、tenantId和operatorId不能为空")
	}

	// 首先检查任务配置是否存在
	config, err := dao.GetTaskConfig(tenantId, taskConfigId)
	if err != nil {
		return err
	}
	if config == nil {
		return errors.New("任务配置不存在")
	}

	sql := `
		UPDATE HUB_TIMER_TASK_CONFIG SET 
			activeFlag = 'N', 
			editTime = ?, 
			editWho = ?, 
			currentVersion = currentVersion + 1,
			oprSeqFlag = ?
		WHERE taskConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	now := time.Now()
	oprSeqFlag := taskConfigId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]

	_, err = dao.db.Exec(context.Background(), sql, []interface{}{
		now, operatorId, oprSeqFlag, taskConfigId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除任务配置失败")
	}

	return nil
}

// ListTaskConfigs 查询任务配置列表（分页）
func (dao *TaskConfigDAO) ListTaskConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.TaskConfig, int, error) {
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
		FROM HUB_TIMER_TASK_CONFIG 
		WHERE tenantId = ? AND activeFlag = 'Y'
	`

	var totalCount struct {
		Total int `db:"total"`
	}

	err := dao.db.QueryOne(ctx, &totalCount, countSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询任务配置总数失败")
	}

	// 查询数据
	offset := (page - 1) * pageSize
	dataSQL := `
		SELECT * FROM HUB_TIMER_TASK_CONFIG 
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY addTime DESC 
		LIMIT ? OFFSET ?
	`

	var configs []*models.TaskConfig
	err = dao.db.Query(ctx, &configs, dataSQL, []interface{}{tenantId, pageSize, offset}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询任务配置列表失败")
	}

	return configs, totalCount.Total, nil
}

// GetTaskConfigsByScheduleType 根据调度类型查询任务配置
func (dao *TaskConfigDAO) GetTaskConfigsByScheduleType(ctx context.Context, tenantId string, scheduleType int) ([]*models.TaskConfig, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_TASK_CONFIG 
		WHERE tenantId = ? AND scheduleType = ? AND activeFlag = 'Y'
		ORDER BY taskPriority DESC, addTime ASC
	`

	var configs []*models.TaskConfig
	err := dao.db.Query(ctx, &configs, query, []interface{}{tenantId, scheduleType}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "根据调度类型查询任务配置失败")
	}

	return configs, nil
} 