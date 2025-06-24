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

// SchedulerConfigDAO 调度器配置数据访问对象
type SchedulerConfigDAO struct {
	db database.Database
}

// NewSchedulerConfigDAO 创建调度器配置DAO
func NewSchedulerConfigDAO(db database.Database) *SchedulerConfigDAO {
	return &SchedulerConfigDAO{
		db: db,
	}
}

// generateSchedulerConfigId 生成调度器配置ID
func (dao *SchedulerConfigDAO) generateSchedulerConfigId() string {
	now := time.Now()
	timeStr := now.Format("20060102150405")
	randomStr := random.GenerateRandomString(4)
	return fmt.Sprintf("SCHD%s%s", timeStr, randomStr)
}

// isSchedulerConfigIdExists 检查调度器配置ID是否已存在
func (dao *SchedulerConfigDAO) isSchedulerConfigIdExists(ctx context.Context, schedulerConfigId string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM HUB_TIMER_SCHEDULER_CONFIG WHERE schedulerConfigId = ?`
	
	var result struct {
		Count int `db:"count"`
	}
	
	err := dao.db.QueryOne(ctx, &result, query, []interface{}{schedulerConfigId}, true)
	if err != nil {
		return false, err
	}
	
	return result.Count > 0, nil
}

// generateUniqueSchedulerConfigId 生成唯一的调度器配置ID
func (dao *SchedulerConfigDAO) generateUniqueSchedulerConfigId(ctx context.Context) (string, error) {
	const maxAttempts = 10
	
	for attempt := 0; attempt < maxAttempts; attempt++ {
		schedulerConfigId := dao.generateSchedulerConfigId()
		
		exists, err := dao.isSchedulerConfigIdExists(ctx, schedulerConfigId)
		if err != nil {
			return "", huberrors.WrapError(err, "检查调度器配置ID是否存在失败")
		}
		
		if !exists {
			return schedulerConfigId, nil
		}
		
		time.Sleep(time.Millisecond)
	}
	
	return "", errors.New("生成唯一调度器配置ID失败，已达到最大尝试次数")
}

// AddSchedulerConfig 添加调度器配置
func (dao *SchedulerConfigDAO) AddSchedulerConfig(ctx context.Context, config *models.SchedulerConfig, operatorId string) error {
	if config.TenantId == "" || config.SchedulerName == "" || config.SchedulerInstanceId == "" {
		return errors.New("tenantId、schedulerName和schedulerInstanceId不能为空")
	}

	// 自动生成调度器配置ID
	if config.SchedulerConfigId == "" {
		generatedId, err := dao.generateUniqueSchedulerConfigId(ctx)
		if err != nil {
			return huberrors.WrapError(err, "生成调度器配置ID失败")
		}
		config.SchedulerConfigId = generatedId
	}

	// 设置自动填充字段
	now := time.Now()
	config.AddTime = now
	config.AddWho = operatorId
	config.EditTime = now
	config.EditWho = operatorId
	config.OprSeqFlag = config.SchedulerConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]
	config.CurrentVersion = 1
	config.ActiveFlag = "Y"

	// 设置默认值
	if config.MaxWorkers == 0 {
		config.MaxWorkers = 5
	}
	if config.QueueSize == 0 {
		config.QueueSize = 100
	}
	if config.DefaultTimeoutSeconds == 0 {
		config.DefaultTimeoutSeconds = 1800
	}
	if config.DefaultRetries == 0 {
		config.DefaultRetries = 3
	}
	if config.SchedulerStatus == 0 {
		config.SchedulerStatus = 1 // 默认停止状态
	}

	_, err := dao.db.Insert(ctx, "HUB_TIMER_SCHEDULER_CONFIG", config, true)
	if err != nil {
		return huberrors.WrapError(err, "添加调度器配置失败")
	}

	return nil
}

// GetSchedulerConfig 根据租户ID和配置ID获取调度器配置
func (dao *SchedulerConfigDAO) GetSchedulerConfig(tenantId, schedulerConfigId string) (*models.SchedulerConfig, error) {
	if schedulerConfigId == "" || tenantId == "" {
		return nil, errors.New("schedulerConfigId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_SCHEDULER_CONFIG 
		WHERE schedulerConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var config models.SchedulerConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{schedulerConfigId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询调度器配置失败")
	}

	return &config, nil
}

// GetSchedulerConfigByInstanceId 根据调度器实例ID获取配置
func (dao *SchedulerConfigDAO) GetSchedulerConfigByInstanceId(tenantId, schedulerInstanceId string) (*models.SchedulerConfig, error) {
	if schedulerInstanceId == "" || tenantId == "" {
		return nil, errors.New("schedulerInstanceId和tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_SCHEDULER_CONFIG 
		WHERE schedulerInstanceId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	var config models.SchedulerConfig
	err := dao.db.QueryOne(context.Background(), &config, query, []interface{}{schedulerInstanceId, tenantId}, true)

	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, huberrors.WrapError(err, "查询调度器配置失败")
	}

	return &config, nil
}

// UpdateSchedulerConfig 更新调度器配置
func (dao *SchedulerConfigDAO) UpdateSchedulerConfig(ctx context.Context, config *models.SchedulerConfig, operatorId string) error {
	if config.SchedulerConfigId == "" || config.TenantId == "" {
		return errors.New("schedulerConfigId和tenantId不能为空")
	}

	// 首先获取当前配置
	currentConfig, err := dao.GetSchedulerConfig(config.TenantId, config.SchedulerConfigId)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("调度器配置不存在")
	}

	// 更新修改信息
	config.EditTime = time.Now()
	config.EditWho = operatorId
	config.CurrentVersion = currentConfig.CurrentVersion + 1
	config.OprSeqFlag = config.SchedulerConfigId + "_" + strings.ReplaceAll(time.Now().String(), ".", "")[:8]

	sql := `
		UPDATE HUB_TIMER_SCHEDULER_CONFIG SET
			schedulerName = ?, schedulerInstanceId = ?, maxWorkers = ?, queueSize = ?,
			defaultTimeoutSeconds = ?, defaultRetries = ?, schedulerStatus = ?,
			lastStartTime = ?, lastStopTime = ?, serverName = ?, serverIp = ?, serverPort = ?,
			totalTaskCount = ?, runningTaskCount = ?, lastHeartbeatTime = ?,
			reserved1 = ?, reserved2 = ?, reserved3 = ?, reserved4 = ?, reserved5 = ?,
			reserved6 = ?, reserved7 = ?, reserved8 = ?, reserved9 = ?, reserved10 = ?,
			noteText = ?, editTime = ?, editWho = ?, currentVersion = ?, oprSeqFlag = ?
		WHERE schedulerConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err = dao.db.Exec(ctx, sql, []interface{}{
		config.SchedulerName, config.SchedulerInstanceId, config.MaxWorkers, config.QueueSize,
		config.DefaultTimeoutSeconds, config.DefaultRetries, config.SchedulerStatus,
		config.LastStartTime, config.LastStopTime, config.ServerName, config.ServerIp, config.ServerPort,
		config.TotalTaskCount, config.RunningTaskCount, config.LastHeartbeatTime,
		config.Reserved1, config.Reserved2, config.Reserved3, config.Reserved4, config.Reserved5,
		config.Reserved6, config.Reserved7, config.Reserved8, config.Reserved9, config.Reserved10,
		config.NoteText, config.EditTime, config.EditWho, config.CurrentVersion, config.OprSeqFlag,
		config.SchedulerConfigId, config.TenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新调度器配置失败")
	}

	return nil
}

// UpdateSchedulerStatus 更新调度器状态
func (dao *SchedulerConfigDAO) UpdateSchedulerStatus(ctx context.Context, tenantId, schedulerInstanceId string, status int, operatorId string) error {
	if tenantId == "" || schedulerInstanceId == "" {
		return errors.New("tenantId和schedulerInstanceId不能为空")
	}

	now := time.Now()
	oprSeqFlag := schedulerInstanceId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]

	var sql string
	var args []interface{}

	if status == 2 { // 运行中
		sql = `
			UPDATE HUB_TIMER_SCHEDULER_CONFIG SET
				schedulerStatus = ?, lastStartTime = ?, editTime = ?, editWho = ?, 
				currentVersion = currentVersion + 1, oprSeqFlag = ?
			WHERE schedulerInstanceId = ? AND tenantId = ? AND activeFlag = 'Y'
		`
		args = []interface{}{status, now, now, operatorId, oprSeqFlag, schedulerInstanceId, tenantId}
	} else { // 停止或暂停
		sql = `
			UPDATE HUB_TIMER_SCHEDULER_CONFIG SET
				schedulerStatus = ?, lastStopTime = ?, editTime = ?, editWho = ?, 
				currentVersion = currentVersion + 1, oprSeqFlag = ?
			WHERE schedulerInstanceId = ? AND tenantId = ? AND activeFlag = 'Y'
		`
		args = []interface{}{status, now, now, operatorId, oprSeqFlag, schedulerInstanceId, tenantId}
	}

	_, err := dao.db.Exec(ctx, sql, args, true)
	if err != nil {
		return huberrors.WrapError(err, "更新调度器状态失败")
	}

	return nil
}

