package dao

import (
	"context"
	"fmt"
	"gateway/internal/types/timertypes"
	"gateway/pkg/database"
	"strings"
	"time"
)

// TimerTaskQuery 定义任务查询条件结构
type TimerTaskQuery struct {
	// 基础查询条件
	TaskId        *string `json:"taskId,omitempty"`
	TenantId      *string `json:"tenantId,omitempty"`
	TaskName      *string `json:"taskName,omitempty"`
	SchedulerId   *string `json:"schedulerId,omitempty"`
	SchedulerName *string `json:"schedulerName,omitempty"`

	// 状态查询
	TaskStatus     *int    `json:"taskStatus,omitempty"`
	TaskStatusList []int   `json:"taskStatusList,omitempty"`
	ActiveFlag     *string `json:"activeFlag,omitempty"`

	// 调度类型查询
	ScheduleType     *int  `json:"scheduleType,omitempty"`
	ScheduleTypeList []int `json:"scheduleTypeList,omitempty"`

	// 优先级查询
	TaskPriority *int `json:"taskPriority,omitempty"`
	MinPriority  *int `json:"minPriority,omitempty"`
	MaxPriority  *int `json:"maxPriority,omitempty"`

	// 执行器配置查询
	ExecutorType  *string `json:"executorType,omitempty"`
	ToolConfigId  *string `json:"toolConfigId,omitempty"`
	OperationType *string `json:"operationType,omitempty"`

	// 时间范围查询
	StartTimeFrom   *time.Time `json:"startTimeFrom,omitempty"`
	StartTimeTo     *time.Time `json:"startTimeTo,omitempty"`
	EndTimeFrom     *time.Time `json:"endTimeFrom,omitempty"`
	EndTimeTo       *time.Time `json:"endTimeTo,omitempty"`
	NextRunTimeFrom *time.Time `json:"nextRunTimeFrom,omitempty"`
	NextRunTimeTo   *time.Time `json:"nextRunTimeTo,omitempty"`
	LastRunTimeFrom *time.Time `json:"lastRunTimeFrom,omitempty"`
	LastRunTimeTo   *time.Time `json:"lastRunTimeTo,omitempty"`

	// 执行统计查询
	MinRunCount     *int64 `json:"minRunCount,omitempty"`
	MaxRunCount     *int64 `json:"maxRunCount,omitempty"`
	MinSuccessCount *int64 `json:"minSuccessCount,omitempty"`
	MaxSuccessCount *int64 `json:"maxSuccessCount,omitempty"`
	MinFailureCount *int64 `json:"minFailureCount,omitempty"`
	MaxFailureCount *int64 `json:"maxFailureCount,omitempty"`

	// 最后执行结果查询
	LastResultSuccess   *string `json:"lastResultSuccess,omitempty"`
	LastExecutionStatus *int    `json:"lastExecutionStatus,omitempty"`

	// 创建和修改时间查询
	AddTimeFrom  *time.Time `json:"addTimeFrom,omitempty"`
	AddTimeTo    *time.Time `json:"addTimeTo,omitempty"`
	EditTimeFrom *time.Time `json:"editTimeFrom,omitempty"`
	EditTimeTo   *time.Time `json:"editTimeTo,omitempty"`
	AddWho       *string    `json:"addWho,omitempty"`
	EditWho      *string    `json:"editWho,omitempty"`

	// 模糊查询
	TaskNameLike        *string `json:"taskNameLike,omitempty"`
	TaskDescriptionLike *string `json:"taskDescriptionLike,omitempty"`
	NoteTextLike        *string `json:"noteTextLike,omitempty"`

	// 排序和分页
	OrderBy        string `json:"orderBy,omitempty"`        // 排序字段
	OrderDirection string `json:"orderDirection,omitempty"` // ASC/DESC
	PageNum        int    `json:"pageNum,omitempty"`        // 页码，从1开始
	PageSize       int    `json:"pageSize,omitempty"`       // 每页大小
	Offset         int    `json:"offset,omitempty"`         // 偏移量
	Limit          int    `json:"limit,omitempty"`          // 限制数量
}

// CountResult count查询结果结构体
type CountResult struct {
	Count int64 `json:"count" db:"count"`
}

// TimerTaskQueryResult 查询结果结构
type TimerTaskQueryResult struct {
	Tasks      []timertypes.TimerTask `json:"tasks"`
	Total      int64                  `json:"total"`
	PageNum    int                    `json:"pageNum"`
	PageSize   int                    `json:"pageSize"`
	TotalPages int                    `json:"totalPages"`
}

// TimerTaskDAO 任务数据访问对象
type TimerTaskDAO struct {
	db database.Database
}

// NewTimerTaskDAO 创建任务DAO实例
func NewTimerTaskDAO(db database.Database) *TimerTaskDAO {
	return &TimerTaskDAO{db: db}
}

