package core

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gateway/internal/gateway/core"
)

func TestNewContext(t *testing.T) {
	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()

	// 创建上下文
	ctx := core.NewContext(writer, req)

	// 验证基本属性
	assert.NotNil(t, ctx, "上下文不应该为nil")
	assert.Equal(t, req, ctx.Request, "请求对象应该匹配")
	assert.Equal(t, writer, ctx.Writer, "响应写入器应该匹配")
	assert.NotNil(t, ctx.Ctx, "上下文对象不应该为nil")
	assert.NotNil(t, ctx.Cancel, "取消函数不应该为nil")
	assert.False(t, ctx.IsResponded(), "初始状态不应该已响应")
	assert.Empty(t, ctx.GetErrors(), "初始错误列表应该为空")
}

func TestContextSetAndGet(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 测试设置和获取基本值
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)
	ctx.Set("key3", true)

	// 验证Get方法
	value1, exists1 := ctx.Get("key1")
	assert.True(t, exists1, "key1 应该存在")
	assert.Equal(t, "value1", value1, "value1 应该匹配")

	value2, exists2 := ctx.Get("key2")
	assert.True(t, exists2, "key2 应该存在")
	assert.Equal(t, 42, value2, "value2 应该匹配")

	value3, exists3 := ctx.Get("key3")
	assert.True(t, exists3, "key3 应该存在")
	assert.Equal(t, true, value3, "value3 应该匹配")

	// 测试不存在的键
	value4, exists4 := ctx.Get("nonexistent")
	assert.False(t, exists4, "不存在的键应该返回false")
	assert.Nil(t, value4, "不存在的键值应该为nil")
}

func TestContextMustGet(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 设置值
	ctx.Set("existing_key", "test_value")

	// 测试MustGet存在的键
	value, err := ctx.MustGet("existing_key")
	assert.NoError(t, err, "MustGet不应该返回错误")
	assert.Equal(t, "test_value", value, "MustGet应该返回正确的值")

	// 测试MustGet不存在的键（应该panic）
	assert.Panics(t, func() {
		ctx.MustGet("nonexistent_key")
	}, "MustGet不存在的键应该panic")
}

func TestContextTypedGetters(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 设置不同类型的值
	ctx.Set("string_key", "string_value")
	ctx.Set("int_key", 123)
	ctx.Set("bool_key", true)
	ctx.Set("wrong_type", 456) // 故意设置错误类型

	// 测试GetString
	strValue, strOK := ctx.GetString("string_key")
	assert.True(t, strOK, "GetString应该成功")
	assert.Equal(t, "string_value", strValue, "字符串值应该匹配")

	strValue2, strOK2 := ctx.GetString("wrong_type")
	assert.False(t, strOK2, "GetString错误类型应该失败")
	assert.Empty(t, strValue2, "失败时应该返回空字符串")

	// 测试GetInt
	intValue, intOK := ctx.GetInt("int_key")
	assert.True(t, intOK, "GetInt应该成功")
	assert.Equal(t, 123, intValue, "整数值应该匹配")

	intValue2, intOK2 := ctx.GetInt("string_key")
	assert.False(t, intOK2, "GetInt错误类型应该失败")
	assert.Equal(t, 0, intValue2, "失败时应该返回0")

	// 测试GetBool
	boolValue, boolOK := ctx.GetBool("bool_key")
	assert.True(t, boolOK, "GetBool应该成功")
	assert.True(t, boolValue, "布尔值应该匹配")

	boolValue2, boolOK2 := ctx.GetBool("int_key")
	assert.False(t, boolOK2, "GetBool错误类型应该失败")
	assert.False(t, boolValue2, "失败时应该返回false")
}

func TestContextRouteAndService(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 测试目标URL
	testURL := "http://backend-service:8080/api/users"
	ctx.SetTargetURL(testURL)
	assert.Equal(t, testURL, ctx.GetTargetURL(), "目标URL应该匹配")

	// 测试路由ID
	routeID := "users-route"
	ctx.SetRouteID(routeID)
	assert.Equal(t, routeID, ctx.GetRouteID(), "路由ID应该匹配")

	// 测试服务ID
	serviceID := "user-service"
	ctx.SetServiceIDs([]string{serviceID})
	serviceIDs := ctx.GetServiceIDs()
	assert.Len(t, serviceIDs, 1, "服务ID数组应该包含1个元素")
	assert.Equal(t, serviceID, serviceIDs[0], "服务ID应该匹配")

	// 测试匹配路径
	matchedPath := "/api/users/**"
	ctx.SetMatchedPath(matchedPath)
	assert.Equal(t, matchedPath, ctx.GetMatchedPath(), "匹配路径应该匹配")
}

