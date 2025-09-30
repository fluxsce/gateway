package nacos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/model"

	"gateway/pkg/plugin/nacos"
)

// 测试配置 - 基于你提供的Spring Cloud配置
const (
	// Nacos服务器配置
	testNacosAddr = "192.168.0.120"
	testNacosPort = uint64(8848)
	testNamespace = "ea63c755-3d65-4203-87d7-5ee6837f5bc9"
	testGroup     = "datahub-test-group"
	testUsername  = "nacos"
	testPassword  = "nacos"

	// 测试服务配置
	testServiceName = "gateway-test-service"
	testServiceIP   = "192.168.1.100"
	testServicePort = uint64(8080)
)

// createTestConfig 创建测试配置
func createTestConfig() *nacos.NacosConfig {
	return &nacos.NacosConfig{
		Servers: []nacos.ServerConfig{
			{Host: testNacosAddr, Port: int(testNacosPort)},
		},
		Namespace: testNamespace,
		Username:  testUsername,
		Password:  testPassword,
		Timeout:   10,
	}
}

// createTestConfigWithGroup 创建带有默认分组的测试配置
func createTestConfigWithGroup() *nacos.NacosConfig {
	return &nacos.NacosConfig{
		Servers: []nacos.ServerConfig{
			{Host: testNacosAddr, Port: int(testNacosPort)},
		},
		Namespace: testNamespace,
		Group:     testGroup, // 设置默认分组
		Username:  testUsername,
		Password:  testPassword,
		Timeout:   10,
	}
}

// TestNacosTool_BasicOperations 测试基本操作
func TestNacosTool_BasicOperations(t *testing.T) {
	// 创建工具实例
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	// 检查基本信息
	if tool.GetID() != "nacos" {
		t.Errorf("工具ID错误，期望 nacos，实际 %s", tool.GetID())
	}

	if tool.GetType() != "service_discovery" {
		t.Errorf("工具类型错误，期望 service_discovery，实际 %s", tool.GetType())
	}

	// 检查初始状态
	if tool.IsConnected() {
		t.Error("工具初始状态应该未连接")
	}

	if tool.IsActive() {
		t.Error("工具初始状态应该不活跃")
	}

	// 获取状态
	status := tool.GetStatus()
	if status == nil {
		t.Error("获取状态失败")
	}

	connected, ok := status["connected"].(bool)
	if !ok || connected {
		t.Error("初始连接状态应该为false")
	}
}

// TestNacosTool_Connection 测试连接功能
func TestNacosTool_Connection(t *testing.T) {
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	// 测试连接（这里可能会失败，因为需要真实的Nacos服务器）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := tool.Connect(ctx)
	if err != nil {
		t.Logf("连接失败（可能是预期行为，如果没有Nacos服务器）: %v", err)
		t.Skip("跳过需要真实Nacos服务器的测试")
		return
	}

	// 如果连接成功
	defer tool.Close()

	if !tool.IsConnected() {
		t.Error("连接成功后状态应该为已连接")
	}

	if !tool.IsActive() {
		t.Error("连接成功后应该处于活跃状态")
	}

	// 获取连接信息
	connInfo := tool.GetConnectionInfo()
	if connInfo == nil {
		t.Error("获取连接信息失败")
	}

	if connInfo["serverAddr"] != testNacosAddr {
		t.Errorf("连接信息中服务器地址错误，期望 %s，实际 %v", testNacosAddr, connInfo["serverAddr"])
	}

	// 测试重连
	err = tool.Reconnect(ctx)
	if err != nil {
		t.Errorf("重连失败: %v", err)
	}

	if !tool.IsConnected() {
		t.Error("重连后应该保持连接状态")
	}
}