// UpdateSchedulerHeartbeat 更新调度器心跳
func (dao *SchedulerConfigDAO) UpdateSchedulerHeartbeat(ctx context.Context, tenantId, schedulerInstanceId string, 
	totalTaskCount, runningTaskCount int, operatorId string) error {
	
	if tenantId == "" || schedulerInstanceId == "" {
		return errors.New("tenantId和schedulerInstanceId不能为空")
	}

	now := time.Now()
	oprSeqFlag := schedulerInstanceId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]

	sql := `
		UPDATE HUB_TIMER_SCHEDULER_CONFIG SET
			totalTaskCount = ?, runningTaskCount = ?, lastHeartbeatTime = ?,
			editTime = ?, editWho = ?, currentVersion = currentVersion + 1, oprSeqFlag = ?
		WHERE schedulerInstanceId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		totalTaskCount, runningTaskCount, now, now, operatorId, oprSeqFlag, schedulerInstanceId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "更新调度器心跳失败")
	}

	return nil
}

// DeleteSchedulerConfig 删除调度器配置（软删除）
func (dao *SchedulerConfigDAO) DeleteSchedulerConfig(ctx context.Context, tenantId, schedulerConfigId, operatorId string) error {
	if schedulerConfigId == "" || tenantId == "" {
		return errors.New("schedulerConfigId和tenantId不能为空")
	}

	now := time.Now()
	oprSeqFlag := schedulerConfigId + "_" + strings.ReplaceAll(now.String(), ".", "")[:8]

	sql := `
		UPDATE HUB_TIMER_SCHEDULER_CONFIG SET
			activeFlag = 'N', editTime = ?, editWho = ?, 
			currentVersion = currentVersion + 1, oprSeqFlag = ?
		WHERE schedulerConfigId = ? AND tenantId = ? AND activeFlag = 'Y'
	`

	_, err := dao.db.Exec(ctx, sql, []interface{}{
		now, operatorId, oprSeqFlag, schedulerConfigId, tenantId,
	}, true)

	if err != nil {
		return huberrors.WrapError(err, "删除调度器配置失败")
	}

	return nil
}

// ListSchedulerConfigs 查询调度器配置列表（分页）
func (dao *SchedulerConfigDAO) ListSchedulerConfigs(ctx context.Context, tenantId string, page, pageSize int) ([]*models.SchedulerConfig, int, error) {
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
	countQuery := "SELECT COUNT(*) as count FROM HUB_TIMER_SCHEDULER_CONFIG WHERE tenantId = ? AND activeFlag = 'Y'"
	var result struct {
		Count int `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, []interface{}{tenantId}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询调度器配置总数失败")
	}

	// 如果没有数据，直接返回
	if result.Count == 0 {
		return make([]*models.SchedulerConfig, 0), 0, nil
	}

	// 查询分页数据
	offset := (page - 1) * pageSize
	query := `
		SELECT * FROM HUB_TIMER_SCHEDULER_CONFIG 
		WHERE tenantId = ? AND activeFlag = 'Y' 
		ORDER BY editTime DESC 
		LIMIT ? OFFSET ?
	`

	var configs []*models.SchedulerConfig
	err = dao.db.Query(ctx, &configs, query, []interface{}{tenantId, pageSize, offset}, true)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "查询调度器配置列表失败")
	}

	return configs, result.Count, nil
}

