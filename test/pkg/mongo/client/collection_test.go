package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/pkg/mongo/types"
)

// TestCollectionCRUD 测试集合的完整CRUD操作
func TestCollectionCRUD(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_crud_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	t.Run("Basic Collection Properties", func(t *testing.T) {
		assert.Equal(t, collectionName, collection.Name(), "集合名称应该匹配")
		assert.NotNil(t, collection.Database(), "数据库引用不应为空")
		assert.Equal(t, testDatabase.Name(), collection.Database().Name(), "数据库名称应该匹配")
	})

	// 测试数据
	testDocs := []types.Document{
		{
			"_id":     "doc1",
			"name":    "张三",
			"age":     25,
			"city":    "北京",
			"score":   85.5,
			"active":  true,
			"tags":    []string{"golang", "mongodb"},
			"profile": types.Document{"level": "beginner", "points": 100},
			"created": time.Now(),
		},
		{
			"_id":     "doc2",
			"name":    "李四",
			"age":     30,
			"city":    "上海",
			"score":   92.0,
			"active":  true,
			"tags":    []string{"python", "mongodb"},
			"profile": types.Document{"level": "intermediate", "points": 200},
			"created": time.Now(),
		},
		{
			"_id":     "doc3",
			"name":    "王五",
			"age":     28,
			"city":    "广州",
			"score":   78.0,
			"active":  false,
			"tags":    []string{"java", "mysql"},
			"profile": types.Document{"level": "advanced", "points": 300},
			"created": time.Now(),
		},
	}

	t.Run("InsertOne", func(t *testing.T) {
		result, err := collection.InsertOne(ctx, testDocs[0], nil)
		assert.NoError(t, err, "插入单个文档应该成功")
		assert.NotNil(t, result, "插入结果不应为空")
		assert.Equal(t, "doc1", result.InsertedID, "插入的ID应该匹配")
		t.Logf("插入文档ID: %v", result.InsertedID)
	})

	t.Run("InsertMany", func(t *testing.T) {
		result, err := collection.InsertMany(ctx, testDocs[1:], nil)
		assert.NoError(t, err, "批量插入应该成功")
		assert.NotNil(t, result, "插入结果不应为空")
		assert.Len(t, result.InsertedIDs, 2, "应该插入2个文档")
		assert.Contains(t, result.InsertedIDs, "doc2", "应该包含doc2")
		assert.Contains(t, result.InsertedIDs, "doc3", "应该包含doc3")
		t.Logf("批量插入IDs: %v", result.InsertedIDs)
	})

	t.Run("Count", func(t *testing.T) {
		// 计算所有文档
		count, err := collection.Count(ctx, types.Filter{}, nil)
		assert.NoError(t, err, "计数应该成功")
		assert.Equal(t, int64(3), count, "应该有3个文档")

		// 计算活跃用户
		activeCount, err := collection.Count(ctx, types.Filter{"active": true}, nil)
		assert.NoError(t, err, "计数活跃用户应该成功")
		assert.Equal(t, int64(2), activeCount, "应该有2个活跃用户")

		// 测试计数选项
		limitCount, err := collection.Count(ctx, types.Filter{}, &types.CountOptions{
			Limit: &[]int64{2}[0],
		})
		assert.NoError(t, err, "带限制的计数应该成功")
		assert.Equal(t, int64(2), limitCount, "限制计数应该返回2")

		t.Logf("总文档数: %d, 活跃用户数: %d, 限制计数: %d", count, activeCount, limitCount)
	})

	t.Run("FindOne", func(t *testing.T) {
		// 基本查找
		result := collection.FindOne(ctx, types.Filter{"_id": "doc1"}, nil)
		assert.NotNil(t, result, "查找结果不应为空")

		var doc types.Document
		err := result.Decode(&doc)
		assert.NoError(t, err, "解码应该成功")
		assert.Equal(t, "张三", doc["name"], "姓名应该匹配")

		// 带投影的查找
		result = collection.FindOne(ctx, types.Filter{"_id": "doc1"}, &types.FindOneOptions{
			Projection: types.Document{"name": 1, "age": 1},
		})
		assert.NotNil(t, result, "投影查找结果不应为空")

		var projectedDoc types.Document
		err = result.Decode(&projectedDoc)
		assert.NoError(t, err, "投影解码应该成功")
		assert.Equal(t, "张三", projectedDoc["name"], "投影姓名应该匹配")
		assert.Equal(t, float64(25), toFloat64(projectedDoc["age"]), "投影年龄应该匹配")
		assert.NotContains(t, projectedDoc, "city", "投影结果不应包含city")

		t.Logf("查找到的文档: %v", doc)
		t.Logf("投影文档: %v", projectedDoc)
	})

	t.Run("Find", func(t *testing.T) {
		// 查找所有活跃用户
		cursor, err := collection.Find(ctx, types.Filter{"active": true}, nil)
		assert.NoError(t, err, "查找多个文档应该成功")
		assert.NotNil(t, cursor, "游标不应为空")

		var docs []types.Document
		err = cursor.All(ctx, &docs)
		assert.NoError(t, err, "解码所有文档应该成功")
		assert.Len(t, docs, 2, "应该找到2个活跃用户")

		// 带排序和限制的查找
		cursor, err = collection.Find(ctx, types.Filter{}, &types.FindOptions{
			Sort:  types.Document{"age": 1}, // 按年龄升序
			Limit: &[]int64{2}[0],
		})
		assert.NoError(t, err, "排序查找应该成功")

		var sortedDocs []types.Document
		err = cursor.All(ctx, &sortedDocs)
		assert.NoError(t, err, "解码排序文档应该成功")
		assert.Len(t, sortedDocs, 2, "应该返回2个文档")

		// 验证排序
		if len(sortedDocs) >= 2 {
			age1 := toFloat64(sortedDocs[0]["age"])
			age2 := toFloat64(sortedDocs[1]["age"])
			assert.LessOrEqual(t, age1, age2, "第一个文档年龄应该小于等于第二个")
		}

		t.Logf("活跃用户数: %d", len(docs))
		t.Logf("排序后的前2个用户年龄: %v, %v", sortedDocs[0]["age"], sortedDocs[1]["age"])
	})

	t.Run("UpdateOne", func(t *testing.T) {
		// 更新单个文档
		update := types.Update{
			"$set": types.Document{
				"score":   90.0,
				"updated": time.Now(),
			},
			"$inc": types.Document{
				"profile.points": 50,
			},
		}

		result, err := collection.UpdateOne(ctx, types.Filter{"_id": "doc1"}, update, nil)
		assert.NoError(t, err, "更新单个文档应该成功")
		assert.NotNil(t, result, "更新结果不应为空")
		assert.Equal(t, int64(1), result.MatchedCount, "应该匹配1个文档")
		assert.Equal(t, int64(1), result.ModifiedCount, "应该修改1个文档")

		// 验证更新结果
		findResult := collection.FindOne(ctx, types.Filter{"_id": "doc1"}, nil)
		var updatedDoc types.Document
		err = findResult.Decode(&updatedDoc)
		assert.NoError(t, err, "查找更新后的文档应该成功")
		assert.Equal(t, 90.0, updatedDoc["score"], "分数应该被更新")
		assert.Contains(t, updatedDoc, "updated", "应该包含updated字段")

		t.Logf("更新结果: 匹配=%d, 修改=%d", result.MatchedCount, result.ModifiedCount)
	})

	t.Run("UpdateMany", func(t *testing.T) {
		// 更新多个文档
		update := types.Update{
			"$set": types.Document{
				"batch_updated": true,
				"batch_time":    time.Now(),
			},
		}

		result, err := collection.UpdateMany(ctx, types.Filter{"active": true}, update, nil)
		assert.NoError(t, err, "批量更新应该成功")
		assert.NotNil(t, result, "更新结果不应为空")
		assert.Equal(t, int64(2), result.MatchedCount, "应该匹配2个文档")
		assert.Equal(t, int64(2), result.ModifiedCount, "应该修改2个文档")

		// 验证批量更新结果
		cursor, err := collection.Find(ctx, types.Filter{"batch_updated": true}, nil)
		assert.NoError(t, err, "查找批量更新的文档应该成功")

		var updatedDocs []types.Document
		err = cursor.All(ctx, &updatedDocs)
		assert.NoError(t, err, "解码批量更新的文档应该成功")
		assert.Len(t, updatedDocs, 2, "应该有2个文档被批量更新")

		t.Logf("批量更新结果: 匹配=%d, 修改=%d", result.MatchedCount, result.ModifiedCount)
	})

	t.Run("DeleteOne", func(t *testing.T) {
		// 删除单个文档
		result, err := collection.DeleteOne(ctx, types.Filter{"_id": "doc3"}, nil)
		assert.NoError(t, err, "删除单个文档应该成功")
		assert.NotNil(t, result, "删除结果不应为空")
		assert.Equal(t, int64(1), result.DeletedCount, "应该删除1个文档")

		// 验证文档已删除
		findResult := collection.FindOne(ctx, types.Filter{"_id": "doc3"}, nil)
		var deletedDoc types.Document
		err = findResult.Decode(&deletedDoc)
		assert.Error(t, err, "查找已删除的文档应该失败")

		// 验证总数减少
		count, err := collection.Count(ctx, types.Filter{}, nil)
		assert.NoError(t, err, "计数应该成功")
		assert.Equal(t, int64(2), count, "删除后应该剩余2个文档")

		t.Logf("删除结果: 删除数=%d", result.DeletedCount)
	})

	t.Run("DeleteMany", func(t *testing.T) {
		// 删除多个文档
		result, err := collection.DeleteMany(ctx, types.Filter{"active": true}, nil)
		assert.NoError(t, err, "批量删除应该成功")
		assert.NotNil(t, result, "删除结果不应为空")
		assert.Equal(t, int64(2), result.DeletedCount, "应该删除2个文档")

		// 验证所有文档已删除
		count, err := collection.Count(ctx, types.Filter{}, nil)
		assert.NoError(t, err, "计数应该成功")
		assert.Equal(t, int64(0), count, "批量删除后应该没有文档")

		t.Logf("批量删除结果: 删除数=%d", result.DeletedCount)
	})
}

