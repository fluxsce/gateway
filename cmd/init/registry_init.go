package init

import (
	"context"
	"fmt"
	registryinit "gateway/internal/registry/init" // 导入注册中心初始化包
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// InitRegistry 初始化注册中心
// 这是注册中心初始化的统一入口函数，负责启动服务注册与发现功能
// 参数:
//
//	ctx: 上下文对象
//	db: 数据库连接实例
//	tenantId: 租户ID，如果为空则使用默认租户
//
// 返回:
//
//	bool: true表示注册中心初始化并启动成功，false表示未使用注册中心
func InitRegistry(ctx context.Context, db database.Database, tenantId string) bool {
	logger.InfoWithTrace(ctx, "开始初始化注册中心")

	// 验证数据库连接
	if db == nil {
		logger.ErrorWithTrace(ctx, "初始化注册中心失败：数据库连接不能为空")
		return false
	}
	// 准备注册中心配置
	registryConfig := &registryinit.RegistryConfig{
		TenantId:   tenantId,
		ActiveFlag: "Y", // 默认只加载活动状态的服务
	}

	// 初始化注册中心
	_, success := registryinit.InitializeRegistry(ctx, db, registryConfig)
	if !success {
		logger.WarnWithTrace(ctx, "注册中心未初始化，系统将以不使用注册中心模式运行")
		return false
	}

	// 启动注册中心服务
	if err := registryinit.StartRegistry(ctx); err != nil {
		logger.ErrorWithTrace(ctx, "启动注册中心服务失败", "error", err)
		return false
	}

	logger.InfoWithTrace(ctx, "注册中心初始化完成并成功启动")
	return true
}

// StopRegistry 停止注册中心服务
// 停止所有注册中心相关服务，包括服务发现、健康检查等
// 如果注册中心未初始化成功，则不执行任何操作
// 参数:
//
//	ctx: 上下文对象
//
// 返回:
//
//	error: 停止失败时返回错误信息
func StopRegistry(ctx context.Context) error {
	// 检查注册中心是否已初始化
	if !GetRegistryStatus() {
		logger.InfoWithTrace(ctx, "注册中心未初始化或未启用，无需停止")
		return nil
	}

	logger.InfoWithTrace(ctx, "开始停止注册中心服务")

	// 调用内部包的StopRegistry方法
	if err := registryinit.StopRegistry(ctx); err != nil {
		return fmt.Errorf("停止注册中心服务失败: %w", err)
	}

	logger.InfoWithTrace(ctx, "注册中心服务已成功停止")
	return nil
}

// GetRegistryStatus 获取注册中心状态
// 检查注册中心是否正常运行
// 返回:
//
//	bool: true表示注册中心正常运行，false表示未初始化或已停止
func GetRegistryStatus() bool {
	return registryinit.IsRegistryReady()
}

// InitRegistryWithConfig 带配置检查的注册中心初始化
// 整合了配置检查、参数读取和初始化流程
// 参数:
//   - ctx: 上下文对象
//   - db: 数据库连接实例
//
// 返回:
//   - error: 初始化失败时返回错误信息（注意：注册中心初始化失败不会返回error，而是记录警告）
func InitRegistryWithConfig(ctx context.Context, db database.Database) error {
	// 检查是否启用注册中心
	if !config.GetBool("app.registry.enabled", true) {
		logger.Info("注册中心未启用，跳过初始化")
		return nil
	}

	// 获取默认租户ID
	tenantId := config.GetString("app.registry.tenant_id", "default")

	// 初始化注册中心 - 现在返回的是布尔值，表示是否成功初始化
	success := InitRegistry(ctx, db, tenantId)
	if !success {
		logger.Warn("注册中心未能成功初始化，系统将以不使用注册中心模式运行")
		// 注意：这里不将其视为致命错误，而是允许系统继续运行
	} else {
		logger.Info("注册中心初始化并启动成功")
	}

	return nil
}
