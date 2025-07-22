package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gohub/pkg/config"
	"gohub/pkg/mongo/client"
	"gohub/pkg/mongo/factory"
	"gohub/pkg/mongo/types"
)

var (
	testClient   *client.Client
	testDatabase types.MongoDatabase
	testManager  *factory.Manager
)

// TestMain 测试主函数，用于测试前的初始化和清理
func TestMain(m *testing.M) {
	// 设置测试环境
	if err := setupTestEnvironment(); err != nil {
		panic("Failed to setup test environment: " + err.Error())
	}
	
	// 运行测试
	code := m.Run()
	
	// 清理测试环境
	teardownTestEnvironment()
	
	os.Exit(code)
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() error {
	// 获取项目根目录
	workspaceRoot := getWorkspaceRoot()
	configPath := filepath.Join(workspaceRoot, "configs", "database.yaml")
	
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，跳过需要真实连接的测试
		return nil
	}
	
	// 加载配置文件
	if err := config.LoadConfigFile(configPath); err != nil {
		return err
	}
	
	// 检查MongoDB是否启用
	if !config.GetBool("mongo.enabled", false) {
		// MongoDB未启用，跳过连接测试
		return nil
	}
	
	// 创建factory manager
	testManager = factory.NewManager()
	
	// 从配置加载连接
	connections, err := factory.LoadAllMongoConnections(configPath)
	if err != nil {
		// 连接失败是可以接受的，因为可能没有真实的MongoDB服务
		return nil
	}
	
	// 获取默认连接
	defaultConnName := config.GetString("mongo.default", "")
	if defaultConnName == "" {
		// 没有默认连接配置
		return nil
	}
	
	if conn, exists := connections[defaultConnName]; exists {
		testClient = conn
		testDatabase = testClient.Database("gohub_test")
	}
	
	return nil
}

// teardownTestEnvironment 清理测试环境
func teardownTestEnvironment() {
	if testManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// 清理测试数据库
		if testDatabase != nil {
			//testDatabase.Drop(ctx)
		}
		
		// 关闭所有连接
		testManager.CloseAll(ctx)
	}
}

