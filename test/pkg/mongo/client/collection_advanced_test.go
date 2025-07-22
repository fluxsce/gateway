package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gateway/pkg/mongo/types"
)

// TestCollectionAggregate 测试集合的聚合操作
func TestCollectionAggregate(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_aggregate_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	// 准备聚合测试数据
	testDocs := []types.Document{
		{
			"_id":        1,
			"name":       "张三",
			"department": "技术部",
			"salary":     8000,
			"age":        25,
			"city":       "北京",
			"skills":     []string{"golang", "mongodb", "redis"},
			"projects": []types.Document{
				{"name": "项目A", "status": "完成", "hours": 120},
				{"name": "项目B", "status": "进行中", "hours": 80},
			},
			"joinDate": time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			"_id":        2,
			"name":       "李四",
			"department": "技术部",
			"salary":     12000,
			"age":        30,
			"city":       "上海",
			"skills":     []string{"python", "mongodb", "elasticsearch"},
			"projects": []types.Document{
				{"name": "项目C", "status": "完成", "hours": 150},
				{"name": "项目D", "status": "完成", "hours": 100},
			},
			"joinDate": time.Date(2021, 6, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			"_id":        3,
			"name":       "王五",
			"department": "销售部",
			"salary":     6000,
			"age":        28,
			"city":       "广州",
			"skills":     []string{"销售", "客户管理"},
			"projects": []types.Document{
				{"name": "销售项目E", "status": "进行中", "hours": 60},
			},
			"joinDate": time.Date(2022, 3, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			"_id":        4,
			"name":       "赵六",
			"department": "技术部",
			"salary":     15000,
			"age":        35,
			"city":       "深圳",
			"skills":     []string{"java", "spring", "mysql", "mongodb"},
			"projects": []types.Document{
				{"name": "项目F", "status": "完成", "hours": 200},
				{"name": "项目G", "status": "完成", "hours": 180},
				{"name": "项目H", "status": "进行中", "hours": 90},
			},
			"joinDate": time.Date(2020, 9, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			"_id":        5,
			"name":       "孙七",
			"department": "销售部",
			"salary":     7000,
			"age":        26,
			"city":       "成都",
			"skills":     []string{"市场营销", "数据分析"},
			"projects": []types.Document{
				{"name": "销售项目I", "status": "完成", "hours": 90},
				{"name": "销售项目J", "status": "进行中", "hours": 45},
			},
			"joinDate": time.Date(2023, 1, 8, 0, 0, 0, 0, time.UTC),
		},
	}

	// 插入测试数据
	_, err := collection.InsertMany(ctx, testDocs, nil)
	require.NoError(t, err, "插入聚合测试数据应该成功")

	t.Run("BasicAggregation", func(t *testing.T) {
		// 基本聚合：按部门分组统计平均薪资
		pipeline := types.Pipeline{
			{
				"$group": types.Document{
					"_id":        "$department",
					"avgSalary":  types.Document{"$avg": "$salary"},
					"totalCount": types.Document{"$sum": 1},
					"maxSalary":  types.Document{"$max": "$salary"},
					"minSalary":  types.Document{"$min": "$salary"},
				},
			},
			{
				"$sort": types.Document{"avgSalary": -1},
			},
		}

		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "基本聚合应该成功")
		assert.NotNil(t, cursor, "聚合游标不应为空")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码聚合结果应该成功")
		assert.Len(t, results, 2, "应该有2个部门的统计结果")

		// 验证聚合结果
		for _, result := range results {
			department := result["_id"].(string)
			avgSalary := toFloat64(result["avgSalary"])
			totalCount := toInt64(result["totalCount"])

			if department == "技术部" {
				assert.Equal(t, int64(3), totalCount, "技术部应该有3个员工")
				assert.Equal(t, float64(11666.666666666666), avgSalary, "技术部平均薪资应该正确")
			} else if department == "销售部" {
				assert.Equal(t, int64(2), totalCount, "销售部应该有2个员工")
				assert.Equal(t, float64(6500), avgSalary, "销售部平均薪资应该正确")
			}
		}

		t.Logf("部门统计结果: %+v", results)
	})

	t.Run("MatchAndProject", func(t *testing.T) {
		// 聚合管道：筛选技术部员工并投影特定字段
		pipeline := types.Pipeline{
			{
				"$match": types.Document{
					"department": "技术部",
					"salary":     types.Document{"$gte": 8000},
				},
			},
			{
				"$project": types.Document{
					"name":       1,
					"salary":     1,
					"age":        1,
					"skillCount": types.Document{"$size": "$skills"},
					"isHighSalary": types.Document{
						"$gte": []interface{}{"$salary", 12000},
					},
				},
			},
			{
				"$sort": types.Document{"salary": -1},
			},
		}

		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "筛选和投影聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码筛选投影结果应该成功")
		assert.Len(t, results, 3, "应该有3个技术部高薪员工")

		// 验证投影结果结构
		for _, result := range results {
			assert.Contains(t, result, "name", "结果应该包含name字段")
			assert.Contains(t, result, "salary", "结果应该包含salary字段")
			assert.Contains(t, result, "age", "结果应该包含age字段")
			assert.Contains(t, result, "skillCount", "结果应该包含skillCount字段")
			assert.Contains(t, result, "isHighSalary", "结果应该包含isHighSalary字段")
			assert.NotContains(t, result, "department", "结果不应该包含department字段")
			assert.NotContains(t, result, "city", "结果不应该包含city字段")
		}

		t.Logf("技术部高薪员工: %+v", results)
	})

	t.Run("UnwindAndGroup", func(t *testing.T) {
		// 聚合管道：展开技能数组并统计技能频次
		pipeline := types.Pipeline{
			{
				"$unwind": "$skills",
			},
			{
				"$group": types.Document{
					"_id":       "$skills",
					"count":     types.Document{"$sum": 1},
					"employees": types.Document{"$push": "$name"},
				},
			},
			{
				"$sort": types.Document{"count": -1},
			},
			{
				"$limit": 5,
			},
		}

		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "展开和分组聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码展开分组结果应该成功")
		assert.True(t, len(results) <= 5, "结果数量应该不超过5个")

		// 验证技能统计结果
		foundMongoDB := false
		for _, result := range results {
			skill := result["_id"].(string)
			count := toInt64(result["count"])

			// 处理 MongoDB 驱动返回的数组类型
			var employees []interface{}
			if empArray, ok := result["employees"].([]interface{}); ok {
				employees = empArray
			} else if empPrimitive, ok := result["employees"].(primitive.A); ok {
				employees = []interface{}(empPrimitive)
			} else {
				t.Errorf("无法解析 employees 字段类型: %T", result["employees"])
				continue
			}

			if skill == "mongodb" {
				foundMongoDB = true
				assert.Equal(t, int64(3), count, "mongodb技能应该有3个员工")
				assert.Len(t, employees, 3, "mongodb技能的员工列表应该有3个")
			}

			assert.Greater(t, count, int64(0), "技能计数应该大于0")
			assert.Len(t, employees, int(count), "员工列表长度应该等于计数")
		}

		assert.True(t, foundMongoDB, "应该找到mongodb技能")
		t.Logf("技能统计结果: %+v", results)
	})

	t.Run("ComplexAggregation", func(t *testing.T) {
		// 复杂聚合：统计每个城市的员工情况和项目信息
		pipeline := types.Pipeline{
			{
				"$unwind": "$projects",
			},
			{
				"$group": types.Document{
					"_id":           "$city",
					"employeeCount": types.Document{"$addToSet": "$name"},
					"totalProjects": types.Document{"$sum": 1},
					"completedProjects": types.Document{
						"$sum": types.Document{
							"$cond": types.Document{
								"if":   types.Document{"$eq": []interface{}{"$projects.status", "完成"}},
								"then": 1,
								"else": 0,
							},
						},
					},
					"totalHours":      types.Document{"$sum": "$projects.hours"},
					"avgSalaryByCity": types.Document{"$avg": "$salary"},
				},
			},
			{
				"$project": types.Document{
					"city":              "$_id",
					"employeeCount":     types.Document{"$size": "$employeeCount"},
					"totalProjects":     1,
					"completedProjects": 1,
					"inProgressProjects": types.Document{
						"$subtract": []interface{}{"$totalProjects", "$completedProjects"},
					},
					"totalHours":      1,
					"avgSalaryByCity": types.Document{"$round": []interface{}{"$avgSalaryByCity", 2}},
					"completionRate": types.Document{
						"$round": []interface{}{
							types.Document{
								"$multiply": []interface{}{
									types.Document{
										"$divide": []interface{}{"$completedProjects", "$totalProjects"},
									},
									100,
								},
							},
							2,
						},
					},
				},
			},
			{
				"$sort": types.Document{"totalHours": -1},
			},
		}

		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "复杂聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码复杂聚合结果应该成功")
		assert.Greater(t, len(results), 0, "应该有聚合结果")

		// 验证复杂聚合结果结构
		for _, result := range results {
			city := result["city"].(string)
			employeeCount := toInt64(result["employeeCount"])
			totalProjects := toInt64(result["totalProjects"])
			completedProjects := toInt64(result["completedProjects"])
			totalHours := toInt64(result["totalHours"])
			completionRate := toFloat64(result["completionRate"])

			assert.NotEmpty(t, city, "城市名称不应为空")
			assert.Greater(t, employeeCount, int64(0), "员工数量应该大于0")
			assert.Greater(t, totalProjects, int64(0), "项目总数应该大于0")
			assert.GreaterOrEqual(t, completedProjects, int64(0), "完成项目数应该大于等于0")
			assert.Greater(t, totalHours, int64(0), "总工时应该大于0")
			assert.GreaterOrEqual(t, completionRate, float64(0), "完成率应该大于等于0")
			assert.LessOrEqual(t, completionRate, float64(100), "完成率应该小于等于100")
		}

		t.Logf("城市项目统计结果: %+v", results)
	})
}

