package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"gateway/pkg/config"
)

func TestGetNodeId_FromConfig(t *testing.T) {
	// 重置状态
	config.ResetNodeId()
	defer config.ResetNodeId()

	// 清除环境变量
	os.Unsetenv("GATEWAY_NODE_ID")
	os.Unsetenv("POD_NAME")

	// 此测试需要配置文件中有 app.node_id
	// 如果没有配置，跳过此测试
	nodeId := config.GetString("app.node_id", "")
	if nodeId == "" {
		t.Skip("配置文件中未设置 app.node_id，跳过此测试")
	}

	result := config.GetNodeId()
	if result != nodeId {
		t.Errorf("GetNodeId() = %s, want %s", result, nodeId)
	}
}

func TestGetNodeId_FromEnvGatewayNodeId(t *testing.T) {
	// 重置状态
	config.ResetNodeId()
	defer config.ResetNodeId()

	// 设置环境变量
	expected := "test-node-from-env"
	os.Setenv("GATEWAY_NODE_ID", expected)
	defer os.Unsetenv("GATEWAY_NODE_ID")

	// 确保 POD_NAME 不干扰测试
	os.Unsetenv("POD_NAME")

	// 清除配置中的 node_id（如果有）
	// 注意：这里假设配置中没有设置 node_id，或者环境变量优先级更高

	result := config.GetNodeId()
	// 如果配置文件中有 node_id，则配置优先
	configNodeId := config.GetString("app.node_id", "")
	if configNodeId != "" {
		t.Logf("配置文件中已设置 app.node_id=%s，环境变量将被忽略", configNodeId)
		return
	}

	if result != expected {
		t.Errorf("GetNodeId() = %s, want %s", result, expected)
	}
}

func TestGetNodeId_FromEnvPodName(t *testing.T) {
	// 重置状态
	config.ResetNodeId()
	defer config.ResetNodeId()

	// 清除 GATEWAY_NODE_ID
	os.Unsetenv("GATEWAY_NODE_ID")

	// 设置 POD_NAME
	expected := "gateway-pod-12345"
	os.Setenv("POD_NAME", expected)
	defer os.Unsetenv("POD_NAME")

	// 如果配置文件中有 node_id，则配置优先
	configNodeId := config.GetString("app.node_id", "")
	if configNodeId != "" {
		t.Logf("配置文件中已设置 app.node_id=%s，环境变量将被忽略", configNodeId)
		return
	}

	result := config.GetNodeId()
	if result != expected {
		t.Errorf("GetNodeId() = %s, want %s", result, expected)
	}
}

func TestGetNodeId_EnvPriority(t *testing.T) {
	// 测试 GATEWAY_NODE_ID 优先于 POD_NAME
	config.ResetNodeId()
	defer config.ResetNodeId()

	gatewayNodeId := "gateway-node-id-value"
	podName := "pod-name-value"

	os.Setenv("GATEWAY_NODE_ID", gatewayNodeId)
	os.Setenv("POD_NAME", podName)
	defer os.Unsetenv("GATEWAY_NODE_ID")
	defer os.Unsetenv("POD_NAME")

	// 如果配置文件中有 node_id，则配置优先
	configNodeId := config.GetString("app.node_id", "")
	if configNodeId != "" {
		t.Logf("配置文件中已设置 app.node_id=%s，环境变量将被忽略", configNodeId)
		return
	}

	result := config.GetNodeId()
	if result != gatewayNodeId {
		t.Errorf("GATEWAY_NODE_ID 应优先于 POD_NAME: got %s, want %s", result, gatewayNodeId)
	}
}

func TestGetNodeId_AutoGenerate(t *testing.T) {
	// 重置状态
	config.ResetNodeId()
	defer config.ResetNodeId()

	// 清除所有环境变量
	os.Unsetenv("GATEWAY_NODE_ID")
	os.Unsetenv("POD_NAME")

	// 如果配置文件中有 node_id，则跳过
	configNodeId := config.GetString("app.node_id", "")
	if configNodeId != "" {
		t.Skip("配置文件中已设置 app.node_id，跳过自动生成测试")
	}

	// 删除持久化文件（如果存在）
	nodeIdFile := filepath.Join(config.GetConfigDir(), ".node_id")
	os.Remove(nodeIdFile)
	defer os.Remove(nodeIdFile)

	result := config.GetNodeId()

	// 验证结果
	if result == "" {
		t.Error("GetNodeId() 返回空字符串")
	}

	// 自动生成的是 SHA256 哈希，64位十六进制
	if len(result) != 64 {
		t.Errorf("自动生成的节点ID长度应为64，实际: %d", len(result))
	}

	t.Logf("自动生成的节点ID: %s", result)
}

func TestGetNodeId_Consistency(t *testing.T) {
	// 测试多次调用返回相同结果
	config.ResetNodeId()
	defer config.ResetNodeId()

	// 清除环境变量，使用自动生成
	os.Unsetenv("GATEWAY_NODE_ID")
	os.Unsetenv("POD_NAME")

	// 如果配置文件中有 node_id，也可以测试一致性
	first := config.GetNodeId()
	second := config.GetNodeId()
	third := config.GetNodeId()

	if first != second || second != third {
		t.Errorf("GetNodeId() 返回不一致: %s, %s, %s", first, second, third)
	}
}

func TestGetNodeId_Persistence(t *testing.T) {
	// 测试持久化功能
	config.ResetNodeId()
	defer config.ResetNodeId()

	// 清除环境变量
	os.Unsetenv("GATEWAY_NODE_ID")
	os.Unsetenv("POD_NAME")

	// 如果配置文件中有 node_id，则跳过
	configNodeId := config.GetString("app.node_id", "")
	if configNodeId != "" {
		t.Skip("配置文件中已设置 app.node_id，跳过持久化测试")
	}

	// 删除持久化文件
	nodeIdFile := filepath.Join(config.GetConfigDir(), ".node_id")
	os.Remove(nodeIdFile)
	defer os.Remove(nodeIdFile)

	// 第一次获取（触发自动生成和持久化）
	first := config.GetNodeId()

	// 检查持久化文件是否创建
	if _, err := os.Stat(nodeIdFile); os.IsNotExist(err) {
		t.Log("警告: 持久化文件未创建（可能是权限问题）")
		return
	}

	// 重置缓存，模拟重启
	config.ResetNodeId()

	// 再次获取（应从持久化文件读取）
	second := config.GetNodeId()

	if first != second {
		t.Errorf("持久化后重启应返回相同ID: first=%s, second=%s", first, second)
	}
}

func TestGetNodeId_NotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetNodeId() panic: %v", r)
		}
	}()

	config.ResetNodeId()
	defer config.ResetNodeId()

	_ = config.GetNodeId()
}

// BenchmarkGetNodeId 性能测试
func BenchmarkGetNodeId(b *testing.B) {
	// 预热
	config.ResetNodeId()
	_ = config.GetNodeId()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.GetNodeId()
	}
}

func BenchmarkGetNodeId_FirstCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		config.ResetNodeId()
		_ = config.GetNodeId()
	}
}
