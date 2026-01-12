package gatewayapp

import (
	"context"
	"path/filepath"
	"strings"

	"gateway/internal/gateway/bootstrap"
	gatewayconfig "gateway/internal/gateway/config"
	"gateway/internal/gateway/loader"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
)

// 版本信息
const (
	Version = "1.0.0"
)

// GatewayApp 网关应用管理器
type GatewayApp struct {
	pool bootstrap.GatewayPool
	db   database.Database
}

// NewGatewayApp 创建网关应用实例
func NewGatewayApp() *GatewayApp {
	return &GatewayApp{
		pool: bootstrap.GetGlobalPool(),
	}
}

// Init 初始化网关应用
func (app *GatewayApp) Init(db database.Database) error {
	// 设置数据库连接
	app.db = db

	// 检查是否启用网关
	if !config.GetBool("app.gateway.enabled", false) {
		logger.Info("网关应用已禁用，跳过初始化")
		return nil
	}

	logger.Info("初始化 Gateway API 网关...", "version", Version)

	// 加载网关配置并创建实例
	if err := app.loadGatewayFromConfig(); err != nil {
		return huberrors.WrapError(err, "加载网关配置失败")
	}

	logger.Info("网关应用初始化完成")
	return nil
}

// Start 启动所有网关实例
func (app *GatewayApp) Start() error {
	// 检查是否启用网关
	if !config.GetBool("app.gateway.enabled", false) {
		logger.Info("网关应用已禁用，跳过启动")
		return nil
	}

	logger.Info("启动所有网关实例...")

	// 启动连接池中的所有网关实例
	if err := app.pool.StartAll(); err != nil {
		return huberrors.WrapError(err, "启动网关实例失败")
	}

	// 记录启动状态
	runningCount := len(app.pool.GetRunningGateways())
	totalCount := app.pool.Count()

	logger.Info("网关启动完成",
		"version", Version,
		"total_instances", totalCount,
		"running_instances", runningCount)

	return nil
}

// Stop 停止所有网关实例
func (app *GatewayApp) Stop() error {
	logger.Info("停止所有网关实例...")

	// 停止连接池中的所有网关实例
	if err := app.pool.StopAll(); err != nil {
		return huberrors.WrapError(err, "停止网关实例失败")
	}

	logger.Info("所有网关实例已停止")
	return nil
}

// GetStatus 获取网关状态
func (app *GatewayApp) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"version":           Version,
		"enabled":           config.GetBool("app.gateway.enabled", false),
		"total_instances":   app.pool.Count(),
		"running_instances": len(app.pool.GetRunningGateways()),
		"instance_ids":      app.pool.GetInstanceIDs(),
	}
	return status
}

// loadGatewayFromConfig 从配置加载网关实例
func (app *GatewayApp) loadGatewayFromConfig() error {
	// 获取配置源
	configSource := config.GetString("app.gateway.configSource", "yaml")

	switch strings.ToLower(configSource) {
	case "database":
		return app.loadFromDatabase()
	case "yaml", "json":
		return app.loadFromFile()
	default:
		return huberrors.NewError("不支持的配置源: %s", configSource)
	}
}

// loadFromFile 从文件加载网关配置
func (app *GatewayApp) loadFromFile() error {
	// 使用统一的配置文件路径构建方式
	configFile := config.GetString("app.gateway.config_file", "")
	if configFile == "" {
		// 如果配置中没有指定文件路径，使用默认路径
		configFile = config.GetConfigPath("gateway.yaml")
	} else {
		// 如果指定了相对路径，且不是绝对路径，则基于配置目录构建
		if !filepath.IsAbs(configFile) && !strings.HasPrefix(configFile, "./") && !strings.HasPrefix(configFile, "../") {
			// 如果是纯文件名，则基于配置目录构建
			configFile = config.GetConfigPath(configFile)
		}
	}

	logger.Info("从文件加载网关配置", "file", configFile)

	// 选择配置加载器
	var configLoader interface {
		LoadConfig(string) (*gatewayconfig.GatewayConfig, error)
	}

	if strings.HasSuffix(strings.ToLower(configFile), ".json") {
		configLoader = loader.NewJSONConfigLoader()
	} else {
		configLoader = loader.NewYAMLConfigLoader()
	}

	// 加载配置
	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return huberrors.WrapError(err, "加载配置文件失败: %s", configFile)
	}

	// 创建网关实例
	return app.createGatewayInstance(cfg, configFile)
}

// loadFromDatabase 从数据库加载网关配置
func (app *GatewayApp) loadFromDatabase() error {
	logger.Info("从数据库加载网关配置")

	// 直接从数据库查询所有活动状态的网关实例
	var gatewayInstances []struct {
		TenantID          string `db:"tenantId"`
		GatewayInstanceID string `db:"gatewayInstanceId"`
	}

	// 查询所有活动状态的网关实例
	query := `SELECT tenantId, gatewayInstanceId 
			  FROM HUB_GW_INSTANCE 
			  WHERE activeFlag = 'Y'`

	// 执行查询
	ctx := context.Background()
	err := app.db.Query(ctx, &gatewayInstances, query, nil, true)
	if err != nil {
		return huberrors.WrapError(err, "查询网关实例失败")
	}

	// 检查是否找到实例
	if len(gatewayInstances) == 0 {
		logger.Warn("未找到活动的网关实例，跳过网关配置加载")
		return nil
	}

	// 记录找到的实例数量
	logger.Info("从数据库加载网关实例", "count", len(gatewayInstances))

	// 为每个实例加载配置并创建网关实例
	for _, instance := range gatewayInstances {
		// 创建数据库配置加载器，使用实例对应的租户ID
		dbLoader := loader.NewDatabaseConfigLoader(app.db, instance.TenantID)

		// 加载网关配置
		cfg, err := dbLoader.LoadGatewayConfig(instance.GatewayInstanceID)
		if err != nil {
			logger.Error("加载网关配置失败", err,
				"tenantId", instance.TenantID,
				"instanceId", instance.GatewayInstanceID)
			continue
		}

		// 创建网关实例
		if err := app.createGatewayInstance(cfg, "database:"+instance.GatewayInstanceID); err != nil {
			logger.Error("创建网关实例失败", err,
				"tenantId", instance.TenantID,
				"instanceId", instance.GatewayInstanceID)
			continue
		}

		logger.Info("网关实例配置加载成功",
			"tenantId", instance.TenantID,
			"instanceId", instance.GatewayInstanceID)
	}

	return nil
}

// createGatewayInstance 创建网关实例并添加到连接池
func (app *GatewayApp) createGatewayInstance(cfg *gatewayconfig.GatewayConfig, source string) error {
	// 创建网关工厂
	factory := bootstrap.NewGatewayFactory()

	// 创建网关实例并添加到连接池
	_, err := factory.CreateGatewayWithPool(cfg, source)
	if err != nil {
		return huberrors.WrapError(err, "创建网关实例失败")
	}

	instanceID := cfg.InstanceID
	if instanceID == "" {
		instanceID = cfg.Base.Listen
	}

	logger.Info("网关实例创建成功",
		"instanceId", instanceID,
		"listen", cfg.Base.Listen,
		"source", source)

	return nil
}