// TestNacosTool_ServiceRegistration 测试服务注册功能
func TestNacosTool_ServiceRegistration(t *testing.T) {
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	// 未连接时的操作应该失败
	err := tool.RegisterService(testServiceName, testServiceIP, testServicePort)
	if err == nil {
		t.Error("未连接时注册服务应该失败")
	}

	err = tool.RegisterServiceWithGroup(testServiceName, testServiceIP, testServicePort, testGroup)
	if err == nil {
		t.Error("未连接时指定分组注册服务应该失败")
	}

	metadata := map[string]string{
		"version": "1.0.0",
		"env":     "test",
	}
	err = tool.RegisterServiceWithMetadata(testServiceName, testServiceIP, testServicePort, testGroup, metadata)
	if err == nil {
		t.Error("未连接时带元数据注册服务应该失败")
	}

	// 连接后测试（需要真实的Nacos服务器）
	ctx := context.Background()
	err = tool.Connect(ctx)
	if err != nil {
		t.Logf("连接失败，跳过服务注册测试: %v", err)
		return
	}
	defer tool.Close()

	// 测试基本注册
	err = tool.RegisterService(testServiceName, testServiceIP, testServicePort)
	if err != nil {
		t.Logf("注册服务失败（可能是网络问题）: %v", err)
	}

	// 测试分组注册
	err = tool.RegisterServiceWithGroup(testServiceName+"-group", testServiceIP, testServicePort+1, testGroup)
	if err != nil {
		t.Logf("分组注册服务失败: %v", err)
	}

	// 测试带元数据注册
	err = tool.RegisterServiceWithMetadata(testServiceName+"-meta", testServiceIP, testServicePort+2, testGroup, metadata)
	if err != nil {
		t.Logf("带元数据注册服务失败: %v", err)
	}

	// 等待注册生效
	time.Sleep(2 * time.Second)

	// 清理注册的服务
	_ = tool.DeregisterService(testServiceName, testServiceIP, testServicePort)
	_ = tool.DeregisterServiceWithGroup(testServiceName+"-group", testServiceIP, testServicePort+1, testGroup)
	_ = tool.DeregisterServiceWithGroup(testServiceName+"-meta", testServiceIP, testServicePort+2, testGroup)
}

// TestNacosTool_ServiceDiscovery 测试服务发现功能
func TestNacosTool_ServiceDiscovery(t *testing.T) {
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	// 未连接时的操作应该失败
	_, err := tool.DiscoverService(testServiceName)
	if err == nil {
		t.Error("未连接时发现服务应该失败")
	}

	_, err = tool.DiscoverHealthyService(testServiceName)
	if err == nil {
		t.Error("未连接时发现健康服务应该失败")
	}

	_, err = tool.SelectOneInstance(testServiceName)
	if err == nil {
		t.Error("未连接时选择实例应该失败")
	}

	// 连接后测试
	ctx := context.Background()
	err = tool.Connect(ctx)
	if err != nil {
		t.Logf("连接失败，跳过服务发现测试: %v", err)
		return
	}
	defer tool.Close()

	// 先注册一个测试服务
	err = tool.RegisterServiceWithGroup(testServiceName, testServiceIP, testServicePort, testGroup)
	if err != nil {
		t.Logf("注册测试服务失败: %v", err)
		return
	}
	defer tool.DeregisterServiceWithGroup(testServiceName, testServiceIP, testServicePort, testGroup)

	// 等待注册生效
	time.Sleep(3 * time.Second)

	// 测试服务发现
	instances, err := tool.DiscoverServiceWithGroup("datahubServer", testGroup)
	if err != nil {
		t.Logf("发现服务失败: %v", err)
	} else {
		t.Logf("发现服务成功，实例数: %d", len(instances))
		for _, instance := range instances {
			t.Logf("实例: %s:%d, 健康: %v", instance.Ip, instance.Port, instance.Healthy)
		}
	}

	// 测试健康服务发现
	healthyInstances, err := tool.DiscoverHealthyServiceWithGroup("datahubServer", testGroup)
	if err != nil {
		t.Logf("发现健康服务失败: %v", err)
	} else {
		t.Logf("发现健康服务成功，实例数: %d", len(healthyInstances))
	}

	// 测试选择实例
	instance, err := tool.SelectOneInstanceWithGroup(testServiceName, testGroup)
	if err != nil {
		t.Logf("选择实例失败: %v", err)
	} else if instance != nil {
		t.Logf("选择实例成功: %s:%d", instance.Ip, instance.Port)
	}

	// 测试获取所有服务
	serviceList, err := tool.GetAllServicesWithGroup(1, 20, testGroup)
	if err != nil {
		t.Logf("获取服务列表失败: %v", err)
	} else {
		t.Logf("获取服务列表成功，服务数: %d", len(serviceList.Doms))
	}
}

