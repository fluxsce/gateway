package mongo

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"gateway/pkg/logger"
	mongoclient "gateway/pkg/mongo/client"
)

// defaultIndexName 集合默认主键索引，对齐时必须保留。
const defaultIndexName = "_id_"

// desiredIndexName 返回命令声明的索引名。
func desiredIndexName(cmd MongoIndexCommand) string {
	if cmd.IndexModel.Options != nil && cmd.IndexModel.Options.Name != "" {
		return cmd.IndexModel.Options.Name
	}
	return cmd.Description
}

// keepIndexNames 启动后该集合只保留默认 _id 与最小集声明的名字。
func keepIndexNames(cmds []MongoIndexCommand) map[string]struct{} {
	keep := map[string]struct{}{defaultIndexName: {}}
	for _, cmd := range cmds {
		if name := desiredIndexName(cmd); name != "" {
			keep[name] = struct{}{}
		}
	}
	return keep
}

// isIndexAlreadyExistsError 目标索引已经按同名建好，对齐后视为成功，不能 drop 重建。
func isIndexAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

// isIndexConflictError 同键但选项/名字不一致，删掉冲突索引后再建最小集。
func isIndexConflictError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "IndexOptionsConflict") ||
		strings.Contains(msg, "IndexKeySpecsConflict")
}

// isIndexPresentError 已存在或同键冲突，需在 drop 旧索引后再确认。
func isIndexPresentError(err error) bool {
	return isIndexAlreadyExistsError(err) || isIndexConflictError(err)
}

