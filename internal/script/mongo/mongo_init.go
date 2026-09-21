package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/logger"
	mongoclient "gateway/pkg/mongo/client"
	mongofactory "gateway/pkg/mongo/factory"
)

// MongoScriptExecutionResult MongoDB脚本执行结果
type MongoScriptExecutionResult struct {
	// ConnectionName MongoDB连接名称
	ConnectionName string

	// ScriptFile 脚本文件路径
	ScriptFile string

	// Success 执行是否成功
	Success bool

	// Error 执行错误信息（如果有）
	Error error

	// Duration 执行耗时
	Duration time.Duration

	// CommandsExecuted 成功执行的命令数量
	CommandsExecuted int

	// CommandsFailed 失败的命令数量
	CommandsFailed int

	// Output 脚本执行输出
	Output string
}

// GetMongoConnection 获取 MongoDB 数据源连接
// 返回:
//   - *mongofactory.Client: MongoDB 数据库连接实例，如果未配置则返回 nil
func GetMongoConnection() (*mongofactory.Manager, error) {
	// 尝试获取默认 MongoDB 连接
	client, err := mongofactory.GetDefaultConnection()
	if err != nil || client == nil {
		logger.Debug("未配置 MongoDB 连接或连接未启用")
		return nil, nil
	}

	logger.Debug("成功获取 MongoDB 连接")
	return mongofactory.NewManager(), nil
}

// IsMongoEnabled 检查 MongoDB 是否已启用
// 返回:
//   - bool: true 表示 MongoDB 已配置且启用
func IsMongoEnabled() bool {
	_, err := mongofactory.GetDefaultConnection()
	return err == nil
}