// TestCollectionAggregateOptions 测试聚合操作选项
func TestCollectionAggregateOptions(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_aggregate_options_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	// 准备大量测试数据
	docs := make([]types.Document, 100)
	for i := 0; i < 100; i++ {
		docs[i] = types.Document{
			"_id":       i,
			"name":      "user_" + time.Now().Format("20060102150405"),
			"category":  "category_" + string(rune('A'+(i%5))),
			"value":     float64(i * 10),
			"timestamp": time.Now().Add(time.Duration(i) * time.Hour),
			"tags":      []string{"tag1", "tag2", "tag3"},
			"metadata":  types.Document{"source": "test", "version": i % 3},
		}
	}

	_, err := collection.InsertMany(ctx, docs, nil)
	require.NoError(t, err, "插入聚合选项测试数据应该成功")

	t.Run("AggregateWithBatchSize", func(t *testing.T) {
		pipeline := types.Pipeline{
			{
				"$group": types.Document{
					"_id":      "$category",
					"count":    types.Document{"$sum": 1},
					"avgValue": types.Document{"$avg": "$value"},
				},
			},
			{
				"$sort": types.Document{"count": -1},
			},
		}

		// 使用较小的批次大小
		batchSize := int32(2)
		options := &types.AggregateOptions{
			BatchSize: &batchSize,
		}

		cursor, err := collection.Aggregate(ctx, pipeline, options)
		assert.NoError(t, err, "带批次大小的聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码带批次大小的聚合结果应该成功")
		assert.Len(t, results, 5, "应该有5个分类的统计结果")

		t.Logf("带批次大小的聚合结果: %+v", results)
	})

	t.Run("AggregateWithAllowDiskUse", func(t *testing.T) {
		// 复杂聚合管道，可能需要磁盘存储
		pipeline := types.Pipeline{
			{
				"$unwind": "$tags",
			},
			{
				"$group": types.Document{
					"_id": types.Document{
						"category": "$category",
						"tag":      "$tags",
					},
					"documents": types.Document{"$push": "$$ROOT"},
					"count":     types.Document{"$sum": 1},
				},
			},
			{
				"$sort": types.Document{"count": -1},
			},
		}

		allowDiskUse := true
		options := &types.AggregateOptions{
			AllowDiskUse: &allowDiskUse,
		}

		cursor, err := collection.Aggregate(ctx, pipeline, options)
		assert.NoError(t, err, "允许磁盘使用的聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码允许磁盘使用的聚合结果应该成功")
		assert.Greater(t, len(results), 0, "应该有聚合结果")

		t.Logf("允许磁盘使用的聚合结果数量: %d", len(results))
	})

	t.Run("AggregateWithMaxTime", func(t *testing.T) {
		pipeline := types.Pipeline{
			{
				"$group": types.Document{
					"_id":        "$category",
					"totalValue": types.Document{"$sum": "$value"},
				},
			},
		}

		// 设置最大执行时间
		maxTime := 30 * time.Second
		options := &types.AggregateOptions{
			MaxTime: &maxTime,
		}

		cursor, err := collection.Aggregate(ctx, pipeline, options)
		assert.NoError(t, err, "带最大时间的聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码带最大时间的聚合结果应该成功")
		assert.Len(t, results, 5, "应该有5个分类的汇总结果")

		t.Logf("带最大时间的聚合结果: %+v", results)
	})
}

