package proxy

import (
	"net/http/httptest"
	"testing"
	"time"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/proxy"
	"gohub/internal/gateway/handler/service"
)

// MockServiceManager 简单的mock服务管理器
type MockServiceManager struct {
	nodes map[string]*service.NodeConfig
	err   error
}

func (m *MockServiceManager) SelectNode(serviceID string, ctx *core.Context) (*service.NodeConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	if node, exists := m.nodes[serviceID]; exists {
		return node, nil
	}
	return &service.NodeConfig{
		ID:      "test-node",
		URL:     "http://localhost:8081",
		Weight:  1,
		Health:  true,
		Enabled: true,
	}, nil
}

// 实现ServiceManager接口的其他方法（测试用简单实现）
func (m *MockServiceManager) AddService(config *service.ServiceConfig) error { return nil }
func (m *MockServiceManager) RemoveService(serviceID string) error { return nil }
func (m *MockServiceManager) GetService(serviceID string) (*service.ServiceConfig, bool) {
	return nil, false
}
func (m *MockServiceManager) ListServices() []*service.ServiceConfig { return nil }
func (m *MockServiceManager) AddNode(serviceID string, node *service.NodeConfig) error { return nil }
func (m *MockServiceManager) RemoveNode(serviceID, nodeID string) error { return nil }
func (m *MockServiceManager) UpdateNodeHealth(serviceID, nodeID string, healthy bool) error {
	return nil
}
func (m *MockServiceManager) UpdateNodeStatus(serviceID, nodeID string, enabled bool) error {
	return nil
}
func (m *MockServiceManager) GetHealthyNodes(serviceID string) ([]*service.NodeConfig, error) {
	return nil, nil
}
func (m *MockServiceManager) GetUnhealthyNodes(serviceID string) ([]*service.NodeConfig, error) {
	return nil, nil
}
func (m *MockServiceManager) GetAllNodes(serviceID string) ([]*service.NodeConfig, error) {
	return nil, nil
}
func (m *MockServiceManager) UpdateService(service *service.ServiceConfig) error { return nil }
func (m *MockServiceManager) GetServiceStats(serviceID string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockServiceManager) RecordServiceSuccess(serviceID string, responseTime time.Duration) {}
func (m *MockServiceManager) RecordServiceFailure(serviceID string)                             {}
func (m *MockServiceManager) Close() error                                                     { return nil }

// TestWebSocketUpgradeHandler_IsWebSocketUpgrade 测试WebSocket升级检测
func TestWebSocketUpgradeHandler_IsWebSocketUpgrade(t *testing.T) {
	mockServiceManager := &MockServiceManager{
		nodes: make(map[string]*service.NodeConfig),
	}

	handler := proxy.NewWebSocketUpgradeHandler(mockServiceManager, nil)

	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{
			name: "Valid WebSocket Upgrade Request",
			headers: map[string]string{
				"Connection":            "upgrade",
				"Upgrade":               "websocket",
				"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
				"Sec-WebSocket-Version": "13",
			},
			expected: true,
		},
		{
			name: "Missing Connection Header",
			headers: map[string]string{
				"Upgrade":               "websocket",
				"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
				"Sec-WebSocket-Version": "13",
			},
			expected: false,
		},
		{
			name: "Wrong Connection Header",
			headers: map[string]string{
				"Connection":            "keep-alive",
				"Upgrade":               "websocket",
				"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
				"Sec-WebSocket-Version": "13",
			},
			expected: false,
		},
		{
			name: "Missing Upgrade Header",
			headers: map[string]string{
				"Connection":            "upgrade",
				"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
				"Sec-WebSocket-Version": "13",
			},
			expected: false,
		},
		{
			name: "Wrong Upgrade Header",
			headers: map[string]string{
				"Connection":            "upgrade",
				"Upgrade":               "h2c",
				"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
				"Sec-WebSocket-Version": "13",
			},
			expected: false,
		},
		{
			name: "Missing WebSocket Key",
			headers: map[string]string{
				"Connection":            "upgrade",
				"Upgrade":               "websocket",
				"Sec-WebSocket-Version": "13",
			},
			expected: false,
		},
		{
			name: "Wrong WebSocket Version",
			headers: map[string]string{
				"Connection":            "upgrade",
				"Upgrade":               "websocket",
				"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
				"Sec-WebSocket-Version": "8",
			},
			expected: false,
		},
		{
			name: "Case Insensitive Headers",
			headers: map[string]string{
				"CONNECTION":            "UPGRADE",
				"UPGRADE":               "WEBSOCKET",
				"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
				"Sec-WebSocket-Version": "13",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)

			// 设置头部
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := handler.IsWebSocketUpgrade(req)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestHTTPProxy_Integration 测试HTTP代理集成
func TestHTTPProxy_Integration(t *testing.T) {
	config := proxy.ProxyConfig{
		Type:    "http",
		Enabled: true,
		Name:    "test-http-proxy",
	}

	mockServiceManager := &MockServiceManager{
		nodes: make(map[string]*service.NodeConfig),
	}
	httpProxy, err := proxy.NewHTTPProxy(config, mockServiceManager)
	if err != nil {
		t.Fatalf("Failed to create HTTP proxy: %v", err)
	}

	// 测试HTTP代理基本功能
	if httpProxy == nil {
		t.Fatal("HTTP proxy should not be nil")
	}

	// 测试配置验证
	if err := httpProxy.Validate(); err != nil {
		t.Fatalf("HTTP proxy validation failed: %v", err)
	}

	// 测试代理类型
	if httpProxy.GetType() != "http" {
		t.Errorf("Expected proxy type 'http', got '%s'", httpProxy.GetType())
	}

	// 测试代理名称
	if httpProxy.GetName() != "test-http-proxy" {
		t.Errorf("Expected proxy name 'test-http-proxy', got '%s'", httpProxy.GetName())
	}

	// 测试是否启用
	if !httpProxy.IsEnabled() {
		t.Error("HTTP proxy should be enabled")
	}
} 