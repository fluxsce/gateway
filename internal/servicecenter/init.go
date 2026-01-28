package servicecenter

import (
	"context"
	"fmt"

	"gateway/internal/servicecenter/manager"
	"gateway/internal/servicecenter/server"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// ServiceCenter 服务中心全局实例
var ServiceCenter *manager.ServiceCenterManager

// Init 初始化服务中心
func Init(ctx context.Context, db database.Database) (*manager.ServiceCenterManager, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库实例不能为空")
	}

	// 创建服务中心管理器
	ServiceCenter = manager.NewServiceCenterManager(db)

	return ServiceCenter, nil
}

// StartAll 启动所有服务中心实例
// 默认使用 "default" 租户，加载所有环境的实例配置
func StartAll(ctx context.Context) error {
	if ServiceCenter == nil {
		return fmt.Errorf("服务中心管理器未初始化，请先调用 Init()")
	}

	// 默认使用 "default" 租户
	tenantId := "default"

	// 从数据库加载所有实例配置并创建 Server（所有环境）
	if err := ServiceCenter.LoadAllInstancesFromDB(ctx, tenantId); err != nil {
		return fmt.Errorf("加载实例配置失败: %w", err)
	}

	// 启动所有实例
	var startErrors []string
	count := 0
	ServiceCenter.ForEachInstance(func(instanceName string, srv *server.Server) error {
		if !srv.IsRunning() {
			// 从 Server 获取配置
			config := srv.GetConfig()
			if config == nil {
				startErrors = append(startErrors, fmt.Sprintf("%s: 无法获取配置", instanceName))
				return nil
			}

			// 获取环境信息
			env := config.Environment
			if env == "" {
				env = "dev" // 默认环境
			}

			if err := ServiceCenter.StartInstance(ctx, tenantId, instanceName, env); err != nil {
				startErrors = append(startErrors, fmt.Sprintf("%s: %v", instanceName, err))
				logger.Error("启动服务中心实例失败", err, "instanceName", instanceName)
			} else {
				count++
			}
		}
		return nil
	})

	if len(startErrors) > 0 {
		return fmt.Errorf("部分服务中心实例启动失败: %v", startErrors)
	}

	logger.Info("所有服务中心实例启动成功", "count", count)
	return nil
}

// StopAll 停止所有服务中心实例
func StopAll(ctx context.Context) error {
	if ServiceCenter == nil {
		return fmt.Errorf("服务中心管理器未初始化")
	}

	var errors []string
	count := 0

	ServiceCenter.ForEachInstance(func(instanceName string, srv *server.Server) error {
		if srv.IsRunning() {
			if err := ServiceCenter.StopInstance(ctx, instanceName); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", instanceName, err))
				logger.Error("停止服务中心实例失败", err, "instanceName", instanceName)
			} else {
				count++
			}
		}
		return nil
	})

	if len(errors) > 0 {
		return fmt.Errorf("部分服务中心实例停止失败: %v", errors)
	}

	logger.Info("所有服务中心实例停止成功", "count", count)
	return nil
}

// GetManager 获取服务中心管理器
func GetManager() *manager.ServiceCenterManager {
	return ServiceCenter
}
