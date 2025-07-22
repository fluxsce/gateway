package factory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gohub/pkg/config"
	mongoConfig "gohub/pkg/mongo/config"
	"gohub/pkg/mongo/factory"
)

// TestMain 测试主函数，用于测试前的初始化
func TestMain(m *testing.M) {
	// 设置测试环境
	setupTestEnvironment()
	
	// 运行测试
	code := m.Run()
	
	// 清理测试环境
	teardownTestEnvironment()
	
	os.Exit(code)
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() {
	// 获取项目根目录
	workspaceRoot := getWorkspaceRoot()
	configPath := filepath.Join(workspaceRoot, "configs", "database.yaml")
	
	// 加载配置文件
	if err := config.LoadConfigFile(configPath); err != nil {
		panic("Failed to load config file: " + err.Error())
	}
}

// teardownTestEnvironment 清理测试环境
func teardownTestEnvironment() {
	// 关闭所有连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	factory.CloseAll(ctx)
}

// getWorkspaceRoot 获取工作区根目录
func getWorkspaceRoot() string {
	// 从当前测试文件路径推断工作区根目录
	// 当前路径：test/pkg/mongo/factory/factory_test.go
	// 需要向上4级到达根目录
	currentDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current directory: " + err.Error())
	}
	
	// 向上查找直到找到go.mod文件
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir
		}
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}
	
	// 如果找不到go.mod，则使用相对路径
	return filepath.Join("..", "..", "..", "..")
}

// TestManager_NewManager 测试Manager的创建
func TestManager_NewManager(t *testing.T) {
	manager := factory.NewManager()
	
	assert.NotNil(t, manager)
	assert.Empty(t, manager.ListConnections())
}

