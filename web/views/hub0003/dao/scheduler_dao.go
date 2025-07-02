package hub0003dao

import (
	"context"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/database/sqlutils"
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
	query := "SELECT * FROM " + scheduler.TableName() + " WHERE tenantId = ? AND schedulerId = ?"
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
		"WHERE tenantId = ? AND schedulerId = ?"
	
	args := []interface{}{
		scheduler.SchedulerName, scheduler.SchedulerInstanceId, scheduler.MaxWorkers, scheduler.QueueSize,
		scheduler.DefaultTimeoutSeconds, scheduler.DefaultRetries, scheduler.SchedulerStatus, scheduler.LastStartTime, scheduler.LastStopTime,
		scheduler.ServerName, scheduler.ServerIp, scheduler.ServerPort, scheduler.TotalTaskCount, scheduler.RunningTaskCount, scheduler.LastHeartbeatTime,
		scheduler.EditTime, scheduler.EditWho, scheduler.OprSeqFlag, scheduler.NoteText,
		scheduler.TenantId, scheduler.SchedulerId,
	}
	
	return dao.db.Exec(ctx, query, args, true)
}

// Delete 删除调度器配置（物理删除）
func (dao *SchedulerDao) Delete(ctx context.Context, tenantId, schedulerId, editWho string) (int64, error) {
	query := "DELETE FROM " + (&hub0003models.TimerScheduler{}).TableName() + 
		" WHERE tenantId = ? AND schedulerId = ?"
	return dao.db.Exec(ctx, query, []interface{}{tenantId, schedulerId}, true)
}

// Query 查询调度器配置列表
func (dao *SchedulerDao) Query(ctx context.Context, params map[string]interface{}, page, pageSize int) ([]*hub0003models.TimerScheduler, int64, error) {
	var schedulers []*hub0003models.TimerScheduler
	
	// 构建查询条件
	whereClause := "WHERE 1=1 "
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
	
	// 构建基础查询语句
	baseQuery := "SELECT * FROM " + (&hub0003models.TimerScheduler{}).TableName() + " " + whereClause + "ORDER BY editTime DESC"
	
	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, err
	}
	
	// 执行统计查询
	var result struct {
		Count int64 `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &result, countQuery, args, true)
	if err != nil {
		return nil, 0, err
	}
	
	// 如果没有数据，直接返回
	if result.Count == 0 {
		return []*hub0003models.TimerScheduler{}, 0, nil
	}
	
	// 创建分页信息
	paginationInfo := sqlutils.NewPaginationInfo(page, pageSize)
	
	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)
	
	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, paginationInfo)
	if err != nil {
		return nil, 0, err
	}
	
	// 合并查询参数
	allArgs := append(args, paginationArgs...)
	
	// 执行分页查询
	err = dao.db.Query(ctx, &schedulers, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, 0, err
	}
	
	return schedulers, result.Count, nil
}

// UpdateStatus 更新调度器状态
func (dao *SchedulerDao) UpdateStatus(ctx context.Context, tenantId, schedulerId string, status int, editWho string) (int64, error) {
	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)
	
	// 获取当前时间函数
	timeFunc, err := sqlutils.GetCurrentTimeFunction(dbType)
	if err != nil {
		return 0, err
	}
	
	var lastStartTime, lastStopTime string
	
	// 根据状态设置开始或停止时间
	if status == 2 { // 运行中
		lastStartTime = timeFunc
		lastStopTime = "lastStopTime"
	} else if status == 1 { // 停止
		lastStartTime = "lastStartTime"
		lastStopTime = timeFunc
	} else { // 其他状态
		lastStartTime = "lastStartTime"
		lastStopTime = "lastStopTime"
	}
	
	query := "UPDATE " + (&hub0003models.TimerScheduler{}).TableName() + 
		fmt.Sprintf(" SET schedulerStatus = ?, lastStartTime = %s, lastStopTime = %s, ", lastStartTime, lastStopTime) +
		fmt.Sprintf("editWho = ?, editTime = %s, currentVersion = currentVersion + 1 ", timeFunc) +
		"WHERE tenantId = ? AND schedulerId = ?"
	
	return dao.db.Exec(ctx, query, []interface{}{status, editWho, tenantId, schedulerId}, true)
}

// UpdateHeartbeat 更新调度器心跳
func (dao *SchedulerDao) UpdateHeartbeat(ctx context.Context, tenantId, schedulerId string, taskCount, runningCount int) (int64, error) {
	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)
	
	// 获取当前时间函数
	timeFunc, err := sqlutils.GetCurrentTimeFunction(dbType)
	if err != nil {
		return 0, err
	}
	
	query := "UPDATE " + (&hub0003models.TimerScheduler{}).TableName() + 
		fmt.Sprintf(" SET lastHeartbeatTime = %s, totalTaskCount = ?, runningTaskCount = ?, ", timeFunc) +
		fmt.Sprintf("editTime = %s, currentVersion = currentVersion + 1 ", timeFunc) +
		"WHERE tenantId = ? AND schedulerId = ? AND schedulerStatus = 2"
	
	return dao.db.Exec(ctx, query, []interface{}{taskCount, runningCount, tenantId, schedulerId}, true)
} 