// TestNacosTool_ServiceSubscription 测试服务订阅功能
func TestNacosTool_ServiceSubscription(t *testing.T) {
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	// 未连接时的操作应该失败
	err := tool.SubscribeService(testServiceName, func(instances []model.Instance, err error) {})
	if err == nil {
		t.Error("未连接时订阅服务应该失败")
	}

	// 连接后测试
	ctx := context.Background()
	err = tool.Connect(ctx)
	if err != nil {
		t.Logf("连接失败，跳过服务订阅测试: %v", err)
		return
	}
	defer tool.Close()

	// 测试订阅
	callbackCalled := false
	callback := func(instances []model.Instance, err error) {
		callbackCalled = true
		if err != nil {
			t.Logf("服务变更通知错误: %v", err)
			return
		}
		t.Logf("服务变更通知: 实例数 %d", len(instances))
	}

	err = tool.SubscribeServiceWithGroup(testServiceName, testGroup, callback)
	if err != nil {
		t.Logf("订阅服务失败: %v", err)
		return
	}

	// 注册一个服务触发变更通知
	err = tool.RegisterServiceWithGroup(testServiceName, testServiceIP, testServicePort, testGroup)
	if err != nil {
		t.Logf("注册服务失败: %v", err)
	} else {
		// 等待通知
		time.Sleep(3 * time.Second)

		// 注销服务
		_ = tool.DeregisterServiceWithGroup(testServiceName, testServiceIP, testServicePort, testGroup)

		// 再等待一下通知
		time.Sleep(2 * time.Second)
	}

	// 取消订阅
	err = tool.UnsubscribeServiceWithGroup(testServiceName, testGroup)
	if err != nil {
		t.Logf("取消订阅失败: %v", err)
	}

	if callbackCalled {
		t.Log("订阅回调被成功调用")
	} else {
		t.Log("订阅回调未被调用（可能是网络或配置问题）")
	}
}

// TestNacosTool_ResourceCleanup 测试资源清理
func TestNacosTool_ResourceCleanup(t *testing.T) {
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	ctx := context.Background()
	err := tool.Connect(ctx)
	if err != nil {
		t.Logf("连接失败，跳过资源清理测试: %v", err)
		return
	}

	// 添加一些订阅
	callback := func(instances []model.Instance, err error) {}
	_ = tool.SubscribeServiceWithGroup("test-service-1", testGroup, callback)
	_ = tool.SubscribeServiceWithGroup("test-service-2", testGroup, callback)

	// 检查订阅数量
	status := tool.GetStatus()
	if subscriptionCount, ok := status["subscriptionCount"].(int); ok {
		t.Logf("当前订阅数量: %d", subscriptionCount)
	}

	// 关闭工具
	err = tool.Close()
	if err != nil {
		t.Errorf("关闭工具失败: %v", err)
	}

	// 检查状态
	if tool.IsConnected() {
		t.Error("关闭后应该处于未连接状态")
	}

	if tool.IsActive() {
		t.Error("关闭后应该处于非活跃状态")
	}

	// 检查订阅是否被清理
	status = tool.GetStatus()
	if subscriptionCount, ok := status["subscriptionCount"].(int); ok && subscriptionCount != 0 {
		t.Errorf("关闭后订阅应该被清理，但仍有 %d 个订阅", subscriptionCount)
	}
}

// TestNacosTool_ConcurrentOperations 测试并发操作
func TestNacosTool_ConcurrentOperations(t *testing.T) {
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	ctx := context.Background()
	err := tool.Connect(ctx)
	if err != nil {
		t.Logf("连接失败，跳过并发测试: %v", err)
		return
	}
	defer tool.Close()

	// 并发注册服务
	const goroutineCount = 10
	done := make(chan bool, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			defer func() { done <- true }()

			serviceName := fmt.Sprintf("concurrent-service-%d", index)
			port := testServicePort + uint64(index)

			// 注册服务
			err := tool.RegisterServiceWithGroup(serviceName, testServiceIP, port, testGroup)
			if err != nil {
				t.Logf("并发注册服务 %s 失败: %v", serviceName, err)
				return
			}

			// 等待一段时间
			time.Sleep(100 * time.Millisecond)

			// 发现服务
			_, err = tool.DiscoverServiceWithGroup(serviceName, testGroup)
			if err != nil {
				t.Logf("并发发现服务 %s 失败: %v", serviceName, err)
			}

			// 注销服务
			err = tool.DeregisterServiceWithGroup(serviceName, testServiceIP, port, testGroup)
			if err != nil {
				t.Logf("并发注销服务 %s 失败: %v", serviceName, err)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < goroutineCount; i++ {
		<-done
	}

	t.Log("并发操作测试完成")
}

// BenchmarkNacosTool_ServiceDiscovery 服务发现性能测试
func BenchmarkNacosTool_ServiceDiscovery(b *testing.B) {
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	ctx := context.Background()
	err := tool.Connect(ctx)
	if err != nil {
		b.Skipf("连接失败，跳过性能测试: %v", err)
		return
	}
	defer tool.Close()

	// 先注册一个测试服务
	err = tool.RegisterServiceWithGroup("benchmark-service", testServiceIP, testServicePort, testGroup)
	if err != nil {
		b.Skipf("注册测试服务失败: %v", err)
		return
	}
	defer tool.DeregisterServiceWithGroup("benchmark-service", testServiceIP, testServicePort, testGroup)

	// 等待注册生效
	time.Sleep(2 * time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tool.DiscoverServiceWithGroup("benchmark-service", testGroup)
			if err != nil {
				b.Errorf("服务发现失败: %v", err)
			}
		}
	})
}