// TestCollectionOptions 测试集合操作的各种选项
func TestCollectionOptions(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_options_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	// 准备测试数据
	docs := []types.Document{
		{"_id": 1, "name": "Alice", "age": 25, "score": 85},
		{"_id": 2, "name": "Bob", "age": 30, "score": 90},
		{"_id": 3, "name": "Charlie", "age": 35, "score": 78},
		{"_id": 4, "name": "Diana", "age": 28, "score": 92},
		{"_id": 5, "name": "Eve", "age": 32, "score": 88},
	}

	_, err := collection.InsertMany(ctx, docs, nil)
	require.NoError(t, err, "插入测试数据应该成功")

	t.Run("FindOptions", func(t *testing.T) {
		// 测试排序
		cursor, err := collection.Find(ctx, types.Filter{}, &types.FindOptions{
			Sort: types.Document{"age": -1}, // 按年龄降序
		})
		assert.NoError(t, err, "排序查找应该成功")

		var sortedDocs []types.Document
		err = cursor.All(ctx, &sortedDocs)
		assert.NoError(t, err, "解码排序文档应该成功")
		assert.Len(t, sortedDocs, 5, "应该返回所有文档")

		// 验证排序
		for i := 0; i < len(sortedDocs)-1; i++ {
			age1 := toFloat64(sortedDocs[i]["age"])
			age2 := toFloat64(sortedDocs[i+1]["age"])
			assert.GreaterOrEqual(t, age1, age2, "年龄应该降序排列")
		}

		// 测试限制和跳过
		cursor, err = collection.Find(ctx, types.Filter{}, &types.FindOptions{
			Sort:  types.Document{"age": 1},
			Skip:  &[]int64{1}[0],
			Limit: &[]int64{2}[0],
		})
		assert.NoError(t, err, "带skip和limit的查找应该成功")

		var limitedDocs []types.Document
		err = cursor.All(ctx, &limitedDocs)
		assert.NoError(t, err, "解码限制文档应该成功")
		assert.Len(t, limitedDocs, 2, "应该返回2个文档")

		// 测试投影
		cursor, err = collection.Find(ctx, types.Filter{}, &types.FindOptions{
			Projection: types.Document{"name": 1, "age": 1},
		})
		assert.NoError(t, err, "投影查找应该成功")

		var projectedDocs []types.Document
		err = cursor.All(ctx, &projectedDocs)
		assert.NoError(t, err, "解码投影文档应该成功")

		for _, doc := range projectedDocs {
			assert.Contains(t, doc, "name", "投影结果应该包含name")
			assert.Contains(t, doc, "age", "投影结果应该包含age")
			assert.NotContains(t, doc, "score", "投影结果不应该包含score")
		}

		t.Logf("排序后的年龄: %v", extractAges(sortedDocs))
		t.Logf("限制查询返回的文档数: %d", len(limitedDocs))
		t.Logf("投影查询返回的字段: %v", getDocumentKeys(projectedDocs[0]))
	})

	t.Run("InsertOptions", func(t *testing.T) {
		// 测试 InsertOne 选项
		doc := types.Document{"_id": 6, "name": "Frank", "age": 40}

		result, err := collection.InsertOne(ctx, doc, &types.InsertOneOptions{
			BypassDocumentValidation: &[]bool{false}[0],
		})
		assert.NoError(t, err, "带选项的插入应该成功")
		assert.Equal(t, float64(6), toFloat64(result.InsertedID), "插入的ID应该匹配")

		// 测试 InsertMany 选项
		newDocs := []types.Document{
			{"_id": 7, "name": "Grace", "age": 26},
			{"_id": 8, "name": "Henry", "age": 31},
		}

		result2, err := collection.InsertMany(ctx, newDocs, &types.InsertManyOptions{
			Ordered: &[]bool{true}[0],
		})
		assert.NoError(t, err, "带选项的批量插入应该成功")
		assert.Len(t, result2.InsertedIDs, 2, "应该插入2个文档")

		t.Logf("插入选项测试完成，新增文档数: %d", len(result2.InsertedIDs)+1)
	})

	t.Run("UpdateOptions", func(t *testing.T) {
		// 测试 upsert 选项
		update := types.Update{
			"$set": types.Document{
				"name": "Upserted User",
				"age":  99,
			},
		}

		result, err := collection.UpdateOne(ctx, types.Filter{"_id": 999}, update, &types.UpdateOptions{
			Upsert: &[]bool{true}[0],
		})
		assert.NoError(t, err, "upsert更新应该成功")
		assert.Equal(t, int64(0), result.MatchedCount, "应该匹配0个文档")
		assert.Equal(t, int64(1), result.UpsertedCount, "应该upsert 1个文档")
		assert.Equal(t, float64(999), toFloat64(result.UpsertedID), "upsert的ID应该匹配")

		// 验证 upsert 结果
		findResult := collection.FindOne(ctx, types.Filter{"_id": 999}, nil)
		var upsertedDoc types.Document
		err = findResult.Decode(&upsertedDoc)
		assert.NoError(t, err, "查找upsert文档应该成功")
		assert.Equal(t, "Upserted User", upsertedDoc["name"], "upsert的名称应该匹配")

		t.Logf("Upsert结果: 匹配=%d, 修改=%d, Upsert=%d, UpsertID=%v",
			result.MatchedCount, result.ModifiedCount, result.UpsertedCount, result.UpsertedID)
	})
}