// listIndexNames 列出集合上已有索引名。
func listIndexNames(ctx context.Context, coll *mongoclient.Collection) ([]string, error) {
	cursor, err := coll.ListIndexes(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var names []string
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		name, _ := doc["name"].(string)
		if name != "" {
			names = append(names, name)
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

// dropObsoleteIndexes 删除不在保留名单中的索引（历史多余索引与错误选项留下的同键索引）。
// 先建最小集再删，避免查询空窗。删失败不中断启动。
func dropObsoleteIndexes(ctx context.Context, coll *mongoclient.Collection, collectionName string, keep map[string]struct{}) (dropped []string, errs []error) {
	names, err := listIndexNames(ctx, coll)
	if err != nil {
		return nil, []error{fmt.Errorf("列出索引失败: %w", err)}
	}
	for _, name := range names {
		if _, ok := keep[name]; ok {
			continue
		}
		if err := coll.DropIndex(ctx, name); err != nil {
			if strings.Contains(err.Error(), "index not found") {
				continue
			}
			logger.Warn("删除历史 Mongo 索引失败，将在下次启动重试",
				"collection", collectionName,
				"index", name,
				"error", err)
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		dropped = append(dropped, name)
		logger.Info("已删除历史 Mongo 索引",
			"collection", collectionName,
			"index", name)
	}
	return dropped, errs
}

// ensureCollectionIndexes 为单个系统日志集合建最小集，再自动删掉该集合上其余索引；TTL 被同键旧索引挡住时会重建。
// 调用方必须先用 IsManagedMongoLogCollection 过滤，本函数再拦一层，避免误删其它集合索引。
func ensureCollectionIndexes(ctx context.Context, coll *mongoclient.Collection, collectionName string, cmds []MongoIndexCommand) (executed, failed int, output string) {
	if !IsManagedMongoLogCollection(collectionName) {
		logger.Warn("拒绝为非系统日志集合对齐索引", "collection", collectionName)
		return 0, 0, fmt.Sprintf("  跳过：%s 不在系统日志集合白名单\n", collectionName)
	}

	var b strings.Builder
	keep := keepIndexNames(cmds)
	existing, listErr := listIndexNames(ctx, coll)
	if listErr != nil {
		logger.Warn("列出 Mongo 索引失败，将按创建结果对齐",
			"collection", collectionName,
			"error", listErr)
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, name := range existing {
		existingSet[name] = struct{}{}
	}

	pending := make([]MongoIndexCommand, 0, len(cmds))
	for i, cmd := range cmds {
		label := desiredIndexName(cmd)
		if _, ok := existingSet[label]; ok && label != "" {
			executed++
			b.WriteString(fmt.Sprintf("  %d. %s 已存在，跳过创建\n", i+1, label))
			logger.Info("Mongo 索引已存在，跳过创建",
				"collection", collectionName,
				"index", label)
			continue
		}
		logger.Info("正在创建 Mongo 索引，大集合可能较久",
			"collection", collectionName,
			"index", label)
		name, err := coll.CreateDriverIndex(ctx, cmd.IndexModel.ToMongoIndexModel())
		if err != nil {
			if isIndexPresentError(err) {
				logger.Info("Mongo 索引已存在或同键冲突，对齐后再确认",
					"collection", collectionName,
					"index", label,
					"error", err)
				b.WriteString(fmt.Sprintf("  %d. %s (已存在或同键，待对齐)\n", i+1, label))
				pending = append(pending, cmd)
				continue
			}
			logger.Warn("创建 Mongo 索引失败",
				"collection", collectionName,
				"index", label,
				"error", err)
			b.WriteString(fmt.Sprintf("  %d. %s 失败: %v\n", i+1, label, err))
			failed++
			continue
		}
		executed++
		b.WriteString(fmt.Sprintf("  %d. %s 已创建\n", i+1, name))
		logger.Info("Mongo 索引创建成功",
			"collection", collectionName,
			"index", name)
	}

	// 新索引尚未全部到位时不能删旧索引，否则列表/详情会空窗扫表
	if failed > 0 {
		logger.Warn("最小集未全部创建成功，跳过删除旧索引，下次启动再对齐",
			"collection", collectionName,
			"failed", failed)
		b.WriteString("  最小集未就绪，保留旧索引\n")
		return executed, failed, b.String()
	}

	dropped, dropErrs := dropObsoleteIndexes(ctx, coll, collectionName, keep)
	if len(dropped) > 0 {
		b.WriteString(fmt.Sprintf("  已自动删除旧索引: %s\n", strings.Join(dropped, ", ")))
	}
	if len(dropErrs) > 0 {
		failed += len(dropErrs)
		for _, e := range dropErrs {
			b.WriteString(fmt.Sprintf("  删除旧索引失败: %v\n", e))
		}
	}

	for _, cmd := range pending {
		label := desiredIndexName(cmd)
		name, err := coll.CreateDriverIndex(ctx, cmd.IndexModel.ToMongoIndexModel())
		if err == nil {
			executed++
			b.WriteString(fmt.Sprintf("  %s 对齐后创建成功\n", name))
			logger.Info("Mongo 索引对齐后创建成功",
				"collection", collectionName,
				"index", name)
			continue
		}
		if isIndexAlreadyExistsError(err) {
			executed++
			b.WriteString(fmt.Sprintf("  %s 已对齐\n", label))
			continue
		}
		if isIndexConflictError(err) && label != "" && label != defaultIndexName {
			if dropErr := coll.DropIndex(ctx, label); dropErr != nil && !strings.Contains(dropErr.Error(), "index not found") {
				logger.Warn("删除选项不匹配的 Mongo 索引失败",
					"collection", collectionName,
					"index", label,
					"error", dropErr)
				b.WriteString(fmt.Sprintf("  %s 选项冲突且无法重建: %v\n", label, dropErr))
				failed++
				continue
			}
			name, err = coll.CreateDriverIndex(ctx, cmd.IndexModel.ToMongoIndexModel())
			if err == nil {
				executed++
				b.WriteString(fmt.Sprintf("  %s 已按最小集重建（含 TTL/键顺序）\n", name))
				logger.Info("Mongo 索引已按最小集重建",
					"collection", collectionName,
					"index", name)
				continue
			}
		}
		logger.Warn("对齐后创建 Mongo 索引仍失败",
			"collection", collectionName,
			"index", label,
			"error", err)
		b.WriteString(fmt.Sprintf("  %s 对齐后仍失败: %v\n", label, err))
		failed++
	}

	return executed, failed, b.String()
}