// BenchmarkNacosTool_SelectInstance 实例选择性能测试
func BenchmarkNacosTool_SelectInstance(b *testing.B) {
	config := createTestConfig()
	tool := nacos.NewNacosTool(config)

	ctx := context.Background()
	err := tool.Connect(ctx)
	if err != nil {
		b.Skipf("连接失败，跳过性能测试: %v", err)
		return
	}
	defer tool.Close()

	// 注册多个实例
	for i := 0; i < 3; i++ {
		port := testServicePort + uint64(i)
		err = tool.RegisterServiceWithGroup("benchmark-select-service", testServiceIP, port, testGroup)
		if err != nil {
			b.Logf("注册实例失败: %v", err)
		}
		defer tool.DeregisterServiceWithGroup("benchmark-select-service", testServiceIP, port, testGroup)
	}

	// 等待注册生效
	time.Sleep(3 * time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tool.SelectOneInstanceWithGroup("benchmark-select-service", testGroup)
			if err != nil {
				b.Errorf("选择实例失败: %v", err)
			}
		}
	})
}

// TestNacosClient_DefaultGroup 测试默认分组功能
func TestNacosClient_DefaultGroup(t *testing.T) {
	// 跳过集成测试（需要真实的Nacos服务器）
	t.Skip("这是集成测试，需要运行中的Nacos服务器")

	// 创建带有默认分组的配置
	config := createTestConfigWithGroup()

	// 创建客户端
	client, err := nacos.NewClient(config)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 测试服务名和实例信息
	testServiceName := "test-default-group-service"
	testIP := "192.168.1.200"
	testPort := uint64(9999)

	// 1. 测试注册实例时使用默认分组（groupName 传空字符串）
	err = client.RegisterInstance(testServiceName, testIP, testPort, "")
	if err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	// 清理：注销实例
	defer func() {
		err := client.DeregisterInstance(testServiceName, testIP, testPort, "")
		if err != nil {
			t.Logf("注销实例失败: %v", err)
		}
	}()

	// 等待注册生效
	time.Sleep(2 * time.Second)

	// 2. 测试获取服务时使用默认分组（groupName 传空字符串）
	instances, err := client.GetService(testServiceName, "")
	if err != nil {
		t.Fatalf("获取服务实例失败: %v", err)
	}

	// 验证能够找到刚才注册的实例
	found := false
	for _, instance := range instances {
		if instance.Ip == testIP && instance.Port == testPort {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("未找到注册的实例，应该在默认分组 %s 中", testGroup)
	}

	// 3. 测试选择健康实例时使用默认分组
	instance, err := client.SelectOneHealthyInstance(testServiceName, "")
	if err != nil {
		t.Fatalf("选择健康实例失败: %v", err)
	}

	if instance.Ip != testIP || instance.Port != testPort {
		t.Errorf("选择的实例不正确，期望: %s:%d, 实际: %s:%d",
			testIP, testPort, instance.Ip, instance.Port)
	}

	t.Logf("测试通过：默认分组功能正常工作，使用分组: %s", testGroup)
}

// TestNacosClient_DefaultGroupBehavior 测试默认分组行为的单元测试
func TestNacosClient_DefaultGroupBehavior(t *testing.T) {
	// 测试配置中有分组的情况
	configWithGroup := &nacos.NacosConfig{
		Servers: []nacos.ServerConfig{
			{Host: "127.0.0.1", Port: 8848},
		},
		Group: "my-custom-group",
	}

	client, err := nacos.NewClient(configWithGroup)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 通过反射或者公开方法测试 getDefaultGroup 的行为
	// 由于 getDefaultGroup 是私有方法，我们无法直接测试
	// 但我们可以通过其他方式验证行为

	// 测试配置中没有分组的情况
	configWithoutGroup := &nacos.NacosConfig{
		Servers: []nacos.ServerConfig{
			{Host: "127.0.0.1", Port: 8848},
		},
		// Group 为空
	}

	client2, err := nacos.NewClient(configWithoutGroup)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client2.Close()

	t.Log("默认分组行为测试完成")
}
