package loader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"gohub/internal/gateway/loader"
	"gohub/pkg/config"
	"gohub/pkg/database"
	_ "gohub/pkg/database/alldriver" // 导入数据库驱动
	"gohub/pkg/logger"
)

// TestDatabaseConfigLoaderExportYAML 测试从数据库加载网关配置并导出为YAML
func TestDatabaseConfigLoaderExportYAML(t *testing.T) {
	// 1. 初始化配置系统
	if err := initTestConfig(); err != nil {
		t.Fatalf("初始化配置失败: %v", err)
	}
	
	// 2. 初始化日志系统
	if err := initTestLogger(); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}
	
	// 3. 初始化数据库连接
	db, err := initTestDatabase()
	if err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}
	defer func() {
		if err := database.CloseAllConnections(); err != nil {
			t.Logf("关闭数据库连接时发生错误: %v", err)
		}
	}()
	
	// 4. 测试配置加载和导出
	if err := testLoadConfigAndExportYAML(t, db); err != nil {
		t.Fatalf("测试加载配置并导出YAML失败: %v", err)
	}
}

// initTestConfig 初始化测试配置
func initTestConfig() error {
	// 获取当前工作目录并构建配置路径
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}
	
	// 构建配置目录路径 (假设从test/gateway/loader回到项目根目录)
	configDir := filepath.Join(workDir, "..", "..", "..", "configs")
	
	// 加载配置文件
	options := config.LoadOptions{
		ClearExisting: false,
		AllowOverride: true,
	}
	
	if err := config.LoadConfig(configDir, options); err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}
	
	return nil
}

// initTestLogger 初始化测试日志系统
func initTestLogger() error {
	if err := logger.Setup(); err != nil {
		return fmt.Errorf("设置日志系统失败: %w", err)
	}
	return nil
}

// initTestDatabase 初始化测试数据库连接
func initTestDatabase() (database.Database, error) {
	// 构建数据库配置文件路径
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取工作目录失败: %w", err)
	}
	
	configPath := filepath.Join(workDir, "..", "..", "..", "configs", "database.yaml")
	
	// 获取默认连接名称
	defaultConn := config.GetString("database.default", "mysql")
	if defaultConn == "" {
		return nil, fmt.Errorf("未指定默认数据库连接")
	}
	
	// 加载所有数据库连接
	dbConnections, err := database.LoadAllConnections(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载数据库连接失败: %w", err)
	}
	
	// 获取默认连接
	db, ok := dbConnections[defaultConn]
	if !ok {
		return nil, fmt.Errorf("默认数据库连接 '%s' 未找到或未启用", defaultConn)
	}
	
	logger.Info("测试数据库连接成功", 
		"default", defaultConn,
		"driver", db.GetDriver())
	
	return db, nil
}

// testLoadConfigAndExportYAML 测试加载配置并导出YAML
func testLoadConfigAndExportYAML(t *testing.T, db database.Database) error {
	// 1. 创建数据库配置加载器
	tenantId := "default"  // 使用默认租户ID
	dbLoader := loader.NewDatabaseConfigLoader(db, tenantId)
	
	// 2. 定义要测试的网关实例ID列表
	instanceIds := []string{"GW20250620110323YWZI"} // 可以根据实际数据库中的数据调整
	
	// 3. 为每个实例ID加载配置并导出
	for _, instanceID := range instanceIds {
		t.Logf("正在测试实例ID: %s", instanceID)
		
		// 加载网关配置
		gatewayConfig, err := dbLoader.LoadGatewayConfig(instanceID)
		if err != nil {
			t.Logf("加载网关实例 '%s' 配置失败: %v", instanceID, err)
			continue // 继续测试下一个实例
		}
		
		t.Logf("成功加载网关实例 '%s' 配置", instanceID)
		t.Logf("  - 实例名称: %s", gatewayConfig.Base.Name)
		t.Logf("  - 监听地址: %s", gatewayConfig.Base.Listen)
		t.Logf("  - 路由数量: %d", len(gatewayConfig.Router.Routes))
		
		// 导出为YAML
		if err := exportConfigToYAML(t, gatewayConfig, instanceID); err != nil {
			t.Errorf("导出实例 '%s' 配置为YAML失败: %v", instanceID, err)
			continue
		}
		
		t.Logf("成功导出实例 '%s' 配置为YAML", instanceID)
	}
	
	return nil
}

// exportConfigToYAML 导出配置为YAML文件
func exportConfigToYAML(t *testing.T, gatewayConfig interface{}, instanceID string) error {
	// 1. 确保导出目录存在
	exportDir := "export"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("创建导出目录失败: %w", err)
	}
	
	// 2. 生成文件名
	filename := fmt.Sprintf("database_loaded_%s_config.yaml", instanceID)
	filePath := filepath.Join(exportDir, filename)
	
	// 3. 将配置转换为YAML
	yamlData, err := yaml.Marshal(gatewayConfig)
	if err != nil {
		return fmt.Errorf("序列化配置为YAML失败: %w", err)
	}
	
	// 4. 写入文件
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return fmt.Errorf("写入YAML文件失败: %w", err)
	}
	
	t.Logf("配置已导出到: %s", filePath)
	return nil
}

// TestDatabaseConnection 单独测试数据库连接
func TestDatabaseConnection(t *testing.T) {
	// 初始化配置
	if err := initTestConfig(); err != nil {
		t.Fatalf("初始化配置失败: %v", err)
	}
	
	// 初始化日志
	if err := initTestLogger(); err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}
	
	// 测试数据库连接
	db, err := initTestDatabase()
	if err != nil {
		t.Fatalf("数据库连接测试失败: %v", err)
	}
	defer func() {
		if err := database.CloseAllConnections(); err != nil {
			t.Logf("关闭数据库连接时发生错误: %v", err)
		}
	}()
	
	// 执行简单查询测试连接
	ctx := context.Background()
	
	// 测试查询网关实例表
	query := "SELECT COUNT(*) as count FROM HUB_GATEWAY_INSTANCE WHERE activeFlag = 'Y'"
	var result struct {
		Count int `db:"count"`
	}
	
	err = db.QueryOne(ctx, &result, query, nil, false)
	if err != nil {
		t.Logf("查询网关实例表失败: %v", err)
		t.Log("这可能是因为表不存在或数据库连接有问题")
	} else {
		t.Logf("数据库连接成功，找到 %d 个启用的网关实例", result.Count)
	}
} 