// TestCollectionDrop 测试集合删除
func TestCollectionDrop(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_drop_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 插入一些测试数据
	docs := []types.Document{
		{"_id": 1, "name": "test1"},
		{"_id": 2, "name": "test2"},
	}

	_, err := collection.InsertMany(ctx, docs, nil)
	require.NoError(t, err, "插入测试数据应该成功")

	// 验证集合存在
	count, err := collection.Count(ctx, types.Filter{}, nil)
	assert.NoError(t, err, "计数应该成功")
	assert.Equal(t, int64(2), count, "应该有2个文档")

	// 验证集合在数据库中存在
	collectionNames, err := testDatabase.ListCollectionNames(ctx, nil)
	assert.NoError(t, err, "列出集合名称应该成功")
	assert.Contains(t, collectionNames, collectionName, "集合应该存在于数据库中")

	// 删除集合
	err = collection.Drop(ctx)
	assert.NoError(t, err, "删除集合应该成功")

	// 验证集合已被删除
	collectionNamesAfterDrop, err := testDatabase.ListCollectionNames(ctx, nil)
	assert.NoError(t, err, "删除后列出集合名称应该成功")
	assert.NotContains(t, collectionNamesAfterDrop, collectionName, "集合应该不存在于数据库中")

	t.Logf("集合 %s 已成功删除", collectionName)
}