// TestCollectionAggregateErrorCases 测试聚合操作的错误情况
func TestCollectionAggregateErrorCases(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_aggregate_errors_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	// 插入少量测试数据
	docs := []types.Document{
		{"_id": 1, "name": "test1", "value": 10},
		{"_id": 2, "name": "test2", "value": 20},
	}

	_, err := collection.InsertMany(ctx, docs, nil)
	require.NoError(t, err, "插入错误测试数据应该成功")

	t.Run("InvalidPipelineStage", func(t *testing.T) {
		// 无效的聚合阶段
		pipeline := types.Pipeline{
			{
				"$invalidStage": types.Document{
					"field": "value",
				},
			},
		}

		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.Error(t, err, "无效的聚合阶段应该失败")
		assert.Nil(t, cursor, "错误情况下游标应该为空")
		assert.Contains(t, err.Error(), "failed to execute aggregation", "错误信息应该说明聚合执行失败")
	})

	t.Run("InvalidFieldReference", func(t *testing.T) {
		// 引用不存在的字段
		pipeline := types.Pipeline{
			{
				"$group": types.Document{
					"_id":      "$nonexistentField",
					"count":    types.Document{"$sum": 1},
					"avgValue": types.Document{"$avg": "$anotherNonexistentField"},
				},
			},
		}

		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "引用不存在字段的聚合应该成功执行")
		assert.NotNil(t, cursor, "游标应该存在")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码应该成功")

		// 不存在的字段会导致所有文档分组到 null 组
		assert.Len(t, results, 1, "应该有一个null分组")
		assert.Nil(t, results[0]["_id"], "分组ID应该为null")
	})

	t.Run("EmptyPipeline", func(t *testing.T) {
		// 空的聚合管道
		pipeline := types.Pipeline{}

		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "空管道聚合应该成功")
		assert.NotNil(t, cursor, "游标应该存在")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码空管道结果应该成功")
		assert.Len(t, results, 2, "空管道应该返回所有文档")
	})

	t.Run("AggregateOnEmptyCollection", func(t *testing.T) {
		// 在空集合上执行聚合
		emptyCollectionName := "test_empty_aggregate_" + time.Now().Format("20060102150405")
		emptyCollection := testDatabase.Collection(emptyCollectionName)

		defer func() {
			if err := emptyCollection.Drop(ctx); err != nil {
				t.Logf("清理空集合失败: %v", err)
			}
		}()

		pipeline := types.Pipeline{
			{
				"$group": types.Document{
					"_id":      "$category",
					"count":    types.Document{"$sum": 1},
					"avgValue": types.Document{"$avg": "$value"},
				},
			},
		}

		cursor, err := emptyCollection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "空集合聚合应该成功")
		assert.NotNil(t, cursor, "游标应该存在")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码空集合聚合结果应该成功")
		assert.Len(t, results, 0, "空集合聚合应该返回空结果")
	})
}