// BuildWhere 构建WHERE条件语句
func (q *TimerTaskQuery) BuildWhere() (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// 基础条件
	if q.TaskId != nil {
		conditions = append(conditions, "taskId = ?")
		args = append(args, *q.TaskId)
	}

	if q.TenantId != nil {
		conditions = append(conditions, "tenantId = ?")
		args = append(args, *q.TenantId)
	}

	if q.TaskName != nil {
		conditions = append(conditions, "taskName = ?")
		args = append(args, *q.TaskName)
	}

	if q.SchedulerId != nil {
		conditions = append(conditions, "schedulerId = ?")
		args = append(args, *q.SchedulerId)
	}

	if q.SchedulerName != nil {
		conditions = append(conditions, "schedulerName = ?")
		args = append(args, *q.SchedulerName)
	}

	// 状态条件
	if q.TaskStatus != nil {
		conditions = append(conditions, "taskStatus = ?")
		args = append(args, *q.TaskStatus)
	}

	if len(q.TaskStatusList) > 0 {
		placeholders := make([]string, len(q.TaskStatusList))
		for i, status := range q.TaskStatusList {
			placeholders[i] = "?"
			args = append(args, status)
		}
		conditions = append(conditions, fmt.Sprintf("taskStatus IN (%s)", strings.Join(placeholders, ",")))
	}

	if q.ActiveFlag != nil {
		conditions = append(conditions, "activeFlag = ?")
		args = append(args, *q.ActiveFlag)
	}

	// 调度类型条件
	if q.ScheduleType != nil {
		conditions = append(conditions, "scheduleType = ?")
		args = append(args, *q.ScheduleType)
	}

	if len(q.ScheduleTypeList) > 0 {
		placeholders := make([]string, len(q.ScheduleTypeList))
		for i, scheduleType := range q.ScheduleTypeList {
			placeholders[i] = "?"
			args = append(args, scheduleType)
		}
		conditions = append(conditions, fmt.Sprintf("scheduleType IN (%s)", strings.Join(placeholders, ",")))
	}

	// 优先级条件
	if q.TaskPriority != nil {
		conditions = append(conditions, "taskPriority = ?")
		args = append(args, *q.TaskPriority)
	}

	if q.MinPriority != nil {
		conditions = append(conditions, "taskPriority >= ?")
		args = append(args, *q.MinPriority)
	}

	if q.MaxPriority != nil {
		conditions = append(conditions, "taskPriority <= ?")
		args = append(args, *q.MaxPriority)
	}

	// 执行器配置条件
	if q.ExecutorType != nil {
		conditions = append(conditions, "executorType = ?")
		args = append(args, *q.ExecutorType)
	}

	if q.ToolConfigId != nil {
		conditions = append(conditions, "toolConfigId = ?")
		args = append(args, *q.ToolConfigId)
	}

	if q.OperationType != nil {
		conditions = append(conditions, "operationType = ?")
		args = append(args, *q.OperationType)
	}

	// 时间范围条件
	if q.StartTimeFrom != nil {
		conditions = append(conditions, "startTime >= ?")
		args = append(args, *q.StartTimeFrom)
	}

	if q.StartTimeTo != nil {
		conditions = append(conditions, "startTime <= ?")
		args = append(args, *q.StartTimeTo)
	}

	if q.EndTimeFrom != nil {
		conditions = append(conditions, "endTime >= ?")
		args = append(args, *q.EndTimeFrom)
	}

	if q.EndTimeTo != nil {
		conditions = append(conditions, "endTime <= ?")
		args = append(args, *q.EndTimeTo)
	}

	if q.NextRunTimeFrom != nil {
		conditions = append(conditions, "nextRunTime >= ?")
		args = append(args, *q.NextRunTimeFrom)
	}

	if q.NextRunTimeTo != nil {
		conditions = append(conditions, "nextRunTime <= ?")
		args = append(args, *q.NextRunTimeTo)
	}

	if q.LastRunTimeFrom != nil {
		conditions = append(conditions, "lastRunTime >= ?")
		args = append(args, *q.LastRunTimeFrom)
	}

	if q.LastRunTimeTo != nil {
		conditions = append(conditions, "lastRunTime <= ?")
		args = append(args, *q.LastRunTimeTo)
	}

	// 执行统计条件
	if q.MinRunCount != nil {
		conditions = append(conditions, "runCount >= ?")
		args = append(args, *q.MinRunCount)
	}

	if q.MaxRunCount != nil {
		conditions = append(conditions, "runCount <= ?")
		args = append(args, *q.MaxRunCount)
	}

	if q.MinSuccessCount != nil {
		conditions = append(conditions, "successCount >= ?")
		args = append(args, *q.MinSuccessCount)
	}

	if q.MaxSuccessCount != nil {
		conditions = append(conditions, "successCount <= ?")
		args = append(args, *q.MaxSuccessCount)
	}

	if q.MinFailureCount != nil {
		conditions = append(conditions, "failureCount >= ?")
		args = append(args, *q.MinFailureCount)
	}

	if q.MaxFailureCount != nil {
		conditions = append(conditions, "failureCount <= ?")
		args = append(args, *q.MaxFailureCount)
	}

	// 最后执行结果条件
	if q.LastResultSuccess != nil {
		conditions = append(conditions, "lastResultSuccess = ?")
		args = append(args, *q.LastResultSuccess)
	}

	if q.LastExecutionStatus != nil {
		conditions = append(conditions, "lastExecutionStatus = ?")
		args = append(args, *q.LastExecutionStatus)
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
	if q.TaskNameLike != nil {
		conditions = append(conditions, "taskName LIKE ?")
		args = append(args, "%"+*q.TaskNameLike+"%")
	}

	if q.TaskDescriptionLike != nil {
		conditions = append(conditions, "taskDescription LIKE ?")
		args = append(args, "%"+*q.TaskDescriptionLike+"%")
	}

	if q.NoteTextLike != nil {
		conditions = append(conditions, "noteText LIKE ?")
		args = append(args, "%"+*q.NoteTextLike+"%")
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	return where, args
}

// BuildOrderBy 构建ORDER BY语句
func (q *TimerTaskQuery) BuildOrderBy() string {
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
func (q *TimerTaskQuery) BuildPagination() string {
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

// QueryTasks 查询任务列表
func (dao *TimerTaskDAO) QueryTasks(ctx context.Context, query *TimerTaskQuery) (*TimerTaskQueryResult, error) {
	tableName := (&timertypes.TimerTask{}).TableName()

	// 构建查询条件
	where, args := query.BuildWhere()
	orderBy := query.BuildOrderBy()
	pagination := query.BuildPagination()

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) as count FROM %s %s", tableName, where)
	var countResult CountResult
	err := dao.db.QueryOne(ctx, &countResult, countSQL, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询任务总数失败: %w", err)
	}
	total := countResult.Count

	// 查询数据
	selectSQL := fmt.Sprintf("SELECT * FROM %s %s %s %s", tableName, where, orderBy, pagination)
	var tasks []timertypes.TimerTask
	err = dao.db.Query(ctx, &tasks, selectSQL, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %w", err)
	}

	// 计算分页信息
	result := &TimerTaskQueryResult{
		Tasks: tasks,
		Total: total,
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

// GetTaskById 根据ID获取任务
func (dao *TimerTaskDAO) GetTaskById(ctx context.Context, tenantId, taskId string) (*timertypes.TimerTask, error) {
	tableName := (&timertypes.TimerTask{}).TableName()
	sql := fmt.Sprintf("SELECT * FROM %s WHERE tenantId = ? AND taskId = ? AND activeFlag = ?", tableName)

	var task timertypes.TimerTask
	err := dao.db.QueryOne(ctx, &task, sql, []interface{}{tenantId, taskId, timertypes.ActiveFlagYes}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	return &task, nil
}

// CreateTask 创建任务
func (dao *TimerTaskDAO) CreateTask(ctx context.Context, task *timertypes.TimerTask) error {
	// 设置活动标记为Y
	task.ActiveFlag = timertypes.ActiveFlagYes

	tableName := (&timertypes.TimerTask{}).TableName()
	_, err := dao.db.Insert(ctx, tableName, task, true)
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}
	return nil
}

// UpdateTask 更新任务
func (dao *TimerTaskDAO) UpdateTask(ctx context.Context, task *timertypes.TimerTask) error {
	tableName := (&timertypes.TimerTask{}).TableName()
	where := "tenantId = ? AND taskId = ?"
	args := []interface{}{task.TenantId, task.TaskId}

	_, err := dao.db.Update(ctx, tableName, task, where, args, true)
	if err != nil {
		return fmt.Errorf("更新任务失败: %w", err)
	}
	return nil
}

// DeleteTask 删除任务（物理删除）
func (dao *TimerTaskDAO) DeleteTask(ctx context.Context, tenantId, taskId string) error {
	tableName := (&timertypes.TimerTask{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND taskId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, taskId}, true)
	if err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}
	return nil
}

// GetTasksByScheduler 根据调度器ID获取任务列表
func (dao *TimerTaskDAO) GetTasksByScheduler(ctx context.Context, tenantId, schedulerId string) ([]timertypes.TimerTask, error) {
	activeFlag := timertypes.ActiveFlagYes

	query := &TimerTaskQuery{
		TenantId:       &tenantId,
		SchedulerId:    &schedulerId,
		ActiveFlag:     &activeFlag,
		OrderBy:        "taskPriority",
		OrderDirection: "DESC",
	}

	result, err := dao.QueryTasks(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Tasks, nil
}

// GetPendingTasks 获取待执行的任务
func (dao *TimerTaskDAO) GetPendingTasks(ctx context.Context, tenantId string, maxCount int) ([]timertypes.TimerTask, error) {
	status := timertypes.TaskStatusPending
	activeFlag := timertypes.ActiveFlagYes
	now := time.Now()

	query := &TimerTaskQuery{
		TenantId:       &tenantId,
		TaskStatus:     &status,
		ActiveFlag:     &activeFlag,
		NextRunTimeTo:  &now,
		OrderBy:        "taskPriority",
		OrderDirection: "DESC",
		Limit:          maxCount,
	}

	result, err := dao.QueryTasks(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.Tasks, nil
}
