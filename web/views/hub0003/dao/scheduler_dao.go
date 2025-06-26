package hub0003dao

import (
	"context"
	"gohub/pkg/database"
	hub0003models "gohub/web/views/hub0003/models"
)

// SchedulerDao 调度器数据访问对象
type SchedulerDao struct {
	db database.Database
}

// NewSchedulerDao 创建调度器数据访问对象
func NewSchedulerDao(db database.Database) *SchedulerDao {
	return &SchedulerDao{db: db}
}

// Add 添加调度器配置
func (dao *SchedulerDao) Add(ctx context.Context, scheduler *hub0003models.TimerScheduler) (int64, error) {
	return dao.db.Insert(ctx, scheduler.TableName(), scheduler, true)
}

// GetById 根据ID获取调度器配置
func (dao *SchedulerDao) GetById(ctx context.Context, tenantId, schedulerId string) (*hub0003models.TimerScheduler, error) {
	scheduler := &hub0003models.TimerScheduler{}
	query := "SELECT * FROM " + scheduler.TableName() + " WHERE tenantId = ? AND schedulerId = ? AND activeFlag = 'Y'"
	err := dao.db.QueryOne(ctx, scheduler, query, []interface{}{tenantId, schedulerId}, true)
	if err != nil {
		return nil, err
	}
	return scheduler, nil
}

// Update 更新调度器配置
func (dao *SchedulerDao) Update(ctx context.Context, scheduler *hub0003models.TimerScheduler) (int64, error) {
	query := "UPDATE " + scheduler.TableName() + " SET schedulerName = ?, schedulerInstanceId = ?, maxWorkers = ?, queueSize = ?, " +
		"defaultTimeoutSeconds = ?, defaultRetries = ?, schedulerStatus = ?, lastStartTime = ?, lastStopTime = ?, " +
		"serverName = ?, serverIp = ?, serverPort = ?, totalTaskCount = ?, runningTaskCount = ?, lastHeartbeatTime = ?, " +
		"editTime = ?, editWho = ?, oprSeqFlag = ?, currentVersion = currentVersion + 1, noteText = ? " +
		"WHERE tenantId = ? AND schedulerId = ? AND activeFlag = 'Y'"
	
	args := []interface{}{
		scheduler.SchedulerName, scheduler.SchedulerInstanceId, scheduler.MaxWorkers, scheduler.QueueSize,
		scheduler.DefaultTimeoutSeconds, scheduler.DefaultRetries, scheduler.SchedulerStatus, scheduler.LastStartTime, scheduler.LastStopTime,
		scheduler.ServerName, scheduler.ServerIp, scheduler.ServerPort, scheduler.TotalTaskCount, scheduler.RunningTaskCount, scheduler.LastHeartbeatTime,
		scheduler.EditTime, scheduler.EditWho, scheduler.OprSeqFlag, scheduler.NoteText,
		scheduler.TenantId, scheduler.SchedulerId,
	}
	
	return dao.db.Exec(ctx, query, args, true)
}

// Delete 删除调度器配置（逻辑删除）
func (dao *SchedulerDao) Delete(ctx context.Context, tenantId, schedulerId, editWho string) (int64, error) {
	query := "UPDATE " + (&hub0003models.TimerScheduler{}).TableName() + " SET activeFlag = 'N', editWho = ?, editTime = NOW() " +
		"WHERE tenantId = ? AND schedulerId = ? AND activeFlag = 'Y'"
	return dao.db.Exec(ctx, query, []interface{}{editWho, tenantId, schedulerId}, true)
}

// Query 查询调度器配置列表
func (dao *SchedulerDao) Query(ctx context.Context, params map[string]interface{}, page, pageSize int) ([]*hub0003models.TimerScheduler, int64, error) {
	var schedulers []*hub0003models.TimerScheduler
	
	// 构建查询条件
	whereClause := "WHERE activeFlag = 'Y' "
	args := []interface{}{}
	
	if tenantId, ok := params["tenantId"].(string); ok && tenantId != "" {
		whereClause += "AND tenantId = ? "
		args = append(args, tenantId)
	}
	
	if schedulerName, ok := params["schedulerName"].(string); ok && schedulerName != "" {
		whereClause += "AND schedulerName LIKE ? "
		args = append(args, "%"+schedulerName+"%")
	}
	
	if schedulerStatus, ok := params["schedulerStatus"].(int); ok && schedulerStatus > 0 {
		whereClause += "AND schedulerStatus = ? "
		args = append(args, schedulerStatus)
	}
	
	// 计算总记录数
	countQuery := "SELECT COUNT(*) FROM " + (&hub0003models.TimerScheduler{}).TableName() + " " + whereClause
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err := dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	total := result.Count
	
	// 查询数据
	query := "SELECT * FROM " + (&hub0003models.TimerScheduler{}).TableName() + " " + whereClause +
		"ORDER BY editTime DESC LIMIT ?, ?"
	args = append(args, (page-1)*pageSize, pageSize)
	
	err = dao.db.Query(ctx, &schedulers, query, args, true)
	if err != nil {
		return nil, 0, err
	}
	
	return schedulers, total, nil
}

// UpdateStatus 更新调度器状态
func (dao *SchedulerDao) UpdateStatus(ctx context.Context, tenantId, schedulerId string, status int, editWho string) (int64, error) {
	var lastStartTime, lastStopTime string
	
	// 根据状态设置开始或停止时间
	if status == 2 { // 运行中
		lastStartTime = "NOW()"
		lastStopTime = "lastStopTime"
	} else if status == 1 { // 停止
		lastStartTime = "lastStartTime"
		lastStopTime = "NOW()"
	} else { // 其他状态
		lastStartTime = "lastStartTime"
		lastStopTime = "lastStopTime"
	}
	
	query := "UPDATE " + (&hub0003models.TimerScheduler{}).TableName() + 
		" SET schedulerStatus = ?, lastStartTime = " + lastStartTime + ", lastStopTime = " + lastStopTime + 
		", editWho = ?, editTime = NOW(), currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND schedulerId = ? AND activeFlag = 'Y'"
	
	return dao.db.Exec(ctx, query, []interface{}{status, editWho, tenantId, schedulerId}, true)
}

// UpdateHeartbeat 更新调度器心跳
func (dao *SchedulerDao) UpdateHeartbeat(ctx context.Context, tenantId, schedulerId string, taskCount, runningCount int) (int64, error) {
	query := "UPDATE " + (&hub0003models.TimerScheduler{}).TableName() + 
		" SET lastHeartbeatTime = NOW(), totalTaskCount = ?, runningTaskCount = ?, " +
		"editTime = NOW(), currentVersion = currentVersion + 1 " +
		"WHERE tenantId = ? AND schedulerId = ? AND activeFlag = 'Y' AND schedulerStatus = 2"
	
	return dao.db.Exec(ctx, query, []interface{}{taskCount, runningCount, tenantId, schedulerId}, true)
} 