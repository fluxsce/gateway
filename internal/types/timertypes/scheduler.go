package timertypes

import (
	"time"
)

// TimerScheduler 定义调度器配置，对应数据库表 HUB_TIMER_SCHEDULER
type TimerScheduler struct {
	// 主键信息
	SchedulerId       string `json:"schedulerId" db:"schedulerId"`
	TenantId          string `json:"tenantId" db:"tenantId"`
	
	// 基础信息
	SchedulerName     string `json:"schedulerName" db:"schedulerName"`
	SchedulerInstanceId string `json:"schedulerInstanceId" db:"schedulerInstanceId"`
	
	// 调度器配置
	MaxWorkers        int    `json:"maxWorkers" db:"maxWorkers"`
	QueueSize         int    `json:"queueSize" db:"queueSize"`
	DefaultTimeoutSeconds int64 `json:"defaultTimeoutSeconds" db:"defaultTimeoutSeconds"`
	DefaultRetries    int    `json:"defaultRetries" db:"defaultRetries"`
	
	// 调度器状态
	SchedulerStatus   int    `json:"schedulerStatus" db:"schedulerStatus"`
	LastStartTime     *time.Time `json:"lastStartTime" db:"lastStartTime"`
	LastStopTime      *time.Time `json:"lastStopTime" db:"lastStopTime"`
	
	// 服务器信息
	ServerName        *string `json:"serverName" db:"serverName"`
	ServerIp          *string `json:"serverIp" db:"serverIp"`
	ServerPort        *int    `json:"serverPort" db:"serverPort"`
	
	// 监控信息
	TotalTaskCount    int    `json:"totalTaskCount" db:"totalTaskCount"`
	RunningTaskCount  int    `json:"runningTaskCount" db:"runningTaskCount"`
	LastHeartbeatTime *time.Time `json:"lastHeartbeatTime" db:"lastHeartbeatTime"`
	
	// 通用字段
	AddTime           time.Time `json:"addTime" db:"addTime"`
	AddWho            string    `json:"addWho" db:"addWho"`
	EditTime          time.Time `json:"editTime" db:"editTime"`
	EditWho           string    `json:"editWho" db:"editWho"`
	OprSeqFlag        string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion    int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag        string    `json:"activeFlag" db:"activeFlag"`
	NoteText          *string   `json:"noteText" db:"noteText"`
	Reserved1         *string   `json:"reserved1" db:"reserved1"`
	Reserved2         *string   `json:"reserved2" db:"reserved2"`
	Reserved3         *string   `json:"reserved3" db:"reserved3"`
}

// TableName 返回数据库表名
func (TimerScheduler) TableName() string {
	return "HUB_TIMER_SCHEDULER"
}

// GetStatusName 获取调度器状态名称
func (s *TimerScheduler) GetStatusName() string {
	switch s.SchedulerStatus {
	case SchedulerStatusStopped:
		return "已停止"
	case SchedulerStatusRunning:
		return "运行中"
	case SchedulerStatusPaused:
		return "已暂停"
	default:
		return "未知状态"
	}
}

// IsActive 检查调度器是否处于活动状态
func (s *TimerScheduler) IsActive() bool {
	return s.ActiveFlag == ActiveFlagYes
}

// IsRunning 检查调度器是否正在运行
func (s *TimerScheduler) IsRunning() bool {
	return s.SchedulerStatus == SchedulerStatusRunning
}

// CanStart 检查调度器是否可以启动
func (s *TimerScheduler) CanStart() bool {
	return s.IsActive() && s.SchedulerStatus != SchedulerStatusRunning
}

// CanStop 检查调度器是否可以停止
func (s *TimerScheduler) CanStop() bool {
	return s.IsActive() && s.SchedulerStatus == SchedulerStatusRunning
}

// Start 启动调度器
func (s *TimerScheduler) Start() {
	s.SchedulerStatus = SchedulerStatusRunning
	now := time.Now()
	s.LastStartTime = &now
}

// Stop 停止调度器
func (s *TimerScheduler) Stop() {
	s.SchedulerStatus = SchedulerStatusStopped
	now := time.Now()
	s.LastStopTime = &now
}

// Pause 暂停调度器
func (s *TimerScheduler) Pause() {
	s.SchedulerStatus = SchedulerStatusPaused
}

// UpdateHeartbeat 更新心跳时间
func (s *TimerScheduler) UpdateHeartbeat() {
	now := time.Now()
	s.LastHeartbeatTime = &now
}

// IsHealthy 检查调度器是否健康
func (s *TimerScheduler) IsHealthy(maxInterval time.Duration) bool {
	if s.LastHeartbeatTime == nil {
		return false
	}
	return time.Since(*s.LastHeartbeatTime) <= maxInterval
}

// IncrementRunningTask 增加运行中任务计数
func (s *TimerScheduler) IncrementRunningTask() {
	s.RunningTaskCount++
}

// DecrementRunningTask 减少运行中任务计数
func (s *TimerScheduler) DecrementRunningTask() {
	if s.RunningTaskCount > 0 {
		s.RunningTaskCount--
	}
}

// UpdateTaskCounts 更新任务计数
func (s *TimerScheduler) UpdateTaskCounts(total int, running int) {
	s.TotalTaskCount = total
	s.RunningTaskCount = running
} 