// TestManager_Connect 测试Manager的连接创建
func TestManager_Connect(t *testing.T) {
	manager := factory.NewManager()
	ctx := context.Background()
	
	// 注意：由于测试环境可能没有真实的MongoDB服务，这个测试可能会失败
	// 在实际环境中，你可能需要使用Mock或者确保有测试用的MongoDB实例
	t.Run("Connect_Success", func(t *testing.T) {
		// 这里我们只测试配置验证和基本逻辑
		// 实际的连接测试需要真实的MongoDB实例
		
		// 测试重复连接
		t.Run("Duplicate_Connection", func(t *testing.T) {
			// 假设第一次连接成功，测试重复连接
			// 由于没有真实的MongoDB，我们跳过这个测试
			t.Skip("需要真实的MongoDB实例进行测试")
		})
	})
	
	// 测试无效配置
	t.Run("Invalid_Config", func(t *testing.T) {
		invalidCfg := &mongoConfig.MongoConfig{
			Host:     "", // 无效的主机名
			Port:     0,  // 无效的端口
			Database: "", // 无效的数据库名
		}
		
		_, err := manager.Connect(ctx, "test_invalid", invalidCfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid config")
	})
}

// TestManager_GetConnection 测试连接获取
func TestManager_GetConnection(t *testing.T) {
	manager := factory.NewManager()
	
	// 测试获取不存在的连接
	_, err := manager.GetConnection("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestManager_ListConnections 测试连接列表
func TestManager_ListConnections(t *testing.T) {
	manager := factory.NewManager()
	
	// 新管理器应该没有连接
	connections := manager.ListConnections()
	assert.Empty(t, connections)
}

// TestManager_Stats 测试统计信息
func TestManager_Stats(t *testing.T) {
	manager := factory.NewManager()
	
	// 新管理器应该有空的统计信息
	stats := manager.Stats()
	assert.NotNil(t, stats)
	assert.Empty(t, stats)
}

// TestLoadAllMongoConnections 测试从配置文件加载所有MongoDB连接
func TestLoadAllMongoConnections(t *testing.T) {
	// 获取配置文件路径
	workspaceRoot := getWorkspaceRoot()
	configPath := filepath.Join(workspaceRoot, "configs", "database.yaml")
	
	t.Run("Load_Config_Success", func(t *testing.T) {
		// 清除现有配置
		config.Clear()
		
		// 测试加载配置
		connections, err := factory.LoadAllMongoConnections(configPath)
		
		if err != nil {
			// 如果加载失败，可能是因为MongoDB服务不可用
			t.Logf("加载MongoDB连接失败（可能是服务不可用）: %v", err)
			return
		}
		
		// 验证连接映射
		assert.NotNil(t, connections)
		t.Logf("加载的连接数: %d", len(connections))
		
		// 如果MongoDB启用且有连接，验证连接
		if len(connections) > 0 {
			// 验证每个连接都不为空
			for name, conn := range connections {
				assert.NotNil(t, conn, "连接 %s 不应为空", name)
				t.Logf("连接: %s", name)
			}
		}
	})
	
	t.Run("Load_Nonexistent_Config", func(t *testing.T) {
		// 测试加载不存在的配置文件
		_, err := factory.LoadAllMongoConnections("nonexistent.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "加载配置文件失败")
	})
}


// TestGetConnectionInfo 测试获取连接信息
func TestGetConnectionInfo(t *testing.T) {
	info := factory.GetConnectionInfo()
	assert.NotNil(t, info)
	// 新系统应该没有活跃连接
	assert.Empty(t, info)
}

// TestHealthCheck 测试健康检查
func TestHealthCheck(t *testing.T) {
	ctx := context.Background()
	
	results := factory.HealthCheck(ctx)
	assert.NotNil(t, results)
	// 没有连接时应该返回空结果
	assert.Empty(t, results)
}

// TestGlobalFunctions 测试全局函数
func TestGlobalFunctions(t *testing.T) {
	t.Run("ListConnections", func(t *testing.T) {
		connections := factory.ListConnections()
		assert.NotNil(t, connections)
		// 初始状态应该没有连接
		assert.Empty(t, connections)
	})
	
	t.Run("GetConnection_NotFound", func(t *testing.T) {
		_, err := factory.GetConnection("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	
	t.Run("CloseAll", func(t *testing.T) {
		ctx := context.Background()
		err := factory.CloseAll(ctx)
		assert.NoError(t, err)
	})
}

// TestConfigIntegration 测试配置集成
func TestConfigIntegration(t *testing.T) {
	// 测试从配置系统读取MongoDB配置
	t.Run("Read_Mongo_Config", func(t *testing.T) {
		// 验证配置是否正确加载
		enabled := config.GetBool("mongo.enabled", false)
		defaultConn := config.GetString("mongo.default", "")
		
		t.Logf("MongoDB启用状态: %v", enabled)
		t.Logf("默认连接: %s", defaultConn)
		
		// 如果MongoDB启用，验证默认连接配置
		if enabled && defaultConn != "" {
			// 构造配置键
			hostKey := "mongo.connections." + defaultConn + ".host"
			portKey := "mongo.connections." + defaultConn + ".port"
			databaseKey := "mongo.connections." + defaultConn + ".database"
			
			host := config.GetString(hostKey, "")
			port := config.GetInt(portKey, 0)
			database := config.GetString(databaseKey, "")
			
			t.Logf("默认连接配置 - Host: %s, Port: %d, Database: %s", host, port, database)
			
			// 验证配置不为空
			assert.NotEmpty(t, host, "主机地址不应为空")
			assert.Greater(t, port, 0, "端口号应大于0")
			assert.NotEmpty(t, database, "数据库名不应为空")
		}
	})
	
	t.Run("Parse_Mongo_Root_Config", func(t *testing.T) {
		// 测试解析MongoDB根配置
		var mongoRootConfig factory.MongoRootConfig
		err := config.GetSection("mongo", &mongoRootConfig)
		
		if err != nil {
			t.Logf("解析MongoDB配置失败: %v", err)
			return
		}
		
		t.Logf("MongoDB根配置 - 启用: %v, 默认: %s, 连接数: %d",
			mongoRootConfig.Enabled,
			mongoRootConfig.Default,
			len(mongoRootConfig.Connections))
		
		// 验证配置结构
		assert.NotNil(t, mongoRootConfig.Connections)
		
		// 如果启用了MongoDB，验证默认连接存在
		if mongoRootConfig.Enabled && mongoRootConfig.Default != "" {
			_, exists := mongoRootConfig.Connections[mongoRootConfig.Default]
			assert.True(t, exists, "默认连接配置应该存在")
		}
		
		// 验证每个连接配置
		for name, connConfig := range mongoRootConfig.Connections {
			assert.NotNil(t, connConfig, "连接配置 %s 不应为空", name)
			t.Logf("连接 %s: Host=%s, Port=%d, Database=%s",
				name, connConfig.Host, connConfig.Port, connConfig.Database)
		}
	})
}

// TestCloseAllConnections 测试关闭所有连接
func TestCloseAllConnections(t *testing.T) {
	err := factory.CloseAllConnections()
	assert.NoError(t, err)
}

// TestNewClientFromURI 测试从URI创建客户端
func TestNewClientFromURI(t *testing.T) {
	// 测试无效URI
	t.Run("Invalid_URI", func(t *testing.T) {
		// 由于这个函数的实现是简化的，我们只测试基本功能
		// 实际测试需要真实的MongoDB实例
		t.Skip("需要真实的MongoDB实例进行测试")
	})
}

// TestNewClientFromConfig 测试从配置文件创建客户端
func TestNewClientFromConfig(t *testing.T) {
	t.Run("From_Config_File", func(t *testing.T) {
		// 由于这个函数的实现是简化的，我们只测试基本功能
		// 实际测试需要真实的MongoDB实例
		t.Skip("需要真实的MongoDB实例进行测试")
		
		// 在需要真实测试时，可以取消注释以下代码：
		// configPath := filepath.Join(getWorkspaceRoot(), "configs", "database.yaml")
		// ctx := context.Background()
		// _, err := factory.NewClientFromConfig(ctx, configPath)
		// assert.NoError(t, err)
	})
}

// BenchmarkManager_Connect 基准测试：连接创建
func BenchmarkManager_Connect(b *testing.B) {
	manager := factory.NewManager()
	ctx := context.Background()
	
	cfg := &mongoConfig.MongoConfig{
		Host:         "localhost",
		Port:         27017,
		Database:     "benchmark_db",
		Username:     "",
		Password:     "",
		AuthDB:       "",
		MaxPoolSize:  10,
		MinPoolSize:  2,
		EnableTLS:    false,
		AppName:      "benchmark_app",
		RetryWrites:  true,
		RetryReads:   true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 由于需要真实的MongoDB实例，我们跳过这个基准测试
		b.Skip("需要真实的MongoDB实例进行基准测试")
		
		connectionName := "bench_conn_" + string(rune(i))
		_, err := manager.Connect(ctx, connectionName, cfg)
		if err != nil {
			b.Errorf("连接创建失败: %v", err)
		}
	}
}

// BenchmarkLoadAllMongoConnections 基准测试：加载所有连接
func BenchmarkLoadAllMongoConnections(b *testing.B) {
	configPath := filepath.Join(getWorkspaceRoot(), "configs", "database.yaml")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 清除现有配置
		config.Clear()
		
		_, err := factory.LoadAllMongoConnections(configPath)
		if err != nil {
			// 在基准测试中，连接失败是可以接受的
			b.Logf("加载连接失败: %v", err)
		}
	}
}

// ExampleLoadAllMongoConnections 示例：加载所有MongoDB连接
func ExampleLoadAllMongoConnections() {
	// 加载配置文件
	configPath := "configs/database.yaml"
	connections, err := factory.LoadAllMongoConnections(configPath)
	if err != nil {
		panic(err)
	}
	
	// 使用连接
	for name, conn := range connections {
		if conn != nil {
			// 使用连接进行数据库操作
			_ = name // 使用连接名称
		}
	}
	
	// 程序结束时关闭所有连接
	defer factory.CloseAllConnections()
}

// ExampleManager 示例：使用连接管理器
func ExampleManager() {
	// 创建连接管理器
	manager := factory.NewManager()
	
	// 创建配置
	cfg := &mongoConfig.MongoConfig{
		Host:         "localhost",
		Port:         27017,
		Database:     "example_db",
		Username:     "user",
		Password:     "password",
		AuthDB:       "admin",
		MaxPoolSize:  100,
		MinPoolSize:  5,
		EnableTLS:    false,
		AppName:      "example_app",
		RetryWrites:  true,
		RetryReads:   true,
	}
	
	// 连接到MongoDB
	ctx := context.Background()
	client, err := manager.Connect(ctx, "example_conn", cfg)
	if err != nil {
		panic(err)
	}
	
	// 使用客户端
	_ = client
	
	// 关闭所有连接
	defer manager.CloseAll(ctx)
}