// TestCollectionErrorCases 测试集合操作的错误情况
func TestCollectionErrorCases(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_errors_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	t.Run("DuplicateKeyError", func(t *testing.T) {
		// 插入重复ID的文档
		doc := types.Document{"_id": "duplicate_test", "name": "test"}

		_, err := collection.InsertOne(ctx, doc, nil)
		assert.NoError(t, err, "第一次插入应该成功")

		// 再次插入相同ID的文档应该失败
		_, err = collection.InsertOne(ctx, doc, nil)
		assert.Error(t, err, "插入重复ID应该失败")
		assert.Contains(t, err.Error(), "duplicate", "错误信息应该包含duplicate")
	})

	t.Run("InvalidUpdateOperation", func(t *testing.T) {
		// 插入测试文档
		doc := types.Document{"_id": "update_test", "name": "test"}
		_, err := collection.InsertOne(ctx, doc, nil)
		require.NoError(t, err, "插入测试文档应该成功")

		// 无效的更新操作
		invalidUpdate := types.Update{
			"$invalidOperator": types.Document{"name": "invalid"},
		}

		_, err = collection.UpdateOne(ctx, types.Filter{"_id": "update_test"}, invalidUpdate, nil)
		assert.Error(t, err, "无效的更新操作应该失败")
	})

	t.Run("NoMatchingDocuments", func(t *testing.T) {
		// 查找不存在的文档
		result := collection.FindOne(ctx, types.Filter{"_id": "nonexistent"}, nil)
		assert.NotNil(t, result, "查找结果不应为空")

		var doc types.Document
		err := result.Decode(&doc)
		assert.Error(t, err, "解码不存在的文档应该失败")

		// 更新不存在的文档
		update := types.Update{"$set": types.Document{"name": "updated"}}
		updateResult, err := collection.UpdateOne(ctx, types.Filter{"_id": "nonexistent"}, update, nil)
		assert.NoError(t, err, "更新不存在的文档应该成功但无影响")
		assert.Equal(t, int64(0), updateResult.MatchedCount, "应该匹配0个文档")
		assert.Equal(t, int64(0), updateResult.ModifiedCount, "应该修改0个文档")

		// 删除不存在的文档
		deleteResult, err := collection.DeleteOne(ctx, types.Filter{"_id": "nonexistent"}, nil)
		assert.NoError(t, err, "删除不存在的文档应该成功但无影响")
		assert.Equal(t, int64(0), deleteResult.DeletedCount, "应该删除0个文档")
	})
}

