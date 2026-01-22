package service

import (
	"context"
	"fmt"

	"gateway/internal/alert/types"
	"gateway/pkg/logger"
)

// logWriteWorker 日志写入 worker（从队列读取日志并添加到批量缓冲区）
func (s *AlertServiceImpl) logWriteWorker() {
	defer s.wg.Done()

	logger.Info("告警日志写入 worker 启动", "batchSize", s.logBatchSize)

	for {
		select {
		case <-s.ctx.Done():
			logger.Info("告警日志写入 worker 停止")
			return
		case alertLog, ok := <-s.logQueue:
			if !ok {
				logger.Info("告警日志队列已关闭")
				return
			}

			// 添加到批量缓冲区
			if err := s.addToBatch(alertLog); err != nil {
				logger.Error("添加告警日志到批量缓冲区失败", "error", err, "alertLogId", alertLog.AlertLogId)
			}
		}
	}
}

// addToBatch 将日志添加到批量缓冲区
func (s *AlertServiceImpl) addToBatch(alertLog *types.AlertLog) error {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()

	s.batchBuffer = append(s.batchBuffer, alertLog)

	// 如果缓冲区满了，立即刷新
	if len(s.batchBuffer) >= s.logBatchSize {
		ctx := context.Background()
		if err := s.batchWriteDirectly(ctx, s.batchBuffer); err != nil {
			logger.Error("批量写入告警日志失败", "error", err, "count", len(s.batchBuffer))
			return err
		}
		s.batchBuffer = s.batchBuffer[:0]
	}

	return nil
}

// batchWriteDirectly 直接批量写入日志到数据库
func (s *AlertServiceImpl) batchWriteDirectly(ctx context.Context, logs []*types.AlertLog) error {
	if len(logs) == 0 {
		return nil
	}

	if err := s.logDAO.BatchSaveLogs(ctx, logs); err != nil {
		return fmt.Errorf("批量写入告警日志失败: %w", err)
	}

	return nil
}

// flushBatch 刷新批量缓冲区
func (s *AlertServiceImpl) flushBatch() error {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()

	if len(s.batchBuffer) == 0 {
		return nil
	}

	// 保存计数用于日志
	count := len(s.batchBuffer)

	// 执行批量写入
	ctx := context.Background()
	err := s.batchWriteDirectly(ctx, s.batchBuffer)
	if err != nil {
		logger.Error("刷新批量缓冲区失败", "error", err, "count", count)
		return err
	}

	// 清空缓冲区
	s.batchBuffer = s.batchBuffer[:0]
	logger.Debug("刷新批量缓冲区", "count", count)

	return nil
}

// flushWorker 定时刷新 worker（定期刷新批量缓冲区）
func (s *AlertServiceImpl) flushWorker() {
	defer s.wg.Done()

	logger.Info("告警日志刷新 worker 启动", "interval", s.logFlushInterval)

	for {
		select {
		case <-s.ctx.Done():
			logger.Info("告警日志刷新 worker 停止")
			return
		case <-s.flushTicker.C:
			if err := s.flushBatch(); err != nil {
				logger.Error("定时刷新批量缓冲区失败", "error", err)
			}
		}
	}
}
