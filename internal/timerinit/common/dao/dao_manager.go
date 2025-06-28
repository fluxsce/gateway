package dao

import (
	"context"
	"fmt"
	"gohub/pkg/database"
)

// DAOManager 数据访问对象管理器
// 统一管理所有的DAO实例，提供便捷的数据访问接口
type DAOManager struct {
	db database.Database
	
	// TimerTypes相关DAO
	TaskDAO      *TimerTaskDAO
	SchedulerDAO *TimerSchedulerDAO
	
	// ToolTypes相关DAO
	ToolConfigDAO      *ToolConfigDAO
	ToolConfigGroupDAO *ToolConfigGroupDAO
}

// NewDAOManager 创建DAO管理器实例
func NewDAOManager(db database.Database) *DAOManager {
	return &DAOManager{
		db:                 db,
		TaskDAO:           NewTimerTaskDAO(db),
		SchedulerDAO:      NewTimerSchedulerDAO(db),
		ToolConfigDAO:     NewToolConfigDAO(db),
		ToolConfigGroupDAO: NewToolConfigGroupDAO(db),
	}
}

// GetDatabase 获取数据库连接
func (dm *DAOManager) GetDatabase() database.Database {
	return dm.db
}


// HealthCheck 健康检查
// 检查数据库连接是否正常
func (dm *DAOManager) HealthCheck(ctx context.Context) error {
	if dm.db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	
	return dm.db.Ping(ctx)
}

// Close 关闭数据库连接
func (dm *DAOManager) Close() error {
	if dm.db != nil {
		return dm.db.Close()
	}
	return nil
}

// === 便捷方法：Timer相关 ===

// GetTaskDAO 获取任务DAO
func (dm *DAOManager) GetTaskDAO() *TimerTaskDAO {
	return dm.TaskDAO
}

// GetSchedulerDAO 获取调度器DAO
func (dm *DAOManager) GetSchedulerDAO() *TimerSchedulerDAO {
	return dm.SchedulerDAO
}

// === 便捷方法：Tool相关 ===

// GetToolConfigDAO 获取工具配置DAO
func (dm *DAOManager) GetToolConfigDAO() *ToolConfigDAO {
	return dm.ToolConfigDAO
}

// GetToolConfigGroupDAO 获取工具配置分组DAO
func (dm *DAOManager) GetToolConfigGroupDAO() *ToolConfigGroupDAO {
	return dm.ToolConfigGroupDAO
}

// === 统计查询方法 ===

// GetTaskStats 获取任务统计信息
func (dm *DAOManager) GetTaskStats(ctx context.Context, tenantId string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 查询任务总数
	totalSQL := "SELECT COUNT(*) FROM HUB_TIMER_TASK WHERE tenantId = ? AND activeFlag = 'Y'"
	var total int64
	err := dm.db.QueryOne(ctx, &total, totalSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询任务总数失败: %w", err)
	}
	stats["total"] = total
	
	// 查询各状态任务数量
	statusSQL := `
		SELECT taskStatus, COUNT(*) as count 
		FROM HUB_TIMER_TASK 
		WHERE tenantId = ? AND activeFlag = 'Y' 
		GROUP BY taskStatus
	`
	var statusCounts []struct {
		TaskStatus int   `db:"taskStatus"`
		Count      int64 `db:"count"`
	}
	err = dm.db.Query(ctx, &statusCounts, statusSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询任务状态统计失败: %w", err)
	}
	
	statusMap := make(map[string]int64)
	for _, sc := range statusCounts {
		switch sc.TaskStatus {
		case 1:
			statusMap["pending"] = sc.Count
		case 2:
			statusMap["running"] = sc.Count
		case 3:
			statusMap["completed"] = sc.Count
		case 4:
			statusMap["failed"] = sc.Count
		case 5:
			statusMap["cancelled"] = sc.Count
		}
	}
	stats["status"] = statusMap
	
	return stats, nil
}

// GetSchedulerStats 获取调度器统计信息
func (dm *DAOManager) GetSchedulerStats(ctx context.Context, tenantId string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 查询调度器总数
	totalSQL := "SELECT COUNT(*) FROM HUB_TIMER_SCHEDULER WHERE tenantId = ? AND activeFlag = 'Y'"
	var total int64
	err := dm.db.QueryOne(ctx, &total, totalSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询调度器总数失败: %w", err)
	}
	stats["total"] = total
	
	// 查询运行中的调度器数量
	runningSQL := "SELECT COUNT(*) FROM HUB_TIMER_SCHEDULER WHERE tenantId = ? AND activeFlag = 'Y' AND schedulerStatus = 2"
	var running int64
	err = dm.db.QueryOne(ctx, &running, runningSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询运行中调度器数量失败: %w", err)
	}
	stats["running"] = running
	
	// 查询停止的调度器数量
	stoppedSQL := "SELECT COUNT(*) FROM HUB_TIMER_SCHEDULER WHERE tenantId = ? AND activeFlag = 'Y' AND schedulerStatus = 1"
	var stopped int64
	err = dm.db.QueryOne(ctx, &stopped, stoppedSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询停止调度器数量失败: %w", err)
	}
	stats["stopped"] = stopped
	
	return stats, nil
}

// GetConfigStats 获取工具配置统计信息
func (dm *DAOManager) GetConfigStats(ctx context.Context, tenantId string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 查询配置总数
	totalSQL := "SELECT COUNT(*) FROM HUB_TOOL_CONFIG WHERE tenantId = ? AND activeFlag = 'Y'"
	var total int64
	err := dm.db.QueryOne(ctx, &total, totalSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询工具配置总数失败: %w", err)
	}
	stats["total"] = total
	
	// 查询启用的配置数量
	enabledSQL := "SELECT COUNT(*) FROM HUB_TOOL_CONFIG WHERE tenantId = ? AND activeFlag = 'Y' AND configStatus = 'Y'"
	var enabled int64
	err = dm.db.QueryOne(ctx, &enabled, enabledSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询启用配置数量失败: %w", err)
	}
	stats["enabled"] = enabled
	
	// 查询各工具类型配置数量
	typeSQL := `
		SELECT toolType, COUNT(*) as count 
		FROM HUB_TOOL_CONFIG 
		WHERE tenantId = ? AND activeFlag = 'Y' 
		GROUP BY toolType
	`
	var typeCounts []struct {
		ToolType string `db:"toolType"`
		Count    int64  `db:"count"`
	}
	err = dm.db.Query(ctx, &typeCounts, typeSQL, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询工具类型统计失败: %w", err)
	}
	
	typeMap := make(map[string]int64)
	for _, tc := range typeCounts {
		typeMap[tc.ToolType] = tc.Count
	}
	stats["types"] = typeMap
	
	return stats, nil
}

