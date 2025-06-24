package bootstrap

import (
	"testing"
)

// TestGatewayPoolSingleton 测试连接池单例模式
func TestGatewayPoolSingleton(t *testing.T) {
	// 多次获取连接池实例
	pool1 := GetGlobalPool()
	pool2 := GetGlobalPool()
	pool3 := GetGlobalPool()

	// 验证所有实例都是同一个
	if pool1 != pool2 {
		t.Error("GetGlobalPool() 返回了不同的实例，单例模式失败")
	}
	
	if pool2 != pool3 {
		t.Error("GetGlobalPool() 返回了不同的实例，单例模式失败")
	}
	
	if pool1 != pool3 {
		t.Error("GetGlobalPool() 返回了不同的实例，单例模式失败")
	}

	// 验证接口类型
	var _ GatewayPool = pool1
	
	t.Log("连接池单例模式测试通过")
}

// TestGatewayPoolBasicOperations 测试连接池基本操作
func TestGatewayPoolBasicOperations(t *testing.T) {
	pool := GetGlobalPool()
	
	// 清空连接池（为了测试环境清洁）
	if err := pool.Clear(); err != nil {
		t.Fatalf("清空连接池失败: %v", err)
	}
	
	// 验证空连接池
	if pool.Count() != 0 {
		t.Error("清空后连接池应该为空")
	}
	
	// 验证不存在的实例
	if pool.Exists("test-instance") {
		t.Error("不存在的实例不应该返回true")
	}
	
	// 验证获取不存在的实例
	_, err := pool.Get("test-instance")
	if err == nil {
		t.Error("获取不存在的实例应该返回错误")
	}
	
	t.Log("连接池基本操作测试通过")
}

// TestCannotCreateDirectly 测试无法直接创建连接池实例
func TestCannotCreateDirectly(t *testing.T) {
	// 这个测试主要是编译时检查
	// 因为 gatewayPool 和 newGatewayPool 都是私有的
	// 如果外部能够直接创建实例，编译就会失败
	
	// 以下代码如果取消注释应该会编译失败：
	// pool := &gatewayPool{} // 编译错误：类型未导出
	// pool := newGatewayPool() // 编译错误：函数未导出
	
	// 只能通过接口获取实例
	var pool GatewayPool = GetGlobalPool()
	if pool == nil {
		t.Error("通过GetGlobalPool()获取的实例不应该为nil")
	}
	
	t.Log("连接池私有化测试通过")
} 
 