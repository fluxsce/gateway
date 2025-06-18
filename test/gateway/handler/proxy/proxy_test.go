package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/proxy"
	"gohub/internal/gateway/handler/service"
)

func TestProxyConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *proxy.ProxyConfig
		description string
	}{
		{
			name: "HTTPProxyConfig",
			config: &proxy.ProxyConfig{
				ID:      "test-http-proxy",
				Enabled: true,
				Name:    "HTTP代理测试",
				Type:    proxy.ProxyTypeHTTP,
				Config: map[string]interface{}{
					"timeout": "30s",
				},
			},
			description: "HTTP代理配置",
		},
		{
			name: "WebSocketProxyConfig",
			config: &proxy.ProxyConfig{
				ID:      "test-ws-proxy",
				Enabled: true,
				Name:    "WebSocket代理测试",
				Type:    proxy.ProxyTypeWebSocket,
				Config: map[string]interface{}{
					"ping_interval": "30s",
				},
			},
			description: "WebSocket代理配置",
		},
		{
			name: "TCPProxyConfig",
			config: &proxy.ProxyConfig{
				ID:      "test-tcp-proxy",
				Enabled: true,
				Name:    "TCP代理测试",
				Type:    proxy.ProxyTypeTCP,
				Config: map[string]interface{}{
					"buffer_size": 32768,
				},
			},
			description: "TCP代理配置",
		},
		{
			name: "UDPProxyConfig",
			config: &proxy.ProxyConfig{
				ID:      "test-udp-proxy",
				Enabled: true,
				Name:    "UDP代理测试",
				Type:    proxy.ProxyTypeUDP,
				Config: map[string]interface{}{
					"buffer_size": 32768,
				},
			},
			description: "UDP代理配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段
			assert.NotEmpty(t, tt.config.ID, "ID不应该为空")
			assert.NotEmpty(t, tt.config.Name, "名称不应该为空")
			assert.True(t, tt.config.Enabled, "应该启用代理")
			assert.NotEmpty(t, string(tt.config.Type), "代理类型不应该为空")
		})
	}
}

func TestProxyTypes(t *testing.T) {
	types := []proxy.ProxyType{
		proxy.ProxyTypeHTTP,
		proxy.ProxyTypeWebSocket,
		proxy.ProxyTypeTCP,
		proxy.ProxyTypeUDP,
	}

	for _, proxyType := range types {
		t.Run(string(proxyType), func(t *testing.T) {
			// 验证代理类型常量不为空
			assert.NotEmpty(t, string(proxyType), "代理类型常量不应该为空")
		})
	}
}

func TestHTTPProxy(t *testing.T) {
	// 创建后端测试服务器
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回请求信息
		response := fmt.Sprintf("Method: %s, Path: %s, Query: %s",
			r.Method, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer backend.Close()

	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	// 添加服务
	svc := &service.ServiceConfig{
		ID:       "test-service",
		Name:     "测试服务",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{
			{
				ID:      "backend1",
				URL:     backend.URL,
				Weight:  1,
				Health:  true,
				Enabled: true,
			},
		},
	}
	err := serviceManager.AddService(svc)
	require.NoError(t, err)

	config := proxy.ProxyConfig{
		ID:      "test-http",
		Enabled: true,
		Name:    "HTTP代理测试",
		Type:    proxy.ProxyTypeHTTP,
		Config: map[string]interface{}{
			"timeout": "30s",
		},
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	proxyHandler, err := proxyFactory.CreateProxy(config)
	require.NoError(t, err, "创建代理处理器失败")
	require.NotNil(t, proxyHandler, "代理处理器不应该为nil")

	tests := []struct {
		name        string
		method      string
		path        string
		query       string
		body        string
		expectCode  int
		description string
	}{
		{
			name:        "GET请求",
			method:      "GET",
			path:        "/api/users",
			query:       "page=1&limit=10",
			expectCode:  http.StatusOK,
			description: "GET请求应该被正确代理",
		},
		{
			name:        "POST请求",
			method:      "POST",
			path:        "/api/users",
			body:        `{"name":"test","email":"test@example.com"}`,
			expectCode:  http.StatusOK,
			description: "POST请求应该被正确代理",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			fullPath := tt.path
			if tt.query != "" {
				fullPath += "?" + tt.query
			}

			req := httptest.NewRequest(tt.method, fullPath, body)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)
			ctx.SetServiceID("test-service") // 设置服务ID

			// 执行代理
			result := proxyHandler.Handle(ctx)

			// 验证结果
			assert.True(t, result, tt.description)
		})
	}
}