// getWorkspaceRoot 获取工作区根目录
func getWorkspaceRoot() string {
	// 从当前测试文件路径推断工作区根目录
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

// skipIfNoConnection 如果没有测试连接则跳过测试
func skipIfNoConnection(t *testing.T) {
	if testClient == nil {
		t.Skip("跳过测试：没有可用的MongoDB连接")
	}
}

// TestClientBasicOperations 测试客户端基本操作
func TestClientBasicOperations(t *testing.T) {
	skipIfNoConnection(t)
	
	ctx := context.Background()
	
	t.Run("Ping", func(t *testing.T) {
		err := testClient.Ping(ctx)
		assert.NoError(t, err, "ping操作应该成功")
	})
	
	t.Run("Database", func(t *testing.T) {
		db := testClient.Database("gohub_test")
		assert.NotNil(t, db, "数据库实例不应为空")
		assert.Equal(t, "gohub_test", db.Name(), "数据库名称应该匹配")
	})
	
	t.Run("ListDatabaseNames", func(t *testing.T) {
		names, err := testClient.ListDatabaseNames(ctx, nil)
		assert.NoError(t, err, "列出数据库名称应该成功")
		assert.NotNil(t, names, "数据库名称列表不应为空")
		t.Logf("发现的数据库: %v", names)
	})
}

// TestDatabaseOperations 测试数据库操作
func TestDatabaseOperations(t *testing.T) {
	skipIfNoConnection(t)
	
	ctx := context.Background()
	
	t.Run("CreateCollection", func(t *testing.T) {
		collectionName := "test_collection_" + time.Now().Format("20060102150405")
		
		err := testDatabase.CreateCollection(ctx, collectionName)
		assert.NoError(t, err, "创建集合应该成功")
		
		// 验证集合是否存在
		names, err := testDatabase.ListCollectionNames(ctx, nil)
		assert.NoError(t, err, "列出集合名称应该成功")
		assert.Contains(t, names, collectionName, "集合应该存在于列表中")
		
		// 清理：删除测试集合
		// err = testDatabase.DropCollection(ctx, collectionName)
		// assert.NoError(t, err, "删除集合应该成功")
	})
	
	t.Run("Collection", func(t *testing.T) {
		collection := testDatabase.Collection("test_collection")
		assert.NotNil(t, collection, "集合实例不应为空")
		assert.Equal(t, "test_collection", collection.Name(), "集合名称应该匹配")
	})
	
	t.Run("ListCollectionNames", func(t *testing.T) {
		names, err := testDatabase.ListCollectionNames(ctx, nil)
		assert.NoError(t, err, "列出集合名称应该成功")
		assert.NotNil(t, names, "集合名称列表不应为空")
		t.Logf("发现的集合: %v", names)
	})
	
	t.Run("RunCommand", func(t *testing.T) {
		// 测试 ping 命令
		pingCmd := types.Document{"ping": 1}
		result := testDatabase.RunCommand(ctx, pingCmd, nil)
		assert.NotNil(t, result, "命令结果不应为空")
		
		var pingResult types.Document
		err := result.Decode(&pingResult)
		assert.NoError(t, err, "解码命令结果应该成功")
		assert.Equal(t, float64(1), pingResult["ok"], "ping命令应该返回ok=1")
		
		t.Logf("Ping命令结果: %v", pingResult)
	})
	
	t.Run("RunCommandServerStatus", func(t *testing.T) {
		// 测试 serverStatus 命令
		statusCmd := types.Document{"serverStatus": 1}
		result := testDatabase.RunCommand(ctx, statusCmd, nil)
		assert.NotNil(t, result, "命令结果不应为空")
		
		var statusResult types.Document
		err := result.Decode(&statusResult)
		assert.NoError(t, err, "解码命令结果应该成功")
		assert.Equal(t, float64(1), statusResult["ok"], "serverStatus命令应该返回ok=1")
		
		// 验证返回的字段
		assert.Contains(t, statusResult, "version", "应该包含version字段")
		assert.Contains(t, statusResult, "uptime", "应该包含uptime字段")
		
		t.Logf("ServerStatus命令结果包含字段: %v", getKeys(statusResult))
	})
	
	t.Run("RunCommandBuildInfo", func(t *testing.T) {
		// 测试 buildInfo 命令
		buildInfoCmd := types.Document{"buildInfo": 1}
		result := testDatabase.RunCommand(ctx, buildInfoCmd, nil)
		assert.NotNil(t, result, "命令结果不应为空")
		
		var buildInfoResult types.Document
		err := result.Decode(&buildInfoResult)
		assert.NoError(t, err, "解码命令结果应该成功")
		assert.Equal(t, float64(1), buildInfoResult["ok"], "buildInfo命令应该返回ok=1")
		
		// 验证返回的字段
		assert.Contains(t, buildInfoResult, "version", "应该包含version字段")
		assert.Contains(t, buildInfoResult, "gitVersion", "应该包含gitVersion字段")
		
		t.Logf("MongoDB版本信息: %v", buildInfoResult["version"])
	})
}

// getKeys 辅助函数，获取 Document 的所有键
func getKeys(doc types.Document) []string {
	keys := make([]string, 0, len(doc))
	for key := range doc {
		keys = append(keys, key)
	}
	return keys
}

// TestCollectionBasicOperations 测试集合基本操作
func TestCollectionBasicOperations(t *testing.T) {
	skipIfNoConnection(t)
	
	ctx := context.Background()
	collectionName := "test_crud_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)
	
	// 确保测试后清理
	defer func() {
		//testDatabase.DropCollection(ctx, collectionName)
	}()
	
	t.Run("InsertOne", func(t *testing.T) {
		doc := types.Document{
			"name":  "测试文档",
			"value": 123,
			"tags":  []string{"test", "mongo"},
		}
		
		result, err := collection.InsertOne(ctx, doc, nil)
		assert.NoError(t, err, "插入文档应该成功")
		assert.NotNil(t, result, "插入结果不应为空")
		assert.NotNil(t, result.InsertedID, "插入的ID不应为空")
		t.Logf("插入的文档ID: %v", result.InsertedID)
	})
	
	t.Run("FindOne", func(t *testing.T) {
		// 先插入一个文档
		doc := types.Document{
			"name":   "查找测试",
			"number": 456,
		}
		insertResult, err := collection.InsertOne(ctx, doc, nil)
		require.NoError(t, err, "插入文档应该成功")
		
		// 查找文档
		filter := types.Document{"name": "查找测试"}
		result := collection.FindOne(ctx, types.Filter(filter), nil)
		assert.NotNil(t, result, "查找结果不应为空")
		
		var foundDoc types.Document
		err = result.Decode(&foundDoc)
		assert.NoError(t, err, "解码文档应该成功")
		assert.Equal(t, "查找测试", foundDoc["name"], "文档名称应该匹配")
		t.Logf("找到的文档: %v", foundDoc)
		t.Logf("插入的ID: %v", insertResult.InsertedID)
	})
	
	t.Run("InsertMany", func(t *testing.T) {
		docs := []types.Document{
			{"batch": 1, "name": "批量文档1"},
			{"batch": 1, "name": "批量文档2"},
			{"batch": 1, "name": "批量文档3"},
		}
		
		result, err := collection.InsertMany(ctx, docs, nil)
		assert.NoError(t, err, "批量插入应该成功")
		assert.NotNil(t, result, "插入结果不应为空")
		assert.Len(t, result.InsertedIDs, 3, "应该插入3个文档")
		t.Logf("批量插入的ID: %v", result.InsertedIDs)
	})
	
	t.Run("Find", func(t *testing.T) {
		// 查找所有batch=1的文档
		filter := types.Document{"batch": 1}
		cursor, err := collection.Find(ctx, types.Filter(filter), nil)
		assert.NoError(t, err, "查找多个文档应该成功")
		assert.NotNil(t, cursor, "游标不应为空")
		
		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码所有文档应该成功")
		assert.GreaterOrEqual(t, len(results), 3, "应该找到至少3个文档")
		t.Logf("找到 %d 个文档", len(results))
	})
	
	t.Run("Count", func(t *testing.T) {
		filter := types.Document{"batch": 1}
		count, err := collection.Count(ctx, types.Filter(filter), nil)
		assert.NoError(t, err, "计数文档应该成功")
		assert.GreaterOrEqual(t, count, int64(3), "应该有至少3个文档")
		t.Logf("文档数量: %d", count)
	})
	
	t.Run("UpdateOne", func(t *testing.T) {
		filter := types.Document{"name": "批量文档1"}
		update := types.Document{"$set": types.Document{"updated": true}}
		
		result, err := collection.UpdateOne(ctx, types.Filter(filter), types.Update(update), nil)
		assert.NoError(t, err, "更新文档应该成功")
		assert.NotNil(t, result, "更新结果不应为空")
		assert.Equal(t, int64(1), result.MatchedCount, "应该匹配1个文档")
		assert.Equal(t, int64(1), result.ModifiedCount, "应该修改1个文档")
		t.Logf("更新结果 - 匹配: %d, 修改: %d", result.MatchedCount, result.ModifiedCount)
	})
	
	t.Run("DeleteOne", func(t *testing.T) {
		filter := types.Document{"name": "批量文档2"}
		
		result, err := collection.DeleteOne(ctx, types.Filter(filter), nil)
		assert.NoError(t, err, "删除文档应该成功")
		assert.NotNil(t, result, "删除结果不应为空")
		assert.Equal(t, int64(1), result.DeletedCount, "应该删除1个文档")
		t.Logf("删除的文档数量: %d", result.DeletedCount)
	})
}

// TestFactoryIntegration 测试工厂集成
func TestFactoryIntegration(t *testing.T) {
	// 测试工厂manager是否能正确获取client
	if testManager == nil {
		t.Skip("跳过测试：factory manager未初始化")
	}
	
	t.Run("ManagerListConnections", func(t *testing.T) {
		connections := testManager.ListConnections()
		assert.NotNil(t, connections, "连接列表不应为空")
		t.Logf("活跃连接: %v", connections)
	})
	
	t.Run("ManagerGetConnection", func(t *testing.T) {
		connections := testManager.ListConnections()
		if len(connections) == 0 {
			t.Skip("没有活跃连接")
		}
		
		// 获取第一个连接
		client, err := testManager.GetConnection(connections[0])
		assert.NoError(t, err, "获取连接应该成功")
		assert.NotNil(t, client, "客户端不应为空")
		
		// 测试连接
		ctx := context.Background()
		err = client.Ping(ctx)
		assert.NoError(t, err, "ping应该成功")
	})
	
	t.Run("ManagerStats", func(t *testing.T) {
		stats := testManager.Stats()
		assert.NotNil(t, stats, "统计信息不应为空")
		t.Logf("连接统计: %v", stats)
	})
	
	t.Run("ManagerHealthCheck", func(t *testing.T) {
		ctx := context.Background()
		results := factory.HealthCheck(ctx)
		assert.NotNil(t, results, "健康检查结果不应为空")
		
		for name, err := range results {
			if err != nil {
				t.Logf("连接 %s 健康检查失败: %v", name, err)
			} else {
				t.Logf("连接 %s 健康检查通过", name)
			}
		}
	})
}

// TestConnectionConfiguration 测试连接配置
func TestConnectionConfiguration(t *testing.T) {
	t.Run("ReadMongoConfig", func(t *testing.T) {
		// 测试读取MongoDB配置
		enabled := config.GetBool("mongo.enabled", false)
		defaultConn := config.GetString("mongo.default", "")
		
		t.Logf("MongoDB启用状态: %v", enabled)
		t.Logf("默认连接: %s", defaultConn)
		
		if enabled && defaultConn != "" {
			// 构造配置键并读取配置
			hostKey := "mongo.connections." + defaultConn + ".host"
			portKey := "mongo.connections." + defaultConn + ".port"
			databaseKey := "mongo.connections." + defaultConn + ".database"
			
			host := config.GetString(hostKey, "")
			port := config.GetInt(portKey, 0)
			database := config.GetString(databaseKey, "")
			
			t.Logf("连接配置 - Host: %s, Port: %d, Database: %s", host, port, database)
			
			// 基本配置验证
			assert.NotEmpty(t, host, "主机地址不应为空")
			assert.Greater(t, port, 0, "端口号应大于0")
			assert.NotEmpty(t, database, "数据库名不应为空")
		}
	})
	
	t.Run("ParseMongoRootConfig", func(t *testing.T) {
		// 测试解析完整的MongoDB根配置
		var mongoRootConfig factory.MongoRootConfig
		err := config.GetSection("mongo", &mongoRootConfig)
		
		if err != nil {
			t.Logf("解析MongoDB配置失败: %v", err)
			return
		}
		
		t.Logf("MongoDB根配置解析成功")
		t.Logf("启用状态: %v", mongoRootConfig.Enabled)
		t.Logf("默认连接: %s", mongoRootConfig.Default)
		t.Logf("连接数量: %d", len(mongoRootConfig.Connections))
		
		// 验证配置结构
		assert.NotNil(t, mongoRootConfig.Connections, "连接配置不应为空")
		
		// 验证每个连接配置
		for name, connConfig := range mongoRootConfig.Connections {
			assert.NotNil(t, connConfig, "连接配置 %s 不应为空", name)
			t.Logf("连接 %s: Host=%s, Port=%d, Database=%s",
				name, connConfig.Host, connConfig.Port, connConfig.Database)
		}
	})
}

// TestClientError 测试客户端错误处理
func TestClientError(t *testing.T) {
	t.Run("DisconnectedClientPing", func(t *testing.T) {
		// 创建一个新的客户端但不连接
		client := client.NewClient()
		
		ctx := context.Background()
		err := client.Ping(ctx)
		assert.Error(t, err, "未连接的客户端ping应该失败")
		assert.Contains(t, err.Error(), "not connected", "错误信息应该包含'not connected'")
	})
	
	t.Run("DisconnectedClientDatabase", func(t *testing.T) {
		// 创建一个新的客户端但不连接
		client := client.NewClient()
		
		// 获取数据库实例应该成功（延迟连接）
		db := client.Database("test")
		assert.NotNil(t, db, "获取数据库实例应该成功")
		assert.Equal(t, "test", db.Name(), "数据库名称应该匹配")
	})
}

// BenchmarkClientOperations 基准测试客户端操作
func BenchmarkClientOperations(b *testing.B) {
	if testClient == nil {
		b.Skip("跳过基准测试：没有可用的MongoDB连接")
	}
	
	ctx := context.Background()
	collection := testDatabase.Collection("benchmark_test")
	
	// 清理测试数据
	defer testDatabase.DropCollection(ctx, "benchmark_test")
	
	b.Run("Ping", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := testClient.Ping(ctx)
			if err != nil {
				b.Errorf("ping失败: %v", err)
			}
		}
	})
	
	b.Run("InsertOne", func(b *testing.B) {
		doc := types.Document{
			"name":  "benchmark",
			"value": 123,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := collection.InsertOne(ctx, doc, nil)
			if err != nil {
				b.Errorf("插入失败: %v", err)
			}
		}
	})
	
	b.Run("FindOne", func(b *testing.B) {
		// 先插入一个文档
		doc := types.Document{
			"name":  "benchmark_find",
			"value": 456,
		}
		collection.InsertOne(ctx, doc, nil)
		
		filter := types.Document{"name": "benchmark_find"}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := collection.FindOne(ctx, types.Filter(filter), nil)
			var foundDoc types.Document
			err := result.Decode(&foundDoc)
			if err != nil {
				b.Errorf("查找失败: %v", err)
			}
		}
	})
}

// ExampleClient 示例：使用MongoDB客户端
func ExampleClient() {
	// 通过factory获取配置的客户端
	configPath := "configs/database.yaml"
	connections, err := factory.LoadAllMongoConnections(configPath)
	if err != nil {
		panic(err)
	}
	
	// 获取默认连接
	client := connections["mongo_main"]
	if client == nil {
		panic("无法获取MongoDB客户端")
	}
	
	// 使用客户端
	ctx := context.Background()
	
	// 测试连接
	err = client.Ping(ctx)
	if err != nil {
		panic(err)
	}
	
	// 获取数据库
	db := client.Database("example")
	
	// 获取集合
	collection := db.Collection("users")
	
	// 插入文档
	doc := types.Document{
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   30,
	}
	
	result, err := collection.InsertOne(ctx, doc, nil)
	if err != nil {
		panic(err)
	}
	
	_ = result // 使用插入结果
	
	// 程序结束时关闭连接
	defer factory.CloseAllConnections()
}