// TestCollectionBenchmarks 集合操作的基准测试
func TestCollectionBenchmarks(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_benchmark_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	// 准备大量测试数据
	docs := make([]types.Document, 1000)
	for i := 0; i < 1000; i++ {
		docs[i] = types.Document{
			"_id":     i,
			"name":    "user_" + time.Now().Format("20060102150405"),
			"age":     20 + (i % 50),
			"score":   float64(60 + (i % 40)),
			"active":  i%2 == 0,
			"created": time.Now(),
		}
	}

	// 批量插入测试数据
	_, err := collection.InsertMany(ctx, docs, nil)
	require.NoError(t, err, "插入基准测试数据应该成功")

	t.Run("BenchmarkOperations", func(t *testing.T) {
		// 测试查找性能
		start := time.Now()
		cursor, err := collection.Find(ctx, types.Filter{"active": true}, nil)
		assert.NoError(t, err, "查找应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码应该成功")

		findDuration := time.Since(start)
		t.Logf("查找 %d 个文档耗时: %v", len(results), findDuration)

		// 测试更新性能
		start = time.Now()
		updateResult, err := collection.UpdateMany(ctx,
			types.Filter{"active": true},
			types.Update{"$set": types.Document{"batch_updated": true}},
			nil)
		assert.NoError(t, err, "批量更新应该成功")

		updateDuration := time.Since(start)
		t.Logf("更新 %d 个文档耗时: %v", updateResult.ModifiedCount, updateDuration)

		// 测试计数性能
		start = time.Now()
		count, err := collection.Count(ctx, types.Filter{"batch_updated": true}, nil)
		assert.NoError(t, err, "计数应该成功")

		countDuration := time.Since(start)
		t.Logf("计数 %d 个文档耗时: %v", count, countDuration)
	})
}

// 辅助函数

// extractAges 提取文档中的年龄字段
func extractAges(docs []types.Document) []float64 {
	ages := make([]float64, len(docs))
	for i, doc := range docs {
		ages[i] = toFloat64(doc["age"])
	}
	return ages
}

// getDocumentKeys 获取文档的所有键
func getDocumentKeys(doc types.Document) []string {
	keys := make([]string, 0, len(doc))
	for key := range doc {
		keys = append(keys, key)
	}
	return keys
}

// toFloat64 将不同类型的数值转换为 float64
func toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	default:
		return 0
	}
}

// toInt64 将不同类型的数值转换为 int64
func toInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	default:
		return 0
	}
}
