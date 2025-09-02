package init

import (
	"context"
	"fmt"

	"gateway/internal/registry/core"
	"gateway/internal/registry/manager"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	TenantId   string `json:"tenantId"`   // 租户ID
	ActiveFlag string `json:"activeFlag"` // 活动标记
}

// InitializeRegistry 初始化注册中心
// 这是主要的入口函数，用于启动注册中心初始化
// 返回管理器实例和初始化是否成功的布尔值
func InitializeRegistry(ctx context.Context, db database.Database, config *RegistryConfig) (core.Manager, bool) {
	if db == nil {
		logger.ErrorWithTrace(ctx, "初始化注册中心失败：数据库连接不能为空")
		return nil, false
	}

	// 创建初始化器
	initializer := NewRegistryInitializer(db)

	// 应用配置
	if config != nil {
		if config.TenantId != "" {
			initializer.SetTenantId(config.TenantId)
		}
		if config.ActiveFlag != "" {
			initializer.SetActiveFlag(config.ActiveFlag)
		}
	}

	// 执行初始化
	logger.InfoWithTrace(ctx, "开始初始化注册中心")
	mgr, success := initializer.Initialize(ctx)
	if !success {
		logger.WarnWithTrace(ctx, "注册中心初始化未完成，将以不使用注册中心模式运行")
		return mgr, false
	}

	return mgr, true
}

// GetRegistryManager 获取注册中心管理器
// 如果管理器未初始化，返回nil
func GetRegistryManager() core.Manager {
	// 直接调用manager包中的GetInstance方法
	return manager.GetInstance()
}

// StartRegistry 启动注册中心服务
// 如果注册中心未初始化，返回错误
func StartRegistry(ctx context.Context) error {
	mgr := GetRegistryManager()
	if mgr == nil {
		return fmt.Errorf("注册中心未初始化")
	}

	return mgr.Start(ctx)
}

// StopRegistry 停止注册中心服务
func StopRegistry(ctx context.Context) error {
	mgr := GetRegistryManager()
	if mgr == nil {
		return nil // 未初始化，无需停止
	}

	return mgr.Stop(ctx)
}

// IsRegistryReady 检查注册中心是否就绪
func IsRegistryReady() bool {
	mgr := GetRegistryManager()
	if mgr == nil {
		return false
	}

	return mgr.IsReady()
}