// TestCollectionAggregatePerformance 聚合操作性能测试
func TestCollectionAggregatePerformance(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_aggregate_perf_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	// 准备性能测试数据
	docs := make([]types.Document, 1000)
	for i := 0; i < 1000; i++ {
		docs[i] = types.Document{
			"_id":       i,
			"userId":    i % 100,                                // 100个不同用户
			"category":  "category_" + string(rune('A'+(i%10))), // 10个不同分类
			"amount":    float64((i % 1000) + 1),                // 金额1-1000
			"timestamp": time.Now().Add(time.Duration(i%24) * time.Hour),
			"status":    []string{"pending", "completed", "cancelled"}[i%3],
			"tags":      []string{"tag1", "tag2", "tag3", "tag4", "tag5"}[:((i % 5) + 1)],
			"metadata": types.Document{
				"source":   []string{"web", "mobile", "api"}[i%3],
				"version":  (i % 5) + 1,
				"priority": []string{"low", "medium", "high"}[i%3],
			},
		}
	}

	_, err := collection.InsertMany(ctx, docs, nil)
	require.NoError(t, err, "插入性能测试数据应该成功")

	t.Run("SimpleGroupingPerformance", func(t *testing.T) {
		// 简单分组聚合性能测试
		pipeline := types.Pipeline{
			{
				"$group": types.Document{
					"_id":         "$category",
					"count":       types.Document{"$sum": 1},
					"totalAmount": types.Document{"$sum": "$amount"},
					"avgAmount":   types.Document{"$avg": "$amount"},
					"maxAmount":   types.Document{"$max": "$amount"},
					"minAmount":   types.Document{"$min": "$amount"},
				},
			},
			{
				"$sort": types.Document{"totalAmount": -1},
			},
		}

		start := time.Now()
		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "简单分组聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码简单分组结果应该成功")

		duration := time.Since(start)
		assert.Len(t, results, 10, "应该有10个分类")

		t.Logf("简单分组聚合 (1000文档, 10分类) 耗时: %v", duration)
		t.Logf("分组结果示例: %+v", results[0])
	})

	t.Run("ComplexAggregationPerformance", func(t *testing.T) {
		// 复杂聚合性能测试
		pipeline := types.Pipeline{
			{
				"$match": types.Document{
					"amount": types.Document{"$gte": 100},
					"status": types.Document{"$in": []string{"completed", "pending"}},
				},
			},
			{
				"$unwind": "$tags",
			},
			{
				"$group": types.Document{
					"_id": types.Document{
						"category": "$category",
						"tag":      "$tags",
						"source":   "$metadata.source",
					},
					"transactionCount": types.Document{"$sum": 1},
					"totalAmount":      types.Document{"$sum": "$amount"},
					"uniqueUsers":      types.Document{"$addToSet": "$userId"},
					"avgAmount":        types.Document{"$avg": "$amount"},
				},
			},
			{
				"$project": types.Document{
					"category":         "$_id.category",
					"tag":              "$_id.tag",
					"source":           "$_id.source",
					"transactionCount": 1,
					"totalAmount":      1,
					"uniqueUserCount":  types.Document{"$size": "$uniqueUsers"},
					"avgAmount":        types.Document{"$round": []interface{}{"$avgAmount", 2}},
				},
			},
			{
				"$sort": types.Document{
					"totalAmount":      -1,
					"transactionCount": -1,
				},
			},
			{
				"$limit": 20,
			},
		}

		start := time.Now()
		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "复杂聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码复杂聚合结果应该成功")

		duration := time.Since(start)
		assert.True(t, len(results) <= 20, "结果数量应该不超过20")

		t.Logf("复杂聚合 (包含match, unwind, group, project, sort, limit) 耗时: %v", duration)
		t.Logf("复杂聚合结果数量: %d", len(results))
		if len(results) > 0 {
			t.Logf("复杂聚合结果示例: %+v", results[0])
		}
	})

	t.Run("LargeResultSetPerformance", func(t *testing.T) {
		// 大结果集聚合性能测试
		pipeline := types.Pipeline{
			{
				"$group": types.Document{
					"_id": types.Document{
						"userId":   "$userId",
						"category": "$category",
						"status":   "$status",
					},
					"transactionCount": types.Document{"$sum": 1},
					"totalAmount":      types.Document{"$sum": "$amount"},
					"transactions": types.Document{"$push": types.Document{
						"id":        "$_id",
						"amount":    "$amount",
						"timestamp": "$timestamp",
					}},
				},
			},
			{
				"$match": types.Document{
					"transactionCount": types.Document{"$gte": 2},
				},
			},
			{
				"$sort": types.Document{"totalAmount": -1},
			},
		}

		start := time.Now()
		cursor, err := collection.Aggregate(ctx, pipeline, nil)
		assert.NoError(t, err, "大结果集聚合应该成功")

		var results []types.Document
		err = cursor.All(ctx, &results)
		assert.NoError(t, err, "解码大结果集聚合结果应该成功")

		duration := time.Since(start)

		t.Logf("大结果集聚合 (包含数组字段) 耗时: %v", duration)
		t.Logf("大结果集聚合结果数量: %d", len(results))

		// 验证结果结构
		if len(results) > 0 {
			firstResult := results[0]
			assert.Contains(t, firstResult, "transactionCount", "结果应该包含transactionCount")
			assert.Contains(t, firstResult, "totalAmount", "结果应该包含totalAmount")
			assert.Contains(t, firstResult, "transactions", "结果应该包含transactions数组")

			transactionCount := toInt64(firstResult["transactionCount"])
			assert.GreaterOrEqual(t, transactionCount, int64(2), "交易数量应该大于等于2")

			// 处理 MongoDB 驱动返回的数组类型
			var transactions []interface{}
			if transArray, ok := firstResult["transactions"].([]interface{}); ok {
				transactions = transArray
			} else if transPrimitive, ok := firstResult["transactions"].(primitive.A); ok {
				transactions = []interface{}(transPrimitive)
			} else {
				t.Errorf("无法解析 transactions 字段类型: %T", firstResult["transactions"])
				return
			}

			assert.Len(t, transactions, int(transactionCount), "交易数组长度应该等于交易数量")
		}
	})
}

