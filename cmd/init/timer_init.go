package init

import (
	"context"
	"fmt"
	"gateway/internal/timerinit/sftp"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/timer"
)

// InitAllTimerTasks 初始化所有定时任务
// 这是定时任务初始化的统一入口函数，会依次初始化各个模块的定时任务
// 参数:
//
//	ctx: 上下文对象
//	db: 数据库连接实例
//	tenantIds: 需要初始化的租户ID列表，如果为空则初始化所有租户
//
// 返回:
//
//	error: 初始化失败时返回错误信息
func InitAllTimerTasks(ctx context.Context, db database.Database, tenantIds ...string) error {
	logger.Info("开始初始化所有定时任务")

	// 验证数据库连接
	if db == nil {
		return fmt.Errorf("数据库连接不能为空")
	}
	if config.GetBool("app.timer.sftp.enabled", false) {
		// 初始化SFTP定时任务
		if err := initSFTPTasks(ctx, db, tenantIds...); err != nil {
			return fmt.Errorf("初始化SFTP定时任务失败: %w", err)
		}
	}

	// 这里可以添加其他类型的定时任务初始化
	// 例如：SSH任务、FTP任务等
	// if err := initSSHTasks(ctx, db, tenantIds...); err != nil {
	//     return fmt.Errorf("初始化SSH定时任务失败: %w", err)
	// }

	logger.Info("所有定时任务初始化完成")
	return nil
}

// StopAllTimerTasks 停止所有定时任务
// 停止指定租户的所有定时任务
// 参数:
//
//	ctx: 上下文对象
//	db: 数据库连接实例
//	tenantIds: 需要停止的租户ID列表，如果为空则停止所有租户的任务
//
// 返回:
//
//	error: 停止失败时返回错误信息
func StopAllTimerTasks() error {
	logger.Info("开始停止所有定时任务")

	timer.GetTimerPool().StopAllSchedulers()

	logger.Info("所有定时任务停止完成")
	return nil
}

// initSFTPTasks 初始化SFTP定时任务
// 内部函数，用于初始化SFTP定时任务
func initSFTPTasks(ctx context.Context, db database.Database, tenantIds ...string) error {
	logger.Info("开始初始化SFTP定时任务")

	if err := sftp.RegisterSFTPTasks(ctx, db, tenantIds...); err != nil {
		return err
	}

	logger.Info("SFTP定时任务初始化完成")
	return nil
}

// 预留的SSH任务初始化函数，当SSH模块实现后可以启用
// func initSSHTasks(ctx context.Context, db database.Database, tenantIds ...string) error {
//     logger.Info("开始初始化SSH定时任务")
//
//     if err := ssh.RegisterSSHTasks(ctx, db, tenantIds...); err != nil {
//         return err
//     }
//
//     logger.Info("SSH定时任务初始化完成")
//     return nil
// }