func TestWebSocketProxy(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	// WebSocket代理测试比较复杂，这里进行基础配置测试
	config := proxy.ProxyConfig{
		ID:      "test-websocket",
		Enabled: true,
		Name:    "WebSocket代理测试",
		Type:    proxy.ProxyTypeWebSocket,
		Config: map[string]interface{}{
			"ping_interval": "30s",
		},
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	proxyHandler, err := proxyFactory.CreateProxy(config)
	require.NoError(t, err)
	require.NotNil(t, proxyHandler)

	// 基础验证WebSocket代理处理器已创建
	assert.Equal(t, proxy.ProxyTypeWebSocket, proxyHandler.GetType())
}

func TestTCPProxy(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	config := proxy.ProxyConfig{
		ID:      "test-tcp",
		Enabled: true,
		Name:    "TCP代理测试",
		Type:    proxy.ProxyTypeTCP,
		Config: map[string]interface{}{
			"buffer_size": 32768,
		},
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	proxyHandler, err := proxyFactory.CreateProxy(config)
	require.NoError(t, err)
	require.NotNil(t, proxyHandler)

	// 基础验证TCP代理处理器已创建
	assert.Equal(t, proxy.ProxyTypeTCP, proxyHandler.GetType())

	// TCP代理不支持HTTP请求
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	result := proxyHandler.Handle(ctx)
	assert.False(t, result, "TCP代理不应该处理HTTP请求")
}

func TestUDPProxy(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	config := proxy.ProxyConfig{
		ID:      "test-udp",
		Enabled: true,
		Name:    "UDP代理测试",
		Type:    proxy.ProxyTypeUDP,
		Config: map[string]interface{}{
			"buffer_size": 32768,
		},
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	proxyHandler, err := proxyFactory.CreateProxy(config)
	require.NoError(t, err)
	require.NotNil(t, proxyHandler)

	// 基础验证UDP代理处理器已创建
	assert.Equal(t, proxy.ProxyTypeUDP, proxyHandler.GetType())

	// UDP代理不支持HTTP请求
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	result := proxyHandler.Handle(ctx)
	assert.False(t, result, "UDP代理不应该处理HTTP请求")
}

func TestProxyErrorHandling(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	config := proxy.ProxyConfig{
		ID:      "test-error",
		Enabled: true,
		Name:    "错误处理测试",
		Type:    proxy.ProxyTypeHTTP,
		Config: map[string]interface{}{
			"timeout": "5s",
		},
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	proxyHandler, err := proxyFactory.CreateProxy(config)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)
	ctx.SetServiceID("nonexistent-service") // 不存在的服务

	result := proxyHandler.Handle(ctx)
	assert.False(t, result, "连接不存在服务应该失败")
}

func TestDisabledProxy(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	config := proxy.ProxyConfig{
		ID:      "test-disabled",
		Enabled: false,
		Name:    "禁用代理测试",
		Type:    proxy.ProxyTypeHTTP,
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	_, err := proxyFactory.CreateProxy(config)

	// 禁用的代理应该无法创建
	assert.Error(t, err, "禁用的代理应该无法创建")
}

func TestDefaultProxyConfig(t *testing.T) {
	defaultConfig := proxy.DefaultProxyConfig

	// 验证默认配置
	assert.Equal(t, "default-proxy", defaultConfig.ID)
	assert.Equal(t, "Default Proxy", defaultConfig.Name)
	assert.True(t, defaultConfig.Enabled, "默认应该启用代理")
	assert.Equal(t, proxy.ProxyTypeHTTP, defaultConfig.Type, "默认代理类型应该是HTTP")
}

func TestProxyInterface(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	config := proxy.ProxyConfig{
		ID:      "test-interface",
		Enabled: true,
		Name:    "接口测试",
		Type:    proxy.ProxyTypeHTTP,
		Config: map[string]interface{}{
			"timeout": "30s",
		},
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	proxyHandler, err := proxyFactory.CreateProxy(config)
	require.NoError(t, err)

	// 验证代理处理器实现了ProxyHandler接口
	var _ proxy.ProxyHandler = proxyHandler

	// 测试接口方法
	assert.Equal(t, proxy.ProxyTypeHTTP, proxyHandler.GetType())
	assert.True(t, proxyHandler.IsEnabled())
	assert.Equal(t, "接口测试", proxyHandler.GetName())
	assert.NoError(t, proxyHandler.Validate())
}

func TestProxyFactory(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	proxyFactory := proxy.NewProxyFactory(serviceManager)

	// 测试获取支持的代理类型
	supportedTypes := proxyFactory.GetSupportedTypes()
	assert.Contains(t, supportedTypes, proxy.ProxyTypeHTTP)
	assert.Contains(t, supportedTypes, proxy.ProxyTypeWebSocket)
	assert.Contains(t, supportedTypes, proxy.ProxyTypeTCP)
	assert.Contains(t, supportedTypes, proxy.ProxyTypeUDP)

	// 测试获取类型描述
	desc := proxyFactory.GetTypeDescription(proxy.ProxyTypeHTTP)
	assert.NotEmpty(t, desc, "HTTP代理类型描述不应该为空")

	// 测试根据类型创建代理
	proxyHandler, err := proxyFactory.CreateProxyByType(proxy.ProxyTypeHTTP)
	require.NoError(t, err)
	assert.Equal(t, proxy.ProxyTypeHTTP, proxyHandler.GetType())
}

func TestProxyConfigValidation(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	proxyFactory := proxy.NewProxyFactory(serviceManager)

	tests := []struct {
		name        string
		config      proxy.ProxyConfig
		expectError bool
		description string
	}{
		{
			name: "ValidConfig",
			config: proxy.ProxyConfig{
				ID:      "valid-proxy",
				Enabled: true,
				Name:    "有效代理",
				Type:    proxy.ProxyTypeHTTP,
			},
			expectError: false,
			description: "有效配置应该通过验证",
		},
		{
			name: "EmptyName",
			config: proxy.ProxyConfig{
				ID:      "empty-name-proxy",
				Enabled: true,
				Name:    "",
				Type:    proxy.ProxyTypeHTTP,
			},
			expectError: true,
			description: "空名称配置应该验证失败",
		},
		{
			name: "InvalidType",
			config: proxy.ProxyConfig{
				ID:      "invalid-type-proxy",
				Enabled: true,
				Name:    "无效类型代理",
				Type:    "invalid-type",
			},
			expectError: true,
			description: "无效类型配置应该验证失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := proxyFactory.ValidateConfig(tt.config)
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestHTTPProxyValidation(t *testing.T) {
	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	config := proxy.ProxyConfig{
		ID:      "test-http-validation",
		Enabled: true,
		Name:    "HTTP代理验证测试",
		Type:    proxy.ProxyTypeHTTP,
		Config: map[string]interface{}{
			"timeout":        "30s",
			"buffer_size":    32768,
			"max_idle_conns": 100,
		},
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	proxyHandler, err := proxyFactory.CreateProxy(config)
	require.NoError(t, err)

	// 验证配置
	err = proxyHandler.Validate()
	assert.NoError(t, err, "有效的HTTP代理配置应该通过验证")
}

// 基准测试
func BenchmarkHTTPProxy(b *testing.B) {
	// 创建后端测试服务器
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	// 创建负载均衡管理器
	factory := service.NewLoadBalancerFactory()
	serviceManager := factory.CreateServiceManager()

	// 添加服务
	service := &service.ServiceConfig{
		ID:       "bench-service",
		Name:     "基准测试服务",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{
			{
				ID:      "backend1",
				URL:     backend.URL,
				Weight:  1,
				Health:  true,
				Enabled: true,
			},
		},
	}
	err := serviceManager.AddService(service)
	require.NoError(b, err)

	config := proxy.ProxyConfig{
		ID:      "bench-http",
		Enabled: true,
		Name:    "HTTP代理基准测试",
		Type:    proxy.ProxyTypeHTTP,
		Config: map[string]interface{}{
			"timeout": "30s",
		},
	}

	proxyFactory := proxy.NewProxyFactory(serviceManager)
	proxyHandler, err := proxyFactory.CreateProxy(config)
	require.NoError(b, err)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)
		ctx.SetServiceID("bench-service")
		proxyHandler.Handle(ctx)
	}
}
