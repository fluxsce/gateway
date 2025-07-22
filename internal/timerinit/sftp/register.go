package sftp

import (
	"context"
	"fmt"

	"gateway/internal/timerinit/common/dao"
	"gateway/internal/timerinit/common/taskinit"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// RegisterSFTPTasks 注册SFTP定时任务
// 这是SFTP定时任务初始化的统一入口函数
// 参数:
//
//	db: 数据库连接实例
//	tenantIds: 需要初始化的租户ID列表，如果为空则初始化所有租户
//
// 返回:
//
//	error: 初始化失败时返回错误信息
func RegisterSFTPTasks(ctx context.Context, db database.Database, tenantIds ...string) error {
	logger.Info("开始注册SFTP定时任务")

	// 创建DAO管理器
	daoManager := dao.NewDAOManager(db)

	// 创建SFTP执行器工厂
	factory := NewSFTPExecutorFactory(daoManager)

	// 创建通用任务注册器
	register := taskinit.NewTaskRegister(db)

	// 使用通用注册器进行任务注册
	if err := register.RegisterTasks(ctx, factory, tenantIds...); err != nil {
		return err
	}

	logger.Info("SFTP定时任务注册完成")
	return nil
}

// RegisterSFTPTasksForTenant 为指定租户注册SFTP定时任务
// 这是针对单个租户的SFTP定时任务初始化函数
// 参数:
//
//	db: 数据库连接实例
//	tenantId: 租户ID
//
// 返回:
//
//	error: 初始化失败时返回错误信息
func RegisterSFTPTasksForTenant(ctx context.Context, db database.Database, tenantId string) error {
	logger.Info("开始为租户注册SFTP定时任务", "tenantId", tenantId)

	// 创建DAO管理器
	daoManager := dao.NewDAOManager(db)

	// 创建SFTP执行器工厂
	factory := NewSFTPExecutorFactory(daoManager)

	// 创建通用任务注册器
	register := taskinit.NewTaskRegister(db)

	// 使用通用注册器进行任务注册
	if err := register.RegisterTasksForTenant(ctx, factory, tenantId); err != nil {
		return err
	}

	logger.Info("租户SFTP定时任务注册完成", "tenantId", tenantId)
	return nil
}

// RegisterSFTPTaskById 根据任务ID注册单个SFTP定时任务
// 这是针对单个任务的SFTP定时任务注册函数
// 参数:
//
//	ctx: 上下文对象
//	db: 数据库连接实例
//	tenantId: 租户ID
//	taskId: 任务ID
//
// 返回:
//
//	error: 注册失败时返回错误信息
func RegisterSFTPTaskById(ctx context.Context, db database.Database, tenantId, taskId string) error {
	logger.Info("开始根据任务ID注册SFTP定时任务", "tenantId", tenantId, "taskId", taskId)

	// 参数验证
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if taskId == "" {
		return fmt.Errorf("任务ID不能为空")
	}

	// 创建DAO管理器
	daoManager := dao.NewDAOManager(db)

	// 检查数据库连接
	if err := daoManager.HealthCheck(ctx); err != nil {
		return fmt.Errorf("数据库连接检查失败: %w", err)
	}

	// 创建SFTP执行器工厂
	factory := NewSFTPExecutorFactory(daoManager)

	// 查询指定任务ID的任务
	executorType := factory.GetExecutorType()
	activeFlag := "Y"

	query := &dao.TimerTaskQuery{
		TenantId:     &tenantId,
		TaskId:       &taskId,
		ExecutorType: &executorType,
		ActiveFlag:   &activeFlag,
	}

	// 执行查询
	result, err := daoManager.GetTaskDAO().QueryTasks(ctx, query)
	if err != nil {
		return fmt.Errorf("查询指定任务失败: %w", err)
	}

	if len(result.Tasks) == 0 {
		logger.Info("未找到指定的任务", "tenantId", tenantId, "taskId", taskId, "executorType", executorType)
		return fmt.Errorf("未找到指定的任务: taskId=%s", taskId)
	}

	if len(result.Tasks) > 1 {
		logger.Warn("找到多个相同ID的任务", "tenantId", tenantId, "taskId", taskId, "taskCount", len(result.Tasks))
	}

	// 获取第一个任务（应该只有一个）
	task := &result.Tasks[0]
	logger.Info("找到指定的任务", "tenantId", tenantId, "taskId", taskId, "taskName", task.TaskName)

	// 创建基础任务初始化器
	initializer := taskinit.NewBaseTaskInitializer(daoManager, factory)

	// 使用基础初始化器的单个任务初始化方法
	if err := initializer.InitializeSingleTask(ctx, task); err != nil {
		logger.Error("根据任务ID注册SFTP任务失败", "error", err, "tenantId", tenantId, "taskId", taskId)
		return fmt.Errorf("注册SFTP任务失败: %w", err)
	}

	logger.Info("根据任务ID注册SFTP定时任务完成", "tenantId", tenantId, "taskId", taskId)
	return nil
}

// ReloadSFTPTasks 重新加载SFTP定时任务
// 停止现有的SFTP任务并重新从数据库加载配置
// 参数:
//
//	db: 数据库连接实例
//	tenantIds: 需要重新加载的租户ID列表，如果为空则重新加载所有租户
//
// 返回:
//
//	error: 重新加载失败时返回错误信息
func ReloadSFTPTasks(ctx context.Context, db database.Database, tenantIds ...string) error {
	logger.Info("开始重新加载SFTP定时任务")

	// 创建DAO管理器
	daoManager := dao.NewDAOManager(db)

	// 创建SFTP执行器工厂
	factory := NewSFTPExecutorFactory(daoManager)

	// 创建通用任务注册器
	register := taskinit.NewTaskRegister(db)

	// 使用通用注册器进行任务重新加载
	return register.ReloadTasks(ctx, factory, tenantIds...)
}

// StopSFTPTasks 停止SFTP定时任务
// 停止指定租户的所有SFTP定时任务
// 参数:
//
//	db: 数据库连接实例
//	tenantIds: 需要停止的租户ID列表，如果为空则停止所有租户的SFTP任务
//
// 返回:
//
//	error: 停止失败时返回错误信息
func StopSFTPTasks(ctx context.Context, db database.Database, tenantIds ...string) error {
	logger.Info("开始停止SFTP定时任务")

	// 创建通用任务注册器
	register := taskinit.NewTaskRegister(db)

	// 使用通用注册器进行任务停止，使用正确的执行器类型
	if err := register.StopTasks(ctx, "SFTP_TRANSFER", tenantIds...); err != nil {
		return fmt.Errorf("停止SFTP任务失败: %w", err)
	}

	logger.Info("SFTP定时任务停止完成")
	return nil
}

// GetSFTPTaskStatus 获取SFTP定时任务状态
// 获取指定租户的SFTP定时任务运行状态统计信息
// 参数:
//
//	db: 数据库连接实例
//	tenantId: 租户ID
//
// 返回:
//
//	map[string]interface{}: 状态统计信息
//	error: 获取失败时返回错误信息
func GetSFTPTaskStatus(ctx context.Context, db database.Database, tenantId string) (map[string]interface{}, error) {
	// 创建通用任务注册器
	register := taskinit.NewTaskRegister(db)

	// 使用通用注册器获取任务状态
	return register.GetTaskStatus(ctx, "sftp", tenantId)
}