// TestCollectionAdvancedCRUD 测试高级CRUD操作
func TestCollectionAdvancedCRUD(t *testing.T) {
	skipIfNoConnection(t)

	ctx := context.Background()
	collectionName := "test_collection_advanced_crud_" + time.Now().Format("20060102150405")
	collection := testDatabase.Collection(collectionName)

	// 确保测试后清理
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Logf("清理集合失败: %v", err)
		}
	}()

	// 准备测试数据
	testDoc := types.Document{
		"_id":     "advanced_test_1",
		"name":    "原始文档",
		"version": 1,
		"status":  "draft",
		"metadata": types.Document{
			"author":      "测试用户",
			"created":     time.Now(),
			"tags":        []string{"test", "document"},
			"permissions": []string{"read", "write"},
		},
		"content": "这是一个测试文档的内容",
	}

	_, err := collection.InsertOne(ctx, testDoc, nil)
	require.NoError(t, err, "插入高级CRUD测试数据应该成功")

	t.Run("ReplaceOne", func(t *testing.T) {
		// 替换整个文档
		replacement := types.Document{
			"_id":     "advanced_test_1",
			"name":    "替换后的文档",
			"version": 2,
			"status":  "published",
			"metadata": types.Document{
				"author":      "新作者",
				"updated":     time.Now(),
				"tags":        []string{"updated", "published"},
				"permissions": []string{"read"},
			},
			"content":   "这是替换后的文档内容",
			"published": true,
		}

		result, err := collection.ReplaceOne(ctx, types.Filter{"_id": "advanced_test_1"}, replacement, nil)
		assert.NoError(t, err, "替换文档应该成功")
		assert.NotNil(t, result, "替换结果不应为空")
		assert.Equal(t, int64(1), result.MatchedCount, "应该匹配1个文档")
		assert.Equal(t, int64(1), result.ModifiedCount, "应该修改1个文档")

		// 验证替换结果
		findResult := collection.FindOne(ctx, types.Filter{"_id": "advanced_test_1"}, nil)
		var replacedDoc types.Document
		err = findResult.Decode(&replacedDoc)
		assert.NoError(t, err, "查找替换后的文档应该成功")
		assert.Equal(t, "替换后的文档", replacedDoc["name"], "文档名称应该被替换")
		assert.Equal(t, float64(2), toFloat64(replacedDoc["version"]), "版本应该被替换")
		assert.Equal(t, true, replacedDoc["published"], "新字段应该存在")

		t.Logf("文档替换成功: %+v", result)
	})

	t.Run("FindOneAndUpdate", func(t *testing.T) {
		// 原子查找并更新
		update := types.Update{
			"$set": types.Document{
				"status":            "reviewed",
				"metadata.reviewer": "审核员",
				"reviewDate":        time.Now(),
			},
			"$inc": types.Document{
				"version": 1,
			},
		}

		// 返回更新后的文档
		returnAfter := types.ReturnDocumentAfter
		result := collection.FindOneAndUpdate(ctx,
			types.Filter{"_id": "advanced_test_1"},
			update,
			&types.FindOneAndUpdateOptions{
				ReturnDocument: &returnAfter,
			})

		assert.NotNil(t, result, "查找并更新结果不应为空")

		var updatedDoc types.Document
		err := result.Decode(&updatedDoc)
		assert.NoError(t, err, "解码查找并更新结果应该成功")
		assert.Equal(t, "reviewed", updatedDoc["status"], "状态应该被更新")
		assert.Equal(t, float64(3), toFloat64(updatedDoc["version"]), "版本应该递增")
		assert.Contains(t, updatedDoc, "reviewDate", "应该包含审核日期")

		// 验证元数据中的审核员字段
		metadata := updatedDoc["metadata"].(types.Document)
		assert.Equal(t, "审核员", metadata["reviewer"], "元数据应该包含审核员")

		t.Logf("原子更新成功，更新后的文档: %+v", updatedDoc)
	})

	t.Run("FindOneAndReplace", func(t *testing.T) {
		// 原子查找并替换
		replacement := types.Document{
			"_id":         "advanced_test_1",
			"name":        "最终版本文档",
			"status":      "final",
			"version":     10,
			"finalizedBy": "系统管理员",
			"finalizedAt": time.Now(),
		}

		// 返回替换前的文档
		returnBefore := types.ReturnDocumentBefore
		result := collection.FindOneAndReplace(ctx,
			types.Filter{"_id": "advanced_test_1"},
			replacement,
			&types.FindOneAndReplaceOptions{
				ReturnDocument: &returnBefore,
			})

		assert.NotNil(t, result, "查找并替换结果不应为空")

		var beforeDoc types.Document
		err := result.Decode(&beforeDoc)
		assert.NoError(t, err, "解码查找并替换结果应该成功")
		assert.Equal(t, "reviewed", beforeDoc["status"], "返回的应该是替换前的文档")
		assert.Equal(t, float64(3), toFloat64(beforeDoc["version"]), "返回的应该是替换前的版本")

		// 验证文档已被替换
		findResult := collection.FindOne(ctx, types.Filter{"_id": "advanced_test_1"}, nil)
		var currentDoc types.Document
		err = findResult.Decode(&currentDoc)
		assert.NoError(t, err, "查找当前文档应该成功")
		assert.Equal(t, "final", currentDoc["status"], "当前文档状态应该是final")
		assert.Equal(t, float64(10), toFloat64(currentDoc["version"]), "当前文档版本应该是10")
		assert.Equal(t, "系统管理员", currentDoc["finalizedBy"], "应该包含最终确认人")

		t.Logf("原子替换成功，替换前文档状态: %s, 当前文档状态: %s",
			beforeDoc["status"], currentDoc["status"])
	})

	t.Run("FindOneAndDelete", func(t *testing.T) {
		// 原子查找并删除
		result := collection.FindOneAndDelete(ctx, types.Filter{"_id": "advanced_test_1"}, nil)
		assert.NotNil(t, result, "查找并删除结果不应为空")

		var deletedDoc types.Document
		err := result.Decode(&deletedDoc)
		assert.NoError(t, err, "解码查找并删除结果应该成功")
		assert.Equal(t, "final", deletedDoc["status"], "返回的应该是被删除的文档")
		assert.Equal(t, "最终版本文档", deletedDoc["name"], "返回的文档名称应该正确")

		// 验证文档已被删除
		findResult := collection.FindOne(ctx, types.Filter{"_id": "advanced_test_1"}, nil)
		var notFoundDoc types.Document
		err = findResult.Decode(&notFoundDoc)
		assert.Error(t, err, "查找已删除的文档应该失败")

		t.Logf("原子删除成功，被删除的文档: %s", deletedDoc["name"])
	})
}
