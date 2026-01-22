package service

import (
	"context"
	"time"

	"gateway/pkg/logger"
)

// cleanupWorker 清理 worker（定期清理旧日志）
func (s *AlertServiceImpl) cleanupWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	logger.Info("告警日志清理 worker 启动", "interval", s.cleanupInterval, "retentionHours", s.logRetentionHours)

	for {
		select {
		case <-s.ctx.Done():
			logger.Info("告警日志清理 worker 停止")
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup 执行清理操作
func (s *AlertServiceImpl) cleanup() {
	ctx := context.Background()
	beforeTime := time.Now().Add(-time.Duration(s.logRetentionHours) * time.Hour)

	affected, err := s.logDAO.CleanupOldLogs(ctx, s.tenantId, beforeTime)
	if err != nil {
		logger.Error("清理旧告警日志失败", "error", err)
		return
	}

	if affected > 0 {
		logger.Info("清理旧告警日志", "count", affected, "beforeTime", beforeTime)
	}
}