// ExecuteMongoScript 执行 MongoDB 初始化脚本
// 使用内置的静态索引命令列表，不再依赖外部 JavaScript 文件
// 参数:
//   - ctx: 上下文对象
//   - scriptDir: 脚本目录路径（保留参数以兼容接口，但不再使用）
//
// 返回:
//   - *MongoScriptExecutionResult: 脚本执行结果
//   - error: 执行失败时返回错误信息
func ExecuteMongoScript(ctx context.Context, scriptDir string) (*MongoScriptExecutionResult, error) {
	startTime := time.Now()

	result := &MongoScriptExecutionResult{
		ConnectionName: "default",
		Success:        false,
		ScriptFile:     "内置静态索引命令",
	}

	// 检查 MongoDB 是否启用
	if !IsMongoEnabled() {
		logger.Info("MongoDB 未启用，跳过脚本执行")
		result.Success = true // 不作为错误处理
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// 获取 MongoDB 客户端
	client, err := mongofactory.GetDefaultConnection()
	if err != nil {
		result.Error = fmt.Errorf("获取 MongoDB 连接失败: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	logger.Info("开始执行 MongoDB 初始化脚本（使用内置命令）",
		"connection", result.ConnectionName)

	// 执行索引创建命令
	executed, failed, output, err := executeMongoIndexCommands(ctx, client)
	result.CommandsExecuted = executed
	result.CommandsFailed = failed
	result.Output = output

	if err != nil {
		result.Error = fmt.Errorf("执行 MongoDB 脚本失败: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	result.Success = true
	result.Duration = time.Since(startTime)

	logger.Info("MongoDB 脚本执行完成",
		"executed", executed,
		"failed", failed,
		"duration", result.Duration)

	return result, nil
}

// executeMongoIndexCommands 执行 MongoDB 索引创建命令
// 参数:
//   - ctx: 上下文对象
//   - client: MongoDB 客户端
//
// 返回:
//   - int: 成功执行的命令数
//   - int: 失败的命令数
//   - string: 执行输出
//   - error: 执行错误
func executeMongoIndexCommands(ctx context.Context, client *mongoclient.Client) (int, int, string, error) {
	// 获取所有索引命令
	commands := GetMongoInitCommands()
	commandsByCollection := GetIndexCommandsByCollection()

	executed := 0
	failed := 0
	var outputBuilder strings.Builder

	logger.Info("MongoDB 索引命令加载完成",
		"total_commands", len(commands),
		"collections", len(commandsByCollection))

	outputBuilder.WriteString("开始创建 MongoDB 索引...\n\n")

	// 按集合分组执行，且只处理白名单中的日志集合
	for collName, collCommands := range commandsByCollection {
		if !IsManagedMongoLogCollection(collName) {
			logger.Warn("跳过非系统日志集合的索引对齐", "collection", collName)
			outputBuilder.WriteString(fmt.Sprintf("[%s] 跳过：不在系统日志集合白名单\n\n", collName))
			continue
		}
		outputBuilder.WriteString(fmt.Sprintf("[%s] 创建索引...\n", collName))
		logger.Info(fmt.Sprintf("开始为集合 %s 创建索引", collName),
			"collection", collName,
			"index_count", len(collCommands))

		// 获取数据库
		db, err := client.DefaultDatabase()
		if err != nil {
			errMsg := fmt.Sprintf("无法获取默认数据库: %v", err)
			logger.Error(errMsg, "error", err)
			outputBuilder.WriteString(fmt.Sprintf("  错误: %s\n", errMsg))
			failed += len(collCommands)
			continue
		}

		collection := db.Collection(collName)
		coll, ok := collection.(*mongoclient.Collection)
		if !ok {
			logger.Warn("集合实现不是 *client.Collection，跳过索引对齐",
				"collection", collName)
			failed += len(collCommands)
			outputBuilder.WriteString("  错误: 集合类型不支持有序建索引与自动删旧索引\n\n")
			continue
		}

		created, createFailed, detail := ensureCollectionIndexes(ctx, coll, collName, collCommands)
		executed += created
		failed += createFailed
		outputBuilder.WriteString(detail)
		outputBuilder.WriteString(fmt.Sprintf("[%s] 索引对齐完成\n\n", collName))
	}

	// 输出汇总信息
	outputBuilder.WriteString("MongoDB 索引对齐完成（自动建最小集并删除旧索引）\n")
	outputBuilder.WriteString(fmt.Sprintf("- 成功创建/对齐: %d\n", executed))
	outputBuilder.WriteString(fmt.Sprintf("- 失败: %d\n", failed))
	outputBuilder.WriteString("- TTL: 30天；旧的多余索引由启动流程删除，无需手工 drop\n")

	logger.Info("MongoDB 索引创建完成",
		"total_executed", executed,
		"total_failed", failed,
		"total_commands", len(commands))

	return executed, failed, outputBuilder.String(), nil
}

// InitializeMongoScript 初始化 MongoDB 脚本（集成到数据库初始化流程）
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - *MongoScriptExecutionResult: 执行结果
//   - error: 执行失败时返回错误信息
func InitializeMongoScript(ctx context.Context) (*MongoScriptExecutionResult, error) {
	logger.Info("开始 MongoDB 索引初始化（使用内置命令）")
	return ExecuteMongoScript(ctx, "")
}

// GetMongoCommandsInfo 获取 MongoDB 命令信息
// 返回:
//   - map[string]interface{}: 命令信息
func GetMongoCommandsInfo() map[string]interface{} {
	commands := GetMongoInitCommands()
	commandsByCollection := GetIndexCommandsByCollection()

	// 统计索引类型
	uniqueCount := 0
	ttlCount := 0
	sparseCount := 0

	for _, cmd := range commands {
		if cmd.IndexModel.Options != nil {
			if cmd.IndexModel.Options.Unique {
				uniqueCount++
			}
			if cmd.IndexModel.Options.ExpireAfterSeconds != nil {
				ttlCount++
			}
			if cmd.IndexModel.Options.Sparse {
				sparseCount++
			}
		}
	}

	info := map[string]interface{}{
		"total_commands": len(commands),
		"collections":    len(commandsByCollection),
		"unique_indexes": uniqueCount,
		"ttl_indexes":    ttlCount,
		"sparse_indexes": sparseCount,
		"source":         "内置静态命令列表",
	}

	return info
}