func TestContextErrors(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 初始状态
	assert.False(t, ctx.HasErrors(), "初始状态不应该有错误")
	assert.Empty(t, ctx.GetErrors(), "初始错误列表应该为空")
	assert.Nil(t, ctx.GetLatestError(), "初始最新错误应该为nil")

	// 添加错误
	err1 := assert.AnError
	err2 := context.DeadlineExceeded

	ctx.AddError(err1)
	assert.True(t, ctx.HasErrors(), "添加错误后应该有错误")
	assert.Len(t, ctx.GetErrors(), 1, "应该有1个错误")
	assert.Equal(t, err1, ctx.GetLatestError(), "最新错误应该是err1")

	ctx.AddError(err2)
	assert.True(t, ctx.HasErrors(), "添加第二个错误后应该有错误")
	assert.Len(t, ctx.GetErrors(), 2, "应该有2个错误")
	assert.Equal(t, err2, ctx.GetLatestError(), "最新错误应该是err2")

	// 验证错误顺序
	errors := ctx.GetErrors()
	assert.Equal(t, err1, errors[0], "第一个错误应该是err1")
	assert.Equal(t, err2, errors[1], "第二个错误应该是err2")
}

func TestContextElapsed(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	// 检查经过时间
	elapsed := ctx.Elapsed()
	assert.Greater(t, elapsed, time.Duration(0), "经过时间应该大于0")
	assert.Less(t, elapsed, time.Second, "经过时间应该少于1秒")
}

func TestContextJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 测试数据
	data := map[string]interface{}{
		"message": "success",
		"code":    200,
		"data":    []string{"item1", "item2"},
	}

	// 发送JSON响应
	ctx.JSON(200, data)

	// 验证响应
	assert.True(t, ctx.IsResponded(), "应该标记为已响应")
	assert.Equal(t, 200, writer.Code, "状态码应该是200")
	assert.Contains(t, writer.Header().Get("Content-Type"), "application/json", "Content-Type应该是JSON")
	assert.NotEmpty(t, writer.Body.String(), "响应体不应该为空")
}

func TestContextString(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 发送字符串响应
	ctx.String(200, "Hello %s", "World")

	// 验证响应
	assert.True(t, ctx.IsResponded(), "应该标记为已响应")
	assert.Equal(t, 200, writer.Code, "状态码应该是200")
	assert.Equal(t, "Hello World", writer.Body.String(), "响应体应该匹配")
}

func TestContextAbort(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 中止请求
	abortData := map[string]string{"error": "unauthorized"}
	ctx.Abort(401, abortData)

	// 验证响应
	assert.True(t, ctx.IsResponded(), "应该标记为已响应")
	assert.Equal(t, 401, writer.Code, "状态码应该是401")
	assert.Contains(t, writer.Header().Get("Content-Type"), "application/json", "Content-Type应该是JSON")
	assert.Contains(t, writer.Body.String(), "unauthorized", "响应体应该包含错误信息")
}

func TestContextPathParams(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 设置路径参数
	params := map[string]string{
		"id":   "123",
		"name": "test",
	}
	ctx.SetPathParams(params)

	// 获取路径参数
	resultParams := ctx.GetPathParams()
	assert.Equal(t, params, resultParams, "路径参数应该匹配")
	assert.Equal(t, "123", resultParams["id"], "id参数应该匹配")
	assert.Equal(t, "test", resultParams["name"], "name参数应该匹配")
}

func TestContextReset(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 设置一些数据
	ctx.Set("key", "value")
	ctx.SetRouteID("test-route")
	ctx.SetServiceIDs([]string{"test-service"})
	ctx.AddError(assert.AnError)

	// 验证数据已设置
	assert.True(t, ctx.HasErrors())
	_, exists := ctx.Get("key")
	assert.True(t, exists)

	// 重置上下文
	ctx.Reset()

	// 验证数据已清空
	assert.False(t, ctx.HasErrors())
	assert.Empty(t, ctx.GetErrors())
	assert.Empty(t, ctx.GetRouteID())
	assert.Empty(t, ctx.GetServiceIDs())
	assert.Empty(t, ctx.GetTargetURL())
	assert.Empty(t, ctx.GetMatchedPath())
	assert.False(t, ctx.IsResponded())

	_, exists = ctx.Get("key")
	assert.False(t, exists, "重置后数据应该被清空")
}

func TestContextCancel(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 验证上下文没有被取消
	select {
	case <-ctx.Ctx.Done():
		t.Fatal("上下文不应该被取消")
	default:
		// 正常情况
	}

	// 取消上下文
	ctx.Cancel()

	// 验证上下文已被取消
	select {
	case <-ctx.Ctx.Done():
		// 正常情况，上下文已被取消
	case <-time.After(100 * time.Millisecond):
		t.Fatal("上下文应该被取消")
	}
}

func TestContextConcurrency(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 并发设置和获取数据
	const numGoroutines = 100
	const numOperations = 10

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)

				// 设置值
				ctx.Set(key, value)

				// 立即获取值
				if gotValue, exists := ctx.Get(key); exists {
					assert.Equal(t, value, gotValue, "并发获取的值应该匹配")
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// 基准测试
func BenchmarkContextSetGet(b *testing.B) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000) // 限制键的数量避免内存无限增长
		ctx.Set(key, i)
		ctx.Get(key)
	}
}

func BenchmarkContextJSON(b *testing.B) {
	req := httptest.NewRequest("GET", "/test", nil)

	data := map[string]interface{}{
		"message": "benchmark test",
		"code":    200,
		"data":    []int{1, 2, 3, 4, 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)
		ctx.JSON(200, data)
	}
}

func BenchmarkContextElapsed(b *testing.B) {
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Elapsed()
	}
}
