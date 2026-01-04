package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/logger"
	mongoclient "gateway/pkg/mongo/client"
	mongofactory "gateway/pkg/mongo/factory"
	mongotypes "gateway/pkg/mongo/types"
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

	// 按集合分组执行
	for collName, collCommands := range commandsByCollection {
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

		// 逐个创建索引
		for i, cmd := range collCommands {
			logger.Debug("准备创建索引",
				"collection", collName,
				"index", cmd.IndexModel.Options.Name,
				"description", cmd.Description)

			// 使用 MongoDB types.IndexModel 创建索引
			// 将 bson.D 转换为 types.Document (bson.M)
			keys := mongotypes.Document{}
			for _, elem := range cmd.IndexModel.Keys {
				keys[elem.Key] = elem.Value
			}

			indexModel := mongotypes.IndexModel{
				Keys: keys,
			}

			// 获取集合
			collection := db.Collection(collName)

			// 创建索引
			indexName, err := collection.CreateIndex(ctx, indexModel)
			if err != nil {
				// 检查是否是索引已存在错误
				if strings.Contains(err.Error(), "already exists") ||
					strings.Contains(err.Error(), "IndexOptionsConflict") {
					logger.Info("索引已存在，跳过创建",
						"collection", collName,
						"index", cmd.IndexModel.Options.Name)
					executed++
					outputBuilder.WriteString(fmt.Sprintf("  %d. %s (已存在)\n", i+1, cmd.IndexModel.Options.Name))
					continue
				}

				logger.Warn("创建索引失败，继续执行后续索引",
					"collection", collName,
					"index", cmd.IndexModel.Options.Name,
					"error", err)
				failed++
				outputBuilder.WriteString(fmt.Sprintf("  %d. %s 失败: %v\n", i+1, cmd.IndexModel.Options.Name, err))
				continue
			}

			executed++
			outputBuilder.WriteString(fmt.Sprintf("  %d. %s (%s)\n", i+1, indexName, cmd.Description))

			logger.Info("索引创建成功",
				"collection", collName,
				"index", indexName,
				"description", cmd.Description)

			// 每执行 5 个索引记录一次进度
			if executed%5 == 0 {
				logger.Info("MongoDB 索引创建进度",
					"executed", executed,
					"failed", failed,
					"total", len(commands))
			}
		}

		outputBuilder.WriteString(fmt.Sprintf("[%s] 索引创建完成\n\n", collName))
	}

	// 输出汇总信息
	outputBuilder.WriteString("MongoDB 索引创建完成\n")
	outputBuilder.WriteString(fmt.Sprintf("- 成功创建: %d 个索引\n", executed))
	outputBuilder.WriteString(fmt.Sprintf("- 失败/跳过: %d 个索引\n", failed))
	outputBuilder.WriteString("- TTL设置: 30天自动清理过期数据\n")

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
