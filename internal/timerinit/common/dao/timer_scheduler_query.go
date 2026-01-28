package dao

import (
	"context"
	"fmt"
	"gateway/internal/types/timertypes"
	"gateway/pkg/database"
	"strings"
	"time"
)

// TimerSchedulerQuery 定义调度器查询条件结构
type TimerSchedulerQuery struct {
	// 基础查询条件
	SchedulerId         *string `json:"schedulerId,omitempty"`
	TenantId            *string `json:"tenantId,omitempty"`
	SchedulerName       *string `json:"schedulerName,omitempty"`
	SchedulerInstanceId *string `json:"schedulerInstanceId,omitempty"`

	// 状态查询
	SchedulerStatus     *int    `json:"schedulerStatus,omitempty"`
	SchedulerStatusList []int   `json:"schedulerStatusList,omitempty"`
	ActiveFlag          *string `json:"activeFlag,omitempty"`

	// 配置查询
	MinMaxWorkers     *int   `json:"minMaxWorkers,omitempty"`
	MaxMaxWorkers     *int   `json:"maxMaxWorkers,omitempty"`
	MinQueueSize      *int   `json:"minQueueSize,omitempty"`
	MaxQueueSize      *int   `json:"maxQueueSize,omitempty"`
	MinDefaultTimeout *int64 `json:"minDefaultTimeout,omitempty"`
	MaxDefaultTimeout *int64 `json:"maxDefaultTimeout,omitempty"`
	MinDefaultRetries *int   `json:"minDefaultRetries,omitempty"`
	MaxDefaultRetries *int   `json:"maxDefaultRetries,omitempty"`

	// 服务器信息查询
	ServerName *string `json:"serverName,omitempty"`
	ServerIp   *string `json:"serverIp,omitempty"`
	ServerPort *int    `json:"serverPort,omitempty"`

	// 监控信息查询
	MinTotalTaskCount   *int `json:"minTotalTaskCount,omitempty"`
	MaxTotalTaskCount   *int `json:"maxTotalTaskCount,omitempty"`
	MinRunningTaskCount *int `json:"minRunningTaskCount,omitempty"`
	MaxRunningTaskCount *int `json:"maxRunningTaskCount,omitempty"`

	// 时间范围查询
	LastStartTimeFrom     *time.Time `json:"lastStartTimeFrom,omitempty"`
	LastStartTimeTo       *time.Time `json:"lastStartTimeTo,omitempty"`
	LastStopTimeFrom      *time.Time `json:"lastStopTimeFrom,omitempty"`
	LastStopTimeTo        *time.Time `json:"lastStopTimeTo,omitempty"`
	LastHeartbeatTimeFrom *time.Time `json:"lastHeartbeatTimeFrom,omitempty"`
	LastHeartbeatTimeTo   *time.Time `json:"lastHeartbeatTimeTo,omitempty"`

	// 创建和修改时间查询
	AddTimeFrom  *time.Time `json:"addTimeFrom,omitempty"`
	AddTimeTo    *time.Time `json:"addTimeTo,omitempty"`
	EditTimeFrom *time.Time `json:"editTimeFrom,omitempty"`
	EditTimeTo   *time.Time `json:"editTimeTo,omitempty"`
	AddWho       *string    `json:"addWho,omitempty"`
	EditWho      *string    `json:"editWho,omitempty"`

	// 模糊查询
	SchedulerNameLike *string `json:"schedulerNameLike,omitempty"`
	ServerNameLike    *string `json:"serverNameLike,omitempty"`
	NoteTextLike      *string `json:"noteTextLike,omitempty"`

	// 健康状态查询
	HealthCheckInterval *time.Duration `json:"healthCheckInterval,omitempty"` // 心跳检查间隔
	OnlyHealthy         *bool          `json:"onlyHealthy,omitempty"`         // 只返回健康的调度器
	OnlyRunning         *bool          `json:"onlyRunning,omitempty"`         // 只返回运行中的调度器

	// 排序和分页
	OrderBy        string `json:"orderBy,omitempty"`        // 排序字段
	OrderDirection string `json:"orderDirection,omitempty"` // ASC/DESC
	PageNum        int    `json:"pageNum,omitempty"`        // 页码，从1开始
	PageSize       int    `json:"pageSize,omitempty"`       // 每页大小
	Offset         int    `json:"offset,omitempty"`         // 偏移量
	Limit          int    `json:"limit,omitempty"`          // 限制数量
}

// TimerSchedulerQueryResult 查询结果结构
type TimerSchedulerQueryResult struct {
	Schedulers []timertypes.TimerScheduler `json:"schedulers"`
	Total      int64                       `json:"total"`
	PageNum    int                         `json:"pageNum"`
	PageSize   int                         `json:"pageSize"`
	TotalPages int                         `json:"totalPages"`
}