// GetSchedulerConfigsByStatus 根据调度器状态查询配置
func (dao *SchedulerConfigDAO) GetSchedulerConfigsByStatus(ctx context.Context, tenantId string, status int) ([]*models.SchedulerConfig, error) {
	if tenantId == "" {
		return nil, errors.New("tenantId不能为空")
	}

	query := `
		SELECT * FROM HUB_TIMER_SCHEDULER_CONFIG 
		WHERE tenantId = ? AND schedulerStatus = ? AND activeFlag = 'Y'
		ORDER BY editTime DESC
	`

	var configs []*models.SchedulerConfig
	err := dao.db.Query(ctx, &configs, query, []interface{}{tenantId, status}, true)
	if err != nil {
		return nil, huberrors.WrapError(err, "查询调度器配置列表失败")
	}

	return configs, nil
}

// QuerySchedulerConfigs 查询调度器配置列表
func (dao *SchedulerConfigDAO) QuerySchedulerConfigs(ctx context.Context, tenantId, schedulerName, status string, pageNum, pageSize int) (int64, []*models.SchedulerConfig, error) {
	if tenantId == "" {
		return 0, nil, errors.New("tenantId不能为空")
	}

	// 构建查询条件
	conditions := []string{"tenantId = ? AND activeFlag = 'Y'"}
	args := []interface{}{tenantId}

	if schedulerName != "" {
		conditions = append(conditions, "schedulerName LIKE ?")
		args = append(args, "%"+schedulerName+"%")
	}

	if status != "" {
		conditions = append(conditions, "schedulerStatus = ?")
		args = append(args, status)
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) as count FROM HUB_TIMER_SCHEDULER_CONFIG WHERE %s", whereClause)
	var result struct {
		Count int64 `db:"count"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return 0, nil, huberrors.WrapError(err, "查询调度器配置总数失败")
	}

	// 如果没有数据，直接返回
	if result.Count == 0 {
		return 0, make([]*models.SchedulerConfig, 0), nil
	}

	// 查询分页数据
	offset := (pageNum - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT * FROM HUB_TIMER_SCHEDULER_CONFIG 
		WHERE %s 
		ORDER BY editTime DESC 
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pageSize, offset)
	var configs []*models.SchedulerConfig
	err = dao.db.Query(ctx, &configs, query, args, true)
	if err != nil {
		return 0, nil, huberrors.WrapError(err, "查询调度器配置列表失败")
	}

	return result.Count, configs, nil
} 