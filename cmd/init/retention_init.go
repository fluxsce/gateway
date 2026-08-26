package init

import (
	"context"
	"sync"

	"gateway/internal/retention"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

var (
	retentionSched *retention.Scheduler
	retentionOnce  sync.Once
)

// LoadSettings 把 HUB_SYS_SETTING 灌入进程缓存。网关转发前调用，避免 ${VAR} 展开时缓存仍为空。
func LoadSettings(ctx context.Context, db database.Database) {
	retention.ReloadSettings(ctx, db)
}

// StartRetention 启动租户级数据生命周期调度。网关访问日志仍由实例级 cleaner 处理。
func StartRetention(ctx context.Context, db database.Database) {
	if db == nil {
		return
	}
	retentionOnce.Do(func() {
		retention.ReloadSettings(ctx, db)
		retentionSched = retention.NewScheduler(db, config.GetNodeId(), retention.NewJobs(db))
		retentionSched.Start(ctx)
		logger.Info("生命周期调度初始化完成")
	})
}

// StopRetention 停止生命周期调度。
func StopRetention() {
	if retentionSched != nil {
		retentionSched.Stop()
	}
}