// TimerSchedulerDAO 调度器数据访问对象
type TimerSchedulerDAO struct {
	db database.Database
}

// NewTimerSchedulerDAO 创建调度器DAO实例
func NewTimerSchedulerDAO(db database.Database) *TimerSchedulerDAO {
	return &TimerSchedulerDAO{db: db}
}

// BuildWhere 构建WHERE条件语句
func (q *TimerSchedulerQuery) BuildWhere() (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// 基础条件
	if q.SchedulerId != nil {
		conditions = append(conditions, "schedulerId = ?")
		args = append(args, *q.SchedulerId)
	}

	if q.TenantId != nil {
		conditions = append(conditions, "tenantId = ?")
		args = append(args, *q.TenantId)
	}

	if q.SchedulerName != nil {
		conditions = append(conditions, "schedulerName = ?")
		args = append(args, *q.SchedulerName)
	}

	if q.SchedulerInstanceId != nil {
		conditions = append(conditions, "schedulerInstanceId = ?")
		args = append(args, *q.SchedulerInstanceId)
	}

	// 状态条件
	if q.SchedulerStatus != nil {
		conditions = append(conditions, "schedulerStatus = ?")
		args = append(args, *q.SchedulerStatus)
	}

	if len(q.SchedulerStatusList) > 0 {
		placeholders := make([]string, len(q.SchedulerStatusList))
		for i, status := range q.SchedulerStatusList {
			placeholders[i] = "?"
			args = append(args, status)
		}
		conditions = append(conditions, fmt.Sprintf("schedulerStatus IN (%s)", strings.Join(placeholders, ",")))
	}

	if q.ActiveFlag != nil {
		conditions = append(conditions, "activeFlag = ?")
		args = append(args, *q.ActiveFlag)
	}

	// 配置条件
	if q.MinMaxWorkers != nil {
		conditions = append(conditions, "maxWorkers >= ?")
		args = append(args, *q.MinMaxWorkers)
	}

	if q.MaxMaxWorkers != nil {
		conditions = append(conditions, "maxWorkers <= ?")
		args = append(args, *q.MaxMaxWorkers)
	}

	if q.MinQueueSize != nil {
		conditions = append(conditions, "queueSize >= ?")
		args = append(args, *q.MinQueueSize)
	}

	if q.MaxQueueSize != nil {
		conditions = append(conditions, "queueSize <= ?")
		args = append(args, *q.MaxQueueSize)
	}

	if q.MinDefaultTimeout != nil {
		conditions = append(conditions, "defaultTimeoutSeconds >= ?")
		args = append(args, *q.MinDefaultTimeout)
	}

	if q.MaxDefaultTimeout != nil {
		conditions = append(conditions, "defaultTimeoutSeconds <= ?")
		args = append(args, *q.MaxDefaultTimeout)
	}

	if q.MinDefaultRetries != nil {
		conditions = append(conditions, "defaultRetries >= ?")
		args = append(args, *q.MinDefaultRetries)
	}

	if q.MaxDefaultRetries != nil {
		conditions = append(conditions, "defaultRetries <= ?")
		args = append(args, *q.MaxDefaultRetries)
	}

	// 服务器信息条件
	if q.ServerName != nil {
		conditions = append(conditions, "serverName = ?")
		args = append(args, *q.ServerName)
	}

	if q.ServerIp != nil {
		conditions = append(conditions, "serverIp = ?")
		args = append(args, *q.ServerIp)
	}

	if q.ServerPort != nil {
		conditions = append(conditions, "serverPort = ?")
		args = append(args, *q.ServerPort)
	}

	// 监控信息条件
	if q.MinTotalTaskCount != nil {
		conditions = append(conditions, "totalTaskCount >= ?")
		args = append(args, *q.MinTotalTaskCount)
	}

	if q.MaxTotalTaskCount != nil {
		conditions = append(conditions, "totalTaskCount <= ?")
		args = append(args, *q.MaxTotalTaskCount)
	}

	if q.MinRunningTaskCount != nil {
		conditions = append(conditions, "runningTaskCount >= ?")
		args = append(args, *q.MinRunningTaskCount)
	}

	if q.MaxRunningTaskCount != nil {
		conditions = append(conditions, "runningTaskCount <= ?")
		args = append(args, *q.MaxRunningTaskCount)
	}

	// 时间范围条件
	if q.LastStartTimeFrom != nil {
		conditions = append(conditions, "lastStartTime >= ?")
		args = append(args, *q.LastStartTimeFrom)
	}

	if q.LastStartTimeTo != nil {
		conditions = append(conditions, "lastStartTime <= ?")
		args = append(args, *q.LastStartTimeTo)
	}

	if q.LastStopTimeFrom != nil {
		conditions = append(conditions, "lastStopTime >= ?")
		args = append(args, *q.LastStopTimeFrom)
	}

	if q.LastStopTimeTo != nil {
		conditions = append(conditions, "lastStopTime <= ?")
		args = append(args, *q.LastStopTimeTo)
	}

	if q.LastHeartbeatTimeFrom != nil {
		conditions = append(conditions, "lastHeartbeatTime >= ?")
		args = append(args, *q.LastHeartbeatTimeFrom)
	}

	if q.LastHeartbeatTimeTo != nil {
		conditions = append(conditions, "lastHeartbeatTime <= ?")
		args = append(args, *q.LastHeartbeatTimeTo)
	}

	// 创建和修改时间条件
	if q.AddTimeFrom != nil {
		conditions = append(conditions, "addTime >= ?")
		args = append(args, *q.AddTimeFrom)
	}

	if q.AddTimeTo != nil {
		conditions = append(conditions, "addTime <= ?")
		args = append(args, *q.AddTimeTo)
	}

	if q.EditTimeFrom != nil {
		conditions = append(conditions, "editTime >= ?")
		args = append(args, *q.EditTimeFrom)
	}

	if q.EditTimeTo != nil {
		conditions = append(conditions, "editTime <= ?")
		args = append(args, *q.EditTimeTo)
	}

	if q.AddWho != nil {
		conditions = append(conditions, "addWho = ?")
		args = append(args, *q.AddWho)
	}

	if q.EditWho != nil {
		conditions = append(conditions, "editWho = ?")
		args = append(args, *q.EditWho)
	}

	// 模糊查询条件
	if q.SchedulerNameLike != nil {
		conditions = append(conditions, "schedulerName LIKE ?")
		args = append(args, "%"+*q.SchedulerNameLike+"%")
	}

	if q.ServerNameLike != nil {
		conditions = append(conditions, "serverName LIKE ?")
		args = append(args, "%"+*q.ServerNameLike+"%")
	}

	if q.NoteTextLike != nil {
		conditions = append(conditions, "noteText LIKE ?")
		args = append(args, "%"+*q.NoteTextLike+"%")
	}

	// 健康状态条件
	if q.OnlyHealthy != nil && *q.OnlyHealthy {
		if q.HealthCheckInterval != nil {
			healthThreshold := time.Now().Add(-*q.HealthCheckInterval)
			conditions = append(conditions, "lastHeartbeatTime >= ?")
			args = append(args, healthThreshold)
		}
	}

	if q.OnlyRunning != nil && *q.OnlyRunning {
		conditions = append(conditions, "schedulerStatus = ?")
		args = append(args, timertypes.SchedulerStatusRunning)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	return where, args
}

// BuildOrderBy 构建ORDER BY语句
func (q *TimerSchedulerQuery) BuildOrderBy() string {
	if q.OrderBy == "" {
		return "ORDER BY addTime DESC"
	}

	direction := "ASC"
	if strings.ToUpper(q.OrderDirection) == "DESC" {
		direction = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", q.OrderBy, direction)
}

// BuildPagination 构建分页语句
func (q *TimerSchedulerQuery) BuildPagination() string {
	if q.Limit > 0 {
		if q.Offset > 0 {
			return fmt.Sprintf("LIMIT %d OFFSET %d", q.Limit, q.Offset)
		}
		return fmt.Sprintf("LIMIT %d", q.Limit)
	}

	if q.PageSize > 0 {
		offset := 0
		if q.PageNum > 1 {
			offset = (q.PageNum - 1) * q.PageSize
		}
		return fmt.Sprintf("LIMIT %d OFFSET %d", q.PageSize, offset)
	}

	return ""
}

// QuerySchedulers 查询调度器列表
func (dao *TimerSchedulerDAO) QuerySchedulers(ctx context.Context, query *TimerSchedulerQuery) (*TimerSchedulerQueryResult, error) {
	tableName := (&timertypes.TimerScheduler{}).TableName()

	// 构建查询条件
	where, args := query.BuildWhere()
	orderBy := query.BuildOrderBy()
	pagination := query.BuildPagination()

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tableName, where)
	var total int64
	err := dao.db.QueryOne(ctx, &total, countSQL, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询调度器总数失败: %w", err)
	}

	// 查询数据
	selectSQL := fmt.Sprintf("SELECT * FROM %s %s %s %s", tableName, where, orderBy, pagination)
	var schedulers []timertypes.TimerScheduler
	err = dao.db.Query(ctx, &schedulers, selectSQL, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询调度器列表失败: %w", err)
	}

	// 计算分页信息
	result := &TimerSchedulerQueryResult{
		Schedulers: schedulers,
		Total:      total,
	}

	if query.PageSize > 0 {
		result.PageNum = query.PageNum
		if result.PageNum == 0 {
			result.PageNum = 1
		}
		result.PageSize = query.PageSize
		result.TotalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}

	return result, nil
}

// GetSchedulerById 根据ID获取调度器
func (dao *TimerSchedulerDAO) GetSchedulerById(ctx context.Context, tenantId, schedulerId string) (*timertypes.TimerScheduler, error) {
	tableName := (&timertypes.TimerScheduler{}).TableName()
	sql := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND schedulerId = ? AND activeFlag = ?", tableName)

	var scheduler timertypes.TimerScheduler
	err := dao.db.QueryOne(ctx, &scheduler, sql, []interface{}{tenantId, schedulerId, timertypes.ActiveFlagYes}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询调度器失败: %w", err)
	}

	return &scheduler, nil
}

// CreateScheduler 创建调度器
func (dao *TimerSchedulerDAO) CreateScheduler(ctx context.Context, scheduler *timertypes.TimerScheduler) error {
	// 设置活动标记为Y
	scheduler.ActiveFlag = timertypes.ActiveFlagYes

	tableName := (&timertypes.TimerScheduler{}).TableName()
	_, err := dao.db.Insert(ctx, tableName, scheduler, true)
	if err != nil {
		return fmt.Errorf("创建调度器失败: %w", err)
	}
	return nil
}

// UpdateScheduler 更新调度器
func (dao *TimerSchedulerDAO) UpdateScheduler(ctx context.Context, scheduler *timertypes.TimerScheduler) error {
	tableName := (&timertypes.TimerScheduler{}).TableName()
	where := "tenantId = ? AND schedulerId = ?"
	args := []interface{}{scheduler.TenantId, scheduler.SchedulerId}

	_, err := dao.db.Update(ctx, tableName, scheduler, where, args, true, true)
	if err != nil {
		return fmt.Errorf("更新调度器失败: %w", err)
	}
	return nil
}

// DeleteScheduler 删除调度器（软删除）
func (dao *TimerSchedulerDAO) DeleteScheduler(ctx context.Context, tenantId, schedulerId string) error {
	tableName := (&timertypes.TimerScheduler{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = 'N' WHERE tenantId = ? AND schedulerId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, schedulerId}, true)
	if err != nil {
		return fmt.Errorf("删除调度器失败: %w", err)
	}
	return nil
}

// GetRunningSchedulers 获取运行中的调度器
func (dao *TimerSchedulerDAO) GetRunningSchedulers(ctx context.Context, tenantId string) ([]timertypes.TimerScheduler, error) {
	onlyRunning := true
	activeFlag := timertypes.ActiveFlagYes

	query := &TimerSchedulerQuery{
		TenantId:       &tenantId,
		ActiveFlag:     &activeFlag,
		OnlyRunning:    &onlyRunning,
		OrderBy:        "schedulerName",
		OrderDirection: "ASC",
	}

	result, err := dao.QuerySchedulers(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Schedulers, nil
}

// GetHealthySchedulers 获取健康的调度器
func (dao *TimerSchedulerDAO) GetHealthySchedulers(ctx context.Context, tenantId string, healthCheckInterval time.Duration) ([]timertypes.TimerScheduler, error) {
	onlyHealthy := true
	activeFlag := timertypes.ActiveFlagYes

	query := &TimerSchedulerQuery{
		TenantId:            &tenantId,
		ActiveFlag:          &activeFlag,
		OnlyHealthy:         &onlyHealthy,
		HealthCheckInterval: &healthCheckInterval,
		OrderBy:             "schedulerName",
		OrderDirection:      "ASC",
	}

	result, err := dao.QuerySchedulers(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Schedulers, nil
}

// UpdateHeartbeat 更新调度器心跳
func (dao *TimerSchedulerDAO) UpdateHeartbeat(ctx context.Context, tenantId, schedulerId string) error {
	tableName := (&timertypes.TimerScheduler{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET lastHeartbeatTime = ? WHERE tenantId = ? AND schedulerId = ?", tableName)

	now := time.Now()
	_, err := dao.db.Exec(ctx, sql, []interface{}{now, tenantId, schedulerId}, true)
	if err != nil {
		return fmt.Errorf("更新调度器心跳失败: %w", err)
	}
	return nil
}
